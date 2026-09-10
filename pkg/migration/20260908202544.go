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
	"fmt"
	"slices"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type projectAncestors20260908202544 struct {
	AncestorID int64 `xorm:"bigint not null pk"`
	ProjectID  int64 `xorm:"bigint not null pk index"`
	Depth      int64 `xorm:"int not null"`
}

func (projectAncestors20260908202544) TableName() string {
	return "project_ancestors"
}

type projects20260908202544 struct {
	ID              int64  `xorm:"bigint autoincr not null unique pk"`
	ParentProjectID *int64 `xorm:"bigint INDEX null"`
}

func (projects20260908202544) TableName() string {
	return "projects"
}

const projectAncestorsBatch20260908202544 = 500

// Frozen copy of models.RebuildProjectAncestors - migrations must not follow later model changes.
func backfillProjectAncestors20260908202544(tx *xorm.Engine) error {
	allProjects := []*projects20260908202544{}
	if err := tx.Cols("id", "parent_project_id").Find(&allProjects); err != nil {
		return fmt.Errorf("could not get projects: %w", err)
	}

	parents := make(map[int64]int64, len(allProjects))
	for _, project := range allProjects {
		// Roots are null now, but rows written before that still store 0.
		if project.ParentProjectID != nil {
			parents[project.ID] = *project.ParentProjectID
			continue
		}
		parents[project.ID] = 0
	}

	// Replacing one batch at a time lets a re-run drop rows which drifted, without the whole
	// table ever being empty: Migrate has no transaction, so a replica may already serve from it.
	for chunk := range slices.Chunk(allProjects, projectAncestorsBatch20260908202544) {
		ids := make([]int64, 0, len(chunk))
		rows := make([]*projectAncestors20260908202544, 0, len(chunk))

		for _, project := range chunk {
			ids = append(ids, project.ID)
			rows = append(rows, &projectAncestors20260908202544{ProjectID: project.ID, AncestorID: project.ID})

			visited := map[int64]bool{project.ID: true}
			depth := int64(0)
			for current := parents[project.ID]; current != 0; current = parents[current] {
				// A missing parent or a cycle ends the chain: the remaining links are treated
				// as absent so the rows still describe a tree.
				if _, exists := parents[current]; !exists || visited[current] {
					break
				}
				visited[current] = true
				depth++
				rows = append(rows, &projectAncestors20260908202544{ProjectID: project.ID, AncestorID: current, Depth: depth})
			}
		}

		if _, err := tx.Where(builder.In("project_id", ids)).Delete(&projectAncestors20260908202544{}); err != nil {
			return fmt.Errorf("could not delete existing project ancestors: %w", err)
		}

		for insertChunk := range slices.Chunk(rows, projectAncestorsBatch20260908202544) {
			if _, err := tx.Insert(insertChunk); err != nil {
				return fmt.Errorf("could not insert project ancestors: %w", err)
			}
		}
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260908202544",
		Description: "Add and backfill the project_ancestors closure table",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync(projectAncestors20260908202544{}); err != nil { //nolint:forbidigo // brand-new table, nothing to drop
				return fmt.Errorf("could not create the project_ancestors table: %w", err)
			}

			return backfillProjectAncestors20260908202544(tx)
		},
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
