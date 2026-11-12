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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Login CSRF (GHSA-frf4-3prf-3vch): a cross-site <form> must not mint a session.
func TestFormURLEncodedBodiesRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	login := url.Values{"username": {"user1"}, "password": {"12345678"}}.Encode()

	for _, path := range []string{"/api/v1/login", "/api/v2/login"} {
		t.Run(path, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodPost, path, login, "", "application/x-www-form-urlencoded")
			assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code, rec.Body.String())
			assert.NotContains(t, rec.Body.String(), "token")
			assert.Empty(t, rec.Result().Cookies())
		})
	}

	t.Run("content type parameters and case", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/login", login, "", "Application/X-WWW-Form-Urlencoded; charset=UTF-8")
		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code, rec.Body.String())
	})

	t.Run("json login still works", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/login", `{"username":"user1","password":"12345678"}`, "", "application/json")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})

	// RFC 6749 requires form bodies on the token endpoint.
	for _, path := range []string{"/api/v1/oauth/token", "/api/v2/oauth/token"} {
		t.Run(path+" accepts forms", func(t *testing.T) {
			form := url.Values{"grant_type": {"password"}, "client_id": {"vikunja"}}.Encode()
			rec := humaRequest(t, e, http.MethodPost, path, form, "", "application/x-www-form-urlencoded")
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		})
	}
}
