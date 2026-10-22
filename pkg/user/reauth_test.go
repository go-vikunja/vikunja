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
	"errors"
	"testing"

	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
