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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func sessionAuthKey(t *testing.T, userID int64) string {
	t.Helper()
	key, err := SessionAuthCacheKey(userID, "some-session")
	require.NoError(t, err)
	return key
}

func sessionAuthGeneration(t *testing.T, userID int64) int64 {
	t.Helper()
	gen, _, err := keyvalue.GetInt64(sessionAuthGenerationKey(userID))
	require.NoError(t, err)
	return gen
}

func TestInvalidateSessionAuthCache(t *testing.T) {
	before := sessionAuthKey(t, 2)
	other := sessionAuthKey(t, 1)
	assert.Equal(t, before, sessionAuthKey(t, 2), "the key must be stable between invalidations")

	require.NoError(t, InvalidateSessionAuthCache(2))

	assert.NotEqual(t, before, sessionAuthKey(t, 2))
	assert.Equal(t, other, sessionAuthKey(t, 1), "other users' cached sessions must stay")
}

// GHSA-4qvp-948j-4cgv: a request racing the transaction may cache the pre-commit state.
func TestStatusChangesInvalidateSessionAuthCache(t *testing.T) {
	for name, setStatus := range map[string]func(s *xorm.Session, u *User) error{
		"SetUserStatus": func(s *xorm.Session, u *User) error { return SetUserStatus(s, u, StatusDisabled) },
		"SetStatus":     func(s *xorm.Session, u *User) error { return u.SetStatus(s, StatusDisabled) },
	} {
		t.Run(name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			before := sessionAuthGeneration(t, 2)
			require.NoError(t, setStatus(s, &User{ID: 2}))
			assert.Equal(t, before+1, sessionAuthGeneration(t, 2), "must invalidate before commit")

			require.NoError(t, s.Commit())
			events.DispatchPending(context.Background(), s)

			assert.Equal(t, before+2, sessionAuthGeneration(t, 2), "must invalidate again once DispatchPending returns")
		})
		t.Run(name+" rollback", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			before := sessionAuthGeneration(t, 2)
			require.NoError(t, setStatus(s, &User{ID: 2}))
			require.NoError(t, s.Rollback())
			events.CleanupPending(s)
			events.DispatchPending(context.Background(), s)

			assert.Equal(t, before+1, sessionAuthGeneration(t, 2))
		})
	}
}
