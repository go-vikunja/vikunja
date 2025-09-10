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
}
