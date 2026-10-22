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
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHuma_ErrorShapeIsRFC9457 asserts once that v2 errors use
// application/problem+json. Applies to every v2 resource; labels are the fixture.
func TestHuma_ErrorShapeIsRFC9457(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("403 Forbidden", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels/6", "", token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())

		ct := rec.Header().Get("Content-Type")
		assert.Contains(t, ct, "application/problem+json", "forbidden response must use RFC 9457 content type; got %q", ct)

		var body huma.ErrorModel
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
		assert.Equal(t, http.StatusForbidden, body.Status)
		assert.NotEmpty(t, body.Title, "title is required by RFC 9457")
	})

	t.Run("422 Validation", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/labels", `{"title":""}`, token, "")
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())

		ct := rec.Header().Get("Content-Type")
		assert.Contains(t, ct, "application/problem+json", "validation response must use RFC 9457 content type; got %q", ct)

		var body huma.ErrorModel
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
		assert.Equal(t, http.StatusUnprocessableEntity, body.Status)
		require.NotEmpty(t, body.Errors, "validation errors must include structured per-field details")
		var foundTitleError bool
		for _, detail := range body.Errors {
			if strings.Contains(detail.Location, "title") {
				foundTitleError = true
				break
			}
		}
		assert.True(t, foundTitleError, "expected at least one error detail locating `title`; got %+v", body.Errors)
	})
}

// TestHuma_HealthcheckFailureDoesNotLeakCause drives /api/v2/health into its
// 500 path end-to-end. The endpoint is unauthenticated, so the underlying
// DB/Redis driver error — which carries hosts, ports and credentials — must
// never reach the response body.
func TestHuma_HealthcheckFailureDoesNotLeakCause(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Redis is enabled but never initialized in tests, so health.Check fails.
	config.RedisEnabled.Set(true)
	defer config.RedisEnabled.Set(false)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/health", "", "", "")
	require.Equal(t, http.StatusInternalServerError, rec.Code, "body: %s", rec.Body.String())

	var body huma.ErrorModel
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body: %s", rec.Body.String())
	assert.Empty(t, body.Errors, "the healthcheck must not expose the underlying cause")
	assert.NotContains(t, rec.Body.String(), "redis", "body: %s", rec.Body.String())
}
