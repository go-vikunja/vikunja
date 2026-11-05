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

package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin what completion stores where the rrule library's own DST and duplicate-value
// handling differs between versions (rrule-go v1.8.3 and v1.9.0 both measured against these
// cases). The engine steps in floating wall-clock time and localises the
// result itself, so a library upgrade must leave every one of these unchanged.
func TestRRuleLibraryPin_SubDailyAndDuplicateValues(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	lordHowe := mustZone(t, "Australia/Lord_Howe")
	cases := []struct {
		name   string
		zone   *time.Location
		rule   string
		due    string
		n      int
		want   string // successive due dates, RFC 3339 in zone
		reason string
	}{
		{"New York gap, hourly", ny, "FREQ=HOURLY", "2027-03-14 00:30 -05:00", 4,
			"2027-03-14T01:30:00-05:00 2027-03-14T03:30:00-04:00 2027-03-14T04:30:00-04:00 2027-03-14T05:30:00-04:00", ""},
		{"New York gap, every 30 minutes", ny, "FREQ=MINUTELY;INTERVAL=30", "2027-03-14 01:00 -05:00", 5,
			"2027-03-14T01:30:00-05:00 2027-03-14T03:00:00-04:00 2027-03-14T03:30:00-04:00 2027-03-14T04:00:00-04:00 2027-03-14T04:30:00-04:00", ""},
		{"New York gap, BYHOUR skips 02:00 on the change day", ny, "FREQ=DAILY;BYHOUR=1,2,3", "2027-03-13 01:00 -05:00", 6,
			"2027-03-13T02:00:00-05:00 2027-03-13T03:00:00-05:00 2027-03-14T01:00:00-05:00 2027-03-14T03:00:00-04:00 2027-03-15T01:00:00-04:00 2027-03-15T02:00:00-04:00", ""},
		{"New York gap, weekly COUNT skips the gap day", ny, "FREQ=WEEKLY;COUNT=3", "2027-03-07 02:30 -05:00", 4,
			"2027-03-21T02:30:00-04:00 2027-03-28T02:30:00-04:00", "count_exhausted"},
		{"New York fold, hourly COUNT", ny, "FREQ=HOURLY;COUNT=5", "2027-11-07 00:30 -04:00", 6,
			"2027-11-07T01:30:00-04:00 2027-11-07T01:30:00-05:00 2027-11-07T02:30:00-05:00 2027-11-07T03:30:00-05:00", "count_exhausted"},
		{"New York fold, every 30 minutes", ny, "FREQ=MINUTELY;INTERVAL=30", "2027-11-07 00:30 -04:00", 6,
			"2027-11-07T01:00:00-04:00 2027-11-07T01:30:00-04:00 2027-11-07T01:00:00-05:00 2027-11-07T01:30:00-05:00 2027-11-07T02:00:00-05:00 2027-11-07T02:30:00-05:00", ""},
		{"New York fold, BYHOUR", ny, "FREQ=DAILY;BYHOUR=0,1,2", "2027-11-06 00:00 -04:00", 6,
			"2027-11-06T01:00:00-04:00 2027-11-06T02:00:00-04:00 2027-11-07T00:00:00-04:00 2027-11-07T01:00:00-04:00 2027-11-07T02:00:00-05:00 2027-11-08T00:00:00-05:00", ""},
		{"Lord Howe 30-minute gap, hourly", lordHowe, "FREQ=HOURLY", "2026-10-04 01:00 +10:30", 4,
			"2026-10-04T02:30:00+11:00 2026-10-04T03:30:00+11:00 2026-10-04T04:30:00+11:00 2026-10-04T05:30:00+11:00", ""},
		{"Lord Howe 30-minute gap, every 15 minutes", lordHowe, "FREQ=MINUTELY;INTERVAL=15", "2026-10-04 01:30 +10:30", 6,
			"2026-10-04T01:45:00+10:30 2026-10-04T02:30:00+11:00 2026-10-04T02:45:00+11:00 2026-10-04T03:00:00+11:00 2026-10-04T03:15:00+11:00 2026-10-04T03:30:00+11:00", ""},
		{"Lord Howe 30-minute fold, every 15 minutes", lordHowe, "FREQ=MINUTELY;INTERVAL=15", "2027-04-04 01:15 +11:00", 6,
			"2027-04-04T01:30:00+11:00 2027-04-04T01:45:00+11:00 2027-04-04T01:30:00+10:30 2027-04-04T01:45:00+10:30 2027-04-04T02:00:00+10:30 2027-04-04T02:15:00+10:30", ""},
		{"Lord Howe 30-minute fold, hourly", lordHowe, "FREQ=HOURLY", "2027-04-04 00:30 +11:00", 4,
			"2027-04-04T01:30:00+11:00 2027-04-04T02:00:00+10:30 2027-04-04T03:00:00+10:30 2027-04-04T04:00:00+10:30", ""},
		{"a stored duplicate BYHOUR keeps its COUNT", time.UTC, "FREQ=DAILY;BYHOUR=9,9;COUNT=3", "2026-01-01 08:00 +00:00", 5,
			"2026-01-01T09:00:00Z 2026-01-02T09:00:00Z", "count_exhausted"},
		{"a stored duplicate BYMINUTE is refused by the policy", time.UTC, "FREQ=HOURLY;BYMINUTE=0,0,30;COUNT=4", "2026-01-01 08:00 +00:00", 5,
			"", "policy_rejected"},
		{"BYSETPOS on the gap day skips the month (known limit)", ny, "FREQ=MONTHLY;BYDAY=SU;BYSETPOS=2", "2027-02-14 02:30 -05:00", 3,
			"2027-04-11T02:30:00-04:00 2027-05-09T02:30:00-04:00 2027-06-13T02:30:00-04:00", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			due := ts(t, c.zone, c.due)
			ans := stepOccurrences(nextOccurrenceInput{Rule: c.rule, Due: dated(due), Now: due.Add(time.Minute), Zone: c.zone}, c.n)
			got := make([]string, 0, len(ans.Occurrences))
			for _, o := range ans.Occurrences {
				got = append(got, o.In(c.zone).Format(time.RFC3339))
			}
			assert.Equal(t, c.want, strings.Join(got, " "))
			assert.Equal(t, c.reason, ans.Reason)
		})
	}
}

// NormalizeRRule on rule text without a DTSTART: unchanged by the v1.9.0 candidate.
func TestRRuleLibraryPin_NormalizeRRuleText(t *testing.T) {
	cases := []struct {
		raw, want string
		ok        bool
	}{
		{"FREQ=DAILY;UNTIL=20260309T020000", "FREQ=DAILY;UNTIL=20260309T020000Z", true}, // floating UNTIL read as UTC
		{"FREQ=DAILY;UNTIL=20260309", "FREQ=DAILY;UNTIL=20260309T000000Z", true},
		{"RRULE:FREQ=WEEKLY;BYDAY=MO", "FREQ=WEEKLY;BYDAY=MO", true},
		{"FREQ=MONTHLY;BYDAY=4FR", "FREQ=MONTHLY;BYDAY=+4FR", true},
		{"FREQ=DAILY;BYHOUR=9,9", "", false},
	}
	for _, c := range cases {
		got, ok := NormalizeRRule(c.raw)
		assert.Equal(t, c.ok, ok, "%q", c.raw)
		assert.Equal(t, c.want, got, "%q", c.raw)
	}
}

// Rule text that carries a DTSTART, inline (DTSTART=) or as its own line, reaches the parser
// through stored older data, CalDAV import (the RRULE value may carry an inline DTSTART) and the
// import gate. This table is rrule-go v1.8.3's behaviour, which Vikunja relies on. On the v1.9.0
// parser alone 7 of the 8 rows fail (measured 2026-10-29: refused pairs, UNTIL's Z dropped);
// parseRuleText keeps all 8 (TestParseRuleText_LibraryIsStricter shows it is the reason).
func TestRRuleLibraryCompat_DTSTARTRules(t *testing.T) {
	cases := []struct {
		rule, normalized string
		mismatchedPair   bool // v1.9.0 refuses it
		floatingOrDate   bool // v1.9.0 drops UNTIL's Z
	}{
		{"FREQ=DAILY;DTSTART=20260301T090000;UNTIL=20260310T090000Z", "FREQ=DAILY;UNTIL=20260310T090000Z", true, false},
		{"FREQ=DAILY;DTSTART=20260301T090000;UNTIL=20260310T090000", "FREQ=DAILY;UNTIL=20260310T090000Z", false, true},
		{"FREQ=DAILY;DTSTART=20260301T090000Z;UNTIL=20260310T090000", "FREQ=DAILY;UNTIL=20260310T090000Z", true, false},
		{"FREQ=DAILY;DTSTART=20260301T090000Z;UNTIL=20260310T090000Z", "FREQ=DAILY;UNTIL=20260310T090000Z", false, false},
		{"FREQ=DAILY;DTSTART=20260301;UNTIL=20260310", "FREQ=DAILY;UNTIL=20260310T000000Z", false, true},
		{"FREQ=DAILY;DTSTART=20260301;UNTIL=20260310T090000Z", "FREQ=DAILY;UNTIL=20260310T090000Z", true, false},
		{"DTSTART:20260301T090000\nRRULE:FREQ=DAILY;UNTIL=20260310T090000Z", "FREQ=DAILY;UNTIL=20260310T090000Z", true, false},
		{"DTSTART:20260301T090000\nRRULE:FREQ=DAILY;UNTIL=20260310T090000", "FREQ=DAILY;UNTIL=20260310T090000Z", false, true},
	}
	due := time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)
	for _, c := range cases {
		t.Run(c.rule, func(t *testing.T) {
			got, ok := NormalizeRRule(c.rule)
			assert.True(t, ok, "normalises")
			assert.Equal(t, c.normalized, got)
			assert.NotNil(t, taskRepeatFromRRule(c.rule), "has a structured projection")
			ans := stepOccurrences(nextOccurrenceInput{Rule: c.rule, Due: dated(due), Now: due.Add(time.Minute), Zone: time.UTC}, 3)
			assert.Empty(t, ans.Reason, "completion advances")
			assert.Len(t, ans.Occurrences, 3)
			imported := &Task{Repeats: c.rule}
			require.NoError(t, DropUnimportableRepeat(imported), "the import keeps it")
			assert.Equal(t, c.rule, imported.Repeats)
		})
	}
}
