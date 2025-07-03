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
	"xorm.io/xorm"
)

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

	user, err := GetUserByID(s, token.UserID)
	if err != nil {
		return
	}

	user.Status = StatusActive
	err = removeTokens(s, user, TokenEmailConfirm)
	if err != nil {
		return
	}
	_, err = s.
		Where("id = ?", user.ID).
		Cols("status").
		Update(user)
	return
}
