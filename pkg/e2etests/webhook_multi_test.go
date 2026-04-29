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
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebhookFailingSiblingDoesNotBlockOthers reproduces bug #2569:
// When two webhooks are configured on a project and the first one fails,
// the second one must still receive its payload.
func TestWebhookFailingSiblingDoesNotBlockOthers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	// Good webhook server — captures the delivered payload.
	goodReceived := make(chan []byte, 4)
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		select {
		case goodReceived <- body:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer good.Close()

	// Bad webhook server — always responds 500 so sendWebhookPayload errors out.
	badHits := make(chan struct{}, 16)
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case badHits <- struct{}{}:
		default:
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer bad.Close()

	// Clean slate + insert two webhooks listening on task.updated for project 1.
	// Use explicit ids so the bad webhook (id=10) is strictly ordered before
	// the good one (id=11) once WebhookListener.Handle applies its ORDER BY id.
	// This makes the test independent of auto-increment state and DB insert
	// ordering. Delete the fixture webhook id=1 (example.com target) first so
	// it does not pollute this test with unrelated delivery failures.
	require.NoError(t, db.LoadFixtures())
	s := db.NewSession()
	defer s.Close()
	_, err = s.Where("id = ?", 1).Delete(&models.Webhook{})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		ID:          10,
		TargetURL:   bad.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		ID:          11,
		TargetURL:   good.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	// Trigger task.updated — this drives the full async pipeline.
	rec, err := testUpdateWithUser(e, t, &testuser1,
		map[string]string{"projecttask": "1"},
		`{"title":"2569 sibling test"}`,
	)
	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), `"title":"2569 sibling test"`)

	// The good webhook MUST receive the payload, even though the bad one
	// is iterated first and fails.
	select {
	case <-goodReceived:
		// success
	case <-time.After(10 * time.Second):
		t.Fatal("good webhook did not receive payload within 10s — #2569 regression")
	}
}

// TestWebhookGoodDeliveredOnceDespiteSiblingRetries verifies that when one
// webhook keeps failing and retries via the watermill middleware, a
// sibling webhook that succeeds on the first attempt is NOT re-delivered
// as a side effect of those retries.
func TestWebhookGoodDeliveredOnceDespiteSiblingRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	// Count how many times the good webhook is invoked.
	var goodCount int32
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&goodCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer good.Close()

	// Bad webhook: always 500 — will exhaust all retries. Count every hit so
	// we can assert the watermill retry middleware actually retried the
	// failing delivery rather than giving up after one attempt.
	var badCount int32
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&badCount, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer bad.Close()

	require.NoError(t, db.LoadFixtures())
	s := db.NewSession()
	defer s.Close()
	_, err = s.Where("id = ?", 1).Delete(&models.Webhook{})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		TargetURL:   good.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		TargetURL:   bad.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	rec, err := testUpdateWithUser(e, t, &testuser1,
		map[string]string{"projecttask": "1"},
		`{"title":"2569 no-duplicate test"}`,
	)
	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), `"title":"2569 no-duplicate test"`)

	// Wait long enough for the bad webhook to exhaust its retries.
	// InitEventsForTesting configures 3 retries at 50ms initial, 2x multiplier,
	// max 1s — so the total window is ~350ms. Wait 3s to be safe.
	time.Sleep(3 * time.Second)

	got := atomic.LoadInt32(&goodCount)
	assert.Equal(t, int32(1), got,
		"good webhook should be delivered exactly once; got %d deliveries", got)

	// The bad webhook should have been retried by the watermill retry
	// middleware. We assert at least 3 attempts (1 initial + 2 retries) to
	// prove retries actually fired — we don't assert an exact count because
	// watermill's gochannel pubsub resends nacked messages, so a permanently
	// failing delivery runs through multiple retry cycles within the wait
	// window rather than stopping at MaxRetries+1.
	gotBad := atomic.LoadInt32(&badCount)
	assert.GreaterOrEqual(t, gotBad, int32(3),
		"bad webhook should be retried at least 3 times; got %d", gotBad)
}

// TestWebhookFlakyRetriesSucceed verifies that a webhook which fails a
// couple of times before succeeding is eventually delivered via the
// watermill retry middleware — proving that returning an error from the
// delivery listener still triggers retries.
func TestWebhookFlakyRetriesSucceed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	var hits int32
	done := make(chan struct{}, 1)
	flaky := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n < 3 {
			http.Error(w, "try again", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		select {
		case done <- struct{}{}:
		default:
		}
	}))
	defer flaky.Close()

	require.NoError(t, db.LoadFixtures())
	s := db.NewSession()
	defer s.Close()
	_, err = s.Where("id = ?", 1).Delete(&models.Webhook{})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		TargetURL:   flaky.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	rec, err := testUpdateWithUser(e, t, &testuser1,
		map[string]string{"projecttask": "1"},
		`{"title":"2569 flaky test"}`,
	)
	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), `"title":"2569 flaky test"`)

	select {
	case <-done:
		// Third attempt succeeded — retries worked.
	case <-time.After(10 * time.Second):
		t.Fatalf("flaky webhook never succeeded after %d attempts", atomic.LoadInt32(&hits))
	}

	// A little grace for any late in-flight retries before we drop the server.
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(3), atomic.LoadInt32(&hits),
		"expected exactly 3 attempts (2 failures + 1 success)")
}

// TestWebhookDeletedBetweenFanoutAndDelivery is a unit-level test of the
// delivery listener: if the underlying webhook row is gone by the time the
// delivery listener runs, Handle must return nil (no retry) and not error.
func TestWebhookDeletedBetweenFanoutAndDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	// Build a delivery event for a webhook id that does not exist.
	evt := &models.WebhookDeliveryEvent{
		WebhookID: 9_999_999,
		Payload: &models.WebhookPayload{
			EventName: "task.updated",
			Time:      time.Now(),
			Data:      map[string]interface{}{},
		},
	}
	body, err := json.Marshal(evt)
	require.NoError(t, err)

	msg := message.NewMessage(watermill.NewUUID(), body)
	listener := &models.WebhookDeliveryListener{}

	// Handle must return nil (quiet drop, no retry).
	require.NoError(t, listener.Handle(msg))
}
