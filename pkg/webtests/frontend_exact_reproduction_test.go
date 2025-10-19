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

package webtests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFrontendExactCalls tests the exact API calls that the frontend makes
// to ensure we're not missing any edge cases that could cause empty UI
func TestFrontendExactCalls(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Test various task IDs to see if any show the empty UI issue
	taskIDs := []string{"1", "2", "3", "4", "5"}

	for _, taskID := range taskIDs {
		t.Run("TaskID_"+taskID, func(t *testing.T) {
			// Test the exact endpoint the frontend calls
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
			req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Accept", "application/json")
			rec := httptest.NewRecorder()

			// Execute the request
			e.ServeHTTP(rec, req)

			t.Logf("Task %s - Status: %d", taskID, rec.Code)

			if rec.Code == 200 {
				var task models.Task
				err = json.Unmarshal(rec.Body.Bytes(), &task)
				require.NoError(t, err, "Task %s should return valid JSON", taskID)

				// Log key fields for debugging
				t.Logf("Task %s - Title: %s", taskID, task.Title)
				t.Logf("Task %s - CreatedBy: %v", taskID, task.CreatedBy != nil)
				t.Logf("Task %s - Assignees: %v (nil: %v)", taskID, len(task.Assignees), task.Assignees == nil)
				t.Logf("Task %s - Labels: %v (nil: %v)", taskID, len(task.Labels), task.Labels == nil)
				t.Logf("Task %s - Reminders: %v (nil: %v)", taskID, len(task.Reminders), task.Reminders == nil)
				t.Logf("Task %s - Reactions: %v (nil: %v)", taskID, len(task.Reactions), task.Reactions == nil)
				t.Logf("Task %s - Identifier: %s", taskID, task.Identifier)

				// Check for the critical frontend compatibility issues
				assert.NotNil(t, task.Assignees, "Task %s assignees should not be nil", taskID)
				assert.NotNil(t, task.Labels, "Task %s labels should not be nil", taskID)
				assert.NotNil(t, task.Reminders, "Task %s reminders should not be nil", taskID)
				assert.NotNil(t, task.Reactions, "Task %s reactions should not be nil", taskID)
				assert.NotNil(t, task.CreatedBy, "Task %s created_by should not be nil", taskID)
				assert.NotEmpty(t, task.Identifier, "Task %s identifier should not be empty", taskID)
			} else if rec.Code == 403 {
				t.Logf("Task %s - Forbidden (expected for some tasks)", taskID)
			} else if rec.Code == 404 {
				t.Logf("Task %s - Not Found (task doesn't exist)", taskID)
			} else {
				t.Errorf("Task %s - Unexpected status: %d, Body: %s", taskID, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestFrontendTaskDetailWithComments tests if the issue appears when frontend
// requests task details with comments expanded
func TestFrontendTaskDetailWithComments(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Test with different expand parameters that frontend might use
	expandOptions := []string{
		"",
		"comments",
		"subtasks",
		"reactions",
		"comments,subtasks",
		"comments,reactions",
		"subtasks,reactions",
		"comments,subtasks,reactions",
	}

	for _, expand := range expandOptions {
		t.Run("Expand_"+expand, func(t *testing.T) {
			url := "/api/v1/tasks/1"
			if expand != "" {
				url += "?expand=" + expand
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			t.Logf("Expand '%s' - Status: %d", expand, rec.Code)

			if rec.Code == 200 {
				var task models.Task
				err = json.Unmarshal(rec.Body.Bytes(), &task)
				require.NoError(t, err, "Expand '%s' should return valid JSON", expand)

				// Check if expand parameters affect the null value issue
				assert.NotNil(t, task.Assignees, "Expand '%s' - assignees should not be nil", expand)
				assert.NotNil(t, task.Labels, "Expand '%s' - labels should not be nil", expand)
				assert.NotNil(t, task.Reminders, "Expand '%s' - reminders should not be nil", expand)
				assert.NotNil(t, task.Reactions, "Expand '%s' - reactions should not be nil", expand)
			}
		})
	}
}

// TestTaskDetailDirectModelCall tests if the issue is in the service layer
// by calling the TaskService directly
func TestTaskDetailDirectModelCall(t *testing.T) {
	// This test bypasses the HTTP layer to test the service directly
	t.Skip("TODO: Implement direct service layer test")
}
