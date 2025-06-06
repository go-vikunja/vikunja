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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
	"xorm.io/xorm"
)

// PasswordReset holds the data to reset a password
type PasswordReset struct {
	// The previously issued reset token.
	Token string `json:"token"`
	// The new password for this user.
	NewPassword string `json:"new_password"`
}

// ResetPassword resets a users password
func ResetPassword(s *xorm.Session, reset *PasswordReset) (err error) {

	// Check if the password is not empty
	if reset.NewPassword == "" {
		return ErrNoUsernamePassword{}
	}

	if reset.Token == "" {
		return ErrNoPasswordResetToken{}
	}

	// Check if we have a token
	token, err := getToken(s, reset.Token, TokenPasswordReset)
	if err != nil {
		return err
	}
	if token == nil {
		return ErrInvalidPasswordResetToken{Token: reset.Token}
	}

	user, err := GetUserByID(s, token.UserID)
	if err != nil {
		return
	}

	// Hash the password
	user.Password, err = HashPassword(reset.NewPassword)
	if err != nil {
		return
	}

	err = removeTokens(s, user, TokenEmailConfirm)
	if err != nil {
		return
	}

	user.Status = StatusActive
	_, err = s.
		Cols("password", "status").
		Where("id = ?", user.ID).
		Update(user)
	if err != nil {
		return
	}

	// Dont send a mail if no mailer is configured
	if !config.MailerEnabled.GetBool() {
		return
	}

	// Send a mail to the user to notify it his password was changed.
	n := &PasswordChangedNotification{
		User: user,
	}

	err = notifications.Notify(user, n)
	return
}

// PasswordTokenRequest defines the request format for password reset resqest
type PasswordTokenRequest struct {
	Email string `json:"email" valid:"email,length(0|250)" maxLength:"250"`
}

// RequestUserPasswordResetTokenByEmail inserts a random token to reset a users password into the databsse
func RequestUserPasswordResetTokenByEmail(s *xorm.Session, tr *PasswordTokenRequest) (err error) {
	if tr.Email == "" {
		return ErrNoUsernamePassword{}
	}

	// Check if the user exists
	user, err := GetUserWithEmail(s, &User{Email: tr.Email})
	if err != nil {
		return
	}

	return RequestUserPasswordResetToken(s, user)
}

// RequestUserPasswordResetToken sends a user a password reset email.
func RequestUserPasswordResetToken(s *xorm.Session, user *User) (err error) {
	token, err := generateToken(s, user, TokenPasswordReset)
	if err != nil {
		return
	}

	// Dont send a mail if no mailer is configured
	if !config.MailerEnabled.GetBool() {
		return
	}

	n := &ResetPasswordNotification{
		User:  user,
		Token: token,
	}

	err = notifications.Notify(user, n)
	return
}
