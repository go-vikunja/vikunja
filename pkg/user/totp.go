// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"image"

	"code.vikunja.io/api/pkg/config"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTP holds a user's totp setting in the database.
type TOTP struct {
	ID     int64  `xorm:"bigint autoincr not null unique pk" json:"-"`
	UserID int64  `xorm:"bigint not null" json:"-"`
	Secret string `xorm:"text not null" json:"secret"`
	// The totp entry will only be enabled after the user verified they have a working totp setup.
	Enabled bool `xorm:"null" json:"enabled"`
	// The totp url used to be able to enroll the user later
	URL string `xorm:"text null" json:"url"`
}

// TableName holds the table name for totp secrets
func (t *TOTP) TableName() string {
	return "totp"
}

// TOTPPasscode is used to validate a users totp passcode
type TOTPPasscode struct {
	User     *User  `json:"-"`
	Passcode string `json:"passcode"`
}

// TOTPEnabledForUser checks if totp is enabled for a user - not if it is activated, use GetTOTPForUser to check that.
func TOTPEnabledForUser(user *User) (bool, error) {
	if !config.ServiceEnableTotp.GetBool() {
		return false, nil
	}
	t := &TOTP{}
	_, err := x.Where("user_id = ?", user.ID).Get(t)
	return t.Enabled, err
}

// GetTOTPForUser returns the current state of totp settings for the user.
func GetTOTPForUser(user *User) (t *TOTP, err error) {
	t = &TOTP{}
	exists, err := x.Where("user_id = ?", user.ID).Get(t)
	if err != nil {
		return
	}
	if !exists {
		return nil, ErrTOTPNotEnabled{}
	}

	return
}

// EnrollTOTP creates a new TOTP entry for the user - it does not enable it yet.
func EnrollTOTP(user *User) (t *TOTP, err error) {
	isEnrolled, err := x.Where("user_id = ?", user.ID).Exist(&TOTP{})
	if err != nil {
		return
	}
	if isEnrolled {
		return nil, ErrTOTPAlreadyEnabled{}
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Vikunja",
		AccountName: user.Username,
	})
	if err != nil {
		return
	}

	t = &TOTP{
		UserID:  user.ID,
		Secret:  key.Secret(),
		Enabled: false,
		URL:     key.URL(),
	}
	_, err = x.Insert(t)
	return
}

// EnableTOTP enables totp for a user. The provided passcode is used to verify the user has a working totp setup.
func EnableTOTP(passcode *TOTPPasscode) (err error) {
	t, err := ValidateTOTPPasscode(passcode)
	if err != nil {
		return
	}

	_, err = x.
		Where("id = ?", t.ID).
		Cols("enabled").
		Update(&TOTP{Enabled: true})
	return
}

// DisableTOTP removes all totp settings for a user.
func DisableTOTP(user *User) (err error) {
	_, err = x.Where("user_id = ?", user.ID).Delete(&TOTP{})
	return
}

// ValidateTOTPPasscode validated totp codes of users.
func ValidateTOTPPasscode(passcode *TOTPPasscode) (t *TOTP, err error) {
	t, err = GetTOTPForUser(passcode.User)
	if err != nil {
		return
	}

	if !totp.Validate(passcode.Passcode, t.Secret) {
		return nil, ErrInvalidTOTPPasscode{Passcode: passcode.Passcode}
	}

	return
}

// GetTOTPQrCodeForUser returns a qrcode for a user's totp setting
func GetTOTPQrCodeForUser(user *User) (qrcode image.Image, err error) {
	t, err := GetTOTPForUser(user)
	if err != nil {
		return
	}

	key, err := otp.NewKeyFromURL(t.URL)
	if err != nil {
		return
	}
	return key.Image(300, 300)
}
