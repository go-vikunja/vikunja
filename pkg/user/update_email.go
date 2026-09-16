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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
	"xorm.io/xorm"
)

const emailConfirmationResendCooldown = time.Minute

// EmailUpdate is the data structure to update a user's email address
type EmailUpdate struct {
	User *User `json:"-"`
	// The new email address. Needs to be a valid email address.
	NewEmail string `json:"new_email" valid:"email,length(0|250),required"`
	// The password of the user for confirmation.
	Password string `json:"password"`
}

func EmailUpdateMessage() string {
	if config.MailerEnabled.GetBool() {
		return "We sent you an email with a link to confirm your new email address. Your current address stays active until you confirm."
	}

	return "Your email address was updated."
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

// UpdateEmail stages the address as PendingEmail until confirmed; with the mailer off it applies immediately.
func UpdateEmail(s *xorm.Session, update *EmailUpdate) (err error) {
	if err := checkEmailNotTaken(s, update.NewEmail); err != nil {
		return err
	}

	update.User, err = GetUserWithEmail(s, &User{ID: update.User.ID})
	if err != nil {
		return
	}

	if !config.MailerEnabled.GetBool() {
		update.User.Email = update.NewEmail
		update.User.PendingEmail = ""
		_, err = s.
			Where("id = ?", update.User.ID).
			Cols("email", "pending_email").
			Update(update.User)
		return
	}

	update.User.PendingEmail = update.NewEmail
	_, err = s.
		Where("id = ?", update.User.ID).
		Cols("pending_email").
		Update(update.User)
	if err != nil {
		return
	}

	err = sendEmailConfirmation(s, update.User)
	if err != nil {
		return
	}

	return notifications.Notify(update.User, &EmailChangeRequestedNotification{
		User:     update.User,
		NewEmail: update.NewEmail,
	}, s)
}

// CancelEmailUpdate discards a pending email change and its confirm tokens.
func CancelEmailUpdate(s *xorm.Session, u *User) (err error) {
	u, err = GetUserWithEmail(s, &User{ID: u.ID})
	if err != nil {
		return
	}
	if u.PendingEmail == "" {
		return ErrNoPendingEmail{UserID: u.ID}
	}

	err = removeTokens(s, u, TokenEmailConfirm)
	if err != nil {
		return
	}
	_, err = s.
		Where("id = ?", u.ID).
		Cols("pending_email").
		Update(&User{PendingEmail: ""})
	return
}

// ResendEmailConfirmation sends a fresh confirm link for a pending email change.
func ResendEmailConfirmation(s *xorm.Session, u *User) (err error) {
	u, err = GetUserWithEmail(s, &User{ID: u.ID})
	if err != nil {
		return
	}
	if u.PendingEmail == "" {
		return ErrNoPendingEmail{UserID: u.ID}
	}

	return sendEmailConfirmation(s, u)
}

func emailConfirmationCooldownKey(userID int64) string {
	return "email_confirm_sent_" + strconv.FormatInt(userID, 10)
}

func sendEmailConfirmation(s *xorm.Session, u *User) (err error) {
	// Kept outside the db so the cooldown survives cancelling the change or rotating tokens.
	cooldownKey := emailConfirmationCooldownKey(u.ID)
	lastSent, exists, err := keyvalue.Get(cooldownKey)
	if err != nil {
		return
	}
	if exists {
		var sentAt int64
		switch v := lastSent.(type) {
		case int64:
			sentAt = v
		case float64:
			sentAt = int64(v)
		}
		if time.Since(time.Unix(sentAt, 0)) < emailConfirmationResendCooldown {
			return ErrEmailConfirmationCooldown{UserID: u.ID}
		}
	}

	err = removeTokens(s, u, TokenEmailConfirm)
	if err != nil {
		return
	}

	token, err := generateToken(s, u, TokenEmailConfirm)
	if err != nil {
		return
	}

	// RouteForMail sends to User.Email; swap in the pending address on a copy.
	recipient := *u
	recipient.Email = u.PendingEmail

	n := &EmailConfirmNotification{
		User:         &recipient,
		IsNew:        false,
		ConfirmToken: token.ClearTextToken,
	}

	err = notifications.Notify(&recipient, n, s)
	if err != nil {
		return
	}

	return keyvalue.Put(cooldownKey, time.Now().Unix())
}

func checkEmailNotTaken(s *xorm.Session, email string) error {
	user := &User{}
	has, err := s.Where("email = ?", email).Get(user)
	if err != nil {
		return err
	}
	if has {
		return ErrUserEmailExists{UserID: user.ID, Email: email}
	}
	return nil
}
