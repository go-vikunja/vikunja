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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenForUser returns a JWT bearer token for the given user, suitable for the
// "Authorization: Bearer ..." header on full ServeHTTP requests.
func tokenForUser(t *testing.T, u *user.User) string {
	token, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	return token
}

// TestHumaLabel_Create_ReadOne_via_OAS31Route exercises the Huma-served
// /labels endpoint with the same auth + fixtures the legacy tests use.
// This proves:
//  1. humaecho5 adapter dispatches correctly
//  2. JWT middleware still populates the echo.Context
//  3. auth.GetAuthFromContext fishes it back out
//  4. DoCreate / DoReadOne wire up through to the model
//  5. Response JSON shape matches what the frontend expects
func TestHumaLabel_Create_ReadOne_via_OAS31Route(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	token := tokenForUser(t, &testuser1)

	// 1) PUT /api/v1/oas3/labels — create a label via the Huma-mounted route
	createReq := httptest.NewRequest(http.MethodPut, "/api/v1/oas3/labels",
		strings.NewReader(`{"title":"spike","hex_color":"abcdef"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	require.Equalf(t, http.StatusOK, createRec.Code,
		"unexpected status %d; body=%q", createRec.Code, createRec.Body.String())

	var created map[string]any
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))
	assert.Equal(t, "spike", created["title"])

	rawID, ok := created["id"]
	require.Truef(t, ok, "response body has no id field: %q", createRec.Body.String())
	// JSON numbers decode to float64.
	idFloat, ok := rawID.(float64)
	require.True(t, ok, "id field is not a number")
	require.NotZero(t, int64(idFloat))
	id := strconv.FormatInt(int64(idFloat), 10)

	// 2) GET /api/v1/oas3/labels/{id} — read it back
	readReq := httptest.NewRequest(http.MethodGet, "/api/v1/oas3/labels/"+id, nil)
	readReq.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	readRec := httptest.NewRecorder()
	e.ServeHTTP(readRec, readReq)

	require.Equalf(t, http.StatusOK, readRec.Code,
		"unexpected status %d; body=%q", readRec.Code, readRec.Body.String())

	var fetched map[string]any
	require.NoError(t, json.Unmarshal(readRec.Body.Bytes(), &fetched))
	assert.Equal(t, "spike", fetched["title"])
	assert.InDelta(t, idFloat, fetched["id"], 0)
}

// TestHumaLabel_OpenAPISpecContainsLabelPaths proves the spec is served
// and includes the Label routes.
func TestHumaLabel_OpenAPISpecContainsLabelPaths(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oas3/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equalf(t, http.StatusOK, rec.Code,
		"unexpected status %d; body=%q", rec.Code, rec.Body.String())

	body := rec.Body.String()
	assert.Contains(t, body, `"openapi":"3.1`)
	assert.Contains(t, body, `/labels`)
	assert.Contains(t, body, `/labels/{id}`)
}

// TestHumaLabel_ForbiddenErrorShape ensures a 403 returns
// {"message": "Forbidden"} and NOT RFC 9457 problem+json.
func TestHumaLabel_ForbiddenErrorShape(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// user 1 creates a label via the Huma route...
	creatorToken := tokenForUser(t, &testuser1)
	createReq := httptest.NewRequest(http.MethodPut, "/api/v1/oas3/labels",
		strings.NewReader(`{"title":"forbidden-target"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, "Bearer "+creatorToken)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	require.Equalf(t, http.StatusOK, createRec.Code,
		"create failed with %d: %q", createRec.Code, createRec.Body.String())

	var created map[string]any
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))
	idFloat, ok := created["id"].(float64)
	require.True(t, ok)
	id := strconv.FormatInt(int64(idFloat), 10)

	// ...another user (user 10) tries to delete it.
	attackerToken := tokenForUser(t, &testuser10)
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/oas3/labels/"+id, nil)
	delReq.Header.Set(echo.HeaderAuthorization, "Bearer "+attackerToken)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	assert.Equalf(t, http.StatusForbidden, delRec.Code,
		"expected 403, got %d; body=%q", delRec.Code, delRec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(delRec.Body.Bytes(), &body))
	assert.Equal(t, "Forbidden", body["message"])

	// RFC 9457 problem+json would put these on the payload; Vikunja's legacy
	// shape must stay clean.
	_, hasType := body["type"]
	_, hasTitle := body["title"]
	assert.Falsef(t, hasType, "unexpected RFC 9457 field 'type' in body %q", delRec.Body.String())
	assert.Falsef(t, hasTitle, "unexpected RFC 9457 field 'title' in body %q", delRec.Body.String())
}
