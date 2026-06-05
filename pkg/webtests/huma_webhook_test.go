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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaWebhook ports the v1 webhook coverage (TestWebhook) to /api/v2 and
// extends it to the full permission matrix the v1 model enforces but its thin
// webtest never exercised. Project webhooks are nested under
// /projects/{project}/webhooks{,/{webhook}} — list, create, update, delete; there
// is deliberately no ReadOne (webhooks carry secrets).
//
// Permission gradient — Webhook.CanRead delegates to Project.CanRead (any share
// level), while Can{Create,Update,Delete} delegate to Project.CanUpdate →
// Project.CanWrite. The same user walks every rung by switching the parent path:
//   - project 1  (owned by testuser1): can do everything; holds fixture webhook #1
//   - project 9  (read share):  CAN list, CANNOT create/update/delete (webhook #2)
//   - project 10 (write share): CAN list/create/update/delete (webhook #3)
//   - project 11 (admin share): CAN list/create/update/delete (webhook #4)
//   - project 2  (no access, owned by user3): forbidden on everything
func TestHumaWebhook(t *testing.T) {
	// availableWebhookEvents is populated by RegisterListeners(), which the
	// webtests harness does not call. Register the one event the fixtures and
	// these cases use so Create/Update validation accepts it.
	models.RegisterEventForWebhook(&models.TaskUpdatedEvent{})

	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/webhooks",
		idParam:  "webhook",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// All harnesses share owned's Echo: each setupTestEnv() regenerates the JWT
	// signing secret, so independent instances would invalidate each other's tokens.
	on := func(projectID string) *webHandlerTestV2 {
		return &webHandlerTestV2{
			user:     &testuser1,
			basePath: "/api/v2/projects/" + projectID + "/webhooks",
			idParam:  "webhook",
			t:        t,
			e:        owned.e,
		}
	}
	readShared := on("9")
	writeShared := on("10")
	adminShared := on("11")
	forbidden := on("2")

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// project 1 has exactly fixture webhook #1.
			ids := webhookIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{1}, ids, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"target_url"`)
		})
		t.Run("Secret and basic auth credentials are never exposed", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `webhook-user`)
			assert.NotContains(t, rec.Body.String(), `webhook-password`)
			assert.NotContains(t, rec.Body.String(), `webhook-secret-fixture`)
		})
		t.Run("Read-only share can list", func(t *testing.T) {
			rec, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := webhookIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{2}, ids, "body: %s", rec.Body.String())
		})
		t.Run("Write share can list", func(t *testing.T) {
			rec, err := writeShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := webhookIDsFromReadAll(t, rec.Body.Bytes())
			assert.ElementsMatch(t, []int64{3}, ids, "body: %s", rec.Body.String())
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"target_url":"https://example.com/new","events":["task.updated"]}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"target_url":"https://example.com/new"`)
			// parent project comes from the URL.
			assert.Contains(t, rec.Body.String(), `"project_id":1`)
		})
		t.Run("Secret and basic auth are not echoed back", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil,
				`{"target_url":"https://example.com/secret","events":["task.updated"],"secret":"top-secret","basic_auth_user":"u","basic_auth_password":"p"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotContains(t, rec.Body.String(), `top-secret`)
			assert.NotContains(t, rec.Body.String(), `"basic_auth_user":"u"`)
			assert.NotContains(t, rec.Body.String(), `"basic_auth_password":"p"`)
		})
		t.Run("Admin share can create", func(t *testing.T) {
			rec, err := adminShared.testCreateWithUser(nil, nil, `{"target_url":"https://example.com/admin","events":["task.updated"]}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"project_id":11`)
		})
		t.Run("Write share can create", func(t *testing.T) {
			rec, err := writeShared.testCreateWithUser(nil, nil, `{"target_url":"https://example.com/write","events":["task.updated"]}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"project_id":10`)
		})
		t.Run("Read share cannot create", func(t *testing.T) {
			_, err := readShared.testCreateWithUser(nil, nil, `{"target_url":"https://example.com/nope","events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"target_url":"https://example.com/nope","events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Invalid event", func(t *testing.T) {
			// An unregistered event name → InvalidFieldError, which v1 surfaces as
			// 412 Precondition Failed (ValidationHTTPError.HTTPCode); v2 mirrors it.
			_, err := owned.testCreateWithUser(nil, nil, `{"target_url":"https://example.com/x","events":["not.a.real.event"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
		})
		t.Run("Missing target url", func(t *testing.T) {
			// Create rejects a non-http target_url via InvalidFieldError → 412.
			_, err := owned.testCreateWithUser(nil, nil, `{"events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal - only events change", func(t *testing.T) {
			// Update persists only the events list (model writes Cols("events")).
			// Send a different target_url and confirm the stored value is untouched.
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"webhook": "1"},
				`{"events":["task.updated"],"target_url":"https://example.com/ignored"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"id":1`)

			rec, err = owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `https://example.com/webhook-fixture`,
				"target_url must stay the fixture value; only events are mutable")
			assert.NotContains(t, rec.Body.String(), `https://example.com/ignored`)
		})
		t.Run("Write share can update", func(t *testing.T) {
			rec, err := writeShared.testUpdateWithUser(nil, map[string]string{"webhook": "3"}, `{"events":["task.updated"]}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
		t.Run("Admin share can update", func(t *testing.T) {
			rec, err := adminShared.testUpdateWithUser(nil, map[string]string{"webhook": "4"}, `{"events":["task.updated"]}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
		t.Run("Read share cannot update", func(t *testing.T) {
			// webhook #2 lives in project 9 (read share); CanUpdate needs write.
			_, err := readShared.testUpdateWithUser(nil, map[string]string{"webhook": "2"}, `{"events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			// webhook #5 lives in project 2, which user1 cannot access at all.
			// canDoWebhook resolves the parent from the webhook row, so the URL
			// project is irrelevant — the real project (2) gates the check.
			_, err := forbidden.testUpdateWithUser(nil, map[string]string{"webhook": "5"}, `{"events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// canDoWebhook returns false for a missing webhook → 403, not 404.
			_, err := owned.testUpdateWithUser(nil, map[string]string{"webhook": "9999"}, `{"events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Read share cannot delete", func(t *testing.T) {
			_, err := readShared.testDeleteWithUser(nil, map[string]string{"webhook": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden", func(t *testing.T) {
			// webhook #5 lives in project 2, which user1 cannot access at all.
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"webhook": "5"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share can delete", func(t *testing.T) {
			rec, err := writeShared.testDeleteWithUser(nil, map[string]string{"webhook": "3"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Admin share can delete", func(t *testing.T) {
			rec, err := adminShared.testDeleteWithUser(nil, map[string]string{"webhook": "4"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"webhook": "1"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// canDoWebhook returns false for a missing webhook → 403.
			_, err := owned.testDeleteWithUser(nil, map[string]string{"webhook": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})
}

// TestHumaWebhook_DisabledByConfig confirms RegisterWebhookRoutes skips the
// resource entirely when webhooks.enabled is false, so the v2 routes 404 rather
// than running with the feature toggled off.
func TestHumaWebhook_DisabledByConfig(t *testing.T) {
	// setupTestEnv loads fixtures and resets config to defaults (webhooks on).
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.WebhooksEnabled.Set(false)
	defer config.WebhooksEnabled.Set(true)

	// Rebuild the router so RegisterWebhookRoutes re-reads the now-disabled flag.
	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	token := humaTokenFor(t, &testuser1)
	rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/webhooks", "", token, "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "route must be absent when webhooks are disabled; body: %s", rec.Body.String())
}

// webhookIDsFromReadAll pulls the webhook IDs out of a v2 paginated list body so
// the visible set can be asserted exactly rather than via substring matching.
func webhookIDsFromReadAll(t *testing.T, body []byte) []int64 {
	t.Helper()
	var resp struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids := make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
}
