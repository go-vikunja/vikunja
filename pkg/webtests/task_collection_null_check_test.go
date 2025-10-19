// Test task collection endpoints that the frontend might use for task lists
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

// TestTaskCollectionEndpoints tests the endpoints that show task lists
// which might be where the frontend is seeing empty/null values
func TestTaskCollectionEndpoints(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Test common task collection endpoints that frontend uses
	endpoints := []struct {
		name        string
		url         string
		expectTasks bool
	}{
		{"AllTasks", "/api/v1/tasks/all", true},
		{"ProjectTasks", "/api/v1/projects/1/views/1/tasks", true},
		{"ProjectTasksDefault", "/api/v1/projects/1/tasks", true},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint.url, nil)
			req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			t.Logf("%s - Status: %d", endpoint.name, rec.Code)

			if rec.Code == 200 {
				var tasks []*models.Task
				err = json.Unmarshal(rec.Body.Bytes(), &tasks)
				require.NoError(t, err, "%s should return valid JSON array", endpoint.name)

				if endpoint.expectTasks {
					assert.NotEmpty(t, tasks, "%s should return some tasks", endpoint.name)
				}

				// Check the first task for null value issues
				if len(tasks) > 0 {
					task := tasks[0]
					t.Logf("%s - First task: %s", endpoint.name, task.Title)
					t.Logf("%s - Assignees: %v (nil: %v)", endpoint.name, len(task.Assignees), task.Assignees == nil)
					t.Logf("%s - Labels: %v (nil: %v)", endpoint.name, len(task.Labels), task.Labels == nil)
					t.Logf("%s - Reminders: %v (nil: %v)", endpoint.name, len(task.Reminders), task.Reminders == nil)
					t.Logf("%s - Reactions: %v (nil: %v)", endpoint.name, len(task.Reactions), task.Reactions == nil)
					t.Logf("%s - CreatedBy: %v", endpoint.name, task.CreatedBy != nil)
					t.Logf("%s - Identifier: %s", endpoint.name, task.Identifier)

					// These should NOT be null after our fix
					assert.NotNil(t, task.Assignees, "%s - assignees should not be nil", endpoint.name)
					assert.NotNil(t, task.Labels, "%s - labels should not be nil", endpoint.name)
					assert.NotNil(t, task.Reminders, "%s - reminders should not be nil", endpoint.name)
					assert.NotNil(t, task.Reactions, "%s - reactions should not be nil", endpoint.name)
					assert.NotNil(t, task.CreatedBy, "%s - created_by should not be nil", endpoint.name)
					assert.NotEmpty(t, task.Identifier, "%s - identifier should not be empty", endpoint.name)
				}
			} else {
				t.Logf("%s - Error response: %s", endpoint.name, rec.Body.String())
			}
		})
	}
}
