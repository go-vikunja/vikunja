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
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

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
	// Insert bad FIRST so it has the lower id and would be iterated before good.
	// Delete the fixture webhook id=1 (example.com target) first so it
	// does not pollute this test with unrelated delivery failures.
	require.NoError(t, db.LoadFixtures())
	s := db.NewSession()
	defer s.Close()
	_, err = s.Where("id = ?", 1).Delete(&models.Webhook{})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		TargetURL:   bad.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
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

	// Bad webhook: always 500 — will exhaust all retries.
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
}
