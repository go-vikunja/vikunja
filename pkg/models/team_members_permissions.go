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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// CanCreate checks if the user can add a new tem member
func (tm *TeamMember) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return tm.IsAdmin(s, a)
}

// CanDelete checks if the user can delete a new team member
func (tm *TeamMember) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	u, err := user.GetUserByUsername(s, tm.Username)
	if err != nil {
		return false, err
	}
	if u.ID == a.GetID() {
		return true, nil
	}
	return tm.IsAdmin(s, a)
}

// CanUpdate checks if the user can modify a team member's permission
func (tm *TeamMember) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return tm.IsAdmin(s, a)
}

// IsAdmin checks if the user is team admin
func (tm *TeamMember) IsAdmin(s *xorm.Session, a web.Auth) (bool, error) {
	// Don't allow anything if we're dealing with a project share here
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	// A user can add a member to a team if he is admin of that team
	exists, err := s.
		Where("user_id = ? AND team_id = ? AND admin = ?", a.GetID(), tm.TeamID, true).
		Get(&TeamMember{})
	return exists, err
}
