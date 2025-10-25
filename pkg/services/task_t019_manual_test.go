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

//go:build manual_test
// +build manual_test

package services

// T019 Manual Test - Saved Filter Bug Verification
//
// PURPOSE:
// This test file was created during T019 bug investigation to test against the
// actual production database at ./tmp/vikunja.db. It validates that saved filters
// work correctly with real data that exhibited the bug.
//
// WHEN TO USE:
// - Manual verification against production database after fixes
// - Reproducing the bug with real-world data
// - Validating that fix works with actual frontend-created saved filters
//
// HOW TO RUN:
// 1. Ensure you have a database at ./tmp/vikunja.db (or modify dbPath below)
// 2. Run: VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test -v -tags manual_test -run TestTaskService_T019_RealDatabase ./pkg/services/
//
// IMPORTANT NOTES:
// - This test uses ABSOLUTE PATH: /home/aron/projects/vikunja/tmp/vikunja.db
// - Change this path if running on different machine or directory
// - This test is excluded from normal test runs (requires -tags manual_test)
// - See automated tests in task_test.go for standard test coverage
//
// RELATED:
// - T019 fix: Set AllowNullCheck: false for subtable filters
// - Automated tests: TestTaskService_SavedFilter_WithView_T019
// - Documentation: specs/007-fix-saved-filters/T027-TEST-FINDINGS.md
//
// STATUS: Kept for historical reference and manual validation capability

import (
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// TestTaskService_T019_RealDatabase tests saved filter with the actual production database
// This test is meant to be run manually against the real database to verify the fix
//
// Run with: cd /home/aron/projects/vikunja && VIKUNJA_SERVICE_ROOTPATH=/home/aron/projects/vikunja go test -v -tags manual_test -run TestTaskService_T019_RealDatabase ./pkg/services/
func TestTaskService_T019_RealDatabase(t *testing.T) {
	// Connect to the real database (use absolute path)
	dbPath := "/home/aron/projects/vikunja/tmp/vikunja.db"
	engine, err := xorm.NewEngine("sqlite3", dbPath)
	require.NoError(t, err, "Failed to connect to database at "+dbPath)
	defer engine.Close()

	s := engine.NewSession()
	defer s.Close()

	ts := NewTaskService(engine)
	u := &user.User{ID: 1}

	t.Run("Test saved filter with real database", func(t *testing.T) {
		// The real database has:
		// - Saved filter ID 1 (maps to project -2)
		// - Filter: "done = false && labels = 6"
		// - View ID 21 for this saved filter
		// - Expected result: Only task 22 (the only task with done=false and label 6)

		projectID := int64(-2) // Saved filter ID 1
		viewID := int64(21)

		// First, verify the saved filter exists and has the correct filter
		var savedFilter models.SavedFilter
		has, err := s.ID(1).Get(&savedFilter)
		require.NoError(t, err)
		require.True(t, has, "Saved filter ID 1 should exist")
		t.Logf("Saved filter: ID=%d, Title='%s'", savedFilter.ID, savedFilter.Title)
		if savedFilter.Filters != nil {
			t.Logf("Filter expression: '%s'", savedFilter.Filters.Filter)
		}

		// Test WITHOUT view (baseline)
		t.Run("Without view ID", func(t *testing.T) {
			collection := &models.TaskCollection{
				ProjectID:          projectID,
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
				Expand:             []models.TaskCollectionExpandable{},
			}

			result, resultCount, totalItems, err := ts.GetAllWithFullFiltering(s, collection, u, "", 1, 50)
			require.NoError(t, err)

			tasks, ok := result.([]*models.Task)
			require.True(t, ok, "Result should be a task array")

			t.Logf("WITHOUT view: returned %d tasks (total: %d)", resultCount, totalItems)
			for i, task := range tasks {
				t.Logf("  Task %d: ID=%d, Title='%s', Done=%v", i+1, task.ID, task.Title, task.Done)
			}

			// Expected: Only task 22 should be returned
			require.Equal(t, 1, resultCount, "Should return exactly 1 task (task 22)")
			if resultCount > 0 {
				require.Equal(t, int64(22), tasks[0].ID, "Should return task ID 22")
			}
		})

		// Test WITH view (reproducing frontend scenario)
		t.Run("WITH view ID 21 (frontend scenario)", func(t *testing.T) {
			collection := &models.TaskCollection{
				ProjectID:          projectID,
				ProjectViewID:      viewID,
				FilterIncludeNulls: false,
				FilterTimezone:     "GMT",
				SortByArr:          []string{"position"},
				OrderByArr:         []string{"asc"},
				Expand:             []models.TaskCollectionExpandable{},
			}

			result, resultCount, totalItems, err := ts.GetAllWithFullFiltering(s, collection, u, "", 1, 50)
			require.NoError(t, err)

			tasks, ok := result.([]*models.Task)
			require.True(t, ok, "Result should be a task array")

			t.Logf("WITH view: returned %d tasks (total: %d)", resultCount, totalItems)
			for i, task := range tasks {
				t.Logf("  Task %d: ID=%d, Title='%s', Done=%v", i+1, task.ID, task.Title, task.Done)
			}

			// This is the T019 bug test: Should return ONLY task 22, not all 26 tasks
			if resultCount != 1 {
				t.Errorf("BUG DETECTED: Expected 1 task, got %d tasks", resultCount)
				t.Logf("This means the saved filter is NOT being applied!")
				for i, task := range tasks {
					t.Logf("  Unexpected task %d: ID=%d, Title='%s'", i+1, task.ID, task.Title)
				}
			} else {
				t.Logf("SUCCESS: Filter correctly applied, returning only 1 task")
			}

			require.Equal(t, 1, resultCount, "Should return exactly 1 task (task 22)")
			if resultCount > 0 {
				require.Equal(t, int64(22), tasks[0].ID, "Should return task ID 22")
			}
		})

		// Also test what ALL tasks would look like (to confirm the bug would be obvious)
		t.Run("Baseline: All tasks for user (no filter)", func(t *testing.T) {
			var allTasks []models.Task
			err := s.Where("1=1").Limit(50).Find(&allTasks)
			require.NoError(t, err)

			t.Logf("BASELINE: Total tasks in database accessible to user: %d", len(allTasks))
			t.Logf("If the bug exists, saved filter would return all %d tasks instead of just 1", len(allTasks))
		})
	})
}
