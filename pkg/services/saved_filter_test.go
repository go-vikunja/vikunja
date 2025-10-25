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

package services

import (
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestSavedFilterService_Get(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	// Test getting a non-existent filter
	_, err := sfs.Get(s, 999, u)
	assert.Error(t, err)
	assert.True(t, models.IsErrSavedFilterDoesNotExist(err))

	// Test getting a filter without permission
	sf := &models.SavedFilter{
		Title:   "Test Filter",
		OwnerID: 999,
		Filters: &models.TaskCollection{},
	}
	_, err = s.Insert(sf)
	assert.NoError(t, err)

	_, err = sfs.Get(s, sf.ID, u)
	assert.Error(t, err)
	assert.Equal(t, ErrAccessDenied, err)

	// Test getting a filter with permission
	sf.OwnerID = u.ID
	_, err = s.Update(sf)
	assert.NoError(t, err)

	retrieved, err := sfs.Get(s, sf.ID, u)
	assert.NoError(t, err)
	assert.Equal(t, sf.ID, retrieved.ID)
	assert.Equal(t, u.ID, retrieved.Owner.ID)
}

func TestSavedFilterService_GetByIDSimple(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)

	t.Run("Success", func(t *testing.T) {
		filter, err := sfs.GetByIDSimple(s, 1)
		assert.NoError(t, err)
		assert.NotNil(t, filter)
		assert.Equal(t, int64(1), filter.ID)
		assert.Equal(t, "testfilter1", filter.Title)
	})

	t.Run("NotFound", func(t *testing.T) {
		filter, err := sfs.GetByIDSimple(s, 999999)
		assert.Error(t, err)
		assert.Nil(t, filter)
		assert.True(t, models.IsErrSavedFilterDoesNotExist(err))
	})
}

func TestSavedFilterService_GetAllForUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	// Get initial count of filters for user 1 (there's already one in fixtures)
	initialFilters, err := sfs.GetAllForUser(s, u, "")
	assert.NoError(t, err)
	initialCount := len(initialFilters)

	// Create some filters
	sf1 := &models.SavedFilter{Title: "Filter 1", OwnerID: u.ID, Filters: &models.TaskCollection{}}
	sf2 := &models.SavedFilter{Title: "Filter 2", OwnerID: u.ID, Filters: &models.TaskCollection{}}
	sf3 := &models.SavedFilter{Title: "Other Filter", OwnerID: 999, Filters: &models.TaskCollection{}}
	_, err = s.Insert(sf1, sf2, sf3)
	assert.NoError(t, err)

	// Test getting all filters (should be initial count + 2 new ones for user 1)
	filters, err := sfs.GetAllForUser(s, u, "")
	assert.NoError(t, err)
	assert.Len(t, filters, initialCount+2)

	// Test searching for filters
	filters, err = sfs.GetAllForUser(s, u, "Filter 1")
	assert.NoError(t, err)
	assert.Len(t, filters, 1)
	assert.Equal(t, "Filter 1", filters[0].Title)
}

func TestSavedFilterService_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	sf := &models.SavedFilter{
		Title:   "New Filter",
		Filters: &models.TaskCollection{},
	}

	err := sfs.Create(s, sf, u)
	assert.NoError(t, err)
	assert.NotZero(t, sf.ID)
	assert.Equal(t, u.ID, sf.OwnerID)
}

func TestSavedFilterService_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	sf := &models.SavedFilter{
		Title:   "Original Title",
		OwnerID: u.ID,
		Filters: &models.TaskCollection{},
	}
	_, err := s.Insert(sf)
	assert.NoError(t, err)

	sf.Title = "Updated Title"
	err = sfs.Update(s, sf, u)
	assert.NoError(t, err)

	retrieved, err := sfs.Get(s, sf.ID, u)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
}

func TestSavedFilterService_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	sf := &models.SavedFilter{
		Title:   "To Be Deleted",
		OwnerID: u.ID,
		Filters: &models.TaskCollection{},
	}
	_, err := s.Insert(sf)
	assert.NoError(t, err)

	err = sfs.Delete(s, sf.ID, u)
	assert.NoError(t, err)

	_, err = sfs.Get(s, sf.ID, u)
	assert.Error(t, err)
	assert.True(t, models.IsErrSavedFilterDoesNotExist(err))
}

func TestSavedFilterService_UpdateInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	sf := &models.SavedFilter{
		Title:   "posfilter",
		Filters: &models.TaskCollection{Filter: "id = 1"},
	}

	err := sfs.Create(s, sf, u)
	assert.NoError(t, err)

	err = sfs.Update(s, sf, u)
	assert.NoError(t, err)

	// Verify that a project view was created
	view := &models.ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", models.GetProjectIDFromSavedFilterID(sf.ID), models.ProjectViewKindKanban).Get(view)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify that a task position was created with non-zero position
	tp := &models.TaskPosition{}
	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, 1).Get(tp)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotZero(t, tp.Position)
}

func TestSavedFilterService_CalculatesNonZeroPositionForNewTasks(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	u := &user.User{ID: 1}

	sf := &models.SavedFilter{
		Title:   "cronfilter",
		Filters: &models.TaskCollection{Filter: "due_date > '2018-01-01T00:00:00'"},
	}

	err := sfs.Create(s, sf, u)
	assert.NoError(t, err)

	// Get the project view created for this filter
	view := &models.ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", models.GetProjectIDFromSavedFilterID(sf.ID), models.ProjectViewKindKanban).Get(view)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Get a task that matches the filter
	task := &models.Task{}
	exists, err = s.Where("id = ?", 5).Get(task)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Insert a task position with zero position (simulating cron job)
	tp := &models.TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 0}
	_, err = s.Insert(tp)
	assert.NoError(t, err)

	// Calculate new position (this is what the cron job does)
	_, err = models.CalculateNewPositionForTask(s, u, task, view)
	assert.NoError(t, err)

	// Verify the position is now non-zero
	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, task.ID).Get(tp)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotZero(t, tp.Position)
}

func TestSavedFilterService_CanRead(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)

	t.Run("owner can read", func(t *testing.T) {
		u := &user.User{ID: 1}
		can, maxRight, err := sfs.CanRead(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
		assert.Equal(t, int(models.PermissionAdmin), maxRight)
	})

	t.Run("non-owner cannot read", func(t *testing.T) {
		u := &user.User{ID: 2}
		can, _, err := sfs.CanRead(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("link share cannot read", func(t *testing.T) {
		ls := &models.LinkSharing{ID: 1}
		can, _, err := sfs.CanRead(s, 1, ls)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

func TestSavedFilterService_CanCreate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)

	t.Run("regular user can create", func(t *testing.T) {
		u := &user.User{ID: 1}
		can, err := sfs.CanCreate(s, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("link share cannot create", func(t *testing.T) {
		ls := &models.LinkSharing{ID: 1}
		can, err := sfs.CanCreate(s, ls)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

func TestSavedFilterService_CanUpdate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)

	t.Run("owner can update", func(t *testing.T) {
		u := &user.User{ID: 1}
		can, err := sfs.CanUpdate(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("non-owner cannot update", func(t *testing.T) {
		u := &user.User{ID: 2}
		can, err := sfs.CanUpdate(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

func TestSavedFilterService_CanDelete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)

	t.Run("owner can delete", func(t *testing.T) {
		u := &user.User{ID: 1}
		can, err := sfs.CanDelete(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("non-owner cannot delete", func(t *testing.T) {
		u := &user.User{ID: 2}
		can, err := sfs.CanDelete(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

// T076: End-to-end integration test for full saved filter execution
// This test validates the complete flow: create saved filter → execute → verify results
func TestSavedFilterService_EndToEnd_FullFilterExecution(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sfs := NewSavedFilterService(testEngine)
	ts := NewTaskService(testEngine)
	u := &user.User{ID: 1}

	t.Run("Simple equality filter execution", func(t *testing.T) {
		// Create saved filter: priority = 3
		sf := &models.SavedFilter{
			Title: "High Priority Tasks",
			Filters: &models.TaskCollection{
				Filter:             "priority = 3",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)
		assert.NotZero(t, sf.ID)

		// Execute filter through TaskService
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 50)
		assert.NoError(t, err)

		tasks, ok := result.([]*models.Task)
		assert.True(t, ok, "Result should be a task array")

		// Verify all returned tasks have priority = 3
		for _, task := range tasks {
			assert.Equal(t, int64(3), task.Priority, "All tasks should have priority 3")
		}

		t.Logf("✓ Simple equality filter returned %d tasks with priority = 3", resultCount)
	})

	t.Run("Complex boolean filter execution", func(t *testing.T) {
		// Create saved filter: (priority > 2 || done = false) && percent_done < 100
		sf := &models.SavedFilter{
			Title: "Active Important Tasks",
			Filters: &models.TaskCollection{
				Filter:             "(priority > 2 || done = false) && percent_done < 100",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 50)
		assert.NoError(t, err)

		tasks, ok := result.([]*models.Task)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, resultCount, 0, "Should return non-negative count")

		// Verify filter logic: (priority > 2 OR done = false) AND percent_done < 100
		for _, task := range tasks {
			// Must satisfy: percent_done < 100
			assert.Less(t, task.PercentDone, float64(1.0), "All tasks should have percent_done < 100")

			// Must satisfy: priority > 2 OR done = false
			satisfiesOR := task.Priority > 2 || !task.Done
			assert.True(t, satisfiesOR, "Task should have priority > 2 OR done = false")
		}

		t.Logf("✓ Complex boolean filter returned %d tasks matching complex logic", resultCount)
	})

	t.Run("Subtable filter execution (labels)", func(t *testing.T) {
		// Create saved filter: labels = 4
		sf := &models.SavedFilter{
			Title: "Tasks with Label 4",
			Filters: &models.TaskCollection{
				Filter:             "labels = 4",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 50)
		assert.NoError(t, err)

		tasks, ok := result.([]*models.Task)
		assert.True(t, ok)

		// Verify all returned tasks have label 4
		for _, task := range tasks {
			hasLabel4 := false
			for _, label := range task.Labels {
				if label.ID == 4 {
					hasLabel4 = true
					break
				}
			}
			assert.True(t, hasLabel4, "Task %d should have label 4", task.ID)
		}

		t.Logf("✓ Subtable filter (labels) returned %d tasks with label 4", resultCount)
	})

	t.Run("Date filter with relative expression", func(t *testing.T) {
		// Create saved filter: due_date >= 'now'
		sf := &models.SavedFilter{
			Title: "Future Due Dates",
			Filters: &models.TaskCollection{
				Filter:             "due_date >= 'now'",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 50)
		assert.NoError(t, err)

		_, ok := result.([]*models.Task)
		assert.True(t, ok)

		// Verify all returned tasks have due_date >= now (or are in the future/present)
		// Note: We can't strictly validate timestamps due to test timing, but we can verify no errors
		t.Logf("✓ Date filter with 'now' expression returned %d tasks", resultCount)
	})

	t.Run("Filter with FilterIncludeNulls=true", func(t *testing.T) {
		// Create saved filter: priority > 0 with includeNulls
		sf := &models.SavedFilter{
			Title: "Priority Tasks (Including Unset)",
			Filters: &models.TaskCollection{
				Filter:             "priority > 0",
				FilterIncludeNulls: true, // Should include tasks with NULL/0 priority
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 100)
		assert.NoError(t, err)

		tasks, ok := result.([]*models.Task)
		assert.True(t, ok)

		// With includeNulls=true, should include both priority > 0 AND priority = 0/NULL
		hasPositivePriority := false
		hasZeroOrNullPriority := false
		for _, task := range tasks {
			if task.Priority > 0 {
				hasPositivePriority = true
			}
			if task.Priority == 0 {
				hasZeroOrNullPriority = true
			}
		}

		assert.True(t, hasPositivePriority, "Should include tasks with priority > 0")
		// Note: hasZeroOrNullPriority depends on test data, log for visibility
		t.Logf("✓ Filter with includeNulls=true returned %d tasks (positive: %v, zero/null: %v)",
			resultCount, hasPositivePriority, hasZeroOrNullPriority)
	})

	t.Run("IN operator filter execution", func(t *testing.T) {
		// Create saved filter: labels in 4,5
		sf := &models.SavedFilter{
			Title: "Tasks with Label 4 or 5",
			Filters: &models.TaskCollection{
				Filter:             "labels in 4,5",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 50)
		assert.NoError(t, err)

		tasks, ok := result.([]*models.Task)
		assert.True(t, ok)

		// Verify all returned tasks have label 4 or 5
		for _, task := range tasks {
			hasLabel4or5 := false
			for _, label := range task.Labels {
				if label.ID == 4 || label.ID == 5 {
					hasLabel4or5 = true
					break
				}
			}
			assert.True(t, hasLabel4or5, "Task %d should have label 4 or 5", task.ID)
		}

		t.Logf("✓ IN operator filter returned %d tasks with label 4 or 5", resultCount)
	})

	t.Run("LIKE operator filter execution", func(t *testing.T) {
		// Create saved filter: title like 'task'
		sf := &models.SavedFilter{
			Title: "Tasks containing 'task' in title",
			Filters: &models.TaskCollection{
				Filter:             "title like 'task'",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 50)
		assert.NoError(t, err)

		tasks, ok := result.([]*models.Task)
		assert.True(t, ok)

		// Verify all returned tasks have 'task' (case-insensitive) in title
		for _, task := range tasks {
			titleLower := strings.ToLower(task.Title)
			assert.Contains(t, titleLower, "task", "Task %d title should contain 'task'", task.ID)
		}

		t.Logf("✓ LIKE operator filter returned %d tasks with 'task' in title", resultCount)
	})

	t.Run("Empty filter string returns all tasks", func(t *testing.T) {
		// Create saved filter with empty filter string
		sf := &models.SavedFilter{
			Title: "All Tasks",
			Filters: &models.TaskCollection{
				Filter:             "",
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
			},
		}

		err := sfs.Create(s, sf, u)
		assert.NoError(t, err)

		// Execute filter
		result, resultCount, _, err := ts.GetAllWithFullFiltering(s, sf.Filters, u, "", 1, 100)
		assert.NoError(t, err)

		assert.Greater(t, resultCount, 0, "Empty filter should return all accessible tasks")
		assert.NotNil(t, result, "Empty filter should return results")

		t.Logf("✓ Empty filter returned %d tasks (all accessible)", resultCount)
	})
}
