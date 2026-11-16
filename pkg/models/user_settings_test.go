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

package models

import (
	"context"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

func TestChangeUserPassword(t *testing.T) {
	t.Run("verifies the password of the authenticated account", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, user.UpdateUserPassword(s, &user.User{ID: 2}, "user2password"))

		// The username points at another account, e.g. a stale token claim.
		err := ChangeUserPassword(context.Background(), s, &user.User{ID: 1, Username: "user2"}, "user2password", "newpassword")
		require.Error(t, err)
		assert.True(t, user.IsErrWrongUsernameOrPassword(err))
	})
	t.Run("changes the password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, ChangeUserPassword(context.Background(), s, &user.User{ID: 1}, "12345678", "newpassword"))
		require.NoError(t, s.Commit())

		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.NoError(t, user.CheckUserPassword(u, "newpassword"))
		db.AssertMissing(t, "sessions", map[string]any{"user_id": 1})
	})
}

func TestSetUserPasswordAndInvalidateSessions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	events.ClearDispatchedEvents()
	s := db.NewSession()
	defer s.Close()

	require.NoError(t, SetUserPasswordAndInvalidateSessions(s, &user.User{ID: 2}, "newpassword"))
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)

	u, err := user.GetUserByID(s, 2)
	require.NoError(t, err)
	require.NoError(t, user.CheckUserPassword(u, "newpassword"))
	db.AssertMissing(t, "sessions", map[string]any{"user_id": 2})
	assert.Equal(t, &SessionsRevokedEvent{UserID: 2}, singleDispatchedEvent[*SessionsRevokedEvent](t))
}

func TestResetInvalidFrontendSettings(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	// Batches of 2 spread the six users over several batches.
	defer func(batch int) { frontendSettingsResetBatch = batch }(frontendSettingsResetBatch)
	frontendSettingsResetBatch = 2
	s := db.NewSession()
	defer s.Close()

	values := map[int64]string{
		1: `{"color_schema":"dark"}`,
		2: `{"a":"` + strings.Repeat("a", MaxFrontendSettingsSize-8) + `"}`,
		3: `{"a":"` + strings.Repeat("a", MaxFrontendSettingsSize-7) + `"}`,
		4: `[1,2]`,
		5: `"dark"`,
	}
	// Postgres stores frontend_settings in a json column, which rejects invalid JSON.
	if db.Type() != schemas.POSTGRES {
		values[6] = `{"color_schema":`
	}
	for id, value := range values {
		_, err := s.Exec("UPDATE users SET frontend_settings = ? WHERE id = ?", value, id)
		require.NoError(t, err)
	}

	require.NoError(t, ResetInvalidFrontendSettings(s))
	require.NoError(t, s.Commit())

	stored, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"color_schema": "dark"}, stored.FrontendSettings)
	db.AssertMissing(t, "users", map[string]any{"id": 2, "frontend_settings": nil})
	for id := range values {
		if id > 2 {
			db.AssertExists(t, "users", map[string]any{"id": id, "frontend_settings": nil}, false)
		}
	}
}
