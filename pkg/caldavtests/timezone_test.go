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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Captured from iOS Reminders 27.2 with the phone in America/Indiana/Indianapolis.
// iOS creates a reminder with an empty PUT and sends its dates in a second one.
var (
	iosDateOnlyVTodo = crlf(`BEGIN:VCALENDAR
CALSCALE:GREGORIAN
PRODID:-//Apple Inc.//iOS 27.2//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20261030T012750Z
DTSTAMP:20261030T012755Z
DTSTART;VALUE=DATE:20261031
DUE;VALUE=DATE:20261031
LAST-MODIFIED:20261030T012754Z
STATUS:NEEDS-ACTION
SUMMARY:Q3
UID:6587DBFA-D0BD-4FC0-93F8-E2283D63AC8C
END:VTODO
END:VCALENDAR
`)
	iosTimedVTodo = crlf(`BEGIN:VCALENDAR
CALSCALE:GREGORIAN
PRODID:-//Apple Inc.//iOS 27.2//EN
VERSION:2.0
BEGIN:VTIMEZONE
TZID:America/Indiana/Indianapolis
BEGIN:DAYLIGHT
DTSTART:20070311T020000
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU
TZNAME:EDT
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
END:DAYLIGHT
BEGIN:STANDARD
DTSTART:20071104T020000
RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU
TZNAME:EST
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
END:STANDARD
END:VTIMEZONE
BEGIN:VTODO
CREATED:20261030T012804Z
DTSTAMP:20261030T012816Z
DTSTART;TZID=America/Indiana/Indianapolis:20261031T090000
DUE;TZID=America/Indiana/Indianapolis:20261031T090000
LAST-MODIFIED:20261030T012815Z
STATUS:NEEDS-ACTION
SUMMARY:Q4
UID:4696EA29-6858-42B2-9DA8-4A2BEACE4CC9
BEGIN:VALARM
ACTION:DISPLAY
DESCRIPTION:Reminder
TRIGGER;VALUE=DATE-TIME:20261031T130000Z
UID:9C3A1A7D-56E8-4154-B5D2-294F8B9E95BF
X-WR-ALARMUID:9C3A1A7D-56E8-4154-B5D2-294F8B9E95BF
END:VALARM
END:VTODO
END:VCALENDAR
`)
)

func crlf(s string) string {
	return strings.ReplaceAll(s, "\n", "\r\n")
}

func setUserTimeZone(t *testing.T, tz string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("id = ?", testuser15.ID).Cols("timezone").Update(&user.User{Timezone: tz})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// putAndGetIOS stores a captured iOS reminder and returns what a client reads back.
func putAndGetIOS(t *testing.T, vtodo, uid string) string {
	t.Helper()
	e := setupTestEnv(t)
	setUserTimeZone(t, "America/Indiana/Indianapolis")

	path := "/dav/projects/36/" + uid + ".ics"
	rec := caldavPUT(t, e, path, vtodo)
	require.Equal(t, http.StatusCreated, rec.Code, "PUT failed. Body:\n%s", rec.Body.String())

	rec = caldavGET(t, e, path)
	require.Equal(t, http.StatusOK, rec.Code, "GET failed. Body:\n%s", rec.Body.String())
	parseICalFromResponse(t, rec)
	return rec.Body.String()
}

func TestTimeZone(t *testing.T) {
	t.Run("iOS date-only reminder stays date-only", func(t *testing.T) {
		body := putAndGetIOS(t, iosDateOnlyVTodo, "6587DBFA-D0BD-4FC0-93F8-E2283D63AC8C")

		assert.Contains(t, body, "DTSTART;VALUE=DATE:20261031")
		assert.Contains(t, body, "DUE;VALUE=DATE:20261031")
		assert.NotContains(t, body, "BEGIN:VTIMEZONE")
	})

	t.Run("iOS timed reminder comes back in the user's time zone", func(t *testing.T) {
		body := putAndGetIOS(t, iosTimedVTodo, "4696EA29-6858-42B2-9DA8-4A2BEACE4CC9")

		assert.Contains(t, body, "DTSTART;TZID=America/Indiana/Indianapolis:20261031T090000")
		assert.Contains(t, body, "DUE;TZID=America/Indiana/Indianapolis:20261031T090000")
		assert.Contains(t, body, "BEGIN:VTIMEZONE")
		assert.Contains(t, body, "TZID:America/Indiana/Indianapolis")
		assert.Contains(t, body, "TRIGGER;VALUE=DATE-TIME:20261031T130000Z")
	})

	t.Run("without a user time zone the server's applies", func(t *testing.T) {
		e := setupTestEnv(t)

		path := "/dav/projects/36/4696EA29-6858-42B2-9DA8-4A2BEACE4CC9.ics"
		rec := caldavPUT(t, e, path, iosTimedVTodo)
		require.Equal(t, http.StatusCreated, rec.Code, "PUT failed. Body:\n%s", rec.Body.String())

		rec = caldavGET(t, e, path)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "DUE:20261031T130000Z")
		assert.NotContains(t, rec.Body.String(), "BEGIN:VTIMEZONE")
	})
}
