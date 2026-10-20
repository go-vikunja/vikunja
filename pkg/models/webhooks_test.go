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

package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookErrorResponseBodyIsTruncated(t *testing.T) {
	const bodyMarker = "A"
	const bodySize = 5 << 20

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(bytes.Repeat([]byte(bodyMarker), bodySize))
	}))
	defer ts.Close()

	// httptest binds to loopback, which the SSRF-safe client blocks by default.
	previousAllowNonRoutable := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
	previousClient := webhookClient
	webhookClient = nil

	logDir := t.TempDir()
	log.ConfigureStandardLogger(true, "file", logDir, "ERROR", "text")

	t.Cleanup(func() {
		config.OutgoingRequestsAllowNonRoutableIPs.Set(previousAllowNonRoutable)
		webhookClient = previousClient
		log.InitLogger()
	})

	w := &Webhook{ID: 42, TargetURL: ts.URL}
	err := w.sendWebhookPayload(&WebhookPayload{EventName: "task.updated"})
	require.Error(t, err)

	logged, err := os.ReadFile(filepath.Join(logDir, "standard.log"))
	require.NoError(t, err)

	assert.Contains(t, string(logged), "from webhook 42")
	loggedBodyBytes := strings.Count(string(logged), bodyMarker)
	assert.Positive(t, loggedBodyBytes, "the response body should still be logged for diagnostics")
	assert.LessOrEqual(t, loggedBodyBytes, maxWebhookErrorBodySize, "the response body must not be read and logged unbounded")
}

// Target URLs often embed the receiver's credential, so reading them needs write access.
func TestWebhook_ProjectReadNeedsWrite(t *testing.T) {
	u := &user.User{ID: 1}

	for _, tc := range []struct {
		name      string
		projectID int64
		can       bool
	}{
		{name: "owner", projectID: 1, can: true},
		{name: "read share", projectID: 9, can: false},
		{name: "write share", projectID: 10, can: true},
		{name: "no access", projectID: 2, can: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			can, _, err := (&Webhook{ProjectID: tc.projectID}).CanRead(s, u)
			require.NoError(t, err)
			assert.Equal(t, tc.can, can, "CanRead")

			_, _, _, err = (&Webhook{ProjectID: tc.projectID}).ReadAll(s, u, "", 1, 50)
			if tc.can {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.True(t, IsErrGenericForbidden(err), "ReadAll: %v", err)
			}
		})
	}

	t.Run("archived project stays listable", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&Webhook{TargetURL: "https://example.com/archived", Events: []string{"task.updated"}, ProjectID: 22, CreatedByID: 1})
		require.NoError(t, err)
		result, _, _, err := (&Webhook{ProjectID: 22}).ReadAll(s, u, "", 1, 50)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestWebhook_JSONNeverSetsOwnership(t *testing.T) {
	w := &Webhook{ProjectID: 3}
	err := json.Unmarshal([]byte(`{"target_url":"https://example.com/x","user_id":1,"project_id":5}`), w)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/x", w.TargetURL)
	assert.Equal(t, int64(3), w.ProjectID)
	assert.Zero(t, w.UserID)
}

func TestWebhook_CanUpdateAndDeleteRouteClass(t *testing.T) {
	u := &user.User{ID: 1}

	for _, tc := range []struct {
		name    string
		webhook Webhook
		can     bool
	}{
		{name: "project webhook via its project", webhook: Webhook{ID: 1, ProjectID: 1}, can: true},
		// Webhook 3 lives in project 10, which user 1 can write too.
		{name: "project webhook via another writable project", webhook: Webhook{ID: 3, ProjectID: 1}, can: false},
		// Webhook 8 is user 1's own user-level webhook.
		{name: "user webhook via project 0", webhook: Webhook{ID: 8}, can: false},
		{name: "user webhook via a project", webhook: Webhook{ID: 8, ProjectID: 1}, can: false},
		{name: "own user webhook via user route", webhook: Webhook{ID: 8, UserID: 1}, can: true},
		{name: "other user's webhook via user route", webhook: Webhook{ID: 6, UserID: 1}, can: false},
		{name: "project webhook via user route", webhook: Webhook{ID: 1, UserID: 1}, can: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			w := tc.webhook
			can, err := w.CanUpdate(s, u)
			require.NoError(t, err)
			assert.Equal(t, tc.can, can, "CanUpdate")

			w = tc.webhook
			can, err = w.CanDelete(s, u)
			require.NoError(t, err)
			assert.Equal(t, tc.can, can, "CanDelete")
		})
	}
}
