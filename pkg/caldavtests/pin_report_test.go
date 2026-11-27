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
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func calendarQueryWithTodoFilter(todoFilter string) string {
	return `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-query xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:D="DAV:">
  <D:prop>
    <D:getetag/>
  </D:prop>
  <C:filter>
    <C:comp-filter name="VCALENDAR">
      <C:comp-filter name="VTODO">` + todoFilter + `</C:comp-filter>
    </C:comp-filter>
  </C:filter>
</C:calendar-query>`
}

func etagOnly(hrefs ...string) []expectedDAVResponse {
	out := make([]expectedDAVResponse, 0, len(hrefs))
	for _, href := range hrefs {
		out = append(out, expectedDAVResponse{
			href: href,
			ok:   map[xml.Name]string{davName("getetag"): anyValue},
		})
	}
	return out
}

func project36TaskHrefs() []string {
	hrefs := make([]string, 0, len(project36TaskTitles))
	for href := range project36TaskTitles {
		hrefs = append(hrefs, href)
	}
	return hrefs
}

func TestPinReportCalendarMultiget(t *testing.T) {
	e := setupTestEnv(t)
	const missing = "/dav/projects/36/nonexistent-uid.ics"

	rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarMultiget(
		pinTaskHref,
		missing,
		// Exists and is readable, but lives outside the requested collection: silently skipped.
		"/dav/projects/38/uid-caldav-test-child-task-2.ics",
	))

	require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
	assertDAVResponses(t, rec, []expectedDAVResponse{
		{
			href: pinTaskHref,
			ok: map[xml.Name]string{
				davName("getetag"):       anyValue,
				calName("calendar-data"): anyValue,
			},
		},
		{
			href:   missing,
			status: http.StatusNotFound,
		},
	})
}

func TestPinReportUnknownType(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavREPORT(t, e, "/dav/projects/36", `<?xml version="1.0" encoding="utf-8" ?>
<C:free-busy-query xmlns:C="urn:ietf:params:xml:ns:caldav">
  <C:time-range start="20060104T140000Z" end="20060105T220000Z"/>
</C:free-busy-query>`)

	assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestPinReportCalendarQuery(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		hrefs  []string
	}{
		{
			name:  "VTODO comp-filter only",
			hrefs: project36TaskHrefs(),
		},
		{
			// No fixture task falls into this range; time-range is not evaluated for VTODOs.
			name:   "time-range",
			filter: `<C:time-range start="20300101T000000Z" end="20300102T000000Z"/>`,
			hrefs:  project36TaskHrefs(),
		},
		{
			name:   "prop-filter text-match",
			filter: `<C:prop-filter name="SUMMARY"><C:text-match collation="i;ascii-casemap">Child</C:text-match></C:prop-filter>`,
			hrefs:  []string{"/dav/projects/36/uid-caldav-test-child-task.ics"},
		},
		{
			name:   "prop-filter text-match is case-insensitive",
			filter: `<C:prop-filter name="SUMMARY"><C:text-match collation="i;ascii-casemap">child TASK</C:text-match></C:prop-filter>`,
			hrefs:  []string{"/dav/projects/36/uid-caldav-test-child-task.ics"},
		},
		{
			name:   "prop-filter text-match without hits",
			filter: `<C:prop-filter name="SUMMARY"><C:text-match collation="i;ascii-casemap">no such summary</C:text-match></C:prop-filter>`,
		},
	}

	e := setupTestEnv(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := caldavREPORT(t, e, "/dav/projects/36", calendarQueryWithTodoFilter(tc.filter))

			require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
			assertDAVResponses(t, rec, etagOnly(tc.hrefs...))
		})
	}
}
