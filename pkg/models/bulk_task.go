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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// BulkTask is the definition of a bulk update task
type BulkTask struct {
	// A project of task ids to update
	IDs   []int64 `json:"task_ids"`
	Tasks []*Task `json:"-"`
	Task
}

// BulkTaskServiceProvider defines the interface for bulk task service operations
type BulkTaskServiceProvider interface {
	GetTasksByIDs(s *xorm.Session, taskIDs []int64) ([]*Task, error)
	CanUpdate(s *xorm.Session, taskIDs []int64, a web.Auth) (bool, error)
	Update(s *xorm.Session, taskIDs []int64, taskUpdate *Task, assignees []*user.User, a web.Auth) error
}

var bulkTaskService BulkTaskServiceProvider

// RegisterBulkTaskService injects the service implementation into the models layer
func RegisterBulkTaskService(service BulkTaskServiceProvider) {
	bulkTaskService = service
}

// @Deprecated: Use BulkTaskService.GetTasksByIDs instead
func (bt *BulkTask) checkIfTasksAreOnTheSameProject(s *xorm.Session) (err error) {
	// Get the tasks - delegate to service
	tasks, err := bulkTaskService.GetTasksByIDs(s, bt.IDs)
	if err != nil {
		return err
	}
	bt.Tasks = tasks

	if len(bt.Tasks) == 0 {
		return ErrBulkTasksNeedAtLeastOne{}
	}

	// Check if all tasks are in the same project
	var firstProjectID = bt.Tasks[0].ProjectID
	for _, t := range bt.Tasks {
		if t.ProjectID != firstProjectID {
			return ErrBulkTasksMustBeInSameProject{firstProjectID, t.ProjectID}
		}
	}

	return nil
}

// @Deprecated: Use BulkTaskService.CanUpdate instead
// CanUpdate checks if a user is allowed to update a task
func (bt *BulkTask) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return bulkTaskService.CanUpdate(s, bt.IDs, a)
}

// @Deprecated: Use BulkTaskService.Update instead
// Update updates a bunch of tasks at once
// @Summary Update a bunch of tasks at once
// @Description Updates a bunch of tasks at once. This includes marking them as done. Note: although you could supply another ID, it will be ignored. Use task_ids instead.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task body models.BulkTask true "The task object. Looks like a normal task, the only difference is it uses an array of project_ids to update."
// @Success 200 {object} models.Task "The updated task object."
// @Failure 400 {object} web.HTTPError "Invalid task object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task (aka its project)"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/bulk [post]
func (bt *BulkTask) Update(s *xorm.Session, a web.Auth) (err error) {
	return bulkTaskService.Update(s, bt.IDs, &bt.Task, bt.Assignees, a)
}
