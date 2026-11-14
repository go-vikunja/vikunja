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
	"net/http/httptest"
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
		out = append(out, expectedDAVResponse{href: href, ok: map[xml.Name]string{davName("getetag"): anyValue}})
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
	// Exists and is readable, but lives outside the requested collection.
	const outside = "/dav/projects/38/uid-caldav-test-child-task-2.ics"

	rec := caldavREPORT(t, e, "/dav/projects/36", ReportCalendarMultiget(pinTaskHref, missing, outside))

	require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
	assertDAVResponses(t, rec, []expectedDAVResponse{
		{
			href: pinTaskHref,
			ok: map[xml.Name]string{
				davName("getetag"):       anyValue,
				calName("calendar-data"): anyValue,
			},
		},
		{href: missing, status: http.StatusNotFound},
		{href: outside, status: http.StatusNotFound},
	})
}

func TestPinReportUnknownType(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavREPORT(t, e, "/dav/projects/36", `<?xml version="1.0" encoding="utf-8" ?>
<C:free-busy-query xmlns:C="urn:ietf:params:xml:ns:caldav">
  <C:time-range start="20060104T140000Z" end="20060105T220000Z"/>
</C:free-busy-query>`)

	assertSupportedReport(t, rec)
}

// RFC 3253 §3.6: a report the resource does not support fails the supported-report precondition.
func assertSupportedReport(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusForbidden, rec.Code, "Body:\n%s", rec.Body.String())
	var body struct {
		XMLName xml.Name  `xml:"DAV: error"`
		Failed  []xmlNode `xml:",any"`
	}
	require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &body), "Body:\n%s", rec.Body.String())
	require.Len(t, body.Failed, 1)
	assert.Equal(t, davName("supported-report"), body.Failed[0].XMLName)
}

func TestPinReportMalformedBody(t *testing.T) {
	e := setupTestEnv(t)

	for name, body := range map[string]string{"empty": "", "not XML": "<<<", "truncated": `<C:calendar-query xmlns:C="urn:ietf:params:xml:ns:caldav">`} {
		t.Run(name, func(t *testing.T) {
			rec := caldavREPORT(t, e, "/dav/projects/36", body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "Body:\n%s", rec.Body.String())

			// The body is only looked at once access is granted.
			rec = caldavRequest(t, e, "REPORT", "/dav/projects/36", body, map[string]string{"Authorization": basicAuthHeader(testuser1.Username, fixturePassword)})
			assert.Equal(t, http.StatusNotFound, rec.Code, "Body:\n%s", rec.Body.String())
		})
	}
}

func TestPinReportCalendarQuery(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		hrefs  []string
	}{
		{name: "VTODO comp-filter only", hrefs: project36TaskHrefs()},
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

const (
	reportSyncCollection = `<?xml version="1.0" encoding="utf-8" ?>
<D:sync-collection xmlns:D="DAV:">
  <D:sync-token/>
  <D:sync-level>1</D:sync-level>
  <D:prop><D:getetag/></D:prop>
</D:sync-collection>`
	reportFreeBusy = `<?xml version="1.0" encoding="utf-8" ?>
<C:free-busy-query xmlns:C="urn:ietf:params:xml:ns:caldav">
  <C:time-range start="20060104T140000Z" end="20060105T220000Z"/>
</C:free-busy-query>`
)

func TestPinReportTargets(t *testing.T) {
	task := []expectedDAVResponse{{
		href: pinTaskHref,
		ok:   map[xml.Name]string{davName("getetag"): anyValue, calName("calendar-data"): anyValue},
	}}
	// They hold no calendar data, so they take no REPORT at all.
	const noReports = http.StatusMethodNotAllowed

	tests := []struct {
		name string
		path string
		// nil: the status alone is pinned.
		query, multiget             []expectedDAVResponse
		queryStatus, multigetStatus int
		syncStatus, unknownStatus   int
		// The Allow header of a 405.
		allow string
	}{
		{
			name: "entry", path: "/dav/", allow: allowEntry,
			queryStatus: noReports, multigetStatus: noReports,
			syncStatus: noReports, unknownStatus: noReports,
		},
		{
			name: "principal", path: user15Principal, allow: allowEntry,
			queryStatus: noReports, multigetStatus: noReports,
			syncStatus: noReports, unknownStatus: noReports,
		},
		{
			name: "well-known", path: "/.well-known/caldav", allow: allowEntry,
			queryStatus: noReports, multigetStatus: noReports,
			syncStatus: noReports, unknownStatus: noReports,
		},
		{
			name: "home set", path: "/dav/projects", allow: allowHomeSet,
			queryStatus: noReports, multigetStatus: noReports,
			syncStatus: noReports, unknownStatus: noReports,
		},
		{
			// The filter and the hrefs are ignored, the task answers for itself.
			name: "task", path: pinTaskHref,
			query: task, multiget: task, queryStatus: 207, multigetStatus: 207,
			syncStatus: http.StatusForbidden, unknownStatus: http.StatusForbidden,
		},
		{
			name: "missing task", path: "/dav/projects/36/nonexistent-uid.ics",
			queryStatus: http.StatusNotFound, multigetStatus: http.StatusNotFound,
			syncStatus: http.StatusNotFound, unknownStatus: http.StatusNotFound,
		},
	}

	e := setupTestEnv(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			check := func(t *testing.T, body string, status int, want []expectedDAVResponse) {
				t.Helper()
				rec := caldavREPORT(t, e, tc.path, body)
				if status == http.StatusForbidden {
					assertSupportedReport(t, rec)
					return
				}
				require.Equal(t, status, rec.Code, "Body:\n%s", rec.Body.String())
				if status == http.StatusMethodNotAllowed {
					assert.Equal(t, tc.allow, rec.Header().Get("Allow"))
				}
				if want == nil {
					return
				}
				assertDAVResponses(t, rec, want)
			}

			t.Run("calendar-query", func(t *testing.T) { check(t, ReportCalendarQuery, tc.queryStatus, tc.query) })
			t.Run("calendar-multiget", func(t *testing.T) {
				check(t, ReportCalendarMultiget(pinTaskHref, "/dav/projects/38/uid-caldav-test-child-task-2.ics"), tc.multigetStatus, tc.multiget)
			})
			t.Run("sync-collection", func(t *testing.T) { check(t, reportSyncCollection, tc.syncStatus, nil) })
			t.Run("unknown", func(t *testing.T) { check(t, reportFreeBusy, tc.unknownStatus, nil) })
		})
	}
}

func TestPinReportCalendarQueryWithoutFilter(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavREPORT(t, e, "/dav/projects/36", `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-query xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:D="DAV:">
  <D:prop><D:getetag/></D:prop>
</C:calendar-query>`)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Body:\n%s", rec.Body.String())
}

// A collection without tasks answers a calendar-query with itself.
func TestPinReportCalendarQueryOnEmptyCollection(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavREPORT(t, e, "/dav/projects/-1", ReportCalendarQuery)

	require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
	assertDAVResponses(t, rec, []expectedDAVResponse{{
		href:     "/dav/projects/-1",
		ok:       map[xml.Name]string{davName("getetag"): anyValue},
		notFound: []xml.Name{calName("calendar-data")},
	}})
}
