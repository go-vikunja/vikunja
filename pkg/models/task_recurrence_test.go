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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func TestTaskRecurrenceUndo(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	owner := &user.User{ID: 1}
	due := time.Now().UTC().Truncate(time.Second).Add(48 * time.Hour)
	task := &Task{
		Title: "Recurring task", Description: `<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><p>Check</p></li></ul>`,
		ProjectID: 1, DueDate: due, StartDate: due.Add(-time.Hour), EndDate: due.Add(time.Hour),
		RepeatAfter: 86400, Reminders: []*TaskReminder{{Reminder: due.Add(-time.Minute)}},
	}
	require.NoError(t, task.Create(s, owner))
	_, err := s.ID(2).Cols("limit").Update(&Bucket{})
	require.NoError(t, err)
	require.NoError(t, updateTaskBucket(s, owner, &TaskBucket{TaskID: task.ID, BucketID: 2, ProjectViewID: 4, ProjectID: 1}))
	require.NoError(t, updateTaskPosition(s, owner, &TaskPosition{TaskID: task.ID, ProjectViewID: 4, Position: 123456789}))
	require.NoError(t, s.Commit())

	stored, err := GetTaskSimple(s, &Task{ID: task.ID})
	require.NoError(t, err)
	before, err := snapshotTaskOccurrence(s, stored, []int64{4})
	require.NoError(t, err)
	require.EqualValues(t, 2, before.Buckets[0].BucketID)

	complete := &TaskRecurrence{TaskID: task.ID}
	can, err := complete.CanCreate(s, owner)
	require.NoError(t, err)
	require.True(t, can)
	can, err = complete.CanCreate(s, &user.User{ID: 13})
	require.True(t, err != nil || !can)
	require.NoError(t, complete.Create(s, owner))
	require.NoError(t, s.Commit())
	require.False(t, complete.Task.Done)
	require.True(t, complete.Task.DueDate.Equal(due.Add(24*time.Hour)))
	require.NotEmpty(t, complete.UndoToken)
	advanced, err := snapshotTaskOccurrence(s, *complete.Task, []int64{4})
	require.NoError(t, err)
	require.EqualValues(t, 1, advanced.Buckets[0].BucketID)
	require.NotEqual(t, before.Description, advanced.Description)
	require.False(t, advanced.DoneAt.IsZero())

	undo := &TaskRecurrence{TaskID: task.ID, UndoToken: complete.UndoToken}
	require.NoError(t, undo.Create(s, owner))
	require.NoError(t, s.Commit())
	restored, err := GetTaskSimple(s, &Task{ID: task.ID})
	require.NoError(t, err)
	after, err := snapshotTaskOccurrence(s, restored, []int64{4})
	require.NoError(t, err)
	require.Equal(t, before.Description, after.Description)
	require.True(t, before.DoneAt.Equal(after.DoneAt))
	require.True(t, before.DueDate.Equal(after.DueDate))
	require.True(t, before.StartDate.Equal(after.StartDate))
	require.True(t, before.EndDate.Equal(after.EndDate))
	require.Equal(t, before.Buckets, after.Buckets)
	require.Equal(t, before.Positions, after.Positions)
	require.Len(t, after.Reminders, 1)
	require.True(t, before.Reminders[0].Reminder.Equal(after.Reminders[0].Reminder))
	require.ErrorIs(t, (&TaskRecurrence{TaskID: task.ID, UndoToken: complete.UndoToken}).Create(s, owner), errTaskOccurrenceConflict{})
	require.NoError(t, s.Rollback())

	second := &TaskRecurrence{TaskID: task.ID}
	require.NoError(t, second.Create(s, owner))
	require.NoError(t, (&Task{ID: task.ID, DueDate: due.Add(72 * time.Hour)}).updateSingleTask(s, owner, []string{"due_date"}))
	require.ErrorIs(t, (&TaskRecurrence{TaskID: task.ID, UndoToken: second.UndoToken}).Create(s, owner), errTaskOccurrenceConflict{})
}
