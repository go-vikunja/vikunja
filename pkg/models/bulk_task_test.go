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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkTask_Update(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("successful update across projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 6}

		bt := &BulkTask{
			TaskIDs: []int64{15, 16},
			Fields:  []string{"title"},
			Values:  &Task{Title: "bulkupdated"},
		}

		allowed, err := bt.CanUpdate(s, u)
		require.NoError(t, err)
		require.True(t, allowed)

		err = bt.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 15, "title": "bulkupdated", "done": false}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 16, "title": "bulkupdated", "done": false}, false)
	})

	t.Run("unauthorized task prevents update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTask{
			TaskIDs: []int64{10, 14},
			Fields:  []string{"title"},
			Values:  &Task{Title: "bulkupdated"},
		}

		allowed, err := bt.CanUpdate(s, u)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("invalid field", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTask{
			TaskIDs: []int64{10},
			Fields:  []string{"invalid"},
			Values:  &Task{Title: "bulkupdated"},
		}

		allowed, err := bt.CanUpdate(s, u)
		require.NoError(t, err)
		require.True(t, allowed)

		err = bt.Update(s, u)
		require.Error(t, err)
		assert.IsType(t, ErrInvalidTaskColumn{}, err)
	})

	t.Run("update done_at when bulk marking tasks done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTask{
			TaskIDs: []int64{1, 3},
			Fields:  []string{"done"},
			Values:  &Task{Done: true},
		}

		allowed, err := bt.CanUpdate(s, u)
		require.NoError(t, err)
		require.True(t, allowed)

		err = bt.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 1, "done": false, "done_at": nil})
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 3, "done": false, "done_at": nil})

		require.Len(t, bt.Tasks, 2)
		assert.NotZero(t, bt.Tasks[0].DoneAt)
		assert.NotZero(t, bt.Tasks[1].DoneAt)
	})
}

func TestBulkTaskDoneAtOverride(t *testing.T) {
	u := &user.User{ID: 1}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	userProvidedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	bt := &BulkTask{
		TaskIDs: []int64{1, 3},
		Fields:  []string{"done", "done_at"},
		Values:  &Task{Done: true, DoneAt: userProvidedTime},
	}

	err := bt.Update(s, u)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	require.Len(t, bt.Tasks, 2)
	assert.True(t, bt.Tasks[0].Done)
	assert.True(t, bt.Tasks[1].Done)
	assert.NotEqual(t, userProvidedTime, bt.Tasks[0].DoneAt)
	assert.NotEqual(t, userProvidedTime, bt.Tasks[1].DoneAt)
	assert.WithinDuration(t, time.Now(), bt.Tasks[0].DoneAt, time.Second*2)
	assert.WithinDuration(t, time.Now(), bt.Tasks[1].DoneAt, time.Second*2)
}
