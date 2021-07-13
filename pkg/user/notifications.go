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
)

// EmailConfirmNotification represents a EmailConfirmNotification notification
type EmailConfirmNotification struct {
	User         *User
	IsNew        bool
	ConfirmToken string
}

// ToMail returns the mail notification for EmailConfirmNotification
func (n *EmailConfirmNotification) ToMail() *notifications.Mail {

	subject := n.User.GetName() + ", please confirm your email address at Vikunja"
	if n.IsNew {
		subject = n.User.GetName() + " + Vikunja = <3"
	}

	nn := notifications.NewMail().
		Subject(subject).
		Greeting("Hi " + n.User.GetName() + ",")

	if n.IsNew {
		nn.Line("Welcome to Vikunja!")
	}

	return nn.
		Line("To confirm your email address, click the link below:").
		Action("Confirm your email address", config.ServiceFrontendurl.GetString()+"?userEmailConfirm="+n.ConfirmToken).
		Line("Have a nice day!")
}

// ToDB returns the EmailConfirmNotification notification in a format which can be saved in the db
func (n *EmailConfirmNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *EmailConfirmNotification) Name() string {
	return ""
}

// PasswordChangedNotification represents a PasswordChangedNotification notification
type PasswordChangedNotification struct {
	User *User
}

// ToMail returns the mail notification for PasswordChangedNotification
func (n *PasswordChangedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Your Password on Vikunja was changed").
		Greeting("Hi " + n.User.GetName() + ",").
		Line("Your account password was successfully changed.").
		Line("If this wasn't you, it could mean someone compromised your account. In this case contact your server's administrator.")
}

// ToDB returns the PasswordChangedNotification notification in a format which can be saved in the db
func (n *PasswordChangedNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *PasswordChangedNotification) Name() string {
	return ""
}

// ResetPasswordNotification represents a ResetPasswordNotification notification
type ResetPasswordNotification struct {
	User  *User
	Token *Token
}

// ToMail returns the mail notification for ResetPasswordNotification
func (n *ResetPasswordNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Reset your password on Vikunja").
		Greeting("Hi "+n.User.GetName()+",").
		Line("To reset your password, click the link below:").
		Action("Reset your password", config.ServiceFrontendurl.GetString()+"?userPasswordReset="+n.Token.Token).
		Line("This link will be valid for 24 hours.").
		Line("Have a nice day!")
}

// ToDB returns the ResetPasswordNotification notification in a format which can be saved in the db
func (n *ResetPasswordNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *ResetPasswordNotification) Name() string {
	return ""
}
