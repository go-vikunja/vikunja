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
	readOnlyEdit := "can be removed (send repeat: null) or replaced over CalDAV"

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
