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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaLinkPreview covers auth, input validation and SSRF protection of the
// link-preview endpoint. OpenGraph parsing itself is unit-tested in
// pkg/modules/linkpreview without touching the network.
func TestHumaLinkPreview(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/link-preview?url=http://example.com", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Rejects non-http scheme", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/link-preview?url=ftp://example.com/x", "", token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Rejects empty url", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/link-preview?url=", "", token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Blocks loopback address (SSRF)", func(t *testing.T) {
		// The SSRF-safe client refuses to dial non-routable IPs, so a loopback
		// target fails before any connection is made.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/link-preview?url=http://127.0.0.1:80/", "", token, "")
		assert.Equal(t, http.StatusBadGateway, rec.Code, "body: %s", rec.Body.String())
	})
}
