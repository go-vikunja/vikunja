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
	"math"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// positionsOf returns the position each task has in a view, in the order the tasks were
// passed in.
func positionsOf(t *testing.T, s *xorm.Session, viewID int64, tasks []*Task) []float64 {
	positions := []float64{}
	for _, task := range tasks {
		p := &TaskPosition{}
		has, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, viewID).Get(p)
		require.NoError(t, err)
		require.True(t, has, "task %d has no position in view %d", task.ID, viewID)
		positions = append(positions, p.Position)
	}
	return positions
}

func TestBulkTaskCreate_Create(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("keeps the order the tasks were passed in", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "first", ProjectID: 1},
			{Title: "second", ProjectID: 1},
			{Title: "third", ProjectID: 1},
		}}

		require.NoError(t, bt.Create(s, u))

		// View 1 has tasks at position 2 and 4, so the batch is spread over the gap below 2
		positions := positionsOf(t, s, 1, bt.Tasks)
		assert.Equal(t, []float64{0.5, 1, 1.5}, positions)

		// Indexes count up per project, without the caller having to say so
		assert.Equal(t, []int64{35, 36, 37}, []int64{bt.Tasks[0].Index, bt.Tasks[1].Index, bt.Tasks[2].Index})
	})

	t.Run("falls back to the default position in a view without positions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "first", ProjectID: 1},
			{Title: "second", ProjectID: 1},
		}}

		require.NoError(t, bt.Create(s, u))

		// View 2 of project 1 has no positioned tasks at all
		positions := positionsOf(t, s, 2, bt.Tasks)
		assert.Equal(t, []float64{
			float64(bt.Tasks[0].Index) * math.Pow(2, 16),
			float64(bt.Tasks[1].Index) * math.Pow(2, 16),
		}, positions)
		assert.Less(t, positions[0], positions[1])
	})

	t.Run("keeps a position the caller passed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "with position", ProjectID: 1, Position: 12345},
			{Title: "without position", ProjectID: 1},
		}}

		require.NoError(t, bt.Create(s, u))

		// The passed position does not take up a slot, so the other task is placed as if
		// it were the only one in the batch
		assert.Equal(t, []float64{12345, 1}, positionsOf(t, s, 1, bt.Tasks))
	})

	t.Run("numbers and places tasks per project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "first in 1", ProjectID: 1},
			{Title: "first in 22", ProjectID: 22},
			{Title: "second in 1", ProjectID: 1},
		}}

		require.NoError(t, bt.Create(s, u))

		inProjectOne := []*Task{bt.Tasks[0], bt.Tasks[2]}
		assert.Equal(t, []float64{2.0 / 3, 4.0 / 3}, positionsOf(t, s, 1, inProjectOne))
		assert.Equal(t, []int64{35, 36}, []int64{bt.Tasks[0].Index, bt.Tasks[2].Index})

		// Project 22 already holds two tasks, so it numbers its own and places the task in
		// its own views
		assert.Equal(t, int64(3), bt.Tasks[1].Index)
		assert.Equal(t, []float64{3 * math.Pow(2, 16)}, positionsOf(t, s, 85, []*Task{bt.Tasks[1]}))
	})

	t.Run("recalculates when the gap is too small for the batch", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Where("project_view_id = ?", 1).Delete(&TaskPosition{})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 1, ProjectViewID: 1, Position: MinPositionSpacing})
		require.NoError(t, err)

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "first", ProjectID: 1},
			{Title: "second", ProjectID: 1},
			{Title: "third", ProjectID: 1},
		}}

		require.NoError(t, bt.Create(s, u))

		positions := positionsOf(t, s, 1, bt.Tasks)
		assert.Less(t, positions[0], positions[1])
		assert.Less(t, positions[1], positions[2])
		for _, p := range positions {
			assert.GreaterOrEqual(t, p, MinPositionSpacing)
		}

		// The recalculation moved the existing task out of the way instead of letting the
		// batch collide with it
		existing := positionsOf(t, s, 1, []*Task{{ID: 1}})
		assert.Greater(t, existing[0], positions[2])
	})

	t.Run("creates nothing when one task fails", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "created before the failing one", ProjectID: 1},
			{Title: "", ProjectID: 1},
		}}

		err := bt.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskCannotBeEmpty(err))
		require.NoError(t, s.Rollback())

		db.AssertMissing(t, "tasks", map[string]interface{}{"title": "created before the failing one"})
	})

	t.Run("empty batch", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{}
		err := bt.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBulkTasksNeedAtLeastOne(err))
	})
}

func TestBulkTaskCreate_CanCreate(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("own project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{{Title: "test", ProjectID: 1}}}
		can, err := bt.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("project shared with write access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{{Title: "test", ProjectID: 10}}}
		can, err := bt.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("read only project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{{Title: "test", ProjectID: 3}}}
		can, err := bt.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("one project without write access rejects the whole batch", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{Tasks: []*Task{
			{Title: "allowed", ProjectID: 1},
			{Title: "not allowed", ProjectID: 3},
		}}
		can, err := bt.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("empty batch", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bt := &BulkTaskCreate{}
		can, err := bt.CanCreate(s, u)
		require.Error(t, err)
		assert.False(t, can)
	})
}
