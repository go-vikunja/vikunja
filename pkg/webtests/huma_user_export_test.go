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
	"bytes"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/files"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaUserExport covers the v2 data-export endpoints. Fixture topology
// (pkg/db/fixtures/users.yml + files.yml):
//   - user1:  local, password 12345678, export_file_id 1 (file row exists, no bytes).
//   - user14: non-local (OIDC), no password to confirm.
//   - user15: local, no export.
func TestHumaUserExport(t *testing.T) {
	t.Run("Request with the correct password", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/request",
			`{"password":"12345678"}`, humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "requested data export")
	})

	t.Run("Request with a wrong password is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/request",
			`{"password":"wrong-password"}`, humaTokenFor(t, &testuser1), "")
		require.NotEqual(t, http.StatusOK, rec.Code,
			"a wrong password must not start an export; body: %s", rec.Body.String())
	})

	t.Run("Request as a non-local user skips the password", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/request",
			`{}`, humaTokenFor(t, &testuser14), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Download streams the export bytes", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user1's export points at file 1; setupTestEnv resets storage, so write
		// real bytes for it (size matches the fixture's declared 100 bytes).
		content := bytes.Repeat([]byte("v"), 100)
		require.NoError(t, (&files.File{ID: 1, Size: uint64(len(content))}).Save(bytes.NewReader(content)))

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/download",
			`{"password":"12345678"}`, humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, content, rec.Body.Bytes(), "the streamed export bytes must match")
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "test")
	})

	t.Run("Download with a wrong password is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/download",
			`{"password":"wrong-password"}`, humaTokenFor(t, &testuser1), "")
		require.NotEqual(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Download without an export returns 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/download",
			`{"password":"12345678"}`, humaTokenFor(t, &testuser15), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Download with a missing physical file returns 404", func(t *testing.T) {
		// user1 has export_file_id 1, but setupTestEnv leaves its bytes unwritten.
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/export/download",
			`{"password":"12345678"}`, humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Status returns the export metadata", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/export", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"id":1`)
		assert.Contains(t, rec.Body.String(), `"expires"`)
	})

	t.Run("Status without an export returns null", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/export", "", humaTokenFor(t, &testuser15), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.JSONEq(t, "null", rec.Body.String())
	})

	t.Run("Unauthenticated request is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/export", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}
