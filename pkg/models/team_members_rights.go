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

// CanCreate checks if the user can add a new tem member
func (tm *TeamMember) CanCreate(a web.Auth) (bool, error) {
	return tm.IsAdmin(a)
}

// CanDelete checks if the user can delete a new team member
func (tm *TeamMember) CanDelete(a web.Auth) (bool, error) {
	return tm.IsAdmin(a)
}

// IsAdmin checks if the user is team admin
func (tm *TeamMember) IsAdmin(a web.Auth) (bool, error) {
	// A user can add a member to a team if he is admin of that team
	exists, err := x.Where("user_id = ? AND team_id = ? AND admin = ?", a.GetID(), tm.TeamID, true).
		Get(&TeamMember{})
	return exists, err
}
