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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timePtr(t time.Time) *time.Time { return &t }

// Fixture access graph (pkg/db/fixtures): project 1 is owned by user1 only
// (everyone else a stranger); task 1 lives in project 1. Project 3 is owned by
// user3, with user1 and user2 granted read. user4 has access to neither.
// Entries: 1 = user1 on task 1, 2 = user1 on project 1, 3 = user3 on project 3.

func TestTimeEntry_CanRead(t *testing.T) {
	tests := []struct {
		name    string
		entryID int64
		auth    web.Auth
		wantCan bool
		wantErr func(error) bool
	}{
		{"owner reads task entry", 1, &user.User{ID: 1}, true, nil},
		{"owner reads project entry", 2, &user.User{ID: 1}, true, nil},
		{"reader reads other user's entry on a shared project", 3, &user.User{ID: 1}, true, nil},
		{"stranger denied on owned project", 1, &user.User{ID: 4}, false, nil},
		{"stranger denied on shared project", 3, &user.User{ID: 4}, false, nil},
		{"link share denied", 1, &LinkSharing{ID: 1, ProjectID: 1}, false, nil},
		{"missing entry is a 404", 999, &user.User{ID: 1}, false, IsErrTimeEntryDoesNotExist},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			can, _, err := (&TimeEntry{ID: tt.entryID}).CanRead(s, tt.auth)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, tt.wantErr(err), "unexpected error type: %v", err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCan, can)
		})
	}
}

func TestTimeEntry_CanCreate(t *testing.T) {
	tests := []struct {
		name    string
		entry   *TimeEntry
		auth    web.Auth
		wantCan bool
		wantErr func(error) bool
	}{
		{"on a task in an owned project", &TimeEntry{TaskID: 1}, &user.User{ID: 1}, true, nil},
		{"on an owned project", &TimeEntry{ProjectID: 1}, &user.User{ID: 1}, true, nil},
		{"on a readable project", &TimeEntry{ProjectID: 3}, &user.User{ID: 1}, true, nil},
		{"stranger denied", &TimeEntry{ProjectID: 1}, &user.User{ID: 4}, false, nil},
		{"both task and project is invalid", &TimeEntry{TaskID: 1, ProjectID: 1}, &user.User{ID: 1}, false, IsErrTimeEntryInvalidContainer},
		{"neither task nor project is invalid", &TimeEntry{}, &user.User{ID: 1}, false, IsErrTimeEntryInvalidContainer},
		{"link share denied", &TimeEntry{ProjectID: 1}, &LinkSharing{ID: 1, ProjectID: 1}, false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			can, err := tt.entry.CanCreate(s, tt.auth)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, tt.wantErr(err), "unexpected error type: %v", err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCan, can)
		})
	}
}

// Entry 3 is authored by user3; user1 can read project 3 but is not the author,
// so it can read but not modify.
func TestTimeEntry_CanModify(t *testing.T) {
	tests := []struct {
		name    string
		entryID int64
		auth    web.Auth
		wantCan bool
	}{
		{"author modifies own entry", 1, &user.User{ID: 1}, true},
		{"author modifies own entry on shared project", 3, &user.User{ID: 3}, true},
		{"reader who is not author cannot modify", 3, &user.User{ID: 1}, false},
		{"stranger cannot modify", 3, &user.User{ID: 4}, false},
		{"link share cannot modify", 1, &LinkSharing{ID: 1, ProjectID: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			canUpdate, err := (&TimeEntry{ID: tt.entryID}).CanUpdate(s, tt.auth)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCan, canUpdate, "CanUpdate")

			canDelete, err := (&TimeEntry{ID: tt.entryID}).CanDelete(s, tt.auth)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCan, canDelete, "CanDelete")
		})
	}
}

// Guards the data leak: ReadAll must return only entries on tasks/projects the
// caller can read, since DoReadAll runs no permission check.
func TestTimeEntry_ReadAll(t *testing.T) {
	tests := []struct {
		name    string
		auth    web.Auth
		wantIDs []int64
	}{
		{"user sees every readable entry", &user.User{ID: 1}, []int64{1, 2, 3, 4}},
		{"user sees only entries on projects they can read", &user.User{ID: 2}, []int64{3}},
		{"stranger sees nothing", &user.User{ID: 4}, []int64{}},
		{"link share sees nothing", &LinkSharing{ID: 1, ProjectID: 1}, []int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			result, count, total, err := (&TimeEntry{}).ReadAll(s, tt.auth, "", 1, 50)
			require.NoError(t, err)
			entries, ok := result.([]*TimeEntry)
			require.True(t, ok)

			gotIDs := make([]int64, 0, len(entries))
			for _, e := range entries {
				gotIDs = append(gotIDs, e.ID)
			}
			assert.ElementsMatch(t, tt.wantIDs, gotIDs)
			assert.Equal(t, len(tt.wantIDs), count)
			assert.Equal(t, int64(len(tt.wantIDs)), total)
		})
	}
}

// Filtering reuses the task filter grammar. user1 can read entries 1,2,4
// (project 1) and 3 (project 3, shared) — the filter only narrows that set.
func TestTimeEntry_ReadAll_Filter(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		wantIDs []int64
		wantErr bool
	}{
		{"by user", "user_id = 3", []int64{3}, false},
		{"by task", "task_id = 1", []int64{1, 4}, false},
		{"by project unions task-attached entries", "project_id = 1", []int64{1, 2, 4}, false},
		{"by project negated", "project_id != 1", []int64{3}, false},
		{"by start time", "start_time > '2018-12-01T11:00:00+00:00'", []int64{2, 3, 4}, false},
		{"running timers via null end_time", "end_time = null", []int64{4}, false},
		{"compound and", "user_id = 1 && end_time = null", []int64{4}, false},
		{"compound or", "user_id = 3 || task_id = 1", []int64{1, 3, 4}, false},
		{"in list", "user_id in 1,3", []int64{1, 2, 3, 4}, false},
		{"comment is not filterable", "comment = whatever", nil, true},
		{"unknown field errors", "bogus = 1", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			te := &TimeEntry{Filter: tt.filter}
			result, _, _, err := te.ReadAll(s, &user.User{ID: 1}, "", 1, 50)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			entries, ok := result.([]*TimeEntry)
			require.True(t, ok)
			gotIDs := make([]int64, 0, len(entries))
			for _, e := range entries {
				gotIDs = append(gotIDs, e.ID)
			}
			assert.ElementsMatch(t, tt.wantIDs, gotIDs)
		})
	}
}

// Search matches the entry comment. Comments: 1="Time entry on task 1",
// 2/3 contain "Standalone", 4="Running timer".
func TestTimeEntry_ReadAll_Search(t *testing.T) {
	tests := []struct {
		name    string
		search  string
		wantIDs []int64
	}{
		{"matches a comment", "Running", []int64{4}},
		{"is case-insensitive", "running", []int64{4}},
		{"matches several", "Standalone", []int64{2, 3}},
		{"no match", "nothing matches this", []int64{}},
		{"empty search returns all readable", "", []int64{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			result, _, _, err := (&TimeEntry{}).ReadAll(s, &user.User{ID: 1}, tt.search, 1, 50)
			require.NoError(t, err)
			entries, ok := result.([]*TimeEntry)
			require.True(t, ok)
			gotIDs := make([]int64, 0, len(entries))
			for _, e := range entries {
				gotIDs = append(gotIDs, e.ID)
			}
			assert.ElementsMatch(t, tt.wantIDs, gotIDs)
		})
	}
}

func TestTimeEntry_Create(t *testing.T) {
	t.Run("manual entry keeps its start time and is owned by the caller", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		start := time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)
		end := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
		te := &TimeEntry{TaskID: 1, StartTime: start, EndTime: &end, Comment: "work"}
		require.NoError(t, te.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())

		assert.Equal(t, int64(1), te.UserID)
		assert.True(t, te.StartTime.Equal(start))
		db.AssertExists(t, "time_entries", map[string]interface{}{
			"id":      te.ID,
			"user_id": 1,
			"task_id": 1,
			"comment": "work",
		}, false)
	})

	t.Run("defaults the start time to now when none is given", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{TaskID: 1}
		require.NoError(t, te.Create(s, &user.User{ID: 1}))
		assert.False(t, te.StartTime.IsZero())
	})

	t.Run("a completed manual entry leaves a running timer alone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// entry 4 is user1's running timer
		manual := &TimeEntry{
			TaskID:    1,
			StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
			EndTime:   timePtr(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)),
		}
		require.NoError(t, manual.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())

		running := &TimeEntry{}
		exists, err := s.Where("id = ?", 4).Get(running)
		require.NoError(t, err)
		require.True(t, exists)
		assert.Nil(t, running.EndTime, "a manual entry must not stop the running timer")
	})

	t.Run("auto-stops the caller's running timer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		a := &user.User{ID: 1}

		first := &TimeEntry{TaskID: 1}
		require.NoError(t, first.Create(s, a))
		require.Nil(t, first.EndTime, "first timer should be running")

		second := &TimeEntry{TaskID: 1}
		require.NoError(t, second.Create(s, a))
		require.NoError(t, s.Commit())

		reloaded := &TimeEntry{}
		exists, err := s.Where("id = ?", first.ID).Get(reloaded)
		require.NoError(t, err)
		require.True(t, exists)
		assert.NotNil(t, reloaded.EndTime, "first timer should have been auto-stopped")
		assert.Nil(t, second.EndTime, "second timer should still be running")
	})
}

// A running timer (no end) must round-trip as a NULL end_time: found by the
// null filter and serialized as JSON null, never the 0001-01-01 zero sentinel.
func TestTimeEntry_RunningTimerEndTimeIsNull(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	a := &user.User{ID: 1}

	te := &TimeEntry{TaskID: 1, StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)}
	require.NoError(t, te.Create(s, a))
	require.NoError(t, s.Commit())

	reloaded, err := getTimeEntryByID(s, te.ID)
	require.NoError(t, err)

	marshalled, err := json.Marshal(reloaded)
	require.NoError(t, err)
	assert.Contains(t, string(marshalled), `"end_time":null`)
	assert.NotContains(t, string(marshalled), "0001-01-01")

	// Stored as NULL, so the null filter matches it (not just the fixtures).
	found := &TimeEntry{Filter: "end_time = null"}
	result, _, _, err := found.ReadAll(s, a, "", 1, 50)
	require.NoError(t, err)
	ids := []int64{}
	for _, e := range result.([]*TimeEntry) {
		ids = append(ids, e.ID)
	}
	assert.Contains(t, ids, te.ID)
}

// Regression guard: the permission check must not clobber the update payload.
func TestTimeEntry_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	a := &user.User{ID: 1}

	te := &TimeEntry{
		ID:        1,
		TaskID:    1,
		StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   timePtr(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)),
		Comment:   "updated comment",
	}

	can, err := te.CanUpdate(s, a) // the handler calls this before Update
	require.NoError(t, err)
	require.True(t, can)
	require.NoError(t, te.Update(s, a))
	require.NoError(t, s.Commit())

	assert.Equal(t, "updated comment", te.Comment)
	db.AssertExists(t, "time_entries", map[string]interface{}{
		"id":      1,
		"comment": "updated comment",
	}, false)
}

func TestTimeEntry_UpdateReassignsContainer(t *testing.T) {
	validTimes := func(te *TimeEntry) {
		te.StartTime = time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)
		te.EndTime = timePtr(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC))
	}

	t.Run("moves an entry from a task to a project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		a := &user.User{ID: 1}

		// Entry 1 is on task 1; move it onto project 1 directly.
		te := &TimeEntry{ID: 1, ProjectID: 1}
		validTimes(te)

		can, err := te.CanUpdate(s, a)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, te.Update(s, a))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "time_entries", map[string]interface{}{
			"id":         1,
			"task_id":    0,
			"project_id": 1,
		}, false)
	})

	t.Run("rejects an update that sets both task and project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := (&TimeEntry{ID: 1, TaskID: 1, ProjectID: 1}).CanUpdate(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTimeEntryInvalidContainer(err))
	})

	t.Run("an omitted container keeps the existing one", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		a := &user.User{ID: 1}

		// Entry 1 is on task 1; update only the comment, no container set.
		te := &TimeEntry{ID: 1, Comment: "kept on task"}
		validTimes(te)

		can, err := te.CanUpdate(s, a)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, te.Update(s, a))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "time_entries", map[string]interface{}{
			"id":         1,
			"task_id":    1,
			"project_id": 0,
			"comment":    "kept on task",
		}, false)
	})
}

func TestTimeEntry_UpdateReopenGuard(t *testing.T) {
	a := &user.User{ID: 1}
	someStart := time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)

	t.Run("rejects clearing the end of a completed entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Entry 1 is completed; a nil end would reopen it as a running timer.
		te := &TimeEntry{ID: 1, TaskID: 1, StartTime: someStart} // EndTime nil
		can, err := te.CanUpdate(s, a)
		require.NoError(t, err)
		require.True(t, can)

		err = te.Update(s, a)
		require.Error(t, err)
		assert.True(t, IsErrTimeEntryAlreadyEnded(err), "unexpected error type: %v", err)
	})

	t.Run("allows editing a running entry while it stays running", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Entry 4 is user1's running timer; keeping it running (nil end) is fine.
		te := &TimeEntry{ID: 4, TaskID: 1, StartTime: someStart, Comment: "edited"} // EndTime nil
		can, err := te.CanUpdate(s, a)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, te.Update(s, a))
	})
}

func TestTimeEntry_RejectsInvertedInterval(t *testing.T) {
	a := &user.User{ID: 1}
	start := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	before := time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)

	t.Run("create rejects an end before the start", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{TaskID: 1, StartTime: start, EndTime: timePtr(before)}
		err := te.Create(s, a)
		require.Error(t, err)
		assert.True(t, IsErrTimeEntryEndBeforeStart(err), "unexpected error type: %v", err)
	})

	t.Run("create allows an end equal to the start", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{TaskID: 1, StartTime: start, EndTime: timePtr(start)}
		require.NoError(t, te.Create(s, a))
	})

	t.Run("create allows a running timer with no end", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{TaskID: 1, StartTime: start} // EndTime nil
		require.NoError(t, te.Create(s, a))
	})

	t.Run("update rejects an end before the start", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Entry 1 is user1's completed entry.
		te := &TimeEntry{ID: 1, TaskID: 1, StartTime: start, EndTime: timePtr(before)}
		can, err := te.CanUpdate(s, a)
		require.NoError(t, err)
		require.True(t, can)

		err = te.Update(s, a)
		require.Error(t, err)
		assert.True(t, IsErrTimeEntryEndBeforeStart(err), "unexpected error type: %v", err)
	})
}

func TestTimeEntry_StopRunningTimer(t *testing.T) {
	t.Run("stops the caller's running timer and returns it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		entry, err := StopRunningTimer(s, &user.User{ID: 1}) // entry 4
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Equal(t, int64(4), entry.ID)
		assert.NotNil(t, entry.EndTime)

		reloaded := &TimeEntry{}
		_, err = s.Where("id = ?", 4).Get(reloaded)
		require.NoError(t, err)
		assert.NotNil(t, reloaded.EndTime, "end time should be persisted")
	})

	t.Run("errors when no timer is running", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := StopRunningTimer(s, &user.User{ID: 2}) // user2 has no entries
		require.Error(t, err)
		assert.True(t, IsErrNoRunningTimer(err), "unexpected error type: %v", err)
	})

	t.Run("denies a link share and leaves the matching user's timer running", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Share id 1 collides with user 1, whose entry 4 is a running timer.
		_, err := StopRunningTimer(s, &LinkSharing{ID: 1, ProjectID: 1})
		require.Error(t, err)
		assert.True(t, IsErrGenericForbidden(err), "unexpected error type: %v", err)

		running := &TimeEntry{}
		exists, err := s.Where("id = ?", 4).Get(running)
		require.NoError(t, err)
		require.True(t, exists)
		assert.Nil(t, running.EndTime, "the user's timer must not have been stopped by a link share")
	})
}

func TestTimeEntry_Events(t *testing.T) {
	u := &user.User{ID: 1}
	someStart := time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)
	someEnd := timePtr(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC))

	t.Run("create dispatches created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{TaskID: 1, StartTime: someStart, EndTime: someEnd}
		require.NoError(t, te.Create(s, u))
		require.NoError(t, s.Commit())
		events.DispatchPending(s)
		events.AssertDispatched(t, &TimeEntryCreatedEvent{})
	})

	t.Run("update dispatches updated", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{ID: 1, TaskID: 1, StartTime: someStart, EndTime: someEnd, Comment: "edited"}
		can, err := te.CanUpdate(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, te.Update(s, u))
		require.NoError(t, s.Commit())
		events.DispatchPending(s)
		events.AssertDispatched(t, &TimeEntryUpdatedEvent{})
	})

	t.Run("delete dispatches deleted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, (&TimeEntry{ID: 1}).Delete(s, u))
		require.NoError(t, s.Commit())
		events.DispatchPending(s)
		events.AssertDispatched(t, &TimeEntryDeletedEvent{})
	})

	t.Run("starting a timer dispatches created plus updated for the auto-stopped entry", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		// entry 4 is user1's running timer; a new running timer auto-stops it
		require.NoError(t, (&TimeEntry{TaskID: 1}).Create(s, u))
		require.NoError(t, s.Commit())
		events.DispatchPending(s)
		events.AssertDispatched(t, &TimeEntryCreatedEvent{})
		events.AssertDispatched(t, &TimeEntryUpdatedEvent{})
	})

	t.Run("a completed manual entry dispatches only created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		te := &TimeEntry{TaskID: 1, StartTime: someStart, EndTime: someEnd}
		require.NoError(t, te.Create(s, u))
		require.NoError(t, s.Commit())
		events.DispatchPending(s)
		assert.Equal(t, 1, events.CountDispatchedEvents((&TimeEntryCreatedEvent{}).Name()))
		assert.Equal(t, 0, events.CountDispatchedEvents((&TimeEntryUpdatedEvent{}).Name()), "a completed manual entry must not auto-stop")
	})

	t.Run("StopRunningTimer dispatches updated", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		_, err := StopRunningTimer(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		events.DispatchPending(s)
		events.AssertDispatched(t, &TimeEntryUpdatedEvent{})
	})
}

func TestTimeEntry_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	require.NoError(t, (&TimeEntry{ID: 1}).Delete(s, &user.User{ID: 1}))
	require.NoError(t, s.Commit())
	db.AssertMissing(t, "time_entries", map[string]interface{}{"id": 1})
}

func TestTimeEntry_TaskCount(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("attaches counts for a licensed, non-share caller", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		task1 := &Task{ID: 1} // fixtures: time entries 1 and 4 are attached to task 1
		task2 := &Task{ID: 2} // no time entries
		taskMap := map[int64]*Task{1: task1, 2: task2}

		require.NoError(t, addTimeEntriesCountToTasks(s, u, []int64{1, 2}, taskMap))

		require.NotNil(t, task1.TimeEntriesCount)
		assert.Equal(t, int64(2), *task1.TimeEntriesCount)
		require.NotNil(t, task2.TimeEntriesCount)
		assert.Equal(t, int64(0), *task2.TimeEntriesCount)
	})

	t.Run("leaves the count unset for a link share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		task1 := &Task{ID: 1}
		taskMap := map[int64]*Task{1: task1}
		require.NoError(t, addTimeEntriesCountToTasks(s, &LinkSharing{ID: 1}, []int64{1}, taskMap))
		assert.Nil(t, task1.TimeEntriesCount, "link shares must not learn time-entry counts")
	})

	t.Run("leaves the count unset when the feature is unlicensed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.ResetForTests() // feature disabled

		task1 := &Task{ID: 1}
		taskMap := map[int64]*Task{1: task1}
		require.NoError(t, addTimeEntriesCountToTasks(s, u, []int64{1}, taskMap))
		assert.Nil(t, task1.TimeEntriesCount, "an unlicensed instance must not expose counts")
	})
}
