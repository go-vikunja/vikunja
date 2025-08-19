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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectV2GetAll(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("Normal", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
		assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
		assert.Contains(t, rec.Body.String(), `"title":"Test3"`)  // Shared directly via users_project
		assert.Contains(t, rec.Body.String(), `"title":"Test12"`) // Shared via parent project
		assert.NotContains(t, rec.Body.String(), `"title":"Test5"`)
		assert.NotContains(t, rec.Body.String(), `"title":"Test22"`) // Archived directly
		assert.Contains(t, rec.Body.String(), `"_links":{`)
	})

	t.Run("Search", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects?s=Test1", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `Test1`)
		assert.NotContains(t, rec.Body.String(), `Test2`)
		assert.NotContains(t, rec.Body.String(), `Test3`)
		assert.NotContains(t, rec.Body.String(), `Test4`)
		assert.NotContains(t, rec.Body.String(), `Test5`)
	})

	t.Run("Normal with archived projects", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects?is_archived=true", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `Test1`)
		assert.NotContains(t, rec.Body.String(), `Test2"`)
		assert.Contains(t, rec.Body.String(), `Test3`)  // Shared directly via users_project
		assert.Contains(t, rec.Body.String(), `Test12`) // Shared via parent project
		assert.NotContains(t, rec.Body.String(), `Test5`)
		assert.Contains(t, rec.Body.String(), `Test21`) // Archived through project
		assert.Contains(t, rec.Body.String(), `Test22`) // Archived directly
	})

	t.Run("Pagination", func(t *testing.T) {
		// There are more than 3 projects for user1
		rec, err := th.Request(t, "GET", "/api/v2/projects?page=1&per_page=3", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("x-pagination-total-pages"))
		assert.Equal(t, "3", rec.Header().Get("x-pagination-result-count"))
	})
}

func TestProjectV2Create(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("Normal", func(t *testing.T) {
		rec, err := th.Request(t, "POST", "/api/v2/projects", strings.NewReader(`{"title":"new project"}`))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"title":"new project"`)
		assert.Contains(t, rec.Body.String(), `"_links":{`)
	})

	t.Run("Empty title", func(t *testing.T) {
		rec, err := th.Request(t, "POST", "/api/v2/projects", strings.NewReader(`{"title":""}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestProjectV2Delete(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)

		rec, err := th.Request(t, "DELETE", "/api/v2/projects/1", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify it's gone
		rec, err = th.Request(t, "GET", "/api/v2/projects/1", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Nonexisting", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)
		rec, err := th.Request(t, "DELETE", "/api/v2/projects/9999", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)
		// User 1 has no permissions on project 2
		rec, err := th.Request(t, "DELETE", "/api/v2/projects/2", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestProjectV2Update(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)

		payload := `{"title":"Updated Title", "description":"Updated Description"}`
		rec, err := th.Request(t, "PUT", "/api/v2/projects/1", strings.NewReader(payload))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"title":"Updated Title"`)
		assert.Contains(t, rec.Body.String(), `"description":"Updated Description"`)
		assert.Contains(t, rec.Body.String(), `"_links":{`)
	})

	t.Run("Nonexisting", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)
		payload := `{"title":"Updated Title"}`
		rec, err := th.Request(t, "PUT", "/api/v2/projects/9999", strings.NewReader(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)
		payload := `{"title":"Updated Title"}`
		rec, err := th.Request(t, "PUT", "/api/v2/projects/2", strings.NewReader(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("EmptyTitle", func(t *testing.T) {
		th := NewTestHelper(t)
		th.Login(t, &testuser1)
		payload := `{"title":""}`
		rec, err := th.Request(t, "PUT", "/api/v2/projects/1", strings.NewReader(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestProjectV2Get(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("Normal", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/1", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
		assert.Contains(t, rec.Body.String(), `"_links":{`)
	})
	t.Run("Nonexisting", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/9999", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
	t.Run("Forbidden", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/2", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
