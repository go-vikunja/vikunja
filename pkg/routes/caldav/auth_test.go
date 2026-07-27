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

package caldav

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
	caldavTokenUser15Valid = "tk_caldav_api_token_test_00000000aabbccdd" // owner_id 15, caldav scope
	nonCaldavTokenUser15   = "tk_nocaldav_token_test_000000005678efab"   // owner_id 15, no caldav scope
)

func newAuthContext() *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dav/principals/user15/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestBasicAuthAPIToken(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("rejects token shorter than prefix+8 without panicking", func(t *testing.T) {
		// Real tokens are far longer; anything shorter is invalid by
		// construction and would otherwise panic inside GetTokenFromTokenString.
		for _, short := range []string{"tk_", "tk_a", "tk_abc", "tk_1234567"} {
			c := newAuthContext()
			ok, err := BasicAuth(c, "user15", short)
			require.NoError(t, err)
			assert.False(t, ok, "short token %q must be rejected", short)
			assert.Nil(t, c.Get("userBasicAuth"))
		}
	})

	t.Run("rejects unknown token", func(t *testing.T) {
		c := newAuthContext()
		ok, err := BasicAuth(c, "user15", "tk_nonexistent_token_value_aaaaaaaaaaaaaaaa")
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("rejects token without caldav scope", func(t *testing.T) {
		c := newAuthContext()
		ok, err := BasicAuth(c, "user15", nonCaldavTokenUser15)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("rejects token whose owner username does not match", func(t *testing.T) {
		c := newAuthContext()
		ok, err := BasicAuth(c, "wrongname", caldavTokenUser15Valid)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("accepts valid token with caldav scope", func(t *testing.T) {
		c := newAuthContext()
		ok, err := BasicAuth(c, "user15", caldavTokenUser15Valid)
		require.NoError(t, err)
		assert.True(t, ok)
		u, is := c.Get("userBasicAuth").(*user.User)
		require.True(t, is)
		assert.Equal(t, int64(15), u.ID)
	})
}
