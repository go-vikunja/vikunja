// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/utils"
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

	update.User.IsActive = false
	update.User.Email = update.NewEmail
	update.User.EmailConfirmToken = utils.MakeRandomString(64)
	_, err = s.
		Where("id = ?", update.User.ID).
		Cols("email", "is_active", "email_confirm_token").
		Update(update.User)
	if err != nil {
		return
	}

	// Send the confirmation mail
	if !config.MailerEnabled.GetBool() {
		return
	}

	// Send the user a mail with a link to confirm the mail
	n := &EmailConfirmNotification{
		User:  update.User,
		IsNew: false,
	}

	err = notifications.Notify(update.User, n)
	return
}
