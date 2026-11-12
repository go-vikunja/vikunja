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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// occurrenceFixture is a task as stored on task 1, and what to ask of it.
type occurrenceFixture struct {
	name           string
	rule           string
	due            time.Time
	fromCompletion bool
	start, end     time.Time
	reminder       time.Time // an absolute reminder; zero for none
	relative       bool      // adds a reminder one hour before the due date
	done           bool      // stored already done
	now            time.Time
	n              int
	wantLen        int    // how many dates completion produces: the non-vacuity check
	wantReason     string // the loop's reason; "" when all n are found
}

var (
	occNow = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	occDue = time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
)

// occurrenceZoneCorpus runs in UTC, America/New_York and Asia/Kolkata, set as the acting user's
// zone. Several rows cross New York's 2026-11-01 fall-back as well.
var occurrenceZoneCorpus = []occurrenceFixture{
	// The plain frequencies: each completion moves the due date.
	{name: "daily", rule: "FREQ=DAILY", due: occDue, now: occNow, n: 5, wantLen: 5},
	// The same task already done: marking it done again moves nothing.
	{name: "daily, already done", rule: "FREQ=DAILY", done: true, due: occDue, now: occNow, n: 5, wantLen: 0, wantReason: reasonTaskDone},
	{name: "weekly on two days", rule: "FREQ=WEEKLY;BYDAY=MO,TH", due: occDue, now: occNow, n: 5, wantLen: 5},
	// Months without a 31st are skipped.
	{name: "monthly on the 31st", rule: "FREQ=MONTHLY;BYMONTHDAY=31", due: occDue, now: occNow, n: 5, wantLen: 5},
	{name: "monthly last Friday", rule: "FREQ=MONTHLY;BYDAY=-1FR", due: occDue, now: occNow, n: 5, wantLen: 5},
	// Only leap years match.
	{name: "yearly on 29 February", rule: "FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=29", due: occDue, now: occNow, n: 3, wantLen: 3},
	// rrule-go's extension, as a CalDAV client may store it.
	{name: "yearly Easter Monday", rule: "FREQ=YEARLY;BYEASTER=1", due: occDue, now: occNow, n: 3, wantLen: 3},
	// Sub-daily: elapsed time, not the wall clock.
	{name: "every 6 hours", rule: "FREQ=HOURLY;INTERVAL=6", due: occDue, now: occNow, n: 5, wantLen: 5},
	// Start, end and reminders are carried; a relative one is not the engine's.
	{
		name: "daily with start, end and reminders", rule: "FREQ=DAILY", due: occDue,
		start: occDue.Add(-2 * time.Hour), end: occDue.Add(time.Hour), reminder: occDue.Add(-30 * time.Minute), relative: true,
		now: occNow, n: 5, wantLen: 5,
	},
	// COUNT is decremented after each successful step: three remain, the current one included.
	{name: "COUNT=3 asked for 5", rule: "FREQ=DAILY;COUNT=3", due: occDue, now: occNow, n: 5, wantLen: 2, wantReason: string(reasonCountExhausted)},
	{name: "COUNT=1", rule: "FREQ=DAILY;COUNT=1", due: occDue, now: occNow, n: 5, wantLen: 0, wantReason: string(reasonCountExhausted)},
	// With a fixed now, UNTIL ending within N reports no_occurrence (paula finding 5).
	{name: "UNTIL within N", rule: "FREQ=DAILY;UNTIL=20261025T090000Z", due: occDue, now: occNow, n: 5, wantLen: 2, wantReason: string(reasonNoOccurrence)},
	{
		name: "UNTIL already passed", rule: "FREQ=DAILY;UNTIL=20261020T000000Z", due: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC),
		now: occNow, n: 5, wantLen: 0, wantReason: string(reasonUntilPassed),
	},
	// Repeat from completion: each step is completed on the previous due date.
	{name: "daily from completion", rule: "FREQ=DAILY", fromCompletion: true, due: occDue, now: occNow, n: 5, wantLen: 5},
	{name: "weekly from completion", rule: "FREQ=WEEKLY", fromCompletion: true, due: occDue, now: occNow, n: 4, wantLen: 4},
	// A request clock with fractions: the query matches what completion stores, to the second.
	{
		name: "fractional clock", rule: "FREQ=DAILY", due: time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC),
		now: occNow.Add(123 * time.Millisecond), n: 3, wantLen: 3,
	},
	{
		name: "fractional clock from completion", rule: "FREQ=HOURLY;INTERVAL=6", fromCompletion: true, due: occDue,
		now: occNow.Add(123 * time.Millisecond), n: 3, wantLen: 3,
	},
}

// occurrenceNewYorkCorpus runs with the acting user in America/New_York. The daylight-saving rows
// assert equality with completion; the dates themselves, including the skipped spring-forward
// occurrence (RFC 5545 §3.3.10), are pinned in TestTaskOccurrencesSpringForwardSkip.
var occurrenceNewYorkCorpus = []occurrenceFixture{
	{name: "daily 20:00 across spring forward", rule: "FREQ=DAILY", due: time.Date(2027, 3, 12, 1, 0, 0, 0, time.UTC), now: time.Date(2027, 3, 11, 12, 0, 0, 0, time.UTC), n: 7, wantLen: 7},
	{name: "daily 02:30 across spring forward", rule: "FREQ=DAILY", due: time.Date(2027, 3, 11, 7, 30, 0, 0, time.UTC), now: time.Date(2027, 3, 11, 6, 0, 0, 0, time.UTC), n: 7, wantLen: 7},
	{name: "daily 20:00 across fall back", rule: "FREQ=DAILY", due: time.Date(2027, 11, 5, 0, 0, 0, 0, time.UTC), now: time.Date(2027, 11, 4, 12, 0, 0, 0, time.UTC), n: 7, wantLen: 7},
	{name: "daily 02:30 across fall back", rule: "FREQ=DAILY", due: time.Date(2027, 11, 4, 6, 30, 0, 0, time.UTC), now: time.Date(2027, 11, 4, 5, 0, 0, 0, time.UTC), n: 7, wantLen: 7},
	// 01:30 occurs twice on 2027-11-07.
	{name: "daily 01:30 across the fall-back overlap", rule: "FREQ=DAILY", due: time.Date(2027, 11, 4, 5, 30, 0, 0, time.UTC), now: time.Date(2027, 11, 4, 4, 0, 0, 0, time.UTC), n: 7, wantLen: 7},
	{name: "daily 02:30 from completion across spring forward", rule: "FREQ=DAILY", fromCompletion: true, due: time.Date(2027, 3, 12, 7, 30, 0, 0, time.UTC), now: time.Date(2027, 3, 12, 7, 30, 0, 0, time.UTC), n: 5, wantLen: 5},
}

// TestTaskOccurrencesEqualCompletion is the query's core invariant: on every
// fixture, the query's dates equal those of N real completions through the update path, read back
// from the database, and the loop's carried state equals the stored task after each completion.
func TestTaskOccurrencesEqualCompletion(t *testing.T) {
	for _, zone := range []string{"UTC", "America/New_York", "Asia/Kolkata"} {
		for _, fx := range occurrenceZoneCorpus {
			t.Run(zone+"/"+fx.name, func(t *testing.T) {
				assertOccurrencesEqualCompletion(t, fx, zone, "")
			})
		}
	}
	for _, fx := range occurrenceNewYorkCorpus {
		t.Run("America/New_York/"+fx.name, func(t *testing.T) {
			assertOccurrencesEqualCompletion(t, fx, "America/New_York", "")
		})
	}

	// The zone order is completion's: the acting user, then the project owner. User 2, sharing
	// project 1, acts in these rows.
	weekly := occurrenceFixture{
		rule: "FREQ=WEEKLY;BYDAY=MO,WE,FR", due: time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC),
		now: time.Date(2026, 10, 27, 0, 5, 0, 0, time.UTC), n: 5, wantLen: 5,
	}
	t.Run("acting user's zone over the owner's", func(t *testing.T) {
		assertOccurrencesEqualCompletion(t, weekly, "America/New_York", "Asia/Kolkata")
	})
	t.Run("acting user without a zone falls back to the owner's", func(t *testing.T) {
		assertOccurrencesEqualCompletion(t, weekly, "", "Asia/Kolkata")
	})
}

// TestTaskOccurrencesZoneDiscriminates is the positive control for the zone rows: the weekly
// fixture's dates differ between New York and the server zone, so a loop that ignored the zone
// would fail them.
func TestTaskOccurrencesZoneDiscriminates(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}
	task := &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=WEEKLY;BYDAY=MO,WE,FR", DueDate: time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)}
	now := time.Date(2026, 10, 27, 0, 5, 0, 0, time.UTC)

	setUserZone(t, s, 1, "America/New_York")
	ny, err := TaskOccurrences(s, u, task, 3, now)
	require.NoError(t, err)
	setUserZone(t, s, 1, "")
	server, err := TaskOccurrences(s, u, task, 3, now)
	require.NoError(t, err)

	require.Len(t, ny.Occurrences, 3)
	require.Len(t, server.Occurrences, 3)
	assert.False(t, ny.Occurrences[0].Equal(server.Occurrences[0]), "New York %v, server %v", ny.Occurrences[0], server.Occurrences[0])
}

func assertOccurrencesEqualCompletion(t *testing.T, fx occurrenceFixture, userZone, ownerZone string) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}
	if ownerZone != "" {
		// User 2 acts, with write access to project 1; its owner, user 1, has ownerZone.
		u = &user.User{ID: 2, Username: "user2"}
		_, err := s.Insert(&ProjectUser{UserID: 2, ProjectID: 1, Permission: PermissionWrite})
		require.NoError(t, err)
		setUserZone(t, s, 1, ownerZone)
	}
	setUserZone(t, s, u.ID, userZone)
	seedOccurrenceTask(t, s, fx)

	task := &Task{ID: 1}
	require.NoError(t, task.ReadOne(s, u))
	answer, err := TaskOccurrences(s, u, task, fx.n, fx.now)
	require.NoError(t, err)
	carried := carriedStates(task, resolveTaskZone(s, u, task.ProjectID), fx.n, fx.now)

	// The query is an on-time-completion forecast, so the real
	// completions it is compared with happen on time too. For a task that repeats from the current
	// date with a future due date, the first one is on that due date instead of at now; this changes
	// only the rows "daily from completion", "weekly from completion" and "fractional clock from
	// completion", which encoded completion at now. Every other row completes at now.
	clock := fx.now
	if fx.fromCompletion && fx.due.After(fx.now) {
		clock = fx.due
	}
	completed, stopped, last := completeTask1(t, s, u, fx.n, clock, fx.fromCompletion)

	require.Len(t, completed, fx.wantLen, "completion's dates: the fixture must discriminate")
	require.Len(t, answer.Occurrences, len(completed), "query %v, completion %v", answer.Occurrences, completed)
	for i := range completed {
		assert.True(t, answer.Occurrences[i].Equal(completed[i].DueDate), "occurrence %d: query %v, completion %v", i, answer.Occurrences[i], completed[i].DueDate)
	}
	if stopped {
		assert.NotEmpty(t, answer.Reason, "completion stopped, so the query must say why")
		lastDue := fx.due
		if len(completed) > 0 {
			lastDue = completed[len(completed)-1].DueDate
		}
		assert.True(t, last.DueDate.Equal(lastDue), "the completion that stopped left the due date: got %v, want %v", last.DueDate, lastDue)
	} else {
		assert.Empty(t, answer.Reason)
	}
	assert.Equal(t, fx.wantReason, answer.Reason)

	// The carried state, step by step, against the stored task after each completion.
	require.Len(t, carried, len(completed))
	for i, stored := range completed {
		in := carried[i]
		assert.True(t, in.Due.At.Equal(answer.Occurrences[i]), "step %d: carried due", i)
		assert.Equal(t, stored.Repeats, in.Rule, "step %d: rule", i)
		assert.True(t, in.Start.At.Equal(stored.StartDate), "step %d: start: carried %v, stored %v", i, in.Start.At, stored.StartDate)
		assert.True(t, in.End.At.Equal(stored.EndDate), "step %d: end: carried %v, stored %v", i, in.End.At, stored.EndDate)
		absolute := absoluteReminders(stored.Reminders)
		require.Len(t, in.Reminders, len(absolute), "step %d: reminders", i)
		for j, rem := range absolute {
			assert.True(t, in.Reminders[j].At.Equal(rem.Reminder), "step %d: reminder %d: carried %v, stored %v", i, j, in.Reminders[j].At, rem.Reminder)
		}
	}
}

// seedOccurrenceTask stores fx on task 1 directly, as a CalDAV client or an earlier completion
// would have stored it.
func seedOccurrenceTask(t *testing.T, s *xorm.Session, fx occurrenceFixture) {
	t.Helper()
	cols := []string{"repeats", "repeats_from_current_date", "due_date", "done"}
	if !fx.start.IsZero() {
		cols = append(cols, "start_date")
	}
	if !fx.end.IsZero() {
		cols = append(cols, "end_date")
	}
	_, err := s.ID(1).Cols(cols...).Update(&Task{
		Done: fx.done, Repeats: fx.rule, RepeatsFromCurrentDate: fx.fromCompletion, DueDate: fx.due,
		StartDate: fx.start, EndDate: fx.end,
	})
	require.NoError(t, err)
	if !fx.reminder.IsZero() {
		_, err = s.Insert(&TaskReminder{TaskID: 1, Reminder: fx.reminder})
		require.NoError(t, err)
	}
	if fx.relative {
		_, err = s.Insert(&TaskReminder{TaskID: 1, Reminder: fx.due.Add(-time.Hour), RelativeTo: ReminderRelationDueDate, RelativePeriod: -3600})
		require.NoError(t, err)
	}
}

// carriedStates is the loop's input after each successful step, from the same start the query
// uses (forecastFirstCompletion): nextOccurrence and nextOccurrenceStep, as stepOccurrences calls
// them, after the same refusals TaskOccurrences applies.
func carriedStates(task *Task, zone *time.Location, n int, now time.Time) []nextOccurrenceInput {
	if occurrencesRefusal(task) != "" {
		return nil
	}
	in := completionInput(task, absoluteReminders(task.Reminders), forecastFirstCompletion(task, now), zone)
	var states []nextOccurrenceInput
	for range n {
		res := nextOccurrence(in)
		if res.Reason != "" {
			break
		}
		in = nextOccurrenceStep(in, res)
		states = append(states, in)
	}
	return states
}

// completeTask1 completes task 1 up to n times through the field-masked update path, as a v2
// {"done": true} does, and returns the stored task, with its reminders, after each completion that
// reopened it. The clock is now; for a task that repeats from the current date, each later
// completion happens on the stored due date. stopped is true when a completion left the task done;
// last is then the task as that completion stored it.
func completeTask1(t *testing.T, s *xorm.Session, u *user.User, n int, now time.Time, fromCompletion bool) (completed []*Task, stopped bool, last *Task) {
	t.Helper()
	orig := timeNow
	t.Cleanup(func() { timeNow = orig })
	clock := now
	for range n {
		pinned := clock
		timeNow = func() time.Time { return pinned }
		require.NoError(t, (&Task{ID: 1, Done: true}).updateSingleTask(s, u, []string{"done"}))

		stored := storedTask1(t, s)
		if stored.Done {
			return completed, true, stored
		}
		reminders, err := getRemindersForTasks(s, []int64{1})
		require.NoError(t, err)
		stored.Reminders = reminders
		completed = append(completed, stored)
		if fromCompletion {
			clock = stored.DueDate
		}
	}
	return completed, false, nil
}

func TestTaskOccurrencesReasons(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}

	for _, tc := range []struct {
		name string
		task *Task
		want string
	}{
		{"no rule", &Task{ID: 1, ProjectID: 1, DueDate: occDue}, "no_rule"},
		{"no due date", &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=DAILY"}, "no_due_date"},
		{"count exhausted", &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=DAILY;COUNT=1", DueDate: occDue}, "count_exhausted"},
		{"until passed", &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=DAILY;UNTIL=20261020T000000Z", DueDate: occDue}, "until_passed"},
		// A stored rule that no longer parses, built directly: no write path accepts it.
		{"invalid rule", &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=FORTNIGHTLY", DueDate: occDue}, "invalid_rule"},
		// The check order: no_rule, then no_due_date, then task_done.
		{"done without a rule", &Task{ID: 1, ProjectID: 1, Done: true, DueDate: occDue}, "no_rule"},
		{"done without a due date", &Task{ID: 1, ProjectID: 1, Done: true, Repeats: "FREQ=DAILY"}, "no_due_date"},
		{"done", &Task{ID: 1, ProjectID: 1, Done: true, Repeats: "FREQ=DAILY", DueDate: occDue}, "task_done"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			answer, err := TaskOccurrences(s, u, tc.task, 5, occNow)
			require.NoError(t, err)
			assert.Equal(t, tc.want, answer.Reason)
			assert.NotNil(t, answer.Occurrences)
			assert.Empty(t, answer.Occurrences)
		})
	}

	// A saved task that is already done, read back from the database, paired with the same task
	// not done.
	t.Run("task done", func(t *testing.T) {
		for _, done := range []bool{true, false} {
			seedOccurrenceTask(t, s, occurrenceFixture{rule: "FREQ=DAILY", due: occDue, done: done})
			task := &Task{ID: 1}
			require.NoError(t, task.ReadOne(s, u))
			require.Equal(t, done, task.Done)

			answer, err := TaskOccurrences(s, u, task, 5, occNow)
			require.NoError(t, err)
			if done {
				assert.Equal(t, "task_done", answer.Reason)
				assert.NotNil(t, answer.Occurrences)
				assert.Empty(t, answer.Occurrences)
			} else {
				assert.Empty(t, answer.Reason)
				assert.Len(t, answer.Occurrences, 5, "positive control: the same task not done lists dates")
			}
		}
	})

	t.Run("a count outside 1 to the cap is invalid", func(t *testing.T) {
		task := &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=DAILY", DueDate: occDue}
		for _, n := range []int{0, -1, MaxOccurrencesCount + 1} {
			answer, err := TaskOccurrences(s, u, task, n, occNow)
			assert.True(t, IsErrInvalidData(err), "n=%d: got %v", n, err)
			assert.Nil(t, answer)

			previewed, err := PreviewOccurrences(s, u, PreviewInput{Rule: "FREQ=DAILY", DueDate: &occDue, Count: n}, occNow)
			assert.True(t, IsErrInvalidData(err), "preview n=%d: got %v", n, err)
			assert.Nil(t, previewed)
		}
		answer, err := TaskOccurrences(s, u, task, MaxOccurrencesCount, occNow)
		require.NoError(t, err)
		assert.Len(t, answer.Occurrences, MaxOccurrencesCount, "the cap itself is allowed")
	})
}

// TestTaskOccurrencesNoMutation: the query writes neither the task, its reminders nor the
// database, on a fixture where it does return dates.
func TestTaskOccurrencesNoMutation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}
	setUserZone(t, s, 1, "America/New_York")
	seedOccurrenceTask(t, s, occurrenceFixture{
		rule: "FREQ=DAILY;COUNT=10", due: time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC),
		start: time.Date(2027, 3, 14, 6, 0, 0, 0, time.UTC), end: time.Date(2027, 3, 14, 9, 0, 0, 0, time.UTC),
		reminder: time.Date(2027, 3, 14, 7, 0, 0, 0, time.UTC), relative: true,
	})

	task := &Task{ID: 1}
	require.NoError(t, task.ReadOne(s, u))
	require.Len(t, task.Reminders, 2)
	snapshot := *task
	snapshot.Reminders = make([]*TaskReminder, len(task.Reminders))
	for i, rem := range task.Reminders {
		r := *rem
		snapshot.Reminders[i] = &r
	}
	storedBefore := storedTask1(t, s)
	remindersBefore, err := getRemindersForTasks(s, []int64{1})
	require.NoError(t, err)

	answer, err := TaskOccurrences(s, u, task, 5, time.Date(2027, 3, 14, 8, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Len(t, answer.Occurrences, 5, "the same call lists dates")

	assert.Equal(t, &snapshot, task, "the task and its reminders are unchanged")
	assert.Equal(t, "FREQ=DAILY;COUNT=10", task.Repeats)
	assert.Equal(t, storedBefore, storedTask1(t, s), "the stored row is unchanged")
	remindersAfter, err := getRemindersForTasks(s, []int64{1})
	require.NoError(t, err)
	assert.Equal(t, remindersBefore, remindersAfter, "the stored reminders are unchanged")
}

// TestTaskOccurrencesPrecision: every date is whole seconds in the server zone, as stored, even
// from a fractional clock and a fractional in-memory due date.
func TestTaskOccurrencesPrecision(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}
	setUserZone(t, s, 1, "Asia/Kolkata")
	now := occNow.Add(123456789 * time.Nanosecond)

	for _, task := range []*Task{
		{ID: 1, ProjectID: 1, Repeats: "FREQ=HOURLY;INTERVAL=6", RepeatsFromCurrentDate: true, DueDate: occDue},
		{ID: 1, ProjectID: 1, Repeats: "FREQ=DAILY", DueDate: occDue.Add(500 * time.Millisecond)},
	} {
		answer, err := TaskOccurrences(s, u, task, 5, now)
		require.NoError(t, err)
		require.Len(t, answer.Occurrences, 5)
		for i, d := range answer.Occurrences {
			assert.Zero(t, d.Nanosecond(), "occurrence %d: %v", i, d)
			assert.Equal(t, config.GetTimeZone().String(), d.Location().String(), "occurrence %d", i)
		}
	}
}

// TestPreviewOccurrencesZone: the given zone, else the caller's, else the server's; a link share
// has no zone of its own and no project owner applies; an unusable name is ErrInvalidTimezone.
// The weekly rule's dates differ between New York, Kolkata and the server zone.
func TestPreviewOccurrencesZone(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}
	now := time.Date(2026, 10, 27, 0, 5, 0, 0, time.UTC)
	weeklyDue := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	in := PreviewInput{Rule: "FREQ=WEEKLY;BYDAY=MO,WE,FR", DueDate: &weeklyDue, Count: 3}
	preview := func(t *testing.T, a web.Auth, zone string) []time.Time {
		t.Helper()
		in := in
		in.TimeZone = zone
		answer, err := PreviewOccurrences(s, a, in, now)
		require.NoError(t, err)
		require.Len(t, answer.Occurrences, 3)
		return answer.Occurrences
	}

	setUserZone(t, s, 1, "America/New_York")
	callerNY := preview(t, u, "")
	setUserZone(t, s, 1, "Asia/Kolkata")
	callerKolkata := preview(t, u, "")
	explicitNY := preview(t, u, "America/New_York")
	setUserZone(t, s, 1, "")
	serverFallback := preview(t, u, "")
	explicitServer := preview(t, u, config.GetTimeZone().String())

	assert.Equal(t, callerNY, explicitNY, "the given zone wins over the caller's")
	assert.NotEqual(t, callerKolkata, explicitNY, "positive control: the zones discriminate")
	assert.Equal(t, explicitServer, serverFallback, "a caller without a zone gets the server's")
	assert.NotEqual(t, callerNY, serverFallback)

	t.Run("a link share falls through to the server zone", func(t *testing.T) {
		setUserZone(t, s, 3, "America/New_York") // the owner of link share 2's project
		assert.Equal(t, serverFallback, preview(t, &LinkSharing{ID: 2}, ""))
	})
	t.Run("an unusable zone is ErrInvalidTimezone", func(t *testing.T) {
		for _, name := range []string{"Mars/Olympus", "Local"} {
			in := in
			in.TimeZone = name
			answer, err := PreviewOccurrences(s, u, in, now)
			assert.True(t, IsErrInvalidTimezone(err), "%s: got %v", name, err)
			assert.NotContains(t, err.Error(), "<nil>", "%s: the message says why", name)
			assert.Nil(t, answer)
		}
	})
	t.Run("no rule and no due date", func(t *testing.T) {
		answer, err := PreviewOccurrences(s, u, PreviewInput{DueDate: in.DueDate, Count: 3}, now)
		require.NoError(t, err)
		assert.Equal(t, "no_rule", answer.Reason)
		answer, err = PreviewOccurrences(s, u, PreviewInput{Rule: in.Rule, Count: 3}, now)
		require.NoError(t, err)
		assert.Equal(t, "no_due_date", answer.Reason)
	})
}

// TestPreviewOccurrencesEqualNewTask: a preview equals the query on a newly created task with the
// same rule and due date, compared against the saved task, never a constant.
func TestPreviewOccurrencesEqualNewTask(t *testing.T) {
	for _, tc := range []struct {
		name           string
		repeat         *TaskRepeat
		due, now       time.Time
		fromCompletion bool
	}{
		// Crosses New York's 2027-03-14 gap at 02:30.
		{name: "daily 02:30 across spring forward", repeat: &TaskRepeat{Freq: "daily", Interval: 1}, due: time.Date(2027, 3, 13, 7, 30, 0, 0, time.UTC), now: time.Date(2027, 3, 13, 8, 0, 0, 0, time.UTC)},
		// Due on 2027-03-14 at 03:30 EDT, as a client normalising 02:30 would send it.
		{name: "daily due on the gap day", repeat: &TaskRepeat{Freq: "daily", Interval: 1}, due: time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC), now: time.Date(2027, 3, 14, 8, 0, 0, 0, time.UTC)},
		{name: "weekly on two days", repeat: &TaskRepeat{Freq: "weekly", Interval: 1, ByDay: []string{"mo", "th"}}, due: occDue, now: occNow},
		{name: "daily from completion", repeat: &TaskRepeat{Freq: "daily", Interval: 1}, due: occDue, now: occNow.Add(123 * time.Millisecond), fromCompletion: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			u := &user.User{ID: 1, Username: "user1"}
			setUserZone(t, s, 1, "America/New_York")

			created := &Task{Title: "preview twin", ProjectID: 1, Repeat: tc.repeat, DueDate: tc.due, RepeatsFromCurrentDate: tc.fromCompletion}
			require.NoError(t, created.Create(s, u))
			saved := &Task{ID: created.ID}
			require.NoError(t, saved.ReadOne(s, u))
			require.NotEmpty(t, saved.Repeats)

			want, err := TaskOccurrences(s, u, saved, 7, tc.now)
			require.NoError(t, err)
			got, err := PreviewOccurrences(s, u, PreviewInput{Rule: saved.Repeats, DueDate: &tc.due, FromCompletion: tc.fromCompletion, Count: 7}, tc.now)
			require.NoError(t, err)

			require.Len(t, want.Occurrences, 7, "positive control: the saved task lists dates")
			assert.Equal(t, want, &got.OccurrenceAnswer)
			assert.Equal(t, DueDateSourceExplicit, got.DueDateSource)
		})
	}
}

// TestTaskOccurrencesSpringForwardSkip pins the dates that occurrenceNewYorkCorpus only checks
// for equality with completion: an occurrence at a nonexistent local time is skipped and does not
// count toward COUNT (RFC 5545 §3.3.10), and a DAILY rule with BYEASTER lists Easter
// days only.
func TestTaskOccurrencesSpringForwardSkip(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	at := func(y int, m time.Month, d, h, mi int) time.Time { return time.Date(y, m, d, h, mi, 0, 0, ny).UTC() }
	// 02:30 on 2027-03-12 (EST); 2027-03-14 02:30 does not exist in New York.
	due := at(2027, 3, 12, 2, 30)
	now := due.Add(30 * time.Minute)

	for _, tc := range []struct {
		name       string
		rule       string
		due, now   time.Time
		count      int
		want       []time.Time
		wantReason string
	}{
		{
			name: "daily 02:30 skips the gap day", rule: "FREQ=DAILY", due: due, now: now, count: 4,
			want: []time.Time{at(2027, 3, 13, 2, 30), at(2027, 3, 15, 2, 30), at(2027, 3, 16, 2, 30), at(2027, 3, 17, 2, 30)},
		},
		{
			// COUNT=4 includes the due date: three more, and the skipped day is not one of them.
			name: "the skipped occurrence is not counted", rule: "FREQ=DAILY;COUNT=4", due: due, now: now, count: 10,
			want:       []time.Time{at(2027, 3, 13, 2, 30), at(2027, 3, 15, 2, 30), at(2027, 3, 16, 2, 30)},
			wantReason: string(reasonCountExhausted),
		},
		{
			// 364 days is 52 weeks, so every candidate is a Sunday at 02:30: 2027-03-14, 2028-03-12,
			// 2029-03-11, 2030-03-10 and 2031-03-09 are each that year's spring-forward Sunday and are
			// skipped by nextWholePeriod's loop; 2032-03-07 is a week before the 2032 gap.
			name: "a long period skips five gap days in a row", rule: "FREQ=DAILY;INTERVAL=364",
			due: at(2026, 3, 15, 2, 30), now: at(2026, 3, 15, 3, 0), count: 1,
			want: []time.Time{at(2032, 3, 7, 2, 30)},
		},
		{
			// Every instance is in the spring-forward gap (second Sunday of March, 02:30), so each
			// candidate is skipped and the iteration still ends, with no date.
			name: "an all-gap rule lists nothing", rule: "FREQ=YEARLY;BYMONTH=3;BYDAY=2SU;BYHOUR=2;BYMINUTE=30",
			due: at(2026, 3, 1, 2, 30), now: at(2026, 3, 1, 3, 0), count: 3,
			want: []time.Time{}, wantReason: string(reasonNoOccurrence),
		},
		{
			// Easter: 2027-03-28, 2028-04-16, 2029-04-01, 2030-04-21.
			name: "DAILY BYEASTER lists Easter days only", rule: "FREQ=DAILY;BYEASTER=0",
			due: at(2027, 3, 28, 9, 0), now: at(2027, 3, 28, 10, 0), count: 3,
			want: []time.Time{at(2028, 4, 16, 9, 0), at(2029, 4, 1, 9, 0), at(2030, 4, 21, 9, 0)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			u := &user.User{ID: 1, Username: "user1"}

			start := time.Now()
			answer, err := PreviewOccurrences(s, u, PreviewInput{Rule: tc.rule, DueDate: &tc.due, TimeZone: "America/New_York", Count: tc.count}, tc.now)
			require.NoError(t, err)
			// The same ceiling as TestTaskOccurrencesTiming: skipped candidates must not make the
			// search unbounded.
			assert.Less(t, time.Since(start), 2*time.Second)
			require.Len(t, answer.Occurrences, len(tc.want), "got %v", answer.Occurrences)
			for i, w := range tc.want {
				assert.True(t, answer.Occurrences[i].Equal(w), "date %d: got %v, want %v", i, answer.Occurrences[i], w)
			}
			assert.Equal(t, tc.wantReason, answer.Reason)
		})
	}
}

// TestTaskOccurrencesTiming bounds a query for MaxOccurrencesCount dates on the worst rules of
// TestNextOccurrenceTiming, at the same anchors and against the same ceilings (4 s for dense monthly,
// 2 s otherwise). After the first
// step the anchor is recent, so only that step pays the horizon's cost: the query costs about one
// completion plus 99 cheap steps. Measured 2026-10-25 on a developer machine: the worst case, dense
// monthly at the horizon, took 0.42-0.43 s for 100 dates against 0.39-0.40 s for one completion
// (TestNextOccurrenceTiming, same run); every other row stayed under 30 ms. Re-measured 2026-10-28
// after the spring-forward skip: dense monthly 0.42 s; the all-gap YEARLY rule ends in
// under 1 ms with no_occurrence (now pinned in TestTaskOccurrencesSpringForwardSkip); every other
// row stayed under 30 ms. Dense monthly has its own ceiling since rrule-go v1.9.0, whose civil-time
// resolution costs about 1.8x per candidate: on an i5, 0.90-1.00 s on v1.8.3
// against 1.74-1.89 s on v1.9.0 unloaded, and 2.88-3.37 s under load. The cap stays at 100. The
// measured durations are logged.
func TestTaskOccurrencesTiming(t *testing.T) {
	const ceiling = 2 * time.Second
	const denseCeiling = 4 * time.Second
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	horizon := now.AddDate(-200, 0, 1)
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}

	timed := func(t *testing.T, zone, rule string, due time.Time) (*OccurrenceAnswer, time.Duration) {
		t.Helper()
		setUserZone(t, s, 1, zone)
		start := time.Now()
		answer, err := TaskOccurrences(s, u, &Task{ID: 1, ProjectID: 1, Repeats: rule, DueDate: due}, MaxOccurrencesCount, now)
		elapsed := time.Since(start)
		require.NoError(t, err)
		t.Logf("timing %s %s: %v (%d dates, reason %q)", zone, t.Name(), elapsed, len(answer.Occurrences), answer.Reason)
		return answer, elapsed
	}

	for _, zone := range []string{"America/New_York", "Asia/Kolkata"} {
		t.Run("dense monthly in "+zone, func(t *testing.T) {
			answer, elapsed := timed(t, zone, worstInBudgetMonthly, horizon)
			assert.Less(t, elapsed, denseCeiling)
			assert.Len(t, answer.Occurrences, MaxOccurrencesCount)
		})
		t.Run("never-matching in "+zone, func(t *testing.T) {
			var worst time.Duration
			for _, rule := range neverMatchingInBoundRules() {
				answer, elapsed := timed(t, zone, rule, time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC))
				assert.Empty(t, answer.Occurrences, "never matches")
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
		// The gap-shifted 02:30 used to collide with 03:30 and count once; it is now
		// skipped (RFC 5545 §3.3.10), so this row times the skip inside nextCounted, not a collision.
		{"COUNT across skipped gap times", "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=30;COUNT=10000"},
		{"plain DAILY COUNT", "FREQ=DAILY;COUNT=10000"},
		{"never-matching DAILY BYSETPOS with COUNT", neverMatchingInBoundRules()[0] + ";COUNT=10000"},
		{"chronological COUNT", "FREQ=DAILY;BYHOUR=2,3;BYMINUTE=0,30;BYSETPOS=2,3;COUNT=10000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, elapsed := timed(t, "America/New_York", tc.rule, horizon)
			assert.Less(t, elapsed, ceiling)
		})
		// At the horizon a COUNT=10000 rule is exhausted at the first step; from a recent due date
		// it takes all MaxOccurrencesCount steps, each decrementing COUNT.
		t.Run(tc.name+" from a recent due date", func(t *testing.T) {
			answer, elapsed := timed(t, "America/New_York", tc.rule, now.AddDate(0, 0, -1))
			assert.Less(t, elapsed, ceiling)
			if !strings.Contains(tc.name, "never-matching") {
				assert.Len(t, answer.Occurrences, MaxOccurrencesCount)
			}
		})
	}
}

// forecastRow is one on-time-completion forecast row, evaluated on
// both list surfaces: TaskOccurrences (the acting user's zone) and PreviewOccurrences (an explicit
// zone).
type forecastRow struct {
	name       string
	zone       string
	rule       string
	due, now   time.Time
	count      int
	want       []time.Time
	wantReason string
}

// TestOccurrencesOnTimeForecast pins the forecast rows F-8 to F-14 below (F-7, a generated due
// date, is in TestOccurrencePreviewInitialDue). Europe/Berlin and HOURLY;INTERVAL=1 from completion unless
// stated; 2026-10-28 is CET (UTC+1).
func TestOccurrencesOnTimeForecast(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	b := func(h, m, sec int) time.Time { return time.Date(2026, 10, 28, h, m, sec, 0, berlin) }
	hours := func(from time.Time, n int) []time.Time {
		out := make([]time.Time, n)
		for i := range out {
			out[i] = from.Add(time.Duration(i) * time.Hour)
		}
		return out
	}
	const hourly = "FREQ=HOURLY;INTERVAL=1"

	rows := []forecastRow{
		{name: "F-8 future explicit due", zone: "Europe/Berlin", rule: hourly, due: b(18, 0, 0), now: b(10, 0, 0), count: 3, want: hours(b(19, 0, 0), 3)},
		// Overdue: completed at now.
		{name: "F-9 overdue due", zone: "Europe/Berlin", rule: hourly, due: b(8, 0, 0), now: b(10, 0, 0), count: 3, want: hours(b(11, 0, 0), 3)},
		{name: "F-10 due equal to now", zone: "Europe/Berlin", rule: hourly, due: b(10, 0, 0), now: b(10, 0, 0), count: 3, want: hours(b(11, 0, 0), 3)},
		{name: "F-11 COUNT=1", zone: "Europe/Berlin", rule: hourly + ";COUNT=1", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: []time.Time{}, wantReason: string(reasonCountExhausted)},
		// The display count never replenishes COUNT.
		{name: "F-12 COUNT=3 asked for 10", zone: "Europe/Berlin", rule: hourly + ";COUNT=3", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: hours(b(19, 0, 0), 2), wantReason: string(reasonCountExhausted)},
		// UNTIL is inclusive. The reason is observed, not assumed: run
		// 2026-10-28, the shared engine ends both lists, on both
		// surfaces, with no_occurrence: the same reason the query gives
		// a rule that ends by UNTIL within the list ("UNTIL within N" above).
		{name: "F-13 UNTIL on an instance", zone: "Europe/Berlin", rule: hourly + ";UNTIL=20261028T190000Z", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: hours(b(19, 0, 0), 2), wantReason: string(reasonNoOccurrence)},
		{name: "F-13 UNTIL one second before an instance", zone: "Europe/Berlin", rule: hourly + ";UNTIL=20261028T185959Z", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: hours(b(19, 0, 0), 1), wantReason: string(reasonNoOccurrence)},
		// Spring forward: 03:30 EDT is one elapsed hour after 01:30 EST; no 02:xx.
		{
			name: "F-14 across spring forward", zone: "America/New_York", rule: hourly,
			due: time.Date(2026, 3, 8, 1, 30, 0, 0, ny), now: time.Date(2026, 3, 7, 23, 0, 0, 0, ny), count: 3,
			want: []time.Time{time.Date(2026, 3, 8, 3, 30, 0, 0, ny), time.Date(2026, 3, 8, 4, 30, 0, 0, ny), time.Date(2026, 3, 8, 5, 30, 0, 0, ny)},
		},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			assertForecastRow(t, row)
		})
	}

	t.Run("F-14 has no 02:xx and one elapsed hour", func(t *testing.T) {
		row := rows[len(rows)-1]
		assert.Equal(t, time.Hour, row.want[0].Sub(row.due))
		for _, d := range row.want {
			assert.NotEqual(t, 2, d.In(ny).Hour())
		}
	})

	t.Run("not from completion is unchanged by the forecast", func(t *testing.T) {
		// Positive control: the same future due date without repeat-from-completion steps from the
		// due date by the rule.
		task := &Task{Repeats: hourly, DueDate: b(18, 0, 0)}
		assert.True(t, forecastFirstCompletion(task, b(10, 0, 0)).Equal(b(10, 0, 0)))
		task.RepeatsFromCurrentDate = true
		assert.True(t, forecastFirstCompletion(task, b(10, 0, 0)).Equal(b(18, 0, 0)))
		assert.True(t, forecastFirstCompletion(task, b(19, 0, 0).Add(500*time.Millisecond)).Equal(b(19, 0, 0)), "now is truncated to the second")
	})
}

func assertForecastRow(t *testing.T, row forecastRow) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1, Username: "user1"}

	preview, err := PreviewOccurrences(s, u, PreviewInput{Rule: row.rule, DueDate: &row.due, FromCompletion: true, TimeZone: row.zone, Count: row.count}, row.now)
	require.NoError(t, err)
	setUserZone(t, s, 1, row.zone)
	query, err := TaskOccurrences(s, u, &Task{ID: 1, ProjectID: 1, Repeats: row.rule, RepeatsFromCurrentDate: true, DueDate: row.due}, row.count, row.now)
	require.NoError(t, err)

	for surface, answer := range map[string]*OccurrenceAnswer{"preview": &preview.OccurrenceAnswer, "query": query} {
		require.Len(t, answer.Occurrences, len(row.want), "%s: got %v", surface, answer.Occurrences)
		prev := row.due
		for i, w := range row.want {
			got := answer.Occurrences[i]
			assert.True(t, got.Equal(w), "%s date %d: got %v, want %v", surface, i, got, w)
			assert.True(t, got.After(prev), "%s date %d: %v is not after %v", surface, i, got, prev)
			prev = got
		}
		assert.Equal(t, row.wantReason, answer.Reason, surface)
	}
}

// TestOccurrencePreviewInitialDue pins the preview's generated anchor (rows P-1 to P-7c and F-7
// below) and its explicit one (the P-10 subtest): the due date is
// selectInitialDue's at the operation's now, the occurrences follow it strictly, and a reason may
// sit beside a valid due date. Europe/Berlin is CET (UTC+1) on 2026-10-28 and 2026-10-29.
func TestOccurrencePreviewInitialDue(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	b := func(d, h, m, sec int) time.Time { return time.Date(2026, 10, d, h, m, sec, 0, berlin) }
	n := func(d, h, m int) time.Time { return time.Date(2023, 3, d, h, m, 0, 0, ny) }

	for _, tc := range []struct {
		name       string
		rule       string
		hhmm       string
		zone       string
		now        time.Time
		fromCompl  bool
		count      int
		wantDue    time.Time // zero: no due date
		want       []time.Time
		wantReason string
	}{
		{name: "P-1 DAILY 09:00 after 09:00", rule: "FREQ=DAILY", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), count: 2, wantDue: b(29, 9, 0, 0), want: []time.Time{b(30, 9, 0, 0), b(31, 9, 0, 0)}},
		// The cutoff is strictly after now.
		{name: "P-2 now exactly 09:00:00", rule: "FREQ=DAILY", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 9, 0, 0), count: 1, wantDue: b(29, 9, 0, 0), want: []time.Time{b(30, 9, 0, 0)}},
		// 2023-03-12 02:30 does not exist in New York: that candidate is skipped.
		{name: "P-3 DAILY 02:30 on the gap day", rule: "FREQ=DAILY", hhmm: "02:30", zone: "America/New_York", now: n(12, 1, 0), count: 1, wantDue: n(13, 2, 30), want: []time.Time{n(14, 2, 30)}},
		{name: "P-4 MINUTELY;INTERVAL=15 from 02:30 on the gap day", rule: "FREQ=MINUTELY;INTERVAL=15", hhmm: "02:30", zone: "America/New_York", now: n(12, 1, 0), count: 2, wantDue: n(12, 3, 0), want: []time.Time{n(12, 3, 15), n(12, 3, 30)}},
		// COUNT is ignored by the search and counts the selected date as instance 1.
		{name: "P-5 COUNT=1 with the seed before now", rule: "FREQ=DAILY;COUNT=1", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), count: 10, wantDue: b(29, 9, 0, 0), want: []time.Time{}, wantReason: string(reasonCountExhausted)},
		{name: "P-6 COUNT=3 asked for 5", rule: "FREQ=DAILY;COUNT=3", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), count: 5, wantDue: b(29, 9, 0, 0), want: []time.Time{b(30, 9, 0, 0), b(31, 9, 0, 0)}, wantReason: string(reasonCountExhausted)},
		// UNTIL is inclusive: 2026-10-29 09:00 CET is 08:00Z.
		{name: "P-7a UNTIL on the first candidate", rule: "FREQ=DAILY;UNTIL=20261029T080000Z", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), count: 3, wantDue: b(29, 9, 0, 0), want: []time.Time{}, wantReason: string(reasonNoOccurrence)},
		{name: "P-7b UNTIL one second before the first candidate", rule: "FREQ=DAILY;UNTIL=20261029T075959Z", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), count: 3, want: []time.Time{}, wantReason: string(reasonNoOccurrence)},
		{name: "P-7c UNTIL before now", rule: "FREQ=DAILY;UNTIL=20261027T000000Z", hhmm: "09:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), count: 3, want: []time.Time{}, wantReason: string(reasonUntilPassed)},
		// A generated future due date with an on-time forecast.
		{name: "F-7 generated 18:00 HOURLY from completion", rule: "FREQ=HOURLY;INTERVAL=1", hhmm: "18:00", zone: "Europe/Berlin", now: b(28, 10, 0, 0), fromCompl: true, count: 3, wantDue: b(28, 18, 0, 0), want: []time.Time{b(28, 19, 0, 0), b(28, 20, 0, 0), b(28, 21, 0, 0)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			u := &user.User{ID: 1, Username: "user1"}

			answer, err := PreviewOccurrences(s, u, PreviewInput{Rule: tc.rule, InitialDueTime: tc.hhmm, FromCompletion: tc.fromCompl, TimeZone: tc.zone, Count: tc.count}, tc.now)
			require.NoError(t, err)
			assert.Equal(t, tc.zone, answer.TimeZone)
			assert.True(t, answer.EvaluatedAt.Equal(tc.now))
			assert.Equal(t, time.UTC, answer.EvaluatedAt.Location())
			assert.Equal(t, tc.wantReason, answer.Reason)
			require.NotNil(t, answer.Occurrences)
			if tc.wantDue.IsZero() {
				assert.Nil(t, answer.DueDate)
				assert.Empty(t, answer.DueDateSource)
				assert.Empty(t, answer.Occurrences)
				return
			}
			require.NotNil(t, answer.DueDate)
			assert.True(t, answer.DueDate.Equal(tc.wantDue), "due: got %v, want %v", answer.DueDate, tc.wantDue)
			assert.True(t, answer.DueDate.After(tc.now), "a generated due date is strictly after now")
			assert.Equal(t, DueDateSourceGenerated, answer.DueDateSource)

			// The same date a write selects (preview == save).
			hour, minute, err := ParseInitialDueTime(tc.hhmm)
			require.NoError(t, err)
			zone, err := time.LoadLocation(tc.zone)
			require.NoError(t, err)
			selected, reason := selectInitialDue(tc.rule, hour, minute, zone, tc.now)
			require.Empty(t, reason)
			assert.True(t, selected.Equal(*answer.DueDate))

			require.Len(t, answer.Occurrences, len(tc.want), "got %v", answer.Occurrences)
			prev := *answer.DueDate
			for i, w := range tc.want {
				assert.True(t, answer.Occurrences[i].Equal(w), "date %d: got %v, want %v", i, answer.Occurrences[i], w)
				assert.True(t, answer.Occurrences[i].After(prev), "date %d is strictly after the previous one", i)
				prev = answer.Occurrences[i]
			}
		})
	}

	t.Run("P-10 explicit due date: the query's list, source explicit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 1, Username: "user1"}
		setUserZone(t, s, 1, "Europe/Berlin")
		due := b(29, 9, 0, 0).Add(400 * time.Millisecond)
		answer, err := PreviewOccurrences(s, u, PreviewInput{Rule: "FREQ=WEEKLY;BYDAY=MO,TH", DueDate: &due, Count: 4}, b(28, 10, 0, 0))
		require.NoError(t, err)
		want, err := TaskOccurrences(s, u, &Task{ID: 1, ProjectID: 1, Repeats: "FREQ=WEEKLY;BYDAY=MO,TH", DueDate: due.Truncate(time.Second)}, 4, b(28, 10, 0, 0))
		require.NoError(t, err)
		require.Len(t, want.Occurrences, 4, "positive control")
		assert.Equal(t, want, &answer.OccurrenceAnswer)
		assert.Equal(t, DueDateSourceExplicit, answer.DueDateSource)
		require.NotNil(t, answer.DueDate)
		assert.True(t, answer.DueDate.Equal(due.Truncate(time.Second)), "echoed to the second")
		assert.Equal(t, "Europe/Berlin", answer.TimeZone, "the caller's zone")
	})

	t.Run("no due date at all keeps the query's no_due_date", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		answer, err := PreviewOccurrences(s, &user.User{ID: 1}, PreviewInput{Rule: "FREQ=DAILY", Count: 3}, b(28, 10, 0, 0))
		require.NoError(t, err)
		assert.Equal(t, reasonNoDueDate, answer.Reason)
		assert.Nil(t, answer.DueDate)
		assert.Empty(t, answer.DueDateSource)
	})

	t.Run("both anchors are a conflict, as the preview maps it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		due := b(29, 9, 0, 0)
		answer, err := PreviewOccurrences(s, &user.User{ID: 1}, PreviewInput{Rule: "FREQ=DAILY", DueDate: &due, InitialDueTime: "09:00", Count: 3}, b(28, 10, 0, 0))
		assert.Nil(t, answer)
		conflict, ok := InitialDueConflict(true, "09:00", true, false).(ErrInitialDueConflict)
		require.True(t, ok)
		assert.Equal(t, conflict.PreviewError(), err)
		assert.True(t, IsErrInvalidData(conflict.WriteError()), "positive control: the same conflict maps differently on a write")
	})

	t.Run("zone: timezone, then the project's, then the caller's", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		// User 2, without a zone, previews; project 1's owner, user 1, is in Asia/Kolkata.
		setUserZone(t, s, 1, "Asia/Kolkata")
		setUserZone(t, s, 2, "")
		u := &user.User{ID: 2, Username: "user2"}
		in := PreviewInput{Rule: "FREQ=DAILY", InitialDueTime: "09:00", Count: 1}
		zoneOf := func(in PreviewInput) string {
			answer, err := PreviewOccurrences(s, u, in, b(28, 10, 0, 0))
			require.NoError(t, err)
			return answer.TimeZone
		}
		assert.Equal(t, config.GetTimeZone().String(), zoneOf(in), "no context: the server's")
		in.ProjectID = 1
		assert.Equal(t, "Asia/Kolkata", zoneOf(in), "the project's owner, as resolveTaskZone resolves it")
		in.TimeZone = "Europe/Berlin"
		assert.Equal(t, "Europe/Berlin", zoneOf(in), "timezone wins")
	})
}
