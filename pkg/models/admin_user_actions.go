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

// loadAdminTargetUser fetches a user by ID for the admin actions, returning
// ErrUserDoesNotExist for an invalid ID or a missing row.
func loadAdminTargetUser(s *xorm.Session, id int64) (*user.User, error) {
	if id < 1 {
		return nil, user.ErrUserDoesNotExist{UserID: id}
	}
	target := &user.User{ID: id}
	has, err := s.Get(target)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, user.ErrUserDoesNotExist{UserID: id}
	}
	return target, nil
}

// SetUserAdminFlag sets a user's instance-admin flag. Demoting the last
// reachable admin is refused via GuardLastAdmin. It does not commit; the caller
// owns the transaction.
func SetUserAdminFlag(s *xorm.Session, id int64, isAdmin bool) (*user.User, error) {
	target, err := loadAdminTargetUser(s, id)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		if err := user.GuardLastAdmin(s, target); err != nil {
			return nil, err
		}
	}

	target.IsAdmin = isAdmin
	if _, err := s.ID(target.ID).Cols("is_admin").Update(target); err != nil {
		return nil, err
	}
	return target, nil
}

// SetUserStatusAsAdmin sets a user's account status. Moving the last reachable
// admin out of Active is refused via GuardLastAdmin (any non-Active status
// blocks login, so it is equivalent to demotion). It does not commit; the caller
// owns the transaction.
func SetUserStatusAsAdmin(s *xorm.Session, id int64, status user.Status) (*user.User, error) {
	target, err := loadAdminTargetUser(s, id)
	if err != nil {
		return nil, err
	}

	if target.IsAdmin && status != user.StatusActive {
		if err := user.GuardLastAdmin(s, target); err != nil {
			return nil, err
		}
	}

	if err := user.SetUserStatus(s, target, status); err != nil {
		return nil, err
	}
	// Reflect the change on the returned struct; GetUserByID refuses disabled accounts.
	target.Status = status
	return target, nil
}

// DeleteUserAsAdmin removes a user. mode "now" deletes immediately; any other
// value triggers the email-confirmation self-deletion flow. Deleting the last
// reachable admin is refused via GuardLastAdmin. It does not commit; the caller
// owns the transaction.
func DeleteUserAsAdmin(s *xorm.Session, id int64, mode string) error {
	target, err := loadAdminTargetUser(s, id)
	if err != nil {
		return err
	}

	if err := user.GuardLastAdmin(s, target); err != nil {
		return err
	}

	if mode == "now" {
		return DeleteUser(s, target)
	}
	return user.RequestDeletion(s, target)
}
