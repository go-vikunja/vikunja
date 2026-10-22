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
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/routes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebsocketUpgradeRateLimit asserts the /ws upgrade endpoint keeps its
// per-IP floor with the global rate limiter disabled - its default. The upgrade
// is unauthenticated, so without the floor anyone could spawn connections
// (2 goroutines + a socket each) at will.
func TestWebsocketUpgradeRateLimit(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	require.False(t, config.RateLimitEnabled.GetBool(), "the global rate limiter must default to off for this test to be meaningful")

	previousLimit := config.RateLimitNoAuthRoutesLimit.GetInt64()
	config.RateLimitNoAuthRoutesLimit.Set(2)
	defer config.RateLimitNoAuthRoutesLimit.Set(previousLimit)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	// Both versions share one limiter, so the budget runs down across them
	// instead of granting a fresh one per API version.
	for _, path := range []string{"/api/v1/ws", "/api/v2/ws"} {
		t.Run("within limit "+path, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, path, "", "", "")
			assert.NotEqual(t, http.StatusTooManyRequests, rec.Code)
			assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Limit"))
		})
	}

	for _, path := range []string{"/api/v1/ws", "/api/v2/ws"} {
		t.Run("throttled "+path, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, path, "", "", "")
			assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
			assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
		})
	}
}
