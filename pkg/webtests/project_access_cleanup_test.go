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
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-hjx8-qv73-f7cm
func TestProjectUserRemovalDeletesWebhooks(t *testing.T) {
	for name, tc := range map[string]struct {
		path        string
		wantStatus  int
		webhookKept bool
	}{
		"v1":                    {path: "/api/v1/projects/10/users/user1", wantStatus: http.StatusOK},
		"v2 deleting webhooks":  {path: "/api/v2/projects/10/users/user1?clear_outside_webhooks=true", wantStatus: http.StatusNoContent},
		"v2 keeping webhooks":   {path: "/api/v2/projects/10/users/user1?clear_outside_webhooks=false", wantStatus: http.StatusNoContent, webhookKept: true},
		"v2 without a decision": {path: "/api/v2/projects/10/users/user1", wantStatus: http.StatusConflict, webhookKept: true},
	} {
		t.Run(name, func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			// user1 has write access to project 10, owned by user6.
			w := insertWebhook(t, 10, 1)

			rec := apiTokenReq(e, http.MethodDelete, tc.path, userJWT(t, 6), "")
			assert.Equal(t, tc.wantStatus, rec.Code, rec.Body.String())
			if tc.webhookKept {
				db.AssertExists(t, "webhooks", map[string]interface{}{"id": w.ID}, false)
			} else {
				db.AssertMissing(t, "webhooks", map[string]interface{}{"id": w.ID})
			}
			db.AssertExists(t, "webhooks", map[string]interface{}{"id": 3}, false)
		})
	}
}

func TestProjectUserRemovalConflictKeepsShare(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	insertWebhook(t, 10, 1)

	rec := apiTokenReq(e, http.MethodDelete, "/api/v2/projects/10/users/user1", userJWT(t, 6), "")
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	assertErrorCode(t, rec, models.ErrCodeOutsideWebhooksNeedConfirmation)
	db.AssertExists(t, "users_projects", map[string]interface{}{"project_id": 10, "user_id": 1}, false)
}

func TestProjectMoveAsksAboutOutsideWebhooks(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	owner := userJWT(t, 1)

	create := func(body string) int64 {
		rec := apiTokenReq(e, http.MethodPost, "/api/v2/projects", owner, body)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var p struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))
		return p.ID
	}
	parent := create(`{"title":"outside webhooks parent"}`)
	child := create(fmt.Sprintf(`{"title":"outside webhooks child","parent_project_id":%d}`, parent))
	target := create(`{"title":"outside webhooks target"}`)

	rec := apiTokenReq(e, http.MethodPost, fmt.Sprintf("/api/v2/projects/%d/users", parent), owner, `{"username":"user2","permission":1}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	w := insertWebhook(t, child, 2)

	patch := func(body string) *httptest.ResponseRecorder {
		return humaRequest(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/projects/%d", child), body, owner, "application/merge-patch+json")
	}

	rec = patch(fmt.Sprintf(`{"parent_project_id":%d}`, target))
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	assertErrorCode(t, rec, models.ErrCodeOutsideWebhooksNeedConfirmation)
	db.AssertExists(t, "projects", map[string]interface{}{"id": child, "parent_project_id": parent}, false)
	db.AssertExists(t, "webhooks", map[string]interface{}{"id": w.ID}, false)

	rec = patch(fmt.Sprintf(`{"parent_project_id":%d,"clear_outside_webhooks":false}`, target))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "clear_outside_webhooks")
	db.AssertExists(t, "projects", map[string]interface{}{"id": child, "parent_project_id": target}, false)
	db.AssertExists(t, "webhooks", map[string]interface{}{"id": w.ID}, false)
}

func insertWebhook(t *testing.T, projectID, createdByID int64) *models.Webhook {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	w := &models.Webhook{
		TargetURL:   "https://example.com/removed-user",
		Events:      []string{"task.updated"},
		ProjectID:   projectID,
		CreatedByID: createdByID,
	}
	_, err := s.Insert(w)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return w
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code int) {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, code, body.Code)
}
