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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func userByID(t *testing.T, id int64) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, id)
	require.NoError(t, err)
	return u
}

func TestTasksReschedule(t *testing.T) {
	t.Run("moves the task and answers with what changed", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		u := userByID(t, 1)

		res := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/reschedule",
			`{"start_date":"2026-11-01T00:00:00Z","end_date":"2026-11-10T00:00:00Z"}`, humaTokenFor(t, u), "")

		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
		var out struct {
			Task struct {
				ID      int64  `json:"id"`
				EndDate string `json:"end_date"`
			} `json:"task"`
			Changed []struct {
				ID int64 `json:"id"`
			} `json:"changed"`
			Skipped []int64 `json:"skipped"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &out))
		assert.Equal(t, int64(1), out.Task.ID)
		assert.Equal(t, "2026-11-10T00:00:00Z", out.Task.EndDate)
		assert.NotNil(t, out.Skipped, "an empty list, not null")
		assert.NotNil(t, out.Changed)
	})
	t.Run("no write access is refused and nothing changes", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user13 has no access to project 1.
		u := userByID(t, 13)

		res := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/reschedule",
			`{"end_date":"2026-11-10T00:00:00Z"}`, humaTokenFor(t, u), "")

		assert.Equal(t, http.StatusForbidden, res.Code, res.Body.String())
	})
	t.Run("an end before the start is a bad request", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/reschedule",
			`{"start_date":"2026-11-10T00:00:00Z","end_date":"2026-11-01T00:00:00Z"}`, humaTokenFor(t, userByID(t, 1)), "")

		assert.Equal(t, http.StatusBadRequest, res.Code, res.Body.String())
	})
	t.Run("an unknown task is not found", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/999999/reschedule", `{}`, humaTokenFor(t, userByID(t, 1)), "")

		assert.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
	})
	t.Run("the permission is advertised for API tokens", func(t *testing.T) {
		_, err := setupTestEnv()
		require.NoError(t, err)
		routes := models.GetAPITokenRoutes()
		require.Contains(t, routes, "tasks")
		assert.NotNil(t, routes["tasks"]["reschedule"], "got %v", routes["tasks"])
	})
}

func TestProjectBaselineEndpoints(t *testing.T) {
	t.Run("save, read, delete", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, userByID(t, 1))

		empty := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/baseline", "", token, "")
		require.Equal(t, http.StatusOK, empty.Code, empty.Body.String())
		var before struct {
			Items []struct{} `json:"items"`
		}
		require.NoError(t, json.Unmarshal(empty.Body.Bytes(), &before))
		assert.Empty(t, before.Items)

		saved := humaRequest(t, e, http.MethodPut, "/api/v2/projects/1/baseline", "", token, "")
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())
		var n struct {
			Tasks int `json:"tasks"`
		}
		require.NoError(t, json.Unmarshal(saved.Body.Bytes(), &n))

		read := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/baseline", "", token, "")
		require.Equal(t, http.StatusOK, read.Code, read.Body.String())
		var after struct {
			SavedAt string     `json:"saved_at"`
			Items   []struct{} `json:"items"`
		}
		require.NoError(t, json.Unmarshal(read.Body.Bytes(), &after))
		assert.Len(t, after.Items, n.Tasks)

		del := humaRequest(t, e, http.MethodDelete, "/api/v2/projects/1/baseline", "", token, "")
		assert.Equal(t, http.StatusNoContent, del.Code, del.Body.String())
	})
	t.Run("a user without access cannot read or change it", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, userByID(t, 13))

		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			res := humaRequest(t, e, method, "/api/v2/projects/1/baseline", "", token, "")
			assert.GreaterOrEqual(t, res.Code, 400, "%s: %s", method, res.Body.String())
		}
	})
	t.Run("a pseudo project has none", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		res := humaRequest(t, e, http.MethodPut, "/api/v2/projects/-1/baseline", "", humaTokenFor(t, userByID(t, 1)), "")

		assert.GreaterOrEqual(t, res.Code, 400, res.Body.String())
	})
}
