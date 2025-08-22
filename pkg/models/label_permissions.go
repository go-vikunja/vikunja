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
	"xorm.io/builder"
	"xorm.io/xorm"
)

// CanUpdate checks if a user can update a label
func (l *Label) CanUpdate(s *xorm.Session, u *user.User) (bool, error) {
	return l.isLabelOwner(s, u) // Only owners should be allowed to update a label
}

// CanDelete checks if a user can delete a label
func (l *Label) CanDelete(s *xorm.Session, u *user.User) (bool, error) {
	return l.isLabelOwner(s, u) // Only owners should be allowed to delete a label
}

// CanRead checks if a user can read a label
func (l *Label) CanRead(s *xorm.Session, u *user.User) (bool, int, error) {
	return l.hasAccessToLabel(s, u)
}

// CanCreate checks if the user can create a label
// Currently a dummy.
func (l *Label) CanCreate(_ *xorm.Session, u *user.User) (bool, error) {
	if u == nil {
		return false, nil
	}

	return true, nil
}

func (l *Label) isLabelOwner(s *xorm.Session, u *user.User) (bool, error) {
	if u == nil {
		return false, nil
	}

	lorig, err := getLabelByIDSimple(s, l.ID)
	if err != nil {
		return false, err
	}
	return lorig.CreatedByID == u.ID, nil
}

// Helper method to check if a user can see a specific label
func (l *Label) hasAccessToLabel(s *xorm.Session, u *user.User) (has bool, maxPermission int, err error) {
	var where builder.Cond
	var createdByID int64
	if u != nil {
		where = builder.In("project_id", getUserProjectsStatement(u.ID, "", false).Select("l.id"))
		createdByID = u.ID
	}

	cond := builder.In("label_tasks.task_id",
		builder.
			Select("id").
			From("tasks").
			Where(where),
	)

	ll := &LabelTask{}
	has, err = s.Table("labels").
		Select("label_tasks.*").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where("label_tasks.label_id is not null OR labels.created_by_id = ?", createdByID).
		Or(cond).
		And("labels.id = ?", l.ID).
		Exist(ll)
	if err != nil {
		return
	}

	// Since the permission depends on the task the label is associated with, we need to check that too.
	if ll.TaskID > 0 {
		t := &Task{ID: ll.TaskID}
		_, maxPermission, err = t.CanRead(s, u)
		if err != nil {
			return
		}
	}

	return
}
