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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Webhook{}
		},
		t: t,
	}
	t.Run("ReadAll", func(t *testing.T) {
		t.Run("should not expose BasicAuth credentials", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, map[string]string{"project": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"target_url"`)
			assert.NotContains(t, rec.Body.String(), `webhook-user`)
			assert.NotContains(t, rec.Body.String(), `webhook-password`)
			assert.NotContains(t, rec.Body.String(), `webhook-secret-fixture`)
		})
		t.Run("Read-only share cannot list", func(t *testing.T) {
			_, err := testHandler.testReadAllWithUser(nil, map[string]string{"project": "9"})
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
		})
		t.Run("Write share can list", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, map[string]string{"project": "10"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `webhook-write-share`)
		})
	})
	// Project routes must never reach user-level webhooks, which token scopes exclude.
	t.Run("Project route ignores user_id from body", func(t *testing.T) {
		models.RegisterEventForWebhook(&models.TaskUpdatedEvent{})
		rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "1"},
			`{"target_url":"https://example.com/body-user","events":["task.updated"],"user_id":1}`)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"project_id":1`)
		assert.Contains(t, rec.Body.String(), `"user_id":0`)
	})
	t.Run("Project 0 cannot create a user webhook", func(t *testing.T) {
		models.RegisterUserDirectedEventForWebhook(&models.TaskReminderFiredEvent{})
		_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "0"},
			`{"target_url":"https://example.com/pwned","events":["task.reminder.fired"],"user_id":1,"project_id":0}`)
		require.Error(t, err)
		db.AssertMissing(t, "webhooks", map[string]interface{}{"target_url": "https://example.com/pwned"})
	})
	t.Run("Project route cannot delete a user webhook", func(t *testing.T) {
		// Webhook 8 is user 1's own user-level webhook.
		_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "1", "webhook": "8"})
		require.Error(t, err)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": 8}, false)
	})
}
