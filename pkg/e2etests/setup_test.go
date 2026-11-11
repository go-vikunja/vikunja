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
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupE2ETestEnvDrainsPendingHandlers covers the flake where a webhook
// delivery handler left over from a finished test still held a DB session while
// the next test seeded fixtures, making LoadFixtures fail with
// "database table is locked: webhooks".
//
// Cancelling the test context is not enough: sendWebhookPayload issues its HTTP
// request with context.Background(), so an in-flight delivery outlives the
// router shutdown. The webhook target below blocks until releaseAfter has
// elapsed; if the second setup returns before that, it did not drain.
func TestSetupE2ETestEnvDrainsPendingHandlers(t *testing.T) {
	const releaseAfter = 500 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e, err := setupE2ETestEnv(ctx)
	require.NoError(t, err)

	delivering := make(chan struct{}, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	// Deferred below target.Close() so it runs first: Close blocks on the
	// in-flight delivery, which only returns once release is closed.
	releaseNow := func() { releaseOnce.Do(func() { close(release) }) }

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case delivering <- struct{}{}:
		default:
		}
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	// Drop the example.com fixture webhook so only the blocking target fires.
	require.NoError(t, db.LoadFixtures())
	s := db.NewSession()
	_, err = s.Where("id = ?", 1).Delete(&models.Webhook{})
	require.NoError(t, err)
	_, err = s.Insert(&models.Webhook{
		ID:          20,
		TargetURL:   target.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())

	_, err = testUpdateWithUser(e, t, &testuser1,
		map[string]string{"projecttask": "1"},
		`{"title":"drain test"}`,
	)
	require.NoError(t, err)

	select {
	case <-delivering:
	case <-time.After(10 * time.Second):
		t.Fatal("webhook delivery handler never started")
	}

	cancel()

	var released atomic.Bool
	time.AfterFunc(releaseAfter, func() {
		released.Store(true)
		releaseNow()
	})
	defer releaseNow()

	nextCtx, cancelNext := context.WithCancel(context.Background())
	defer cancelNext()

	_, err = setupE2ETestEnv(nextCtx)
	require.NoError(t, err)

	assert.True(t, released.Load(), "setupE2ETestEnv seeded fixtures while a webhook handler was still holding a DB session")
}
