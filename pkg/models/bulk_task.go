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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// BulkTask represents a bulk task update payload.
type BulkTask struct {
	*Task   `json:"-" xorm:"-"`
	TaskIDs []int64  `json:"task_ids"`
	Fields  []string `json:"fields"`
	Values  *Task    `json:"values"`
	Tasks   []*Task  `json:"tasks,omitempty"`
}

// CanUpdate checks if the user can update all provided tasks.
func (bt *BulkTask) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	tasks, err := GetTasksSimpleByIDs(s, bt.TaskIDs)
	if err != nil {
		return false, err
	}
	if len(tasks) == 0 {
		return false, ErrBulkTasksNeedAtLeastOne{}
	}
	// ensure user can write to each involved project
	projects := map[int64]struct{}{}
	for _, t := range tasks {
		projects[t.ProjectID] = struct{}{}
	}
	for pid := range projects {
		l := &Project{ID: pid}
		can, err := l.CanWrite(s, a)
		if err != nil || !can {
			return false, err
		}
	}
	// if tasks are moved to another project, check destination permission
	if bt.Values != nil && bt.Values.ProjectID != 0 {
		l := &Project{ID: bt.Values.ProjectID}
		can, err := l.CanWrite(s, a)
		if err != nil || !can {
			return false, err
		}
	}
	return true, nil
}

// Update updates multiple tasks at once.
// @Summary Update multiple tasks
// @Description Updates multiple tasks atomically. All provided tasks must be writable by the user.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param bulkTask body models.BulkTask true "Bulk task update payload"
// @Success 200 {array} models.Task "Updated tasks"
// @Failure 400 {object} web.HTTPError "Invalid request"
// @Failure 403 {object} web.HTTPError "The user does not have access to the tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/bulk [post]
func (bt *BulkTask) Update(s *xorm.Session, a web.Auth) (err error) {
	if bt.Values == nil {
		bt.Values = &Task{}
	}
	tasks, err := updateTasks(s, a, bt.Values, bt.TaskIDs, bt.Fields)
	if err != nil {
		return err
	}
	bt.Tasks = tasks
	return nil
}
