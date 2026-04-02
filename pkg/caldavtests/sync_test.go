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

package caldavtests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestETagBehavior(t *testing.T) {
	t.Run("GET returns ETag header", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		assert.Equal(t, http.StatusOK, rec.Code)

		etag := rec.Header().Get("ETag")
		assert.NotEmpty(t, etag, "GET should return an ETag header")
		// ETag must be a quoted string per HTTP spec
		assert.True(t, len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"',
			"ETag must be a quoted string. Got: %s", etag)
	})

	t.Run("Same resource returns same ETag on repeated GET", func(t *testing.T) {
		e := setupTestEnv(t)

		rec1 := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		rec2 := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")

		etag1 := rec1.Header().Get("ETag")
		etag2 := rec2.Header().Get("ETag")

		assert.Equal(t, etag1, etag2,
			"Same resource should return same ETag on consecutive GETs")
	})

	t.Run("ETag changes after PUT update", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create a task
		vtodo := NewVTodo("etag-change-test", "ETag Change Test").Build()
		rec1 := caldavPUT(t, e, "/dav/projects/36/etag-change-test.ics", vtodo)
		require.Equal(t, http.StatusCreated, rec1.Code)

		// Get its ETag
		rec2 := caldavGET(t, e, "/dav/projects/36/etag-change-test.ics")
		etag1 := rec2.Header().Get("ETag")

		// ETag uses second-precision timestamps, so we must wait to ensure a different value
		time.Sleep(time.Second)

		// Update the task
		vtodoUpdated := NewVTodo("etag-change-test", "ETag Change Test UPDATED").
			DtStamp(time.Now().Add(time.Second).UTC()).
			Build()
		rec3 := caldavPUT(t, e, "/dav/projects/36/etag-change-test.ics", vtodoUpdated)
		require.True(t, rec3.Code >= 200 && rec3.Code < 300)

		// Get the new ETag
		rec4 := caldavGET(t, e, "/dav/projects/36/etag-change-test.ics")
		etag2 := rec4.Header().Get("ETag")

		if etag1 != "" && etag2 != "" {
			assert.NotEqual(t, etag1, etag2,
				"ETag should change after task is updated. Before: %s, After: %s", etag1, etag2)
		}
	})

	t.Run("PROPFIND ETag matches GET ETag", func(t *testing.T) {
		t.Skip("Known bug: caldav-go formats ETags differently in HTTP headers vs XML properties")
		e := setupTestEnv(t)

		// Get ETag via GET
		getResp := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		getETag := getResp.Header().Get("ETag")

		// Get ETag via PROPFIND
		propfindResp := caldavPROPFIND(t, e, "/dav/projects/36/uid-caldav-test.ics", "0", PropfindResourceProperties)
		ms := parseMultistatus(t, propfindResp)
		require.Len(t, ms.Responses, 1)
		propfindETag := getSuccessfulProp(t, ms.Responses[0]).GetETag

		if getETag != "" && propfindETag != "" {
			assert.Equal(t, getETag, propfindETag,
				"ETag from GET and PROPFIND should match")
		}
	})

	t.Run("Different tasks have different ETags", func(t *testing.T) {
		e := setupTestEnv(t)

		rec1 := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		rec2 := caldavGET(t, e, "/dav/projects/36/uid-caldav-test-parent-task.ics")

		etag1 := rec1.Header().Get("ETag")
		etag2 := rec2.Header().Get("ETag")

		if etag1 != "" && etag2 != "" {
			assert.NotEqual(t, etag1, etag2,
				"Different tasks should have different ETags")
		}
	})
}

func TestCTagBehavior(t *testing.T) {
	// Apple CalendarServer getctag extension:
	// A collection-level tag that changes when any resource within is modified.
	// Used by DAVx5, Thunderbird, Apple clients for efficient sync.

	t.Run("PROPFIND on collection returns getctag", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		body := rec.Body.String()

		// Check if getctag is present (it may not be — this documents the behavior)
		assert.Contains(t, body, "getctag",
			"PROPFIND on collection should include getctag property.\n"+
				"If this fails, getctag is not implemented — clients will sync less efficiently.\n"+
				"Body:\n%s", body)
	})

	t.Run("CTag changes after task is added", func(t *testing.T) {
		e := setupTestEnv(t)

		// Get initial CTag
		rec1 := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec1, 207)
		ms1 := parseMultistatus(t, rec1)
		ctag1 := getSuccessfulProp(t, ms1.Responses[0]).GetCTag

		// Add a task
		vtodo := NewVTodo("ctag-test-add", "CTag Test Add").Build()
		addRec := caldavPUT(t, e, "/dav/projects/36/ctag-test-add.ics", vtodo)
		require.True(t, addRec.Code >= 200 && addRec.Code < 300)

		// Get new CTag
		rec2 := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec2, 207)
		ms2 := parseMultistatus(t, rec2)
		ctag2 := getSuccessfulProp(t, ms2.Responses[0]).GetCTag

		if ctag1 != "" && ctag2 != "" {
			assert.NotEqual(t, ctag1, ctag2,
				"CTag should change after a task is added. Before: %s, After: %s", ctag1, ctag2)
		}
	})

	t.Run("CTag changes after task is deleted", func(t *testing.T) {
		e := setupTestEnv(t)

		// Get initial CTag
		rec1 := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec1, 207)
		ms1 := parseMultistatus(t, rec1)
		ctag1 := getSuccessfulProp(t, ms1.Responses[0]).GetCTag

		// Delete a task
		delRec := caldavDELETE(t, e, "/dav/projects/36/uid-caldav-test.ics")
		require.Equal(t, http.StatusNoContent, delRec.Code)

		// Get new CTag
		rec2 := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec2, 207)
		ms2 := parseMultistatus(t, rec2)
		ctag2 := getSuccessfulProp(t, ms2.Responses[0]).GetCTag

		if ctag1 != "" && ctag2 != "" {
			assert.NotEqual(t, ctag1, ctag2,
				"CTag should change after a task is deleted. Before: %s, After: %s", ctag1, ctag2)
		}
	})
}

func TestConditionalRequests(t *testing.T) {
	// RFC 4918 requires support for conditional requests using ETags.
	// If-Match prevents lost updates (optimistic concurrency).
	// If-None-Match prevents unnecessary downloads.

	t.Run("PUT with matching If-Match succeeds", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create a task and get its ETag
		vtodo := NewVTodo("if-match-test", "If-Match Test").Build()
		caldavPUT(t, e, "/dav/projects/36/if-match-test.ics", vtodo)

		getRec := caldavGET(t, e, "/dav/projects/36/if-match-test.ics")
		etag := getRec.Header().Get("ETag")
		require.NotEmpty(t, etag, "Need an ETag for this test")

		// Update with correct If-Match
		vtodoUpdated := NewVTodo("if-match-test", "If-Match Test Updated").Build()
		rec := caldavRequest(t, e, http.MethodPut, "/dav/projects/36/if-match-test.ics", vtodoUpdated, map[string]string{
			"Content-Type": "text/calendar; charset=utf-8",
			"If-Match":     etag,
		})

		assert.True(t, rec.Code >= 200 && rec.Code < 300,
			"PUT with matching If-Match should succeed. Got %d", rec.Code)
	})

	t.Run("PUT with non-matching If-Match returns 412", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create a task
		vtodo := NewVTodo("if-match-fail", "If-Match Fail Test").Build()
		caldavPUT(t, e, "/dav/projects/36/if-match-fail.ics", vtodo)

		// Try to update with wrong ETag
		vtodoUpdated := NewVTodo("if-match-fail", "Should Not Update").Build()
		rec := caldavRequest(t, e, http.MethodPut, "/dav/projects/36/if-match-fail.ics", vtodoUpdated, map[string]string{
			"Content-Type": "text/calendar; charset=utf-8",
			"If-Match":     `"wrong-etag"`,
		})

		assert.Equal(t, http.StatusPreconditionFailed, rec.Code,
			"PUT with non-matching If-Match should return 412 Precondition Failed. Got %d.\n"+
				"If this fails, the server doesn't support conditional PUT — a common CalDAV bug.", rec.Code)
	})

	t.Run("GET with matching If-None-Match returns 304", func(t *testing.T) {
		t.Skip("Known limitation: caldav-go does not implement If-None-Match conditional requests")
		e := setupTestEnv(t)

		// Get the task and its ETag
		rec1 := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		etag := rec1.Header().Get("ETag")
		require.NotEmpty(t, etag, "Need an ETag for this test")

		// Request again with If-None-Match
		rec2 := caldavRequest(t, e, http.MethodGet, "/dav/projects/36/uid-caldav-test.ics", "", map[string]string{
			"If-None-Match": etag,
		})

		assert.Equal(t, http.StatusNotModified, rec2.Code,
			"GET with matching If-None-Match should return 304 Not Modified. Got %d.\n"+
				"If this fails, the server doesn't support conditional GET — clients re-download unnecessarily.", rec2.Code)
	})
}
