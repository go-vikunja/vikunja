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
