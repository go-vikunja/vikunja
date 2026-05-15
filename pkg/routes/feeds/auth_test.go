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

package feeds

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	feedsTokenUser13Valid  = "tk_feeds_access_token_user_0013_feed0013"    // owner_id 13, feeds scope
	caldavOnlyToken        = "tk_caldav_api_token_test_00000000aabbccdd"   // owner_id 15, caldav scope only
	expiredTasksToken      = "tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8" // expired
	tasksScopedTokenOwner1 = "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e" // owner_id 1, no feeds scope
)

func newContext() *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/feeds/notifications.atom", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestBasicAuth(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("rejects non-token password without touching db", func(t *testing.T) {
		c := newContext()
		ok, err := BasicAuth(c, "user1", "plaintextpassword")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Nil(t, c.Get("userBasicAuth"))
	})

	t.Run("rejects unknown token", func(t *testing.T) {
		c := newContext()
		ok, err := BasicAuth(c, "user1", "tk_nonexistent_token_value_aaaaaaaaaaaaaaaa")
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("rejects token whose owner username does not match", func(t *testing.T) {
		c := newContext()
		// feedsTokenUser13Valid belongs to user 13; supply a different username.
		ok, err := BasicAuth(c, "wrongname", feedsTokenUser13Valid)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("rejects token without feeds scope", func(t *testing.T) {
		c := newContext()
		ok, err := BasicAuth(c, "user1", tasksScopedTokenOwner1)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("rejects token with caldav scope but no feeds scope", func(t *testing.T) {
		c := newContext()
		ok, err := BasicAuth(c, "user15", caldavOnlyToken)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		c := newContext()
		ok, err := BasicAuth(c, "user1", expiredTasksToken)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("accepts valid token with feeds scope", func(t *testing.T) {
		c := newContext()
		ok, err := BasicAuth(c, "user13", feedsTokenUser13Valid)
		require.NoError(t, err)
		assert.True(t, ok)
		u, is := c.Get("userBasicAuth").(*user.User)
		require.True(t, is)
		assert.Equal(t, int64(13), u.ID)
	})
}
