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
	"fmt"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSavedFilterUpdateInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "posfilter",
		Filters: &TaskCollection{Filter: "id = 1"},
	}

	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	err = sf.Update(s, u)
	require.NoError(t, err)

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	tp := &TaskPosition{}
	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, 1).Get(tp)
	require.NoError(t, err)
	require.True(t, exists)
	assert.NotZero(t, tp.Position)
}

func TestCronInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "cronfilter",
		Filters: &TaskCollection{Filter: "due_date > '2018-01-01T00:00:00'"},
	}

	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	task := &Task{}
	exists, err = s.Where("id = ?", 5).Get(task)
	require.NoError(t, err)
	require.True(t, exists)

	// Force the task to a zero position in this view to simulate the unhealed
	// state. A task only ever has one position row per view, so update it if it
	// already exists (e.g. created with the filter) instead of inserting a duplicate.
	tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 0}
	hasPosition, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, view.ID).Exist(&TaskPosition{})
	require.NoError(t, err)
	if hasPosition {
		_, err = s.Where("task_id = ? AND project_view_id = ?", task.ID, view.ID).Cols("position").Update(tp)
	} else {
		_, err = s.Insert(tp)
	}
	require.NoError(t, err)

	_, err = calculateNewPositionForTask(s, u, task, view)
	require.NoError(t, err)

	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, task.ID).Get(tp)
	require.NoError(t, err)
	require.True(t, exists)
	assert.NotZero(t, tp.Position)
}

// Task 51 matches "done = false" but is soft-deleted — the bucket heal must not
// resurrect it. The existing heal tests are existence-only and would pass that.
func TestSavedFilterHealDoesNotResurrectSoftDeletedTasks(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "no-resurrect",
		Filters: &TaskCollection{Filter: "done = false"},
	}

	u := &user.User{ID: 1}
	require.NoError(t, sf.Create(s, u))
	require.NoError(t, sf.Update(s, u))
	require.NoError(t, s.Commit())

	db.AssertMissing(t, "task_buckets", map[string]interface{}{"task_id": 51})
	db.AssertMissing(t, "task_positions", map[string]interface{}{"task_id": 51})
}

func TestCronCreatesNonZeroPositions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Create a saved filter
	sf := &SavedFilter{
		Title:   "cron-test-filter",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	// Get the kanban view for this filter
	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	// Simulate what the cron does: call RecalculateTaskPositions
	err = RecalculateTaskPositions(s, view, u)
	require.NoError(t, err)

	// Verify no positions are 0
	zeroCount, err := s.Where("project_view_id = ? AND position = 0", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.Zero(t, zeroCount, "No positions should be 0")

	// Verify all tasks have positions
	positionCount, err := s.Where("project_view_id = ?", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.NotZero(t, positionCount, "Should have positions")
}

func TestFilterUpdateCreatesNonZeroPositions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Create a saved filter
	sf := &SavedFilter{
		Title:   "update-test-filter",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	// Update the filter (this triggers position creation)
	err = sf.Update(s, u)
	require.NoError(t, err)

	// Get the kanban view
	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	// Verify no positions are 0
	zeroCount, err := s.Where("project_view_id = ? AND position = 0", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.Zero(t, zeroCount, "No positions should be 0 after filter update")
}

func TestMultipleNewTasksGetDistinctPositions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Create a saved filter that matches multiple tasks
	sf := &SavedFilter{
		Title:   "multi-task-filter",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	err = sf.Update(s, u)
	require.NoError(t, err)

	// Get the kanban view
	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	// Get all positions
	var positions []*TaskPosition
	err = s.Where("project_view_id = ?", view.ID).Find(&positions)
	require.NoError(t, err)

	// Verify all positions are unique
	seen := make(map[float64]int64)
	for _, p := range positions {
		if existingTaskID, exists := seen[p.Position]; exists {
			t.Errorf("Position %f is duplicated between tasks %d and %d",
				p.Position, existingTaskID, p.TaskID)
		}
		seen[p.Position] = p.TaskID
	}
}

func TestTaskFetchCreatesPositionsOnDemand(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	// Create a saved filter
	sf := &SavedFilter{
		Title:   "on-demand-position-filter",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	err := sf.Create(s, u)
	require.NoError(t, err)

	// Get the list view for this filter
	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindList).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	// Delete any existing positions to simulate a fresh state (before cron runs)
	_, err = s.Where("project_view_id = ?", view.ID).Delete(&TaskPosition{})
	require.NoError(t, err)

	// Verify NO positions exist now
	existingCount, err := s.Where("project_view_id = ?", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.Zero(t, existingCount, "No positions should exist after deletion")

	// Fetch tasks for the view - this should trigger on-demand position creation
	tc := &TaskCollection{
		ProjectID:     view.ProjectID,
		ProjectViewID: view.ID,
	}
	result, _, _, err := tc.ReadAll(s, u, "", 1, 50)
	require.NoError(t, err)

	tasks := result.([]*Task)
	require.NotEmpty(t, tasks, "Should have tasks matching the filter")

	// Verify all returned tasks have non-zero positions
	for _, task := range tasks {
		assert.NotZero(t, task.Position,
			"Task %d (%s) should have non-zero position", task.ID, task.Title)
	}

	// Verify positions were created in database
	createdCount, err := s.Where("project_view_id = ?", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.NotZero(t, createdCount, "Positions should have been created")

	// Verify no zero positions
	zeroCount, err := s.Where("project_view_id = ? AND position = 0", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.Zero(t, zeroCount, "No positions should be zero")
}

func TestIssue724_SortingOnFilteredViews(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	// Create a saved filter
	sf := &SavedFilter{
		Title:   "issue-724-filter",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	err := sf.Create(s, u)
	require.NoError(t, err)

	// Get the list view for this filter
	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindList).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	// Fetch tasks for the view (simulating what the API does)
	tc := &TaskCollection{
		ProjectID:     view.ProjectID,
		ProjectViewID: view.ID,
	}

	// This should trigger position creation
	result, _, _, err := tc.ReadAll(s, u, "", 1, 50)
	require.NoError(t, err)

	tasks := result.([]*Task)

	// Verify all returned tasks have non-zero positions
	for _, task := range tasks {
		assert.NotZero(t, task.Position,
			"Task %d (%s) should have non-zero position", task.ID, task.Title)
	}

	// Verify positions in database are all non-zero
	zeroCount, err := s.Where("project_view_id = ? AND position = 0", view.ID).Count(&TaskPosition{})
	require.NoError(t, err)
	assert.Zero(t, zeroCount,
		"No position=0 records should exist in database for view %d", view.ID)
}

// A task which starts matching a saved filter has no position row in the filter's views yet.
// Its position must be created before the fetch query runs, so it appears at the top on the
// very first fetch instead of landing at the bottom and jumping to the top on the next one.
func TestSavedFilterNewTaskAtTopOnFirstFetch(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	sf := &SavedFilter{
		Title:   "first-fetch-position",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	require.NoError(t, sf.Create(s, u))

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindList).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	// Position all currently matching tasks so the task created below is the only one
	// without a position row, as after the position cron ran.
	require.NoError(t, RecalculateTaskPositions(s, view, u))

	task := &Task{Title: "new task matching the filter", ProjectID: 1}
	require.NoError(t, task.Create(s, u))

	fetchTaskIndex := func() int {
		tc := &TaskCollection{
			ProjectID:     getProjectIDFromSavedFilterID(sf.ID),
			ProjectViewID: view.ID,
			SortBy:        []string{"position"},
			OrderBy:       []string{"asc"},
		}
		result, _, _, err := tc.ReadAll(s, u, "", 1, 100)
		require.NoError(t, err)
		tasks, ok := result.([]*Task)
		require.True(t, ok)
		for i, tt := range tasks {
			if tt.ID == task.ID {
				return i
			}
		}
		t.Fatalf("task %d not found in filter results", task.ID)
		return -1
	}

	first := fetchTaskIndex()
	assert.Equal(t, 0, first, "newly matching task must be at the top on the first fetch")
	assert.Equal(t, first, fetchTaskIndex(), "task order must not change between fetches")
}

// The heal must cover all matching tasks without a position, not just those on the fetched
// page — otherwise tasks beyond the current page keep reordering until the cron runs.
func TestSavedFilterHealsPositionsBeyondFetchedPage(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	sf := &SavedFilter{
		Title:   "heal-all-pages",
		Filters: &TaskCollection{Filter: "done = false"},
	}
	require.NoError(t, sf.Create(s, u))

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?",
		getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindList).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, RecalculateTaskPositions(s, view, u))

	newTasks := make([]*Task, 0, 3)
	for i := 0; i < 3; i++ {
		task := &Task{Title: fmt.Sprintf("unpositioned task %d", i), ProjectID: 1}
		require.NoError(t, task.Create(s, u))
		newTasks = append(newTasks, task)
	}

	// Fetch a single page smaller than the number of matching tasks
	tc := &TaskCollection{
		ProjectID:     getProjectIDFromSavedFilterID(sf.ID),
		ProjectViewID: view.ID,
		SortBy:        []string{"position"},
		OrderBy:       []string{"asc"},
	}
	_, _, _, err = tc.ReadAll(s, u, "", 1, 2)
	require.NoError(t, err)

	for _, task := range newTasks {
		hasPosition, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, view.ID).Exist(&TaskPosition{})
		require.NoError(t, err)
		assert.True(t, hasPosition,
			"task %d must get a position row after one fetch, even when not on the fetched page", task.ID)
	}
}
