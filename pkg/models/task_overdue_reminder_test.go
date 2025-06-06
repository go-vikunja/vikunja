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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUndoneOverDueTasks(t *testing.T) {
	t.Run("no undone tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-01-01T01:13:00Z")
		require.NoError(t, err)
		tasks, err := getUndoneOverdueTasks(s, now)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
	t.Run("undone overdue", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T09:00:00Z")
		require.NoError(t, err)
		uts, err := getUndoneOverdueTasks(s, now)
		require.NoError(t, err)
		assert.Len(t, uts, 1)
		assert.Len(t, uts[1].tasks, 2)
		// The tasks don't always have the same order, so we only check their presence, not their position.
		var task5Present bool
		var task6Present bool
		for _, t := range uts[1].tasks {
			if t.ID == 5 {
				task5Present = true
			}
			if t.ID == 6 {
				task6Present = true
			}
		}
		assert.Truef(t, task5Present, "expected task 5 to be present but was not")
		assert.Truef(t, task6Present, "expected task 6 to be present but was not")
	})
	t.Run("done overdue", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-11-01T01:13:00Z")
		require.NoError(t, err)
		tasks, err := getUndoneOverdueTasks(s, now)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
}
