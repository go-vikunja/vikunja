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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaTaskBulkCreate covers the v2 bulk create contract: atomic creation into the project from the URL, guarded by one write check on it.
func TestHumaTaskBulkCreate(t *testing.T) {
	base := &webHandlerTestV2{user: &testuser1, t: t}
	require.NoError(t, base.ensureEnv())

	bulkPost := func(projectID string, u *user.User, payload string) (*models.BulkTaskCreation, error) {
		h := &webHandlerTestV2{user: u, basePath: "/api/v2/projects/" + projectID + "/tasks/bulk", t: t, e: base.e}
		rec, err := h.serve(http.MethodPost, h.basePath, payload)
		if err != nil {
			return nil, err
		}
		assert.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		result := &models.BulkTaskCreation{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), result))
		return result, nil
	}

	t.Run("Create multiple tasks", func(t *testing.T) {
		result, err := bulkPost("1", &testuser1, `{"tasks":[{"title":"bulk a"},{"title":"bulk b"},{"title":"bulk c"}]}`)
		require.NoError(t, err)
		require.Len(t, result.Tasks, 3)
		for _, task := range result.Tasks {
			assert.NotZero(t, task.ID)
			assert.NotZero(t, task.Index)
			assert.Equal(t, int64(1), task.ProjectID)
		}
		assert.Equal(t, "bulk a", result.Tasks[0].Title)
		assert.Equal(t, "bulk c", result.Tasks[2].Title)
	})

	t.Run("URL project wins over body project_id", func(t *testing.T) {
		result, err := bulkPost("1", &testuser1, `{"tasks":[{"title":"body project ignored","project_id":2}]}`)
		require.NoError(t, err)
		require.Len(t, result.Tasks, 1)
		assert.Equal(t, int64(1), result.Tasks[0].ProjectID)
	})

	t.Run("Empty batch", func(t *testing.T) {
		// minItems:"1" on Tasks is enforced by Huma before the handler runs.
		_, err := bulkPost("1", &testuser1, `{"tasks":[]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
	})

	t.Run("Empty title fails Huma validation", func(t *testing.T) {
		// minLength:"1" on Task.Title is enforced by Huma before the handler runs.
		_, err := bulkPost("1", &testuser1, `{"tasks":[{"title":"ok"},{"title":""}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
	})

	t.Run("Invalid nested field fails at the boundary", func(t *testing.T) {
		// hex_color is caught by Huma's schema, repeat_after by its `valid:` tag.
		for _, tc := range []struct {
			name     string
			payload  string
			location string
		}{
			{"hex_color", `{"tasks":[{"title":"ok"},{"title":"bad color","hex_color":"ff00ff11"}]}`, "body.tasks[1].hex_color"},
			{"repeat_after", `{"tasks":[{"title":"ok"},{"title":"bad repeat","repeat_after":-1}]}`, "body.tasks[1].repeat_after"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				h := &webHandlerTestV2{user: &testuser1, basePath: "/api/v2/projects/1/tasks/bulk", t: t, e: base.e}
				rec, err := h.serve(http.MethodPost, h.basePath, tc.payload)
				require.Error(t, err)
				require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())

				var body huma.ErrorModel
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
				var locations []string
				for _, detail := range body.Errors {
					locations = append(locations, detail.Location)
				}
				assert.Contains(t, locations, tc.location)
			})
		}
	})

	t.Run("Invalid task names its index", func(t *testing.T) {
		// The ten-year cap is a model-level rule, so it fails with the payload index instead of as a field error.
		_, err := bulkPost("1", &testuser1, `{"tasks":[{"title":"ok"},{"title":"bad repeat","repeat_after":999999999999}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		assertHandlerErrorCode(t, err, models.ErrCodeInvalidTaskInBulkCreation)
		assert.Contains(t, err.Error(), "index 1")
	})

	t.Run("Forbidden - no access to the project", func(t *testing.T) {
		// User 6 has no access to project 1.
		_, err := bulkPost("1", &testuser6, `{"tasks":[{"title":"nope"}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("Forbidden - read-only share", func(t *testing.T) {
		// User 2 has read-only access to project 3.
		_, err := bulkPost("3", &testuser2, `{"tasks":[{"title":"nope"}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("Nonexistent project", func(t *testing.T) {
		_, err := bulkPost("99999", &testuser1, `{"tasks":[{"title":"nope"}]}`)
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
	})

	t.Run("Link share with write access", func(t *testing.T) {
		// Link share 2 has write access to project 2.
		token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
			ID:          2,
			Hash:        "test2",
			ProjectID:   2,
			Permission:  models.PermissionWrite,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		})
		require.NoError(t, err)
		rec := humaRequest(t, base.e, http.MethodPost, "/api/v2/projects/2/tasks/bulk", `{"tasks":[{"title":"via link share"}]}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"via link share"`)
	})

	t.Run("Link share read-only is forbidden", func(t *testing.T) {
		// Link share 1 has read-only access to project 1.
		token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
			ID:          1,
			Hash:        "test",
			ProjectID:   1,
			Permission:  models.PermissionRead,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		})
		require.NoError(t, err)
		rec := humaRequest(t, base.e, http.MethodPost, "/api/v2/projects/1/tasks/bulk", `{"tasks":[{"title":"nope"}]}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})
}
