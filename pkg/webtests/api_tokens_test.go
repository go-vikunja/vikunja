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

package webtests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokenRoutesIncludesCaldav(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/routes", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"caldav"`)
	assert.Contains(t, res.Body.String(), `"access"`)
}

func TestAPIToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			u, err := auth.GetAuthFromClaims(c)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			return c.JSON(http.StatusOK, u)
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e") // Token 1
		require.NoError(t, h(c))
		// check if the request handlers "see" the request as if it came directly from that user
		assert.Contains(t, res.Body.String(), `"username":"user1"`)
	})
	t.Run("invalid token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_loremipsumdolorsitamet")
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("expired token", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8") // Token 2
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("valid token, invalid scope", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e")
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("disabled user token rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_disabled_user_test_token_000000001234abcd") // Token 4 (disabled user 17)
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("locked user token rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		req.Header.Set(echo.HeaderAuthorization, "Bearer tk_locked_user_test_token_0000000012345678") // Token 5 (locked user 18)
		require.NoError(t, h(c))
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), `"code":11`)
	})
	t.Run("jwt", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := routes.SetupTokenMiddleware()(func(c *echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
		require.NoError(t, err)

		req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
		require.NoError(t, h(c))
	})
}

func apiTokenReq(e *echo.Echo, method, target, jwt, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	return res
}

func userJWT(t *testing.T, id int64) string {
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, id)
	require.NoError(t, err)
	jwt, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	return jwt
}

// GHSA-vvcv-vpph-h844: link_share id 2 (hash test2) collides with user id 2,
// who owns api_token id 3. A colliding link-share principal must not be able to
// list, create, or delete that user's API tokens.
func TestAPITokenLinkShareCollision(t *testing.T) {
	linkShareJWT := func(t *testing.T) string {
		jwt, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
			ID:          2,
			Hash:        "test2",
			ProjectID:   2,
			Permission:  models.PermissionWrite,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		})
		require.NoError(t, err)
		return jwt
	}

	const createBody = `{"title":"collision","permissions":{"tasks":["read_all"]},"expires_at":"2099-01-01T00:00:00Z"}`

	t.Run("link share GET is forbidden and leaks no metadata", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := apiTokenReq(e, http.MethodGet, "/api/v1/tokens", linkShareJWT(t), "")
		assert.Equal(t, http.StatusForbidden, res.Code)
		assert.NotContains(t, res.Body.String(), "test token 3")
	})

	t.Run("link share PUT is forbidden and creates no row", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := apiTokenReq(e, http.MethodPut, "/api/v1/tokens", linkShareJWT(t), createBody)
		assert.Equal(t, http.StatusForbidden, res.Code)

		s := db.NewSession()
		defer s.Close()
		count, err := s.Where("owner_id = ?", 2).Count(&models.APIToken{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "no token must be created for the colliding user")
	})

	t.Run("link share DELETE is forbidden and retains the row", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := apiTokenReq(e, http.MethodDelete, "/api/v1/tokens/3", linkShareJWT(t), "")
		assert.Equal(t, http.StatusForbidden, res.Code)

		s := db.NewSession()
		defer s.Close()
		exists, err := s.Where("id = ?", 3).Exist(&models.APIToken{})
		require.NoError(t, err)
		assert.True(t, exists, "the target token must be retained")
	})

	t.Run("regular user positive controls", func(t *testing.T) {
		t.Run("GET", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			res := apiTokenReq(e, http.MethodGet, "/api/v1/tokens", userJWT(t, 2), "")
			assert.Equal(t, http.StatusOK, res.Code)
			assert.Contains(t, res.Body.String(), "test token 3")
		})
		t.Run("PUT", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			res := apiTokenReq(e, http.MethodPut, "/api/v1/tokens", userJWT(t, 2), createBody)
			assert.Equal(t, http.StatusCreated, res.Code)
		})
		t.Run("DELETE", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			res := apiTokenReq(e, http.MethodDelete, "/api/v1/tokens/3", userJWT(t, 2), "")
			assert.Equal(t, http.StatusOK, res.Code)
		})
	})

	t.Run("bot owner positive controls", func(t *testing.T) {
		// user 21 owns bot user 23.
		t.Run("create for own bot", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			res := apiTokenReq(e, http.MethodPut, "/api/v1/tokens", userJWT(t, 21),
				`{"title":"bot","owner_id":23,"permissions":{"tasks":["read_all"]},"expires_at":"2099-01-01T00:00:00Z"}`)
			assert.Equal(t, http.StatusCreated, res.Code)

			s := db.NewSession()
			defer s.Close()
			exists, err := s.Where("owner_id = ?", 23).Exist(&models.APIToken{})
			require.NoError(t, err)
			assert.True(t, exists)
		})
		t.Run("delete own bot token", func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			// First create a bot-owned token, then delete it.
			createRes := apiTokenReq(e, http.MethodPut, "/api/v1/tokens", userJWT(t, 21),
				`{"title":"bot","owner_id":23,"permissions":{"tasks":["read_all"]},"expires_at":"2099-01-01T00:00:00Z"}`)
			require.Equal(t, http.StatusCreated, createRes.Code)

			botToken := &models.APIToken{}
			s := db.NewSession()
			_, err = s.Where("owner_id = ?", 23).Get(botToken)
			require.NoError(t, err)
			// Close before the request: an open session holds the SQLite table
			// lock and the delete handler's own session would deadlock.
			s.Close()

			delRes := apiTokenReq(e, http.MethodDelete, "/api/v1/tokens/"+strconv.FormatInt(botToken.ID, 10), userJWT(t, 21), "")
			assert.Equal(t, http.StatusOK, delRes.Code)
		})
	})
}
