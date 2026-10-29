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
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

func TestParseTodosTimeZone(t *testing.T) {
	indy := mustLoadLocation(t, "America/Indiana/Indianapolis")
	stamp := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		loc  *time.Location
		todo *Todo
		want string
	}{
		{
			name: "timed due date is written with a TZID and its zone period",
			loc:  indy,
			todo: &Todo{
				UID:       "timed",
				Summary:   "Timed",
				Timestamp: stamp,
				DueDate:   time.Date(2026, 10, 29, 17, 0, 0, 0, indy),
				Done:      true,
				Completed: time.Date(2026, 10, 28, 9, 0, 0, 0, indy),
				Created:   stamp,
				Updated:   stamp,
			},
			want: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//Vikunja//EN
BEGIN:VTIMEZONE
TZID:America/Indiana/Indianapolis
BEGIN:DAYLIGHT
DTSTART:20260308T020000
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VTODO
UID:timed
DTSTAMP:20261028T120000Z
SUMMARY:Timed
COMPLETED:20261028T130000Z
STATUS:COMPLETED
DUE;TZID=America/Indiana/Indianapolis:20261029T170000
CREATED:20261028T120000Z
LAST-MODIFIED:20261028T120000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "midnight start and due are written as dates",
			loc:  indy,
			todo: &Todo{
				UID:       "all-day",
				Summary:   "All day",
				Timestamp: stamp,
				Start:     time.Date(2026, 10, 30, 0, 0, 0, 0, indy),
				DueDate:   time.Date(2026, 10, 30, 0, 0, 0, 0, indy),
				Updated:   stamp,
			},
			want: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//Vikunja//EN
BEGIN:VTODO
UID:all-day
DTSTAMP:20261028T120000Z
SUMMARY:All day
DTSTART;VALUE=DATE:20261030
STATUS:NEEDS-ACTION
DUE;VALUE=DATE:20261030
LAST-MODIFIED:20261028T120000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "a timed due date keeps a midnight start timed",
			loc:  indy,
			todo: &Todo{
				UID:       "mixed",
				Summary:   "Mixed",
				Timestamp: stamp,
				Start:     time.Date(2026, 10, 30, 0, 0, 0, 0, indy),
				DueDate:   time.Date(2026, 10, 30, 17, 0, 0, 0, indy),
				Updated:   stamp,
			},
			want: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//Vikunja//EN
BEGIN:VTIMEZONE
TZID:America/Indiana/Indianapolis
BEGIN:DAYLIGHT
DTSTART:20260308T020000
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VTODO
UID:mixed
DTSTAMP:20261028T120000Z
SUMMARY:Mixed
DTSTART;TZID=America/Indiana/Indianapolis:20261030T000000
STATUS:NEEDS-ACTION
DUE;TZID=America/Indiana/Indianapolis:20261030T170000
LAST-MODIFIED:20261028T120000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "an all-day span across the DST change stays all-day",
			loc:  indy,
			todo: &Todo{
				UID:       "span",
				Summary:   "Span",
				Timestamp: stamp,
				Start:     time.Date(2026, 10, 31, 0, 0, 0, 0, indy),
				End:       time.Date(2026, 11, 2, 0, 0, 0, 0, indy),
				Updated:   stamp,
			},
			want: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//Vikunja//EN
BEGIN:VTODO
UID:span
DTSTAMP:20261028T120000Z
SUMMARY:Span
DTSTART;VALUE=DATE:20261031
DTEND;VALUE=DATE:20261102
STATUS:NEEDS-ACTION
LAST-MODIFIED:20261028T120000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "monthly repeat takes the day of month from the written zone",
			loc:  indy,
			todo: &Todo{
				UID:        "monthly",
				Summary:    "Monthly",
				Timestamp:  stamp,
				DueDate:    time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC), // Oct 31, 22:00 EDT
				RepeatMode: models.TaskRepeatModeMonth,
				Updated:    stamp,
			},
			want: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//Vikunja//EN
BEGIN:VTIMEZONE
TZID:America/Indiana/Indianapolis
BEGIN:DAYLIGHT
DTSTART:20260308T020000
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VTODO
UID:monthly
DTSTAMP:20261028T120000Z
SUMMARY:Monthly
STATUS:NEEDS-ACTION
DUE;TZID=America/Indiana/Indianapolis:20261031T220000
RRULE:FREQ=MONTHLY;BYMONTHDAY=31
LAST-MODIFIED:20261028T120000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "UTC under another name keeps the Z form",
			loc:  mustLoadLocation(t, "Etc/GMT"),
			todo: &Todo{
				UID:       "gmt",
				Summary:   "GMT",
				Timestamp: stamp,
				DueDate:   time.Date(2026, 10, 29, 17, 0, 0, 0, time.UTC),
				Updated:   stamp,
			},
			want: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//Vikunja//EN
BEGIN:VTODO
UID:gmt
DTSTAMP:20261028T120000Z
SUMMARY:GMT
STATUS:NEEDS-ACTION
DUE:20261029T170000Z
LAST-MODIFIED:20261028T120000Z
END:VTODO
END:VCALENDAR`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTodos(&Config{
				Name:     "test",
				ProdID:   "Vikunja",
				TimeZone: tt.loc,
			}, []*Todo{tt.todo})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTodosTimeZoneOneObservancePerPeriod(t *testing.T) {
	indy := mustLoadLocation(t, "America/Indiana/Indianapolis")
	todo := func(uid string, due time.Time) *Todo {
		return &Todo{
			UID:     uid,
			Summary: uid,
			DueDate: due,
		}
	}

	got := ParseTodos(&Config{
		Name:     "test",
		ProdID:   "Vikunja",
		TimeZone: indy,
	}, []*Todo{
		todo("december", time.Date(2026, 12, 1, 9, 0, 0, 0, indy)),
		todo("october", time.Date(2026, 10, 29, 9, 0, 0, 0, indy)),
		todo("november", time.Date(2026, 11, 20, 9, 0, 0, 0, indy)),
	})

	assert.Equal(t, 1, strings.Count(got, "BEGIN:VTIMEZONE"))
	assert.Contains(t, got, `BEGIN:VTIMEZONE
TZID:America/Indiana/Indianapolis
BEGIN:DAYLIGHT
DTSTART:20260308T020000
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
END:DAYLIGHT
BEGIN:STANDARD
DTSTART:20261101T020000
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
END:STANDARD
END:VTIMEZONE
BEGIN:VTODO`)
}

// resolveOffset applies RFC 5545 §3.6.5 to observances without RRULE: the one
// with the latest onset at or before the local time decides the offset.
func resolveOffset(t *testing.T, vtimezone, local string) string {
	t.Helper()
	var bestOnset, bestOffset, onset string
	for _, line := range strings.Split(vtimezone, "\n") {
		switch {
		case strings.HasPrefix(line, "DTSTART:"):
			onset = strings.TrimPrefix(line, "DTSTART:")
		case strings.HasPrefix(line, "TZOFFSETTO:"):
			if onset <= local && onset > bestOnset {
				bestOnset = onset
				bestOffset = strings.TrimPrefix(line, "TZOFFSETTO:")
			}
		}
	}
	require.NotEmpty(t, bestOffset, "no observance applies to %s in\n%s", local, vtimezone)
	return bestOffset
}

func TestVTimezoneResolvesWrittenTimes(t *testing.T) {
	zones := map[string][]time.Time{
		"America/Indiana/Indianapolis": {
			time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC),
			time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC),
			time.Date(2026, 11, 1, 5, 30, 0, 0, time.UTC), // 01:30 EDT, the first of the repeated hour
		},
		"Europe/Berlin": {
			time.Date(2026, 3, 29, 1, 30, 0, 0, time.UTC), // 03:30 CEST, right after spring forward
			time.Date(2026, 12, 24, 18, 0, 0, 0, time.UTC),
		},
		"Asia/Kolkata": {
			time.Date(2026, 6, 1, 4, 0, 0, 0, time.UTC),
		},
		"Australia/Sydney": {
			time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		},
		"Australia/Lord_Howe": {
			time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		},
		"Etc/GMT+5": {
			time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	for name, instants := range zones {
		t.Run(name, func(t *testing.T) {
			zone := newCalendarZone(mustLoadLocation(t, name))
			var locals []string
			for _, instant := range instants {
				prop := zone.dateProperty("DUE", instant, false)
				require.True(t, strings.HasPrefix(prop, "DUE;TZID="+name+":"), prop)
				locals = append(locals, strings.TrimPrefix(prop, "DUE;TZID="+name+":"))
			}

			vtz := zone.vtimezone()
			for i, instant := range instants {
				_, offset := instant.In(zone.loc).Zone()
				assert.Equal(t, formatUTCOffset(offset), resolveOffset(t, vtz, locals[i]), "%s at %s", name, locals[i])
			}
		})
	}
}

func TestDatePropertyRepeatedHour(t *testing.T) {
	zone := newCalendarZone(mustLoadLocation(t, "America/Indiana/Indianapolis"))

	// 01:30 EST on Nov 1 shares its wall time with 01:30 EDT an hour earlier.
	assert.Equal(t, "DUE:20261101T063000Z", zone.dateProperty("DUE", time.Date(2026, 11, 1, 6, 30, 0, 0, time.UTC), false))
	assert.Equal(t, "DUE;TZID=America/Indiana/Indianapolis:20261101T013000", zone.dateProperty("DUE", time.Date(2026, 11, 1, 5, 30, 0, 0, time.UTC), false))
	assert.Equal(t, "DUE;TZID=America/Indiana/Indianapolis:20261101T020000", zone.dateProperty("DUE", time.Date(2026, 11, 1, 7, 0, 0, 0, time.UTC), false))
}

func TestFormatUTCOffset(t *testing.T) {
	tests := map[int]string{
		0:                      "+0000",
		-5 * 3600:              "-0500",
		5*3600 + 30*60:         "+0530",
		10*3600 + 30*60:        "+1030",
		-(5*3600 + 45*60):      "-0545",
		-(5*3600 + 50*60 + 36): "-055036",
	}
	for seconds, want := range tests {
		t.Run(strconv.Itoa(seconds), func(t *testing.T) {
			assert.Equal(t, want, formatUTCOffset(seconds))
		})
	}
}

func TestNewCalendarZone(t *testing.T) {
	for _, name := range []string{"UTC", "GMT", "Etc/UTC", "Zulu"} {
		assert.True(t, newCalendarZone(mustLoadLocation(t, name)).utc, name)
	}
	for _, name := range []string{"Europe/London", "Etc/GMT+5", "America/Indiana/Indianapolis"} {
		assert.False(t, newCalendarZone(mustLoadLocation(t, name)).utc, name)
	}
	assert.True(t, newCalendarZone(nil).utc)
	assert.True(t, newCalendarZone(mustLoadLocation(t, "Local")).utc)
}

func TestTimeZoneRoundTrip(t *testing.T) {
	indy := mustLoadLocation(t, "America/Indiana/Indianapolis")
	tasks := []*models.Task{
		{
			UID:     "timed-summer",
			Title:   "timed-summer",
			DueDate: time.Date(2026, 10, 29, 17, 0, 0, 0, indy),
		},
		{
			UID:     "timed-winter",
			Title:   "timed-winter",
			DueDate: time.Date(2026, 11, 3, 7, 0, 0, 0, indy),
		},
		{
			UID:       "all-day",
			Title:     "all-day",
			StartDate: time.Date(2026, 10, 30, 0, 0, 0, 0, indy),
			DueDate:   time.Date(2026, 10, 30, 0, 0, 0, 0, indy),
		},
	}

	for _, task := range tasks {
		t.Run(task.UID, func(t *testing.T) {
			exported := GetCaldavTodosForTasks(
				&models.ProjectWithTasksAndBuckets{},
				[]*models.TaskWithComments{{Task: *task}},
				indy,
			)
			parsed, _, err := ParseTaskFromVTODO(exported, indy)
			require.NoError(t, err)
			assert.True(t, task.DueDate.Equal(parsed.DueDate), "due: want %s, got %s\n%s", task.DueDate, parsed.DueDate, exported)
			assert.True(t, task.StartDate.Equal(parsed.StartDate), "start: want %s, got %s\n%s", task.StartDate, parsed.StartDate, exported)
		})
	}
}

func TestParseTaskFromVTODOReadsLocalTimesInZone(t *testing.T) {
	indy := mustLoadLocation(t, "America/Indiana/Indianapolis")
	task, _, err := ParseTaskFromVTODO(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//EN
BEGIN:VTODO
UID:local-times
SUMMARY:Local times
DTSTART;VALUE=DATE:20261030
DUE:20261030T150000
END:VTODO
END:VCALENDAR`, indy)
	require.NoError(t, err)

	assert.True(t, task.StartDate.Equal(time.Date(2026, 10, 30, 0, 0, 0, 0, indy)), "start: %s", task.StartDate)
	assert.True(t, task.DueDate.Equal(time.Date(2026, 10, 30, 15, 0, 0, 0, indy)), "due: %s", task.DueDate)
}
