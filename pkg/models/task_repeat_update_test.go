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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// TestResolveIncomingRepeat pins the write rules for a task's recurrence. Masked updates never
// reach this function.
func TestResolveIncomingRepeat(t *testing.T) {
	const (
		simple     = "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,FR"
		easter     = "FREQ=YEARLY;BYEASTER=0"
		monthEnd   = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1"
		countThree = "FREQ=DAILY;INTERVAL=1;COUNT=3"
		// Barred by rrulePolicyRejects: sub-daily with a BY* part.
		rejected = "FREQ=HOURLY;BYHOUR=9"
	)
	readOnlyEdit := "can be removed (send repeat: null) or replaced with repeats_rrule (v2) or over CalDAV"

	cases := []struct {
		name        string
		stored      string
		incoming    *TaskRepeat
		incomingRaw string
		want        repeatDecision
		wantErr     string // substring of the ErrInvalidData message; "" means no error
	}{
		// Create (stored == "").
		{name: "create with nil stores nothing", want: repeatDecision{}},
		{
			name:     "create with a structured rule",
			incoming: &TaskRepeat{Freq: "daily", Interval: 2},
			want:     repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=2"},
		},
		{
			name:     "create with a policy-rejected rule",
			incoming: &TaskRepeat{Freq: "hourly", Interval: 1, ByHour: []int{9}},
			wantErr:  "sub-daily",
		},
		{
			name:     "create with an invalid frequency",
			incoming: &TaskRepeat{Freq: "fortnightly"},
			wantErr:  "Invalid frequency",
		},

		// An equal echo keeps the stored text byte-identical.
		{
			name:     "equal echo of a simple rule",
			stored:   simple,
			incoming: &TaskRepeat{Freq: "weekly", Interval: 1, ByDay: []string{"mo", "fr"}},
			want:     repeatDecision{Repeats: simple, Keep: true},
		},
		{
			name:     "equal echo with a server-decremented count",
			stored:   countThree,
			incoming: &TaskRepeat{Freq: "daily", Interval: 1, Count: 3},
			want:     repeatDecision{Repeats: countThree, Keep: true},
		},
		{
			name:     "equal echo with a different list order",
			stored:   simple,
			incoming: &TaskRepeat{Freq: "weekly", Interval: 1, ByDay: []string{"fr", "mo"}},
			want:     repeatDecision{Repeats: simple, Keep: true},
		},
		{
			name:     "equal echo of the month-end rule keeps its part order",
			stored:   monthEnd,
			incoming: taskRepeatFromRRule(monthEnd),
			want:     repeatDecision{Repeats: monthEnd, Keep: true},
		},
		{
			name:     "equal echo of a stored rule the policy now rejects",
			stored:   rejected,
			incoming: &TaskRepeat{Freq: "hourly", Interval: 1, ByHour: []int{9}},
			want:     repeatDecision{Repeats: rejected, Keep: true},
		},
		{
			name:     "echo of a read-only rule carrying the flag",
			stored:   easter,
			incoming: &TaskRepeat{Freq: "yearly", Interval: 1, ReadOnly: true},
			want:     repeatDecision{Repeats: easter, Keep: true},
		},
		{
			name:     "echo of a read-only rule without the flag",
			stored:   easter,
			incoming: &TaskRepeat{Freq: "yearly", Interval: 1},
			want:     repeatDecision{Repeats: easter, Keep: true},
		},

		// Stale COUNT: the same rule with a higher COUNT than the
		// stored remaining one is a client echoing a COUNT read before the server decremented
		// it, so the stored COUNT stays. A lower COUNT or any other change is a deliberate edit.
		{
			name:     "stale higher count on an otherwise equal rule keeps the stored count",
			stored:   "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incoming: &TaskRepeat{Freq: "daily", Interval: 1, Count: 3},
			want:     repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1;COUNT=2", Keep: true},
		},
		{
			name:     "stale higher count on a read-only rule keeps it instead of refusing",
			stored:   "FREQ=YEARLY;BYEASTER=0;COUNT=2",
			incoming: &TaskRepeat{Freq: "yearly", Interval: 1, Count: 3},
			want:     repeatDecision{Repeats: "FREQ=YEARLY;BYEASTER=0;COUNT=2", Keep: true},
		},
		{
			name:     "lower count is a deliberate edit",
			stored:   "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incoming: &TaskRepeat{Freq: "daily", Interval: 1, Count: 1},
			want:     repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1;COUNT=1"},
		},
		{
			name:     "removing the count is a deliberate edit",
			stored:   "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incoming: &TaskRepeat{Freq: "daily", Interval: 1},
			want:     repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1"},
		},
		{
			name:     "higher count with a changed freq is a deliberate edit",
			stored:   "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incoming: &TaskRepeat{Freq: "weekly", Interval: 1, Count: 3},
			want:     repeatDecision{Repeats: "FREQ=WEEKLY;INTERVAL=1;COUNT=3"},
		},
		{
			name:     "higher count with a changed interval is a deliberate edit",
			stored:   "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incoming: &TaskRepeat{Freq: "daily", Interval: 2, Count: 3},
			want:     repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=2;COUNT=3"},
		},
		{
			name:     "adding a count to an unlimited rule is a deliberate edit",
			stored:   "FREQ=DAILY;INTERVAL=1",
			incoming: &TaskRepeat{Freq: "daily", Interval: 1, Count: 3},
			want:     repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1;COUNT=3"},
		},

		// The same stale-COUNT rule for raw RRULE text from CalDAV.
		{
			name:        "raw rule with a stale higher count keeps the stored count",
			stored:      "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incomingRaw: "FREQ=DAILY;COUNT=3",
			want:        repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1;COUNT=2", Keep: true},
		},
		{
			name:        "raw read-only rule with a stale higher count keeps the stored count",
			stored:      "FREQ=YEARLY;BYEASTER=0;COUNT=2",
			incomingRaw: "FREQ=YEARLY;BYEASTER=0;COUNT=4",
			want:        repeatDecision{Repeats: "FREQ=YEARLY;BYEASTER=0;COUNT=2", Keep: true},
		},
		{
			name:        "raw rule with a stale count but a different BYEASTER replaces",
			stored:      "FREQ=YEARLY;BYEASTER=0;COUNT=2",
			incomingRaw: "FREQ=YEARLY;BYEASTER=1;COUNT=4",
			want:        repeatDecision{Repeats: "FREQ=YEARLY;BYEASTER=1;COUNT=4"},
		},
		{
			name:        "raw rule with a lower count replaces",
			stored:      "FREQ=DAILY;INTERVAL=1;COUNT=2",
			incomingRaw: "FREQ=DAILY;INTERVAL=1;COUNT=1",
			want:        repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1;COUNT=1"},
		},
		{
			name:        "raw rule adding a count to an unlimited rule replaces",
			stored:      "FREQ=DAILY;INTERVAL=1",
			incomingRaw: "FREQ=DAILY;INTERVAL=1;COUNT=3",
			want:        repeatDecision{Repeats: "FREQ=DAILY;INTERVAL=1;COUNT=3"},
		},

		// A read-only rule refuses anything but an equal echo or a clear.
		{
			name:     "edit of a read-only rule",
			stored:   easter,
			incoming: &TaskRepeat{Freq: "yearly", Interval: 2},
			wantErr:  readOnlyEdit,
		},
		{
			name:     "edit of a read-only rule that only drops the flag's parts",
			stored:   "FREQ=DAILY;DTSTART=20260101T090000Z",
			incoming: &TaskRepeat{Freq: "weekly", Interval: 1},
			wantErr:  readOnlyEdit,
		},

		// Nil clears, whatever the stored rule.
		{name: "nil on a read-only rule clears", stored: easter, want: repeatDecision{}},
		{name: "nil on a simple rule clears", stored: simple, want: repeatDecision{}},

		// A different valid rule replaces; COUNT comes from the request.
		{
			name:     "deliberate replacement",
			stored:   simple,
			incoming: &TaskRepeat{Freq: "monthly", Interval: 1, ByMonthDay: []int{15}},
			want:     repeatDecision{Repeats: "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15"},
		},
		{
			name:     "replacement takes count from the request",
			stored:   countThree,
			incoming: &TaskRepeat{Freq: "weekly", Interval: 1, Count: 5},
			want:     repeatDecision{Repeats: "FREQ=WEEKLY;INTERVAL=1;COUNT=5"},
		},
		{
			// The tradeoff of the stale-COUNT rule: raising only the COUNT of an unchanged
			// rule looks exactly like a stale echo, so the stored COUNT is kept.
			name:     "raising only the count of an unchanged rule keeps the stored count",
			stored:   countThree,
			incoming: &TaskRepeat{Freq: "daily", Interval: 1, Count: 5},
			want:     repeatDecision{Repeats: countThree, Keep: true},
		},

		// An invalid or policy-rejected replacement is refused.
		{
			name:     "policy-rejected replacement",
			stored:   simple,
			incoming: &TaskRepeat{Freq: "minutely", Interval: 1, ByDay: []string{"mo"}},
			wantErr:  "sub-daily",
		},
		{
			name:     "replacement with a bad until",
			stored:   simple,
			incoming: &TaskRepeat{Freq: "daily", Until: func() *string { s := "tomorrow"; return &s }()},
			wantErr:  "Invalid until",
		},

		// Raw RRULE text from a trusted caller (CalDAV, duplication, import).
		{
			name:        "raw equal to the stored rule keeps it without validation",
			stored:      rejected,
			incomingRaw: rejected,
			want:        repeatDecision{Repeats: rejected, Keep: true},
		},
		{
			name:        "raw from CalDAV replaces the stored rule",
			stored:      simple,
			incomingRaw: easter,
			want:        repeatDecision{Repeats: easter},
		},
		{
			name:        "raw on create is stored verbatim",
			incomingRaw: monthEnd,
			want:        repeatDecision{Repeats: monthEnd},
		},
		{
			name:        "raw with an agreeing structured repeat",
			stored:      simple,
			incoming:    taskRepeatFromRRule(easter),
			incomingRaw: easter,
			want:        repeatDecision{Repeats: easter},
		},
		{
			name:        "raw with a disagreeing structured repeat",
			stored:      simple,
			incoming:    &TaskRepeat{Freq: "daily", Interval: 1},
			incomingRaw: easter,
			wantErr:     "disagree",
		},
		{
			name:        "raw that the policy rejects",
			stored:      simple,
			incomingRaw: rejected,
			wantErr:     "sub-daily",
		},
		{
			name:        "raw that does not parse",
			stored:      simple,
			incomingRaw: "not a rule",
			wantErr:     "Invalid RRULE",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolveIncomingRepeat(c.stored, c.incoming, c.incomingRaw)
			if c.wantErr != "" {
				require.Error(t, err)
				require.True(t, IsErrInvalidData(err), "want ErrInvalidData, got %T: %v", err, err)
				assert.Contains(t, err.(ErrInvalidData).Message, c.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}

	t.Run("read-only refusal names both ways out", func(t *testing.T) {
		_, err := resolveIncomingRepeat(easter, &TaskRepeat{Freq: "yearly", Interval: 2}, "")
		require.Error(t, err)
		msg := err.(ErrInvalidData).Message
		assert.Contains(t, msg, "repeat: null")
		assert.Contains(t, msg, "CalDAV")
	})
}

// TestTaskRepeatRuleTextWrites pins the v2 rule text (Task.RepeatsText) through the real Create
// and Update paths. Each "kept" case is paired with a "replaced" case on the same fixture.
func TestTaskRepeatRuleTextWrites(t *testing.T) {
	u := &user.User{ID: 1}
	const (
		easter = "FREQ=YEARLY;BYEASTER=1"
		// Non-canonical part order, so keeping it proves the text is byte-identical.
		easterReordered = "BYEASTER=1;FREQ=YEARLY"
		weekly          = "FREQ=WEEKLY;BYDAY=MO"
	)

	// readWithText reads the fixture task as a v2 client would see it: the structured
	// projection plus the stored text as repeats_rrule, filled from the read.
	readWithText := func(t *testing.T, s *xorm.Session) *Task {
		t.Helper()
		upd := readForEcho(t, s, u)
		upd.RepeatsText = strPtr(storedRepeats(t, s, repeatingFixtureTask))
		return upd
	}

	t.Run("create with text stores the canonical rule and a read-only projection", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: "easter", ProjectID: 1, RepeatsText: strPtr("rrule:freq=yearly;byeaster=1")}
		require.NoError(t, task.Create(s, u))

		assert.Equal(t, easter, storedRepeats(t, s, task.ID))
		require.NotNil(t, task.Repeat, "the response carries the projection")
		assert.True(t, task.Repeat.ReadOnly)
	})
	t.Run("create with refused text fails and stores nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: "bad", ProjectID: 1, RepeatsText: strPtr("FREQ=DAILY;COUNT=5;UNTIL=20271231T235959Z")}
		err := task.Create(s, u)
		require.Error(t, err)
		require.True(t, IsErrInvalidData(err), "got %v", err)
		assert.Contains(t, err.(ErrInvalidData).Message, "COUNT and UNTIL cannot both be set")
		assert.Zero(t, task.ID)
	})
	t.Run("create with empty text and a structured repeat stores the repeat", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := &Task{Title: "structured", ProjectID: 1, Repeat: &TaskRepeat{Freq: "daily", Interval: 2}, RepeatsText: strPtr("")}
		require.NoError(t, task.Create(s, u))
		assert.Equal(t, "FREQ=DAILY;INTERVAL=2", storedRepeats(t, s, task.ID))
	})

	t.Run("text echo with repeat nil clears (the frontend clear shape)", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, easterReordered)

		upd := readWithText(t, s)
		upd.Repeat = nil
		require.NoError(t, upd.Update(s, u))

		assert.Empty(t, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("text echo with the projection keeps the rule byte-identical (a PATCH title edit)", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, easterReordered)

		upd := readWithText(t, s)
		require.NotNil(t, upd.Repeat)
		upd.Title = "renamed"
		require.NoError(t, upd.Update(s, u))

		assert.Equal(t, easterReordered, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("changed text with the old read-only projection replaces the rule", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, easterReordered)

		upd := readWithText(t, s)
		require.True(t, upd.Repeat.ReadOnly)
		upd.RepeatsText = strPtr("FREQ=YEARLY;BYEASTER=2")
		require.NoError(t, upd.Update(s, u))

		assert.Equal(t, "FREQ=YEARLY;BYEASTER=2", storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("changed text with the old projection replaces the rule (a PATCH text edit)", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.RepeatsText = strPtr("FREQ=WEEKLY;BYDAY=TU")
		require.NoError(t, upd.Update(s, u))

		assert.Equal(t, "FREQ=WEEKLY;BYDAY=TU", storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("empty text with repeat nil clears", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.Repeat = nil
		upd.RepeatsText = strPtr("")
		require.NoError(t, upd.Update(s, u))

		assert.Empty(t, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("empty text with the echoed projection clears", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.RepeatsText = strPtr("")
		require.NoError(t, upd.Update(s, u))

		assert.Empty(t, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("empty text with a changed repeat is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.Repeat = &TaskRepeat{Freq: "daily", Interval: 1}
		upd.RepeatsText = strPtr("")
		err := upd.Update(s, u)
		require.Error(t, err)
		require.True(t, IsErrInvalidData(err), "got %v", err)
		assert.Contains(t, err.(ErrInvalidData).Message, "repeats_rrule clears the recurrence but repeat changes it")
		assert.Equal(t, weekly, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("changed text with a disagreeing changed repeat is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.Repeat = &TaskRepeat{Freq: "daily", Interval: 1}
		upd.RepeatsText = strPtr("FREQ=WEEKLY;BYDAY=TU")
		err := upd.Update(s, u)
		require.Error(t, err)
		require.True(t, IsErrInvalidData(err), "got %v", err)
		assert.Equal(t, "The structured repeat (repeat) and the rule text (repeats_rrule) disagree; set only one of them.", err.(ErrInvalidData).Message)
		assert.Equal(t, weekly, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("changed text with an agreeing changed repeat stores the text", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.Repeat = &TaskRepeat{Freq: "weekly", Interval: 1, ByDay: []string{"tu"}}
		upd.RepeatsText = strPtr("FREQ=WEEKLY;BYDAY=TU")
		require.NoError(t, upd.Update(s, u))
		assert.Equal(t, "FREQ=WEEKLY;BYDAY=TU", storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("refused text fails with 2002 and leaves the stored rule", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := readWithText(t, s)
		upd.RepeatsText = strPtr("FREQ=DAILY;UNTIL=20271231")
		err := upd.Update(s, u)
		require.Error(t, err)
		require.True(t, IsErrInvalidData(err), "got %v", err)
		assert.Equal(t, ErrCodeInvalidData, err.(ErrInvalidData).HTTPError().Code)
		assert.Equal(t, weekly, storedRepeats(t, s, repeatingFixtureTask))
	})

	// Text ValidateRuleText refuses is never read as a respelled echo, whatever
	// repeat says. Each stored rule is spelled non-canonically, so a kept echo (byte-identical)
	// is told apart from a replacement (the canonical spelling).
	invalidEchoes := []struct {
		name, stored, invalid, respelled, wantErr string
	}{
		{
			name:      "floating UNTIL",
			stored:    "UNTIL=20271231T235959Z;FREQ=DAILY",
			invalid:   "FREQ=DAILY;UNTIL=20271231T235959",
			respelled: "freq=daily;until=20271231t235959z",
			wantErr:   "must be a UTC date-time",
		},
		{
			name:      "COUNT=0",
			stored:    "INTERVAL=1;FREQ=DAILY",
			invalid:   "INTERVAL=1;FREQ=DAILY;COUNT=0",
			respelled: "freq=daily;interval=1",
			wantErr:   "COUNT 0 must be a positive integer",
		},
	}
	for _, c := range invalidEchoes {
		for _, nilRepeat := range []bool{false, true} {
			shape := "an echoed repeat"
			if nilRepeat {
				shape = "a nil repeat"
			}
			t.Run(c.name+" with "+shape+" is refused with 2002", func(t *testing.T) {
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				storeRepeats(t, s, c.stored)

				upd := readWithText(t, s)
				require.NotNil(t, upd.Repeat)
				if nilRepeat {
					upd.Repeat = nil
				}
				upd.RepeatsText = strPtr(c.invalid)
				err := upd.Update(s, u)
				require.Error(t, err)
				require.True(t, IsErrInvalidData(err), "got %v", err)
				assert.Equal(t, ErrCodeInvalidData, err.(ErrInvalidData).HTTPError().Code)
				assert.Contains(t, err.(ErrInvalidData).Message, c.wantErr)
				assert.Equal(t, c.stored, storedRepeats(t, s, repeatingFixtureTask))
			})
			t.Run(c.name+": the valid respelling with "+shape+" is still an echo", func(t *testing.T) {
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				storeRepeats(t, s, c.stored)

				upd := readWithText(t, s)
				require.NotNil(t, upd.Repeat)
				if nilRepeat {
					upd.Repeat = nil
				}
				upd.RepeatsText = strPtr(c.respelled)
				require.NoError(t, upd.Update(s, u))
				if nilRepeat {
					// An echo hands the decision to repeat, and a nil repeat clears; changed
					// text would have stored the rule instead.
					assert.Empty(t, storedRepeats(t, s, repeatingFixtureTask))
				} else {
					assert.Equal(t, c.stored, storedRepeats(t, s, repeatingFixtureTask), "kept byte-identical")
				}
			})
		}
	}

	t.Run("no text with a raw rule (the CalDAV shape) behaves as before", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		// The raw rule from a CalDAV client replaces the stored one ...
		upd := &Task{ID: repeatingFixtureTask, Title: "caldav", Repeats: "INTERVAL=2;FREQ=DAILY"}
		require.NoError(t, upd.Update(s, u))
		assert.Equal(t, "INTERVAL=2;FREQ=DAILY", storedRepeats(t, s, repeatingFixtureTask), "stored verbatim")

		// ... and the same raw rule sent again keeps it.
		upd = &Task{ID: repeatingFixtureTask, Title: "caldav again", Repeats: "INTERVAL=2;FREQ=DAILY"}
		require.NoError(t, upd.Update(s, u))
		assert.Equal(t, "INTERVAL=2;FREQ=DAILY", storedRepeats(t, s, repeatingFixtureTask))
	})

	t.Run("a bulk update leaves the stored text alone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, easterReordered)

		_, err := updateTasks(s, u, &Task{Title: "bulk renamed"}, []int64{repeatingFixtureTask}, []string{"title"})
		require.NoError(t, err)
		assert.Equal(t, easterReordered, storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("a mask naming repeats_rrule applies the text to the repeats column", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := &Task{ID: repeatingFixtureTask, RepeatsText: strPtr("FREQ=DAILY;INTERVAL=3")}
		require.NoError(t, upd.updateSingleTask(s, u, []string{"repeats_rrule"}))
		assert.Equal(t, "FREQ=DAILY;INTERVAL=3", storedRepeats(t, s, repeatingFixtureTask))
	})
	t.Run("a mask excluding recurrence ignores the text", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		storeRepeats(t, s, weekly)

		upd := &Task{ID: repeatingFixtureTask, Title: "renamed only", RepeatsText: strPtr("FREQ=DAILY;INTERVAL=3")}
		require.NoError(t, upd.updateSingleTask(s, u, []string{"title"}))
		assert.Equal(t, weekly, storedRepeats(t, s, repeatingFixtureTask))
	})
}

// TestDecrementedRule pins the COUNT decrement completion applies (setTaskDatesRRule) and the
// occurrence query reuses.
func TestDecrementedRule(t *testing.T) {
	next, ok := decrementedRule("FREQ=DAILY;INTERVAL=1;COUNT=3")
	assert.True(t, ok)
	assert.Equal(t, "FREQ=DAILY;INTERVAL=1;COUNT=2", next)

	next, ok = decrementedRule("COUNT=2;FREQ=WEEKLY")
	assert.True(t, ok)
	assert.Equal(t, "FREQ=WEEKLY;COUNT=1", next, "serialised canonically, as completion stores it")

	for _, rule := range []string{"FREQ=DAILY;COUNT=1", "FREQ=DAILY", "", "not a rule"} {
		_, ok := decrementedRule(rule)
		assert.False(t, ok, rule)
	}
}
