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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/license"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixture entries (pkg/db/fixtures/time_entries.yml): 1 = user1 on task 1,
// 2 = user1 on project 1, 3 = user3 on project 3 (user1 can read), 4 = user1's
// running timer on task 1. user1 (testuser1) can read all four.

// The gate is the one v2-specific concern with no model-level equivalent: every
// time-tracking route 404s on an instance without the feature.
func TestHumaTimeEntry_LicenseGate(t *testing.T) {
	t.Run("disabled feature 404s the list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{}) // licensed, but not time tracking
		defer license.ResetForTests()

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("disabled feature 404s timer/stop", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{})
		defer license.ResetForTests()

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/time-entries/timer/stop", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("disabled feature 404s the task-scoped list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{})
		defer license.ResetForTests()

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/time-entries", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("enabled feature serves the list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries", "", humaTokenFor(t, &testuser1), "")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})
}

func TestHumaTimeEntry(t *testing.T) {
	// SetForTests must come after setupTestEnv — the latter re-inits the license to free.
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureTimeTracking})
	defer license.ResetForTests()

	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/time-entries",
		idParam:  "id",
		t:        t,
		e:        e,
	}

	t.Run("ReadAll returns the readable set", func(t *testing.T) {
		rec, err := testHandler.testReadAllWithUser(nil, nil)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int64{1, 2, 3, 4}, timeEntryIDsFromReadAll(t, rec.Body.Bytes()),
			"body: %s", rec.Body.String())
	})

	t.Run("ReadOne", func(t *testing.T) {
		rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"id": "1"})
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"id":1`)
		assert.Contains(t, rec.Body.String(), `"max_permission":`)
	})

	t.Run("ReadOne forbidden for a stranger", func(t *testing.T) {
		// entry 1 is on project 1; user2 has no access to it.
		stranger := webHandlerTestV2{user: &testuser2, basePath: "/api/v2/time-entries", idParam: "id", t: t, e: e}
		_, err := stranger.testReadOneWithUser(nil, map[string]string{"id": "1"})
		require.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})

	t.Run("ReadOne of a missing entry is 404", func(t *testing.T) {
		_, err := testHandler.testReadOneWithUser(nil, map[string]string{"id": "9999"})
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
	})
}

// Create exercises the full handler path (DoCreate → CanCreate → Create →
// commit → DispatchPending) that the model-level tests bypass.
func TestHumaTimeEntry_Create(t *testing.T) {
	t.Run("saving an entry with end_time", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		body := `{"task_id":1,"start_time":"2020-01-01T09:00:00Z","end_time":"2020-01-01T10:00:00Z","comment":"work"}`
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/time-entries", body, humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"user_id":1`)
		assert.Contains(t, rec.Body.String(), `"task_id":1`)
	})

	t.Run("starting a timer without end_time", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		body := `{"task_id":1,"start_time":"2020-01-01T09:00:00Z","comment":"timer"}`
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/time-entries", body, humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"user_id":1`)
		// A running timer's end_time is null on the wire, not the zero-time sentinel.
		assert.Contains(t, rec.Body.String(), `"end_time":null`)
		assert.NotContains(t, rec.Body.String(), "0001-01-01")
	})
}

// The filter param must wire through the route into ReadAll.
func TestHumaTimeEntry_Filter(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureTimeTracking})
	defer license.ResetForTests()
	token := humaTokenFor(t, &testuser1)

	t.Run("by task", func(t *testing.T) {
		q := url.Values{"filter": {"task_id = 1"}}
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries?"+q.Encode(), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.ElementsMatch(t, []int64{1, 4}, timeEntryIDsFromReadAll(t, rec.Body.Bytes()))
	})

	t.Run("running timers via null end_time", func(t *testing.T) {
		q := url.Values{"filter": {"end_time = null"}}
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/time-entries?"+q.Encode(), "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.ElementsMatch(t, []int64{4}, timeEntryIDsFromReadAll(t, rec.Body.Bytes()))
	})
}

func TestHumaTimeEntry_TimerStop(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	license.SetForTests([]license.Feature{license.FeatureTimeTracking})
	defer license.ResetForTests()

	t.Run("stops the caller's running timer", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/time-entries/timer/stop", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"id":4`, "entry 4 is user1's running timer")
		assert.NotContains(t, rec.Body.String(), `"end_time":"0001-01-01`, "end_time must now be set")
	})

	t.Run("404 when the caller has no running timer", func(t *testing.T) {
		// user2 has no entries, so no running timer.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/time-entries/timer/stop", "", humaTokenFor(t, &testuser2), "")
		assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	})
}

func timeEntryIDsFromReadAll(t *testing.T, body []byte) []int64 {
	t.Helper()
	var resp struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(body, &resp), "ReadAll body must be a paginated envelope: %s", string(body))
	ids := make([]int64, 0, len(resp.Items))
	for _, it := range resp.Items {
		ids = append(ids, it.ID)
	}
	return ids
}
