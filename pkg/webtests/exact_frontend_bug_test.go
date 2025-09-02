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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExactFrontendBug tests the EXACT URL that's failing in the real frontend
func TestExactFrontendBug(t *testing.T) {
	// This is the EXACT URL that's returning 500 in the real frontend
	queryString := "sort_by[]=due_date&sort_by[]=id&order_by[]=asc&order_by[]=desc&filter=done+%3D+false&filter_include_nulls=false&s=&filter_timezone=GMT&page=1"

	e, err := setupTestEnv()
	require.NoError(t, err)

	// Test the exact request that's failing
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/all?"+queryString, nil)
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	rec := httptest.NewRecorder()

	// This should reproduce the exact 500 error
	e.ServeHTTP(rec, req)

	t.Logf("Status Code: %d", rec.Code)
	t.Logf("Response Body: %s", rec.Body.String())

	if rec.Code == 500 {
		t.Logf("REPRODUCED 500 ERROR! This proves the bug still exists despite our fix")
		t.Logf("Response: %s", rec.Body.String())
	} else if rec.Code != 200 {
		t.Logf("Non-200 status: %d, Response: %s", rec.Code, rec.Body.String())
	} else {
		t.Logf("SUCCESS: Request worked in test environment (200 OK)")
		t.Logf("Response preview: %.500s", rec.Body.String())
	}

	// Let's also test the binding to understand what's happening
	testReq := httptest.NewRequest(http.MethodGet, "/tasks/all?"+queryString, nil)
	testRec := httptest.NewRecorder()
	c := e.NewContext(testReq, testRec)
	addUserTokenToContext(t, &testuser1, c)

	taskCollection := &models.TaskCollection{}
	bindErr := c.Bind(taskCollection)

	t.Logf("Binding Error: %v", bindErr)
	t.Logf("SortBy: %v", taskCollection.SortBy)
	t.Logf("SortByArr: %v", taskCollection.SortByArr)
	t.Logf("OrderBy: %v", taskCollection.OrderBy)
	t.Logf("OrderByArr: %v", taskCollection.OrderByArr)

	// If the test shows 200 but frontend shows 500, there might be a difference
	// between test environment and real environment
	assert.Equal(t, http.StatusOK, rec.Code, "Frontend URL should work")
}
