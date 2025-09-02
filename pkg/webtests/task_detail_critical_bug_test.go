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

// TestTaskDetailView_Critical_Bug_EmptyData reproduces the critical bug where
// GET /api/v1/tasks/:id returns a "bare" task object without its related data
// (comments, labels, assignees, reactions, etc.) that the frontend expects.
// This test MUST FAIL before the bug is fixed.
func TestTaskDetailView_Critical_Bug_EmptyData(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Test getting task ID 1 (should exist in fixtures)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Execute the request through the full Echo server
	e.ServeHTTP(rec, req)

	// The request should succeed (200 OK)
	assert.Equal(t, http.StatusOK, rec.Code, "GET /api/v1/tasks/:id should return 200 OK")

	// Parse the response
	var task models.Task
	err = json.Unmarshal(rec.Body.Bytes(), &task)
	require.NoError(t, err, "Response should be valid JSON task object")

	t.Logf("Task ID: %d", task.ID)
	t.Logf("Task Title: %s", task.Title)
	t.Logf("Task Project ID: %d", task.ProjectID)
	t.Logf("Task Assignees count: %d", len(task.Assignees))
	t.Logf("Task Labels count: %d", len(task.Labels))
	t.Logf("Task Attachments count: %d", len(task.Attachments))
	t.Logf("Task Related Tasks count: %d", len(task.RelatedTasks))
	t.Logf("Task CreatedBy: %+v", task.CreatedBy)
	t.Logf("Task Identifier: %s", task.Identifier)
	t.Logf("Task IsFavorite: %t", task.IsFavorite)

	// Basic task data should be present
	assert.NotZero(t, task.ID, "Task should have an ID")
	assert.NotEmpty(t, task.Title, "Task should have a title")
	assert.NotZero(t, task.ProjectID, "Task should have a project ID")

	// BUG: The following fields should be populated by AddDetailsToTasks but are likely empty
	// When the bug is fixed, these assertions should pass.
	// For now, we assert they are empty to confirm the bug exists.

	// The bug: These should be populated but aren't
	t.Log("=== CHECKING FOR BUG: Empty related data ===")

	// This test should initially FAIL when we assert for populated data
	// because the bug makes these fields empty/null
	if task.CreatedBy == nil || task.CreatedBy.ID == 0 {
		t.Log("BUG CONFIRMED: CreatedBy is not populated")
	} else {
		t.Log("CreatedBy IS populated - possible bug fix or test environment difference")
	}

	if task.Identifier == "" {
		t.Log("BUG CONFIRMED: Identifier is not populated")
	} else {
		t.Log("Identifier IS populated:", task.Identifier)
	}

	// For tasks that should have assignees/labels in fixtures, check if they're missing
	// From fixture data analysis, task 1 should likely have some related data
	expectedToHaveAssignees := false   // We'll need to check fixtures to know this
	expectedToHaveLabels := false      // We'll need to check fixtures to know this
	expectedToHaveAttachments := false // We'll need to check fixtures to know this

	if expectedToHaveAssignees && len(task.Assignees) == 0 {
		t.Log("BUG CONFIRMED: Expected assignees but got none")
	}
	if expectedToHaveLabels && len(task.Labels) == 0 {
		t.Log("BUG CONFIRMED: Expected labels but got none")
	}
	if expectedToHaveAttachments && len(task.Attachments) == 0 {
		t.Log("BUG CONFIRMED: Expected attachments but got none")
	}

	// The critical assertion: These should be populated for proper v1 compatibility
	// This test should FAIL until the bug is fixed
	assert.NotNil(t, task.CreatedBy, "CreatedBy should be populated (BUG: currently null)")
	assert.NotEqual(t, 0, task.CreatedBy.ID, "CreatedBy.ID should be populated (BUG: currently 0)")
	assert.NotEmpty(t, task.Identifier, "Task identifier should be populated (BUG: currently empty)")
}

// TestTaskDetailView_ExpectedFields verifies what fields we expect to be populated
// in a fully enriched task object from the v1 API for frontend compatibility.
func TestTaskDetailView_ExpectedFields(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Test getting task ID 1 (should exist in fixtures)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Execute the request
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response
	var task models.Task
	err = json.Unmarshal(rec.Body.Bytes(), &task)
	require.NoError(t, err)

	// Log the complete response structure for debugging
	t.Logf("=== COMPLETE TASK RESPONSE STRUCTURE ===")
	responseBytes, _ := json.MarshalIndent(task, "", "  ")
	t.Logf("Full task JSON:\n%s", string(responseBytes))

	// What the frontend expects from v1 API:
	// 1. Basic task fields (ID, Title, Description, etc.)
	// 2. CreatedBy user object
	// 3. Task identifier (project.identifier + "-" + task.index or "#" + task.index)
	// 4. Assignees array (even if empty)
	// 5. Labels array (even if empty)
	// 6. Attachments array (even if empty)
	// 7. RelatedTasks map (even if empty)
	// 8. IsFavorite boolean
	// 9. Reminders array (even if empty)

	// Assert the minimum required structure for v1 compatibility
	assert.NotNil(t, task.Assignees, "Assignees should be initialized (not null)")
	assert.NotNil(t, task.Labels, "Labels should be initialized (not null)")
	assert.NotNil(t, task.Attachments, "Attachments should be initialized (not null)")
	assert.NotNil(t, task.RelatedTasks, "RelatedTasks should be initialized (not null)")
	assert.NotNil(t, task.Reminders, "Reminders should be initialized (not null)")

	// These are the critical fields for the bug
	assert.NotNil(t, task.CreatedBy, "CreatedBy should be populated")
	if task.CreatedBy != nil {
		assert.NotEqual(t, 0, task.CreatedBy.ID, "CreatedBy.ID should be populated")
		assert.NotEmpty(t, task.CreatedBy.Username, "CreatedBy.Username should be populated")
	}
	assert.NotEmpty(t, task.Identifier, "Task identifier should be populated")
}

// TestTaskDetailView_CompareWithTaskCollection tests if individual task retrieval
// provides the same level of data enrichment as task collections.
func TestTaskDetailView_CompareWithTaskCollection(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Get the same task through the collection endpoint
	collectionReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/views/1/tasks", nil)
	collectionReq.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	collectionReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	collectionRec := httptest.NewRecorder()

	e.ServeHTTP(collectionRec, collectionReq)
	assert.Equal(t, http.StatusOK, collectionRec.Code)

	var collectionTasks []*models.Task
	err = json.Unmarshal(collectionRec.Body.Bytes(), &collectionTasks)
	require.NoError(t, err)
	require.NotEmpty(t, collectionTasks, "Collection should return tasks")

	// Find task ID 1 in the collection
	var collectionTask *models.Task
	for _, task := range collectionTasks {
		if task.ID == 1 {
			collectionTask = task
			break
		}
	}
	require.NotNil(t, collectionTask, "Task ID 1 should be in the collection")

	// Get the same task through the detail endpoint
	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	detailReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	detailRec := httptest.NewRecorder()

	e.ServeHTTP(detailRec, detailReq)
	assert.Equal(t, http.StatusOK, detailRec.Code)

	var detailTask models.Task
	err = json.Unmarshal(detailRec.Body.Bytes(), &detailTask)
	require.NoError(t, err)

	// Compare the enrichment levels - they should be the same
	t.Logf("=== COLLECTION TASK vs DETAIL TASK COMPARISON ===")
	t.Logf("Collection CreatedBy: %+v", collectionTask.CreatedBy)
	t.Logf("Detail CreatedBy: %+v", detailTask.CreatedBy)
	t.Logf("Collection Identifier: %s", collectionTask.Identifier)
	t.Logf("Detail Identifier: %s", detailTask.Identifier)
	t.Logf("Collection Assignees count: %d", len(collectionTask.Assignees))
	t.Logf("Detail Assignees count: %d", len(detailTask.Assignees))

	// Both should have the same level of enrichment
	assert.Equal(t, collectionTask.CreatedBy != nil, detailTask.CreatedBy != nil, "CreatedBy enrichment should match")
	assert.Equal(t, collectionTask.Identifier, detailTask.Identifier, "Identifier should match")
	assert.Equal(t, len(collectionTask.Assignees), len(detailTask.Assignees), "Assignees count should match")
	assert.Equal(t, len(collectionTask.Labels), len(detailTask.Labels), "Labels count should match")
}
