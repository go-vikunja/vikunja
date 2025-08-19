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

func TestProjectV2TasksGetAll(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("Normal", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/1/tasks", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"title":"task #1"`)
		assert.Contains(t, rec.Body.String(), `"title":"task #2 done"`)
		assert.NotContains(t, rec.Body.String(), `"title":"task #13 basic other project"`)
	})

	t.Run("Pagination", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/1/tasks?page=1&per_page=3", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("x-pagination-total-pages"))
		assert.Equal(t, "3", rec.Header().Get("x-pagination-result-count"))
	})

	t.Run("Filtering", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/1/tasks?s="+url.QueryEscape("task #1"), nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"title":"task #1"`)
		assert.NotContains(t, rec.Body.String(), `"title":"task #2 done"`)
	})

	t.Run("Forbidden", func(t *testing.T) {
		rec, err := th.Request(t, "GET", "/api/v2/projects/2/tasks", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
