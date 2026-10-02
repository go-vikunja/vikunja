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
	"context"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func day(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }

func datedTask(t *testing.T, owner *user.User, projectID int64, title string, start, end time.Time) *Task {
	t.Helper()
	return mustCreate(t, owner, &Task{Title: title, ProjectID: projectID, StartDate: start, EndDate: end})
}

func precede(t *testing.T, owner *user.User, a, b *Task) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, (&TaskRelation{TaskID: a.ID, OtherTaskID: b.ID, RelationKind: RelationKindPreceeds}).Create(s, owner))
	require.NoError(t, s.Commit())
}

func reschedule(t *testing.T, as *user.User, id int64, start, end time.Time) (*RescheduleResult, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	res, err := RescheduleTask(s, as, id, start, end)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return res, nil
}

func datesOf(t *testing.T, id int64) (time.Time, time.Time) {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	task, err := GetTaskByIDSimple(s, id)
	require.NoError(t, err)
	return task.StartDate, task.EndDate
}

func TestRescheduleTask(t *testing.T) {
	t.Run("a successor that would start too early moves and keeps its duration", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(5))
		b := datedTask(t, jane, jane.DefaultProjectID, "B", day(5), day(8))
		precede(t, jane, a, b)

		res, err := reschedule(t, jane, a.ID, day(1), day(10))

		require.NoError(t, err)
		require.Len(t, res.Changed, 1)
		start, end := datesOf(t, b.ID)
		assert.True(t, start.Equal(day(10)), "starts when A ends: %v", start)
		assert.True(t, end.Equal(day(13)), "keeps its 3 days: %v", end)
		_, aEnd := datesOf(t, a.ID)
		assert.True(t, aEnd.Equal(day(10)))
	})
	t.Run("the push cascades down a chain", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		b := datedTask(t, jane, jane.DefaultProjectID, "B", day(3), day(5))
		c := datedTask(t, jane, jane.DefaultProjectID, "C", day(5), day(6))
		precede(t, jane, a, b)
		precede(t, jane, b, c)

		res, err := reschedule(t, jane, a.ID, day(1), day(7))

		require.NoError(t, err)
		assert.Len(t, res.Changed, 2)
		bs, be := datesOf(t, b.ID)
		cs, ce := datesOf(t, c.ID)
		assert.True(t, bs.Equal(day(7)) && be.Equal(day(9)), "%v %v", bs, be)
		assert.True(t, cs.Equal(day(9)) && ce.Equal(day(10)), "%v %v", cs, ce)
	})
	t.Run("a successor that already starts late enough is left alone, and moving earlier pulls nobody", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		b := datedTask(t, jane, jane.DefaultProjectID, "B", day(10), day(12))
		precede(t, jane, a, b)

		later, err := reschedule(t, jane, a.ID, day(1), day(8))
		require.NoError(t, err)
		assert.Empty(t, later.Changed, "day 10 is after day 8")

		earlier, err := reschedule(t, jane, a.ID, day(1), day(2))
		require.NoError(t, err)
		assert.Empty(t, earlier.Changed)
		bs, _ := datesOf(t, b.ID)
		assert.True(t, bs.Equal(day(10)), "never pulled earlier")
	})
	t.Run("the latest predecessor wins when a task has two", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		x := datedTask(t, jane, jane.DefaultProjectID, "X", day(1), day(9))
		b := datedTask(t, jane, jane.DefaultProjectID, "B", day(3), day(5))
		precede(t, jane, a, b)
		precede(t, jane, x, b)

		_, err := reschedule(t, jane, a.ID, day(1), day(4))
		require.NoError(t, err)
		bs, _ := datesOf(t, b.ID)
		assert.True(t, bs.Equal(day(4)), "pushed by A to day 4 for now")

		// Now X, which ends later, is moved: it pushes B further.
		_, err = reschedule(t, jane, x.ID, day(1), day(12))
		require.NoError(t, err)
		bs, _ = datesOf(t, b.ID)
		assert.True(t, bs.Equal(day(12)))
	})
	t.Run("undated and done successors stay", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		undated := mustCreate(t, jane, &Task{Title: "Undated", ProjectID: jane.DefaultProjectID})
		done := mustCreate(t, jane, &Task{Title: "Done", ProjectID: jane.DefaultProjectID, StartDate: day(3), EndDate: day(4), Done: true})
		precede(t, jane, a, undated)
		precede(t, jane, a, done)

		res, err := reschedule(t, jane, a.ID, day(1), day(9))

		require.NoError(t, err)
		assert.Empty(t, res.Changed)
		s, e := datesOf(t, done.ID)
		assert.True(t, s.Equal(day(3)) && e.Equal(day(4)))
		us, ue := datesOf(t, undated.ID)
		assert.True(t, us.IsZero() && ue.IsZero())
	})
	t.Run("a milestone moves to the end date and keeps zero duration", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		m := mustCreate(t, jane, &Task{Title: "Release", ProjectID: jane.DefaultProjectID, IsMilestone: true, StartDate: day(3), EndDate: day(3)})
		precede(t, jane, a, m)

		res, err := reschedule(t, jane, a.ID, day(1), day(6))

		require.NoError(t, err)
		require.Len(t, res.Changed, 1)
		s, e := datesOf(t, m.ID)
		assert.True(t, s.Equal(day(6)) && e.Equal(day(6)), "%v %v", s, e)

		// Moving the milestone itself sets start to end, and it needs an end.
		_, err = reschedule(t, jane, m.ID, day(1), day(9))
		require.NoError(t, err)
		s, e = datesOf(t, m.ID)
		assert.True(t, s.Equal(day(9)) && e.Equal(day(9)))
		_, err = reschedule(t, jane, m.ID, time.Time{}, time.Time{})
		require.Error(t, err)
	})
	t.Run("a successor in a project the caller cannot write is skipped and reported", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		mine := createProjectFor(t, jane, "Mine")
		joes := createProjectFor(t, joe, "Joes")
		shareProject(t, joes.ID, jane.ID, PermissionRead) // read only
		shareProject(t, mine.ID, joe.ID, PermissionRead)
		a := datedTask(t, jane, mine.ID, "A", day(1), day(3))
		b := datedTask(t, joe, joes.ID, "B", day(3), day(5))
		// Joe may relate to A, he can read it.
		precede(t, joe, a, b)

		res, err := reschedule(t, jane, a.ID, day(1), day(9))

		require.NoError(t, err)
		assert.Empty(t, res.Changed)
		assert.Equal(t, []int64{b.ID}, res.Skipped)
		s, _ := datesOf(t, b.ID)
		assert.True(t, s.Equal(day(3)))
	})
	t.Run("a successor in an archived project is skipped, the rest of the move still happens", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		mine := createProjectFor(t, jane, "Mine")
		archived := createProjectFor(t, jane, "Old")
		a := datedTask(t, jane, mine.ID, "A", day(1), day(3))
		inArchived := datedTask(t, jane, archived.ID, "In archived", day(3), day(5))
		fine := datedTask(t, jane, mine.ID, "Fine", day(3), day(4))
		precede(t, jane, a, inArchived)
		precede(t, jane, a, fine)
		s := db.NewSession()
		_, err := s.ID(archived.ID).Cols("is_archived").Update(&Project{IsArchived: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		res, err := reschedule(t, jane, a.ID, day(1), day(9))

		require.NoError(t, err)
		assert.Equal(t, []int64{inArchived.ID}, res.Skipped)
		require.Len(t, res.Changed, 1)
		start, _ := datesOf(t, fine.ID)
		assert.True(t, start.Equal(day(9)))
		archivedStart, _ := datesOf(t, inArchived.ID)
		assert.True(t, archivedStart.Equal(day(3)), "left where it was")
	})
	t.Run("the caller needs write access to the task itself", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))

		_, err := reschedule(t, joe, a.ID, day(1), day(9))

		require.Error(t, err)
		_, end := datesOf(t, a.ID)
		assert.True(t, end.Equal(day(3)), "nothing changed")
	})
	t.Run("an end before the start is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))

		_, err := reschedule(t, jane, a.ID, day(5), day(2))

		require.Error(t, err)
	})
	t.Run("assignees are untouched", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		before := assigneeIDs(t, a.ID)
		require.NotEmpty(t, before, "the inbox task is auto-assigned")

		_, err := reschedule(t, jane, a.ID, day(2), day(4))

		require.NoError(t, err)
		assert.Equal(t, before, assigneeIDs(t, a.ID))
	})
	t.Run("queues an update event per moved task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		b := datedTask(t, jane, jane.DefaultProjectID, "B", day(3), day(4))
		precede(t, jane, a, b)
		events.ClearDispatchedEvents()

		_, err := reschedule(t, jane, a.ID, day(1), day(6))

		require.NoError(t, err)
		assert.Equal(t, 2, events.CountDispatchedEvents((&TaskUpdatedEvent{}).Name()))
	})
}

func TestSuccessorDates(t *testing.T) {
	task := func(start, end time.Time, milestone bool) *Task {
		return &Task{StartDate: start, EndDate: end, IsMilestone: milestone}
	}
	s, e, shift := successorDates(task(day(3), day(5), false), day(10))
	assert.True(t, shift)
	assert.True(t, s.Equal(day(10)) && e.Equal(day(12)))

	_, _, shift = successorDates(task(day(10), day(12), false), day(10))
	assert.False(t, shift, "starts exactly when the predecessor ends")

	s, e, shift = successorDates(task(day(3), time.Time{}, false), day(10))
	assert.True(t, shift)
	assert.True(t, s.Equal(day(10)) && e.IsZero(), "no end stays no end")

	_, _, shift = successorDates(task(time.Time{}, day(5), false), day(10))
	assert.False(t, shift, "no start")

	_, _, shift = successorDates(task(time.Time{}, time.Time{}, true), day(10))
	assert.False(t, shift, "a milestone without an end")
}

func TestPrecedesCycleIsRefused(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	jane := registerPerson(t, "jane", "jane@corp.com")
	a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(2))
	b := datedTask(t, jane, jane.DefaultProjectID, "B", day(2), day(3))
	c := datedTask(t, jane, jane.DefaultProjectID, "C", day(3), day(4))
	precede(t, jane, a, b)
	precede(t, jane, b, c)

	s := db.NewSession()
	defer s.Close()
	err := (&TaskRelation{TaskID: c.ID, OtherTaskID: a.ID, RelationKind: RelationKindPreceeds}).Create(s, jane)
	require.Error(t, err)
	assert.IsType(t, ErrTaskRelationCycle{}, err)
	_ = s.Rollback()

	s2 := db.NewSession()
	defer s2.Close()
	err = (&TaskRelation{TaskID: a.ID, OtherTaskID: c.ID, RelationKind: RelationKindFollows}).Create(s2, jane)
	require.Error(t, err, "A follows C while C follows B follows A is a cycle too")
	_ = s2.Rollback()
}

func TestProjectBaseline(t *testing.T) {
	save := func(t *testing.T, as *user.User, projectID int64) (int, error) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		n, err := SaveProjectBaseline(s, as, projectID)
		if err != nil {
			_ = s.Rollback()
			return 0, err
		}
		require.NoError(t, s.Commit())
		return n, nil
	}
	read := func(t *testing.T, as *user.User, projectID int64) ([]*TaskBaseline, time.Time, error) {
		t.Helper()
		s := db.NewReadSession()
		defer s.Close()
		return GetProjectBaseline(s, as, projectID)
	}

	t.Run("saves the dated tasks and replaces the previous baseline", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		mustCreate(t, jane, &Task{Title: "Undated", ProjectID: jane.DefaultProjectID})

		n, err := save(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		_, err = reschedule(t, jane, a.ID, day(2), day(9))
		require.NoError(t, err)
		rows, savedAt, err := read(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.True(t, rows[0].EndDate.Equal(day(3)), "the baseline keeps the old plan")
		assert.False(t, savedAt.IsZero())

		b := datedTask(t, jane, jane.DefaultProjectID, "B", day(4), day(5))
		n, err = save(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		assert.Equal(t, 2, n, "replaced, now with both")
		rows, _, err = read(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		ids := []int64{rows[0].TaskID, rows[1].TaskID}
		assert.ElementsMatch(t, []int64{a.ID, b.ID}, ids)
	})
	t.Run("a project without a baseline reads as empty", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		rows, savedAt, err := read(t, jane, jane.DefaultProjectID)

		require.NoError(t, err)
		assert.Empty(t, rows)
		assert.True(t, savedAt.IsZero())
	})
	t.Run("clear removes it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		_, err := save(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)

		s := db.NewSession()
		require.NoError(t, ClearProjectBaseline(s, jane, jane.DefaultProjectID))
		require.NoError(t, s.Commit())
		s.Close()

		rows, _, err := read(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
	t.Run("permissions: read is enough to see it, write is needed to change it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		reader := registerPerson(t, "reader", "reader@corp.com")
		stranger := registerPerson(t, "stranger", "stranger@corp.com")
		shareProject(t, jane.DefaultProjectID, reader.ID, PermissionRead)
		datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		_, err := save(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)

		rows, _, err := read(t, reader, jane.DefaultProjectID)
		require.NoError(t, err)
		assert.Len(t, rows, 1)

		_, err = save(t, reader, jane.DefaultProjectID)
		require.Error(t, err, "a reader cannot save")
		s := db.NewSession()
		err = ClearProjectBaseline(s, reader, jane.DefaultProjectID)
		_ = s.Rollback()
		s.Close()
		require.Error(t, err, "a reader cannot clear")

		_, _, err = read(t, stranger, jane.DefaultProjectID)
		require.Error(t, err, "a stranger cannot read")
		_, err = save(t, stranger, jane.DefaultProjectID)
		require.Error(t, err)
	})
	t.Run("a task that moved to another project does not break the new project's baseline", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		other := createProjectFor(t, jane, "Other")
		moved := datedTask(t, jane, jane.DefaultProjectID, "Moved", day(1), day(3))
		_, err := save(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)

		s := db.NewSession()
		_, err = s.ID(moved.ID).Cols("project_id").Update(&Task{ProjectID: other.ID})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		// The old project no longer lists it, and the new one can take its own baseline.
		rows, _, err := read(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		assert.Empty(t, rows)
		n, err := save(t, jane, other.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		rows, _, err = read(t, jane, other.ID)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, moved.ID, rows[0].TaskID)
	})
	t.Run("a deleted task is not returned", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		gone := datedTask(t, jane, jane.DefaultProjectID, "Gone", day(1), day(3))
		_, err := save(t, jane, jane.DefaultProjectID)
		require.NoError(t, err)
		s := db.NewSession()
		_, err = s.ID(gone.ID).Cols("deleted_at").Update(&Task{DeletedAt: time.Now()})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		rows, _, err := read(t, jane, jane.DefaultProjectID)

		require.NoError(t, err)
		assert.Empty(t, rows)
	})
	t.Run("pseudo projects have no baseline", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		_, err := save(t, jane, -1)
		require.Error(t, err)
		_, err = save(t, jane, 0)
		require.Error(t, err)
	})
	t.Run("deleted tasks are not part of it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a := datedTask(t, jane, jane.DefaultProjectID, "A", day(1), day(3))
		s := db.NewSession()
		_, err := s.ID(a.ID).Cols("deleted_at").Update(&Task{DeletedAt: time.Now()})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		n, err := save(t, jane, jane.DefaultProjectID)

		require.NoError(t, err)
		assert.Zero(t, n)
	})
}

func TestIsMilestoneRoundTrip(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	jane := registerPerson(t, "jane", "jane@corp.com")
	m := mustCreate(t, jane, &Task{Title: "M", ProjectID: jane.DefaultProjectID, IsMilestone: true, EndDate: day(5)})
	db.AssertExists(t, "tasks", map[string]interface{}{"id": m.ID, "is_milestone": true}, false)

	s := db.NewSession()
	defer s.Close()
	update := &Task{ID: m.ID, Title: "M", ProjectID: jane.DefaultProjectID, IsMilestone: false, EndDate: day(5)}
	require.NoError(t, update.Update(s, jane))
	require.NoError(t, s.Commit())

	db.AssertExists(t, "tasks", map[string]interface{}{"id": m.ID, "is_milestone": false}, false)
}
