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
	"fmt"
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/require"
)

func TestTaskRecurrenceV2(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)
	due := time.Now().UTC().Truncate(time.Second).Add(48 * time.Hour)
	body := fmt.Sprintf(`{"title":"Daily task","repeat_after":86400,"due_date":%q}`, due.Format(time.RFC3339))
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/tasks", body, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	task := &models.Task{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), task))
	path := fmt.Sprintf("/api/v2/tasks/%d/recurrence", task.ID)

	rec = humaRequest(t, e, http.MethodPost, path, `{}`, humaTokenFor(t, &testuser15), "")
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	rec = humaRequest(t, e, http.MethodPost, path, `{}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	completed := &models.TaskRecurrence{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), completed))
	require.NotEmpty(t, completed.UndoToken)
	require.True(t, due.Add(24*time.Hour).Equal(completed.Task.DueDate))

	body = fmt.Sprintf(`{"undo_token":%q}`, completed.UndoToken)
	rec = humaRequest(t, e, http.MethodPost, path, body, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	restored := &models.TaskRecurrence{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), restored))
	require.Empty(t, restored.UndoToken)
	require.True(t, due.Equal(restored.Task.DueDate))
	require.True(t, task.DoneAt.Equal(restored.Task.DoneAt))

	rec = humaRequest(t, e, http.MethodPost, path, body, token, "")
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
}
