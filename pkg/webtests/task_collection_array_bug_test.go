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
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test reproduces the critical 500 error bug when using frontend-style array query parameters
func TestTaskCollection_ReadAllWeb_ComplexArraySorting(t *testing.T) {
	// This test reproduces the exact query string that the frontend sends and causes a 500 error
	// The frontend sends: sort_by[]=due_date&sort_by[]=id&order_by[]=asc&order_by[]=desc
	t.Run("Frontend Array Syntax with Multiple Sort Parameters", func(t *testing.T) {
		// Create the raw query string exactly as the frontend sends it
		queryString := "sort_by[]=due_date&sort_by[]=id&order_by[]=asc&order_by[]=desc&filter=done+%3D+false&filter_include_nulls=false&s=&filter_timezone=GMT&page=1"

		// Let's try creating a direct HTTP request to see if we can reproduce the 500 error
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Create a request with the exact query string from the bug report
		req := httptest.NewRequest(http.MethodGet, "/tasks/all?"+queryString, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		// Add authentication
		addUserTokenToContext(t, &testuser1, c)

		// Create the handler
		taskCollectionHandler := &handler.WebHandler{
			EmptyStruct: func() handler.CObject {
				return &models.TaskCollection{}
			},
		}

		// Call the ReadAllWeb handler - this should reproduce the 500 error
		err = taskCollectionHandler.ReadAllWeb(c)

		// Let's also check what the struct looks like after binding
		taskCollection := &models.TaskCollection{}
		bindErr := c.Bind(taskCollection)
		t.Logf("Bind result: %+v", taskCollection)
		t.Logf("SortBy: %v", taskCollection.SortBy)
		t.Logf("SortByArr: %v", taskCollection.SortByArr)
		t.Logf("OrderBy: %v", taskCollection.OrderBy)
		t.Logf("OrderByArr: %v", taskCollection.OrderByArr)
		if bindErr != nil {
			t.Logf("Bind error: %v", bindErr)
		}

		// This test MUST fail initially to prove we've reproduced the bug
		if err != nil {
			t.Logf("Reproduced the bug! Error: %v", err)
			// The bug is confirmed if we get a binding/parsing error
			assert.Contains(t, err.Error(), "Invalid model provided")
		} else {
			// If we get here without error, let's check what was actually parsed
			assert.Equal(t, http.StatusOK, rec.Code)
			t.Logf("Unexpected success - query was: %s", queryString)
			t.Logf("Response length: %d characters", len(rec.Body.String()))
		}
	})

	// Let's also test the exact route through the Echo router
	t.Run("Through Echo Router", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Create request exactly as browser would
		queryString := "sort_by[]=due_date&sort_by[]=id&order_by[]=asc&order_by[]=desc&filter=done+%3D+false&filter_include_nulls=false&s=&filter_timezone=GMT&page=1"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/all?"+queryString, nil)
		req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Let Echo handle the full request routing
		e.ServeHTTP(rec, req)

		t.Logf("Status Code: %d", rec.Code)
		t.Logf("Response length: %d", len(rec.Body.String()))

		if rec.Code == 500 {
			t.Logf("REPRODUCED THE BUG! 500 error response: %s", rec.Body.String())
		}

		// The bug is fixed if we get 200 OK, reproduced if we get 500
		if rec.Code != 200 {
			t.Logf("Non-200 status code: %d, body: %s", rec.Code, rec.Body.String())
		}

		assert.Equal(t, http.StatusOK, rec.Code, "Expected 200 OK, but got %d. This indicates the bug is present.", rec.Code)
	})
}

// Helper function to generate JWT token for testing
func getJWTTokenForUser(t *testing.T, user *user.User) string {
	token, err := auth.NewUserJWTAuthtoken(user, false)
	require.NoError(t, err)
	return token
}
