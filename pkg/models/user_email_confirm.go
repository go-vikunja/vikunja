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
