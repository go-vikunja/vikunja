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

// CanCreate checks if the user can create a team <-> project relation
func (tl *TeamProject) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return tl.canDoTeamProject(s, a)
}

// CanDelete checks if the user can delete a team <-> project relation
func (tl *TeamProject) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return tl.canDoTeamProject(s, a)
}

// CanUpdate checks if the user can update a team <-> project relation
func (tl *TeamProject) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return tl.canDoTeamProject(s, a)
}

func (tl *TeamProject) canDoTeamProject(s *xorm.Session, a web.Auth) (bool, error) {
	// Link shares aren't allowed to do anything
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	l := Project{ID: tl.ProjectID}
	return l.IsAdmin(s, a)
}
