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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBulkTaskCreateV2 covers POST /tasks/bulk. Like TestBulkTaskV2 it drives the
// Echo+Huma stack directly, since webHandlerTestV2 only models base[/{id}] paths.
func TestBulkTaskCreateV2(t *testing.T) {
	t.Run("creates all tasks in the order they were passed", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/bulk",
			`{"tasks":[{"title":"bulk first","project_id":1},{"title":"bulk second","project_id":1}]}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		var got struct {
			Tasks []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
				Index int64  `json:"index"`
			} `json:"tasks"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got.Tasks, 2)
		assert.Equal(t, "bulk first", got.Tasks[0].Title)
		assert.Equal(t, "bulk second", got.Tasks[1].Title)
		assert.Less(t, got.Tasks[0].Index, got.Tasks[1].Index)

		db.AssertExists(t, "tasks", map[string]interface{}{"id": got.Tasks[0].ID, "title": "bulk first"}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": got.Tasks[1].ID, "title": "bulk second"}, false)
	})

	t.Run("forbidden when missing write on one involved project", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Project 1 is owned by testuser1, project 3 is a read-only share.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/bulk",
			`{"tasks":[{"title":"shouldnothappen","project_id":1},{"title":"shouldnothappen too","project_id":3}]}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

		db.AssertMissing(t, "tasks", map[string]interface{}{"title": "shouldnothappen"})
	})

	t.Run("creates nothing when one task fails", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/bulk",
			`{"tasks":[{"title":"created before the failing one","project_id":1},{"title":"into a bucket that does not exist","project_id":1,"bucket_id":9999}]}`, token, "")
		require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())

		db.AssertMissing(t, "tasks", map[string]interface{}{"title": "created before the failing one"})
	})

	t.Run("empty batch is rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/bulk", `{"tasks":[]}`, token, "")
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("more tasks than allowed are rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		tasks := make([]string, 0, 201)
		for range 201 {
			tasks = append(tasks, `{"title":"too many","project_id":1}`)
		}

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/bulk",
			`{"tasks":[`+strings.Join(tasks, ",")+`]}`, token, "")
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())

		db.AssertMissing(t, "tasks", map[string]interface{}{"title": "too many"})
	})
}
