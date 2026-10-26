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
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

// zeros returns n copies of "0" joined by commas, the shape of the expansion bomb: rrule-go
// prebuilds |BYHOUR| x |BYMINUTE| x |BYSECOND| timestamps for a daily or longer rule.
func zeros(n int) string {
	return strings.TrimSuffix(strings.Repeat("0,", n), ",")
}

func zeroInts(n int) []int {
	return make([]int, n)
}

// fullList returns lo..hi joined by commas.
func fullList(lo, hi int) string {
	values := make([]string, 0, hi-lo+1)
	for v := lo; v <= hi; v++ {
		values = append(values, strconv.Itoa(v))
	}
	return strings.Join(values, ",")
}

// Dense time sets within every per-list bound. rrule-go materialises days-in-period x time set
// per period, so the yearly one allocated ~769 MB per After() and the monthly one ~102 MB.
var (
	fullTimeSet  = "BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 59) + ";BYSECOND=" + fullList(0, 59)
	denseDaily   = "FREQ=DAILY;" + fullTimeSet
	denseYearly  = "FREQ=YEARLY;BYDAY=MO,TU,WE,TH,FR,SA,SU;" + fullTimeSet
	denseMonthly = "FREQ=MONTHLY;BYMONTHDAY=" + fullList(1, 31) + ";" + fullTimeSet
	// 31 x 24 x 60 x 2 = 89,280: under the earlier 100,000 budget, it burned ~2.5 s per
	// completion on a due date from 1900.
	denseMonthlyEveryMinute = "FREQ=MONTHLY;BYMONTHDAY=" + fullList(1, 31) + ";BYHOUR=" + fullList(0, 23) +
		";BYMINUTE=" + fullList(0, 59) + ";BYSECOND=0,1"
	// 31 x 24 x 13 = 9,672: the densest monthly rule within the budget.
	worstInBudgetMonthly = "FREQ=MONTHLY;BYMONTHDAY=" + fullList(1, 31) + ";BYHOUR=" + fullList(0, 23) +
		";BYMINUTE=" + fullList(0, 12)
)

// bysetposBomb: 429 characters, one occurrence per day, every position
// impossible, so completion scanned ~2.9M days to the year 9999 (~5 s).
var bysetposBomb = "FREQ=DAILY;BYSETPOS=" + fullList(2, 130)

// allocatedBytes runs fn and returns the bytes it allocated.
func allocatedBytes(fn func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	fn()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// The expansion bomb: about 3 KB of rule text that makes rrule-go build ~125M timestamps (~3 GB).
var expansionBomb = "FREQ=DAILY;BYHOUR=" + zeros(500) + ";BYMINUTE=" + zeros(500) + ";BYSECOND=" + zeros(500)

func TestCheckRRuleBounds(t *testing.T) {
	accepted := []string{
		// INTERVAL spans up to the 100-year horizon.
		"FREQ=WEEKLY;INTERVAL=5218",
		"FREQ=SECONDLY;INTERVAL=3155760000",
		"FREQ=MONTHLY;INTERVAL=1200",
		"FREQ=YEARLY;INTERVAL=100",
		// COUNT up to 10,000.
		"FREQ=DAILY;COUNT=10000",
		// BYSETPOS within the per-period set.
		"FREQ=DAILY;BYHOUR=9,17;BYSETPOS=-1,2",
		"FREQ=WEEKLY;BYDAY=MO;BYSETPOS=7",
		// Up to 10 BYSETPOS values.
		"FREQ=YEARLY;BYSETPOS=" + fullList(1, 10),
		"FREQ=DAILY;BYHOUR=9,17;BYMINUTE=0,30",
		"FREQ=DAILY;BYHOUR=0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
		"FREQ=MONTHLY;BYMONTHDAY=28,29,30,31;BYSETPOS=-1",
		"FREQ=MONTHLY;BYMONTHDAY=-1",
		"FREQ=YEARLY;BYYEARDAY=1,-1,366,-366",
		"FREQ=YEARLY;BYWEEKNO=1,-1,53,-53",
		"FREQ=YEARLY;BYMONTH=1,2,3,4,5,6,7,8,9,10,11,12",
		"FREQ=MONTHLY;BYDAY=1MO,MO,-1MO,2TU",
		"FREQ=YEARLY;BYEASTER=0",
		"FREQ=YEARLY;BYEASTER=-80,257",
		"FREQ=DAILY",
		// Ordinals rrule-go evaluates safely: within a month up to the 5th, across a year up to the 53rd.
		"FREQ=MONTHLY;BYDAY=-5MO,5MO",
		"FREQ=YEARLY;BYMONTH=2;BYDAY=-5FR,5FR",
		"FREQ=YEARLY;BYDAY=53MO,-53MO",
		// An ordinal on a weekly rule is meaningless (RFC 5545 allows ordinals only for MONTHLY and
		// YEARLY) but rrule-go ignores it, so a stored one keeps working.
		"FREQ=WEEKLY;BYDAY=2MO",
		// The occurrences-per-period budget (10,000): a yearly rule with a few times of day fits
		// (366 days x 4 times), and so does the densest monthly rule within it (31 x 312 = 9,672).
		"FREQ=YEARLY;BYHOUR=9,17;BYMINUTE=0,30",
		worstInBudgetMonthly,
		"FREQ=MONTHLY;BYMONTHDAY=1,15;BYHOUR=0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23;BYMINUTE=0,15,30,45",
	}
	for _, rule := range accepted {
		t.Run("accepts "+rule, func(t *testing.T) {
			skipUnlessFitsPlatformInt(t, rule)
			opt, err := rrule.StrToROption(rule)
			require.NoError(t, err)
			assert.NoError(t, checkRRuleBounds(opt))
		})
	}

	rejected := map[string]string{
		// INTERVAL overflow: rrule-go loops forever (WEEKLY) or panics (MONTHLY).
		"FREQ=WEEKLY;INTERVAL=2635249153387078802":   "INTERVAL",
		"FREQ=MONTHLY;INTERVAL=9223372036854775807":  "INTERVAL",
		"FREQ=MONTHLY;INTERVAL=1201":                 "INTERVAL",
		"FREQ=YEARLY;INTERVAL=101":                   "INTERVAL",
		"FREQ=WEEKLY;INTERVAL=5219":                  "INTERVAL",
		"FREQ=SECONDLY;INTERVAL=9223372036854775807": "INTERVAL",
		"FREQ=DAILY;INTERVAL=106751991167300":        "INTERVAL",
		// COUNT over 10,000.
		"FREQ=SECONDLY;COUNT=2147483647": "COUNT",
		"FREQ=DAILY;COUNT=10001":         "COUNT",
		// BYSETPOS beyond the per-period set can never select anything; rrule-go scans to 9999.
		bysetposBomb:                              "BYSETPOS",
		"FREQ=WEEKLY;BYDAY=MO;BYSETPOS=8":         "BYSETPOS",
		"FREQ=YEARLY;BYSETPOS=" + fullList(1, 11): "BYSETPOS",
		"FREQ=DAILY;BYHOUR=9,9":                   "BYHOUR",
		"FREQ=DAILY;BYMINUTE=0,30,0":              "BYMINUTE",
		"FREQ=DAILY;BYSECOND=1,1":                 "BYSECOND",
		"FREQ=MONTHLY;BYMONTHDAY=15,15":           "BYMONTHDAY",
		"FREQ=WEEKLY;BYDAY=MO,MO":                 "BYDAY",
		"FREQ=MONTHLY;BYDAY=1MO,+1MO":             "BYDAY",
		"FREQ=YEARLY;BYEASTER=0,0":                "BYEASTER",
		"FREQ=DAILY;BYHOUR=24":                    "BYHOUR",
		"FREQ=DAILY;BYMINUTE=60":                  "BYMINUTE",
		"FREQ=DAILY;BYSECOND=60":                  "BYSECOND",
		"FREQ=MONTHLY;BYDAY=-6MO":                 "BYDAY",
		"FREQ=MONTHLY;BYDAY=10MO":                 "BYDAY",
		"FREQ=MONTHLY;BYDAY=7MO":                  "BYDAY",
		"FREQ=YEARLY;BYMONTH=12;BYDAY=8MO":        "BYDAY",
		"FREQ=YEARLY;BYMONTH=1;BYDAY=-6MO":        "BYDAY",
		denseYearly:                               "occurrences",
		denseMonthly:                              "occurrences",
		denseDaily:                                "occurrences",
		denseMonthlyEveryMinute:                   "occurrences",
		"FREQ=MONTHLY;BYMONTHDAY=0":               "BYMONTHDAY",
		"FREQ=MONTHLY;BYMONTHDAY=32":              "BYMONTHDAY",
		"FREQ=MONTHLY;BYMONTHDAY=-32":             "BYMONTHDAY",
		"FREQ=YEARLY;BYYEARDAY=367":               "BYYEARDAY",
		"FREQ=YEARLY;BYWEEKNO=54":                 "BYWEEKNO",
		"FREQ=YEARLY;BYMONTH=13":                  "BYMONTH",
		"FREQ=MONTHLY;BYMONTHDAY=1;BYSETPOS=0":    "BYSETPOS",
		"FREQ=MONTHLY;BYMONTHDAY=1;BYSETPOS=367":  "BYSETPOS",
		"FREQ=YEARLY;BYDAY=54MO":                  "BYDAY",
		"FREQ=YEARLY;BYEASTER=258":                "BYEASTER",
		"FREQ=YEARLY;BYEASTER=-81":                "BYEASTER",
		expansionBomb:                             "more than once",
	}
	for rule, part := range rejected {
		t.Run("rejects "+rule[:min(len(rule), 60)], func(t *testing.T) {
			skipUnlessFitsPlatformInt(t, rule)
			opt, err := rrule.StrToROption(rule)
			require.NoError(t, err)
			err = checkRRuleBounds(opt)
			require.Error(t, err)
			require.True(t, IsErrInvalidData(err), "want ErrInvalidData, got %T", err)
			assert.Contains(t, err.(ErrInvalidData).Message, part)
		})
	}
}

// TestRRuleExpansionBombIsRejectedBeforeAllocation: every write path rejects the bomb before
// rrule-go expands it, and fast.
func TestRRuleExpansionBombIsRejectedBeforeAllocation(t *testing.T) {
	fast := func(t *testing.T, fn func()) {
		t.Helper()
		start := time.Now()
		fn()
		assert.Less(t, time.Since(start), time.Second, "the bound must reject before expanding the rule")
	}

	t.Run("raw rule text", func(t *testing.T) {
		fast(t, func() {
			err := validateRRule(expansionBomb)
			require.Error(t, err)
			assert.True(t, IsErrInvalidData(err))
		})
	})
	t.Run("structured repeat on create and update", func(t *testing.T) {
		bomb := &TaskRepeat{Freq: "daily", Interval: 1, ByHour: zeroInts(500), ByMinute: zeroInts(500), BySecond: zeroInts(500)}
		fast(t, func() {
			_, err := resolveIncomingRepeat("", bomb, "")
			require.Error(t, err)
			assert.True(t, IsErrInvalidData(err))
			_, err = resolveIncomingRepeat("FREQ=DAILY;INTERVAL=1", bomb, "")
			require.Error(t, err)
			assert.True(t, IsErrInvalidData(err))
		})
	})
	t.Run("raw rule from a trusted caller", func(t *testing.T) {
		fast(t, func() {
			_, err := resolveIncomingRepeat("FREQ=DAILY;INTERVAL=1", nil, expansionBomb)
			require.Error(t, err)
			assert.True(t, IsErrInvalidData(err))
		})
	})
	t.Run("CalDAV and importer normalisation drops it", func(t *testing.T) {
		fast(t, func() {
			_, ok := NormalizeRRule("RRULE:" + expansionBomb)
			assert.False(t, ok)
		})
	})
	t.Run("a normal multi-value rule is still accepted", func(t *testing.T) {
		require.NoError(t, validateRRule("FREQ=DAILY;BYHOUR=9,17;BYMINUTE=0,30"))
		got, err := resolveIncomingRepeat("", &TaskRepeat{Freq: "daily", Interval: 1, ByHour: []int{9, 17}, ByMinute: []int{0, 30}}, "")
		require.NoError(t, err)
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1;BYHOUR=9,17;BYMINUTE=0,30", got.Repeats)
		normalized, ok := NormalizeRRule("RRULE:FREQ=DAILY;BYHOUR=9,17;BYMINUTE=0,30")
		assert.True(t, ok)
		assert.Equal(t, "FREQ=DAILY;BYHOUR=9,17;BYMINUTE=0,30", normalized)
	})
}

// TestEngineBoundsStoredRules: a stored rule written before the bound existed must not be
// expanded unbounded on completion. Duplicate values are dropped (semantics-preserving, so the
// task advances as before); a value outside the domain leaves the task done and its dates
// unchanged, the same as a rule that no longer parses.
func TestEngineBoundsStoredRules(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)
	due := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)

	t.Run("a stored expansion bomb advances like its deduplicated rule, fast", func(t *testing.T) {
		start := time.Now()
		n := completeRepeating(&Task{Repeats: expansionBomb, DueDate: due})
		assert.Less(t, time.Since(start), time.Second)
		assert.False(t, n.Done)
		assert.Equal(t, time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC), n.DueDate.UTC())
		assert.Equal(t, expansionBomb, n.Repeats, "completion does not rewrite the stored rule")
	})
	t.Run("stored duplicates advance as before", func(t *testing.T) {
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY;BYHOUR=9,9,17", DueDate: due})
		assert.False(t, n.Done)
		assert.Equal(t, time.Date(2026, 10, 22, 17, 0, 0, 0, time.UTC), n.DueDate.UTC())
	})
	t.Run("the densest in-budget rule at the horizon's oldest due date completes in bounded time", func(t *testing.T) {
		start := time.Now()
		n := completeRepeating(&Task{Repeats: worstInBudgetMonthly, DueDate: now.AddDate(-200, 0, 1)})
		assert.Less(t, time.Since(start), time.Second, "monthly anchors are not clamped; the budget and the horizon bound the cost (measured 373-413 ms)")
		assert.False(t, n.Done)
		assert.True(t, n.DueDate.After(now))
	})
	t.Run("a stored rule over the occurrences-per-period budget leaves the task done, fast", func(t *testing.T) {
		var n *Task
		start := time.Now()
		allocated := allocatedBytes(func() {
			n = completeRepeating(&Task{Repeats: denseYearly, DueDate: due})
		})
		assert.Less(t, time.Since(start), 100*time.Millisecond)
		assert.Less(t, allocated, uint64(10<<20), "must not materialise the period")
		assert.True(t, n.Done)
		assert.True(t, n.DueDate.IsZero(), "dates are left unchanged")
	})
	t.Run("a stored BYDAY ordinal rrule-go cannot evaluate leaves the task done, without panicking", func(t *testing.T) {
		for _, rule := range []string{"FREQ=MONTHLY;BYDAY=10MO", "FREQ=MONTHLY;BYDAY=-6MO", "FREQ=YEARLY;BYMONTH=12;BYDAY=8MO"} {
			var n *Task
			require.NotPanics(t, func() {
				n = completeRepeating(&Task{Repeats: rule, DueDate: due})
			}, rule)
			assert.True(t, n.Done, rule)
		}
	})
	// A stored rule the creation policy rejects (sub-daily with a BY* part) is not expanded.
	// Some never match and would iterate forever.
	t.Run("a stored policy-rejected rule leaves the task done, without hanging", func(t *testing.T) {
		for _, rule := range []string{
			"FREQ=HOURLY;INTERVAL=2;BYHOUR=1", // never matches: rrule-go does not return
			"FREQ=HOURLY;INTERVAL=1;BYHOUR=9", // would advance, but stops too: the documented cost
		} {
			done := make(chan *Task, 1)
			go func() { done <- completeRepeating(&Task{Repeats: rule, DueDate: due}) }()
			select {
			case n := <-done:
				assert.True(t, n.Done, rule)
				assert.True(t, n.DueDate.IsZero(), "%s: dates are left unchanged", rule)
				assert.Equal(t, rule, n.Repeats, "%s: the stored rule is kept", rule)
			case <-time.After(500 * time.Millisecond):
				t.Fatalf("%s: completion did not return within 500 ms", rule)
			}
		}
	})
	t.Run("a policy-accepted rule with a BY* part still advances", func(t *testing.T) {
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY;BYHOUR=9", DueDate: due})
		assert.False(t, n.Done)
		assert.Equal(t, time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC), n.DueDate.UTC())
	})
	// Stored rules that hang or panic in rrule-go, or take minutes, leave the task done (engine
	// refusals), within a deadline so a regression cannot wedge the run.
	t.Run("stored INTERVAL, COUNT and BYSETPOS payloads leave the task done, without hanging", func(t *testing.T) {
		old := time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC)
		for _, rule := range []string{
			"FREQ=WEEKLY;INTERVAL=2635249153387078802",  // clamp overflow: loops backwards forever
			"FREQ=MONTHLY;INTERVAL=9223372036854775807", // rrule-go: negative index panic
			"FREQ=SECONDLY;COUNT=2147483647",            // hundreds of millions of occurrences
			bysetposBomb,
		} {
			result := make(chan *Task, 1)
			panicked := make(chan any, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						panicked <- r
					}
				}()
				result <- completeRepeating(&Task{Repeats: rule, DueDate: old})
			}()
			select {
			case n := <-result:
				assert.True(t, n.Done, rule)
				assert.True(t, n.DueDate.IsZero(), "%s: dates unchanged", rule)
			case p := <-panicked:
				t.Errorf("%s: completion panicked: %v", rule, p)
			case <-time.After(500 * time.Millisecond):
				t.Errorf("%s: completion did not return within 500 ms", rule)
			}
		}
	})
	t.Run("a stored rule with more than 10 BYSETPOS values leaves the task done", func(t *testing.T) {
		// It would match (24 times a day), so staying done shows the refusal.
		rule := "FREQ=DAILY;BYHOUR=" + fullList(0, 23) + ";BYSETPOS=" + fullList(1, 11)
		result := make(chan *Task, 1)
		go func() { result <- completeRepeating(&Task{Repeats: rule, DueDate: due}) }()
		select {
		case n := <-result:
			assert.True(t, n.Done)
			assert.True(t, n.DueDate.IsZero(), "dates unchanged")
		case <-time.After(500 * time.Millisecond):
			t.Fatal("completion did not return within 500 ms")
		}
	})
	t.Run("a stored rule with 10 distinct BYSETPOS values and a duplicate is still evaluated", func(t *testing.T) {
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY;BYHOUR=" + fullList(0, 23) + ";BYSETPOS=1,1," + fullList(2, 10), DueDate: due})
		assert.False(t, n.Done)
		assert.True(t, n.DueDate.After(now))
	})
	t.Run("the worst never-matching in-bound rules complete in bounded time", func(t *testing.T) {
		// The densest time set within the per-period budget, an impossible date and the 10
		// allowed positions: every period up to the year 9999 is scanned.
		for _, rule := range []string{
			"FREQ=DAILY;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 59) + ";BYSECOND=" + fullList(0, 5) + ";BYSETPOS=" + fullList(1, 10),
			"FREQ=WEEKLY;BYDAY=MO;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 58) + ";BYSETPOS=" + fullList(1, 10),
			"FREQ=MONTHLY;BYDAY=MO;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 12) + ";BYSETPOS=" + fullList(1, 10),
			"FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=0;BYSETPOS=" + fullList(1, 10),
		} {
			opt, err := rrule.StrToROption(rule)
			require.NoError(t, err)
			require.NoError(t, checkRRuleWriteLimits(opt), "the rule is in bounds")
			start := time.Now()
			n := completeRepeating(&Task{Repeats: rule, DueDate: time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC)})
			assert.Less(t, time.Since(start), 500*time.Millisecond, rule[:40]) // measured 12-170 ms
			assert.True(t, n.Done, "never matches")
		}
	})
	t.Run("the largest accepted COUNT completes in bounded time", func(t *testing.T) {
		start := time.Now()
		n := completeRepeating(&Task{Repeats: "FREQ=SECONDLY;COUNT=10000", DueDate: time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC)})
		assert.Less(t, time.Since(start), time.Second)
		assert.True(t, n.Done, "no occurrence is left after now")
	})
	t.Run("INTERVAL at the 100-year horizon advances; beyond it the task stays done", func(t *testing.T) {
		const yearSeconds = 31_557_600
		old := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
		for _, f := range []struct {
			freq  string
			unit  int64
			limit int64
		}{{"SECONDLY", 1, 3_155_760_000}, {"MINUTELY", 60, 52_596_000}, {"HOURLY", 3600, 876_600}, {"DAILY", 86400, 36_525}, {"WEEKLY", 604800, 5_218}} {
			limit := f.limit
			rule := fmt.Sprintf("FREQ=%s;INTERVAL=%d", f.freq, limit)
			if !fitsPlatformInt(rule) {
				continue // 32-bit: the SECONDLY limit does not parse into rrule-go's int
			}
			n := completeRepeating(&Task{Repeats: rule, DueDate: old})
			assert.False(t, n.Done, rule)
			// Completion returns dates in the server zone: the instant is compared, in UTC like the
			// expected value.
			assert.Equal(t, old.Add(time.Duration(limit*f.unit)*time.Second), n.DueDate.UTC(), rule)
			for _, interval := range []int64{limit + 1, 300 * yearSeconds / f.unit, 10_000 * yearSeconds / f.unit} {
				rule := fmt.Sprintf("FREQ=%s;INTERVAL=%d", f.freq, interval)
				n := completeRepeating(&Task{Repeats: rule, DueDate: old})
				assert.True(t, n.Done, "%s: beyond the horizon", rule)
				assert.True(t, n.DueDate.IsZero(), "%s: dates unchanged", rule)
			}
		}
		for _, rule := range []string{"FREQ=MONTHLY;INTERVAL=1201", "FREQ=YEARLY;INTERVAL=101"} {
			n := completeRepeating(&Task{Repeats: rule, DueDate: old})
			assert.True(t, n.Done, rule)
		}
	})
	t.Run("a due date more than 200 years ago is beyond the horizon", func(t *testing.T) {
		// The task stays done, its dates unchanged.
		beyond := now.AddDate(-200, 0, -1)
		within := now.AddDate(-200, 0, 1)
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: beyond})
		assert.True(t, n.Done)
		assert.True(t, n.DueDate.IsZero())
		n = completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: within})
		assert.False(t, n.Done)
		assert.True(t, n.DueDate.After(now))

		// The horizon covers due dates only. Without a due date, old dates take one step each (as
		// upstream's legacy repeat did), which costs nothing, so they are not refused.
		start := now.Add(-48 * time.Hour)
		reminder := time.Date(1700, 1, 1, 9, 0, 0, 0, time.UTC)
		n = completeRepeating(&Task{Repeats: "FREQ=MONTHLY", StartDate: start, Reminders: []*TaskReminder{{Reminder: reminder}}})
		assert.False(t, n.Done, "a no-due-date task with a reminder in 1700 advances")
		// Completion returns dates in the server zone: the instants are compared, in UTC like the
		// expected values.
		assert.Equal(t, start.AddDate(0, 1, 0), n.StartDate.UTC())
		assert.Equal(t, reminder.AddDate(0, 1, 0), reminderTime(t, n, 0).UTC())
		n = completeRepeating(&Task{Repeats: "FREQ=MONTHLY", StartDate: beyond})
		assert.False(t, n.Done)
		// Repeating from the current date, the anchor is now: an old start moves.
		n = completeRepeating(&Task{Repeats: "FREQ=DAILY", StartDate: beyond, RepeatsFromCurrentDate: true})
		assert.False(t, n.Done)
		assert.Equal(t, now.Add(24*time.Hour), n.StartDate.UTC()) // the instant, in UTC
	})
	t.Run("a next date past the year 9999 leaves the task done", func(t *testing.T) {
		// MySQL's DATETIME ends at 9999-12-31; nothing later is stored.
		due := time.Date(9999, 12, 1, 9, 0, 0, 0, time.UTC)
		withFixedNow(t, time.Date(9999, 12, 10, 0, 0, 0, 0, time.UTC))
		for _, rule := range []string{"FREQ=DAILY;INTERVAL=365", "FREQ=WEEKLY;INTERVAL=52"} {
			n := completeRepeating(&Task{Repeats: rule, DueDate: due})
			assert.True(t, n.Done, rule)
			assert.True(t, n.DueDate.IsZero(), "%s: dates unchanged", rule)
		}
		withFixedNow(t, now)
	})
	t.Run("a stored BYEASTER rrule-go cannot evaluate leaves the task done, without panicking", func(t *testing.T) {
		var n *Task
		require.NotPanics(t, func() {
			n = completeRepeating(&Task{Repeats: "FREQ=YEARLY;BYEASTER=300", DueDate: due})
		})
		assert.True(t, n.Done)
		assert.True(t, n.DueDate.IsZero(), "dates are left unchanged")
	})
}

// TestRRuleLengthLimit: the repeats column is varchar(500). A longer rule is refused on API
// writes and copies, and dropped by CalDAV and import normalisation, so PostgreSQL and MySQL
// never see an over-long insert.
func TestRRuleLengthLimit(t *testing.T) {
	days := make([]string, 200)
	for i := range days {
		days[i] = strconv.Itoa(i + 1)
	}
	tooLong := "FREQ=YEARLY;BYYEARDAY=" + strings.Join(days, ",")
	require.Greater(t, len(tooLong), maxRRuleLength)
	opt, err := rrule.StrToROption(tooLong)
	require.NoError(t, err)
	require.NoError(t, checkRRuleBounds(opt), "the rule is within the safety bound; only its length is wrong")

	err = validateRRule(tooLong)
	require.Error(t, err)
	assert.Contains(t, err.(ErrInvalidData).Message, "500 characters")
	_, err = resolveCopiedRepeat(tooLong)
	require.Error(t, err)
	_, ok := NormalizeRRule("RRULE:" + tooLong)
	assert.False(t, ok)

	// The longest such rule that still fits is accepted.
	atLimit := "FREQ=YEARLY;BYYEARDAY=1"
	for i := 2; len(atLimit)+len(","+strconv.Itoa(i)) <= maxRRuleLength; i++ {
		atLimit += "," + strconv.Itoa(i)
	}
	require.Greater(t, len(atLimit), maxRRuleLength-4)
	assert.NoError(t, validateRRule(atLimit))
}

// TestRRuleLengthLimitAfterNormalisation: rrule-go's serialisation can lengthen a rule, e.g. a
// date-only UNTIL gains a time (+8 characters), and completing a COUNT rule writes it back
// serialised. A rule that fits only before serialisation is refused, or completions would fail
// on PostgreSQL and MySQL.
func TestRRuleLengthLimitAfterNormalisation(t *testing.T) {
	raw := "FREQ=YEARLY;UNTIL=20300101;BYYEARDAY=1"
	for i := 2; len(raw)+len(","+strconv.Itoa(i)) <= 499; i++ {
		raw += "," + strconv.Itoa(i)
	}
	require.LessOrEqual(t, len(raw), maxRRuleLength)
	require.GreaterOrEqual(t, len(raw), 493)
	opt, err := rrule.StrToROption(raw)
	require.NoError(t, err)
	require.Greater(t, len(opt.RRuleString()), maxRRuleLength, "the serialised form is longer")

	err = validateRRule(raw)
	require.Error(t, err)
	assert.Contains(t, err.(ErrInvalidData).Message, "500 characters")
	_, err = resolveCopiedRepeat(raw)
	require.Error(t, err)
}

// TestRRuleIntervalHorizon: writes refuse an INTERVAL spanning more than 100 years, on every
// frequency; CalDAV and import normalisation drop it.
func TestRRuleIntervalHorizon(t *testing.T) {
	for _, rule := range []string{
		"FREQ=YEARLY;INTERVAL=100", "FREQ=MONTHLY;INTERVAL=1200", "FREQ=WEEKLY;INTERVAL=5218",
		"FREQ=DAILY;INTERVAL=36525", "FREQ=HOURLY;INTERVAL=876600", "FREQ=MINUTELY;INTERVAL=52596000",
		"FREQ=SECONDLY;INTERVAL=3155760000",
	} {
		if !fitsPlatformInt(rule) {
			continue // 32-bit: does not parse into rrule-go's int
		}
		require.NoError(t, validateRRule(rule), rule)
	}
	for _, rule := range []string{
		"FREQ=YEARLY;INTERVAL=101", "FREQ=MONTHLY;INTERVAL=1201", "FREQ=WEEKLY;INTERVAL=5219",
		"FREQ=DAILY;INTERVAL=36526", "FREQ=HOURLY;INTERVAL=876601", "FREQ=MINUTELY;INTERVAL=52596001",
		"FREQ=SECONDLY;INTERVAL=3155760001", "FREQ=WEEKLY;INTERVAL=2635249153387078802",
	} {
		if !fitsPlatformInt(rule) {
			continue // 32-bit: refused as unparseable rather than by the horizon
		}
		err := validateRRule(rule)
		require.Error(t, err, rule)
		assert.Contains(t, err.(ErrInvalidData).Message, "100 years", rule)
		_, ok := NormalizeRRule(rule)
		assert.False(t, ok, rule)
	}
}

// TestAdvanceTaskDatesRecoversFromAPanic: the bounds keep every known panic out of rrule-go, but
// an evaluation that panics anyway (here a BYDAY ordinal the bounds refuse, passed past them) must
// not take the request down: the task stays done with its dates, reminders and rule unchanged.
func TestAdvanceTaskDatesRecoversFromAPanic(t *testing.T) {
	// A nil option panics inside the evaluation stage whatever rrule-go version is in use, standing in
	// for any panic there; the bounds already keep the library's known panics out of reach.
	const rule = "FREQ=MONTHLY;BYDAY=-1MO"
	var opt *rrule.ROption

	due := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	reminder := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	start := time.Date(2026, 9, 29, 9, 0, 0, 0, time.UTC)
	oldTask := &Task{ID: 99, Repeats: rule, DueDate: due, StartDate: start, Reminders: []*TaskReminder{{Reminder: reminder}}}
	newTask := &Task{ID: 99, Repeats: rule, Done: true, DueDate: due, StartDate: start, Reminders: oldTask.Reminders}

	var moved bool
	require.NotPanics(t, func() {
		moved = advanceTaskDates(oldTask, newTask, opt, rrule.ROption{}, time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC))
	})
	assert.False(t, moved)
	assert.True(t, newTask.Done)
	assert.Equal(t, rule, newTask.Repeats)
	assert.Equal(t, due, newTask.DueDate)
	assert.Equal(t, start, newTask.StartDate)
	require.Len(t, newTask.Reminders, 1)
	assert.Equal(t, reminder, newTask.Reminders[0].Reminder)
}

// TestRRuleBydayOrdinalBelowMonthly: RFC 5545 allows a BYDAY position (2MO) only for MONTHLY and
// YEARLY rules. Writes refuse it on other frequencies; a stored one keeps working, because
// rrule-go ignores the position there.
func TestRRuleBydayOrdinalBelowMonthly(t *testing.T) {
	for _, rule := range []string{"FREQ=WEEKLY;BYDAY=2MO", "FREQ=DAILY;BYDAY=1FR", "FREQ=WEEKLY;BYDAY=MO,-1FR"} {
		err := validateRRule(rule)
		require.Error(t, err, rule)
		assert.Contains(t, err.(ErrInvalidData).Message, "BYDAY", rule)
		_, ok := NormalizeRRule(rule)
		assert.False(t, ok, rule)
	}
	// Sub-daily rules with BYDAY are refused already, by the creation policy.
	require.Error(t, validateRRule("FREQ=HOURLY;BYDAY=1MO"))
	for _, rule := range []string{"FREQ=WEEKLY;BYDAY=MO", "FREQ=DAILY;BYDAY=MO,FR", "FREQ=MONTHLY;BYDAY=2MO", "FREQ=YEARLY;BYDAY=20MO"} {
		require.NoError(t, validateRRule(rule), rule)
	}

	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) // a Thursday
	withFixedNow(t, now)
	n := completeRepeating(&Task{Repeats: "FREQ=WEEKLY;BYDAY=2MO", DueDate: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)})
	assert.False(t, n.Done, "a stored one still advances")
	assert.Equal(t, time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC), n.DueDate.UTC(), "the next Monday")
}

// TestPlainUntilRulesAdvanceArithmetically: a sub-daily rule with UNTIL but no BY* part and no
// COUNT used to be iterated occurrence by occurrence from its due date, so
// FREQ=SECONDLY;UNTIL=20991231T235959Z with a 2020 due date generated 200M+ occurrences per
// completion. It now advances by arithmetic, as without UNTIL, and UNTIL is applied to the result.
// The result matches the UNTIL-less arithmetic path, and equals rrule-go's away from DST
// transitions (the cases below use UTC); across one it steps by elapsed time.
func TestPlainUntilRulesAdvanceArithmetically(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)

	t.Run("an old due date completes in bounded time", func(t *testing.T) {
		for _, rule := range []string{"FREQ=SECONDLY;UNTIL=20991231T235959Z", "FREQ=MINUTELY;UNTIL=20991231T235959Z", "FREQ=HOURLY;UNTIL=20991231T235959Z"} {
			result := make(chan *Task, 1)
			go func() {
				// The oldest due date the horizon still evaluates (200 years back).
				result <- completeRepeating(&Task{Repeats: rule, DueDate: time.Date(1830, 1, 1, 9, 0, 0, 0, time.UTC)})
			}()
			select {
			case n := <-result:
				assert.False(t, n.Done, rule)
				assert.True(t, n.DueDate.After(now), rule)
			case <-time.After(500 * time.Millisecond):
				t.Errorf("%s: completion did not return within 500 ms", rule)
			}
		}
	})

	t.Run("away from DST transitions the next occurrence is what rrule-go gives", func(t *testing.T) {
		dues := []time.Time{
			now.Add(-3 * 24 * time.Hour),
			now.Add(-90 * time.Minute),
			now.Add(-7 * time.Second),
			now.Add(2 * time.Hour), // a future due date advances from itself
		}
		rules := []string{
			"FREQ=SECONDLY;INTERVAL=7;UNTIL=20991231T235959Z",
			"FREQ=MINUTELY;INTERVAL=13;UNTIL=20991231T235959Z",
			"FREQ=HOURLY;INTERVAL=5;UNTIL=20991231T235959Z",
			"FREQ=HOURLY;UNTIL=20261022T150000Z",   // UNTIL soon after now
			"FREQ=MINUTELY;UNTIL=20261022T120000Z", // UNTIL exactly now: nothing after now
			"FREQ=HOURLY;INTERVAL=2;UNTIL=20261022T130000Z",
		}
		for _, rule := range rules {
			for _, due := range dues {
				for _, fromCurrentDate := range []bool{false, true} {
					want := expectedNextByRRuleGo(t, rule, due, fromCurrentDate, now)
					n := completeRepeating(&Task{Repeats: rule, DueDate: due, RepeatsFromCurrentDate: fromCurrentDate})
					if want.IsZero() {
						assert.True(t, n.Done, "%s due %s fromCurrent=%v: no occurrence left", rule, due, fromCurrentDate)
						continue
					}
					assert.False(t, n.Done, "%s due %s fromCurrent=%v", rule, due, fromCurrentDate)
					assert.Equal(t, want, n.DueDate.UTC(), "%s due %s fromCurrent=%v", rule, due, fromCurrentDate)
				}
			}
		}
	})
}

// expectedNextByRRuleGo is the next due date as the iterating path computed it: the rule anchored at
// the due date (or at now when repeating from the current date), first occurrence after now (or
// after a future due date).
func expectedNextByRRuleGo(t *testing.T, rule string, due time.Time, fromCurrentDate bool, now time.Time) time.Time {
	t.Helper()
	r, err := rrule.StrToRRule(rule)
	require.NoError(t, err)
	anchor, from := due, now
	switch {
	case fromCurrentDate:
		anchor = now
	case due.After(now):
		from = due
	}
	r.DTStart(anchor)
	return r.After(from, false).UTC()
}

// TestCompletionKeepsDatesStorable: completion never produces a date outside
// [1000-01-01, 9999-12-31], counting relative reminders as they will be recomputed.
func TestCompletionKeepsDatesStorable(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)

	t.Run("a start date moved before the year 1000", func(t *testing.T) {
		due := time.Date(2030, 10, 23, 12, 0, 0, 0, time.UTC)
		start := time.Date(1000, 10, 23, 12, 0, 0, 0, time.UTC)
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: due, StartDate: start, RepeatsFromCurrentDate: true})
		assert.True(t, n.Done)
		assert.True(t, n.DueDate.IsZero() && n.StartDate.IsZero(), "dates unchanged")
	})
	t.Run("a relative reminder recomputed past 9999", func(t *testing.T) {
		due := time.Date(9999, 12, 29, 12, 0, 0, 0, time.UTC)
		rem := &TaskReminder{Reminder: due.Add(48 * time.Hour), RelativeTo: ReminderRelationDueDate, RelativePeriod: 172800}
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: due, Reminders: []*TaskReminder{rem}})
		assert.True(t, n.Done)
		assert.True(t, n.DueDate.IsZero(), "dates unchanged")
		assert.Equal(t, due.Add(48*time.Hour), rem.Reminder)
	})
	t.Run("a relative reminder within range still advances", func(t *testing.T) {
		due := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
		rem := &TaskReminder{Reminder: due.Add(-time.Hour), RelativeTo: ReminderRelationDueDate, RelativePeriod: -3600}
		n := completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: due, Reminders: []*TaskReminder{rem}})
		assert.False(t, n.Done)
		assert.Equal(t, time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC), n.DueDate.UTC())
	})
}

// TestStorableRangeIsCheckedInUTC: the database stores GMT (pkg/db), so the year range is checked on
// the UTC instant, whatever the service time zone.
func TestStorableRangeIsCheckedInUTC(t *testing.T) {
	west := time.FixedZone("UTC-5", -5*3600)
	east := time.FixedZone("UTC+5", 5*3600)
	withFixedNow(t, time.Date(9999, 12, 30, 12, 0, 0, 0, time.UTC))

	// 9999-12-31T22:00-05:00 is 10000-01-01T03:00Z.
	due := time.Date(9999, 12, 30, 22, 0, 0, 0, west)
	n := completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: due})
	assert.True(t, n.Done, "the next due date is in the year 10000 in UTC")
	assert.True(t, n.DueDate.IsZero(), "dates unchanged")

	// Repeating from the current date, the due date is rebased to now + 1 day and the start keeps
	// its offset to it, landing on 1000-01-01T02:00+05:00, which is 0999-12-31T21:00Z.
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)
	due = time.Date(2030, 10, 23, 12, 0, 0, 0, east)
	next := now.Add(24 * time.Hour)
	start := time.Date(1000, 1, 1, 2, 0, 0, 0, east).Add(due.Sub(next))
	n = completeRepeating(&Task{Repeats: "FREQ=DAILY", DueDate: due, StartDate: start, RepeatsFromCurrentDate: true})
	assert.True(t, n.Done, "the moved start is in the year 999 in UTC")
}

// TestNegativeIntervalAndCount: writes refuse a negative INTERVAL or COUNT, and CalDAV and import
// normalisation drop them. rrule-go reads a negative COUNT as unlimited, so a
// stored one keeps repeating at completion, unchanged.
func TestNegativeIntervalAndCount(t *testing.T) {
	for _, rule := range []string{"FREQ=DAILY;INTERVAL=-5", "FREQ=DAILY;COUNT=-5"} {
		err := validateRRule(rule)
		require.Error(t, err, rule)
		assert.True(t, IsErrInvalidData(err), rule)
		_, ok := NormalizeRRule("RRULE:" + rule)
		assert.False(t, ok, rule)
	}
	_, err := resolveIncomingRepeat("", &TaskRepeat{Freq: "daily", Interval: 1, Count: -5}, "")
	require.Error(t, err)

	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)
	n := completeRepeating(&Task{Repeats: "FREQ=DAILY;COUNT=-5", DueDate: time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)})
	assert.False(t, n.Done, "a stored negative COUNT repeats without limit, as rrule-go reads it")
	assert.Equal(t, time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC), n.DueDate.UTC())
	assert.Equal(t, "FREQ=DAILY;COUNT=-5", n.Repeats, "the stored rule is not rewritten")
}

// fitsPlatformInt reports whether every number in rule fits this platform's int. On 32-bit targets
// (upstream ships arm and 386 builds) rrule-go's ROption fields are 32-bit, so a larger value does
// not parse at all; the cases using one are about 64-bit overflow.
func fitsPlatformInt(rule string) bool {
	for _, n := range regexp.MustCompile(`[0-9]+`).FindAllString(rule, -1) {
		if _, err := strconv.ParseInt(n, 10, strconv.IntSize); err != nil {
			return false
		}
	}
	return true
}

func skipUnlessFitsPlatformInt(t *testing.T, rule string) {
	t.Helper()
	if !fitsPlatformInt(rule) {
		t.Skip("32-bit: a number in this rule does not fit rrule-go's int")
	}
}

// TestPlainSubDailyArithmeticEdges: the arithmetic path for plain sub-daily rules (with or without
// UNTIL) steps by elapsed time, matching the UNTIL-less path and upstream's legacy repeat, and
// equals rrule-go away from DST transitions.
func TestPlainSubDailyArithmeticEdges(t *testing.T) {
	t.Run("UNTIL is compared at second precision", func(t *testing.T) {
		// rrule-go truncates DTSTART to the second, so its last occurrence is 12:01:00; the
		// arithmetic result 12:01:00.5 must not exhaust the rule.
		withFixedNow(t, time.Date(2026, 10, 22, 12, 0, 0, 500_000_000, time.UTC))
		n := completeRepeating(&Task{Repeats: "FREQ=MINUTELY;UNTIL=20261022T120100Z", DueDate: time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC), RepeatsFromCurrentDate: true})
		assert.False(t, n.Done)
		assert.Equal(t, time.Date(2026, 10, 22, 12, 1, 0, 0, time.UTC), n.DueDate.Truncate(time.Second).UTC())
	})

	t.Run("a step count past time.Duration's range stays aligned to the due date", func(t *testing.T) {
		// Three ~99-year steps from 1827 exceed time.Duration, and the old
		// fallback rebased to now + interval (2125-10-23T06:00Z), past UNTIL.
		withFixedNow(t, time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC))
		due := time.Date(1827, 1, 1, 9, 0, 0, 0, time.UTC)
		const step = 3_124_202_400
		want := time.Unix(due.Unix()+3*step, 0).UTC()
		require.Equal(t, time.Date(2124, 1, 3, 15, 0, 0, 0, time.UTC), want)
		for _, rule := range []string{
			"FREQ=SECONDLY;INTERVAL=3124202400;UNTIL=21241231T235959Z",
			"FREQ=SECONDLY;INTERVAL=3124202400",
		} {
			if !fitsPlatformInt(rule) {
				continue
			}
			n := completeRepeating(&Task{Repeats: rule, DueDate: due})
			assert.False(t, n.Done, rule)
			assert.Equal(t, want, n.DueDate.UTC(), rule)
		}
	})

	t.Run("across DST transitions it steps by elapsed time", func(t *testing.T) {
		ny, err := time.LoadLocation("America/New_York")
		require.NoError(t, err)
		// Fall back (2026-11-01 02:00 EDT -> 01:00 EST): three elapsed hours from 00:30 EDT.
		withFixedNow(t, time.Date(2026, 11, 1, 6, 0, 0, 0, time.UTC))
		due := time.Date(2026, 11, 1, 0, 30, 0, 0, ny)
		n := completeRepeating(&Task{Repeats: "FREQ=HOURLY;INTERVAL=3;UNTIL=20271231T000000Z", DueDate: due})
		assert.Equal(t, due.Add(3*time.Hour).UTC(), n.DueDate.UTC())
		// Spring forward (2027-03-14 02:00 EST -> 03:00 EDT): two elapsed 90-minute steps from 00:30 EST.
		withFixedNow(t, time.Date(2027, 3, 14, 7, 0, 0, 0, time.UTC))
		due = time.Date(2027, 3, 14, 0, 30, 0, 0, ny)
		n = completeRepeating(&Task{Repeats: "FREQ=MINUTELY;INTERVAL=90;UNTIL=20271231T000000Z", DueDate: due})
		assert.Equal(t, due.Add(180*time.Minute).UTC(), n.DueDate.UTC())
	})
}

// TestAddRepeatIntervalToTimeStaysAligned: the first step past now is reached from t by whole
// steps even when the step count times the step, or the gap itself, exceeds time.Duration.
func TestAddRepeatIntervalToTimeStaysAligned(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	t.Run("product past time.Duration's range", func(t *testing.T) {
		from := time.Date(1827, 1, 1, 9, 0, 0, 0, time.UTC)
		step := 3_124_202_400 * time.Second
		assert.Equal(t, time.Date(2124, 1, 3, 15, 0, 0, 0, time.UTC), addRepeatIntervalToTime(now, from, step).UTC())
	})
	t.Run("gap past time.Duration's range", func(t *testing.T) {
		from := time.Date(1700, 1, 1, 9, 0, 0, 0, time.UTC) // a Friday
		got := addRepeatIntervalToTime(now, from, 7*24*time.Hour)
		assert.True(t, got.After(now) && !got.Add(-7*24*time.Hour).After(now), "got %v", got)
		assert.Equal(t, time.Friday, got.Weekday())
		assert.Equal(t, 9, got.Hour())
	})
}

// TestNoOccurrenceWithinTheSearchWindowIsLogged: rrule-go only searches dtstart + ~292 years, so
// an infinite calendar rule anchored near the 200-year horizon can have its next occurrence beyond
// that window. The task stays done, and a distinct line says why.
func TestNoOccurrenceWithinTheSearchWindowIsLogged(t *testing.T) {
	withFixedNow(t, time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC))
	var logged []int64
	orig := logNoOccurrenceInWindow
	logNoOccurrenceInWindow = func(taskID int64) { logged = append(logged, taskID) }
	t.Cleanup(func() { logNoOccurrenceInWindow = orig })

	n := completeRepeating(&Task{ID: 7, Repeats: "FREQ=YEARLY;INTERVAL=100;BYMONTH=10;BYMONTHDAY=22", DueDate: time.Date(1826, 10, 23, 12, 0, 0, 0, time.UTC)})
	assert.True(t, n.Done, "behaviour unchanged: the task stays done")
	assert.Equal(t, []int64{7}, logged)

	// A stored negative COUNT is unlimited to rrule-go, so it is an infinite rule too.
	logged = nil
	n = completeRepeating(&Task{ID: 10, Repeats: "FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=30;COUNT=-1", DueDate: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)})
	assert.True(t, n.Done)
	assert.Equal(t, []int64{10}, logged)

	// COUNT or UNTIL running out is exhaustion, not this.
	logged = nil
	n = completeRepeating(&Task{ID: 8, Repeats: "FREQ=DAILY;UNTIL=20261001T000000Z", DueDate: time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)})
	assert.True(t, n.Done)
	n = completeRepeating(&Task{ID: 9, Repeats: "FREQ=MONTHLY;BYMONTHDAY=1;COUNT=2", DueDate: time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC)})
	assert.True(t, n.Done)
	assert.Empty(t, logged)
}

// TestNextOccurrenceTiming measures completion's worst cases outside UTC: the earlier
// fixed-duration engine's worst cases re-run in America/New_York and Asia/Kolkata, and the cases the wall-clock engine adds (one
// look-back pass, chronological COUNT, no clamp under COUNT), each at the 200-year horizon. The
// worst case measured about 0.4 s on a developer machine and 0.66 s on a GitHub-hosted runner; the
// 2 s ceilings leave room for slower CI hardware while still catching a regression to the
// multi-second costs these bounds removed. The measured durations are logged.
func TestNextOccurrenceTiming(t *testing.T) {
	const ceiling = 2 * time.Second
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	withFixedNow(t, now)
	horizon := now.AddDate(-200, 0, 1)
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	kolkata, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	timed := func(t *testing.T, zone *time.Location, old *Task) (*Task, time.Duration) {
		t.Helper()
		start := time.Now()
		n := completeRepeatingIn(old, zone)
		elapsed := time.Since(start)
		t.Logf("timing %s %s: %v", zone, t.Name(), elapsed)
		return n, elapsed
	}

	for _, zone := range []*time.Location{ny, kolkata} {
		t.Run("dense monthly in "+zone.String(), func(t *testing.T) {
			n, elapsed := timed(t, zone, &Task{Repeats: worstInBudgetMonthly, DueDate: horizon})
			assert.Less(t, elapsed, ceiling)
			assert.False(t, n.Done)
			assert.True(t, n.DueDate.After(now))
		})
		t.Run("never-matching in "+zone.String(), func(t *testing.T) {
			var worst time.Duration
			for _, rule := range neverMatchingInBoundRules() {
				n, elapsed := timed(t, zone, &Task{Repeats: rule, DueDate: time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC)})
				assert.True(t, n.Done, "never matches")
				worst = max(worst, elapsed)
			}
			t.Logf("timing %s %s worst: %v", zone, t.Name(), worst)
			assert.Less(t, worst, ceiling)
		})
	}

	for _, tc := range []struct {
		name string
		rule string
	}{
		{"sparse BYSETPOS monthly", "FREQ=MONTHLY;BYMONTHDAY=14;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3"},
		{"sparse BYSETPOS weekly", "FREQ=WEEKLY;BYDAY=SU;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3"},
		{"COUNT collision", "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=30;COUNT=10000"},
		{"plain DAILY COUNT", "FREQ=DAILY;COUNT=10000"},
		{"never-matching DAILY BYSETPOS with COUNT", neverMatchingInBoundRules()[0] + ";COUNT=10000"},
		{"chronological COUNT", "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3;COUNT=10000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, elapsed := timed(t, ny, &Task{Repeats: tc.rule, DueDate: horizon})
			assert.Less(t, elapsed, ceiling)
		})
	}
}

// neverMatchingInBoundRules are the worst never-matching in-bound rules of TestEngineBoundsStoredRules:
// the densest time set within the per-period budget, an impossible date and the 10 allowed
// positions, so every period up to the year 9999 is scanned.
func neverMatchingInBoundRules() []string {
	return []string{
		"FREQ=DAILY;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 59) + ";BYSECOND=" + fullList(0, 5) + ";BYSETPOS=" + fullList(1, 10),
		"FREQ=WEEKLY;BYDAY=MO;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 58) + ";BYSETPOS=" + fullList(1, 10),
		"FREQ=MONTHLY;BYDAY=MO;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=" + fullList(0, 12) + ";BYSETPOS=" + fullList(1, 10),
		"FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=30;BYHOUR=" + fullList(0, 23) + ";BYMINUTE=0;BYSETPOS=" + fullList(1, 10),
	}
}

// forEachZone runs fn in the server zone (GMT in tests), America/New_York and Asia/Kolkata.
func forEachZone(t *testing.T, fn func(t *testing.T, zone *time.Location)) {
	t.Helper()
	for _, zone := range []*time.Location{config.GetTimeZone(), mustZone(t, "America/New_York"), mustZone(t, "Asia/Kolkata")} {
		t.Run(zone.String(), func(t *testing.T) { fn(t, zone) })
	}
}

// withinDeadline runs fn and fails the test when it does not return within d, so a regression to an
// unbounded iteration cannot wedge the run.
func withinDeadline(t *testing.T, d time.Duration, label string, fn func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		fn()
	}()
	select {
	case p := <-done:
		if p != nil {
			t.Errorf("%s: panicked: %v", label, p)
		}
	case <-time.After(d):
		t.Fatalf("%s: did not return within %v", label, d)
	}
}

// TestNextOccurrenceLimitsInEveryZone: the recurrence limits hold in every zone. The horizon,
// bounds, policy and storable-range fixtures of TestEngineBoundsStoredRules,
// TestRRuleIntervalHorizon, TestCompletionKeepsDatesStorable and TestStorableRangeIsCheckedInUTC
// give the same refusal, with the same reason, in the server zone, New York and Kolkata, and
// completion leaves the dates unchanged; the fixtures that advance still advance.
func TestNextOccurrenceLimitsInEveryZone(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	due := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	old := time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC)
	intervalOld := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	refusals := []struct {
		name   string
		in     nextOccurrenceInput
		reason noNextReason
	}{
		{"over the occurrences-per-period budget", nextOccurrenceInput{Rule: denseYearly, Due: dated(due), Now: now}, reasonOutsideBounds},
		{"BYDAY ordinal 10MO", nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYDAY=10MO", Due: dated(due), Now: now}, ""},
		{"BYDAY ordinal -6MO", nextOccurrenceInput{Rule: "FREQ=MONTHLY;BYDAY=-6MO", Due: dated(due), Now: now}, ""},
		{"BYDAY ordinal 8MO", nextOccurrenceInput{Rule: "FREQ=YEARLY;BYMONTH=12;BYDAY=8MO", Due: dated(due), Now: now}, ""},
		{"policy-rejected, never matching", nextOccurrenceInput{Rule: "FREQ=HOURLY;INTERVAL=2;BYHOUR=1", Due: dated(due), Now: now}, reasonPolicyRejected},
		{"policy-rejected", nextOccurrenceInput{Rule: "FREQ=HOURLY;INTERVAL=1;BYHOUR=9", Due: dated(due), Now: now}, reasonPolicyRejected},
		{"INTERVAL clamp overflow", nextOccurrenceInput{Rule: "FREQ=WEEKLY;INTERVAL=2635249153387078802", Due: dated(old), Now: now}, ""},
		{"INTERVAL MaxInt64", nextOccurrenceInput{Rule: "FREQ=MONTHLY;INTERVAL=9223372036854775807", Due: dated(old), Now: now}, ""},
		{"COUNT payload", nextOccurrenceInput{Rule: "FREQ=SECONDLY;COUNT=2147483647", Due: dated(old), Now: now}, ""},
		{"BYSETPOS bomb", nextOccurrenceInput{Rule: bysetposBomb, Due: dated(old), Now: now}, ""},
		{"more than 10 BYSETPOS values", nextOccurrenceInput{Rule: "FREQ=DAILY;BYHOUR=" + fullList(0, 23) + ";BYSETPOS=" + fullList(1, 11), Due: dated(due), Now: now}, ""},
		{"BYEASTER", nextOccurrenceInput{Rule: "FREQ=YEARLY;BYEASTER=300", Due: dated(due), Now: now}, ""},
		{"never matching, in bounds", nextOccurrenceInput{Rule: neverMatchingInBoundRules()[0], Due: dated(time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC)), Now: now}, reasonNoOccurrence},
		{"INTERVAL beyond the horizon", nextOccurrenceInput{Rule: "FREQ=DAILY;INTERVAL=36526", Due: dated(intervalOld), Now: now}, reasonIntervalBeyondHorizon},
		{"INTERVAL beyond the horizon, weekly", nextOccurrenceInput{Rule: "FREQ=WEEKLY;INTERVAL=5219", Due: dated(intervalOld), Now: now}, reasonIntervalBeyondHorizon},
		{"INTERVAL beyond the horizon, yearly", nextOccurrenceInput{Rule: "FREQ=YEARLY;INTERVAL=101", Due: dated(intervalOld), Now: now}, reasonIntervalBeyondHorizon},
		{"due date beyond the horizon", nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(now.AddDate(-200, 0, -1)), Now: now}, reasonAnchorBeyondHorizon},
		{"due date in 1700", nextOccurrenceInput{Rule: "FREQ=DAILY;INTERVAL=3;BYHOUR=9", Due: dated(time.Date(1700, 1, 1, 9, 0, 0, 0, time.UTC)), Now: now}, reasonAnchorBeyondHorizon},
		{"next date past 9999, daily", nextOccurrenceInput{Rule: "FREQ=DAILY;INTERVAL=365", Due: dated(time.Date(9999, 12, 1, 9, 0, 0, 0, time.UTC)), Now: time.Date(9999, 12, 10, 0, 0, 0, 0, time.UTC)}, reasonOutsideStorableRange},
		{"next date past 9999, weekly", nextOccurrenceInput{Rule: "FREQ=WEEKLY;INTERVAL=52", Due: dated(time.Date(9999, 12, 1, 9, 0, 0, 0, time.UTC)), Now: time.Date(9999, 12, 10, 0, 0, 0, 0, time.UTC)}, reasonOutsideStorableRange},
		{"next date past 9999 in UTC", nextOccurrenceInput{Rule: "FREQ=DAILY", Due: dated(time.Date(9999, 12, 30, 22, 0, 0, 0, time.FixedZone("UTC-5", -5*3600))), Now: time.Date(9999, 12, 30, 12, 0, 0, 0, time.UTC)}, reasonOutsideStorableRange},
		{"a start moved before the year 1000", nextOccurrenceInput{Rule: "FREQ=DAILY", FromCompletion: true, Due: dated(time.Date(2030, 10, 23, 12, 0, 0, 0, time.UTC)), Start: dated(time.Date(1000, 10, 23, 12, 0, 0, 0, time.UTC)), Now: now}, reasonOutsideStorableRange},
	}
	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			if !fitsPlatformInt(tc.in.Rule) {
				t.Skip("32-bit: a number in this rule does not fit rrule-go's int")
			}
			// The reason in the server zone is the fixture's; every other zone must give the same.
			want := tc.reason
			forEachZone(t, func(t *testing.T, zone *time.Location) {
				in := tc.in
				in.Zone = zone
				var res nextOccurrenceResult
				withinDeadline(t, 500*time.Millisecond, tc.name, func() { res = nextOccurrence(in) })
				if want == "" {
					want = res.Reason
				}
				require.NotEmpty(t, res.Reason, "refused in every zone")
				assert.Equal(t, want, res.Reason)
				assertNothingMoved(t, in, res)

				withFixedNow(t, in.Now)
				var n *Task
				withinDeadline(t, 500*time.Millisecond, tc.name, func() {
					n = completeRepeatingIn(&Task{Repeats: in.Rule, DueDate: in.Due.At, StartDate: in.Start.At, RepeatsFromCurrentDate: in.FromCompletion}, zone)
				})
				assert.True(t, n.Done, "the task stays done")
				assert.True(t, n.DueDate.IsZero() && n.StartDate.IsZero(), "dates unchanged")
			})
		})
	}

	t.Run("the fixtures that advance still advance", func(t *testing.T) {
		forEachZone(t, func(t *testing.T, zone *time.Location) {
			withFixedNow(t, now)
			for _, old := range []*Task{
				{Repeats: expansionBomb, DueDate: due},
				{Repeats: "FREQ=DAILY;BYHOUR=9,9,17", DueDate: due},
				{Repeats: "FREQ=DAILY;INTERVAL=36525", DueDate: intervalOld},
				{Repeats: "FREQ=WEEKLY;INTERVAL=5218", DueDate: intervalOld},
				{Repeats: "FREQ=DAILY", DueDate: now.AddDate(-200, 0, 1)},
				{Repeats: "FREQ=DAILY;BYHOUR=" + fullList(0, 23) + ";BYSETPOS=1,1," + fullList(2, 10), DueDate: due},
				{Repeats: "FREQ=DAILY", DueDate: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC), Reminders: []*TaskReminder{{RelativeTo: ReminderRelationDueDate, RelativePeriod: -3600}}},
			} {
				start := time.Now()
				n := completeRepeatingIn(old, zone)
				assert.Less(t, time.Since(start), time.Second, old.Repeats[:min(len(old.Repeats), 40)])
				assert.False(t, n.Done, old.Repeats[:min(len(old.Repeats), 40)])
				assert.True(t, n.DueDate.After(old.DueDate), old.Repeats[:min(len(old.Repeats), 40)])
			}
		})
	})

	t.Run("a relative reminder recomputed past 9999 leaves the task done", func(t *testing.T) {
		forEachZone(t, func(t *testing.T, zone *time.Location) {
			withFixedNow(t, now)
			due := time.Date(9999, 12, 29, 12, 0, 0, 0, time.UTC)
			rem := &TaskReminder{Reminder: due.Add(48 * time.Hour), RelativeTo: ReminderRelationDueDate, RelativePeriod: 172800}
			n := completeRepeatingIn(&Task{Repeats: "FREQ=DAILY", DueDate: due, Reminders: []*TaskReminder{rem}}, zone)
			assert.True(t, n.Done)
			assert.True(t, n.DueDate.IsZero(), "dates unchanged")
			assert.Equal(t, due.Add(48*time.Hour), rem.Reminder)
		})
	})
}
