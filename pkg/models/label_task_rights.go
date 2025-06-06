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

// CanCreate checks if a user can add a label to a task
func (lt *LabelTask) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	label, err := getLabelByIDSimple(s, lt.LabelID)
	if err != nil {
		return false, err
	}

	hasAccessTolabel, _, err := label.hasAccessToLabel(s, a)
	if err != nil || !hasAccessTolabel { // If the user doesn't have access to the label, we can error out here
		return false, err
	}

	canDoLabelTask, err := canDoLabelTask(s, lt.TaskID, a)
	if err != nil {
		return false, err
	}

	return hasAccessTolabel && canDoLabelTask, nil
}

// CanDelete checks if a user can delete a label from a task
func (lt *LabelTask) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	canDoLabelTask, err := canDoLabelTask(s, lt.TaskID, a)
	if err != nil {
		return false, err
	}
	if !canDoLabelTask {
		return false, nil
	}

	// We don't care here if the label exists or not. The only relevant thing here is if the relation already exists,
	// throw an error.
	exists, err := s.Exist(&LabelTask{LabelID: lt.LabelID, TaskID: lt.TaskID})
	if err != nil {
		return false, err
	}
	return exists, err
}

// CanCreate determines if a user can update a labeltask
func (ltb *LabelTaskBulk) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return canDoLabelTask(s, ltb.TaskID, a)
}

// Helper function to check if a user can write to a task
// + is able to see the label
// always the same check for either deleting or adding a label to a task
func canDoLabelTask(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	// A user can add a label to a task if he can write to the task
	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return false, err
	}
	return task.CanUpdate(s, a)
}
