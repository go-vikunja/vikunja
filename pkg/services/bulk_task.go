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

package services

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"dario.cat/mergo"
	"xorm.io/xorm"
)

// BulkTaskService handles bulk task update operations
type BulkTaskService struct {
	DB          *xorm.Engine
	TaskService *TaskService
}

// NewBulkTaskService creates a new BulkTaskService
func NewBulkTaskService(db *xorm.Engine) *BulkTaskService {
	return &BulkTaskService{
		DB:          db,
		TaskService: NewTaskService(db),
	}
}

// GetTasksByIDs retrieves tasks by their IDs
func (bts *BulkTaskService) GetTasksByIDs(s *xorm.Session, taskIDs []int64) (tasks []*models.Task, err error) {
	// Validate all IDs are positive
	for _, id := range taskIDs {
		if id < 1 {
			return nil, models.ErrTaskDoesNotExist{ID: id}
		}
	}

	// Fetch tasks
	tasks = []*models.Task{}
	err = s.In("id", taskIDs).Find(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// checkIfTasksAreOnTheSameProject verifies all tasks belong to the same project
func (bts *BulkTaskService) checkIfTasksAreOnTheSameProject(tasks []*models.Task) error {
	if len(tasks) == 0 {
		return models.ErrBulkTasksNeedAtLeastOne{}
	}

	// Check if all tasks are in the same project
	firstProjectID := tasks[0].ProjectID
	for _, t := range tasks {
		if t.ProjectID != firstProjectID {
			return models.ErrBulkTasksMustBeInSameProject{
				ShouldBeID: firstProjectID,
				IsID:       t.ProjectID,
			}
		}
	}

	return nil
}

// CanUpdate checks if a user is allowed to bulk update tasks
func (bts *BulkTaskService) CanUpdate(s *xorm.Session, taskIDs []int64, a web.Auth) (bool, error) {
	// Get the tasks
	tasks, err := bts.GetTasksByIDs(s, taskIDs)
	if err != nil {
		return false, err
	}

	// Check if all tasks are on the same project
	err = bts.checkIfTasksAreOnTheSameProject(tasks)
	if err != nil {
		return false, err
	}

	// Check if user has write access to the project
	project := &models.Project{ID: tasks[0].ProjectID}
	return project.CanWrite(s, a)
}

// Update updates a bunch of tasks at once
// NOTE: This method does NOT check permissions - call CanUpdate first
// The same-project validation is only done in CanUpdate, not in Update
func (bts *BulkTaskService) Update(s *xorm.Session, taskIDs []int64, taskUpdate *models.Task, assignees []*user.User, a web.Auth) error {
	// Get the tasks
	tasks, err := bts.GetTasksByIDs(s, taskIDs)
	if err != nil {
		return err
	}

	// Update each task (no validation - CanUpdate should be called first by the handler)
	for _, oldTask := range tasks {
		// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
		models.UpdateDone(oldTask, taskUpdate)

		// Update the assignees
		if err := oldTask.UpdateTaskAssignees(s, assignees, a); err != nil {
			return err
		}

		// Merge the update into the old task
		// For whatever reason, xorm doesn't detect if done is updated, so we need to update this every time by hand
		// Which is why we merge the actual task struct with the one we got from the user
		// The user struct overrides values in the actual one.
		if err := mergo.Merge(oldTask, taskUpdate, mergo.WithOverride); err != nil {
			return err
		}

		// And because a false is considered to be a null value, we need to explicitly check that case here.
		if !taskUpdate.Done {
			oldTask.Done = false
		}

		// Save the updated task
		_, err = s.ID(oldTask.ID).
			Cols("title",
				"description",
				"done",
				"due_date",
				"reminders",
				"repeat_after",
				"priority",
				"start_date",
				"end_date").
			Update(oldTask)
		if err != nil {
			return err
		}
	}

	return nil
}
