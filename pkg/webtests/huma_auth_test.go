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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaAuthPublic ports the v1 coverage of the public local-account flows
// (register, password reset, email confirm) to /api/v2. These endpoints opt out
// of the global auth, so requests carry no token.
func TestHumaAuthPublic(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	post := func(path, body string) *httptest.ResponseRecorder {
		return humaRequest(t, e, http.MethodPost, path, body, "", "")
	}

	t.Run("Register", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			rec := post("/api/v2/register", `{"username":"newhumauser","password":"12345678","email":"newhuma@example.com"}`)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"username":"newhumauser"`)
		})
		t.Run("already existing username", func(t *testing.T) {
			rec := post("/api/v2/register", `{"username":"user1","password":"12345678","email":"x@example.com"}`)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
		t.Run("empty username", func(t *testing.T) {
			rec := post("/api/v2/register", `{"username":"","password":"12345678","email":"x@example.com"}`)
			assert.GreaterOrEqual(t, rec.Code, http.StatusBadRequest)
		})
	})

	t.Run("Request password reset token", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			rec := post("/api/v2/user/password/token", `{"email":"user1@example.com"}`)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), "Token was sent.")
		})
		t.Run("no user with that email", func(t *testing.T) {
			rec := post("/api/v2/user/password/token", `{"email":"user1000@example.com"}`)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})

	t.Run("Reset password", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			rec := post("/api/v2/user/password/reset", `{"token":"passwordresettesttoken","new_password":"12345678"}`)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), "The password was updated successfully.")
		})
		t.Run("invalid token", func(t *testing.T) {
			rec := post("/api/v2/user/password/reset", `{"token":"invalidtoken","new_password":"12345678"}`)
			assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		})
	})

	t.Run("Confirm email", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			rec := post("/api/v2/user/confirm", `{"token":"tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael"}`)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), "The email was confirmed successfully.")
		})
		t.Run("invalid token", func(t *testing.T) {
			rec := post("/api/v2/user/confirm", `{"token":"invalidToken"}`)
			assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		})
	})
}

// TestHumaRegisterDisabled proves the registration endpoint 404s when
// registration is disabled, mirroring v1.
func TestHumaRegisterDisabled(t *testing.T) {
	config.ServiceEnableRegistration.Set(false)
	defer config.ServiceEnableRegistration.Set(true)

	e, err := setupTestEnv()
	require.NoError(t, err)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/register",
		`{"username":"nope","password":"12345678","email":"nope@example.com"}`, "", "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestHumaLinkShareAuth ports the v1 link-share auth coverage to /api/v2.
func TestHumaLinkShareAuth(t *testing.T) {
	config.ServiceEnableLinkSharing.Set(true)

	e, err := setupTestEnv()
	require.NoError(t, err)

	post := func(share, body string) *httptest.ResponseRecorder {
		return humaRequest(t, e, http.MethodPost, "/api/v2/shares/"+share+"/auth", body, "", "")
	}

	t.Run("without password", func(t *testing.T) {
		rec := post("test", ``)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"token":"`)
		assert.Contains(t, rec.Body.String(), `"project_id":1`)
	})
	t.Run("with password, correct", func(t *testing.T) {
		rec := post("testWithPassword", `{"password":"12345678"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"token":"`)
	})
	t.Run("with password, missing", func(t *testing.T) {
		rec := post("testWithPassword", ``)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Equal(t, models.ErrCodeLinkSharePasswordRequired, problemCode(t, rec))
	})
	t.Run("with password, wrong", func(t *testing.T) {
		rec := post("testWithPassword", `{"password":"wrong"}`)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Equal(t, models.ErrCodeLinkSharePasswordInvalid, problemCode(t, rec))
	})
}

// TestHumaTokenMeta ports the token-introspection and link-share renew
// endpoints to /api/v2.
func TestHumaTokenMeta(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	userToken := humaTokenFor(t, &testuser1)

	t.Run("token test (GET) returns ok", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/token/test", "", userToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"ok"`)
	})
	t.Run("token check (POST) returns 200, not 418", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/token/test", "", userToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"ok"`)
	})
	t.Run("token check unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/token/test", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("routes lists token routes", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/routes", "", userToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var routes map[string]map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &routes))
		assert.Contains(t, routes, "tasks")
	})

	t.Run("renew link-share token", func(t *testing.T) {
		share := &models.LinkSharing{
			ID:          1,
			Hash:        "test",
			ProjectID:   1,
			Permission:  models.PermissionRead,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		}
		shareToken, err := auth.NewLinkShareJWTAuthtoken(share)
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/token", "", shareToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"token":"`)
	})
	t.Run("renew rejects user token", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/token", "", userToken, "")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestHumaOAuth ports the OAuth 2.0 token and authorize flows to /api/v2 and
// exercises both the JSON and the spec-compliant form-urlencoded encodings of
// the token endpoint.
func TestHumaOAuth(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("authorize requires authentication", func(t *testing.T) {
		body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", "abc", "S256", "s")
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/authorize", string(body), "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("full code flow with PKCE (JSON token request)", func(t *testing.T) {
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		challenge := pkceChallenge(verifier)
		code := authorizeV2(t, e, challenge, "xyz")

		body, _ := json.Marshal(map[string]string{ //nolint:errchkjson
			"grant_type":    "authorization_code",
			"code":          code,
			"client_id":     "vikunja",
			"redirect_uri":  "vikunja-flutter://callback",
			"code_verifier": verifier,
		})
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/token", string(body), "", "application/json")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))

		var resp oauth2server.TokenResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "bearer", resp.TokenType)
		assert.NotEmpty(t, resp.RefreshToken)
	})

	t.Run("full code flow with PKCE (form-urlencoded token request)", func(t *testing.T) {
		verifier := "form-encoded-flow-verifier"
		challenge := pkceChallenge(verifier)
		code := authorizeV2(t, e, challenge, "")

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"client_id":     {"vikunja"},
			"redirect_uri":  {"vikunja-flutter://callback"},
			"code_verifier": {verifier},
		}
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/token", form.Encode(), "", "application/x-www-form-urlencoded")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var resp oauth2server.TokenResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
	})

	t.Run("invalid grant type", func(t *testing.T) {
		form := url.Values{"grant_type": {"password"}, "client_id": {"vikunja"}}
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/token", form.Encode(), "", "application/x-www-form-urlencoded")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// authorizeV2 runs the v2 authorize step for testuser1 and returns the code.
func authorizeV2(t *testing.T, e *echo.Echo, challenge, state string) string {
	t.Helper()
	token := humaTokenFor(t, &testuser1)
	body := authorizeRequestBody("code", "vikunja", "vikunja-flutter://callback", challenge, "S256", state)
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/oauth/authorize", string(body), token, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp oauth2server.AuthorizeResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.Code)
	return resp.Code
}

// problemCode pulls the Vikunja numeric error code out of an RFC 9457 body.
func problemCode(t *testing.T, rec *httptest.ResponseRecorder) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body.Code
}
