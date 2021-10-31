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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/notifications"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"
)

const IssuerLocal = `local`

// CreateUser creates a new user and inserts it into the database
func CreateUser(s *xorm.Session, user *User) (newUser *User, err error) {

	if user.Issuer == "" {
		user.Issuer = IssuerLocal
	}

	// Check if we have all needed information
	err = checkIfUserIsValid(user)
	if err != nil {
		return nil, err
	}

	// Check if the user already exists with that username
	err = checkIfUserExists(s, user)
	if err != nil {
		return nil, err
	}

	if user.Issuer == IssuerLocal {
		// Hash the password
		user.Password, err = HashPassword(user.Password)
		if err != nil {
			return nil, err
		}
	}

	user.Status = StatusActive
	user.AvatarProvider = "initials"

	// Insert it
	_, err = s.Insert(user)
	if err != nil {
		return nil, err
	}

	// Get the  full new User
	newUserOut, err := GetUserByID(s, user.ID)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&CreatedEvent{
		User: newUserOut,
	})
	if err != nil {
		return nil, err
	}

	// Dont send a mail if no mailer is configured
	if !config.MailerEnabled.GetBool() || user.Issuer != IssuerLocal {
		return newUserOut, err
	}

	user.Status = StatusEmailConfirmationRequired
	token, err := generateNewToken(s, user, TokenEmailConfirm)
	if err != nil {
		return nil, err
	}

	_, err = s.
		Where("id = ?", user.ID).
		Cols("email", "is_active").
		Update(user)
	if err != nil {
		return
	}

	n := &EmailConfirmNotification{
		User:         user,
		IsNew:        true,
		ConfirmToken: token.Token,
	}

	err = notifications.Notify(user, n)
	return newUserOut, err
}

// HashPassword hashes a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(bytes), err
}

func checkIfUserIsValid(user *User) error {
	if user.Email == "" ||
		(user.Issuer != IssuerLocal && user.Subject == "") ||
		(user.Issuer == IssuerLocal && (user.Password == "" ||
			user.Username == "")) {
		return ErrNoUsernamePassword{}
	}

	return nil
}

func checkIfUserExists(s *xorm.Session, user *User) (err error) {
	exists := true
	_, err = GetUserByUsername(s, user.Username)
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return err
		}
	}
	if exists {
		return ErrUsernameExists{user.ID, user.Username}
	}

	// Check if the user already existst with that email
	exists = true
	userToCheck := &User{
		Email:   user.Email,
		Issuer:  user.Issuer,
		Subject: user.Subject,
	}

	if user.Issuer != IssuerLocal {
		userToCheck.Email = ""
	}

	_, err = getUser(s, userToCheck, false)
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return err
		}
	}
	if exists && user.Issuer == IssuerLocal {
		return ErrUserEmailExists{user.ID, user.Email}
	}

	return nil
}
