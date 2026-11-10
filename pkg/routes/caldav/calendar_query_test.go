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

package caldav

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func calendarQuery(t *testing.T, filter string) string {
	t.Helper()
	const projectID = 36

	body := `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop><D:getetag/></D:prop>
  <C:filter>` + filter + `</C:filter>
</C:calendar-query>`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("REPORT", projectPath(projectID), strings.NewReader(body))
	req.Header.Set("Depth", "1")
	c := echo.New().NewContext(req, rec)
	c.Set("userBasicAuth", caldavFilterUser)
	c.SetPathValues(echo.PathValues{{Name: "project", Value: strconv.FormatInt(projectID, 10)}})

	require.NoError(t, ProjectHandler(c))
	require.Equal(t, http.StatusMultiStatus, rec.Code, rec.Body.String())
	return rec.Body.String()
}

func uidFilter(textMatch string) string {
	return `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO">
  <C:prop-filter name="UID">` + textMatch + `</C:prop-filter>
</C:comp-filter></C:comp-filter>`
}

func TestCalendarQueryFilters(t *testing.T) {
	const otherTaskHref = "/dav/projects/36/uid-caldav-test.ics"
	const parentTaskHref = "/dav/projects/36/uid-caldav-test-parent-task-2.ics"

	t.Run("UID text-match returns only the matching task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := calendarQuery(t, uidFilter(`<C:text-match collation="i;octet">uid-caldav-test-parent-task-2</C:text-match>`))

		assert.Equal(t, 1, strings.Count(body, "<D:response>"), body)
		assert.Contains(t, body, parentTaskHref)
	})

	t.Run("UID text-match without a match returns nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := calendarQuery(t, uidFilter(`<C:text-match collation="i;octet">does-not-exist-12345</C:text-match>`))

		assert.NotContains(t, body, "<D:response>")
	})

	t.Run("plain VTODO comp-filter returns every task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := calendarQuery(t, `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO"/></C:comp-filter>`)

		assert.Contains(t, body, otherTaskHref)
		assert.Contains(t, body, parentTaskHref)
	})

	t.Run("time-range is not applied to VTODOs", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := calendarQuery(t, `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO">
  <C:time-range start="20000101T000000Z" end="20000102T000000Z"/>
</C:comp-filter></C:comp-filter>`)

		assert.Contains(t, body, otherTaskHref)
		assert.Contains(t, body, parentTaskHref)
	})
}
