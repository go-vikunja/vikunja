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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// webhookCapture is a test helper that starts an HTTP server to capture webhook payloads.
type webhookCapture struct {
	server   *httptest.Server
	payloads chan webhookDelivery
	mu       sync.Mutex
	received []webhookDelivery
}

type webhookDelivery struct {
	Body    []byte
	Headers http.Header
}

func newWebhookCapture() *webhookCapture {
	wc := &webhookCapture{
		payloads: make(chan webhookDelivery, 10),
	}
	wc.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		delivery := webhookDelivery{
			Body:    body,
			Headers: r.Header.Clone(),
		}
		wc.mu.Lock()
		wc.received = append(wc.received, delivery)
		wc.mu.Unlock()
		select {
		case wc.payloads <- delivery:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	return wc
}

func (wc *webhookCapture) URL() string {
	return wc.server.URL
}

func (wc *webhookCapture) Close() {
	wc.server.Close()
}

// waitForPayload waits for a webhook payload to arrive within 10 seconds.
func (wc *webhookCapture) waitForPayload(t *testing.T) webhookDelivery {
	t.Helper()
	select {
	case d := <-wc.payloads:
		return d
	case <-time.After(10 * time.Second):
		t.Fatal("Webhook payload not received within timeout")
		return webhookDelivery{} // unreachable
	}
}

// assertNoPayload asserts that no webhook payload arrives within the given duration.
func (wc *webhookCapture) assertNoPayload(t *testing.T, wait time.Duration) {
	t.Helper()
	select {
	case d := <-wc.payloads:
		t.Fatalf("Expected no webhook payload but received one: %s", string(d.Body))
	case <-time.After(wait):
		// success — nothing arrived
	}
}

// insertUserWebhook creates a user-level webhook (no project_id) in the database.
func insertUserWebhook(t *testing.T, userID int64, targetURL string, evts []string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&models.Webhook{
		TargetURL:   targetURL,
		Events:      evts,
		UserID:      userID,
		CreatedByID: userID,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// insertUserWebhookWithSecret creates a user-level webhook with an HMAC secret.
func insertUserWebhookWithSecret(t *testing.T, userID int64, targetURL string, evts []string, secret string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&models.Webhook{
		TargetURL:   targetURL,
		Events:      evts,
		UserID:      userID,
		CreatedByID: userID,
		Secret:      secret,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// clearWebhooks removes all webhook rows from the database.
func clearWebhooks(t *testing.T) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("1=1").Delete(&models.Webhook{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// parseWebhookPayload parses a webhook delivery body into a structured map.
func parseWebhookPayload(t *testing.T, d webhookDelivery) map[string]interface{} {
	t.Helper()
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(d.Body, &payload))
	return payload
}

func TestUserWebhookTaskOverdueE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	insertUserWebhook(t, testuser1.ID, capture.URL(), []string{"task.overdue"})

	// Dispatch a TaskOverdueEvent directly — this simulates the overdue cron job
	err = events.Dispatch(&models.TaskOverdueEvent{
		Task: &models.Task{
			ID:        1,
			Title:     "Overdue task",
			ProjectID: 1,
		},
		User:    &testuser1,
		Project: &models.Project{ID: 1, Title: "Test Project"},
	})
	require.NoError(t, err)

	delivery := capture.waitForPayload(t)
	payload := parseWebhookPayload(t, delivery)

	assert.Equal(t, "task.overdue", payload["event_name"])
	data, ok := payload["data"].(map[string]interface{})
	require.True(t, ok, "payload.data should be a map")
	task, ok := data["task"].(map[string]interface{})
	require.True(t, ok, "payload.data.task should be a map")
	assert.Equal(t, "Overdue task", task["title"])
}

func TestUserWebhookTaskReminderFiredE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	insertUserWebhook(t, testuser1.ID, capture.URL(), []string{"task.reminder.fired"})

	// Dispatch a TaskReminderFiredEvent directly — simulates the reminder cron job
	err = events.Dispatch(&models.TaskReminderFiredEvent{
		Task: &models.Task{
			ID:        1,
			Title:     "Reminder task",
			ProjectID: 1,
		},
		User:    &testuser1,
		Project: &models.Project{ID: 1, Title: "Test Project"},
	})
	require.NoError(t, err)

	delivery := capture.waitForPayload(t)
	payload := parseWebhookPayload(t, delivery)

	assert.Equal(t, "task.reminder.fired", payload["event_name"])
	data, ok := payload["data"].(map[string]interface{})
	require.True(t, ok, "payload.data should be a map")
	task, ok := data["task"].(map[string]interface{})
	require.True(t, ok, "payload.data.task should be a map")
	assert.Equal(t, "Reminder task", task["title"])
}

func TestUserWebhookDoesNotFireForOtherUsers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	// Create a user-level webhook for user 2
	insertUserWebhook(t, 2, capture.URL(), []string{"task.overdue"})

	// Dispatch an overdue event for user 1 — user 2's webhook should NOT fire
	err = events.Dispatch(&models.TaskOverdueEvent{
		Task: &models.Task{
			ID:        1,
			Title:     "Overdue for user 1",
			ProjectID: 1,
		},
		User:    &testuser1,
		Project: &models.Project{ID: 1, Title: "Test Project"},
	})
	require.NoError(t, err)

	capture.assertNoPayload(t, 3*time.Second)
}

func TestUserWebhookAndProjectWebhookBothFire(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	userCapture := newWebhookCapture()
	defer userCapture.Close()
	projectCapture := newWebhookCapture()
	defer projectCapture.Close()

	clearWebhooks(t)

	// User-level webhook for user 1 listening to task.overdue
	insertUserWebhook(t, testuser1.ID, userCapture.URL(), []string{"task.overdue"})

	// Project-level webhook for project 1 listening to task.overdue
	s := db.NewSession()
	_, err = s.Insert(&models.Webhook{
		TargetURL:   projectCapture.URL(),
		Events:      []string{"task.overdue"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	// Dispatch overdue event — both project-level and user-level webhooks should fire
	err = events.Dispatch(&models.TaskOverdueEvent{
		Task: &models.Task{
			ID:        1,
			Title:     "Both webhooks overdue",
			ProjectID: 1,
		},
		User:    &testuser1,
		Project: &models.Project{ID: 1, Title: "Test Project"},
	})
	require.NoError(t, err)

	// Both should receive the payload
	projectDelivery := projectCapture.waitForPayload(t)
	projectPayload := parseWebhookPayload(t, projectDelivery)
	assert.Equal(t, "task.overdue", projectPayload["event_name"])

	userDelivery := userCapture.waitForPayload(t)
	userPayload := parseWebhookPayload(t, userDelivery)
	assert.Equal(t, "task.overdue", userPayload["event_name"])
}

func TestUserWebhookHMACSigning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	secret := "test-hmac-secret-for-user-webhook"
	insertUserWebhookWithSecret(t, testuser1.ID, capture.URL(), []string{"task.overdue"}, secret)

	err = events.Dispatch(&models.TaskOverdueEvent{
		Task: &models.Task{
			ID:        1,
			Title:     "HMAC overdue",
			ProjectID: 1,
		},
		User:    &testuser1,
		Project: &models.Project{ID: 1, Title: "Test Project"},
	})
	require.NoError(t, err)

	delivery := capture.waitForPayload(t)

	// Verify the HMAC signature header is present and correct
	signature := delivery.Headers.Get("X-Vikunja-Signature")
	require.NotEmpty(t, signature, "X-Vikunja-Signature header should be set")

	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write(delivery.Body)
	require.NoError(t, err)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expectedSig, signature, "HMAC signature should match")
}

func TestUserWebhookOnlyMatchesSubscribedEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	// Subscribe only to task.reminder.fired — task.overdue should NOT trigger it
	insertUserWebhook(t, testuser1.ID, capture.URL(), []string{"task.reminder.fired"})

	err = events.Dispatch(&models.TaskOverdueEvent{
		Task: &models.Task{
			ID:        1,
			Title:     "Wrong event",
			ProjectID: 1,
		},
		User:    &testuser1,
		Project: &models.Project{ID: 1, Title: "Test Project"},
	})
	require.NoError(t, err)

	capture.assertNoPayload(t, 3*time.Second)
}

func TestUserWebhookDoesNotFireForProjectEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	// User-level webhook subscribed to task.updated (a non-user-directed event)
	insertUserWebhook(t, testuser1.ID, capture.URL(), []string{"task.updated"})

	// Update task via the web handler — user-level webhooks should NOT fire
	// for project-scoped events like task.updated
	_, err = testUpdateWithUser(e, t, &testuser1,
		map[string]string{"projecttask": "1"},
		`{"title":"Should not trigger user webhook"}`,
	)
	require.NoError(t, err)

	capture.assertNoPayload(t, 3*time.Second)
}

func TestUserWebhookTasksOverdueBatchE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	capture := newWebhookCapture()
	defer capture.Close()

	clearWebhooks(t)
	insertUserWebhook(t, testuser1.ID, capture.URL(), []string{"tasks.overdue"})

	// Dispatch a batch TasksOverdueEvent
	err = events.Dispatch(&models.TasksOverdueEvent{
		Tasks: []*models.Task{
			{ID: 1, Title: "Overdue 1", ProjectID: 1},
			{ID: 2, Title: "Overdue 2", ProjectID: 1},
		},
		User: &user.User{ID: testuser1.ID, Username: testuser1.Username},
		Projects: map[int64]*models.Project{
			1: {ID: 1, Title: "Test Project"},
		},
	})
	require.NoError(t, err)

	delivery := capture.waitForPayload(t)
	payload := parseWebhookPayload(t, delivery)

	assert.Equal(t, "tasks.overdue", payload["event_name"])
	data, ok := payload["data"].(map[string]interface{})
	require.True(t, ok, "payload.data should be a map")
	tasks, ok := data["tasks"].([]interface{})
	require.True(t, ok, "payload.data.tasks should be an array")
	assert.Len(t, tasks, 2)
}
