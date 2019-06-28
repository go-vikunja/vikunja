//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
)

// CanCreate checks if the user can create a team <-> list relation
func (tl *TeamList) CanCreate(a web.Auth) (bool, error) {
	return tl.canDoTeamList(a)
}

// CanDelete checks if the user can delete a team <-> list relation
func (tl *TeamList) CanDelete(a web.Auth) (bool, error) {
	return tl.canDoTeamList(a)
}

// CanUpdate checks if the user can update a team <-> list relation
func (tl *TeamList) CanUpdate(a web.Auth) (bool, error) {
	return tl.canDoTeamList(a)
}

func (tl *TeamList) canDoTeamList(a web.Auth) (bool, error) {
	l := List{ID: tl.ListID}
	return l.IsAdmin(a)
}
