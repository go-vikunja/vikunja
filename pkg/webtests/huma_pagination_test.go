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
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type paginationEnvelope struct {
	Items      []json.RawMessage `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int64             `json:"total_pages"`
}

func paginationEnvelopeFrom(t *testing.T, body []byte) paginationEnvelope {
	t.Helper()
	var env paginationEnvelope
	require.NoError(t, json.Unmarshal(body, &env), "list body must be a paginated envelope: %s", string(body))
	return env
}

// TestHumaListPagination_MaxItemsPerPage pins service.maxitemsperpage as default and silent cap for v2 per_page.
func TestHumaListPagination_MaxItemsPerPage(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	previous := config.ServiceMaxItemsPerPage.GetInt()
	config.ServiceMaxItemsPerPage.Set(3)
	t.Cleanup(func() { config.ServiceMaxItemsPerPage.Set(previous) })

	token := humaTokenFor(t, &testuser1)

	t.Run("omitted per_page uses the configured maximum", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		env := paginationEnvelopeFrom(t, rec.Body.Bytes())
		assert.Len(t, env.Items, 3, "an omitted per_page must page at service.maxitemsperpage; body: %s", rec.Body.String())
		assert.Equal(t, 3, env.PerPage, "the envelope must report the effective per_page")
		assert.EqualValues(t, 7, env.Total)
		assert.EqualValues(t, 3, env.TotalPages, "total_pages must be computed from the effective per_page")
	})

	t.Run("smaller explicit per_page is honoured", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels?per_page=2", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		env := paginationEnvelopeFrom(t, rec.Body.Bytes())
		assert.Len(t, env.Items, 2)
		assert.Equal(t, 2, env.PerPage)
		assert.EqualValues(t, 4, env.TotalPages)
	})

	t.Run("larger explicit per_page is silently capped", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels?per_page=1000", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code,
			"asking for more than the instance allows must be capped, not rejected; body: %s", rec.Body.String())
		env := paginationEnvelopeFrom(t, rec.Body.Bytes())
		assert.Len(t, env.Items, 3, "body: %s", rec.Body.String())
		assert.Equal(t, 3, env.PerPage, "the envelope must report the cap, not the requested value")
		assert.EqualValues(t, 3, env.TotalPages)
	})

	t.Run("a later page uses the capped size too", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels?page=3&per_page=1000", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		env := paginationEnvelopeFrom(t, rec.Body.Bytes())
		assert.Len(t, env.Items, 1, "page 3 of 7 items at 3 per page holds the trailing item; body: %s", rec.Body.String())
		assert.Equal(t, 3, env.Page)
		assert.Equal(t, 3, env.PerPage)
	})

	t.Run("per_page=0 stays rejected", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels?per_page=0", "", token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("negative per_page stays rejected", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/labels?per_page=-1", "", token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})

	// Huma only runs Resolve on embedded structs it discovers, so cover both input shapes.
	t.Run("applies to an input embedding ListParams next to path params", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/tasks?per_page=1000", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		env := paginationEnvelopeFrom(t, rec.Body.Bytes())
		assert.Equal(t, 3, env.PerPage, "body: %s", rec.Body.String())
		assert.Len(t, env.Items, 3)
	})

	t.Run("applies to a bare ListParams input", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/token/caldav?per_page=1000", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		env := paginationEnvelopeFrom(t, rec.Body.Bytes())
		assert.Equal(t, 3, env.PerPage, "body: %s", rec.Body.String())
	})
}
