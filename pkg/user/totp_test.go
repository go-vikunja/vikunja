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
	"fmt"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/utils"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOTPPasscodeCannotBeReused(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Generate a valid TOTP passcode for user10's secret from the fixture
	// user10 has TOTP enabled with this secret in pkg/db/fixtures/totp.yml
	secret := "JBSWY3DPEHPK3PXP" //nolint:gosec
	passcode, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	user := &User{ID: 10}

	// First use should succeed
	_, err = ValidateTOTPPasscode(s, &TOTPPasscode{
		User:     user,
		Passcode: passcode,
	})
	require.NoError(t, err)

	// Second use of the same passcode should fail
	_, err = ValidateTOTPPasscode(s, &TOTPPasscode{
		User:     user,
		Passcode: passcode,
	})
	require.Error(t, err)
	assert.True(t, IsErrTOTPPasscodeUsed(err), "expected ErrTOTPPasscodeUsed, got: %v", err)
}

func TestTOTPPasscodeWhitespaceVariantsCannotBeReused(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	secret := "JBSWY3DPEHPK3PXP" //nolint:gosec
	// Previous window: still valid, but not consumed by TestTOTPPasscodeCannotBeReused.
	passcode, err := totp.GenerateCode(secret, time.Now().Add(-30*time.Second))
	require.NoError(t, err)

	user := &User{ID: 10}

	_, err = ValidateTOTPPasscode(s, &TOTPPasscode{User: user, Passcode: " " + passcode + "\t"})
	require.NoError(t, err)

	for _, variant := range []string{passcode, " " + passcode, passcode + " ", "\n" + passcode + "\n"} {
		_, err = ValidateTOTPPasscode(s, &TOTPPasscode{User: user, Passcode: variant})
		require.Error(t, err, "variant %q", variant)
		assert.True(t, IsErrTOTPPasscodeUsed(err), "variant %q: expected ErrTOTPPasscodeUsed, got: %v", variant, err)
	}

	for _, invalid := range []string{passcode[:3] + " " + passcode[3:], "+" + passcode[1:], passcode + "0", ""} {
		_, err = ValidateTOTPPasscode(s, &TOTPPasscode{User: user, Passcode: invalid})
		require.Error(t, err, "invalid %q", invalid)
		assert.True(t, IsErrInvalidTOTPPasscode(err), "invalid %q: expected ErrInvalidTOTPPasscode, got: %v", invalid, err)
	}
}

func TestHandleFailedTOTPAuthLocksAccount(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	user := &User{ID: 10}

	// keyvalue store is process-global; reset counter between runs.
	_ = keyvalue.Del(user.GetFailedTOTPAttemptsKey())
	gen := sessionAuthGeneration(t, user.ID)

	for i := 0; i < 10; i++ {
		s := db.NewSession()
		HandleFailedTOTPAuth(user)
		require.NoError(t, s.Rollback())
		require.NoError(t, s.Close())
	}

	s := db.NewSession()
	defer s.Close()
	fresh, err := GetUserByID(s, user.ID)
	require.Error(t, err)
	assert.True(t, IsErrAccountLocked(err), "expected ErrAccountLocked, got: %v", err)
	assert.Equal(t, StatusAccountLocked, fresh.Status)
	assert.Equal(t, gen+2, sessionAuthGeneration(t, user.ID), "must invalidate before and after commit")
}

func TestHandleFailedTOTPAuthLockoutCanBeUnlockedByPasswordReset(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	user := &User{ID: 10}
	// keyvalue store is process-global; reset counter between runs.
	_ = keyvalue.Del(user.GetFailedTOTPAttemptsKey())

	for i := 0; i < 10; i++ {
		s := db.NewSession()
		HandleFailedTOTPAuth(user)
		require.NoError(t, s.Rollback())
		require.NoError(t, s.Close())
	}

	// Fetch the token from the DB (not a fresh one) so regressions where
	// the lockout flow fails to persist a reset token are caught.
	s := db.NewSession()
	defer s.Close()
	full, err := GetUserByID(s, user.ID)
	require.Error(t, err)
	require.True(t, IsErrAccountLocked(err))
	require.Equal(t, StatusAccountLocked, full.Status)

	tokens, err := getTokensForKind(s, full, TokenPasswordReset)
	require.NoError(t, err)
	require.NotEmpty(t, tokens, "HandleFailedTOTPAuth should have persisted a password reset token")
	token := tokens[len(tokens)-1]
	// Tokens are stored hashed and the cleartext is never persisted, so overwrite
	// the stored hash with a known value's hash to exercise the reset end-to-end.
	const rawToken = "totp-lockout-reset-token"
	token.Token = utils.Sha256Hex(rawToken)
	_, err = s.Where("id = ?", token.ID).Cols("token").Update(token)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	s2 := db.NewSession()
	defer s2.Close()
	_, err = ResetPassword(s2, &PasswordReset{
		Token:       rawToken,
		NewPassword: "new-password-123",
	})
	require.NoError(t, err)
	require.NoError(t, s2.Commit())

	s3 := db.NewSession()
	defer s3.Close()
	unlocked, err := GetUserByID(s3, user.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusActive, unlocked.Status)
}

func TestEnrollTOTPRequiresPassword(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u, err := GetUserByID(s, 2)
	require.NoError(t, err)
	_ = keyvalue.Del(reauthAttemptsKey(u.ID))

	_, err = EnrollTOTP(context.Background(), s, u, "wrong")
	assert.True(t, IsErrWrongUsernameOrPassword(err), "got %v", err)
	assert.Equal(t, int64(1), kvCount(t, reauthAttemptsKey(u.ID)))
	db.AssertMissing(t, "totp", map[string]any{"user_id": u.ID})

	enrolled, err := EnrollTOTP(context.Background(), s, u, "12345678")
	require.NoError(t, err)
	assert.False(t, enrolled.Enabled)
}

func TestEnableTOTPRequiresPassword(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u, err := GetUserByID(s, 1)
	require.NoError(t, err)
	_ = keyvalue.Del(reauthAttemptsKey(u.ID))

	passcode, err := totp.GenerateCode("HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ", time.Now())
	require.NoError(t, err)

	err = EnableTOTP(context.Background(), s, u, "wrong", passcode)
	assert.True(t, IsErrWrongUsernameOrPassword(err), "got %v", err)

	err = EnableTOTP(context.Background(), s, u, "12345678", passcode)
	require.NoError(t, err)
	enabled, err := TOTPEnabledForUser(s, u)
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestTOTPPasscodeConcurrentReplay(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	const attempts = 128

	for range 10 {
		passcode, err := totp.GenerateCode("JBSWY3DPEHPK3PXP", time.Now())
		require.NoError(t, err)
		key := fmt.Sprintf("totp_used_10_%s", passcode)
		require.NoError(t, keyvalue.Del(key))
		t.Cleanup(func() { _ = keyvalue.Del(key) })

		start := make(chan struct{})
		results := make(chan error, attempts)
		var ready sync.WaitGroup
		ready.Add(attempts)
		for range attempts {
			go func() {
				s := db.NewSession()
				defer s.Close()
				ready.Done()
				<-start
				_, err := ValidateTOTPPasscode(s, &TOTPPasscode{
					User:     &User{ID: 10},
					Passcode: passcode,
				})
				results <- err
			}()
		}
		ready.Wait()
		close(start)

		accepted := 0
		for range attempts {
			err := <-results
			if err == nil {
				accepted++
			} else {
				require.True(t, IsErrTOTPPasscodeUsed(err), "unexpected error: %v", err)
			}
		}
		require.Equal(t, 1, accepted, "a passcode must be accepted once across concurrent logins")
	}
}
