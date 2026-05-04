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
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

// humaTokenFor issues a JWT for a test user via the real auth flow — used
// by the v2-only supplementary tests that drive the full Echo+Huma stack
// directly (bypassing the webHandlerTestV2 harness).
func humaTokenFor(t *testing.T, u *user.User) string {
	t.Helper()
	tok, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	return tok
}

// humaRequest is a one-shot dispatch helper that reuses an already-bootstrapped
// echo.Echo. Used by the v2-only supplementary tests to avoid re-loading
// fixtures between chained calls (create → patch → get).
func humaRequest(t *testing.T, e *echo.Echo, method, path, body, token, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}
