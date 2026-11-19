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
	"encoding/json"
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

func TestNormalizeRRule(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		want   string
		wantOK bool
	}{
		{"plain valid", "FREQ=DAILY;INTERVAL=1", "FREQ=DAILY;INTERVAL=1", true},
		{"strips RRULE prefix", "RRULE:FREQ=DAILY;INTERVAL=1", "FREQ=DAILY;INTERVAL=1", true},
		{"strips lowercase prefix", "rrule:FREQ=WEEKLY;BYDAY=MO", "FREQ=WEEKLY;BYDAY=MO", true},
		{"strips mixed-case prefix", "Rrule:FREQ=WEEKLY;BYDAY=MO", "FREQ=WEEKLY;BYDAY=MO", true},
		{"trims whitespace", "  FREQ=DAILY;INTERVAL=1  ", "FREQ=DAILY;INTERVAL=1", true},
		{"canonicalizes leading zero", "FREQ=MONTHLY;BYMONTHDAY=01", "FREQ=MONTHLY;BYMONTHDAY=1", true},
		{"rejects garbage", "not a rule", "", false},
		{"rejects empty", "", "", false},
		{"rejects policy: sub-daily with BY*", "FREQ=HOURLY;BYHOUR=9", "", false},
		{"allows plain sub-daily", "FREQ=HOURLY;INTERVAL=2", "FREQ=HOURLY;INTERVAL=2", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := NormalizeRRule(c.raw)
			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.want, got)
		})
	}
}

// TestTaskRepeatReadOnly pins when the structured projection is flagged read-only: exactly when
// serialising the projection loses part of the stored rule.
func TestTaskRepeatReadOnly(t *testing.T) {
	cases := []struct {
		name     string
		stored   string
		readOnly bool
	}{
		// INTERVAL parses as 0 while the projection writes INTERVAL=1.
		{"bare daily", "FREQ=DAILY", false},
		{"daily with explicit interval", "FREQ=DAILY;INTERVAL=1", false},
		{"explicit zero interval", "FREQ=DAILY;INTERVAL=0", false},
		{"byday out of canonical order", "FREQ=WEEKLY;BYDAY=FR,MO", false},
		{"nth weekday", "FREQ=MONTHLY;BYDAY=-1FR", false},
		{"month-end migrated rule", "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1", false},
		{"leading zero bymonthday", "FREQ=MONTHLY;BYMONTHDAY=01", false},
		{"until", "FREQ=DAILY;UNTIL=20261231T000000Z", false},
		{"non-default wkst", "FREQ=WEEKLY;WKST=SU;BYDAY=MO", false},
		{"explicit default wkst", "FREQ=WEEKLY;WKST=MO;BYDAY=MO", false},
		{"count", "FREQ=DAILY;COUNT=3", false},
		{"all list parts", "FREQ=YEARLY;BYMONTH=3,1;BYYEARDAY=-1,100;BYWEEKNO=20,1;BYHOUR=9;BYMINUTE=30;BYSECOND=0", false},
		{"byeaster", "FREQ=YEARLY;BYEASTER=0", true},
		{"byeaster with other parts", "FREQ=YEARLY;INTERVAL=2;BYEASTER=-2", true},
		// NormalizeRRule (and so CalDAV) drops DTSTART, so these reach the database only through
		// older data or a trusted raw copy. Both forms rrule-go parses are covered.
		{"dtstart two-line form", "DTSTART:20260101T090000Z\nRRULE:FREQ=DAILY", true},
		{"dtstart inline form", "FREQ=DAILY;DTSTART=20260101T090000Z", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := taskRepeatFromRRule(c.stored)
			require.NotNil(t, r, "a parseable stored rule always has a projection")
			assert.Equal(t, c.readOnly, r.ReadOnly)
		})
	}

	t.Run("NormalizeRRule strips DTSTART, so CalDAV never stores it", func(t *testing.T) {
		got, ok := NormalizeRRule("FREQ=DAILY;DTSTART=20260101T090000Z")
		require.True(t, ok)
		assert.Equal(t, "FREQ=DAILY", got)
		assert.False(t, taskRepeatFromRRule(got).ReadOnly)
	})

	t.Run("NormalizeRRule keeps BYEASTER, so CalDAV can store a read-only rule", func(t *testing.T) {
		got, ok := NormalizeRRule("FREQ=YEARLY;BYEASTER=0")
		require.True(t, ok)
		assert.True(t, taskRepeatFromRRule(got).ReadOnly)
	})
}

// TestTaskRepeatRepresentableRoundTrip pins the serialisation of a representable rule's projection.
func TestTaskRepeatRepresentableRoundTrip(t *testing.T) {
	cases := []struct {
		stored string
		want   string
	}{
		{"FREQ=DAILY", "FREQ=DAILY;INTERVAL=1"},
		{"FREQ=WEEKLY;BYDAY=FR,MO", "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR,MO"},
		{"FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1", "FREQ=MONTHLY;INTERVAL=1;BYSETPOS=-1;BYMONTHDAY=28,29,30,31"},
		{"FREQ=MONTHLY;BYMONTHDAY=01", "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1"},
		{"FREQ=DAILY;UNTIL=20261231T000000Z", "FREQ=DAILY;INTERVAL=1;UNTIL=20261231T000000Z"},
		{"FREQ=WEEKLY;WKST=SU;BYDAY=MO", "FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO"},
		{"FREQ=DAILY;COUNT=3", "FREQ=DAILY;INTERVAL=1;COUNT=3"},
	}
	for _, c := range cases {
		t.Run(c.stored, func(t *testing.T) {
			r := taskRepeatFromRRule(c.stored)
			require.NotNil(t, r)
			got, err := r.toRRule()
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
			assert.False(t, taskRepeatFromRRule(got).ReadOnly)
		})
	}
}

func TestTaskRepeatReadOnlyField(t *testing.T) {
	t.Run("serialises as read_only and is omitted when false", func(t *testing.T) {
		b, err := json.Marshal(&TaskRepeat{Freq: "yearly", Interval: 1, ReadOnly: true})
		require.NoError(t, err)
		assert.Contains(t, string(b), `"read_only":true`)

		b, err = json.Marshal(&TaskRepeat{Freq: "daily", Interval: 1})
		require.NoError(t, err)
		assert.NotContains(t, string(b), "read_only")
	})

	t.Run("does not affect serialisation to RRULE", func(t *testing.T) {
		plain, err := (&TaskRepeat{Freq: "daily", Interval: 2}).toRRule()
		require.NoError(t, err)
		flagged, err := (&TaskRepeat{Freq: "daily", Interval: 2, ReadOnly: true}).toRRule()
		require.NoError(t, err)
		assert.Equal(t, plain, flagged)
	})
}

func TestRepeatsEqual(t *testing.T) {
	until := func(s string) *string { return &s }
	base := func() *TaskRepeat { return &TaskRepeat{Freq: "weekly", Interval: 1, ByDay: []string{"mo", "fr"}} }

	cases := []struct {
		name  string
		a, b  *TaskRepeat
		equal bool
	}{
		{"both nil", nil, nil, true},
		{"left nil", nil, base(), false},
		{"right nil", base(), nil, false},
		{"identical", base(), base(), true},
		{"read_only ignored", base(), func() *TaskRepeat { r := base(); r.ReadOnly = true; return r }(), true},
		{"interval 0 reads as 1", &TaskRepeat{Freq: "daily"}, &TaskRepeat{Freq: "daily", Interval: 1}, true},
		{"interval differs", &TaskRepeat{Freq: "daily", Interval: 2}, &TaskRepeat{Freq: "daily", Interval: 1}, false},
		{"freq case-insensitive", &TaskRepeat{Freq: "DAILY", Interval: 1}, &TaskRepeat{Freq: "daily", Interval: 1}, true},
		{"freq differs", &TaskRepeat{Freq: "daily", Interval: 1}, &TaskRepeat{Freq: "weekly", Interval: 1}, false},
		{"by_day order and case", &TaskRepeat{Freq: "weekly", ByDay: []string{"FR", "mo"}}, base(), true},
		{"by_day differs", &TaskRepeat{Freq: "weekly", ByDay: []string{"tu", "fr"}}, base(), false},
		{"by_day nth differs", &TaskRepeat{Freq: "monthly", ByDay: []string{"1fr"}}, &TaskRepeat{Freq: "monthly", ByDay: []string{"-1fr"}}, false},
		{"nil and empty list", &TaskRepeat{Freq: "daily", ByMonthDay: []int{}}, &TaskRepeat{Freq: "daily"}, true},
		{"by_month_day order", &TaskRepeat{Freq: "monthly", ByMonthDay: []int{31, 28}}, &TaskRepeat{Freq: "monthly", ByMonthDay: []int{28, 31}}, true},
		{"by_month_day differs", &TaskRepeat{Freq: "monthly", ByMonthDay: []int{28}}, &TaskRepeat{Freq: "monthly", ByMonthDay: []int{29}}, false},
		{"by_month differs", &TaskRepeat{Freq: "yearly", ByMonth: []int{1}}, &TaskRepeat{Freq: "yearly", ByMonth: []int{2}}, false},
		{"by_year_day differs", &TaskRepeat{Freq: "yearly", ByYearDay: []int{1}}, &TaskRepeat{Freq: "yearly", ByYearDay: []int{2}}, false},
		{"by_week_no differs", &TaskRepeat{Freq: "yearly", ByWeekNo: []int{1}}, &TaskRepeat{Freq: "yearly", ByWeekNo: []int{2}}, false},
		{"by_set_pos differs", &TaskRepeat{Freq: "monthly", BySetPos: []int{-1}}, &TaskRepeat{Freq: "monthly", BySetPos: []int{1}}, false},
		{"by_hour differs", &TaskRepeat{Freq: "daily", ByHour: []int{9}}, &TaskRepeat{Freq: "daily", ByHour: []int{10}}, false},
		{"by_minute differs", &TaskRepeat{Freq: "daily", ByMinute: []int{0}}, &TaskRepeat{Freq: "daily", ByMinute: []int{30}}, false},
		{"by_second differs", &TaskRepeat{Freq: "daily", BySecond: []int{0}}, &TaskRepeat{Freq: "daily", BySecond: []int{30}}, false},
		{"count differs", &TaskRepeat{Freq: "daily", Count: 3}, &TaskRepeat{Freq: "daily", Count: 2}, false},
		{"until same instant, other offset", &TaskRepeat{Freq: "daily", Until: until("2026-12-31T19:00:00-05:00")}, &TaskRepeat{Freq: "daily", Until: until("2027-01-01T00:00:00Z")}, true},
		{"until differs", &TaskRepeat{Freq: "daily", Until: until("2026-12-31T00:00:00Z")}, &TaskRepeat{Freq: "daily", Until: until("2027-01-01T00:00:00Z")}, false},
		{"until set vs nil", &TaskRepeat{Freq: "daily", Until: until("2026-12-31T00:00:00Z")}, &TaskRepeat{Freq: "daily"}, false},
		{"wkst empty equals mo", &TaskRepeat{Freq: "weekly", Wkst: "MO"}, &TaskRepeat{Freq: "weekly"}, true},
		{"wkst differs", &TaskRepeat{Freq: "weekly", Wkst: "su"}, &TaskRepeat{Freq: "weekly"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.equal, repeatsEqual(c.a, c.b))
			assert.Equal(t, c.equal, repeatsEqual(c.b, c.a), "symmetric")
		})
	}

	t.Run("does not mutate its arguments", func(t *testing.T) {
		a := &TaskRepeat{Freq: "WEEKLY", ByDay: []string{"FR", "MO"}, ByMonthDay: []int{31, 28}}
		repeatsEqual(a, base())
		assert.Equal(t, "WEEKLY", a.Freq)
		assert.Equal(t, []string{"FR", "MO"}, a.ByDay)
		assert.Equal(t, []int{31, 28}, a.ByMonthDay)
	})
}
