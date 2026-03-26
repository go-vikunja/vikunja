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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

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
