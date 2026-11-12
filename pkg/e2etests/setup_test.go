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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/require"
)

// TestSetupE2ETestEnvCancelsInFlightWebhook keeps the target blocked while the
// next setup drains the cancelled handler and replaces database fixtures.
func TestSetupE2ETestEnvCancelsInFlightWebhook(t *testing.T) {

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
	defer releaseNow()

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

	nextCtx, cancelNext := context.WithCancel(context.Background())
	defer cancelNext()

	ready := make(chan error, 1)
	go func() { _, err := setupE2ETestEnv(nextCtx); ready <- err }()
	select {
	case err := <-ready:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("cancelled webhook prevented fixture setup while its target was blocked")
	}
}
