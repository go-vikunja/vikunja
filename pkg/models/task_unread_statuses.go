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

type TaskUnreadStatus struct {
	TaskID          int64 `xorm:"bigint not null unique(task_user)" param:"projecttask"`
	UserID          int64 `xorm:"bigint not null unique(task_user)"`
	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*TaskUnreadStatus) TableName() string {
	return "task_unread_statuses"
}

func (t *TaskUnreadStatus) CanUpdate(_ *xorm.Session, _ web.Auth) (bool, error) {
	return true, nil
}

// Update marks a task as read
// @Summary Mark a task as read
// @Description Marks a task as read for the current user by removing the unread status entry.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Success 200 {object} models.TaskUnreadStatus "The task unread status object."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/read [post]
func (t *TaskUnreadStatus) Update(s *xorm.Session, a web.Auth) error {
	return markTaskAsRead(s, t.TaskID, a)
}

func markTaskAsRead(s *xorm.Session, taskID int64, a web.Auth) error {
	_, err := s.Where("task_id = ? AND user_id = ?", taskID, a.GetID()).
		Delete(&TaskUnreadStatus{})

	return err
}
