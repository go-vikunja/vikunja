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

// TestHumaUserWebhook ports the v1 user-webhook coverage (the per-user sibling of
// the project webhooks tested in TestHumaWebhook) to /api/v2. User webhooks live
// at /user/settings/webhooks{,/{webhook}} — list, events, create, update, delete;
// there is deliberately no ReadOne (webhooks carry credentials).
//
// Ownership gradient — a user webhook is owned by its UserID, and every Can* boils
// down to "are you that user". Fixtures: webhooks #6/#7 belong to user6, #8 to
// user1. The actor is user6 (not user1): the user-webhook e2e tests dispatch
// user-directed events only for users 1 and 2, so user6-owned fixtures never fire
// there. The point of these cases is that user6 sees and mutates only their own
// webhooks and is forbidden on user1's.
func TestHumaUserWebhook(t *testing.T) {
	// availableWebhookEvents / userDirectedWebhookEvents are populated by
	// RegisterListeners(), which the webtests harness does not call. Register the
	// one user-directed event the fixtures and these cases use so Create/Update
	// validation accepts it.
	models.RegisterUserDirectedEventForWebhook(&models.TaskReminderFiredEvent{})

	owner := webHandlerTestV2{
		user:     &testuser6,
		basePath: "/api/v2/user/settings/webhooks",
		idParam:  "webhook",
		t:        t,
	}
	require.NoError(t, owner.ensureEnv())

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal - sees only own webhooks", func(t *testing.T) {
			rec, err := owner.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			ids := webhookIDsFromReadAll(t, rec.Body.Bytes())
			// user6 owns #6 and #7; #8 belongs to user1 and must not appear.
			assert.ElementsMatch(t, []int64{6, 7}, ids, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"target_url"`)
		})
		t.Run("Secret and basic auth credentials are never exposed", func(t *testing.T) {
			rec, err := owner.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `uwh-secret-fixture`)
			assert.NotContains(t, rec.Body.String(), `uwh-basicauth-user`)
			assert.NotContains(t, rec.Body.String(), `uwh-basicauth-pass`)
		})
	})

	t.Run("Events", func(t *testing.T) {
		// The events route reports only user-directed events. task.reminder.fired
		// is registered above; task.updated (project-only) must not be listed.
		token := humaTokenFor(t, &testuser6)
		rec := humaRequest(t, owner.e, http.MethodGet, "/api/v2/user/settings/webhooks/events", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		var events []string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &events), "body: %s", rec.Body.String())
		assert.Contains(t, events, "task.reminder.fired")
		assert.NotContains(t, events, "task.updated")
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owner.testCreateWithUser(nil, nil,
				`{"target_url":"https://example.com/new","events":["task.reminder.fired"]}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"target_url":"https://example.com/new"`)
			// Ownership comes from the token, not the body.
			assert.Contains(t, rec.Body.String(), `"user_id":6`)
		})
		t.Run("Secret and basic auth are not echoed back", func(t *testing.T) {
			rec, err := owner.testCreateWithUser(nil, nil,
				`{"target_url":"https://example.com/secret","events":["task.reminder.fired"],"secret":"top-secret","basic_auth_user":"u","basic_auth_password":"p"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotContains(t, rec.Body.String(), `top-secret`)
			assert.NotContains(t, rec.Body.String(), `"basic_auth_user":"u"`)
			assert.NotContains(t, rec.Body.String(), `"basic_auth_password":"p"`)
		})
		t.Run("Non user-directed event rejected", func(t *testing.T) {
			// task.updated is a project event, not user-directed; Create rejects it
			// → InvalidFieldError, surfaced as 422 on v2.
			_, err := owner.testCreateWithUser(nil, nil,
				`{"target_url":"https://example.com/x","events":["task.updated"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Missing target url", func(t *testing.T) {
			_, err := owner.testCreateWithUser(nil, nil, `{"events":["task.reminder.fired"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal - only events change", func(t *testing.T) {
			rec, err := owner.testUpdateWithUser(nil, map[string]string{"webhook": "6"},
				`{"events":["task.reminder.fired"],"target_url":"https://example.com/ignored"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"id":6`)

			rec, err = owner.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `https://example.com/user-webhook-fixture`,
				"target_url must stay the fixture value; only events are mutable")
			assert.NotContains(t, rec.Body.String(), `https://example.com/ignored`)
		})
		t.Run("Cannot update another user's webhook", func(t *testing.T) {
			// webhook #8 belongs to user1; canDoWebhook resolves ownership from the
			// stored row, so user6 is forbidden regardless of the URL.
			_, err := owner.testUpdateWithUser(nil, map[string]string{"webhook": "8"},
				`{"target_url":"https://example.com/wh","events":["task.reminder.fired"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// canDoWebhook returns false for a missing webhook → 403, not 404.
			_, err := owner.testUpdateWithUser(nil, map[string]string{"webhook": "9999"},
				`{"target_url":"https://example.com/wh","events":["task.reminder.fired"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Cannot delete another user's webhook", func(t *testing.T) {
			_, err := owner.testDeleteWithUser(nil, map[string]string{"webhook": "8"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := owner.testDeleteWithUser(nil, map[string]string{"webhook": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Normal", func(t *testing.T) {
			rec, err := owner.testDeleteWithUser(nil, map[string]string{"webhook": "7"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
	})
}

// TestHumaUserWebhook_DisabledByConfig confirms RegisterUserWebhookRoutes skips
// the resource when webhooks.enabled is false, so the v2 user-webhook routes 404
// rather than running with the feature toggled off.
func TestHumaUserWebhook_DisabledByConfig(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.WebhooksEnabled.Set(false)
	defer config.WebhooksEnabled.Set(true)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	token := humaTokenFor(t, &testuser1)
	rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/webhooks", "", token, "")
	assert.Equal(t, http.StatusNotFound, rec.Code, "route must be absent when webhooks are disabled; body: %s", rec.Body.String())
}
