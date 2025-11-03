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

// CanCreate checks if the user can create a new team
func (t *Team) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	// This is currently a dummy function, later on we could imagine global limits etc.
	return true, nil
}

// CanUpdate checks if the user can update a team
func (t *Team) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return t.IsAdmin(s, a)
}

// CanDelete checks if a user can delete a team
func (t *Team) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return t.IsAdmin(s, a)
}

// IsAdmin returns true when the user is admin of a team
func (t *Team) IsAdmin(s *xorm.Session, a web.Auth) (bool, error) {
	// Don't do anything if we're deadling with a link share auth here
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	// Check if the team exists to be able to return a proper error message if not
	_, err := GetTeamByID(s, t.ID)
	if err != nil {
		return false, err
	}

	return s.Where("team_id = ?", t.ID).
		And("user_id = ?", a.GetID()).
		And("admin = ?", true).
		Get(&TeamMember{})
}

// CanRead returns true if the user has read access to the team
func (t *Team) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	// Check if the user is in the team
	tm := &TeamMember{}
	can, err := s.
		Where("team_id = ?", t.ID).
		And("user_id = ?", a.GetID()).
		Get(tm)

	maxPermissions := 0
	if tm.Admin {
		maxPermissions = int(PermissionAdmin)
	}

	return can, maxPermissions, err
}
