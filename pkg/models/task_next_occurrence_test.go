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
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

// nextOccurrenceNow is the reference "now" of these tests, a Thursday.
var nextOccurrenceNow = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

func dated(at time.Time) datedValue { return datedValue{At: at} }

// assertNothingMoved checks that a refused call returned every input date unchanged.
func assertNothingMoved(t *testing.T, in nextOccurrenceInput, res nextOccurrenceResult) {
	t.Helper()
	assert.True(t, res.Due.At.Equal(in.Due.At), "due: got %v, want %v", res.Due.At, in.Due.At)
	assert.True(t, res.Start.At.Equal(in.Start.At), "start: got %v, want %v", res.Start.At, in.Start.At)
	assert.True(t, res.End.At.Equal(in.End.At), "end: got %v, want %v", res.End.At, in.End.At)
	require.Len(t, res.Reminders, len(in.Reminders))
	for i := range in.Reminders {
		assert.True(t, res.Reminders[i].At.Equal(in.Reminders[i].At), "reminder %d", i)
	}
}

// ruleTooLongForCountWriteBack is a COUNT rule that fits the column as stored, but whose COUNT-1
// serialisation (a date-only UNTIL gains a time) does not, built like
// TestRRuleLengthLimitAfterNormalisation.
func ruleTooLongForCountWriteBack(t *testing.T) string {
	t.Helper()
	raw := "FREQ=YEARLY;COUNT=5;UNTIL=20300101;BYYEARDAY=1"
	for i := 2; len(raw)+len(","+strconv.Itoa(i)) <= 499; i++ {
		raw += "," + strconv.Itoa(i)
	}
	opt, err := rrule.StrToROption(raw)
	require.NoError(t, err)
	opt.Count--
	require.Greater(t, len(opt.RRuleString()), maxRRuleLength, "the COUNT-1 serialisation must exceed the column")
	return raw
}

// TestNextOccurrence_Reasons walks the guard chain in its order: each refusal has
// its stable reason, and nothing moves.
func TestNextOccurrence_Reasons(t *testing.T) {
	now := nextOccurrenceNow
	var logged []int64
	orig := logNoOccurrenceInWindow
	logNoOccurrenceInWindow = func(taskID int64) { logged = append(logged, taskID) }
	t.Cleanup(func() { logNoOccurrenceInWindow = orig })

	_, err := rrule.StrToROption("FREQ=MONTHLY;BYMONTHDAY=32")
	require.NoError(t, err, "the outside_bounds rule must parse, so the bounds are what refuse it")

	cases := []struct {
		name   string
		in     nextOccurrenceInput
		reason noNextReason
	}{
		{"no_zone", nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(now), Now: now}, reasonNoZone},
		{"no_rule", nextOccurrenceInput{Rule: "", Due: dated(now), Now: now, Zone: time.UTC}, reasonNoRule},
		{"invalid_rule", nextOccurrenceInput{Rule: "FREQ=NOPE", Due: dated(now), Now: now, Zone: time.UTC}, reasonInvalidRule},
		{"interval_beyond_horizon", nextOccurrenceInput{Rule: "FREQ=YEARLY;INTERVAL=101", Due: dated(now), Now: now, Zone: time.UTC}, reasonIntervalBeyondHorizon},
		{"anchor_beyond_horizon", nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(now.AddDate(-200, 0, -1)), Now: now, Zone: time.UTC}, reasonAnchorBeyondHorizon},
		{"outside_bounds", nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYMONTHDAY=32", Due: dated(now), Now: now, Zone: time.UTC}, reasonOutsideBounds},
		{"policy_rejected", nextOccurrenceInput{Rule: "FREQ=HOURLY;BYMINUTE=0", Due: dated(now), Now: now, Zone: time.UTC}, reasonPolicyRejected},
		{"rule_too_long", nextOccurrenceInput{Rule: ruleTooLongForCountWriteBack(t), Due: dated(now), Now: now, Zone: time.UTC}, reasonRuleTooLong},
		{"count_exhausted", nextOccurrenceInput{Rule: "FREQ=DAILY;COUNT=1", Due: dated(now), Now: now, Zone: time.UTC}, reasonCountExhausted},
		{"until_passed", nextOccurrenceInput{Rule: "FREQ=DAILY;UNTIL=20261001T000000Z", Due: dated(time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)), Now: now, Zone: time.UTC}, reasonUntilPassed},
		{"no_occurrence, finite", nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYMONTHDAY=1;COUNT=2", Due: dated(time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC)), Now: now, Zone: time.UTC, TaskID: 1}, reasonNoOccurrence},
		{"no_occurrence, infinite", nextOccurrenceInput{Rule: "FREQ=YEARLY;INTERVAL=100;BYMONTH=10;BYMONTHDAY=22", Due: dated(time.Date(1826, 10, 23, 12, 0, 0, 0, time.UTC)), Now: now, Zone: time.UTC, TaskID: 2}, reasonNoOccurrence},
		{"outside_storable_range", nextOccurrenceInput{
			Rule: "FREQ=DAILY;INTERVAL=365",
			Due:  dated(time.Date(9999, 12, 1, 9, 0, 0, 0, time.UTC)),
			// a set start and reminder, to check they are returned unchanged too
			Start:     dated(time.Date(9999, 11, 30, 9, 0, 0, 0, time.UTC)),
			Reminders: []datedValue{dated(time.Date(9999, 11, 29, 9, 0, 0, 0, time.UTC))},
			Now:       time.Date(9999, 12, 10, 0, 0, 0, 0, time.UTC), Zone: time.UTC,
		}, reasonOutsideStorableRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := nextOccurrence(tc.in)
			assert.Equal(t, tc.reason, res.Reason)
			assertNothingMoved(t, tc.in, res)
		})
	}
	assert.Equal(t, []int64{2}, logged, "only the infinite rule logs a search-window exhaustion")

	t.Run("evaluation_failed", func(t *testing.T) {
		// Any panic in the evaluation stage is recovered. A nil option stands in for one: it panics
		// inside the stage whatever rrule-go version is in use, so this does not depend on a
		// library bug the bounds already keep out of reach.
		const rule = "FREQ=MONTHLY;BYDAY=-1MO"
		var opt *rrule.ROption
		in := nextOccurrenceInput{
			Rule: rule, Due: dated(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)),
			Reminders: []datedValue{dated(time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC))},
			Now:       now, Zone: time.UTC, TaskID: 99,
		}
		var res nextOccurrenceResult
		require.NotPanics(t, func() { res = evaluateNextOccurrence(in, opt, rrule.ROption{}) })
		assert.Equal(t, reasonEvaluationFailed, res.Reason)
		assertNothingMoved(t, in, res)
	})

	t.Run("a successful call has no reason", func(t *testing.T) {
		res := nextOccurrence(nextOccurrenceInput{Rule: "FREQ=WEEKLY;BYDAY=MO", Due: dated(now), Now: now, Zone: time.UTC})
		assert.Empty(t, res.Reason)
		assert.True(t, res.Due.At.After(now))
	})
}

// TestNextOccurrence_ReasonCodesAreStable fails when a reason string changes, a code is added,
// or the withdrawn no_anchor comes back.
func TestNextOccurrence_ReasonCodesAreStable(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "task_next_occurrence.go", nil, 0)
	require.NoError(t, err)
	var declared []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		if ident, ok := spec.Type.(*ast.Ident); !ok || ident.Name != "noNextReason" {
			return true
		}
		for _, v := range spec.Values {
			lit, ok := v.(*ast.BasicLit)
			require.True(t, ok, "reason constants are string literals")
			s, err := strconv.Unquote(lit.Value)
			require.NoError(t, err)
			declared = append(declared, s)
		}
		return true
	})
	sort.Strings(declared)
	want := []string{
		"anchor_beyond_horizon", "count_exhausted", "evaluation_failed", "interval_beyond_horizon",
		"invalid_rule", "no_occurrence", "no_rule", "no_zone", "outside_bounds",
		"outside_storable_range", "policy_rejected", "rule_too_long", "until_passed",
	}
	assert.Equal(t, want, declared)
	assert.NotContains(t, declared, "no_anchor")
}

// TestNextOccurrence_Deterministic: the result depends on the inputs alone, not on the process
// zone or the service time zone.
func TestNextOccurrence_Deterministic(t *testing.T) {
	inputs := []nextOccurrenceInput{
		{Rule: "FREQ=WEEKLY;BYDAY=MO,WE,FR", Due: dated(time.Date(2026, 10, 26, 20, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 26, 20, 5, 0, 0, time.UTC), Zone: time.UTC},
		{Rule: "FREQ=MONTHLY;BYMONTHDAY=15", Due: dated(time.Date(2026, 10, 15, 21, 0, 0, 0, time.UTC)), Start: dated(time.Date(2026, 10, 14, 21, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 15, 21, 5, 0, 0, time.UTC), Zone: time.UTC},
		{Rule: "FREQ=MONTHLY;BYMONTHDAY=1", FromCompletion: true, Due: dated(time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)), Now: nextOccurrenceNow, Zone: time.UTC},
	}
	before := make([]nextOccurrenceResult, len(inputs))
	for i, in := range inputs {
		before[i] = nextOccurrence(in)
		require.Empty(t, before[i].Reason, "case %d", i)
	}

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	withProcessZone(t, tokyo)
	origZone := config.ServiceTimeZone.GetString()
	t.Cleanup(func() { config.ServiceTimeZone.Set(origZone) })
	config.ServiceTimeZone.Set("America/Los_Angeles")

	for i, in := range inputs {
		after := nextOccurrence(in)
		assert.Equal(t, before[i].Reason, after.Reason, "case %d", i)
		for _, pair := range [][2]datedValue{{before[i].Due, after.Due}, {before[i].Start, after.Start}, {before[i].End, after.End}} {
			assert.True(t, pair[0].At.Equal(pair[1].At), "case %d: %v vs %v", i, pair[0].At, pair[1].At)
			assert.Equal(t, pair[0].Intended, pair[1].Intended, "case %d", i)
			assert.Equal(t, pair[0].IntendedZone, pair[1].IntendedZone, "case %d", i)
		}
	}
}

// withProcessZone sets the process zone for one test. The helper must not depend on it,
// so this is the one place a test deliberately changes it.
func withProcessZone(t *testing.T, loc *time.Location) {
	t.Helper()
	orig := time.Local                      //nolint:gosmopolitan // proving the helper ignores the process zone
	t.Cleanup(func() { time.Local = orig }) //nolint:gosmopolitan // restores the process zone
	time.Local = loc                        //nolint:gosmopolitan // proving the helper ignores the process zone
}

// nextOccurrenceGoldenCases are in UTC, which has no clock change, so their values are stable.
// The zone cases follow the engine tests below.
var nextOccurrenceGoldenCases = []struct {
	name string
	in   nextOccurrenceInput
	due  time.Time
}{
	{
		name: "weekly BYDAY",
		in:   nextOccurrenceInput{Rule: "FREQ=WEEKLY;BYDAY=MO,WE,FR", Due: dated(time.Date(2026, 10, 26, 20, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 26, 20, 5, 0, 0, time.UTC), Zone: time.UTC},
		due:  time.Date(2026, 10, 28, 20, 0, 0, 0, time.UTC),
	},
	{
		name: "monthly on the 15th",
		in:   nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYMONTHDAY=15", Due: dated(time.Date(2026, 10, 15, 21, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 15, 21, 5, 0, 0, time.UTC), Zone: time.UTC},
		due:  time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC),
	},
	{
		name: "second Tuesday",
		in:   nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYDAY=2TU", Due: dated(time.Date(2026, 10, 13, 22, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 13, 22, 5, 0, 0, time.UTC), Zone: time.UTC},
		due:  time.Date(2026, 11, 10, 22, 0, 0, 0, time.UTC),
	},
	{
		name: "COUNT",
		in:   nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYMONTHDAY=15;COUNT=3", Due: dated(time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC), Zone: time.UTC},
		due:  time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC),
	},
	{
		name: "from completion",
		in:   nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYMONTHDAY=1", FromCompletion: true, Due: dated(time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)), Now: nextOccurrenceNow, Zone: time.UTC},
		due:  time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC), // anchored at Now, so Now's clock time
	},
}

// TestNextOccurrence_Golden pins the UTC golden cases.
func TestNextOccurrence_Golden(t *testing.T) {
	for _, tc := range nextOccurrenceGoldenCases {
		t.Run(tc.name, func(t *testing.T) {
			res := nextOccurrence(tc.in)
			require.Empty(t, res.Reason)
			assert.True(t, res.Due.At.Equal(tc.due), "got %v, want %v", res.Due.At, tc.due)
			assert.Nil(t, res.Due.Intended)
			assert.Empty(t, res.Due.IntendedZone)
			assert.True(t, res.Start.At.IsZero(), "an unset date stays unset")
			assert.True(t, res.End.At.IsZero(), "an unset date stays unset")
		})
	}
}

// TestNextOccurrence_IntendedIsAcceptedAndIgnored: an intended clock time in a zone without a clock
// change near the date fails the gap check (g = 0) and is ignored, never an
// error. TestNextOccurrenceIntended covers the intents that apply.
func TestNextOccurrence_IntendedIsAcceptedAndIgnored(t *testing.T) {
	intended := 2*time.Hour + 30*time.Minute
	tc := nextOccurrenceGoldenCases[1]
	in := tc.in
	in.Due.Intended, in.Due.IntendedZone = &intended, "UTC"
	var res nextOccurrenceResult
	require.NotPanics(t, func() { res = nextOccurrence(in) })
	require.Empty(t, res.Reason)
	assert.True(t, res.Due.At.Equal(tc.due), "got %v, want %v", res.Due.At, tc.due)
}

// TestNextOccurrence_EchoedDatesCarryNoCompanionIntent: a path that returns an input date as it is
// (a guard, an unset date) still returns Start, End and the reminders without Intended, and no date
// with an IntendedZone; the due date's own Intended is echoed, as only a due date carries one.
func TestNextOccurrence_EchoedDatesCarryNoCompanionIntent(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }
	withIntent := func(d time.Time) datedValue {
		v := dz(t, d, "02:30")
		v.IntendedZone = "Europe/Berlin"
		return v
	}
	for _, tc := range []struct {
		name string
		in   nextOccurrenceInput
	}{
		{"no_rule guard", nextOccurrenceInput{Due: withIntent(at("2027-03-14 03:30 -04:00")), Start: withIntent(at("2027-03-14 03:30 -04:00")),
			Reminders: []datedValue{withIntent(at("2027-03-14 03:30 -04:00"))}, Now: at("2027-03-14 04:00 -04:00"), Zone: ny}},
		{"unset end", nextOccurrenceInput{Rule: "FREQ=DAILY", Due: withIntent(at("2027-03-14 03:30 -04:00")), End: datedValue{Intended: tod(t, "02:30"), IntendedZone: "UTC"},
			Now: at("2027-03-14 04:00 -04:00"), Zone: ny}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := nextOccurrence(tc.in)
			for i, d := range append([]datedValue{res.Start, res.End}, res.Reminders...) {
				assert.Nil(t, d.Intended, "date %d", i)
				assert.Empty(t, d.IntendedZone, "date %d", i)
			}
			assert.Empty(t, res.Due.IntendedZone)
			if res.Reason != "" {
				assert.Equal(t, tc.in.Due.Intended, res.Due.Intended, "an unmoved due date echoes its Intended")
			}
			assert.Equal(t, "Europe/Berlin", tc.in.Due.IntendedZone, "the input is not changed")
		})
	}
}

// TestNextOccurrence_DoesNotMutateItsInput: the helper is pure; the caller's reminders slice is
// read, never written.
func TestNextOccurrence_DoesNotMutateItsInput(t *testing.T) {
	reminder := time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC)
	in := nextOccurrenceInput{
		Rule: "FREQ=WEEKLY;BYDAY=MO", Due: dated(time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)),
		Start:     dated(time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)),
		Reminders: []datedValue{dated(reminder)},
		Now:       time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC), Zone: time.UTC,
	}
	res := nextOccurrence(in)
	require.Empty(t, res.Reason)
	assert.Equal(t, reminder, in.Reminders[0].At, "the input reminder is untouched")
	require.Len(t, res.Reminders, 1)
	assert.True(t, res.Reminders[0].At.Equal(time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)), "got %v", res.Reminders[0].At)
	assert.True(t, res.Start.At.Equal(time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)), "got %v", res.Start.At)
}

// TestNextOccurrence_EqualsCompletion: completion stores exactly what nextOccurrence returns for
// the task's dates.
func TestNextOccurrence_EqualsCompletion(t *testing.T) {
	for _, tc := range nextOccurrenceGoldenCases {
		t.Run(tc.name, func(t *testing.T) {
			withFixedNow(t, tc.in.Now)
			res := nextOccurrence(tc.in)
			require.Empty(t, res.Reason)

			task := &Task{Repeats: tc.in.Rule, RepeatsFromCurrentDate: tc.in.FromCompletion, DueDate: tc.in.Due.At}
			n := completeRepeatingIn(task, time.UTC)
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.Equal(res.Due.At), "completion %v, helper %v", n.DueDate, res.Due.At)
		})
	}

	t.Run("COUNT is written back by completion, not the helper", func(t *testing.T) {
		tc := nextOccurrenceGoldenCases[3]
		withFixedNow(t, tc.in.Now)
		n := completeRepeatingIn(&Task{Repeats: tc.in.Rule, DueDate: tc.in.Due.At}, time.UTC)
		opt, err := rrule.StrToROption(n.Repeats)
		require.NoError(t, err)
		assert.Equal(t, 2, opt.Count, "stored rule %q", n.Repeats)
		assert.Equal(t, []int{15}, opt.Bymonthday)
	})

	t.Run("every date and reminder", func(t *testing.T) {
		in := nextOccurrenceInput{
			Rule:      "FREQ=MONTHLY;BYDAY=2TU",
			Due:       dated(time.Date(2026, 10, 13, 22, 0, 0, 0, time.UTC)),
			Start:     dated(time.Date(2026, 10, 12, 8, 0, 0, 0, time.UTC)),
			End:       dated(time.Date(2026, 10, 14, 18, 0, 0, 0, time.UTC)),
			Reminders: []datedValue{dated(time.Date(2026, 10, 13, 21, 0, 0, 0, time.UTC)), dated(time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC))},
			Now:       time.Date(2026, 10, 13, 22, 5, 0, 0, time.UTC),
			Zone:      time.UTC,
		}
		withFixedNow(t, in.Now)
		res := nextOccurrence(in)
		require.Empty(t, res.Reason)

		task := &Task{
			Repeats: in.Rule, DueDate: in.Due.At, StartDate: in.Start.At, EndDate: in.End.At,
			Reminders: []*TaskReminder{
				reminderAt(in.Reminders[0].At),
				{RelativeTo: ReminderRelationDueDate, RelativePeriod: -3600},
				reminderAt(in.Reminders[1].At),
			},
		}
		n := completeRepeatingIn(task, time.UTC)
		assert.False(t, n.Done)
		assert.True(t, n.DueDate.Equal(res.Due.At))
		assert.True(t, n.StartDate.Equal(res.Start.At))
		assert.True(t, n.EndDate.Equal(res.End.At))
		assert.True(t, reminderTime(t, n, 0).Equal(res.Reminders[0].At))
		assert.True(t, reminderTime(t, n, 2).Equal(res.Reminders[1].At))
		assert.True(t, res.Start.At.Equal(time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)), "got %v", res.Start.At)
	})
}

// TestNextOccurrence_UnsetReminderIsNotStepped: a reminder whose At is zero is not set. It keeps
// its index, stays zero, and never reaches the stepping, which would read the clock for a zero
// DTSTART.
func TestNextOccurrence_UnsetReminderIsNotStepped(t *testing.T) {
	set := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)

	t.Run("no due date", func(t *testing.T) {
		in := nextOccurrenceInput{
			Rule:      "FREQ=MONTHLY;BYMONTHDAY=15",
			Reminders: []datedValue{{}, dated(set)},
			Now:       nextOccurrenceNow, Zone: time.UTC,
		}
		first := nextOccurrence(in)
		require.Empty(t, first.Reason)
		require.Len(t, first.Reminders, 2)
		assert.True(t, first.Reminders[0].At.IsZero(), "an unset reminder stays unset, got %v", first.Reminders[0].At)
		assert.True(t, first.Reminders[1].At.Equal(time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)), "got %v", first.Reminders[1].At)

		// rrule-go truncates DTSTART to the second, so a clock read would show after a second.
		time.Sleep(1100 * time.Millisecond)
		second := nextOccurrence(in)
		require.Len(t, second.Reminders, 2)
		for i := range first.Reminders {
			assert.True(t, first.Reminders[i].At.Equal(second.Reminders[i].At), "reminder %d: %v then %v", i, first.Reminders[i].At, second.Reminders[i].At)
		}
	})

	t.Run("no due date, from completion", func(t *testing.T) {
		res := nextOccurrence(nextOccurrenceInput{
			Rule:           "FREQ=MONTHLY;BYMONTHDAY=15",
			FromCompletion: true,
			Reminders:      []datedValue{{}, dated(set)},
			Now:            nextOccurrenceNow, Zone: time.UTC,
		})
		require.Empty(t, res.Reason)
		require.Len(t, res.Reminders, 2)
		assert.True(t, res.Reminders[0].At.IsZero(), "got %v", res.Reminders[0].At)
		assert.True(t, res.Reminders[1].At.Equal(time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)), "the set reminder is the earliest, got %v", res.Reminders[1].At)
	})

	t.Run("with a due date", func(t *testing.T) {
		res := nextOccurrence(nextOccurrenceInput{
			Rule:      "FREQ=MONTHLY;BYMONTHDAY=15",
			Due:       dated(time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)),
			Reminders: []datedValue{{}},
			Now:       time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC), Zone: time.UTC,
		})
		require.Empty(t, res.Reason, "an unset reminder must not make the call refuse")
		require.Len(t, res.Reminders, 1)
		assert.True(t, res.Reminders[0].At.IsZero(), "got %v", res.Reminders[0].At)
		assert.True(t, res.Due.At.Equal(time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)), "got %v", res.Due.At)
	})
}

// ---------------------------------------------------------------------------------------------
// The wall-clock engine. Golden times are New York unless stated. Inputs and expectations are
// written with their numeric UTC offset, so a time in a repeated (fall-back) hour is unambiguous;
// "[02:30]" in a comment is an output intent.
// ---------------------------------------------------------------------------------------------

func mustZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

// ts parses "2006-01-02 15:04 -07:00" (or with seconds) and reads the instant in zone.
func ts(t *testing.T, zone *time.Location, s string) time.Time {
	t.Helper()
	layout := "2006-01-02 15:04 -07:00"
	if len(s) == len("2006-01-02 15:04:05 -07:00") {
		layout = "2006-01-02 15:04:05 -07:00"
	}
	v, err := time.Parse(layout, s)
	require.NoError(t, err)
	return v.In(zone)
}

// tod parses a local time of day, "15:04" or "15:04:05"; "" is no intent.
func tod(t *testing.T, s string) *time.Duration {
	t.Helper()
	if s == "" {
		return nil
	}
	layout := "15:04"
	if len(s) == len("15:04:05") {
		layout = "15:04:05"
	}
	v, err := time.Parse(layout, s)
	require.NoError(t, err)
	d := time.Duration(v.Hour())*time.Hour + time.Duration(v.Minute())*time.Minute + time.Duration(v.Second())*time.Second
	return &d
}

// dz is a date with an intended clock time ("" = none). IntendedZone stays "": it is deprecated,
// never read and never set, so assertDated checks "" for free.
func dz(t *testing.T, at time.Time, intended string) datedValue {
	t.Helper()
	return datedValue{At: at, Intended: tod(t, intended)}
}

// assertDated checks one result date: the instant, the intended time, and that IntendedZone is "".
func assertDated(t *testing.T, label string, got, want datedValue) {
	t.Helper()
	if want.At.IsZero() {
		assert.True(t, got.At.IsZero(), "%s: want unset, got %v", label, got.At)
		return
	}
	assert.True(t, got.At.Equal(want.At), "%s: got %v (%v), want %v", label, got.At, got.At.UTC(), want.At)
	if want.Intended == nil {
		assert.Nil(t, got.Intended, "%s: want no intent, got %v", label, durString(got.Intended))
		assert.Empty(t, got.IntendedZone, "%s", label)
		return
	}
	if assert.NotNil(t, got.Intended, "%s: want intent %v", label, *want.Intended) {
		assert.Equal(t, *want.Intended, *got.Intended, "%s: intent", label)
	}
	assert.Equal(t, want.IntendedZone, got.IntendedZone, "%s: intent zone", label)
}

func durString(d *time.Duration) string {
	if d == nil {
		return "nil"
	}
	return d.String()
}

// assertResult checks a whole result against the expected one.
func assertResult(t *testing.T, got, want nextOccurrenceResult) {
	t.Helper()
	require.Equal(t, want.Reason, got.Reason)
	if want.Reason != "" {
		return
	}
	assertDated(t, "due", got.Due, want.Due)
	assertDated(t, "start", got.Start, want.Start)
	assertDated(t, "end", got.End, want.End)
	require.Len(t, got.Reminders, len(want.Reminders))
	for i := range want.Reminders {
		assertDated(t, "reminder "+strconv.Itoa(i), got.Reminders[i], want.Reminders[i])
	}
}

// fedBack is the input of the next completion: the result's dates, with their intent, completed
// at now in zone (the chain feeds each output Intended back in).
func fedBack(in nextOccurrenceInput, res nextOccurrenceResult, now time.Time, zone *time.Location) nextOccurrenceInput {
	next := in
	next.Due, next.Start, next.End = res.Due, res.Start, res.End
	next.Reminders = append([]datedValue(nil), res.Reminders...)
	next.Now, next.Zone = now, zone
	return next
}

// asStored reads every input date in the server zone, as dates loaded from the database are
// (xorm SetTZLocation): the result must depend on the instants and Zone only, never on the
// location the inputs carry.
func asStored(in nextOccurrenceInput) nextOccurrenceInput {
	server := config.GetTimeZone()
	stored := func(d datedValue) datedValue {
		if !d.At.IsZero() {
			d.At = d.At.In(server)
		}
		return d
	}
	in.Due, in.Start, in.End = stored(in.Due), stored(in.Start), stored(in.End)
	in.Reminders = append([]datedValue(nil), in.Reminders...)
	for i := range in.Reminders {
		in.Reminders[i] = stored(in.Reminders[i])
	}
	return in
}

// assertCompletionAgrees completes the same task through completion in the same zone and checks it
// stores exactly the helper's instants, in the server zone.
func assertCompletionAgrees(t *testing.T, in nextOccurrenceInput, res nextOccurrenceResult) {
	t.Helper()
	in = asStored(in)
	task := &Task{Repeats: in.Rule, RepeatsFromCurrentDate: in.FromCompletion, DueDate: in.Due.At, StartDate: in.Start.At, EndDate: in.End.At}
	for _, rem := range in.Reminders {
		task.Reminders = append(task.Reminders, reminderAt(rem.At))
	}
	withFixedNow(t, in.Now)
	n := completeRepeatingIn(task, in.Zone)
	if res.Reason != "" {
		assert.True(t, n.Done, "completion stays done as the helper refuses (%s)", res.Reason)
		return
	}
	assert.False(t, n.Done)
	server := config.GetTimeZone().String()
	for _, pair := range []struct {
		label     string
		got, want time.Time
	}{{"due", n.DueDate, res.Due.At}, {"start", n.StartDate, res.Start.At}, {"end", n.EndDate, res.End.At}} {
		assert.True(t, pair.got.Equal(pair.want), "completion %s %v, helper %v", pair.label, pair.got, pair.want)
		if !pair.got.IsZero() {
			assert.Equal(t, server, pair.got.Location().String(), "%s is stored in the server zone", pair.label)
		}
	}
	for i := range in.Reminders {
		got := reminderTime(t, n, i)
		assert.True(t, got.Equal(res.Reminders[i].At), "completion reminder %d %v, helper %v", i, got, res.Reminders[i].At)
		assert.Equal(t, server, got.Location().String(), "reminder %d is stored in the server zone", i)
	}
}

// --- explicit now and zone ---

// TestNextOccurrence_ProcessLocalIsIgnored: completion's "now" comes from the process clock, in
// time.Local. A from-completion rule used to be evaluated in that zone (rrule-go evaluates in
// DTSTART's location), so the host's zone decided the date. Now it is evaluated in the completion's
// zone. Sun 2026-10-25 20:00Z is already Monday 05:00 in Tokyo.
func TestNextOccurrence_ProcessLocalIsIgnored(t *testing.T) {
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	complete := func(local *time.Location) time.Time {
		withProcessZone(t, local)
		orig := timeNow
		timeNow = func() time.Time { return now.In(time.Local) } //nolint:gosmopolitan // the process clock reads in the process zone; that is the defect under test
		defer func() { timeNow = orig }()
		n := completeRepeatingIn(&Task{
			Repeats:                "FREQ=WEEKLY;BYDAY=MO",
			RepeatsFromCurrentDate: true,
			Reminders:              []*TaskReminder{reminderAt(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))},
		}, time.UTC)
		require.False(t, n.Done)
		return reminderTime(t, n, 0)
	}
	inUTC := complete(time.UTC)
	inTokyo := complete(mustZone(t, "Asia/Tokyo"))
	assert.True(t, inUTC.Equal(time.Date(2026, 10, 26, 20, 0, 0, 0, time.UTC)), "got %v", inUTC)
	assert.True(t, inTokyo.Equal(inUTC), "time.Local = Tokyo gave %v, UTC gave %v", inTokyo.UTC(), inUTC.UTC())
}

// TestNextOccurrence_CompletionReturnsServerZoneDates: the helper returns its dates in the
// completion's zone; completion stores and returns them in the server zone, as every date
// loaded from the database is, so v1, v2 and CalDAV strings keep their offset. The instants are
// the helper's.
func TestNextOccurrence_CompletionReturnsServerZoneDates(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }
	in := nextOccurrenceInput{
		Rule:      "FREQ=DAILY",
		Due:       dated(at("2027-03-13 20:00 -05:00")),
		Start:     dated(at("2027-03-13 08:00 -05:00")),
		End:       dated(at("2027-03-13 22:00 -05:00")),
		Reminders: []datedValue{dated(at("2027-03-13 19:00 -05:00"))},
		Now:       at("2027-03-13 21:00 -05:00"),
		Zone:      ny,
	}
	res := nextOccurrence(in)
	require.Empty(t, res.Reason)
	for _, d := range []datedValue{res.Due, res.Start, res.End, res.Reminders[0]} {
		assert.Equal(t, "America/New_York", d.At.Location().String(), "the helper returns dates in Zone")
	}
	assertCompletionAgrees(t, in, res)
}

// --- rule attributes ---

// TestNextOccurrenceRuleAttributes: the arithmetic family and the no-due-date policy are decided
// separately. The policy set is exactly the set the earlier fixed-duration engine stepped by a
// fixed duration.
func TestNextOccurrenceRuleAttributes(t *testing.T) {
	for _, tc := range []struct {
		rule   string
		family ruleFamily
		policy noDuePolicy
	}{
		{"FREQ=DAILY", familyWallclock, policyIndependent},
		{"FREQ=WEEKLY;INTERVAL=2", familyWallclock, policyIndependent},
		{"FREQ=DAILY;COUNT=3", familyWallclock, policyOneStep},
		{"FREQ=DAILY;UNTIL=20271231T000000Z", familyWallclock, policyOneStep},
		{"FREQ=DAILY;BYHOUR=9", familyWallclock, policyOneStep},
		{"FREQ=MONTHLY", familyWallclock, policyOneStep},
		{"FREQ=YEARLY", familyWallclock, policyOneStep},
		{"FREQ=HOURLY", familyElapsed, policyIndependent},
		{"FREQ=SECONDLY;INTERVAL=7200", familyElapsed, policyIndependent},
		{"FREQ=HOURLY;UNTIL=20271231T000000Z", familyElapsed, policyOneStep},
		{"FREQ=MINUTELY;COUNT=5", familyElapsed, policyOneStep},
	} {
		t.Run(tc.rule, func(t *testing.T) {
			opt, err := rrule.StrToROption(tc.rule)
			require.NoError(t, err)
			assert.Equal(t, tc.family, familyOf(opt), "family")
			assert.Equal(t, tc.policy, policyOf(opt), "policy")
		})
	}
}

// --- the floating timeline and localise ---

// TestLocaliseWorkedExamples: gap and overlap examples, on dates from 2000 on. localise never
// relies on Go's time.Date normalisation.
func TestLocaliseWorkedExamples(t *testing.T) {
	for _, tc := range []struct {
		zone  string
		asked time.Time // a floating wall clock
		want  time.Time // the instant
	}{
		{"America/New_York", time.Date(2027, 3, 14, 2, 30, 0, 0, time.UTC), time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC)},  // gap: 03:30 EDT
		{"America/New_York", time.Date(2026, 11, 1, 1, 30, 0, 0, time.UTC), time.Date(2026, 11, 1, 5, 30, 0, 0, time.UTC)},  // overlap: 01:30 EDT
		{"Europe/Berlin", time.Date(2027, 3, 28, 2, 30, 0, 0, time.UTC), time.Date(2027, 3, 28, 1, 30, 0, 0, time.UTC)},     // gap: 03:30 CEST
		{"Europe/Berlin", time.Date(2026, 10, 25, 2, 30, 0, 0, time.UTC), time.Date(2026, 10, 25, 0, 30, 0, 0, time.UTC)},   // overlap: 02:30 CEST
		{"Pacific/Auckland", time.Date(2026, 9, 27, 2, 30, 0, 0, time.UTC), time.Date(2026, 9, 26, 14, 30, 0, 0, time.UTC)}, // gap: 03:30 NZDT
		{"Pacific/Auckland", time.Date(2027, 4, 4, 2, 30, 0, 0, time.UTC), time.Date(2027, 4, 3, 13, 30, 0, 0, time.UTC)},   // overlap: 02:30 NZDT
		{"America/Santiago", time.Date(2026, 9, 6, 0, 30, 0, 0, time.UTC), time.Date(2026, 9, 6, 4, 30, 0, 0, time.UTC)},    // gap at midnight: 01:30 -03, same date
	} {
		t.Run(tc.zone+" "+tc.asked.Format("2006-01-02 15:04"), func(t *testing.T) {
			zone := mustZone(t, tc.zone)
			got := localise(tc.asked, zone)
			assert.True(t, got.Equal(tc.want), "got %v (%v), want %v", got, got.UTC(), tc.want.In(zone))
			assert.Equal(t, tc.zone, got.Location().String(), "the result is read in the zone")
		})
	}
}

// localiseTransition is one clock change: the local date, and the wall-clock hour it affects
// (skipped in a gap, repeated in an overlap), [from, from+length).
type localiseTransition struct {
	zone   string
	date   string // 2006-01-02, the local date of the affected wall clock
	from   string // 15:04
	length time.Duration
	gap    bool
}

// localiseTransitions are every 2026-2027 transition of the matrix's DST zones, plus Santiago's
// midnight gap and Nuuk's gap that crosses midnight.
var localiseTransitions = []localiseTransition{
	{"America/New_York", "2026-03-08", "02:00", time.Hour, true},
	{"America/New_York", "2026-11-01", "01:00", time.Hour, false},
	{"America/New_York", "2027-03-14", "02:00", time.Hour, true},
	{"America/New_York", "2027-11-07", "01:00", time.Hour, false},
	{"Europe/Berlin", "2026-03-29", "02:00", time.Hour, true},
	{"Europe/Berlin", "2026-10-25", "02:00", time.Hour, false},
	{"Europe/Berlin", "2027-03-28", "02:00", time.Hour, true},
	{"Europe/Berlin", "2027-10-31", "02:00", time.Hour, false},
	{"Pacific/Auckland", "2026-04-05", "02:00", time.Hour, false},
	{"Pacific/Auckland", "2026-09-27", "02:00", time.Hour, true},
	{"Pacific/Auckland", "2027-04-04", "02:00", time.Hour, false},
	{"Pacific/Auckland", "2027-09-26", "02:00", time.Hour, true},
	{"America/Santiago", "2026-09-06", "00:00", time.Hour, true},
	{"America/Nuuk", "2027-03-27", "23:00", time.Hour, true},
}

// TestLocaliseTransitions probes a time inside every gap and overlap, and one clear of it:
// a gap moves forward by its length from the pre-transition offset, an overlap takes the
// earlier of its two instants, and a clear time is the identity. Zones without DST, at
// non-hour offsets, are the identity on the same dates. float(localise(f)) == f whenever f exists.
func TestLocaliseTransitions(t *testing.T) {
	wall := func(t *testing.T, date, clock string) time.Time {
		v, err := time.Parse("2006-01-02 15:04", date+" "+clock)
		require.NoError(t, err)
		return v // UTC-located: a floating wall clock
	}
	offsetAt := func(zone *time.Location, instant time.Time) time.Duration {
		_, off := instant.In(zone).Zone()
		return time.Duration(off) * time.Second
	}
	for _, tr := range localiseTransitions {
		t.Run(tr.zone+" "+tr.date, func(t *testing.T) {
			zone := mustZone(t, tr.zone)
			inside := wall(t, tr.date, tr.from).Add(tr.length / 2)
			clearTime := wall(t, tr.date, tr.from).Add(-time.Hour)
			before := offsetAt(zone, inside.Add(-48*time.Hour))

			got := localise(inside, zone)
			if tr.gap {
				assert.True(t, got.Equal(inside.Add(-before)), "gap: got %v, want the pre-transition offset's instant %v", got, inside.Add(-before))
				assert.True(t, float(got, zone).Equal(inside.Add(tr.length)), "gap: wall clock %v, want %v", float(got, zone), inside.Add(tr.length))
			} else {
				assert.True(t, float(got, zone).Equal(inside), "overlap: wall clock kept, got %v", float(got, zone))
				later := got.Add(tr.length)
				assert.True(t, float(later, zone).Equal(inside), "overlap: the later instant has the same wall clock")
				assert.False(t, float(got.Add(-tr.length), zone).Equal(inside), "overlap: the result is the earlier of the two")
			}

			gotClear := localise(clearTime, zone)
			assert.True(t, float(gotClear, zone).Equal(clearTime), "clear of the change: identity, got %v", gotClear)
			assert.True(t, gotClear.Equal(clearTime.Add(-offsetAt(zone, gotClear))), "clear of the change: one instant")
		})
	}
	for _, name := range []string{"Asia/Kolkata", "Asia/Kathmandu", "UTC"} {
		zone := mustZone(t, name)
		for _, tr := range localiseTransitions {
			f := wall(t, tr.date, tr.from).Add(tr.length / 2)
			got := localise(f, zone)
			assert.True(t, float(got, zone).Equal(f), "%s %v: identity, got %v", name, f, got)
		}
	}
}

// --- wall-clock stepping and one-pass selection ---

// TestNextOccurrenceWallClock: calendar rules are evaluated on the zone's calendar and wall clock,
// sub-daily ones stay elapsed, and each result is normalised. Each case also goes through
// completion in the same zone.
func TestNextOccurrenceWallClock(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }
	in := func(zone *time.Location, rule string, due, now time.Time) nextOccurrenceInput {
		return nextOccurrenceInput{Rule: rule, Due: dated(due), Now: now, Zone: zone}
	}
	berlin := mustZone(t, "Europe/Berlin")
	auckland := mustZone(t, "Pacific/Auckland")
	santiago := mustZone(t, "America/Santiago")
	cases := []struct {
		name string
		in   nextOccurrenceInput
		want nextOccurrenceResult
	}{
		{"Mon/Wed/Fri", in(ny, "FREQ=WEEKLY;BYDAY=MO,WE,FR", at("2026-10-26 20:00 -04:00"), at("2026-10-26 20:05 -04:00")),
			nextOccurrenceResult{Due: dated(at("2026-10-28 20:00 -04:00"))}},
		{"the 15th", in(ny, "FREQ=MONTHLY;BYMONTHDAY=15", at("2026-10-15 21:00 -04:00"), at("2026-10-15 21:05 -04:00")),
			nextOccurrenceResult{Due: dated(at("2026-11-15 21:00 -05:00"))}},
		{"second Tuesday", in(ny, "FREQ=MONTHLY;BYDAY=2TU", at("2026-10-13 22:00 -04:00"), at("2026-10-13 22:05 -04:00")),
			nextOccurrenceResult{Due: dated(at("2026-11-10 22:00 -05:00"))}},
		{"daily, spring forward", in(ny, "FREQ=DAILY", at("2027-03-13 20:00 -05:00"), at("2027-03-13 20:05 -05:00")),
			nextOccurrenceResult{Due: dated(at("2027-03-14 20:00 -04:00"))}},
		{"weekly, fall back", in(ny, "FREQ=WEEKLY", at("2026-10-31 20:00 -04:00"), at("2026-10-31 20:05 -04:00")),
			nextOccurrenceResult{Due: dated(at("2026-11-07 20:00 -05:00"))}},
		{"Auckland, spring forward", in(auckland, "FREQ=DAILY", ts(t, auckland, "2026-09-26 20:00 +12:00"), ts(t, auckland, "2026-09-26 20:05 +12:00")),
			nextOccurrenceResult{Due: dated(ts(t, auckland, "2026-09-27 20:00 +13:00"))}},
		{"Auckland, fall back", in(auckland, "FREQ=DAILY", ts(t, auckland, "2027-04-03 20:00 +13:00"), ts(t, auckland, "2027-04-03 20:05 +13:00")),
			nextOccurrenceResult{Due: dated(ts(t, auckland, "2027-04-04 20:00 +12:00"))}},
		{"HOURLY;INTERVAL=24 is elapsed", in(ny, "FREQ=HOURLY;INTERVAL=24", at("2027-03-13 20:00 -05:00"), at("2027-03-13 20:05 -05:00")),
			nextOccurrenceResult{Due: dated(at("2027-03-14 21:00 -04:00"))}},
		{"a gap moves forward by the gap", in(ny, "FREQ=DAILY", at("2027-03-13 02:30 -05:00"), at("2027-03-13 03:00 -05:00")),
			nextOccurrenceResult{Due: dz(t, at("2027-03-14 03:30 -04:00"), "02:30")}},
		{"an overlap takes the earlier instant", in(ny, "FREQ=DAILY", at("2026-10-31 01:30 -04:00"), at("2026-10-31 01:35 -04:00")),
			nextOccurrenceResult{Due: dated(at("2026-11-01 01:30 -04:00"))}},
		{"east, Berlin gap", in(berlin, "FREQ=DAILY", ts(t, berlin, "2027-03-27 02:30 +01:00"), ts(t, berlin, "2027-03-27 02:35 +01:00")),
			nextOccurrenceResult{Due: dz(t, ts(t, berlin, "2027-03-28 03:30 +02:00"), "02:30")}},
		{"east, Berlin overlap", in(berlin, "FREQ=DAILY", ts(t, berlin, "2026-10-24 02:30 +02:00"), ts(t, berlin, "2026-10-24 02:35 +02:00")),
			nextOccurrenceResult{Due: dated(ts(t, berlin, "2026-10-25 02:30 +02:00"))}},
		{"midnight gap, Santiago", in(santiago, "FREQ=DAILY", ts(t, santiago, "2026-09-05 00:30 -04:00"), ts(t, santiago, "2026-09-05 00:35 -04:00")),
			nextOccurrenceResult{Due: dz(t, ts(t, santiago, "2026-09-06 01:30 -03:00"), "00:30")}},
		{"from completion keeps the completion's clock", nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true, Due: dated(at("2027-03-13 20:00 -05:00")), Now: at("2027-03-13 19:30 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-14 19:30 -04:00"))}},
		{"monthly: the earliest normalised instant wins", in(ny, "FREQ=MONTHLY;BYMONTHDAY=14;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3", at("2027-02-14 03:00 -05:00"), at("2027-03-01 12:00 -05:00")),
			nextOccurrenceResult{Due: dated(at("2027-03-14 03:00 -04:00"))}},
		{"weekly: the earliest normalised instant wins", in(ny, "FREQ=WEEKLY;BYDAY=SU;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3", at("2027-03-07 03:00 -05:00"), at("2027-03-08 12:00 -05:00")),
			nextOccurrenceResult{Due: dated(at("2027-03-14 03:00 -04:00"))}},
		{"no due date, independent: the start jumps past now on the wall clock", nextOccurrenceInput{Rule: "FREQ=DAILY", Start: dated(at("2027-03-13 02:30 -05:00")), Now: at("2027-03-13 12:00 -05:00"), Zone: ny},
			nextOccurrenceResult{Start: dated(at("2027-03-14 03:30 -04:00"))}}, // a companion: no intent
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := nextOccurrence(asStored(tc.in))
			assertResult(t, res, tc.want)
			assertCompletionAgrees(t, tc.in, res)
		})
	}

	t.Run("Mon/Wed/Fri in the server zone gives the server's calendar", func(t *testing.T) {
		// The same instant completed in GMT is a Tuesday there, so the next is Wednesday 00:00Z,
		// which is Tuesday 20:00 in New York: the defect the user's zone fixes.
		res := nextOccurrence(asStored(in(time.UTC, "FREQ=WEEKLY;BYDAY=MO,WE,FR", at("2026-10-26 20:00 -04:00"), at("2026-10-26 20:05 -04:00"))))
		assertResult(t, res, nextOccurrenceResult{Due: dated(at("2026-10-27 20:00 -04:00"))})
	})

	t.Run("no DST, non-hour offsets keep the local clock every step", func(t *testing.T) {
		for _, name := range []string{"Asia/Kolkata", "Asia/Kathmandu"} {
			zone := mustZone(t, name)
			due := time.Date(2026, 10, 26, 20, 0, 0, 0, zone)
			for _, step := range []struct {
				rule string
				want time.Time
			}{
				{"FREQ=DAILY", time.Date(2026, 10, 27, 20, 0, 0, 0, zone)},
				{"FREQ=WEEKLY", time.Date(2026, 11, 2, 20, 0, 0, 0, zone)},
				{"FREQ=MONTHLY", time.Date(2026, 11, 26, 20, 0, 0, 0, zone)},
				{"FREQ=MONTHLY;BYMONTHDAY=26", time.Date(2026, 11, 26, 20, 0, 0, 0, zone)},
			} {
				res := nextOccurrence(asStored(in(zone, step.rule, due, due.Add(time.Minute))))
				assertResult(t, res, nextOccurrenceResult{Due: dated(step.want)})
			}
		}
	})
}

// TestNextOccurrenceServerZoneControl: in the server zone, on server-zone due dates, the weekly
// and monthly rules above give exactly the earlier fixed-duration engine's results. The values were
// captured from the code before this change, with the same instants as the New York cases.
func TestNextOccurrenceServerZoneControl(t *testing.T) {
	gmt := config.GetTimeZone()
	for _, tc := range []struct {
		rule      string
		due, want time.Time
	}{
		{"FREQ=WEEKLY;BYDAY=MO,WE,FR", time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)},
		{"FREQ=MONTHLY;BYMONTHDAY=15", time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC), time.Date(2026, 11, 15, 1, 0, 0, 0, time.UTC)},
		{"FREQ=MONTHLY;BYDAY=2TU", time.Date(2026, 10, 14, 2, 0, 0, 0, time.UTC), time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.rule, func(t *testing.T) {
			in := nextOccurrenceInput{Rule: tc.rule, Due: dated(tc.due.In(gmt)), Now: tc.due.Add(5 * time.Minute), Zone: gmt}
			res := nextOccurrence(in)
			assertResult(t, res, nextOccurrenceResult{Due: dated(tc.want)})
			withFixedNow(t, in.Now)
			n := completeRepeating(&Task{Repeats: tc.rule, DueDate: tc.due.In(gmt)})
			assert.True(t, n.DueDate.Equal(tc.want), "completion in the server zone: got %v", n.DueDate)
		})
	}
}

// --- UNTIL on normalised instants ---

// TestNextOccurrenceUntil: UNTIL is compared with each occurrence's normalised instant, at second
// precision: around an overlap an occurrence within UNTIL is kept although its floating value is
// past it, and through a gap one past UNTIL is dropped although its floating value is before it.
func TestNextOccurrenceUntil(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }
	for _, tc := range []struct {
		name     string
		rule     string
		due, now time.Time
		want     time.Time
		noneLeft bool
	}{
		{"UNTIL in the second pass of the overlap keeps 01:30 EDT", "FREQ=DAILY;UNTIL=20261101T061500Z", at("2026-10-31 01:30 -04:00"), at("2026-10-31 02:00 -04:00"), at("2026-11-01 01:30 -04:00"), false},
		{"UNTIL in the first pass, before 01:30 EDT", "FREQ=DAILY;UNTIL=20261101T051500Z", at("2026-10-31 01:30 -04:00"), at("2026-10-31 02:00 -04:00"), time.Time{}, true},
		{"UNTIL exactly at the occurrence is included", "FREQ=DAILY;UNTIL=20261101T053000Z", at("2026-10-31 01:30 -04:00"), at("2026-10-31 02:00 -04:00"), at("2026-11-01 01:30 -04:00"), false},
		{"a gap shift past UNTIL is excluded", "FREQ=DAILY;UNTIL=20270314T071500Z", at("2027-03-13 02:30 -05:00"), at("2027-03-13 03:00 -05:00"), time.Time{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := nextOccurrenceInput{Rule: tc.rule, Due: dated(tc.due), Now: tc.now, Zone: ny}
			res := nextOccurrence(asStored(in))
			if tc.noneLeft {
				assertResult(t, res, nextOccurrenceResult{Reason: reasonNoOccurrence})
			} else {
				assertResult(t, res, nextOccurrenceResult{Due: dated(tc.want)})
			}
			assertCompletionAgrees(t, in, res)
		})
	}
}

// --- chronological COUNT ---

// TestNextOccurrenceCount: COUNT counts the distinct normalised instants from DTSTART in time
// order; every occurrence from DTSTART is ranked, and a COUNT rule is never anchor-clamped.
func TestNextOccurrenceCount(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }

	t.Run("a clamped anchor must not revive an exhausted COUNT", func(t *testing.T) {
		// The last of 1000 daily occurrences is about 2.7 years after the due date, already past.
		// The earlier fixed-duration engine clamped the anchor to within a year of now and counted
		// from there.
		now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
		in := nextOccurrenceInput{Rule: "FREQ=DAILY;COUNT=1000", Due: dated(time.Date(2023, 10, 22, 9, 0, 0, 0, time.UTC)), Now: now, Zone: time.UTC}
		res := nextOccurrence(asStored(in))
		assertResult(t, res, nextOccurrenceResult{Reason: reasonNoOccurrence})
		assertCompletionAgrees(t, in, res)
	})
	t.Run("ranks count from DTSTART, not from the look-back skip", func(t *testing.T) {
		in := nextOccurrenceInput{Rule: "FREQ=DAILY;COUNT=5", Due: dated(time.Date(2027, 1, 1, 9, 0, 0, 0, time.UTC)), Now: time.Date(2027, 1, 5, 12, 0, 0, 0, time.UTC), Zone: time.UTC}
		assertResult(t, nextOccurrence(in), nextOccurrenceResult{Reason: reasonNoOccurrence})
	})

	cases := []struct {
		name      string
		in        nextOccurrenceInput
		want      nextOccurrenceResult
		wantCount int
	}{
		{"a collision counts once", nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=30;COUNT=3", Due: dated(at("2027-03-13 03:30 -05:00")), Now: at("2027-03-14 03:45 -04:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-15 02:30 -04:00"))}, 2},
		{"COUNT follows time, not floating order", nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3;COUNT=2", Due: dated(at("2027-03-13 03:00 -05:00")), Now: at("2027-03-13 03:15 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-14 03:00 -04:00"))}, 1},
		{"collision representative (no COUNT)", nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=30", Due: dated(at("2027-03-13 03:30 -05:00")), Start: dated(at("2027-03-13 02:30 -05:00")), Now: at("2027-03-13 04:00 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-14 03:30 -04:00")), Start: dated(at("2027-03-14 03:30 -04:00"))}, 0},
		{"elapsed COUNT keeps native COUNT on UTC", nextOccurrenceInput{Rule: "FREQ=HOURLY;COUNT=5", Due: dated(at("2027-03-14 01:30 -05:00")), Start: dated(at("2027-03-14 01:00 -05:00")), Now: at("2027-03-14 01:45 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-14 03:30 -04:00")), Start: dated(at("2027-03-14 03:00 -04:00"))}, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := nextOccurrence(asStored(tc.in))
			assertResult(t, res, tc.want)
			assertCompletionAgrees(t, tc.in, res)
			if tc.wantCount > 0 {
				withFixedNow(t, tc.in.Now)
				n := completeRepeatingIn(&Task{Repeats: tc.in.Rule, DueDate: tc.in.Due.At, StartDate: tc.in.Start.At}, tc.in.Zone)
				opt, err := rrule.StrToROption(n.Repeats)
				require.NoError(t, err)
				assert.Equal(t, tc.wantCount, opt.Count, "stored COUNT after completion")
			}
		})
	}
}

// --- companions by family ---

// TestNextOccurrenceCompanions: a companion keeps its own local clock time and moves by whole
// local days (wallclock), keeps its calendar offset when the rule names a time, or moves by
// the reference's elapsed delta (sub-daily). It carries no intent.
func TestNextOccurrenceCompanions(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }
	cases := []struct {
		name string
		in   nextOccurrenceInput
		want nextOccurrenceResult
	}{
		{"plain daily across spring forward", nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-13 20:00 -05:00")), Start: dated(at("2027-03-13 08:00 -05:00")), Now: at("2027-03-13 21:00 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-14 20:00 -04:00")), Start: dated(at("2027-03-14 08:00 -04:00"))}},
		{"from completion: each companion keeps its own clock", nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true,
			Due: dated(at("2026-10-26 20:00 -04:00")), Start: dated(at("2026-10-26 08:00 -04:00")), Reminders: []datedValue{dated(at("2026-10-26 19:00 -04:00"))}, Now: at("2026-10-27 19:30 -04:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2026-10-28 19:30 -04:00")), Start: dated(at("2026-10-28 08:00 -04:00")), Reminders: []datedValue{dated(at("2026-10-28 19:00 -04:00"))}}},
		{"from completion across the fall back", nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true,
			Due: dated(at("2026-10-30 20:00 -04:00")), Start: dated(at("2026-10-30 08:00 -04:00")), Now: at("2026-10-31 19:30 -04:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2026-11-01 19:30 -05:00")), Start: dated(at("2026-11-01 08:00 -05:00"))}},
		{"BYHOUR: the calendar offset is kept", nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=9,17",
			Due: dated(at("2026-10-26 17:00 -04:00")), Start: dated(at("2026-10-26 16:00 -04:00")), End: dated(at("2026-10-26 18:00 -04:00")), Now: at("2026-10-26 17:30 -04:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2026-10-27 09:00 -04:00")), Start: dated(at("2026-10-27 08:00 -04:00")), End: dated(at("2026-10-27 10:00 -04:00"))}},
		{"month end", nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYMONTHDAY=28,29,30,31;BYSETPOS=-1",
			Due: dated(at("2027-01-31 20:00 -05:00")), Start: dated(at("2027-01-15 09:00 -05:00")), Now: at("2027-01-31 21:00 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-02-28 20:00 -05:00")), Start: dated(at("2027-02-12 09:00 -05:00"))}},
		{"hourly: companions move by the elapsed delta", nextOccurrenceInput{Rule: "FREQ=HOURLY",
			Due: dated(at("2026-11-01 01:30 -04:00")), Start: dated(at("2026-11-01 00:30 -04:00")), Now: at("2026-11-01 01:40 -04:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2026-11-01 01:30 -05:00")), Start: dated(at("2026-11-01 01:30 -04:00"))}},
		{"hourly, no due date, fall back: each date jumps by elapsed hours", nextOccurrenceInput{Rule: "FREQ=HOURLY",
			Start: dated(at("2026-11-01 00:30 -04:00")), End: dated(at("2026-11-01 01:30 -04:00")), Now: at("2026-11-01 01:40 -04:00"), Zone: ny},
			nextOccurrenceResult{Start: dated(at("2026-11-01 01:30 -05:00")), End: dated(at("2026-11-01 01:30 -05:00"))}},
		{"hourly, no due date, spring forward", nextOccurrenceInput{Rule: "FREQ=HOURLY",
			Start: dated(at("2027-03-14 01:00 -05:00")), End: dated(at("2027-03-14 01:30 -05:00")), Now: at("2027-03-14 01:45 -05:00"), Zone: ny},
			nextOccurrenceResult{Start: dated(at("2027-03-14 03:00 -04:00")), End: dated(at("2027-03-14 03:30 -04:00"))}},
		{"weekly span across spring forward", nextOccurrenceInput{Rule: "FREQ=WEEKLY;BYDAY=SA",
			Due: dated(at("2027-03-06 22:00 -05:00")), Start: dated(at("2027-03-06 22:00 -05:00")), End: dated(at("2027-03-08 06:00 -05:00")), Now: at("2027-03-06 23:00 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2027-03-13 22:00 -05:00")), Start: dated(at("2027-03-13 22:00 -05:00")), End: dated(at("2027-03-15 06:00 -04:00"))}},
		{"one-step end keeps its whole-day gap", nextOccurrenceInput{Rule: "FREQ=WEEKLY;BYDAY=SA",
			Start: dated(at("2027-03-06 22:00 -05:00")), End: dated(at("2027-03-08 06:00 -05:00")), Now: at("2027-03-06 23:00 -05:00"), Zone: ny},
			nextOccurrenceResult{Start: dated(at("2027-03-13 22:00 -05:00")), End: dated(at("2027-03-15 06:00 -04:00"))}},
		{"fall back: calendar order, no clamp", nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=1;BYMINUTE=30",
			Due: dated(at("2026-11-01 01:30 -05:00")), Start: dated(at("2026-11-01 01:45 -04:00")), Now: at("2026-11-01 02:30 -05:00"), Zone: ny},
			nextOccurrenceResult{Due: dated(at("2026-11-02 01:30 -05:00")), Start: dated(at("2026-11-02 01:45 -05:00"))}},
		{"no due date, independent (ratchet)", nextOccurrenceInput{Rule: "FREQ=DAILY",
			Start: dated(time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)), End: dated(time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)), Now: time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC), Zone: time.UTC},
			nextOccurrenceResult{Start: dated(time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)), End: dated(time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC))}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := nextOccurrence(asStored(tc.in))
			assertResult(t, res, tc.want)
			assertCompletionAgrees(t, tc.in, res)
		})
	}

	t.Run("a companion in a gap keeps the shifted time at the next step (a known limit)", func(t *testing.T) {
		in := nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-13 20:00 -05:00")), Start: dated(at("2027-03-13 02:30 -05:00")), Now: at("2027-03-13 21:00 -05:00"), Zone: ny}
		first := nextOccurrence(asStored(in))
		assertResult(t, first, nextOccurrenceResult{Due: dated(at("2027-03-14 20:00 -04:00")), Start: dated(at("2027-03-14 03:30 -04:00"))})
		second := nextOccurrence(asStored(fedBack(in, first, at("2027-03-14 21:00 -04:00"), ny)))
		assertResult(t, second, nextOccurrenceResult{Due: dated(at("2027-03-15 20:00 -04:00")), Start: dated(at("2027-03-15 03:30 -04:00"))})
	})
}

// --- the intended clock time in the helper ---

// TestNextOccurrenceIntended chains completions, feeding each output intent back in: a gap-shifted
// due date returns to its intended clock time at the next step when the gap check accepts
// it. Companions carry no intent, so a gap-shifted or clamped one keeps the shifted time (a known
// limit).
func TestNextOccurrenceIntended(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	at := func(s string) time.Time { return ts(t, ny, s) }

	type step struct {
		now  time.Time
		zone *time.Location
		want nextOccurrenceResult
	}
	chain := func(t *testing.T, in nextOccurrenceInput, steps ...step) {
		t.Helper()
		in.Now, in.Zone = steps[0].now, steps[0].zone
		res := nextOccurrence(asStored(in))
		assertResult(t, res, steps[0].want)
		for i, s := range steps[1:] {
			in = fedBack(in, res, s.now, s.zone)
			res = nextOccurrence(asStored(in))
			t.Logf("step %d", i+2)
			assertResult(t, res, s.want)
		}
	}

	t.Run("the due date returns to 02:30", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-13 02:30 -05:00"))},
			step{at("2027-03-13 03:00 -05:00"), ny, nextOccurrenceResult{Due: dz(t, at("2027-03-14 03:30 -04:00"), "02:30")}},
			step{at("2027-03-14 03:35 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 02:30 -04:00"))}})
	})
	t.Run("in Chicago the gap check rejects the intent (g = 0)", func(t *testing.T) {
		// 07:30Z is 01:30 CST in Chicago, whose clocks change only at 08:00Z: no gap in the 24 h
		// before, so the intent 02:30 is ignored and the due date keeps its Chicago clock.
		chicago := mustZone(t, "America/Chicago")
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-13 02:30 -05:00"))},
			step{at("2027-03-13 03:00 -05:00"), ny, nextOccurrenceResult{Due: dz(t, at("2027-03-14 03:30 -04:00"), "02:30")}},
			step{at("2027-03-14 03:35 -04:00"), chicago, nextOccurrenceResult{Due: dated(ts(t, chicago, "2027-03-15 01:30 -05:00"))}})
	})
	t.Run("companions keep their gap-shifted clocks (a known limit)", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-13 20:00 -05:00")), Start: dated(at("2027-03-13 02:30 -05:00")),
			End: dated(at("2027-03-13 02:45 -05:00")), Reminders: []datedValue{dated(at("2027-03-13 02:15 -05:00"))}},
			step{at("2027-03-13 20:05 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-14 20:00 -04:00")),
				Start: dated(at("2027-03-14 03:30 -04:00")), End: dated(at("2027-03-14 03:45 -04:00")),
				Reminders: []datedValue{dated(at("2027-03-14 03:15 -04:00"))}}},
			step{at("2027-03-14 20:05 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 20:00 -04:00")),
				Start: dated(at("2027-03-15 03:30 -04:00")), End: dated(at("2027-03-15 03:45 -04:00")),
				Reminders: []datedValue{dated(at("2027-03-15 03:15 -04:00"))}}})
	})
	t.Run("overnight span, end in the gap keeps the shifted time (a known limit)", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-12 23:00 -05:00")), Start: dated(at("2027-03-12 22:00 -05:00")), End: dated(at("2027-03-13 02:30 -05:00"))},
			step{at("2027-03-12 23:30 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-13 23:00 -05:00")), Start: dated(at("2027-03-13 22:00 -05:00")),
				End: dated(at("2027-03-14 03:30 -04:00"))}},
			step{at("2027-03-13 23:30 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-14 23:00 -04:00")), Start: dated(at("2027-03-14 22:00 -04:00")),
				End: dated(at("2027-03-15 03:30 -04:00"))}})
	})
	t.Run("partially named times keep the remembered components", func(t *testing.T) {
		for _, tc := range []struct{ rule, due, sun, intent, mon string }{
			{"FREQ=DAILY;BYMINUTE=30", "2027-03-13 02:30 -05:00", "2027-03-14 03:30 -04:00", "02:30", "2027-03-15 02:30 -04:00"},
			{"FREQ=DAILY;BYSECOND=15", "2027-03-13 02:30:15 -05:00", "2027-03-14 03:30:15 -04:00", "02:30:15", "2027-03-15 02:30:15 -04:00"},
			{"FREQ=DAILY;BYHOUR=2", "2027-03-13 02:30 -05:00", "2027-03-14 03:30 -04:00", "02:30", "2027-03-15 02:30 -04:00"},
			{"FREQ=DAILY;BYHOUR=2;BYMINUTE=30", "2027-03-13 02:30 -05:00", "2027-03-14 03:30 -04:00", "02:30", "2027-03-15 02:30 -04:00"}, // a named time in the gap
		} {
			t.Run(tc.rule, func(t *testing.T) {
				chain(t, nextOccurrenceInput{Rule: tc.rule, Due: dated(at(tc.due))},
					step{at("2027-03-13 03:00 -05:00"), ny, nextOccurrenceResult{Due: dz(t, at(tc.sun), tc.intent)}},
					step{at("2027-03-14 04:00 -04:00"), ny, nextOccurrenceResult{Due: dated(at(tc.mon))}})
			})
		}
	})
	t.Run("a clamp on a shifted reference keeps the clamped time (a known limit)", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=2;BYMINUTE=30", Due: dated(at("2027-03-13 02:30 -05:00")), End: dated(at("2027-03-13 03:00 -05:00"))},
			step{at("2027-03-13 04:00 -05:00"), ny, nextOccurrenceResult{Due: dz(t, at("2027-03-14 03:30 -04:00"), "02:30"), End: dated(at("2027-03-14 03:30 -04:00"))}},
			step{at("2027-03-14 04:00 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 02:30 -04:00")), End: dated(at("2027-03-15 03:30 -04:00"))}})
	})
	t.Run("a clamp on a shifted companion keeps the clamped time (a known limit)", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=3;BYMINUTE=10", Due: dated(at("2027-03-13 03:10 -05:00")), Start: dated(at("2027-03-13 02:40 -05:00"))},
			step{at("2027-03-13 04:00 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-14 03:10 -04:00")), Start: dated(at("2027-03-14 03:10 -04:00"))}},
			step{at("2027-03-14 04:00 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 03:10 -04:00")), Start: dated(at("2027-03-15 03:10 -04:00"))}})
	})
	t.Run("Nuuk: the intended civil date is recovered", func(t *testing.T) {
		nuuk := mustZone(t, "America/Nuuk")
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY;INTERVAL=2", Due: dated(ts(t, nuuk, "2027-03-25 23:30 -02:00"))},
			step{ts(t, nuuk, "2027-03-25 23:35 -02:00"), nuuk, nextOccurrenceResult{Due: dz(t, ts(t, nuuk, "2027-03-28 00:30 -01:00"), "23:30")}},
			step{ts(t, nuuk, "2027-03-28 00:35 -01:00"), nuuk, nextOccurrenceResult{Due: dated(ts(t, nuuk, "2027-03-29 23:30 -01:00"))}})
	})
	t.Run("Santiago midnight gap", func(t *testing.T) {
		santiago := mustZone(t, "America/Santiago")
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(ts(t, santiago, "2026-09-05 00:30 -04:00"))},
			step{ts(t, santiago, "2026-09-05 00:35 -04:00"), santiago, nextOccurrenceResult{Due: dz(t, ts(t, santiago, "2026-09-06 01:30 -03:00"), "00:30")}},
			step{ts(t, santiago, "2026-09-06 01:35 -03:00"), santiago, nextOccurrenceResult{Due: dated(ts(t, santiago, "2026-09-07 00:30 -03:00"))}})
	})
	t.Run("no gap, no intent; elapsed rules never carry intent", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(at("2027-03-13 20:00 -05:00"))},
			step{at("2027-03-13 20:05 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-14 20:00 -04:00"))}})
		chain(t, nextOccurrenceInput{Rule: "FREQ=HOURLY;INTERVAL=24", Due: dated(at("2027-03-13 02:30 -05:00")), Start: dated(at("2027-03-13 02:15 -05:00"))},
			step{at("2027-03-13 03:00 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-14 03:30 -04:00")), Start: dated(at("2027-03-14 03:15 -04:00"))}})
	})
	t.Run("from completion: the re-based due date takes now's clock and no intent", func(t *testing.T) {
		in := nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true, Due: dz(t, at("2027-03-14 03:30 -04:00"), "02:30")}
		chain(t, in, step{at("2027-03-14 09:00 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 09:00 -04:00"))}})
	})
	t.Run("from completion, no due date: colliding reminders separate again", func(t *testing.T) {
		// The earliest reminder by input order is the one reference; the
		// other reminder at the same instant is a companion and keeps its own clock (03:30, the
		// shifted time: a known limit), so the two separate again instead of collapsing.
		in := nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true,
			Reminders: []datedValue{dated(at("2027-03-13 01:30 -05:00")), dated(at("2027-03-13 02:30 -05:00"))}}
		first := in
		first.Now, first.Zone = at("2027-03-13 03:30 -05:00"), ny
		assertCompletionAgrees(t, first, nextOccurrence(asStored(first)))
		chain(t, in,
			step{at("2027-03-13 03:30 -05:00"), ny, nextOccurrenceResult{Reminders: []datedValue{
				dated(at("2027-03-14 03:30 -04:00")), dated(at("2027-03-14 03:30 -04:00"))}}},
			step{at("2027-03-14 04:00 -04:00"), ny, nextOccurrenceResult{Reminders: []datedValue{
				dated(at("2027-03-15 04:00 -04:00")), dated(at("2027-03-15 03:30 -04:00"))}}})
	})
	t.Run("an intent beyond the gap is ignored", func(t *testing.T) {
		// 03:30 EDT with intent 01:00: a displacement of 2.5 h exceeds the 1 h gap.
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dz(t, at("2027-03-14 03:30 -04:00"), "01:00")},
			step{at("2027-03-14 03:35 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 03:30 -04:00"))}})
	})
	t.Run("an intent far from any clock change is ignored (g = 0)", func(t *testing.T) {
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dz(t, at("2027-06-14 03:30 -04:00"), "02:30")},
			step{at("2027-06-14 03:35 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-06-15 03:30 -04:00"))}})
	})
	t.Run("IntendedZone is ignored on input (the gap check has no zone condition)", func(t *testing.T) {
		due := dz(t, at("2027-03-14 03:30 -04:00"), "02:30")
		due.IntendedZone = "Europe/Berlin"
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: due},
			step{at("2027-03-14 03:35 -04:00"), ny, nextOccurrenceResult{Due: dated(at("2027-03-15 02:30 -04:00"))}})
	})
	t.Run("an intent on a fall-back day is ignored (g = 0)", func(t *testing.T) {
		// The offset fell by an hour in the 24 h before 08:30Z, so g = 0 and nothing is accepted.
		chain(t, nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dz(t, at("2026-11-01 03:30 -05:00"), "02:30")},
			step{at("2026-11-01 03:35 -05:00"), ny, nextOccurrenceResult{Due: dated(at("2026-11-02 03:30 -05:00"))}})
	})
}

// --- zone x transition matrix ---

// TestNextOccurrenceZoneMatrix runs the engine's scenarios in UTC, America/New_York,
// Europe/Berlin, Asia/Kolkata and Pacific/Auckland, across every 2026-2027 transition of the DST
// zones. New York's values are the literal goldens of the tests above; here each zone asserts the
// property: the same local clock, a gap moves forward, where the due date returns at the next
// step and a companion keeps the shifted time (a known limit), an overlap takes the earlier
// instant, and sub-daily rules are elapsed. The zones without DST see the identity on the same
// dates.
func TestNextOccurrenceZoneMatrix(t *testing.T) {
	nyDates := []string{"2026-03-08", "2026-11-01", "2027-03-14", "2027-11-07"}
	matrix := []struct {
		zone  string
		dates []string
	}{
		{"UTC", nyDates},
		{"America/New_York", nyDates},
		{"Europe/Berlin", []string{"2026-03-29", "2026-10-25", "2027-03-28", "2027-10-31"}},
		{"Asia/Kolkata", nyDates},
		{"Pacific/Auckland", []string{"2026-04-05", "2026-09-27", "2027-04-04", "2027-09-26"}},
	}
	for _, m := range matrix {
		zone := mustZone(t, m.zone)
		for _, date := range m.dates {
			t.Run(m.zone+" "+date, func(t *testing.T) {
				d, err := time.Parse("2006-01-02", date)
				require.NoError(t, err)
				tr := transitionOn(zone, d)
				local := func(day, hour, minute int) time.Time {
					return time.Date(d.Year(), d.Month(), d.Day()+day, hour, minute, 0, 0, zone)
				}
				wall := func(day int, clock time.Duration) time.Time {
					return time.Date(d.Year(), d.Month(), d.Day()+day, 0, 0, 0, 0, time.UTC).Add(clock)
				}
				complete := func(in nextOccurrenceInput) nextOccurrenceResult {
					in.Zone = zone
					res := nextOccurrence(asStored(in))
					require.Empty(t, res.Reason)
					return res
				}

				// The clock is kept across the change, daily and weekly.
				for _, c := range []struct {
					rule string
					days int
				}{{"FREQ=DAILY", 1}, {"FREQ=WEEKLY", 7}} {
					due := local(-c.days, 20, 0)
					res := complete(nextOccurrenceInput{Rule: c.rule, Due: dated(due), Now: due.Add(5 * time.Minute)})
					assertDated(t, c.rule, res.Due, dated(local(0, 20, 0)))
				}
				// From completion, the completion's clock on the next day.
				res := complete(nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true, Due: dated(local(-2, 20, 0)), Now: local(-1, 19, 30)})
				assertDated(t, "from completion", res.Due, dated(local(0, 19, 30)))
				// A sub-daily rule is elapsed.
				due := local(-1, 20, 0)
				res = complete(nextOccurrenceInput{Rule: "FREQ=HOURLY;INTERVAL=24", Due: dated(due), Start: dated(local(-1, 8, 0)), Now: due.Add(5 * time.Minute)})
				assertDated(t, "hourly due", res.Due, dated(due.Add(24*time.Hour)))
				assertDated(t, "hourly start", res.Start, dated(local(-1, 8, 0).Add(24*time.Hour)))

				// Normalisation and the intended clock time at the affected wall clock: the due date
				// (plain and with a named time) and a companion.
				probe := tr.probe // a time of day inside the gap or the overlap, or 02:30 without DST
				h, mi := int(probe/time.Hour), int(probe%time.Hour/time.Minute)
				named := "FREQ=DAILY;BYHOUR=" + strconv.Itoa(h) + ";BYMINUTE=" + strconv.Itoa(mi)
				for _, rule := range []string{"FREQ=DAILY", named} {
					due := local(-1, h, mi)
					in := nextOccurrenceInput{Rule: rule, Due: dated(due), Now: due.Add(5 * time.Minute)}
					first := complete(in)
					assertTransitionResult(t, rule+" due", zone, tr, first.Due, wall(0, probe), true)
					second := complete(fedBack(in, first, first.Due.At.Add(5*time.Minute), zone))
					assertDated(t, rule+" due, next step", second.Due, dated(time.Date(d.Year(), d.Month(), d.Day()+1, h, mi, 0, 0, zone)))
				}
				due = local(-1, 20, 0)
				in := nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(due), Start: dated(local(-1, h, mi)), Now: due.Add(time.Hour)}
				first := complete(in)
				assertDated(t, "companion's due", first.Due, dated(local(0, 20, 0)))
				assertTransitionResult(t, "companion", zone, tr, first.Start, wall(0, probe), false)
				second := complete(fedBack(in, first, first.Due.At.Add(time.Hour), zone))
				kept := time.Duration(0) // a gap-shifted companion keeps the shifted clock (a known limit)
				if tr.change > 0 {
					kept = tr.change
				}
				assertDated(t, "companion, next step", second.Start, dated(time.Date(d.Year(), d.Month(), d.Day()+1, h, mi, 0, 0, zone).Add(kept)))
			})
		}
	}
}

// zoneTransition is a zone's clock change on one local date: its size (positive for a gap,
// negative for an overlap, zero when there is none) and a time of day inside the affected hour.
type zoneTransition struct {
	change time.Duration
	probe  time.Duration
}

// transitionOn finds the zone's clock change on the local date d by scanning its offsets.
func transitionOn(zone *time.Location, d time.Time) zoneTransition {
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, zone).Add(-6 * time.Hour)
	_, prev := start.In(zone).Zone()
	for at := start; at.Before(start.Add(36 * time.Hour)); at = at.Add(15 * time.Minute) {
		_, off := at.In(zone).Zone()
		if off == prev {
			continue
		}
		change := time.Duration(off-prev) * time.Second
		// The affected wall clock starts at the transition read in the earlier offset (a gap) or the
		// later one (an overlap); the probe is the middle of it.
		reading := at.Add(time.Duration(prev) * time.Second)
		if change < 0 {
			reading = at.Add(time.Duration(off) * time.Second)
		}
		magnitude := change
		if magnitude < 0 {
			magnitude = -magnitude
		}
		return zoneTransition{change: change, probe: timeOfDay(reading.UTC()) + magnitude/2}
	}
	return zoneTransition{probe: 2*time.Hour + 30*time.Minute}
}

// assertTransitionResult checks a date asked for at the floating wall clock asked on the change's
// date: a gap moves it forward by the gap, and a due date (remembers) keeps the asked clock time as
// its intent while a companion gets none; an overlap keeps the clock at the earlier instant;
// otherwise it is the identity. IntendedZone is always "".
func assertTransitionResult(t *testing.T, label string, zone *time.Location, tr zoneTransition, got datedValue, asked time.Time, remembers bool) {
	t.Helper()
	assert.Empty(t, got.IntendedZone, label)
	switch {
	case tr.change > 0:
		assert.True(t, float(got.At, zone).Equal(asked.Add(tr.change)), "%s: gap: wall clock %v, want %v", label, float(got.At, zone), asked.Add(tr.change))
		if !remembers {
			assert.Nil(t, got.Intended, "%s: a companion carries no intent", label)
		} else if assert.NotNil(t, got.Intended, "%s: the asked clock is remembered", label) {
			assert.Equal(t, timeOfDay(asked), *got.Intended, label)
		}
	case tr.change < 0:
		assert.True(t, float(got.At, zone).Equal(asked), "%s: overlap: wall clock %v, want %v", label, float(got.At, zone), asked)
		assert.True(t, float(got.At.Add(-tr.change), zone).Equal(asked), "%s: overlap: the earlier instant", label)
		assert.Nil(t, got.Intended, label)
	default:
		assert.True(t, float(got.At, zone).Equal(asked), "%s: identity: wall clock %v, want %v", label, float(got.At, zone), asked)
		assert.Nil(t, got.Intended, label)
	}
}

// ---------------------------------------------------------------------------------------------
// Relative reminders. Expected values are written as UTC instants with the local reading beside
// them.
// ---------------------------------------------------------------------------------------------

func utc(y int, mo time.Month, d, h, mi int) time.Time {
	return time.Date(y, mo, d, h, mi, 0, 0, time.UTC)
}

// relativeReminderCase is one row of a relativeReminderTime table: elapsed is base + period, today's
// value, asserted where it differs so a regression to plain addition fails loudly.
type relativeReminderCase struct {
	name      string
	base      time.Time
	period    int64
	want      time.Time
	differsBy bool // want differs from the elapsed base + period
}

// newYorkRelativeReminderCases is the New York table.
func newYorkRelativeReminderCases() []relativeReminderCase {
	return []relativeReminderCase{
		// due Sun 2027-03-14 20:00 EDT, 1 day before → Sat 20:00 EST (elapsed: Sat 19:00 EST)
		{"1 day before across spring forward", utc(2027, 3, 15, 0, 0), -86400, utc(2027, 3, 14, 1, 0), true},
		// due Sat 2026-11-07 09:00 EST, 1 week before → Sat 10-31 09:00 EDT (elapsed: 10:00 EDT)
		{"1 week before across fall back", utc(2026, 11, 7, 14, 0), -604800, utc(2026, 10, 31, 13, 0), true},
		// start Sat 2027-03-13 20:00 EST, 1 day after → Sun 20:00 EDT (elapsed: Sun 21:00 EDT)
		{"1 day after across spring forward", utc(2027, 3, 14, 1, 0), 86400, utc(2027, 3, 15, 0, 0), true},
		// due Sun 2027-03-14 20:00 EDT, 2 hours before → Sun 18:00 EDT (sub-day: elapsed)
		{"2 hours before is elapsed", utc(2027, 3, 15, 0, 0), -7200, utc(2027, 3, 14, 22, 0), false},
		// due Sun 2027-03-14 20:00 EDT, 1 day 2 hours before → Sat 18:00 EST (elapsed: Sat 17:00 EST)
		{"1 day 2 hours before is days first", utc(2027, 3, 15, 0, 0), -93600, utc(2027, 3, 13, 23, 0), true},
		// due Mon 2027-03-15 02:30 EDT, 1 day before → Sun 02:30 is in the gap → Sun 03:30 EDT
		// (elapsed: Sun 01:30 EST)
		{"1 day before lands in the gap", utc(2027, 3, 15, 6, 30), -86400, utc(2027, 3, 14, 7, 30), true},
		// due Mon 2026-11-02 01:30 EST, 1 day before → Sun 01:30 occurs twice → the earlier, EDT
		// (elapsed: Sun 01:30 EST)
		{"1 day before lands in the overlap", utc(2026, 11, 2, 6, 30), -86400, utc(2026, 11, 1, 5, 30), true},
	}
}

// subDayRelativeReminderCases start at the later occurrence of Sun 2026-11-01 01:30 (EST, 06:30Z)
// with sub-day periods, which never read the zone.
func subDayRelativeReminderCases() []relativeReminderCase {
	base := utc(2026, 11, 1, 6, 30)
	return []relativeReminderCase{
		{"zero period", base, 0, base, false},
		{"2 hours before", base, -7200, utc(2026, 11, 1, 4, 30), false}, // 00:30 EDT
		{"1 hour after", base, 3600, utc(2026, 11, 1, 7, 30), false},    // 02:30 EST
	}
}

func assertRelativeReminderCase(t *testing.T, c relativeReminderCase, zone *time.Location) {
	t.Helper()
	got := relativeReminderTime(c.base, c.period, zone)
	assert.True(t, got.Equal(c.want), "got %v, want %v", got.UTC(), c.want.UTC())
	elapsed := c.base.Add(time.Duration(c.period) * time.Second)
	if c.differsBy {
		assert.False(t, got.Equal(elapsed), "must not be the elapsed %v", elapsed.UTC())
	} else {
		assert.True(t, got.Equal(elapsed), "a sub-day period is elapsed")
	}
}

// TestRelativeReminderTime: whole days of a relative period count calendar days in the zone and
// keep the base's local clock time; the remainder and any sub-day period are elapsed.
func TestRelativeReminderTime(t *testing.T) {
	ny := mustZone(t, "America/New_York")
	for _, c := range newYorkRelativeReminderCases() {
		t.Run(c.name, func(t *testing.T) { assertRelativeReminderCase(t, c, ny) })
	}
}

// TestRelativeReminderTimeSubDayNeverReadsTheZone: a sub-day period is base + period with a nil
// zone too, so an overlap instant keeps its identity.
func TestRelativeReminderTimeSubDayNeverReadsTheZone(t *testing.T) {
	for _, c := range subDayRelativeReminderCases() {
		t.Run(c.name, func(t *testing.T) { assertRelativeReminderCase(t, c, nil) })
	}
}

// TestRelativeReminderTimeNilZoneWholeDays: a nil zone with a whole-day period is a caller bug (a
// zone is always resolved for one). The defensive choice is the elapsed value, today's behaviour.
func TestRelativeReminderTimeNilZoneWholeDays(t *testing.T) {
	base := utc(2027, 3, 15, 0, 0)
	got := relativeReminderTime(base, -93600, nil)
	assert.True(t, got.Equal(base.Add(-93600*time.Second)), "got %v", got.UTC())
}

// TestRelativeReminderTimeBerlinSaver: the saver's zone decides the calendar. Due Sun
// 2027-03-14 20:00 EDT is Mon 03-15 01:00 CET; one day before in Berlin is Sun 03-14 01:00 CET
// (00:00Z, Sat 19:00 EST), not New York's Sat 20:00 EST.
func TestRelativeReminderTimeBerlinSaver(t *testing.T) {
	got := relativeReminderTime(utc(2027, 3, 15, 0, 0), -86400, mustZone(t, "Europe/Berlin"))
	assert.True(t, got.Equal(utc(2027, 3, 14, 0, 0)), "got %v", got.UTC())
}

// TestRelativeReminderTimeZoneMatrix runs the matrix zones. UTC and Kolkata have no clock change,
// so every New York row is the elapsed value there; Berlin and Auckland at their transitions.
func TestRelativeReminderTimeZoneMatrix(t *testing.T) {
	for _, name := range []string{"UTC", "Asia/Kolkata"} {
		zone := mustZone(t, name)
		t.Run(name, func(t *testing.T) {
			for _, c := range append(newYorkRelativeReminderCases(), subDayRelativeReminderCases()...) {
				got := relativeReminderTime(c.base, c.period, zone)
				assert.True(t, got.Equal(c.base.Add(time.Duration(c.period)*time.Second)), "%s: got %v", c.name, got.UTC())
			}
		})
	}

	transitions := map[string][]relativeReminderCase{
		"Europe/Berlin": {
			// spring 2027-03-28: Sun 20:00 CEST → Sat 03-27 20:00 CET
			{"spring forward", utc(2027, 3, 28, 18, 0), -86400, utc(2027, 3, 27, 19, 0), true},
			// Mon 03-29 02:30 CEST → Sun 02:30 is in the gap → 03:30 CEST
			{"spring gap", utc(2027, 3, 29, 0, 30), -86400, utc(2027, 3, 28, 1, 30), true},
			// fall 2026-10-25: Sun 20:00 CET → Sat 10-24 20:00 CEST
			{"fall back", utc(2026, 10, 25, 19, 0), -86400, utc(2026, 10, 24, 18, 0), true},
			// Mon 10-26 02:30 CET → Sun 02:30 occurs twice → the earlier, CEST
			{"fall overlap", utc(2026, 10, 26, 1, 30), -86400, utc(2026, 10, 25, 0, 30), true},
		},
		"Pacific/Auckland": {
			// spring 2026-09-27: Sun 20:00 NZDT → Sat 09-26 20:00 NZST
			{"spring forward", utc(2026, 9, 27, 7, 0), -86400, utc(2026, 9, 26, 8, 0), true},
			// Mon 09-28 02:30 NZDT → Sun 02:30 is in the gap → 03:30 NZDT
			{"spring gap", utc(2026, 9, 27, 13, 30), -86400, utc(2026, 9, 26, 14, 30), true},
			// fall 2027-04-04: Sun 20:00 NZST → Sat 04-03 20:00 NZDT
			{"fall back", utc(2027, 4, 4, 8, 0), -86400, utc(2027, 4, 3, 7, 0), true},
			// Mon 04-05 02:30 NZST → Sun 02:30 occurs twice → the earlier, NZDT
			{"fall overlap", utc(2027, 4, 4, 14, 30), -86400, utc(2027, 4, 3, 13, 30), true},
		},
	}
	for name, cases := range transitions {
		zone := mustZone(t, name)
		for _, c := range cases {
			t.Run(name+"/"+c.name, func(t *testing.T) { assertRelativeReminderCase(t, c, zone) })
		}
	}
}

// TestRelativeReminderTimeIgnoresTheProcessZone: the function is pure; the process zone
// does not change a result.
func TestRelativeReminderTimeIgnoresTheProcessZone(t *testing.T) {
	withProcessZone(t, mustZone(t, "Asia/Tokyo"))

	ny := mustZone(t, "America/New_York")
	for _, c := range newYorkRelativeReminderCases() {
		t.Run(c.name, func(t *testing.T) { assertRelativeReminderCase(t, c, ny) })
	}
}
