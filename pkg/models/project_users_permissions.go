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
	"xorm.io/xorm"
)

// CanCreate checks if the user can create a new user <-> project relation
func (lu *ProjectUser) CanCreate(s *xorm.Session, u *user.User) (bool, error) {
	return lu.canDoProjectUser(s, u)
}

// CanDelete checks if the user can delete a user <-> project relation
func (lu *ProjectUser) CanDelete(s *xorm.Session, u *user.User) (bool, error) {
	return lu.canDoProjectUser(s, u)
}

// CanUpdate checks if the user can update a user <-> project relation
func (lu *ProjectUser) CanUpdate(s *xorm.Session, u *user.User) (bool, error) {
	return lu.canDoProjectUser(s, u)
}

func (lu *ProjectUser) canDoProjectUser(s *xorm.Session, u *user.User) (bool, error) {
	// Link shares aren't allowed to do anything
	if u == nil {
		return false, nil
	}

	// Get the project and check if the user has write access on it
	l := Project{ID: lu.ProjectID}
	return l.IsAdmin(s, u)
}
