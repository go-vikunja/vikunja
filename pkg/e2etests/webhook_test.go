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

package e2etests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskUpdateWebhookE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	// Start a test HTTP server to capture webhook payloads.
	// Use a non-blocking send so retries or duplicate deliveries don't hang.
	webhookReceived := make(chan []byte, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		select {
		case webhookReceived <- body:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Reload fixtures to start from a clean state, then insert
	// a fresh webhook for project 1 listening to "task.updated".
	require.NoError(t, db.LoadFixtures())
	s := db.NewSession()
	defer s.Close()
	_, err = s.Insert(&models.Webhook{
		TargetURL:   ts.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	// Update task 1 via the web handler — this triggers the full pipeline:
	// UpdateWeb → Task.Update() → DispatchOnCommit → s.Commit() →
	// DispatchPending → Dispatch → watermill → WebhookListener.Handle →
	// HTTP POST to ts.URL
	rec, err := testUpdateWithUser(e, t, &testuser1,
		map[string]string{"projecttask": "1"},
		`{"title":"E2E webhook test"}`,
	)
	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), `"title":"E2E webhook test"`)

	// Wait for the webhook payload to arrive via the real async pipeline
	select {
	case body := <-webhookReceived:
		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "task.updated", payload["event_name"])

		data, ok := payload["data"].(map[string]interface{})
		require.True(t, ok, "payload.data should be a map")
		task, ok := data["task"].(map[string]interface{})
		require.True(t, ok, "payload.data.task should be a map")
		assert.Equal(t, "E2E webhook test", task["title"])

	case <-time.After(10 * time.Second):
		t.Fatal("Webhook payload not received within 10s timeout")
	}
}
