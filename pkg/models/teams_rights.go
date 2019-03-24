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

// CanCreate checks if the user can create a new team
func (t *Team) CanCreate(a web.Auth) (bool, error) {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true, nil
}

// CanUpdate checks if the user can update a team
func (t *Team) CanUpdate(a web.Auth) (bool, error) {
	u := getUserForRights(a)

	// Check if the current user is in the team and has admin rights in it
	return x.Where("team_id = ?", t.ID).
		And("user_id = ?", u.ID).
		And("admin = ?", true).
		Get(&TeamMember{})
}

// CanDelete checks if a user can delete a team
func (t *Team) CanDelete(a web.Auth) (bool, error) {
	return t.IsAdmin(a)
}

// IsAdmin returns true when the user is admin of a team
func (t *Team) IsAdmin(a web.Auth) (bool, error) {
	u := getUserForRights(a)

	return x.Where("team_id = ?", t.ID).
		And("user_id = ?", u.ID).
		And("admin = ?", true).
		Get(&TeamMember{})
}

// CanRead returns true if the user has read access to the team
func (t *Team) CanRead(a web.Auth) (bool, error) {
	user := getUserForRights(a)

	// Check if the user is in the team
	return x.Where("team_id = ?", t.ID).
		And("user_id = ?", user.ID).
		Get(&TeamMember{})
}
