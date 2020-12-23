// Copyright2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import "xorm.io/xorm"

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

	// Check if the token is valid
	user := User{}
	has, err := s.
		Where("email_confirm_token = ?", c.Token).
		Get(&user)
	if err != nil {
		return
	}

	if !has {
		return ErrInvalidEmailConfirmToken{Token: c.Token}
	}

	user.IsActive = true
	user.EmailConfirmToken = ""
	_, err = s.
		Where("id = ?", user.ID).
		Cols("is_active", "email_confirm_token").
		Update(&user)
	return
}
