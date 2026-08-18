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
	"time"

	"xorm.io/xorm"
)

// Confirm tokens are exempt from the cleanup cron so registration links never expire; change links must.
const emailChangeTokenValidity = 24 * time.Hour

// EmailConfirm holds the token to confirm a mail address
type EmailConfirm struct {
	// The email confirm token sent via email.
	Token string `json:"token"`
}

// ConfirmEmail handles the confirmation of an email address
func ConfirmEmail(s *xorm.Session, c *EmailConfirm) (err error) {

	// Check if we have an email confirm token
	if c.Token == "" {
		return ErrInvalidEmailConfirmToken{}
	}

	token, err := getToken(s, c.Token, TokenEmailConfirm)
	if err != nil {
		return
	}
	if token == nil {
		return ErrInvalidEmailConfirmToken{Token: c.Token}
	}

	if token.UserID < 1 {
		return ErrInvalidEmailConfirmToken{Token: c.Token}
	}

	user, err := GetUserWithEmail(s, &User{ID: token.UserID})
	if err != nil {
		// A locked account may still activate its registration, but not swap its login address.
		if !IsErrAccountLocked(err) || user.PendingEmail != "" {
			return err
		}
	}

	var cols []string
	if user.PendingEmail != "" {
		if time.Since(token.Created) > emailChangeTokenValidity {
			if err := removeTokenByID(s, user, TokenEmailConfirm, token.ID); err != nil {
				return err
			}
			return ErrInvalidEmailConfirmToken{Token: c.Token}
		}

		// Someone else may have claimed the address since the change was requested.
		if err := checkEmailNotTaken(s, user.PendingEmail); err != nil {
			return err
		}
		user.Email = user.PendingEmail
		user.PendingEmail = ""
		cols = append(cols, "email", "pending_email")
	}

	err = removeTokens(s, user, TokenEmailConfirm)
	if err != nil {
		return
	}

	if user.Status == StatusEmailConfirmationRequired {
		user.Status = StatusActive
		cols = append(cols, "status")
	}

	if len(cols) == 0 {
		return ErrInvalidEmailConfirmToken{Token: c.Token}
	}

	_, err = s.
		Where("id = ?", user.ID).
		Cols(cols...).
		Update(user)
	return
}
