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

	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ErrorResponse represents the expected JSON error structure for standard errors
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents the expected JSON error structure for validation errors
type ValidationErrorResponse struct {
	Code          int      `json:"code"`
	Message       string   `json:"message"`
	InvalidFields []string `json:"invalid_fields"`
}

// TestErrorResponseFormats tests that error responses are correctly serialized to JSON
// This is critical because the error response format is part of the API contract
func TestErrorResponseFormats(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	// Get auth token for testuser1
	token, err := auth.NewUserJWTAuthtoken(&testuser1, false)
	require.NoError(t, err)

	t.Run("validation error returns invalid_fields in JSON body", func(t *testing.T) {
		// Update a project with empty title - this should trigger validation error
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1", strings.NewReader(`{"title":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Should be 412 Precondition Failed for validation errors
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)

		var errResp ValidationErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &errResp)
		require.NoError(t, err, "Response body: %s", rec.Body.String())

		// Verify the error structure includes invalid_fields
		assert.Equal(t, 2002, errResp.Code, "Expected error code 2002 (ErrCodeInvalidData)")
		require.NotEmpty(t, errResp.InvalidFields, "invalid_fields should not be empty")
		require.GreaterOrEqual(t, len(errResp.InvalidFields), 1, "invalid_fields should have at least one element")
		assert.Contains(t, errResp.InvalidFields[0], "title", "invalid_fields should mention 'title'")
	})

	t.Run("bind error returns 400 with message", func(t *testing.T) {
		// Send malformed JSON
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/1", strings.NewReader(`{invalid json`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found error returns 404 with correct structure", func(t *testing.T) {
		// Try to get a project that doesn't exist
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/99999", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &errResp)
		require.NoError(t, err, "Response body: %s", rec.Body.String())

		// Should have a proper error code
		assert.NotZero(t, errResp.Code, "Error code should be non-zero")
		assert.NotEmpty(t, errResp.Message, "Error message should not be empty")
	})

	t.Run("forbidden error returns 403", func(t *testing.T) {
		// Try to access a project owned by user13 (project 20)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/20", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("domain error returns correct code and message", func(t *testing.T) {
		// Try to create a project with a nonexistent parent
		req := httptest.NewRequest(http.MethodPut, "/api/v1/projects", strings.NewReader(`{"title":"Test","parent_project_id":99999}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Should be 404 for nonexistent parent project
		assert.Equal(t, http.StatusNotFound, rec.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &errResp)
		require.NoError(t, err, "Response body: %s", rec.Body.String())

		// Verify the error has proper structure
		assert.NotZero(t, errResp.Code, "Error code should be non-zero")
		assert.NotEmpty(t, errResp.Message, "Error message should not be empty")
	})

	t.Run("unauthorized request returns 401", func(t *testing.T) {
		// Make request without auth token
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1", nil)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
