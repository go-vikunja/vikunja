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

func TestLiveTaskCheck(t *testing.T) {
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

var retrievedTask models.Task
err = json.Unmarshal(rec.Body.Bytes(), &retrievedTask)
require.NoError(t, err)

// Print the actual response for debugging
t.Logf("=== LIVE TASK RESPONSE ===")
t.Logf("Task ID: %d", retrievedTask.ID)
t.Logf("Task Title: %s", retrievedTask.Title)
t.Logf("Assignees: %+v (type: %T)", retrievedTask.Assignees, retrievedTask.Assignees)
t.Logf("Labels: %+v (type: %T)", retrievedTask.Labels, retrievedTask.Labels)
t.Logf("Reminders: %+v (type: %T)", retrievedTask.Reminders, retrievedTask.Reminders)
t.Logf("Reactions: %+v (type: %T)", retrievedTask.Reactions, retrievedTask.Reactions)
t.Logf("CreatedBy: %+v", retrievedTask.CreatedBy)
t.Logf("Identifier: %s", retrievedTask.Identifier)
t.Logf("Raw JSON: %s", rec.Body.String())

// Verify the fields are initialized properly
assert.NotNil(t, retrievedTask.Assignees, "Assignees should not be nil")
assert.NotNil(t, retrievedTask.Labels, "Labels should not be nil")
assert.NotNil(t, retrievedTask.Reminders, "Reminders should not be nil")
assert.NotNil(t, retrievedTask.Reactions, "Reactions should not be nil")
assert.NotZero(t, retrievedTask.CreatedBy.ID, "CreatedBy should be populated")
assert.NotEmpty(t, retrievedTask.Identifier, "Identifier should not be empty")
}
