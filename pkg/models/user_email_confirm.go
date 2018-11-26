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

// EmailConfirm holds the token to confirm a mail address
type EmailConfirm struct {
	Token string `json:"token"`
}

// UserEmailConfirm handles the confirmation of an email address
func UserEmailConfirm(c *EmailConfirm) (err error) {

	// Check if we have an email confirm token
	if c.Token == "" {
		return ErrInvalidEmailConfirmToken{}
	}

	// Check if the token is valid
	user := User{}
	has, err := x.Where("email_confirm_token = ?", c.Token).Get(&user)
	if err != nil {
		return
	}

	if !has {
		return ErrInvalidEmailConfirmToken{Token: c.Token}
	}

	user.IsActive = true
	user.EmailConfirmToken = ""
	_, err = x.Where("id = ?", user.ID).Cols("is_active", "email_confirm_token").Update(&user)
	return
}
