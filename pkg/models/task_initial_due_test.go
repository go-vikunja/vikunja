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
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initialDueRow is one first-occurrence selection: rule + initial_due_time + zone + now, the due
// date it must freeze (or the reason there is none), and optionally one engine step from that due
// date with the stored rule, completed on time.
type initialDueRow struct {
	name         string
	zone         string
	now          string // ts layout, read in zone
	hour, minute int
	rule         string // the stored rule; its COUNT is not part of the search
	want         string // ts layout; "" = no date
	wantReason   noNextReason
	then         string       // the next due date after one engine step from want; "" = not asserted
	thenReason   noNextReason // the reason that step stops with, when it stops
}

func runInitialDueRows(t *testing.T, rows []initialDueRow) {
	t.Helper()
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			zone := mustZone(t, row.zone)
			now := ts(t, zone, row.now)

			got, reason := selectInitialDue(row.rule, row.hour, row.minute, zone, now)

			if row.want == "" {
				assert.True(t, got.IsZero(), "no date expected, got %s", got.UTC())
				assert.Equal(t, row.wantReason, reason)
				return
			}
			want := ts(t, zone, row.want)
			assert.Empty(t, reason)
			if !assert.True(t, want.Equal(got), "want %s, got %s", want.UTC(), got.UTC()) {
				return
			}
			assert.True(t, got.After(now), "a selected due date is strictly after now")

			if row.then == "" && row.thenReason == "" {
				return
			}
			// One completion on the due date with the stored rule: COUNT counts the selected date as
			// instance 1.
			res := nextOccurrence(nextOccurrenceInput{Rule: row.rule, Due: dated(got), Now: got, Zone: zone})
			if row.thenReason != "" {
				assert.Equal(t, row.thenReason, res.Reason)
				return
			}
			assert.Empty(t, res.Reason)
			then := ts(t, zone, row.then)
			assert.True(t, then.Equal(res.Due.At), "then: want %s, got %s", then.UTC(), res.Due.At.UTC())
		})
	}
}

// TestInitialDueGapSeeds: a seed whose wall clock falls in a spring-forward gap keeps
// its nominal phase. The gap candidate is skipped, not shifted, and the search goes on from it.
func TestInitialDueGapSeeds(t *testing.T) {
	runInitialDueRows(t, []initialDueRow{
		{
			// 2023-03-12 02:30 does not exist in New York; the next instance is 03-13 02:30 EDT.
			name: "ny-daily-nominal-gap-seed", zone: "America/New_York",
			now: "2023-03-12 05:00 +00:00", hour: 2, minute: 30, rule: "FREQ=DAILY;COUNT=1",
			want:       "2023-03-13 06:30 +00:00",
			thenReason: reasonCountExhausted, // stored COUNT stays 1: nothing follows
		},
		{
			// 02:30 and 02:45 are in the gap; 03:00 EDT is the first instance, then 03:15 EDT.
			name: "ny-minutely-nominal-gap-seed", zone: "America/New_York",
			now: "2023-03-12 05:00 +00:00", hour: 2, minute: 30, rule: "FREQ=MINUTELY;INTERVAL=15;COUNT=2",
			want: "2023-03-12 07:00 +00:00",
			then: "2023-03-12 07:15 +00:00",
		},
	})
}

// TestInitialDuePhaseFixtures pins the server's initial-phase policy (rows F-1..F-6 below).
// These are not RFC 5545 rows and not preview/write parity rows: the seed fixes phase and implicit
// selectors only. Europe/Berlin, WKST=MO (rrule-go's default).
//
// Observed core result (2026-10-28, first green run of selectInitialDue, TZ=UTC and
// TZ=America/New_York): every row, its due date and its one completion step, equals the expected
// value written before the implementation, when every row returned no date.
func TestInitialDuePhaseFixtures(t *testing.T) {
	runInitialDueRows(t, []initialDueRow{
		{
			name: "F-1 weekly BYDAY=MO from a Wednesday seed", zone: "Europe/Berlin",
			now: "2026-10-28 08:00 +01:00", hour: 9, rule: "FREQ=WEEKLY;BYDAY=MO",
			want: "2026-11-02 09:00 +01:00", then: "2026-11-09 09:00 +01:00",
		},
		{
			// Mon 10-26 is in the seed's week but before the seed; the phase week is the seed's.
			name: "F-2 biweekly BYDAY=MO keeps the seed week's phase", zone: "Europe/Berlin",
			now: "2026-10-28 08:00 +01:00", hour: 9, rule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO",
			want: "2026-11-09 09:00 +01:00", then: "2026-11-23 09:00 +01:00",
		},
		{
			// November has no 31st.
			name: "F-3 monthly BYMONTHDAY=31", zone: "Europe/Berlin",
			now: "2026-10-28 08:00 +01:00", hour: 9, rule: "FREQ=MONTHLY;BYMONTHDAY=31",
			want: "2026-10-31 09:00 +01:00", then: "2026-12-31 09:00 +01:00",
		},
		{
			// Implicit day 31 from the seed; the seed is before now and February has no 31st.
			name: "F-4 monthly with the seed's implicit day 31", zone: "Europe/Berlin",
			now: "2027-01-31 10:00 +01:00", hour: 9, rule: "FREQ=MONTHLY",
			want: "2027-03-31 09:00 +02:00", then: "2027-05-31 09:00 +02:00",
		},
		{
			// The seed itself; the stored COUNT 3 counts it as instance 1.
			name: "F-5 daily COUNT=3 selects the seed", zone: "Europe/Berlin",
			now: "2026-10-28 08:00 +01:00", hour: 9, rule: "FREQ=DAILY;COUNT=3",
			want: "2026-10-28 09:00 +01:00", then: "2026-10-29 09:00 +01:00",
		},
		{
			// COUNT is ignored in the search; completion with the stored COUNT=1 is then exhausted.
			name: "F-6 daily COUNT=1 after the seed's time", zone: "Europe/Berlin",
			now: "2026-10-28 10:00 +01:00", hour: 9, rule: "FREQ=DAILY;COUNT=1",
			want: "2026-10-29 09:00 +01:00", thenReason: reasonCountExhausted,
		},
	})
}

// TestInitialDueUntil: UNTIL is inclusive at second precision, an UNTIL behind now is
// until_passed, and one at or after now but before the first candidate is no_occurrence.
func TestInitialDueUntil(t *testing.T) {
	// DAILY 09:00 Berlin, now 10:00 local: the first candidate is 2026-10-29 09:00 CET = 08:00Z.
	daily0900 := func(name, until, want string, reason noNextReason) initialDueRow {
		return initialDueRow{
			name: name, zone: "Europe/Berlin", now: "2026-10-28 10:00 +01:00", hour: 9,
			rule: "FREQ=DAILY;UNTIL=" + until, want: want, wantReason: reason,
		}
	}
	// DAILY 00:30 Berlin, now 10:00 local: the first candidate 2026-10-29 00:30 CET is
	// 2026-10-28T23:30Z, on the previous UTC day. Each UNTIL is asserted as a UTC instant.
	daily0030 := func(name, until, want string, reason noNextReason) initialDueRow {
		return initialDueRow{
			name: name, zone: "Europe/Berlin", now: "2026-10-28 10:00 +01:00", minute: 30,
			rule: "FREQ=DAILY;UNTIL=" + until, want: want, wantReason: reason,
		}
	}
	runInitialDueRows(t, []initialDueRow{
		daily0900("UNTIL before now", "20261028T080000Z", "", reasonUntilPassed),
		daily0900("UNTIL equal to now", "20261028T090000Z", "", reasonNoOccurrence),
		daily0900("UNTIL one second before the first candidate", "20261029T075959Z", "", reasonNoOccurrence),
		daily0900("UNTIL equal to the first candidate is inclusive", "20261029T080000Z", "2026-10-29 09:00 +01:00", ""),

		daily0030("UTC boundary: UNTIL = candidate - 1s", "20261028T232959Z", "", reasonNoOccurrence),
		daily0030("UTC boundary: UNTIL = candidate", "20261028T233000Z", "2026-10-28 23:30 +00:00", ""),
		daily0030("UTC boundary: UNTIL = candidate + 1s", "20261028T233001Z", "2026-10-28 23:30 +00:00", ""),
		daily0030("UTC boundary: UNTIL = the candidate day's local midnight", "20261028T230000Z", "", reasonNoOccurrence),
	})
}

// TestInitialDueCutoff pins "strictly after now" (spec US1 scenario 2) and a seed in a fall-back
// fold, which takes the zone's normal first instance (localise: the earlier one).
func TestInitialDueCutoff(t *testing.T) {
	runInitialDueRows(t, []initialDueRow{
		{
			name: "now exactly at the seed gives tomorrow", zone: "Europe/Berlin",
			now: "2026-10-28 09:00:00 +01:00", hour: 9, rule: "FREQ=DAILY",
			want: "2026-10-29 09:00 +01:00",
		},
		{
			name: "now one second before the seed gives the seed", zone: "Europe/Berlin",
			now: "2026-10-28 08:59:59 +01:00", hour: 9, rule: "FREQ=DAILY",
			want: "2026-10-28 09:00 +01:00",
		},
		{
			// 02:30 occurs twice on 2026-10-25 in Berlin; the seed is the earlier instant (CEST).
			name: "seed in a fall-back fold takes the earlier instant", zone: "Europe/Berlin",
			now: "2026-10-25 00:00 +02:00", hour: 2, minute: 30, rule: "FREQ=DAILY",
			want: "2026-10-25 02:30 +02:00", then: "2026-10-26 02:30 +01:00",
		},
		{
			// Known limit of the initial-search shim: a sub-daily rule iterates wall clocks, so the
			// repeated 01:00 is one candidate at its earlier instant (before now) and the fold's second
			// pass is never the first due date. Completion then steps elapsed time again.
			name: "sub-daily seed before a fall-back fold skips the fold's second pass", zone: "America/New_York",
			now: "2026-11-01 01:30 -04:00", rule: "FREQ=HOURLY",
			want: "2026-11-01 02:00 -05:00", then: "2026-11-01 03:00 -05:00",
		},
	})
}

// TestInitialDueBoundRow: a rule the guards accept but with no possible
// instance (30 February), so the search itself runs to its bound and ends with no_occurrence.
func TestInitialDueBoundRow(t *testing.T) {
	zone := mustZone(t, "Europe/Berlin")
	now := ts(t, zone, "2026-10-28 10:00 +01:00")

	start := time.Now()
	got, reason := selectInitialDue("FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=30", 9, 0, zone, now)
	elapsed := time.Since(start)

	assert.True(t, got.IsZero())
	assert.Equal(t, reasonNoOccurrence, reason)
	// The ceiling is 500 ms; the measurement is logged so a slow host shows up before it fails.
	t.Logf("exhausting search took %s", elapsed)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

// TestInitialDueGuardRefusals keeps guard refusals apart from the bound row above: they return
// before any evaluation, with the guard's own reason.
func TestInitialDueGuardRefusals(t *testing.T) {
	zone := mustZone(t, "Europe/Berlin")
	now := ts(t, zone, "2026-10-28 10:00 +01:00")
	for rule, want := range map[string]noNextReason{
		"":                         reasonNoRule,
		"FREQ=NOPE":                reasonInvalidRule,
		"FREQ=YEARLY;INTERVAL=101": reasonIntervalBeyondHorizon,
		"FREQ=MINUTELY;BYHOUR=9":   reasonPolicyRejected,
		"FREQ=DAILY;UNTIL=20200101T000000Z;COUNT=2": reasonUntilPassed,
	} {
		got, reason := selectInitialDue(rule, 9, 0, zone, now)
		assert.True(t, got.IsZero(), rule)
		assert.Equal(t, want, reason, rule)
	}
	got, reason := selectInitialDue("FREQ=DAILY", 9, 0, nil, now)
	assert.True(t, got.IsZero())
	assert.Equal(t, reasonNoZone, reason)
}

// TestInitialDueResultInServerZone: the selected instant is returned in the server zone, as
// stepOccurrences returns occurrences, at second precision.
func TestInitialDueResultInServerZone(t *testing.T) {
	zone := mustZone(t, "Europe/Berlin")
	got, reason := selectInitialDue("FREQ=DAILY", 9, 0, zone, ts(t, zone, "2026-10-28 10:00 +01:00"))
	assert.Empty(t, reason)
	assert.Equal(t, config.GetTimeZone().String(), got.Location().String())
	assert.Equal(t, got, got.Truncate(time.Second))
}

// TestInitialDueSearchRuleDropsCountOnly: the evaluation copy drops COUNT and nothing else.
func TestInitialDueSearchRuleDropsCountOnly(t *testing.T) {
	assert.Equal(t, "FREQ=DAILY;INTERVAL=2;UNTIL=20261231T000000Z",
		initialSearchRule("FREQ=DAILY;INTERVAL=2;COUNT=1;UNTIL=20261231T000000Z"))
	assert.Equal(t, "FREQ=WEEKLY;BYDAY=MO", initialSearchRule("FREQ=WEEKLY;BYDAY=MO"))
	assert.Equal(t, "FREQ=NOPE;COUNT=1", initialSearchRule("FREQ=NOPE;COUNT=1"))
}

func TestParseInitialDueTime(t *testing.T) {
	for in, want := range map[string][2]int{"00:00": {0, 0}, "23:59": {23, 59}, "09:00": {9, 0}, "19:45": {19, 45}} {
		hour, minute, err := ParseInitialDueTime(in)
		if assert.NoError(t, err, in) {
			assert.Equal(t, want, [2]int{hour, minute}, in)
		}
	}
	for _, in := range []string{"24:00", "9:00", "09:60", "09:00:00", "", " ", " 09:00", "09:00 ", "+09:00", "09-00", "０９:００"} {
		_, _, err := ParseInitialDueTime(in)
		assert.True(t, IsErrInvalidData(err), "%q must be refused", in)
	}
}

func TestInitialDueConflict(t *testing.T) {
	assert.NoError(t, InitialDueConflict(true, "", false, true), "no initial_due_time never conflicts")
	assert.NoError(t, InitialDueConflict(false, "09:00", true, false), "undated, ruled, open task")

	cases := []struct {
		name                  string
		dueSet, hasRule, done bool
		want                  InitialDueConflictKind
	}{
		{"with a due date", true, true, false, InitialDueConflictDueDate},
		{"without a rule", false, false, false, InitialDueConflictNoRule},
		{"on a done task", false, true, true, InitialDueConflictDone},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := InitialDueConflict(c.dueSet, "09:00", c.hasRule, c.done)
			var conflict ErrInitialDueConflict
			require.ErrorAs(t, err, &conflict)
			assert.Equal(t, c.want, conflict.Kind)

			// The write path is 400 (code 2002), the preview 422; the message is the same.
			write, ok := conflict.WriteError().(ErrInvalidData)
			if assert.True(t, ok) {
				assert.Equal(t, http.StatusBadRequest, write.HTTPError().HTTPCode)
				assert.Equal(t, ErrCodeInvalidData, write.HTTPError().Code)
				assert.Equal(t, conflict.Message, write.Message)
			}
			preview, ok := conflict.PreviewError().(ValidationHTTPError)
			if assert.True(t, ok) {
				assert.Equal(t, conflict.Message, preview.Message)
				assert.Contains(t, preview.InvalidFields, "initial_due_time")
			}
		})
	}
}

func TestInitialDueEngineClock(t *testing.T) {
	// A deterministic outer clock stands in for the real one, so restoring is checked without the
	// wall clock. The original clock comes back even if the test ends early.
	orig := timeNow
	t.Cleanup(func() { timeNow = orig })
	sentinel := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	timeNow = func() time.Time { return sentinel }

	pinned := time.Date(2026, 10, 28, 9, 0, 0, 999_000_000, time.UTC)
	restore := PinEngineClockForTests(pinned)
	assert.Equal(t, pinned.Truncate(time.Second), EngineNow())
	assert.Equal(t, pinned, timeNow(), "completion's clock is the same pin")
	restore()
	assert.Equal(t, sentinel, timeNow(), "restore puts the previous clock back")
	assert.Equal(t, sentinel, EngineNow())
}

// TestInitialDueSubDailyCost measures the costliest accepted sub-daily search: FREQ=SECONDLY from
// a 00:00 seed late in the day, where the floating iteration passes the day's candidates and then
// the look-ahead after the best one (nextIterated). Same ceiling as the bound row.
func TestInitialDueSubDailyCost(t *testing.T) {
	zone := mustZone(t, "Europe/Berlin")
	now := ts(t, zone, "2026-10-28 23:59:30 +01:00")

	start := time.Now()
	got, reason := selectInitialDue("FREQ=SECONDLY", 0, 0, zone, now)
	elapsed := time.Since(start)

	assert.Empty(t, reason)
	assert.True(t, ts(t, zone, "2026-10-28 23:59:31 +01:00").Equal(got), "got %s", got.UTC())
	t.Logf("SECONDLY search took %s", elapsed)
	assert.Less(t, elapsed, 500*time.Millisecond)
}
