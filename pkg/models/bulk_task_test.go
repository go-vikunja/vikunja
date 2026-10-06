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
		var expectedErr ErrInvalidTaskColumn
		assert.ErrorAs(t, err, &expectedErr)
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

	t.Run("bulk update without assignees field keeps assignees", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 30 has users 1 and 2 assigned in the fixtures. A bulk update
		// that only names "priority" in fields must not touch them (#4109).
		bt := &BulkTask{
			TaskIDs: []int64{30},
			Fields:  []string{"priority"},
			Values:  &Task{Priority: 3},
		}

		allowed, err := bt.CanUpdate(s, u)
		require.NoError(t, err)
		require.True(t, allowed)

		err = bt.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{"id": 30, "priority": 3}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": 30, "user_id": 1}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": 30, "user_id": 2}, false)
	})

	t.Run("bulk update with assignees field replaces assignees", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTask{
			TaskIDs: []int64{30},
			Fields:  []string{"assignees"},
			Values:  &Task{Assignees: []*user.User{{ID: 1}}},
		}

		allowed, err := bt.CanUpdate(s, u)
		require.NoError(t, err)
		require.True(t, allowed)

		err = bt.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": 30, "user_id": 1}, false)
		db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": 30, "user_id": 2})
	})

	t.Run("don't update done_at when bulk marking tasks done", func(t *testing.T) {
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
		require.Error(t, err)
		var expectedErr ErrInvalidTaskColumn
		assert.ErrorAs(t, err, &expectedErr)
	})
}
