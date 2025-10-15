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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskLabel_AddLabel tests adding a label to a task via PUT /tasks/:id/labels
func TestTaskLabel_AddLabel(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// First, create a label
	labelPayload := map[string]interface{}{
		"title":     "Test Label for Task",
		"hex_color": "ff0000",
	}
	labelBytes, _ := json.Marshal(labelPayload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", strings.NewReader(string(labelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "Label creation should succeed")

	var createdLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &createdLabel)
	require.NoError(t, err)
	require.NotZero(t, createdLabel.ID)

	// Now add the label to task 1
	addLabelPayload := map[string]interface{}{
		"label_id": createdLabel.ID,
	}
	addLabelBytes, _ := json.Marshal(addLabelPayload)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/tasks/1/labels", strings.NewReader(string(addLabelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	t.Logf("Add label to task response code: %d", rec.Code)
	t.Logf("Add label to task response body: %s", rec.Body.String())

	assert.Equal(t, http.StatusCreated, rec.Code, "Adding label to task should succeed")

	var returnedLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &returnedLabel)
	require.NoError(t, err)
	assert.Equal(t, createdLabel.ID, returnedLabel.ID)
	assert.Equal(t, "Test Label for Task", returnedLabel.Title)
}

// TestTaskLabel_AddLabelWithoutID tests that adding a label without an ID returns an error
func TestTaskLabel_AddLabelWithoutID(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Try to add a label without ID
	addLabelPayload := map[string]interface{}{
		"title": "This should fail",
	}
	addLabelBytes, _ := json.Marshal(addLabelPayload)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/1/labels", strings.NewReader(string(addLabelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Should return bad request when label ID is missing")
}

// TestTaskLabel_RemoveLabel tests removing a label from a task via DELETE /tasks/:id/labels/:labelid
func TestTaskLabel_RemoveLabel(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// First, create and add a label to a task
	labelPayload := map[string]interface{}{
		"title":     "Label to Remove",
		"hex_color": "00ff00",
	}
	labelBytes, _ := json.Marshal(labelPayload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", strings.NewReader(string(labelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var createdLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &createdLabel)
	require.NoError(t, err)

	// Add the label to task 1
	addLabelPayload := map[string]interface{}{
		"label_id": createdLabel.ID,
	}
	addLabelBytes, _ := json.Marshal(addLabelPayload)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/tasks/1/labels", strings.NewReader(string(addLabelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Now remove the label
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tasks/1/labels/%d", createdLabel.ID), nil)
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	t.Logf("Remove label response code: %d", rec.Code)
	t.Logf("Remove label response body: %s", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec.Code, "Removing label from task should succeed")
}

// TestTaskLabel_GetTaskLabels tests getting all labels for a task via GET /tasks/:id/labels
func TestTaskLabel_GetTaskLabels(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Get labels for task 1 (which has labels in fixtures)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/1/labels", nil)
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	t.Logf("Get task labels response code: %d", rec.Code)
	t.Logf("Get task labels response body: %s", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec.Code, "Getting task labels should succeed")

	var labels []*models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &labels)
	require.NoError(t, err)
	assert.NotNil(t, labels, "Should return a labels array")
}

// TestTaskLabel_BulkUpdate tests bulk updating labels on a task via POST /tasks/:id/labels/bulk
func TestTaskLabel_BulkUpdate(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Create two labels
	label1Payload := map[string]interface{}{
		"title":     "Bulk Label 1",
		"hex_color": "ff0000",
	}
	label1Bytes, _ := json.Marshal(label1Payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", strings.NewReader(string(label1Bytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var label1 models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &label1)
	require.NoError(t, err)

	label2Payload := map[string]interface{}{
		"title":     "Bulk Label 2",
		"hex_color": "00ff00",
	}
	label2Bytes, _ := json.Marshal(label2Payload)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/labels", strings.NewReader(string(label2Bytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var label2 models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &label2)
	require.NoError(t, err)

	// Bulk update task 1 with both labels
	bulkPayload := map[string]interface{}{
		"labels": []map[string]interface{}{
			{"id": label1.ID},
			{"id": label2.ID},
		},
	}
	bulkBytes, _ := json.Marshal(bulkPayload)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/tasks/1/labels/bulk", strings.NewReader(string(bulkBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	t.Logf("Bulk update response code: %d", rec.Code)
	t.Logf("Bulk update response body: %s", rec.Body.String())

	assert.Equal(t, http.StatusOK, rec.Code, "Bulk update should succeed")

	var returnedLabels []*models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &returnedLabels)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(returnedLabels), 2, "Should have at least 2 labels")
}

// TestTaskLabel_PermissionDenied tests that users without access cannot add labels
func TestTaskLabel_PermissionDenied(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Create a label as user1
	labelPayload := map[string]interface{}{
		"title":     "User1 Label",
		"hex_color": "ff0000",
	}
	labelBytes, _ := json.Marshal(labelPayload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", strings.NewReader(string(labelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var createdLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &createdLabel)
	require.NoError(t, err)

	// Try to add user1's label to a task that user2 doesn't have access to (task 1 is owned by user1, not shared with user2)
	// But first we need to check if task 1 is accessible to user2 or not from fixtures
	// For this test, we'll try to add a label to task 1 as user2 which should fail
	addLabelPayload := map[string]interface{}{
		"label_id": createdLabel.ID,
	}
	addLabelBytes, _ := json.Marshal(addLabelPayload)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/tasks/1/labels", strings.NewReader(string(addLabelBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser2))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	t.Logf("Permission denied response code: %d", rec.Code)
	t.Logf("Permission denied response body: %s", rec.Body.String())

	// Should return 403 or 404 depending on how permissions are implemented
	assert.Contains(t, []int{http.StatusForbidden, http.StatusNotFound}, rec.Code, "Should deny access to task user doesn't own")
}
