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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
)

const (
	reauthMaxFailures = 10
	reauthWindow      = 15 * time.Minute
)

func reauthAttemptsKey(userID int64) string {
	return "reauth_attempts_" + strconv.FormatInt(userID, 10)
}

// Keyed on the user, not the IP: a stolen session must not allow brute-forcing the password or totp code.
func ThrottleReauth(userID int64, check func() error) error {
	key := reauthAttemptsKey(userID)
	// Reserve before checking so concurrent guesses cannot overshoot the limit.
	attempts, err := keyvalue.IncrByWithTTL(key, 1, reauthWindow)
	if err != nil {
		return err
	}
	if attempts > reauthMaxFailures {
		refundReauthAttempt(key)
		return ErrTooManyReauthAttempts{UserID: userID}
	}

	err = check()
	if !IsErrWrongUsernameOrPassword(err) && !IsErrInvalidTOTPPasscode(err) {
		refundReauthAttempt(key)
	}
	return err
}

func refundReauthAttempt(key string) {
	if _, err := keyvalue.IncrByWithTTL(key, -1, reauthWindow); err != nil {
		log.Errorf("Could not refund re-authentication attempt for %s: %s", key, err)
	}
}

func CheckReauthPassword(u *User, password string) error {
	return ThrottleReauth(u.ID, func() error {
		return CheckUserPassword(u, password)
	})
}
