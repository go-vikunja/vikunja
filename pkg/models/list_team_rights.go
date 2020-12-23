// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

// CanCreate checks if the user can create a team <-> list relation
func (tl *TeamList) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return tl.canDoTeamList(s, a)
}

// CanDelete checks if the user can delete a team <-> list relation
func (tl *TeamList) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return tl.canDoTeamList(s, a)
}

// CanUpdate checks if the user can update a team <-> list relation
func (tl *TeamList) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return tl.canDoTeamList(s, a)
}

func (tl *TeamList) canDoTeamList(s *xorm.Session, a web.Auth) (bool, error) {
	// Link shares aren't allowed to do anything
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	l := List{ID: tl.ListID}
	return l.IsAdmin(s, a)
}
