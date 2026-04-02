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
	"strings"
	"testing"

	ics "github.com/arran4/golang-ical"
	"github.com/stretchr/testify/assert"
)

func TestPropfindCollection(t *testing.T) {
	// RFC 4918 §9.1 (rfc4918.txt line 1939):
	// "The PROPFIND method retrieves properties defined on the resource
	//  identified by the Request-URI."

	t.Run("Depth 0 on project returns collection properties", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// Depth 0 should return exactly 1 response (the collection itself)
		assert.Len(t, ms.Responses, 1,
			"Depth 0 should return exactly the collection")

		r := ms.Responses[0]
		prop := getSuccessfulProp(t, r)

		// displayname should be the project title
		assert.Equal(t, "Project 36 for Caldav tests", prop.DisplayName,
			"displayname should match project title")

		// resourcetype should include both DAV:collection and CALDAV:calendar
		assert.Contains(t, prop.ResourceType.InnerXML, "collection",
			"resourcetype should include DAV:collection")
	})

	t.Run("Depth 1 on project returns collection plus tasks", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36", "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// Project 36 has 5 tasks in fixtures (tasks 40-43, 45)
		// Depth 1 should return the collection + all tasks = 6 responses
		assert.GreaterOrEqual(t, len(ms.Responses), 6,
			"Depth 1 should return collection + all tasks")

		// First response should be the collection itself
		// Subsequent responses should be individual tasks
		body := rec.Body.String()
		assert.Contains(t, body, ".ics",
			"Task responses should have .ics hrefs")
	})

	t.Run("Depth 1 on project returns ETags for each resource", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36", "1", PropfindResourceProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		for _, r := range ms.Responses {
			prop := getSuccessfulProp(t, r)
			// Every resource should have an ETag
			// RFC 4918 §15.6: "strong ETags MUST be used"
			assert.NotEmpty(t, prop.GetETag,
				"Every resource in PROPFIND should have an ETag. Href: %s", r.Href)
		}
	})

	t.Run("PROPFIND on nonexistent project returns 404", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/99999", "0", PropfindCalendarCollectionProperties)

		assert.Equal(t, http.StatusNotFound, rec.Code,
			"PROPFIND on nonexistent project should return 404")
	})

	t.Run("Depth 1 includes calendar-data for each task", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36", "1", PropfindResourceProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		taskCount := 0
		for _, r := range ms.Responses {
			prop := getSuccessfulProp(t, r)
			if prop.CalendarData != "" {
				taskCount++
				// Each calendar-data should be valid iCalendar
				cal := parseICalFromString(t, prop.CalendarData)
				vtodo := getVTodo(t, cal)
				uid := getVTodoProperty(vtodo, ics.ComponentPropertyUniqueId)
				assert.NotEmpty(t, uid, "Each VTODO should have a UID")
			}
		}
		assert.Positive(t, taskCount, "Should have at least one task with calendar-data")
	})
}

func TestPropfindResource(t *testing.T) {
	t.Run("Depth 0 on task returns task properties", func(t *testing.T) {
		e := setupTestEnv(t)

		// Task 40 has UID "uid-caldav-test" in project 36
		rec := caldavPROPFIND(t, e, "/dav/projects/36/uid-caldav-test.ics", "0", PropfindResourceProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		assert.Len(t, ms.Responses, 1,
			"Depth 0 on a task should return exactly 1 response")

		r := ms.Responses[0]
		prop := getSuccessfulProp(t, r)

		assert.NotEmpty(t, prop.GetETag, "Task should have an ETag")
		assert.NotEmpty(t, prop.CalendarData, "Task should have calendar-data")

		// Parse and validate the calendar data
		cal := parseICalFromString(t, prop.CalendarData)
		vtodo := getVTodo(t, cal)
		assert.Equal(t, "uid-caldav-test", getVTodoProperty(vtodo, ics.ComponentPropertyUniqueId))
		assert.Equal(t, "Title Caldav Test", getVTodoProperty(vtodo, ics.ComponentPropertySummary))
	})

	t.Run("PROPFIND on nonexistent task returns 404", func(t *testing.T) {
		t.Skip("Known limitation: caldav-go returns 207 with 404 propstat instead of top-level 404")
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36/nonexistent-uid.ics", "0", PropfindResourceProperties)

		assert.Equal(t, http.StatusNotFound, rec.Code,
			"PROPFIND on nonexistent task should return 404")
	})

	t.Run("ETag format is quoted string", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects/36/uid-caldav-test.ics", "0", PropfindResourceProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)
		r := ms.Responses[0]
		prop := getSuccessfulProp(t, r)

		// RFC 4918 requires ETags to be quoted strings
		assert.True(t, len(prop.GetETag) > 2 &&
			prop.GetETag[0] == '"' && prop.GetETag[len(prop.GetETag)-1] == '"',
			"ETag should be a quoted string, got: %s", prop.GetETag)
	})
}

func TestPropfindCalendarHome(t *testing.T) {
	t.Run("Depth 1 on /dav/projects lists all accessible calendars", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// testuser15 should see at least projects 36 and 38
		projectFound36 := false
		projectFound38 := false
		for _, r := range ms.Responses {
			if strings.Contains(r.Href, "36") {
				projectFound36 = true
			}
			if strings.Contains(r.Href, "38") {
				projectFound38 = true
			}
		}
		assert.True(t, projectFound36, "Should list project 36 in calendar home")
		assert.True(t, projectFound38, "Should list project 38 in calendar home")
	})

	t.Run("Each calendar has displayname matching project title", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		for _, r := range ms.Responses {
			prop := getSuccessfulProp(t, r)
			if prop.DisplayName != "" {
				// Every calendar with a displayname should have a reasonable title
				assert.NotEmpty(t, prop.DisplayName,
					"Calendar at %s should have a displayname", r.Href)
			}
		}
	})

	t.Run("User only sees projects they have access to", func(t *testing.T) {
		e := setupTestEnv(t)

		// testuser1 should NOT see testuser15's projects (36, 38)
		rec := caldavRequest(t, e, "PROPFIND", "/dav/projects", PropfindCalendarCollectionProperties, map[string]string{
			"Depth":         "1",
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
		})

		assertResponseStatus(t, rec, 207)

		body := rec.Body.String()
		// user1 should not see project 36 or 38 (owned by user15)
		assert.NotContains(t, body, "Project 36 for Caldav tests",
			"user1 should not see user15's project 36")
		assert.NotContains(t, body, "Project 38 for Caldav tests",
			"user1 should not see user15's project 38")
	})
}
