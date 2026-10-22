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

// TestHumaRateLimitUnauthenticated covers rate limiting with the default "user"
// kind: v2 rate limits one group which also serves its public routes, so those
// requests reach the middleware without any claims to key on.
func TestHumaRateLimitUnauthenticated(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)

	config.RateLimitEnabled.Set(true)
	config.RateLimitKind.Set("user")
	config.RateLimitLimit.Set(2)
	defer func() {
		config.RateLimitEnabled.Set(false)
		config.RateLimitLimit.Set(100)
	}()

	e := routes.NewEcho()
	routes.RegisterRoutes(e)

	// Both requests are counted against the same ip-derived key, so the
	// remaining budget must run down across them - not reset per path.
	for _, tc := range []struct {
		path          string
		wantRemaining string
	}{
		{"/api/v2/info", "1"},
		{"/api/v2/health", "0"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, tc.path, "", "", "")
			assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Limit"))
			assert.Equal(t, tc.wantRemaining, rec.Header().Get("X-RateLimit-Remaining"))
		})
	}

	t.Run("limits by ip", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/info", "", "", "")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())
		assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
	})
}
