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
	"testing"

	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskByProjectIndex(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	token, err := auth.NewUserJWTAuthtoken(&testuser1, "test-session-id")
	require.NoError(t, err)

	do := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	t.Run("by numeric project id", func(t *testing.T) {
		rec := do("/api/v1/projects/1/tasks/by-index/1")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"id":1`)
	})

	t.Run("by project identifier", func(t *testing.T) {
		// Project 1 has identifier "test1" in fixtures.
		rec := do("/api/v1/projects/test1/tasks/by-index/1")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"id":1`)
	})

	t.Run("unknown project identifier returns 404", func(t *testing.T) {
		rec := do("/api/v1/projects/does-not-exist/tasks/by-index/1")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("numeric-only value always treated as id", func(t *testing.T) {
		// Even if a project had the identifier "999999", a pure-digit value
		// is parsed as an id; the task lookup then fails.
		rec := do("/api/v1/projects/999999/tasks/by-index/1")
		assert.NotEqual(t, http.StatusOK, rec.Code)
	})
}
