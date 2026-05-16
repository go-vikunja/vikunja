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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskRepeatFromRRule(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		r := taskRepeatFromRRule("")
		assert.Nil(t, r)
	})

	t.Run("simple daily", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=DAILY;INTERVAL=1")
		require.NotNil(t, r)
		assert.Equal(t, "daily", r.Freq)
		assert.Equal(t, 1, r.Interval)
		assert.Empty(t, r.ByDay)
		assert.Empty(t, r.ByMonthDay)
	})

	t.Run("weekly with days", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR")
		require.NotNil(t, r)
		assert.Equal(t, "weekly", r.Freq)
		assert.Equal(t, 2, r.Interval)
		assert.Equal(t, []string{"mo", "we", "fr"}, r.ByDay)
	})

	t.Run("monthly with bymonthday", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15")
		require.NotNil(t, r)
		assert.Equal(t, "monthly", r.Freq)
		assert.Equal(t, 1, r.Interval)
		assert.Equal(t, []int{15}, r.ByMonthDay)
	})

	t.Run("yearly with bymonth and byday", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=YEARLY;INTERVAL=1;BYMONTH=3;BYDAY=-1SU")
		require.NotNil(t, r)
		assert.Equal(t, "yearly", r.Freq)
		assert.Equal(t, []int{3}, r.ByMonth)
		assert.Equal(t, []string{"-1su"}, r.ByDay)
	})

	t.Run("with count", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=DAILY;INTERVAL=1;COUNT=10")
		require.NotNil(t, r)
		assert.Equal(t, 10, r.Count)
	})

	t.Run("with until", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=DAILY;INTERVAL=1;UNTIL=20261231T235959Z")
		require.NotNil(t, r)
		require.NotNil(t, r.Until)
		assert.Equal(t, "2026-12-31T23:59:59Z", *r.Until)
	})

	t.Run("with wkst", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=WEEKLY;INTERVAL=1;WKST=SU")
		require.NotNil(t, r)
		assert.Equal(t, "su", r.Wkst)
	})

	t.Run("default interval is 1", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=DAILY")
		require.NotNil(t, r)
		assert.Equal(t, 1, r.Interval)
	})

	t.Run("with bysetpos", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=MONTHLY;BYDAY=MO,TU,WE,TH,FR;BYSETPOS=-1")
		require.NotNil(t, r)
		assert.Equal(t, []int{-1}, r.BySetPos)
	})

	t.Run("with byyearday", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=YEARLY;BYYEARDAY=1,100,200")
		require.NotNil(t, r)
		assert.Equal(t, []int{1, 100, 200}, r.ByYearDay)
	})

	t.Run("with byweekno", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=YEARLY;BYWEEKNO=1,52")
		require.NotNil(t, r)
		assert.Equal(t, []int{1, 52}, r.ByWeekNo)
	})

	t.Run("with byhour byminute bysecond", func(t *testing.T) {
		r := taskRepeatFromRRule("FREQ=DAILY;BYHOUR=9,17;BYMINUTE=0,30;BYSECOND=0")
		require.NotNil(t, r)
		assert.Equal(t, []int{9, 17}, r.ByHour)
		assert.Equal(t, []int{0, 30}, r.ByMinute)
		assert.Equal(t, []int{0}, r.BySecond)
	})

	t.Run("invalid rrule returns nil", func(t *testing.T) {
		r := taskRepeatFromRRule("INVALID")
		assert.Nil(t, r)
	})
}

func TestTaskRepeatToRRule(t *testing.T) {
	t.Run("nil returns empty", func(t *testing.T) {
		var r *TaskRepeat
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Empty(t, s)
	})

	t.Run("empty freq returns empty", func(t *testing.T) {
		r := &TaskRepeat{}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Empty(t, s)
	})

	t.Run("simple daily", func(t *testing.T) {
		r := &TaskRepeat{Freq: "daily", Interval: 1}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1", s)
	})

	t.Run("weekly with days", func(t *testing.T) {
		r := &TaskRepeat{
			Freq:     "weekly",
			Interval: 2,
			ByDay:    []string{"mo", "we", "fr"},
		}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR", s)
	})

	t.Run("monthly with bymonthday", func(t *testing.T) {
		r := &TaskRepeat{
			Freq:       "monthly",
			Interval:   1,
			ByMonthDay: []int{15},
		}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", s)
	})

	t.Run("yearly with nth weekday", func(t *testing.T) {
		r := &TaskRepeat{
			Freq:     "yearly",
			Interval: 1,
			ByMonth:  []int{3},
			ByDay:    []string{"-1su"},
		}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, "FREQ=YEARLY;INTERVAL=1;BYMONTH=3;BYDAY=-1SU", s)
	})

	t.Run("with count", func(t *testing.T) {
		r := &TaskRepeat{Freq: "daily", Interval: 1, Count: 10}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Contains(t, s, "COUNT=10")
	})

	t.Run("with until", func(t *testing.T) {
		until := "2026-12-31T23:59:59Z"
		r := &TaskRepeat{Freq: "daily", Interval: 1, Until: &until}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Contains(t, s, "UNTIL=20261231T235959Z")
	})

	t.Run("with wkst", func(t *testing.T) {
		r := &TaskRepeat{Freq: "weekly", Interval: 1, Wkst: "su"}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Contains(t, s, "WKST=SU")
	})

	t.Run("invalid freq returns error", func(t *testing.T) {
		r := &TaskRepeat{Freq: "invalid"}
		_, err := r.toRRule()
		assert.Error(t, err)
	})

	t.Run("invalid weekday returns error", func(t *testing.T) {
		r := &TaskRepeat{Freq: "weekly", Interval: 1, ByDay: []string{"xx"}}
		_, err := r.toRRule()
		assert.Error(t, err)
	})

	t.Run("case insensitive freq", func(t *testing.T) {
		r := &TaskRepeat{Freq: "DAILY", Interval: 1}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1", s)
	})

	t.Run("with all int slice fields", func(t *testing.T) {
		r := &TaskRepeat{
			Freq:       "yearly",
			Interval:   1,
			ByMonthDay: []int{1},
			ByMonth:    []int{1},
			ByYearDay:  []int{1},
			ByWeekNo:   []int{1},
			BySetPos:   []int{1},
			ByHour:     []int{9},
			ByMinute:   []int{30},
			BySecond:   []int{0},
		}
		s, err := r.toRRule()
		require.NoError(t, err)
		assert.Contains(t, s, "BYMONTHDAY=1")
		assert.Contains(t, s, "BYMONTH=1")
		assert.Contains(t, s, "BYYEARDAY=1")
		assert.Contains(t, s, "BYWEEKNO=1")
		assert.Contains(t, s, "BYSETPOS=1")
		assert.Contains(t, s, "BYHOUR=9")
		assert.Contains(t, s, "BYMINUTE=30")
		assert.Contains(t, s, "BYSECOND=0")
	})
}

func TestTaskRepeatRoundTrip(t *testing.T) {
	t.Run("daily round-trips", func(t *testing.T) {
		original := "FREQ=DAILY;INTERVAL=2"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("weekly with days round-trips", func(t *testing.T) {
		original := "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,WE,FR"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("monthly with bymonthday round-trips", func(t *testing.T) {
		original := "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("complex yearly round-trips", func(t *testing.T) {
		// Note: rrule-go normalizes "4TH" to "+4TH" (explicit positive sign)
		original := "FREQ=YEARLY;INTERVAL=1;BYMONTH=11;BYDAY=+4TH"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("with count round-trips", func(t *testing.T) {
		original := "FREQ=DAILY;INTERVAL=1;COUNT=5"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("with until round-trips", func(t *testing.T) {
		original := "FREQ=DAILY;INTERVAL=1;UNTIL=20261231T235959Z"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("with wkst round-trips", func(t *testing.T) {
		original := "FREQ=WEEKLY;INTERVAL=1;WKST=SU"
		r := taskRepeatFromRRule(original)
		require.NotNil(t, r)
		result, err := r.toRRule()
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})
}
