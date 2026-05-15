// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"math"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// projectViewKindKanban20260410120000 matches the iota value from pkg/models/project_view.go:
// ProjectViewKindList = 0, ProjectViewKindGantt = 1, ProjectViewKindTable = 2, ProjectViewKindKanban = 3
const projectViewKindKanban20260410120000 = 3

// getDescendantProjectIDs20260410120000 returns all descendant project IDs for a given parent project
// by traversing the project hierarchy iteratively using breadth-first search.
// Copied verbatim from pkg/models/task_collection.go
func getDescendantProjectIDs20260410120000(s *xorm.Session, parentProjectID int64) ([]int64, error) {
	var allDescendants []int64
	queue := []int64{parentProjectID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		var childIDs []int64
		err := s.Table("projects").Cols("id").Where("parent_project_id = ?", currentID).Find(&childIDs)
		if err != nil {
			return nil, err
		}

		allDescendants = append(allDescendants, childIDs...)
		queue = append(queue, childIDs...)
	}

	return allDescendants, nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260410120000",
		Description: "Migrate task positions to root project views for hierarchical display",
		Migrate: func(tx *xorm.Engine) error {
			s := tx.NewSession()
			defer s.Close()

			if err := s.Begin(); err != nil {
				return err
			}

			// Step 1: Delete all positions from child project views
			_, err := s.Exec(`
				DELETE FROM task_positions
				WHERE project_view_id IN (
					SELECT pv.id
					FROM project_views pv
					INNER JOIN projects p ON pv.project_id = p.id
					WHERE p.parent_project_id != 0
				)
			`)
			if err != nil {
				return err
			}

			// Step 2: For each root project view (excluding kanban), add positions
			// for child project tasks that don't already have positions

			type rootView struct {
				ViewID    int64 `xorm:"view_id"`
				ProjectID int64 `xorm:"project_id"`
			}
			var rootViews []rootView
			err = s.SQL(`
				SELECT pv.id as view_id, pv.project_id
				FROM project_views pv
				INNER JOIN projects p ON pv.project_id = p.id
				WHERE p.parent_project_id = 0
				  AND pv.view_kind != ?
			`, projectViewKindKanban20260410120000).Find(&rootViews)
			if err != nil {
				return err
			}

			for _, rv := range rootViews {
				// Get all descendant project IDs (not including the root itself)
				descendants, err := getDescendantProjectIDs20260410120000(s, rv.ProjectID)
				if err != nil {
					return err
				}

				if len(descendants) == 0 {
					continue
				}

				// Get existing task IDs that already have positions in this view
				var existingTaskIDs []int64
				err = s.Table("task_positions").
					Cols("task_id").
					Where("project_view_id = ?", rv.ViewID).
					Find(&existingTaskIDs)
				if err != nil {
					return err
				}
				existingSet := make(map[int64]bool, len(existingTaskIDs))
				for _, id := range existingTaskIDs {
					existingSet[id] = true
				}

				// Get tasks from child projects, ordered by priority/due_date/created
				var allChildTaskIDs []int64
				err = s.Table("tasks").
					In("project_id", descendants).
					Where("done = ?", false).
					OrderBy("priority DESC, CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC, created ASC").
					Cols("id").
					Find(&allChildTaskIDs)
				if err != nil {
					return err
				}

				// Filter to only tasks that don't already have positions
				var taskIDs []int64
				for _, id := range allChildTaskIDs {
					if !existingSet[id] {
						taskIDs = append(taskIDs, id)
					}
				}

				if len(taskIDs) == 0 {
					continue
				}

				// Get the current maximum position to place new tasks after existing ones
				var maxPosition float64
				_, err = s.SQL(`
					SELECT COALESCE(MAX(position), 0) FROM task_positions WHERE project_view_id = ?
				`, rv.ViewID).Get(&maxPosition)
				if err != nil {
					return err
				}

				// Calculate positions for new tasks, placing them after existing tasks
				spacing := math.Pow(2, 32) / float64(len(taskIDs)+1)
				for i, taskID := range taskIDs {
					position := maxPosition + spacing*float64(i+1)
					_, err = s.Exec(
						"INSERT INTO task_positions (task_id, project_view_id, position) VALUES (?, ?, ?)",
						taskID, rv.ViewID, position,
					)
					if err != nil {
						return err
					}
				}
			}

			return s.Commit()
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
