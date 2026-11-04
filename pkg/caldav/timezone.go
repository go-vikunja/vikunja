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
	"fmt"
	"sort"
	"time"
)

// calendarZone writes DTSTART, DUE and DTEND in one zone and collects the
// VTIMEZONE observances they need: RFC 5545 §3.2.19 requires one per TZID.
type calendarZone struct {
	loc *time.Location
	utc bool
	// Start of each zone period a written value falls in, to an instant inside it.
	periods map[int64]time.Time
}

func newCalendarZone(loc *time.Location) *calendarZone {
	// "Local" is no IANA name a client could resolve.
	if loc == nil || loc.String() == "Local" {
		loc = time.UTC
	}
	return &calendarZone{
		loc:     loc,
		utc:     isFixedUTC(loc),
		periods: make(map[int64]time.Time),
	}
}

// isFixedUTC is true for UTC under any name (GMT, Etc/UTC, ...), which keeps the Z form.
func isFixedUTC(loc *time.Location) bool {
	now := time.Now().In(loc)
	_, offset := now.Zone()
	start, end := now.ZoneBounds()
	return offset == 0 && start.IsZero() && end.IsZero()
}

// isAllDay is true when every set date is at midnight in the zone. Vikunja has no
// all-day flag and stores a client's date-only DUE as midnight; writing it back as
// VALUE=DATE keeps it all-day. All or nothing, as DTSTART and DUE must share a
// value type (RFC 5545 §3.8.2.3).
func (z *calendarZone) isAllDay(dates ...time.Time) bool {
	var hasDate bool
	for _, d := range dates {
		if d.Unix() <= 0 {
			continue
		}
		hasDate = true
		local := d.In(z.loc)
		if local.Hour() != 0 || local.Minute() != 0 || local.Second() != 0 || local.Nanosecond() != 0 {
			return false
		}
	}
	return hasDate
}

func (z *calendarZone) dateProperty(name string, ts time.Time, allDay bool) string {
	if allDay {
		return name + ";VALUE=DATE:" + ts.In(z.loc).Format(`20060102`)
	}
	if z.utc || isRepeatedLocalTime(ts.In(z.loc)) {
		return name + ":" + makeCalDavTimeFromTimeStamp(ts)
	}

	start, _ := ts.In(z.loc).ZoneBounds()
	if _, seen := z.periods[start.Unix()]; !seen {
		z.periods[start.Unix()] = ts
	}
	return name + ";TZID=" + z.loc.String() + ":" + ts.In(z.loc).Format(DateFormat)
}

// isRepeatedLocalTime is true in the hour repeated after clocks fall back. A local
// time there means its first occurrence (RFC 5545 §3.3.5), so t needs UTC instead.
func isRepeatedLocalTime(t time.Time) bool {
	start, _ := t.ZoneBounds()
	if start.IsZero() {
		return false
	}
	_, offset := t.Zone()
	_, before := start.Add(-time.Nanosecond).Zone()
	return before > offset && t.Before(start.Add(time.Duration(before-offset)*time.Second))
}

// vtimezone holds one observance per zone period a written value falls in. That
// resolves those values, since a client applies the latest onset before a local time.
func (z *calendarZone) vtimezone() string {
	if len(z.periods) == 0 {
		return ""
	}

	starts := make([]int64, 0, len(z.periods))
	for start := range z.periods {
		starts = append(starts, start)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i] < starts[j] })

	vtz := `
BEGIN:VTIMEZONE
TZID:` + z.loc.String()
	for _, s := range starts {
		vtz += observance(z.periods[s].In(z.loc))
	}
	return vtz + `
END:VTIMEZONE`
}

func observance(t time.Time) string {
	name, offset := t.Zone()
	start, _ := t.ZoneBounds()

	// A period with no known start has always applied, so its onset predates any task.
	offsetFrom := offset
	onset := `16010101T000000`
	if !start.IsZero() {
		_, offsetFrom = start.Add(-time.Nanosecond).Zone()
		// An observance's DTSTART is local time in the offset it replaces.
		onset = start.In(time.FixedZone("", offsetFrom)).Format(DateFormat)
	}

	kind := "STANDARD"
	if t.IsDST() {
		kind = "DAYLIGHT"
	}

	return `
BEGIN:` + kind + `
DTSTART:` + onset + `
TZOFFSETFROM:` + formatUTCOffset(offsetFrom) + `
TZOFFSETTO:` + formatUTCOffset(offset) + `
TZNAME:` + escapeICalText(name) + `
END:` + kind
}

// formatUTCOffset formats seconds east of UTC as an RFC 5545 §3.3.14 UTC-OFFSET.
func formatUTCOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	out := fmt.Sprintf("%s%02d%02d", sign, seconds/3600, seconds%3600/60)
	if seconds%60 != 0 {
		out += fmt.Sprintf("%02d", seconds%60)
	}
	return out
}
