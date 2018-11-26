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
	"code.vikunja.io/api/pkg/log"
)

// CanCreate checks if the user can create a new team
func (t *Team) CanCreate(u *User) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}

// CanUpdate checks if the user can update a team
func (t *Team) CanUpdate(u *User) bool {

	// Check if the current user is in the team and has admin rights in it
	exists, err := x.Where("team_id = ?", t.ID).
		And("user_id = ?", u.ID).
		And("admin = ?", true).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Team: %s", err)
		return false
	}

	return exists
}

// CanDelete checks if a user can delete a team
func (t *Team) CanDelete(u *User) bool {
	return t.IsAdmin(u)
}

// IsAdmin returns true when the user is admin of a team
func (t *Team) IsAdmin(u *User) bool {
	exists, err := x.Where("team_id = ?", t.ID).
		And("user_id = ?", u.ID).
		And("admin = ?", true).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Team: %s", err)
		return false
	}
	return exists
}

// CanRead returns true if the user has read access to the team
func (t *Team) CanRead(user *User) bool {
	// Check if the user is in the team
	exists, err := x.Where("team_id = ?", t.ID).
		And("user_id = ?", user.ID).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Team: %s", err)
		return false
	}
	return exists
}
