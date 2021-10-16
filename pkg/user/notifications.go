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
	"strconv"

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

// InvalidTOTPNotification represents a InvalidTOTPNotification notification
type InvalidTOTPNotification struct {
	User *User
}

// ToMail returns the mail notification for InvalidTOTPNotification
func (n *InvalidTOTPNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Someone just tried to login to your Vikunja account, but failed").
		Greeting("Hi "+n.User.GetName()+",").
		Line("Someone just tried to log in into your account with correct username and password but a wrong TOTP passcode.").
		Line("**If this was not you, someone else knows your password. You should set a new one immediately!**").
		Action("Reset your password", config.ServiceFrontendurl.GetString()+"get-password-reset")
}

// ToDB returns the InvalidTOTPNotification notification in a format which can be saved in the db
func (n *InvalidTOTPNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *InvalidTOTPNotification) Name() string {
	return "totp.invalid"
}

// PasswordAccountLockedAfterInvalidTOTOPNotification represents a PasswordAccountLockedAfterInvalidTOTOPNotification notification
type PasswordAccountLockedAfterInvalidTOTOPNotification struct {
	User *User
}

// ToMail returns the mail notification for PasswordAccountLockedAfterInvalidTOTOPNotification
func (n *PasswordAccountLockedAfterInvalidTOTOPNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("We've disabled your account on Vikunja").
		Greeting("Hi " + n.User.GetName() + ",").
		Line("Someone tried to log in with your credentials but failed to provide a valid TOTP passcode.").
		Line("After 10 failed attempts, we've disabled your account and reset your password. To set a new one, follow the instructions in the reset email we just sent you.").
		Line("If you did not receive an email with reset instructions, you can always request a new one at [" + config.ServiceFrontendurl.GetString() + "get-password-reset](" + config.ServiceFrontendurl.GetString() + "get-password-reset).")
}

// ToDB returns the PasswordAccountLockedAfterInvalidTOTOPNotification notification in a format which can be saved in the db
func (n *PasswordAccountLockedAfterInvalidTOTOPNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *PasswordAccountLockedAfterInvalidTOTOPNotification) Name() string {
	return "password.account.locked.after.invalid.totop"
}

// FailedLoginAttemptNotification represents a FailedLoginAttemptNotification notification
type FailedLoginAttemptNotification struct {
	User *User
}

// ToMail returns the mail notification for FailedLoginAttemptNotification
func (n *FailedLoginAttemptNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Someone just tried to login to your Vikunja account, but failed to provide a correct password").
		Greeting("Hi "+n.User.GetName()+",").
		Line("Someone just tried to log in into your account with a wrong password three times in a row.").
		Line("If this was not you, this could be someone else trying to break into your account.").
		Line("To enhance the security of you account you may want to set a stronger password or enable TOTP authentication in the settings:").
		Action("Go to settings", config.ServiceFrontendurl.GetString()+"user/settings")
}

// ToDB returns the FailedLoginAttemptNotification notification in a format which can be saved in the db
func (n *FailedLoginAttemptNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *FailedLoginAttemptNotification) Name() string {
	return "failed.login.attempt"
}

// AccountDeletionConfirmNotification represents a AccountDeletionConfirmNotification notification
type AccountDeletionConfirmNotification struct {
	User         *User
	ConfirmToken string
}

// ToMail returns the mail notification for AccountDeletionConfirmNotification
func (n *AccountDeletionConfirmNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Please confirm the deletion of your Vikunja account").
		Greeting("Hi "+n.User.GetName()+",").
		Line("You have requested the deletion of your account. To confirm this, please click the link below:").
		Action("Confirm the deletion of my account", config.ServiceFrontendurl.GetString()+"?accountDeletionConfirm="+n.ConfirmToken).
		Line("This link will be valid for 24 hours.").
		Line("Once you confirm the deletion we will schedule the deletion of your account in three days and send you another email until then.").
		Line("If you proceed with the deletion of your account, we will remove all of your namespaces, lists and tasks you created. Everything you shared with another user or team will transfer ownership to them.").
		Line("If you did not requested the deletion or changed your mind, you can simply ignore this email.").
		Line("Have a nice day!")
}

// ToDB returns the AccountDeletionConfirmNotification notification in a format which can be saved in the db
func (n *AccountDeletionConfirmNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *AccountDeletionConfirmNotification) Name() string {
	return "user.deletion.confirm"
}

// AccountDeletionNotification represents a AccountDeletionNotification notification
type AccountDeletionNotification struct {
	User               *User
	NotificationNumber int
}

// ToMail returns the mail notification for AccountDeletionNotification
func (n *AccountDeletionNotification) ToMail() *notifications.Mail {
	durationString := "in " + strconv.Itoa(n.NotificationNumber) + " days"

	if n.NotificationNumber == 1 {
		durationString = "tomorrow"
	}

	return notifications.NewMail().
		Subject("Your Vikunja account will be deleted "+durationString).
		Greeting("Hi "+n.User.GetName()+",").
		Line("You recently requested the deletion of your Vikunja account.").
		Line("We will delete your account "+durationString+".").
		Line("If you changed your mind, simply click the link below to cancel the deletion and follow the instructions there:").
		Action("Abort the deletion", config.ServiceFrontendurl.GetString()).
		Line("Have a nice day!")
}

// ToDB returns the AccountDeletionNotification notification in a format which can be saved in the db
func (n *AccountDeletionNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *AccountDeletionNotification) Name() string {
	return "user.deletion"
}

// AccountDeletedNotification represents a AccountDeletedNotification notification
type AccountDeletedNotification struct {
	User *User
}

// ToMail returns the mail notification for AccountDeletedNotification
func (n *AccountDeletedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Your Vikunja Account has been deleted").
		Greeting("Hi " + n.User.GetName() + ",").
		Line("As requested, we've deleted your Vikunja account.").
		Line("This deletion is permanent. If did not create a backup and need your data back now, talk to your administrator.").
		Line("Have a nice day!")
}

// ToDB returns the AccountDeletedNotification notification in a format which can be saved in the db
func (n *AccountDeletedNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *AccountDeletedNotification) Name() string {
	return "user.deleted"
}
