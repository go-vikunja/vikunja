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

package models

import (
	"math"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"
)

type TaskPosition struct {
	// The ID of the task this position is for
	TaskID int64 `xorm:"bigint not null index" json:"task_id" param:"task"`
	// The project view this task is related to
	ProjectViewID int64 `xorm:"bigint not null index" json:"project_view_id"`
	// The position of the task - any task project can be sorted as usual by this parameter.
	// When accessing tasks via kanban buckets, this is primarily used to sort them based on a range
	// We're using a float64 here to make it possible to put any task within any two other tasks (by changing the number).
	// You would calculate the new position between two tasks with something like task3.position = (task2.position - task1.position) / 2.
	// A 64-Bit float leaves plenty of room to initially give tasks a position with 2^16 difference to the previous task
	// which also leaves a lot of room for rearranging and sorting later.
	// Positions are always saved per view. They will automatically be set if you request the tasks through a view
	// endpoint, otherwise they will always be 0. To update them, take a look at the Task Position endpoint.
	Position float64 `xorm:"double not null" json:"position"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (tp *TaskPosition) TableName() string {
	return "task_positions"
}

func (tp *TaskPosition) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: tp.TaskID}
	return t.CanUpdate(s, a)
}

// Update is the handler to update a task position
// @Summary Updates a task position
// @Description Updates a task position.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Param view body models.TaskPosition true "The task position with updated values you want to change."
// @Success 200 {object} models.TaskPosition "The updated task position."
// @Failure 400 {object} web.HTTPError "Invalid task position object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id}/position [post]
func (tp *TaskPosition) Update(s *xorm.Session, a web.Auth) (err error) {

	// Update all positions if the newly saved position is < 0.1
	var shouldRecalculate bool
	var view *ProjectView
	if tp.Position < 0.1 {
		shouldRecalculate = true
		view, err = GetProjectViewByID(s, tp.ProjectViewID)
		if err != nil {
			return err
		}
	}

	exists, err := s.
		Where("task_id = ? AND project_view_id = ?", tp.TaskID, tp.ProjectViewID).
		Exist(&TaskPosition{})
	if err != nil {
		return err
	}

	if !exists {
		_, err = s.Insert(tp)
		if err != nil {
			return
		}
		if shouldRecalculate {
			return RecalculateTaskPositions(s, view, a)
		}
		return nil
	}

	_, err = s.
		Where("task_id = ? AND project_view_id = ?", tp.TaskID, tp.ProjectViewID).
		Cols("project_view_id", "position").
		Update(tp)
	if err != nil {
		return
	}

	if shouldRecalculate {
		return RecalculateTaskPositions(s, view, a)
	}

	return triggerTaskUpdatedEventForTaskID(s, a, tp.TaskID)
}

func RecalculateTaskPositions(s *xorm.Session, view *ProjectView, a web.Auth) (err error) {

	log.Debugf("Recalculating task positions for view %d", view.ID)

	opts := &taskSearchOptions{
		projectViewID: view.ID,
		sortby: []*sortParam{
			{
				projectViewID: view.ID,
				sortBy:        taskPropertyPosition,
				orderBy:       orderAscending,
			},
			{
				sortBy:  taskPropertyID,
				orderBy: orderAscending,
			},
		},
	}

	// Using the collection so that we get all tasks, even in cases where we're dealing with a saved filter underneath
	tc := &TaskCollection{
		ProjectID: view.ProjectID,
	}
	if view.ProjectID < -1 {
		tc.ProjectID = 0

		sf, err := GetSavedFilterSimpleByID(s, GetSavedFilterIDFromProjectID(view.ProjectID))
		if err != nil {
			return err
		}

		opts.filterIncludeNulls = sf.Filters.FilterIncludeNulls
		opts.filterTimezone = sf.Filters.FilterTimezone
		opts.filter = sf.Filters.Filter
		opts.parsedFilters, err = getTaskFiltersFromFilterString(opts.filter, opts.filterTimezone)
		if err != nil {
			return err
		}
	}

	projects, err := getRelevantProjectsFromCollection(s, a, tc)
	if err != nil {
		return err
	}

	for _, p := range projects {
		opts.projectIDs = append(opts.projectIDs, p.ID)
	}

	dbSearcher := &dbTaskSearcher{
		s: s,
		a: a,
	}

	// We're directly using the db here, even if Typesense is configured, because in some edge cases Typesense
	// does not know about all tasks. These tasks then won't have their position recalculated, which means they will
	// seemingly jump around after reloading their project.
	// The real fix here is of course to make sure all tasks are indexed in Typesense, but until that's fixed,
	// this solves the issue of task positions not being saved.
	allTasks, _, err := dbSearcher.Search(opts)
	if err != nil {
		return
	}
	if len(allTasks) == 0 {
		return
	}

	maxPosition := math.Pow(2, 32)
	newPositions := make([]*TaskPosition, 0, len(allTasks))

	for i, task := range allTasks {

		currentPosition := maxPosition / float64(len(allTasks)) * (float64(i + 1))

		newPositions = append(newPositions, &TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			Position:      currentPosition,
		})
	}

	_, err = s.
		Where("project_view_id = ?", view.ID).
		Delete(&TaskPosition{})
	if err != nil {
		return
	}

	count, err := s.Insert(newPositions)
	if err != nil {
		return
	}

	log.Debugf("Inserted %d new positions for %d total tasks in view %d", count, len(allTasks), view.ID)

	return events.Dispatch(&TaskPositionsRecalculatedEvent{
		NewTaskPositions: newPositions,
	})
}

func getPositionsForView(s *xorm.Session, view *ProjectView) (positions []*TaskPosition, err error) {
	positions = []*TaskPosition{}
	err = s.
		Where("project_view_id = ?", view.ID).
		Find(&positions)
	return
}

func calculateNewPositionForTask(s *xorm.Session, a web.Auth, t *Task, view *ProjectView) (*TaskPosition, error) {
	if t.Position == 0 {
		lowestPosition := &TaskPosition{}
		exists, err := s.Where("project_view_id = ?", view.ID).
			OrderBy("position asc").
			Get(lowestPosition)
		if err != nil {
			return nil, err
		}
		if exists {
			if lowestPosition.Position == 0 {
				err = RecalculateTaskPositions(s, view, a)
				if err != nil {
					return nil, err
				}

				lowestPosition = &TaskPosition{}
				_, err = s.Where("project_view_id = ?", view.ID).
					OrderBy("position asc").
					Get(lowestPosition)
				if err != nil {
					return nil, err
				}
			}

			t.Position = lowestPosition.Position / 2
		}
	}

	return &TaskPosition{
		TaskID:        t.ID,
		ProjectViewID: view.ID,
		Position:      calculateDefaultPosition(t.Index, t.Position),
	}, nil
}

func DeleteOrphanedTaskPositions(s *xorm.Session) (count int64, err error) {
	return s.
		Where("task_id not in (select id from tasks) OR project_view_id not in (select id from project_views)").
		Delete(&TaskPosition{})
}
