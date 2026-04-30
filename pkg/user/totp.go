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
	"fmt"
	"image"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"xorm.io/xorm"
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
func TOTPEnabledForUser(s *xorm.Session, user *User) (bool, error) {
	if !config.ServiceEnableTotp.GetBool() {
		return false, nil
	}
	t := &TOTP{}
	_, err := s.Where("user_id = ?", user.ID).Get(t)
	return t.Enabled, err
}

// GetTOTPForUser returns the current state of totp settings for the user.
func GetTOTPForUser(s *xorm.Session, user *User) (t *TOTP, err error) {
	t = &TOTP{}
	exists, err := s.Where("user_id = ?", user.ID).Get(t)
	if err != nil {
		return
	}
	if !exists {
		return nil, ErrTOTPNotEnabled{}
	}

	return
}

// EnrollTOTP creates a new TOTP entry for the user - it does not enable it yet.
func EnrollTOTP(s *xorm.Session, user *User) (t *TOTP, err error) {
	isEnrolled, err := s.Where("user_id = ?", user.ID).Exist(&TOTP{})
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
	_, err = s.Insert(t)
	return
}

// EnableTOTP enables totp for a user. The provided passcode is used to verify the user has a working totp setup.
func EnableTOTP(s *xorm.Session, passcode *TOTPPasscode) (err error) {
	t, err := ValidateTOTPPasscode(s, passcode)
	if err != nil {
		return
	}

	_, err = s.
		Where("id = ?", t.ID).
		Cols("enabled").
		Update(&TOTP{Enabled: true})
	return
}

// DisableTOTP removes all totp settings for a user.
func DisableTOTP(s *xorm.Session, user *User) (err error) {
	_, err = s.
		Where("user_id = ?", user.ID).
		Delete(&TOTP{})
	return
}

// ValidateTOTPPasscode validated totp codes of users.
func ValidateTOTPPasscode(s *xorm.Session, passcode *TOTPPasscode) (t *TOTP, err error) {
	t, err = GetTOTPForUser(s, passcode.User)
	if err != nil {
		return
	}

	if !totp.Validate(passcode.Passcode, t.Secret) {
		return nil, ErrInvalidTOTPPasscode{Passcode: passcode.Passcode}
	}

	// Prevent passcode reuse within the validity window.
	// Store the timestamp when the passcode was used; treat entries older than
	// 90 seconds (30s TOTP window + clock skew) as expired.
	const totpUsedTTL = 90 * time.Second
	usedKey := fmt.Sprintf("totp_used_%s_%s", strconv.FormatInt(passcode.User.ID, 10), passcode.Passcode)
	val, exists, err := keyvalue.Get(usedKey)
	if err != nil {
		return nil, err
	}
	if exists {
		if usedAt, ok := val.(int64); ok && time.Since(time.Unix(usedAt, 0)) < totpUsedTTL {
			return nil, ErrTOTPPasscodeUsed{}
		}
		// Entry expired — allow reuse, overwrite below
	}

	// Mark this passcode as used with the current timestamp
	err = keyvalue.Put(usedKey, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	// Lazily clean up expired entries to prevent unbounded growth
	go cleanupExpiredTOTPKeys(totpUsedTTL)

	return
}

func cleanupExpiredTOTPKeys(ttl time.Duration) {
	keys, err := keyvalue.ListKeys("totp_used_")
	if err != nil {
		return
	}
	for _, key := range keys {
		val, exists, err := keyvalue.Get(key)
		if err != nil || !exists {
			continue
		}
		if usedAt, ok := val.(int64); ok && time.Since(time.Unix(usedAt, 0)) >= ttl {
			_ = keyvalue.Del(key)
		}
	}
}

// GetTOTPQrCodeForUser returns a qrcode for a user's totp setting
func GetTOTPQrCodeForUser(s *xorm.Session, user *User) (qrcode image.Image, err error) {
	t, err := GetTOTPForUser(s, user)
	if err != nil {
		return
	}

	key, err := otp.NewKeyFromURL(t.URL)
	if err != nil {
		return
	}
	return key.Image(300, 300)
}

// HandleFailedTOTPAuth records a failed TOTP attempt and locks the account
// after 10 consecutive failures.
//
// Must not share the caller's session: the login handler rolls its session
// back on auth failure, which would discard the lockout write
// (GHSA-fgfv-pv97-6cmj). Opens its own session and commits independently.
func HandleFailedTOTPAuth(user *User) {
	log.Errorf("Invalid TOTP credentials provided for user %d", user.ID)

	key := user.GetFailedTOTPAttemptsKey()
	if err := keyvalue.IncrBy(key, 1); err != nil {
		log.Errorf("Could not increase failed TOTP attempts for user %d: %s", user.ID, err)
		return
	}

	a, _, err := keyvalue.Get(key)
	if err != nil {
		log.Errorf("Could not get failed TOTP attempts for user %d: %s", user.ID, err)
		return
	}
	// Redis backend returns the counter as a string; in-memory as int64.
	attempts, ok := a.(int64)
	if !ok {
		attemptsStr, ok := a.(string)
		if !ok {
			log.Errorf("Unexpected type for failed TOTP attempts for user %d: %T", user.ID, a)
			return
		}
		attempts, err = strconv.ParseInt(attemptsStr, 10, 64)
		if err != nil {
			log.Errorf("Could not convert failed TOTP attempts to int64 for user %d: %v, value: %s", user.ID, err, attemptsStr)
			return
		}
	}

	if attempts == 3 {
		s := db.NewSession()
		defer s.Close()
		if err := notifications.Notify(user, &InvalidTOTPNotification{User: user}, s); err != nil {
			log.Errorf("Could not send failed TOTP notification to user %d: %s", user.ID, err)
			_ = s.Rollback()
			return
		}
		if err := s.Commit(); err != nil {
			log.Errorf("Could not commit failed TOTP notification for user %d: %s", user.ID, err)
		}
		return
	}

	if attempts < 10 {
		return
	}

	log.Infof("Blocking user account %d after 10 failed TOTP password attempts", user.ID)
	s := db.NewSession()
	defer s.Close()

	if err := RequestUserPasswordResetToken(s, user); err != nil {
		log.Errorf("Could not issue password reset token for user %d after 10 failed TOTP attempts: %s", user.ID, err)
		_ = s.Rollback()
		return
	}
	if err := notifications.Notify(user, &PasswordAccountLockedAfterInvalidTOTPNotification{User: user}, s); err != nil {
		log.Errorf("Could not send password information mail to user %d after 10 failed TOTP attempts: %s", user.ID, err)
		_ = s.Rollback()
		return
	}
	if err := user.SetStatus(s, StatusAccountLocked); err != nil {
		log.Errorf("Could not lock user %d: %s", user.ID, err)
		_ = s.Rollback()
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("Could not commit lockout for user %d: %s", user.ID, err)
	}
}
