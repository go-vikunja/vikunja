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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaNotification mirrors v1's notification routes (notifications.yml has
// no v1 webtest, so this is ported 1:1 from the v1 handler/model behaviour).
// Link-share guards and mark-all live in separate top-level funcs below because
// they need a dedicated echo.Echo — interleaving setupTestEnv() inside the
// webHandlerTestV2 matrix rotates the JWT key out from under its cached env.
//
// Fixture topology (see pkg/db/fixtures/notifications.yml):
//   - #1, #2: belong to user1, both unread.
//   - #3: belongs to user2, unread — must stay invisible/untouched for user1.
func TestHumaNotification(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/notifications",
		idParam:  "notificationid",
		t:        t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Own notifications only", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)

			ids := notificationIDsFromReadAll(t, rec.Body.Bytes())
			// Exact set: user1 sees only their own notifications, not user2's #3.
			assert.ElementsMatch(t, []int64{1, 2}, ids,
				"ReadAll must return exactly {1,2}; body: %s", rec.Body.String())
			assert.NotContains(t, ids, int64(3), "user2's notification #3 must be hidden")
		})
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		t.Run("Normal - mark own as read", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"notificationid": "1"}, `{"read":true}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"id":1`)
			// A read notification carries a non-zero read_at timestamp.
			assert.NotContains(t, rec.Body.String(), `"read_at":"0001-01-01T00:00:00Z"`,
				"read_at must be set after marking read; body: %s", rec.Body.String())
		})
		t.Run("Mark own as unread", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"notificationid": "2"}, `{"read":false}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"read_at":"0001-01-01T00:00:00Z"`,
				"read_at must be zeroed when marking unread; body: %s", rec.Body.String())
		})
		t.Run("Forbidden - other user's notification (#3)", func(t *testing.T) {
			// CanUpdate scopes by notifiable_id; #3 belongs to user2.
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"notificationid": "3"}, `{"read":true}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - nonexistent", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"notificationid": "9999"}, `{"read":true}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})
}

// TestHumaNotification_MarkAllAsRead covers the custom bulk action: it marks
// every notification of the caller as read and leaves other users untouched.
func TestHumaNotification_MarkAllAsRead(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/notifications", "", token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"message":"success"`)

	// Re-list and confirm all of user1's notifications are now read.
	list := humaRequest(t, e, http.MethodGet, "/api/v2/notifications", "", token, "")
	require.Equal(t, http.StatusOK, list.Code, "body: %s", list.Body.String())
	assert.NotContains(t, list.Body.String(), `"read_at":"0001-01-01T00:00:00Z"`,
		"every notification must be read after mark-all; body: %s", list.Body.String())

	// user2's notification #3 must remain untouched (unread).
	otherList := humaRequest(t, e, http.MethodGet, "/api/v2/notifications", "", humaTokenFor(t, &testuser2), "")
	require.Equal(t, http.StatusOK, otherList.Code, "body: %s", otherList.Body.String())
	assert.Contains(t, otherList.Body.String(), `"read_at":"0001-01-01T00:00:00Z"`,
		"another user's notifications must stay unread; body: %s", otherList.Body.String())
}

// TestHumaNotification_LinkShareForbidden ports v1's guard: a link-share auth
// has no notifications, so list / mark-read / mark-all all refuse it (403).
func TestHumaNotification_LinkShareForbidden(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
		ID:          1,
		Hash:        "test",
		ProjectID:   1,
		Permission:  models.PermissionRead,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	})
	require.NoError(t, err)

	t.Run("list", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/notifications", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("mark read", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/notifications/1", `{"read":true}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("mark all read", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/notifications", "", token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}

func notificationIDsFromReadAll(t *testing.T, body []byte) []int64 {
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
