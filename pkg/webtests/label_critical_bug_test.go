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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLabelCreation_PUT_Fixed tests that the critical bug where
// PUT /api/v1/labels returned 404 Not Found has been fixed.
// This test should now PASS after the bug fix.
func TestLabelCreation_PUT_Fixed(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Create a test label payload
	labelPayload := map[string]interface{}{
		"title":       "Test Label",
		"description": "A test label for reproduction",
		"hex_color":   "ff0000",
	}
	payloadBytes, err := json.Marshal(labelPayload)
	require.NoError(t, err)

	// Create the request - This is what the frontend is trying to do
	req := httptest.NewRequest(http.MethodPut, "/api/v1/labels", strings.NewReader(string(payloadBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Execute the request through the full Echo server
	e.ServeHTTP(rec, req)

	// The bug: This should succeed (201 Created) but currently returns 404
	// Once fixed, this assertion should be changed to assert.Equal(t, http.StatusCreated, rec.Code)
	t.Log("Current response code:", rec.Code)
	t.Log("Current response body:", rec.Body.String())

	// The bug has been FIXED: This should now succeed (201 Created)
	// Originally this returned 404, but after the fix it should work
	assert.Equal(t, http.StatusCreated, rec.Code, "PUT /api/v1/labels should now work after the fix")

	// Verify the response contains the created label
	var createdLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &createdLabel)
	require.NoError(t, err, "Response should be valid JSON label object")
	assert.Equal(t, "Test Label", createdLabel.Title)
	assert.Equal(t, "ff0000", createdLabel.HexColor)
	assert.NotZero(t, createdLabel.ID, "Created label should have an ID")
}

// TestLabelCreation_POST_Works tests that POST /api/v1/labels works correctly.
// This helps verify that the issue is specifically with the PUT method routing.
func TestLabelCreation_POST_Works(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Create a test label payload
	labelPayload := map[string]interface{}{
		"title":       "Test Label via POST",
		"description": "A test label via POST method",
		"hex_color":   "00ff00",
	}
	payloadBytes, err := json.Marshal(labelPayload)
	require.NoError(t, err)

	// Create the request using POST (which should work)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", strings.NewReader(string(payloadBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Execute the request through the full Echo server
	e.ServeHTTP(rec, req)

	// This should work correctly
	assert.Equal(t, http.StatusCreated, rec.Code, "POST /api/v1/labels should work correctly")

	// Verify the response contains the created label
	var createdLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &createdLabel)
	require.NoError(t, err)
	assert.Equal(t, "Test Label via POST", createdLabel.Title)
	assert.Equal(t, "00ff00", createdLabel.HexColor)
}

// TestLabelUpdate_PUT_WithID_Works verifies that PUT /api/v1/labels/:id works correctly.
// This helps distinguish between the creation bug and update functionality.
func TestLabelUpdate_PUT_WithID_Works(t *testing.T) {
	// Setup test environment
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Create update payload
	updatePayload := map[string]interface{}{
		"title":       "Updated Label Title",
		"description": "Updated description",
		"hex_color":   "0000ff",
	}
	payloadBytes, err := json.Marshal(updatePayload)
	require.NoError(t, err)

	// Create the request to update label ID 1 (should exist in fixtures)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/labels/1", strings.NewReader(string(payloadBytes)))
	req.Header.Set("Authorization", "Bearer "+getJWTTokenForUser(t, &testuser1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Execute the request through the full Echo server
	e.ServeHTTP(rec, req)

	// This should work correctly (update existing label)
	assert.Equal(t, http.StatusOK, rec.Code, "PUT /api/v1/labels/:id should work correctly")

	// Verify the response contains the updated label
	var updatedLabel models.Label
	err = json.Unmarshal(rec.Body.Bytes(), &updatedLabel)
	require.NoError(t, err)
	assert.Equal(t, "Updated Label Title", updatedLabel.Title)
	assert.Equal(t, "0000ff", updatedLabel.HexColor)
}
