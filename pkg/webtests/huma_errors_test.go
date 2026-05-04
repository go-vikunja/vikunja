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

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHuma_ErrorShapeIsRFC9457 asserts once — across a 403 and a 422 — that
// v2 errors use application/problem+json with a `status` field. This is the
// "changed error responses" deviation from v1 and applies to every v2 resource,
// so the assertion lives once here rather than being duplicated per resource.
// Labels are used as the fixture because they are the only v2 resource today.
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
