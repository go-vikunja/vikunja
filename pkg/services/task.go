// Vikunja is a to-do list application to facilitate your life.
// Adding a comment to force a recompile and check the line number of the error.
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
	"xorm.io/xorm"
)

// TaskService represents a service for managing tasks.
type TaskService struct {
	DB *xorm.Engine
}

// NewTaskService creates a new TaskService.
func NewTaskService(db *xorm.Engine) *TaskService {
	return &TaskService{DB: db}
}

// Update updates a task.
func (ts *TaskService) Update(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	can, err := ts.Can(s, task, u).Write()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	// The old logic used task.Update which did a lot of things.
	// We need to replicate that logic here.
	// For now, we'll just do a simple update.
	if _, err := s.ID(task.ID).AllCols().Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

// TaskPermissions represents the permissions for a task.
type TaskPermissions struct {
	s    *xorm.Session
	task *models.Task
	user *user.User
}

// Can returns a new TaskPermissions struct.
func (ts *TaskService) Can(s *xorm.Session, task *models.Task, u *user.User) *TaskPermissions {
	return &TaskPermissions{s: s, task: task, user: u}
}

// Read checks if the user can read the task.
func (tp *TaskPermissions) Read() (bool, error) {
	if tp.user == nil {
		return false, nil
	}
	can, _, err := tp.task.CanRead(tp.s, tp.user)
	return can, err
}

// Write checks if the user can write to the task.
func (tp *TaskPermissions) Write() (bool, error) {
	if tp.user == nil {
		return false, nil
	}
	can, err := tp.task.CanWrite(tp.s, tp.user)
	return can, err
}
