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

package user

import (
	"context"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
	"xorm.io/xorm"
)

// EmailUpdate is the data structure to update a user's email address
type EmailUpdate struct {
	User *User `json:"-"`
	// The new email address. Needs to be a valid email address.
	NewEmail string `json:"new_email" valid:"email,length(0|250),required"`
	// The password of the user for confirmation.
	Password string `json:"password"`
}

// ChangeUserEmail verifies the user's password, then sets a new email address
// (kicking off confirmation when the mailer is enabled). Shared by the v1 and
// v2 email-update handlers; only HTTP input binding stays in the handlers.
func ChangeUserEmail(ctx context.Context, s *xorm.Session, u *User, password, newEmail string) error {
	verified, err := CheckUserCredentials(ctx, s, &Login{Username: u.Username, Password: password})
	if err != nil {
		return err
	}
	return UpdateEmail(s, &EmailUpdate{User: verified, NewEmail: newEmail})
}

// UpdateEmail lets a user update their email address
func UpdateEmail(s *xorm.Session, update *EmailUpdate) (err error) {

	// Check the email is not already used
	user := &User{}
	has, err := s.Where("email = ?", update.NewEmail).Get(user)
	if err != nil {
		return
	}

	if has {
		return ErrUserEmailExists{UserID: user.ID, Email: update.NewEmail}
	}

	// Set the user as unconfirmed and the new email address
	update.User, err = GetUserWithEmail(s, &User{ID: update.User.ID})
	if err != nil {
		return
	}

	update.User.Email = update.NewEmail

	// Send the confirmation mail
	if !config.MailerEnabled.GetBool() {
		update.User.Status = StatusActive
		_, err = s.
			Where("id = ?", update.User.ID).
			Cols("email", "status").
			Update(update.User)
		return
	}

	update.User.Status = StatusEmailConfirmationRequired
	_, err = s.
		Where("id = ?", update.User.ID).
		Cols("email", "status").
		Update(update.User)
	if err != nil {
		return
	}

	token, err := generateToken(s, update.User, TokenEmailConfirm)
	if err != nil {
		return
	}

	// Send the user a mail with a link to confirm the mail
	n := &EmailConfirmNotification{
		User:         update.User,
		IsNew:        false,
		ConfirmToken: token.ClearTextToken,
	}

	err = notifications.Notify(update.User, n, s)
	return
}
