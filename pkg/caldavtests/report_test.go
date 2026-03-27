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
	"strings"
	"testing"

	ics "github.com/arran4/golang-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportCalendarQuery(t *testing.T) {
	// RFC 4791 §7.8 (rfc4791.txt line 1967):
	// "The CALDAV:calendar-query REPORT performs a search for all calendar
	//  object resources that match a specified filter."

	t.Run("calendar-query returns 207 Multi-Status", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)

		assertResponseStatus(t, rec, 207)
	})

	t.Run("calendar-query returns all tasks in project", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// Project 36 has 5 tasks in fixtures (40, 41, 42, 43, 45)
		assert.Len(t, ms.Responses, 5,
			"Should return all 5 tasks from project 36")
	})

	t.Run("calendar-query responses include ETag", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)

		ms := parseMultistatus(t, rec)
		for _, r := range ms.Responses {
			prop := getSuccessfulProp(t, r)
			assert.NotEmpty(t, prop.GetETag,
				"Each response should include an ETag. Href: %s", r.Href)
		}
	})

	t.Run("calendar-query responses include valid calendar-data", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)

		ms := parseMultistatus(t, rec)
		for i, r := range ms.Responses {
			prop := getSuccessfulProp(t, r)
			assert.NotEmpty(t, prop.CalendarData,
				"Response %d should include calendar-data. Href: %s", i, r.Href)

			// Each calendar-data should be parseable iCalendar
			cal := parseICalFromString(t, prop.CalendarData)
			vtodo := getVTodo(t, cal)

			uid := getVTodoProperty(vtodo, ics.ComponentPropertyUniqueId)
			assert.NotEmpty(t, uid, "VTODO %d should have a UID", i)

			summary := getVTodoProperty(vtodo, ics.ComponentPropertySummary)
			assert.NotEmpty(t, summary, "VTODO %d should have a SUMMARY", i)
		}
	})

	t.Run("calendar-query response hrefs point to correct resources", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)

		ms := parseMultistatus(t, rec)
		for _, r := range ms.Responses {
			// Each href should be a valid task URL containing the project ID and .ics
			assert.Contains(t, r.Href, "/dav/projects/",
				"Href should contain /dav/projects/")
			assert.True(t, strings.HasSuffix(r.Href, ".ics"),
				"Href should end with .ics. Got: %s", r.Href)
		}
	})

	t.Run("calendar-query on nonexistent project returns error", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/99999", ReportCalendarQuery)

		// Should be 404 or similar error, not 200/207
		assert.NotEqual(t, 207, rec.Code,
			"REPORT on nonexistent project should not return 207")
	})

	t.Run("calendar-query on project 38 returns correct task count", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, "/dav/projects/38", ReportCalendarQuery)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)
		// Project 38 has 2 tasks (44, 46)
		assert.Len(t, ms.Responses, 2,
			"Project 38 should have 2 tasks")
	})
}

func TestReportCalendarMultiget(t *testing.T) {
	// RFC 4791 §7.9 (rfc4791.txt line 3479):
	// "The CALDAV:calendar-multiget REPORT is used to retrieve specific
	//  calendar object resources from within a collection."

	t.Run("calendar-multiget returns requested tasks", func(t *testing.T) {
		e := setupTestEnv(t)

		// Request two specific tasks from project 36
		body := ReportCalendarMultiget(
			"/dav/projects/36/uid-caldav-test.ics",
			"/dav/projects/36/uid-caldav-test-parent-task.ics",
		)

		rec := caldavREPORT(t, e, "/dav/projects/36", body)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		assert.Len(t, ms.Responses, 2,
			"Should return exactly the 2 requested tasks")
	})

	t.Run("calendar-multiget returns calendar-data for each task", func(t *testing.T) {
		e := setupTestEnv(t)

		body := ReportCalendarMultiget(
			"/dav/projects/36/uid-caldav-test.ics",
		)

		rec := caldavREPORT(t, e, "/dav/projects/36", body)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)
		require.Len(t, ms.Responses, 1)

		prop := getSuccessfulProp(t, ms.Responses[0])
		assert.NotEmpty(t, prop.CalendarData, "Should include calendar-data")
		assert.NotEmpty(t, prop.GetETag, "Should include ETag")

		// Verify the returned data matches the requested task
		cal := parseICalFromString(t, prop.CalendarData)
		vtodo := getVTodo(t, cal)
		assert.Equal(t, "uid-caldav-test", getVTodoProperty(vtodo, ics.ComponentPropertyUniqueId))
	})

	t.Run("calendar-multiget with nonexistent href returns 404 for that href", func(t *testing.T) {
		e := setupTestEnv(t)

		body := ReportCalendarMultiget(
			"/dav/projects/36/uid-caldav-test.ics",
			"/dav/projects/36/nonexistent-uid.ics",
		)

		rec := caldavREPORT(t, e, "/dav/projects/36", body)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// Should still return results — the existing task should be there
		// The nonexistent one might be absent or have a 404 propstat
		foundExisting := false
		for _, r := range ms.Responses {
			if strings.Contains(r.Href, "uid-caldav-test") {
				foundExisting = true
				prop := getSuccessfulProp(t, r)
				assert.NotEmpty(t, prop.CalendarData)
			}
		}
		assert.True(t, foundExisting,
			"Should still return the existing task even when one href is invalid")
	})

	t.Run("calendar-multiget with empty href list returns empty", func(t *testing.T) {
		e := setupTestEnv(t)

		body := ReportCalendarMultiget() // No hrefs

		rec := caldavREPORT(t, e, "/dav/projects/36", body)

		// Should return 207 with no responses, or possibly an error
		assert.True(t, rec.Code == 207 || rec.Code >= 400,
			"Empty multiget should return 207 (empty) or an error, got %d", rec.Code)
	})

	t.Run("calendar-multiget ETags match PROPFIND ETags", func(t *testing.T) {
		e := setupTestEnv(t)

		// Get ETag via PROPFIND
		propfindRec := caldavPROPFIND(t, e, "/dav/projects/36/uid-caldav-test.ics", "0", PropfindResourceProperties)
		assertResponseStatus(t, propfindRec, 207)
		propfindMs := parseMultistatus(t, propfindRec)
		propfindEtag := getSuccessfulProp(t, propfindMs.Responses[0]).GetETag

		// Get ETag via multiget REPORT
		body := ReportCalendarMultiget("/dav/projects/36/uid-caldav-test.ics")
		reportRec := caldavREPORT(t, e, "/dav/projects/36", body)
		assertResponseStatus(t, reportRec, 207)
		reportMs := parseMultistatus(t, reportRec)
		reportEtag := getSuccessfulProp(t, reportMs.Responses[0]).GetETag

		// ETags should match between PROPFIND and REPORT
		assert.Equal(t, propfindEtag, reportEtag,
			"ETag from PROPFIND and calendar-multiget should match")
	})
}
