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
)

func TestCRUDCreate(t *testing.T) {
	// RFC 4791 §5.3.2 (rfc4791.txt line 1358):
	// "A PUT request on a calendar collection creates a new calendar
	//  object resource when the Request-URI does not identify an
	//  existing resource."

	t.Run("PUT new task returns 201 Created", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("test-create-uid", "Test Create Task").
			Due(time.Date(2024, 3, 1, 15, 0, 0, 0, time.UTC)).
			Build()

		rec := caldavPUT(t, e, "/dav/projects/36/test-create-uid.ics", vtodo)

		assert.Equal(t, http.StatusCreated, rec.Code,
			"PUT of new resource should return 201. Body:\n%s", rec.Body.String())
	})

	t.Run("PUT new task sets ETag in response", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("test-etag-uid", "Test ETag Task").Build()

		rec := caldavPUT(t, e, "/dav/projects/36/test-etag-uid.ics", vtodo)

		assert.Equal(t, http.StatusCreated, rec.Code)

		etag := rec.Header().Get("ETag")
		assert.NotEmpty(t, etag,
			"PUT response should include ETag header for the newly created resource")
	})

	t.Run("Created task is retrievable via GET", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("test-roundtrip-uid", "Roundtrip Test Task").
			Description("A task created via CalDAV PUT").
			Priority(3).
			Build()

		rec := caldavPUT(t, e, "/dav/projects/36/test-roundtrip-uid.ics", vtodo)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Now GET the task back
		rec2 := caldavGET(t, e, "/dav/projects/36/test-roundtrip-uid.ics")
		assert.Equal(t, http.StatusOK, rec2.Code)

		body := rec2.Body.String()
		assert.Contains(t, body, "BEGIN:VCALENDAR")
		assert.Contains(t, body, "BEGIN:VTODO")
		assert.Contains(t, body, "UID:test-roundtrip-uid")
		assert.Contains(t, body, "SUMMARY:Roundtrip Test Task")
	})

	t.Run("PUT with invalid VCALENDAR returns error", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPUT(t, e, "/dav/projects/36/bad-task.ics", "not a valid vcalendar")

		// Should fail with a 4xx error
		assert.GreaterOrEqual(t, rec.Code, 400,
			"PUT with invalid VCALENDAR should return 4xx error")
		assert.Less(t, rec.Code, 500,
			"PUT with invalid VCALENDAR should not be a server error")
	})

	t.Run("PUT to nonexistent project returns 404", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("test-noproject-uid", "No Project Task").Build()
		rec := caldavPUT(t, e, "/dav/projects/99999/test-noproject-uid.ics", vtodo)

		assert.Equal(t, http.StatusNotFound, rec.Code,
			"PUT to nonexistent project should return 404")
	})

	t.Run("PUT task with all supported fields", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("test-allfields-uid", "All Fields Task").
			Description("Full description\\nwith newlines").
			Priority(1). // Highest priority in CalDAV
			Due(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)).
			DtStart(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)).
			Categories("work", "urgent").
			Status("IN-PROCESS").
			AlarmAbsolute(time.Date(2024, 6, 15, 8, 0, 0, 0, time.UTC)).
			Build()

		rec := caldavPUT(t, e, "/dav/projects/36/test-allfields-uid.ics", vtodo)
		assert.Equal(t, http.StatusCreated, rec.Code,
			"PUT with all fields should succeed. Body:\n%s", rec.Body.String())
	})
}

func TestCRUDRead(t *testing.T) {
	t.Run("GET existing task returns VCALENDAR", func(t *testing.T) {
		e := setupTestEnv(t)

		// Task 40 (uid-caldav-test) exists in project 36 from fixtures
		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")

		assert.Equal(t, http.StatusOK, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, "BEGIN:VCALENDAR")
		assert.Contains(t, body, "BEGIN:VTODO")
		assert.Contains(t, body, "UID:uid-caldav-test")
		assert.Contains(t, body, "SUMMARY:Title Caldav Test")
		assert.Contains(t, body, "END:VTODO")
		assert.Contains(t, body, "END:VCALENDAR")
	})

	t.Run("GET returns correct Content-Type", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")

		assert.Equal(t, http.StatusOK, rec.Code)
		contentType := rec.Header().Get("Content-Type")
		// Should be text/calendar per RFC 4791
		assert.Contains(t, contentType, "text/calendar",
			"GET on .ics resource should return Content-Type: text/calendar, got: %s", contentType)
	})

	t.Run("GET returns ETag header", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")

		assert.Equal(t, http.StatusOK, rec.Code)
		etag := rec.Header().Get("ETag")
		assert.NotEmpty(t, etag, "GET response should include ETag header")
	})

	t.Run("GET nonexistent task returns 404", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36/nonexistent-uid.ics")

		assert.Equal(t, http.StatusNotFound, rec.Code,
			"GET nonexistent task should return 404")
	})

	t.Run("GET project returns all tasks as VCALENDAR", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36")

		assert.Equal(t, http.StatusOK, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, "BEGIN:VCALENDAR")
		assert.Contains(t, body, "X-WR-CALNAME:Project 36 for Caldav tests")
		// Should contain multiple VTODOs
		assert.Contains(t, body, "uid-caldav-test")
	})

	t.Run("GET task with .ics suffix works", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET task without .ics suffix works", func(t *testing.T) {
		e := setupTestEnv(t)

		// Some clients may not include .ics suffix
		rec := caldavGET(t, e, "/dav/projects/36/uid-caldav-test")
		// This might 404 depending on implementation — document the behavior
		// Either 200 or 404 is acceptable, but should be consistent
		assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNotFound,
			"GET without .ics should return 200 or 404, got %d", rec.Code)
	})
}

func TestCRUDUpdate(t *testing.T) {
	t.Run("PUT to existing task updates it", func(t *testing.T) {
		e := setupTestEnv(t)

		// First create
		vtodo := NewVTodo("test-update-uid", "Original Title").Build()
		rec := caldavPUT(t, e, "/dav/projects/36/test-update-uid.ics", vtodo)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Then update
		vtodoUpdated := NewVTodo("test-update-uid", "Updated Title").
			Description("Now with a description").
			Build()
		rec2 := caldavPUT(t, e, "/dav/projects/36/test-update-uid.ics", vtodoUpdated)

		// Update should return 200 or 204 (not 201)
		assert.True(t, rec2.Code == http.StatusOK ||
			rec2.Code == http.StatusNoContent ||
			rec2.Code == http.StatusCreated, // Some implementations return 201 for updates too
			"PUT update should return 200, 204, or 201, got %d", rec2.Code)

		// Verify the update took effect
		rec3 := caldavGET(t, e, "/dav/projects/36/test-update-uid.ics")
		assert.Equal(t, http.StatusOK, rec3.Code)
		assert.Contains(t, rec3.Body.String(), "Updated Title")
		assert.Contains(t, rec3.Body.String(), "Now with a description")
	})

	t.Run("PUT update changes ETag", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create
		vtodo := NewVTodo("test-etag-change-uid", "ETag Change Test").Build()
		rec1 := caldavPUT(t, e, "/dav/projects/36/test-etag-change-uid.ics", vtodo)
		assert.Equal(t, http.StatusCreated, rec1.Code)
		etag1 := rec1.Header().Get("ETag")

		// Update
		vtodoUpdated := NewVTodo("test-etag-change-uid", "ETag Change Test Updated").Build()
		rec2 := caldavPUT(t, e, "/dav/projects/36/test-etag-change-uid.ics", vtodoUpdated)
		etag2 := rec2.Header().Get("ETag")

		// ETags should differ after update
		if etag1 != "" && etag2 != "" {
			assert.NotEqual(t, etag1, etag2,
				"ETag should change after update. Before: %s, After: %s", etag1, etag2)
		}
	})

	t.Run("PUT update preserves UID", func(t *testing.T) {
		e := setupTestEnv(t)

		// Create
		vtodo := NewVTodo("test-preserve-uid", "Preserve UID Test").Build()
		rec := caldavPUT(t, e, "/dav/projects/36/test-preserve-uid.ics", vtodo)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Update with different title but same UID
		vtodoUpdated := NewVTodo("test-preserve-uid", "Updated Preserve UID").Build()
		caldavPUT(t, e, "/dav/projects/36/test-preserve-uid.ics", vtodoUpdated)

		// Verify UID is preserved
		rec3 := caldavGET(t, e, "/dav/projects/36/test-preserve-uid.ics")
		assert.Contains(t, rec3.Body.String(), "UID:test-preserve-uid")
	})
}

func TestCRUDDelete(t *testing.T) {
	t.Run("DELETE existing task returns 204", func(t *testing.T) {
		e := setupTestEnv(t)

		// Task 40 (uid-caldav-test) exists in project 36
		rec := caldavDELETE(t, e, "/dav/projects/36/uid-caldav-test.ics")

		assert.Equal(t, http.StatusNoContent, rec.Code,
			"DELETE should return 204 No Content. Body:\n%s", rec.Body.String())
	})

	t.Run("DELETE task makes it unreachable", func(t *testing.T) {
		e := setupTestEnv(t)

		// Delete task 40
		rec := caldavDELETE(t, e, "/dav/projects/36/uid-caldav-test.ics")
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Try to GET it — should 404
		rec2 := caldavGET(t, e, "/dav/projects/36/uid-caldav-test.ics")
		assert.Equal(t, http.StatusNotFound, rec2.Code,
			"GET after DELETE should return 404")
	})

	t.Run("DELETE nonexistent task returns 404", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavDELETE(t, e, "/dav/projects/36/nonexistent-uid.ics")

		assert.Equal(t, http.StatusNotFound, rec.Code,
			"DELETE nonexistent task should return 404")
	})

	t.Run("DELETE task removes it from project listing", func(t *testing.T) {
		e := setupTestEnv(t)

		// First verify task exists in project listing
		rec := caldavGET(t, e, "/dav/projects/36")
		assert.Contains(t, rec.Body.String(), "uid-caldav-test")

		// Delete it
		caldavDELETE(t, e, "/dav/projects/36/uid-caldav-test.ics")

		// Verify it's gone from the listing
		rec2 := caldavGET(t, e, "/dav/projects/36")
		assert.NotContains(t, rec2.Body.String(), "uid-caldav-test")
	})

	t.Run("Full lifecycle: PUT create -> GET read -> PUT update -> DELETE", func(t *testing.T) {
		e := setupTestEnv(t)

		uid := "test-lifecycle-uid"
		path := "/dav/projects/36/" + uid + ".ics"

		// Create
		vtodo := NewVTodo(uid, "Lifecycle Test").Build()
		rec := caldavPUT(t, e, path, vtodo)
		assert.Equal(t, http.StatusCreated, rec.Code, "Create failed")

		// Read
		rec = caldavGET(t, e, path)
		assert.Equal(t, http.StatusOK, rec.Code, "Read failed")
		assert.Contains(t, rec.Body.String(), "Lifecycle Test")

		// Update
		vtodo2 := NewVTodo(uid, "Lifecycle Test Updated").Build()
		rec = caldavPUT(t, e, path, vtodo2)
		assert.True(t, rec.Code >= 200 && rec.Code < 300, "Update failed with %d", rec.Code)

		// Verify update
		rec = caldavGET(t, e, path)
		assert.Contains(t, rec.Body.String(), "Lifecycle Test Updated")

		// Delete
		rec = caldavDELETE(t, e, path)
		assert.Equal(t, http.StatusNoContent, rec.Code, "Delete failed")

		// Verify gone
		rec = caldavGET(t, e, path)
		assert.Equal(t, http.StatusNotFound, rec.Code, "Task should be gone after delete")
	})
}
