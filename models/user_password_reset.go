package models

import (
	"code.vikunja.io/api/models/mail"
	"code.vikunja.io/api/models/utils"
)

// PasswordReset holds the data to reset a password
type PasswordReset struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// UserPasswordReset resets a users password
func UserPasswordReset(reset *PasswordReset) (err error) {

	// Check if the password is not empty
	if reset.NewPassword == "" {
		return ErrNoUsernamePassword{}
	}

	// Check if we have a token
	var user User
	exists, err := x.Where("password_reset_token = ?", reset.Token).Get(&user)
	if err != nil {
		return
	}

	if !exists {
		return ErrInvalidPasswordResetToken{Token: reset.Token}
	}

	// Hash the password
	user.Password, err = hashPassword(reset.NewPassword)
	if err != nil {
		return
	}

	// Save it
	_, err = x.Where("id = ?", user.ID).Update(&user)
	if err != nil {
		return
	}

	// Dont send a mail if we're testing
	if IsTesting {
		return
	}

	// Send a mail to the user to notify it his password was changed.
	data := map[string]interface{}{
		"User": user,
	}

	mail.SendMailWithTemplate(user.Email, "Your password on Vikunja was changed", "password-changed", data)

	return
}

// PasswordTokenRequest defines the request format for password reset resqest
type PasswordTokenRequest struct {
	Username string `json:"user_name"`
}

// RequestUserPasswordResetToken inserts a random token to reset a users password into the databsse
func RequestUserPasswordResetToken(tr *PasswordTokenRequest) (err error) {
	// Check if the user exists
	user, err := GetUser(User{Username: tr.Username})
	if err != nil {
		return
	}

	// Generate a token and save it
	user.PasswordResetToken = utils.MakeRandomString(400)

	// Save it
	_, err = x.Where("id = ?", user.ID).Update(&user)
	if err != nil {
		return
	}

	// Dont send a mail if we're testing
	if IsTesting {
		return
	}

	data := map[string]interface{}{
		"User": user,
	}

	// Send the user a mail with the reset token
	mail.SendMailWithTemplate(user.Email, "Reset your password on Vikunja", "reset-password", data)
	return
}
