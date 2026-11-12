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
	"errors"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestThrottleReauth(t *testing.T) {
	wrong := func() error { return ErrWrongUsernameOrPassword{} }
	ok := func() error { return nil }

	t.Run("blocks after the failure limit, even the right password", func(t *testing.T) {
		const userID = 9001
		_ = keyvalue.Del(reauthAttemptsKey(userID))

		for i := 0; i < reauthMaxFailures; i++ {
			err := ThrottleReauth(userID, wrong)
			require.True(t, IsErrWrongUsernameOrPassword(err), "attempt %d: %v", i, err)
		}

		called := false
		err := ThrottleReauth(userID, func() error {
			called = true
			return nil
		})
		assert.True(t, IsErrTooManyReauthAttempts(err), "got %v", err)
		assert.False(t, called)
	})

	t.Run("counts wrong totp codes", func(t *testing.T) {
		const userID = 9002
		_ = keyvalue.Del(reauthAttemptsKey(userID))

		for i := 0; i < reauthMaxFailures; i++ {
			_ = ThrottleReauth(userID, func() error { return ErrInvalidTOTPPasscode{} })
		}
		assert.True(t, IsErrTooManyReauthAttempts(ThrottleReauth(userID, ok)))
	})

	t.Run("successes and other errors are refunded", func(t *testing.T) {
		const userID = 9003
		_ = keyvalue.Del(reauthAttemptsKey(userID))

		for i := 0; i < reauthMaxFailures*2; i++ {
			require.NoError(t, ThrottleReauth(userID, ok))
			require.EqualError(t, ThrottleReauth(userID, func() error { return errors.New("db down") }), "db down")
		}
		for i := 0; i < reauthMaxFailures-1; i++ {
			_ = ThrottleReauth(userID, wrong)
		}
		require.NoError(t, ThrottleReauth(userID, ok))
	})

	t.Run("is keyed per user", func(t *testing.T) {
		const userID, other = 9004, 9005
		_ = keyvalue.Del(reauthAttemptsKey(userID))
		_ = keyvalue.Del(reauthAttemptsKey(other))

		for i := 0; i < reauthMaxFailures; i++ {
			_ = ThrottleReauth(userID, wrong)
		}
		require.NoError(t, ThrottleReauth(other, ok))
	})
}

func setupReauthUser(t *testing.T, id int64) (*xorm.Session, *User) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	t.Cleanup(func() { s.Close() })
	u, err := GetUserByID(s, id)
	require.NoError(t, err)
	require.NoError(t, keyvalue.Del(reauthAttemptsKey(u.ID)))
	require.NoError(t, keyvalue.Del(u.GetFailedTOTPAttemptsKey()))
	require.NoError(t, keyvalue.Del(u.GetFailedPasswordAttemptsKey()))
	return s, u
}

func kvCount(t *testing.T, key string) int64 {
	n, _, err := keyvalue.GetInt64(key)
	require.NoError(t, err)
	return n
}

func TestReauthenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("a wrong password counts toward the lockout and the throttle", func(t *testing.T) {
		s, u := setupReauthUser(t, 10)

		err := Reauthenticate(ctx, s, u, "wrong", "")
		assert.True(t, IsErrWrongUsernameOrPassword(err), "got %v", err)
		assert.Equal(t, int64(1), kvCount(t, u.GetFailedPasswordAttemptsKey()))
		assert.Equal(t, int64(1), kvCount(t, reauthAttemptsKey(u.ID)))
	})

	t.Run("a user without TOTP only needs the password", func(t *testing.T) {
		s, u := setupReauthUser(t, 1)

		require.NoError(t, Reauthenticate(ctx, s, u, "12345678", ""))
		assert.Equal(t, int64(0), kvCount(t, reauthAttemptsKey(u.ID)))
	})

	t.Run("a missing passcode is refused without counting toward the TOTP lockout", func(t *testing.T) {
		s, u := setupReauthUser(t, 10)

		err := Reauthenticate(ctx, s, u, "12345678", "")
		assert.True(t, IsErrInvalidTOTPPasscode(err), "got %v", err)
		assert.Equal(t, int64(0), kvCount(t, u.GetFailedTOTPAttemptsKey()))
		assert.Equal(t, int64(1), kvCount(t, reauthAttemptsKey(u.ID)))
	})
}
