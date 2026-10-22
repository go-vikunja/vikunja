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
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

var testuser13 = user.User{ID: 13, Username: "user13", Issuer: "local"}

// Fixtures: user 13 owns project 20 and cannot read project 1 (user 1's).
func TestHumaNotification_ProjectPermissions(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	readableID := seedNotification(t, s, 13, &models.TaskCommentNotification{
		Doer:    &user.User{ID: 1},
		Task:    &models.Task{ID: 35, ProjectID: 20, Index: 1, Title: "readable-task-title"},
		Comment: &models.TaskComment{ID: 1, Comment: "readable comment"},
		Project: &models.Project{ID: 20, Title: "readable-project-title"},
	})
	revokedID := seedNotification(t, s, 13, &models.TaskCommentNotification{
		Doer:    &user.User{ID: 1},
		Task:    &models.Task{ID: 1, ProjectID: 1, Index: 1, Title: "revoked-task-title"},
		Comment: &models.TaskComment{ID: 2, Comment: "revoked comment"},
		Project: &models.Project{ID: 1, Title: "revoked-project-title"},
	})
	accountID := seedNotification(t, s, 13, &models.TeamMemberAddedNotification{
		Member: &user.User{ID: 13},
		Doer:   &user.User{ID: 1},
		Team:   &models.Team{ID: 1, Name: "account-level-team"},
	})
	require.NoError(t, s.Commit())

	t.Run("json list", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/notifications", "", humaTokenFor(t, &testuser13), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		ids := notificationIDsFromReadAll(t, rec.Body.Bytes())
		assert.ElementsMatch(t, []int64{readableID, accountID}, ids,
			"the readable and the account-level notification must be returned, the revoked one must not")

		body := rec.Body.String()
		assert.Contains(t, body, "readable-project-title", "body: %s", body)
		assert.NotContains(t, body, "revoked comment",
			"the payload of a filtered notification must not leak; body: %s", body)
		assert.Contains(t, body, "account-level-team", "body: %s", body)
	})

	t.Run("atom feed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/notifications.atom", nil)
		req.SetBasicAuth("user13", feedsTokenUser13)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		require.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "application/atom+xml"),
			"expected atom content type, got %q", rec.Header().Get(echo.HeaderContentType))

		// i18n yields the bare message key under test, so entry titles cannot
		// identify an entry here.
		body := rec.Body.String()
		assert.Contains(t, body, feedEntryID(readableID),
			"the feed must carry a notification about a readable project; body: %s", body)
		assert.NotContains(t, body, feedEntryID(revokedID),
			"the feed must apply the same check as the json endpoint; body: %s", body)
		assert.Contains(t, body, feedEntryID(accountID),
			"account-level notifications must stay in the feed; body: %s", body)
	})

	t.Run("mark as read is refused for a notification the caller cannot read", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/notifications/"+strconv.FormatInt(revokedID, 10),
			`{"read":true}`, humaTokenFor(t, &testuser13), "")
		assert.Equal(t, http.StatusForbidden, rec.Code,
			"marking a notification read echoes its payload back; body: %s", rec.Body.String())
	})

	t.Run("mark as read still works for a readable notification", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/notifications/"+strconv.FormatInt(readableID, 10),
			`{"read":true}`, humaTokenFor(t, &testuser13), "")
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	})
}

func feedEntryID(notificationID int64) string {
	return "vikunja-notification-" + strconv.FormatInt(notificationID, 10)
}

func seedNotification(t *testing.T, s *xorm.Session, notifiableID int64, n notifications.Notification) int64 {
	t.Helper()
	content, err := json.Marshal(n)
	require.NoError(t, err)
	dbn := &notifications.DatabaseNotification{
		NotifiableID: notifiableID,
		Notification: json.RawMessage(content),
		Name:         n.Name(),
		ProjectID:    notifications.ProjectIDOf(n),
	}
	_, err = s.Insert(dbn)
	require.NoError(t, err)
	return dbn.ID
}
