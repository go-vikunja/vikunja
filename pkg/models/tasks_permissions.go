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

// CanDelete checks if the user can delete an task
func (t *Task) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canDoTask(s, a)
}

// CanUpdate determines if a user has the permission to update a project task
func (t *Task) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canDoTask(s, a)
}

// CanCreate determines if a user has the permission to create a project task
func (t *Task) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	// A user can do a task if he has write acces to its project
	l := &Project{ID: t.ProjectID}
	return l.CanWrite(s, a)
}

// CanRead determines if a user can read a task
func (t *Task) CanRead(s *xorm.Session, a web.Auth) (canRead bool, maxPermission int, err error) {
	expand := t.Expand
	if err = t.resolveIDFromProjectAndIndex(s); err != nil {
		return
	}
	*t, err = getTaskByIDForAuth(s, a, t.ID)
	if err != nil {
		return
	}

	t.Expand = expand

	// A user can read a task if it has access to the project
	l := &Project{ID: t.ProjectID}
	return l.CanRead(s, a)
}

// CanWrite checks if a user has write access to a task
func (t *Task) CanWrite(s *xorm.Session, a web.Auth) (canWrite bool, err error) {
	return t.canDoTask(s, a)
}

func (t *Task) canDoTrashedTask(s *xorm.Session, a web.Auth, can func(*Project, *xorm.Session, web.Auth) (bool, error)) (bool, error) {
	trashed, err := getReadableTrashedTask(s, a, t.ID)
	if err != nil {
		return false, err
	}
	// CanWrite lets instance admins past archived projects
	if err := (&Project{ID: trashed.ProjectID}).CheckIsArchived(s); err != nil {
		return false, err
	}
	*t = *trashed
	return can(&Project{ID: t.ProjectID}, s, a)
}

// CanRestore checks whether the auth may restore the trashed task.
func (t *Task) CanRestore(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canDoTrashedTask(s, a, (*Project).CanWrite)
}

// CanDeletePermanently checks whether the auth is admin of the trashed task's project.
// Link shares have no trash, so they can't.
func (t *Task) CanDeletePermanently(s *xorm.Session, a web.Auth) (bool, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return false, nil
	}
	return t.canDoTrashedTask(s, a, (*Project).IsAdmin)
}

// Helper function to check if a user can do stuff on a project task
func (t *Task) canDoTask(s *xorm.Session, a web.Auth) (bool, error) {
	ot, err := getTaskByIDForAuth(s, a, t.ID)
	if err != nil {
		return false, err
	}

	// Check if we're moving the task into a different project to check if the user has sufficient permissions for that on the new project
	if t.ProjectID != 0 && t.ProjectID != ot.ProjectID {
		newProject := &Project{ID: t.ProjectID}
		can, err := newProject.CanWrite(s, a)
		if err != nil {
			return false, err
		}
		if !can {
			return false, ErrGenericForbidden{}
		}
	}

	// A user can do a task if it has write acces to its project
	l := &Project{ID: ot.ProjectID}
	return l.CanWrite(s, a)
}
