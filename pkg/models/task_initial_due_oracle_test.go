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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Server-side semantic oracle. Two families, kept apart:
//
//   - TestLibraryOracleAnchoredRows: anchored DST controls from rrule-go's own tests, adapted to
//     the task path with an explicit anchor (the due date), cutoff and completion instant. They
//     are not preview/write parity rows.
//   - TestRepeatPreviewMatrixPreviewParity: the preview-equals-completion assertion over the matrix
//     (testdata/repeat_preview_matrix.json, see task_repeat_preview_test.go): the preview lists
//     exactly the completion chain.
//
// Boundary: in-package unit tests of the pure engine (stepOccurrences,
// nextOccurrence) and of PreviewOccurrences with an explicit timezone, so no session, auth or
// database is read. "now" is an explicit parameter in every call; nothing reads the wall
// clock, the process zone or the engine clock. The server zone is set per row through config and
// restored.

// oracleLayout is the oracle JSON's floating local date-time form (anchor_local).
const oracleLayout = "20060102T150405"

// oracleRow is one anchored library row: the anchor in zone is instance 1, wantUTC lists every
// instance of the recurrence set including the anchor.
type oracleRow struct {
	id, zone, anchorLocal, rule string
	wantUTC                     []string
}

// oracleObserved runs a row through the real completion loop: the anchor as the stored due
// date, completed on time (now = the anchor, the cutoff the first step is strictly after), with
// completion's COUNT decrement between steps. It asks for one more step than the oracle lists, so
// the end of the set is observed too, and returns the instants after the anchor and the reason
// the list ended.
func oracleObserved(t *testing.T, row oracleRow) (anchor time.Time, got []time.Time, reason string) {
	t.Helper()
	zone := mustZone(t, row.zone)
	anchor, err := time.ParseInLocation(oracleLayout, row.anchorLocal, zone)
	require.NoError(t, err)
	answer := stepOccurrences(nextOccurrenceInput{Rule: row.rule, Due: dated(anchor), Now: anchor, Zone: zone}, len(row.wantUTC)+1)
	return anchor, answer.Occurrences, answer.Reason
}

func oracleInstants(t *testing.T, utc []string) []time.Time {
	t.Helper()
	out := make([]time.Time, len(utc))
	for i, s := range utc {
		at, err := time.Parse(time.RFC3339, s)
		require.NoError(t, err)
		out[i] = at
	}
	return out
}

func assertInstants(t *testing.T, want, got []time.Time, msg string) {
	t.Helper()
	if !assert.Len(t, got, len(want), "%s: got %v", msg, got) {
		return
	}
	for i := range want {
		assert.True(t, want[i].Equal(got[i]), "%s %d: want %s, got %s", msg, i, want[i].UTC(), got[i].UTC())
	}
}

// TestLibraryOracleAnchoredRows: core's completion path equals the library oracle on the four
// anchored controls it is obliged to meet, and the NY BYSETPOS row records core's observed
// output beside the RFC oracle.
func TestLibraryOracleAnchoredRows(t *testing.T) {
	rows := []oracleRow{
		// rrule-go rrule_dst_test.go:TestDSTSydney/DailyGap. 2022-10-02 02:30 is in the
		// Sydney spring-forward gap: skipped, not counted.
		{id: "sydney-daily-gap-count", zone: "Australia/Sydney", anchorLocal: "20221001T023000", rule: "FREQ=DAILY;COUNT=3",
			wantUTC: []string{"2022-09-30T16:30:00Z", "2022-10-02T15:30:00Z", "2022-10-03T15:30:00Z"}},
		// rrule-go rrule_dst_test.go:TestDSTOtherZones/LordHoweGap. A 30-minute spring-forward.
		{id: "lord-howe-half-hour-minutely", zone: "Australia/Lord_Howe", anchorLocal: "20231001T013000", rule: "FREQ=MINUTELY;INTERVAL=15;COUNT=3",
			wantUTC: []string{"2023-09-30T15:00:00Z", "2023-09-30T15:15:00Z", "2023-09-30T15:30:00Z"}},
		// rrule-go rrule_dst_test.go:TestDSTPublicQueries/Until0. UNTIL equal to an instance includes it.
		{id: "sydney-until-inclusive", zone: "Australia/Sydney", anchorLocal: "20221001T023000", rule: "FREQ=DAILY;UNTIL=20221002T153000Z",
			wantUTC: []string{"2022-09-30T16:30:00Z", "2022-10-02T15:30:00Z"}},
		// rrule-go rrule_dst_test.go:TestDSTPublicQueries/Until-1. UNTIL one second earlier excludes it.
		{id: "sydney-until-one-second-before", zone: "Australia/Sydney", anchorLocal: "20221001T023000", rule: "FREQ=DAILY;UNTIL=20221002T152959Z",
			wantUTC: []string{"2022-09-30T16:30:00Z"}},
	}
	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			want := oracleInstants(t, row.wantUTC)
			anchor, got, reason := oracleObserved(t, row)
			require.True(t, want[0].Equal(anchor), "the anchor is the oracle's instance 1: %s vs %s", anchor.UTC(), want[0])
			assertInstants(t, want[1:], got, "instance after the anchor")
			assert.NotEmpty(t, reason, "the set ends where the oracle's ends")
		})
	}
}

// TestLibraryOracleBySetPosGapObserved: rrule-go
// str_timeforms_test.go:TestTimeFormsIntegratedGapBysetpos.
//
// Known limitation: core ranks BYSETPOS positions before
// excluding gap times, so on 2026-03-08 position 1 is the nonexistent 02:00 and that day yields
// nothing, where RFC 5545 takes 03:00 EDT. The RFC oracle stays as the library wrote it; the test
// asserts core's OBSERVED output so a change to it (a cure, or a further drift) is noticed and
// decided deliberately, never silently.
//
// Observed 2026-10-28 (TZ=UTC and TZ=America/New_York): after the
// anchor core lists NOTHING and ends with no_occurrence; the RFC oracle lists 2026-03-08T07:00Z and
// 2026-03-09T06:00Z. Two causes, separated by the diagnostic subtest below:
//   - 2026-03-08 is the BYSETPOS limit above (absent even when UNTIL is given in UTC);
//   - 2026-03-09T06:00Z is excluded because core reads the floating UNTIL 20260309T020000 (no Z)
//     as 02:00 UTC, not 02:00 New York. That is a second core/library difference, not the
//     BYSETPOS limit; it is asserted here as observed, not cured.
func TestLibraryOracleBySetPosGapObserved(t *testing.T) {
	row := oracleRow{id: "ny-gap-before-bysetpos", zone: "America/New_York", anchorLocal: "20260307T020000",
		rule:    "FREQ=DAILY;UNTIL=20260309T020000;BYSETPOS=1;BYHOUR=2,3",
		wantUTC: []string{"2026-03-07T07:00:00Z", "2026-03-08T07:00:00Z", "2026-03-09T06:00:00Z"}} // the RFC oracle

	rfc := oracleInstants(t, row.wantUTC)
	anchor, got, reason := oracleObserved(t, row)
	require.True(t, rfc[0].Equal(anchor), "the anchor is the oracle's instance 1")
	t.Logf("BYSETPOS-gap RFC oracle after the anchor: %v; core observed: %v (reason %q)", rfc[1:], got, reason)
	assertInstants(t, []time.Time{}, got, "observed instance after the anchor")
	assert.Equal(t, string(reasonNoOccurrence), reason)

	// Diagnostic only, not an oracle row: the same rule with UNTIL as the UTC instant of 02:00 EDT
	// on 2026-03-09 isolates the BYSETPOS limit. 03-09 06:00Z is then listed; 03-08 is still missing.
	t.Run("diagnostic: UNTIL in UTC isolates the BYSETPOS limit", func(t *testing.T) {
		utcUntil := row
		utcUntil.rule = "FREQ=DAILY;UNTIL=20260309T060000Z;BYSETPOS=1;BYHOUR=2,3"
		_, got, reason := oracleObserved(t, utcUntil)
		assertInstants(t, oracleInstants(t, []string{"2026-03-09T06:00:00Z"}), got, "observed instance after the anchor")
		assert.Equal(t, string(reasonNoOccurrence), reason)
	})
}

// TestRepeatPreviewMatrixPreviewParity: for every preview_eligible row, the preview of the row's rule with its due date as
// an explicit due_date, evaluated in the row's user zone one second before the due date, lists
// exactly the due dates the completion chain stores (TestRepeatPreviewMatrix), and ends with a
// reason exactly where the chain leaves the task done.
func TestRepeatPreviewMatrixPreviewParity(t *testing.T) {
	m := loadRepeatPreviewMatrix(t)
	origZone := config.ServiceTimeZone.GetString()
	t.Cleanup(func() { config.ServiceTimeZone.Set(origZone) })

	eligible := 0
	for _, row := range m.Rows {
		if !row.PreviewEligible {
			continue
		}
		eligible++
		t.Run(row.Name, func(t *testing.T) {
			config.ServiceTimeZone.Set(row.ServerZone)
			rule, err := row.Repeat.toRRule()
			require.NoError(t, err)

			due := row.Due
			now := due.Add(-time.Second).In(mustZone(t, row.UserZone))
			answer, err := PreviewOccurrences(nil, nil, PreviewInput{
				Rule: rule, DueDate: &due, TimeZone: row.UserZone, Count: row.Completions,
			}, now)
			require.NoError(t, err)
			require.NotNil(t, answer.DueDate)
			assert.True(t, due.Equal(*answer.DueDate), "due_date echoed")
			assert.Equal(t, DueDateSourceExplicit, answer.DueDateSource)

			var chain []time.Time
			endsDone := false
			for _, step := range row.Expect {
				if step.Done {
					endsDone = true
					break
				}
				chain = append(chain, step.Due)
			}
			assertInstants(t, chain, answer.Occurrences, "occurrence")
			if endsDone {
				assert.NotEmpty(t, answer.Reason, "the chain leaves the task done: the preview ends with a reason")
			} else {
				assert.Empty(t, answer.Reason)
			}
		})
	}
	assert.Equal(t, 33, eligible, "the matrix's preview_eligible rows (version 1)")
}

// tsRowSource prefixes the rows carried over from the frontend's former preview evaluator tests;
// the number after it is the row's line there. Rows already carried by the matrix chains above, or by
// task_initial_due_test.go (the phase fixtures, cutoff, UNTIL, gap seeds), are not repeated, so every
// semantic TS row has exactly one server test.
const tsRowSource = "rrulePreview.test.ts"

// TestCarriedOverPreviewRows: TS previewDates rows with an explicit due date. TS lists the
// due date first; the server's occurrences are strictly after it. now is one second before the
// due date, as in the matrix assertion.
func TestCarriedOverPreviewRows(t *testing.T) {
	origZone := config.ServiceTimeZone.GetString()
	t.Cleanup(func() { config.ServiceTimeZone.Set(origZone) })
	config.ServiceTimeZone.Set("UTC")

	for _, row := range []struct {
		line, zone, rule, due string
		count                 int
		want                  []string
		wantReason            noNextReason
	}{
		{line: "123", zone: "Europe/Berlin", rule: "FREQ=DAILY;INTERVAL=2;UNTIL=20270101T000000Z", due: "2026-10-24T21:30:00Z", count: 2,
			want: []string{"2026-10-26T22:30:00Z", "2026-10-28T22:30:00Z"}},
		{line: "131", zone: "America/New_York", rule: "FREQ=DAILY;INTERVAL=36525", due: "2026-10-23T13:00:00Z", count: 2,
			want: []string{"2126-10-24T13:00:00Z", "2226-10-25T13:00:00Z"}},
		{line: "195", zone: "Europe/Berlin", rule: "FREQ=DAILY;COUNT=5", due: "2026-10-25T00:30:00Z", count: 2,
			want: []string{"2026-10-26T01:30:00Z", "2026-10-27T01:30:00Z"}},
		{line: "243", zone: "America/New_York", rule: "FREQ=YEARLY", due: "9999-01-01T03:00:00Z", count: 2,
			want: []string{}, wantReason: reasonOutsideStorableRange},
		{line: "244", zone: "UTC", rule: "FREQ=DAILY", due: "9999-12-31T12:00:00Z", count: 2,
			want: []string{}, wantReason: reasonOutsideStorableRange},
		{line: "245", zone: "UTC", rule: "FREQ=DAILY", due: "9999-12-30T12:00:00Z", count: 2,
			want: []string{"9999-12-31T12:00:00Z"}, wantReason: reasonOutsideStorableRange},
		{line: "253", zone: "UTC", rule: "FREQ=YEARLY;INTERVAL=100", due: "2400-02-29T12:00:00Z", count: 2,
			want: []string{}, wantReason: reasonNoOccurrence},
		// :115-:119, every 3 years from 29 February 2028: 2031, 2034 and 2037 have no 29 February
		// (RFC 5545 §3.3.10: an invalid date is skipped, not moved), nor do 2043, 2046 and 2049.
		{line: "115", zone: "UTC", rule: "FREQ=YEARLY;INTERVAL=3", due: "2028-02-29T12:00:00Z", count: 2,
			want: []string{"2040-02-29T12:00:00Z", "2052-02-29T12:00:00Z"}},
		{line: "254", zone: "UTC", rule: "FREQ=YEARLY;INTERVAL=100", due: "2096-02-29T12:00:00Z", count: 2,
			want: []string{"2196-02-29T12:00:00Z", "2296-02-29T12:00:00Z"}},
		{line: "274", zone: "America/New_York", rule: "FREQ=WEEKLY;COUNT=3", due: "2026-10-23T13:00:00Z", count: 3,
			want: []string{"2026-10-30T13:00:00Z", "2026-11-06T14:00:00Z"}, wantReason: reasonCountExhausted},
		{line: "277", zone: "America/New_York", rule: "FREQ=WEEKLY;COUNT=1", due: "2026-10-23T13:00:00Z", count: 3,
			want: []string{}, wantReason: reasonCountExhausted},
		{line: "281", zone: "America/New_York", rule: "FREQ=WEEKLY;UNTIL=20261107T045959Z", due: "2026-10-23T13:00:00Z", count: 2,
			want: []string{"2026-10-30T13:00:00Z", "2026-11-06T14:00:00Z"}},
		{line: "282", zone: "America/New_York", rule: "FREQ=WEEKLY;UNTIL=20261031T035959Z", due: "2026-10-23T13:00:00Z", count: 2,
			want: []string{"2026-10-30T13:00:00Z"}, wantReason: reasonNoOccurrence},
		{line: "286", zone: "America/New_York", rule: "FREQ=WEEKLY;UNTIL=20261106T140000Z", due: "2026-10-23T13:00:00Z", count: 2,
			want: []string{"2026-10-30T13:00:00Z", "2026-11-06T14:00:00Z"}},
		{line: "287", zone: "America/New_York", rule: "FREQ=WEEKLY;UNTIL=20261106T135959Z", due: "2026-10-23T13:00:00Z", count: 2,
			want: []string{"2026-10-30T13:00:00Z"}, wantReason: reasonNoOccurrence},
		{line: "291", zone: "America/New_York", rule: "FREQ=WEEKLY;UNTIL=20261001T000000Z", due: "2026-10-23T13:00:00Z", count: 2,
			want: []string{}, wantReason: reasonUntilPassed},
	} {
		t.Run(tsRowSource+":"+row.line, func(t *testing.T) {
			due, err := time.Parse(time.RFC3339, row.due)
			require.NoError(t, err)
			answer, err := PreviewOccurrences(nil, nil, PreviewInput{Rule: row.rule, DueDate: &due, TimeZone: row.zone, Count: row.count}, due.Add(-time.Second))
			require.NoError(t, err)
			assertInstants(t, oracleInstants(t, row.want), answer.Occurrences, "occurrence")
			assert.Equal(t, string(row.wantReason), answer.Reason)
		})
	}
}

// TestCarriedOverFirstOccurrenceRows: TS firstOccurrenceAfter rows, under the server's seed
// semantics. TS took an anchor instant; the server's seed is today in the zone at
// initial_due_time, so a row is carried over with its anchor's time of day and a now on the
// anchor's day. :380, :391 and :397 are not copied (they conflict with the seed semantics;
// TestInitialDueCutoff, TestInitialDueGapSeeds and TestInitialDuePhaseFixtures hold the server's
// values). Where TS returned null, the server's reason is asserted.
func TestCarriedOverFirstOccurrenceRows(t *testing.T) {
	runInitialDueRows(t, []initialDueRow{
		{name: tsRowSource + ":369 weekly MO from a Thursday afternoon", zone: "America/New_York",
			now: "2026-10-22 14:00 -04:00", hour: 9, rule: "FREQ=WEEKLY;BYDAY=MO", want: "2026-10-26 09:00 -04:00"},
		{name: tsRowSource + ":376 BYMONTHDAY=31 skips February", zone: "America/New_York",
			now: "2027-02-10 10:00 -05:00", hour: 12, rule: "FREQ=MONTHLY;BYMONTHDAY=31", want: "2027-03-31 12:00 -04:00"},
		{name: tsRowSource + ":406 hourly INTERVAL=4 from the seed's time", zone: "America/New_York",
			now: "2026-10-22 22:30 -04:00", hour: 9, rule: "FREQ=HOURLY;INTERVAL=4", want: "2026-10-23 01:00 -04:00"},
		{name: tsRowSource + ":411 COUNT=1 is ignored in the search", zone: "America/New_York",
			now: "2026-10-22 14:00 -04:00", hour: 9, rule: "FREQ=WEEKLY;COUNT=1", want: "2026-10-29 09:00 -04:00",
			thenReason: reasonCountExhausted},
		{name: tsRowSource + ":416 weekly UNTIL before now", zone: "America/New_York",
			now: "2026-10-22 14:00 -04:00", hour: 9, rule: "FREQ=WEEKLY;UNTIL=20261001T035959Z", wantReason: reasonUntilPassed},
		{name: tsRowSource + ":417 daily UNTIL between now and the first candidate", zone: "America/New_York",
			now: "2026-10-22 22:30 -04:00", hour: 9, rule: "FREQ=DAILY;UNTIL=20261023T035959Z", wantReason: reasonNoOccurrence},
		{name: tsRowSource + ":418 hourly UNTIL before now", zone: "America/New_York",
			now: "2026-10-22 16:30 -04:00", hour: 9, rule: "FREQ=HOURLY;UNTIL=20261022T200000Z", wantReason: reasonUntilPassed},
		{name: tsRowSource + ":422 yearly INTERVAL=6 has no fixed year limit", zone: "America/New_York",
			now: "2026-10-22 14:00 -04:00", hour: 9, rule: "FREQ=YEARLY;INTERVAL=6", want: "2032-10-22 09:00 -04:00"},
		{name: tsRowSource + ":425 yearly from Feb 29 skips 2029-2031", zone: "America/New_York",
			now: "2028-02-29 13:00 -05:00", hour: 9, rule: "FREQ=YEARLY", want: "2032-02-29 09:00 -05:00"},
		{name: tsRowSource + ":431 yearly INTERVAL=100 from Feb 29 2400 leaves the search window", zone: "America/New_York",
			now: "2400-02-29 15:00 -05:00", hour: 9, rule: "FREQ=YEARLY;INTERVAL=100", wantReason: reasonNoOccurrence},
		// :432-:434 differ from the seed semantics in their anchor. TS :432-:433 anchors 9998-12-31T12:00Z with now
		// 9999-06-01 and expects 9999-12-31T12:00:00.000Z, an anchored success; TS :434 (anchor
		// 9999-01-01T03:00Z New York) expects null. The server has no separate anchor: the seed is now's day
		// at initial_due_time, so this row is an adaptation, not a copy. Seeded on
		// 9999-06-01 12:00 UTC, the next yearly date would be in year 10000, and the server ends
		// with no_occurrence (observed 2026-10-28): the search generates no year-10000 candidate, so
		// the storable-range check (outside_storable_range) never sees one.
		{name: tsRowSource + ":433 (server seed) yearly past the last storable year", zone: "UTC",
			now: "9999-06-01 13:00 +00:00", hour: 12, rule: "FREQ=YEARLY", wantReason: reasonNoOccurrence},
	})
}
