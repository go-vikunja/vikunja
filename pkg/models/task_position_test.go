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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPositionConflicts(t *testing.T) {
	t.Run("no conflicts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project view 1 has tasks at positions 2 and 4 - no conflicts
		conflicts, err := findPositionConflicts(s, 1, 2)
		require.NoError(t, err)
		assert.Len(t, conflicts, 1) // Only one task at position 2
	})

	t.Run("finds conflicts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Insert two tasks with the same position
		_, err := s.Insert(&TaskPosition{TaskID: 100, ProjectViewID: 1, Position: 999})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 101, ProjectViewID: 1, Position: 999})
		require.NoError(t, err)

		conflicts, err := findPositionConflicts(s, 1, 999)
		require.NoError(t, err)
		assert.Len(t, conflicts, 2)
	})

	t.Run("no conflicts at nonexistent position", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		conflicts, err := findPositionConflicts(s, 1, 12345)
		require.NoError(t, err)
		assert.Empty(t, conflicts)
	})
}

func TestResolveTaskPositionConflicts(t *testing.T) {
	t.Run("no conflict to resolve", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Single task - no conflict
		conflicts := []*TaskPosition{
			{TaskID: 1, ProjectViewID: 1, Position: 100},
		}
		err := resolveTaskPositionConflicts(s, 1, conflicts)
		require.NoError(t, err)
	})

	t.Run("resolves conflicts with neighbors", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Set up: Create positions at 100, 200 (conflict), 200 (conflict), 300
		_, err := s.Insert(&TaskPosition{TaskID: 100, ProjectViewID: 1, Position: 100})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 101, ProjectViewID: 1, Position: 200})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 102, ProjectViewID: 1, Position: 200})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 103, ProjectViewID: 1, Position: 300})
		require.NoError(t, err)

		conflicts := []*TaskPosition{
			{TaskID: 101, ProjectViewID: 1, Position: 200},
			{TaskID: 102, ProjectViewID: 1, Position: 200},
		}

		err = resolveTaskPositionConflicts(s, 1, conflicts)
		require.NoError(t, err)

		// Check that the positions are now different
		var pos1, pos2 TaskPosition
		_, err = s.Where("task_id = ? AND project_view_id = ?", 101, 1).Get(&pos1)
		require.NoError(t, err)
		_, err = s.Where("task_id = ? AND project_view_id = ?", 102, 1).Get(&pos2)
		require.NoError(t, err)

		assert.NotEqual(t, pos1.Position, pos2.Position)
		// Both should be between 100 and 300
		assert.Greater(t, pos1.Position, 100.0)
		assert.Less(t, pos1.Position, 300.0)
		assert.Greater(t, pos2.Position, 100.0)
		assert.Less(t, pos2.Position, 300.0)
	})

	t.Run("resolves conflicts at start (no left neighbor)", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear existing positions for this view to control test data
		_, err := s.Where("project_view_id = ?", 99).Delete(&TaskPosition{})
		require.NoError(t, err)

		// Set up: positions at 50 (conflict), 50 (conflict), 100
		_, err = s.Insert(&TaskPosition{TaskID: 200, ProjectViewID: 99, Position: 50})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 201, ProjectViewID: 99, Position: 50})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 202, ProjectViewID: 99, Position: 100})
		require.NoError(t, err)

		conflicts := []*TaskPosition{
			{TaskID: 200, ProjectViewID: 99, Position: 50},
			{TaskID: 201, ProjectViewID: 99, Position: 50},
		}

		err = resolveTaskPositionConflicts(s, 99, conflicts)
		require.NoError(t, err)

		// Check positions are unique and between 0 and 100
		var pos1, pos2 TaskPosition
		_, err = s.Where("task_id = ? AND project_view_id = ?", 200, 99).Get(&pos1)
		require.NoError(t, err)
		_, err = s.Where("task_id = ? AND project_view_id = ?", 201, 99).Get(&pos2)
		require.NoError(t, err)

		assert.NotEqual(t, pos1.Position, pos2.Position)
		assert.GreaterOrEqual(t, pos1.Position, 0.0)
		assert.Less(t, pos1.Position, 100.0)
		assert.GreaterOrEqual(t, pos2.Position, 0.0)
		assert.Less(t, pos2.Position, 100.0)
	})

	t.Run("resolves conflicts at end (no right neighbor)", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear existing positions for this view
		_, err := s.Where("project_view_id = ?", 98).Delete(&TaskPosition{})
		require.NoError(t, err)

		// Set up: positions at 100, 200 (conflict), 200 (conflict) - no right neighbor
		_, err = s.Insert(&TaskPosition{TaskID: 300, ProjectViewID: 98, Position: 100})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 301, ProjectViewID: 98, Position: 200})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 302, ProjectViewID: 98, Position: 200})
		require.NoError(t, err)

		conflicts := []*TaskPosition{
			{TaskID: 301, ProjectViewID: 98, Position: 200},
			{TaskID: 302, ProjectViewID: 98, Position: 200},
		}

		err = resolveTaskPositionConflicts(s, 98, conflicts)
		require.NoError(t, err)

		// Check positions are unique and > 100
		var pos1, pos2 TaskPosition
		_, err = s.Where("task_id = ? AND project_view_id = ?", 301, 98).Get(&pos1)
		require.NoError(t, err)
		_, err = s.Where("task_id = ? AND project_view_id = ?", 302, 98).Get(&pos2)
		require.NoError(t, err)

		assert.NotEqual(t, pos1.Position, pos2.Position)
		assert.Greater(t, pos1.Position, 100.0)
		assert.Greater(t, pos2.Position, 100.0)
	})

	t.Run("returns error when spacing exhausted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear existing positions for this view
		_, err := s.Where("project_view_id = ?", 97).Delete(&TaskPosition{})
		require.NoError(t, err)

		// Set up: extremely tight spacing that can't accommodate multiple tasks
		// Gap of 2e-9 with 2 conflicts means spacing of ~6.67e-10 < MinPositionSpacing (1e-9)
		basePos := 100.0
		tinyGap := 1e-9
		_, err = s.Insert(&TaskPosition{TaskID: 400, ProjectViewID: 97, Position: basePos})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 401, ProjectViewID: 97, Position: basePos + tinyGap})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 402, ProjectViewID: 97, Position: basePos + tinyGap})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 403, ProjectViewID: 97, Position: basePos + 2*tinyGap})
		require.NoError(t, err)

		conflicts := []*TaskPosition{
			{TaskID: 401, ProjectViewID: 97, Position: basePos + tinyGap},
			{TaskID: 402, ProjectViewID: 97, Position: basePos + tinyGap},
		}

		err = resolveTaskPositionConflicts(s, 97, conflicts)
		assert.True(t, IsErrNeedsFullRecalculation(err))
	})

	t.Run("handles multiple conflicts deterministically", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear existing positions for this view
		_, err := s.Where("project_view_id = ?", 96).Delete(&TaskPosition{})
		require.NoError(t, err)

		// Set up: 4 tasks at the same position
		_, err = s.Insert(&TaskPosition{TaskID: 504, ProjectViewID: 96, Position: 0})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 501, ProjectViewID: 96, Position: 500})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 503, ProjectViewID: 96, Position: 500})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 502, ProjectViewID: 96, Position: 500})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 500, ProjectViewID: 96, Position: 500})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 505, ProjectViewID: 96, Position: 1000})
		require.NoError(t, err)

		conflicts := []*TaskPosition{
			{TaskID: 501, ProjectViewID: 96, Position: 500},
			{TaskID: 503, ProjectViewID: 96, Position: 500},
			{TaskID: 502, ProjectViewID: 96, Position: 500},
			{TaskID: 500, ProjectViewID: 96, Position: 500},
		}

		err = resolveTaskPositionConflicts(s, 96, conflicts)
		require.NoError(t, err)

		// Fetch all positions and verify they are unique and ordered by task ID
		var positions []*TaskPosition
		err = s.Where("project_view_id = ? AND task_id IN (500, 501, 502, 503)", 96).
			OrderBy("task_id ASC").
			Find(&positions)
		require.NoError(t, err)
		require.Len(t, positions, 4)

		// Positions should be strictly increasing (sorted by task_id)
		for i := 1; i < len(positions); i++ {
			assert.Greater(t, positions[i].Position, positions[i-1].Position,
				"Position for task %d should be greater than task %d",
				positions[i].TaskID, positions[i-1].TaskID)
		}

		// All should be between 0 and 1000
		for _, p := range positions {
			assert.Greater(t, p.Position, 0.0)
			assert.Less(t, p.Position, 1000.0)
		}
	})
}

func TestUpdateTaskPositionWithConflictResolution(t *testing.T) {
	t.Run("resolves conflict on update", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear existing positions for this view
		_, err := s.Where("project_view_id = ?", 95).Delete(&TaskPosition{})
		require.NoError(t, err)

		// Set up: two tasks with different positions
		_, err = s.Insert(&TaskPosition{TaskID: 600, ProjectViewID: 95, Position: 100})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 601, ProjectViewID: 95, Position: 200})
		require.NoError(t, err)

		// Update task 600 to have the same position as task 601
		tp := &TaskPosition{
			TaskID:        600,
			ProjectViewID: 95,
			Position:      200,
		}

		err = updateTaskPosition(s, nil, tp)
		require.NoError(t, err)

		// Verify both tasks now have unique positions
		var pos1, pos2 TaskPosition
		_, err = s.Where("task_id = ? AND project_view_id = ?", 600, 95).Get(&pos1)
		require.NoError(t, err)
		_, err = s.Where("task_id = ? AND project_view_id = ?", 601, 95).Get(&pos2)
		require.NoError(t, err)

		assert.NotEqual(t, pos1.Position, pos2.Position)
	})
}

func TestRepairTaskPositions(t *testing.T) {
	t.Run("no duplicates to repair", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear all positions and set up clean data with no duplicates
		_, err := s.Where("project_view_id = ?", 94).Delete(&TaskPosition{})
		require.NoError(t, err)

		_, err = s.Insert(&TaskPosition{TaskID: 700, ProjectViewID: 94, Position: 100})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 701, ProjectViewID: 94, Position: 200})
		require.NoError(t, err)

		result, err := RepairTaskPositions(s, false)
		require.NoError(t, err)

		// View 94 should be scanned but not repaired (no duplicates)
		assert.GreaterOrEqual(t, result.ViewsScanned, 1)
		assert.Empty(t, result.Errors)
	})

	t.Run("repairs duplicates in view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear and set up duplicates
		_, err := s.Where("project_view_id = ?", 93).Delete(&TaskPosition{})
		require.NoError(t, err)

		_, err = s.Insert(&TaskPosition{TaskID: 800, ProjectViewID: 93, Position: 100})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 801, ProjectViewID: 93, Position: 200})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 802, ProjectViewID: 93, Position: 200}) // Duplicate!
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 803, ProjectViewID: 93, Position: 300})
		require.NoError(t, err)

		result, err := RepairTaskPositions(s, false)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, result.ViewsRepaired, 1)
		assert.GreaterOrEqual(t, result.TasksAffected, 2)
		assert.Empty(t, result.Errors)

		// Verify positions are now unique
		var pos1, pos2 TaskPosition
		_, err = s.Where("task_id = ? AND project_view_id = ?", 801, 93).Get(&pos1)
		require.NoError(t, err)
		_, err = s.Where("task_id = ? AND project_view_id = ?", 802, 93).Get(&pos2)
		require.NoError(t, err)

		assert.NotEqual(t, pos1.Position, pos2.Position)
	})

	t.Run("dry run reports without changes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Clear and set up duplicates
		_, err := s.Where("project_view_id = ?", 92).Delete(&TaskPosition{})
		require.NoError(t, err)

		_, err = s.Insert(&TaskPosition{TaskID: 900, ProjectViewID: 92, Position: 500})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 901, ProjectViewID: 92, Position: 500}) // Duplicate!
		require.NoError(t, err)

		result, err := RepairTaskPositions(s, true) // dry run
		require.NoError(t, err)

		assert.GreaterOrEqual(t, result.ViewsRepaired, 1)
		assert.GreaterOrEqual(t, result.TasksAffected, 2)

		// Verify positions are still duplicates (dry run shouldn't change them)
		var pos1, pos2 TaskPosition
		_, err = s.Where("task_id = ? AND project_view_id = ?", 900, 92).Get(&pos1)
		require.NoError(t, err)
		_, err = s.Where("task_id = ? AND project_view_id = ?", 901, 92).Get(&pos2)
		require.NoError(t, err)

		assert.InDelta(t, pos1.Position, pos2.Position, 0) // Still duplicates
	})

	t.Run("handles multiple views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Set up duplicates in two different views
		_, err := s.Where("project_view_id IN (90, 91)").Delete(&TaskPosition{})
		require.NoError(t, err)

		// View 90: duplicates
		_, err = s.Insert(&TaskPosition{TaskID: 1000, ProjectViewID: 90, Position: 100})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 1001, ProjectViewID: 90, Position: 100})
		require.NoError(t, err)

		// View 91: duplicates
		_, err = s.Insert(&TaskPosition{TaskID: 1002, ProjectViewID: 91, Position: 200})
		require.NoError(t, err)
		_, err = s.Insert(&TaskPosition{TaskID: 1003, ProjectViewID: 91, Position: 200})
		require.NoError(t, err)

		result, err := RepairTaskPositions(s, false)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, result.ViewsRepaired, 2)
		assert.GreaterOrEqual(t, result.TasksAffected, 4)
		assert.Empty(t, result.Errors)
	})
}
