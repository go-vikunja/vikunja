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
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskAttachmentUploadSize(t *testing.T) {
	tests := []struct {
		name           string
		fileSize       int64
		expectedStatus int
		configMaxSize  string
	}{
		{
			name:           "Upload file within 32MB boundary",
			fileSize:       30 * 1024 * 1024, // 30MB
			expectedStatus: http.StatusOK,
			configMaxSize:  "50MB",
		},
		{
			name:           "Upload file above old 32MB limit",
			fileSize:       35 * 1024 * 1024, // 35MB
			expectedStatus: http.StatusOK,
			configMaxSize:  "50MB",
		},
		{
			name:           "Upload file exceeding configured limit",
			fileSize:       55 * 1024 * 1024, // 55MB
			expectedStatus: http.StatusRequestEntityTooLarge,
			configMaxSize:  "50MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set config BEFORE creating Echo instance
			oldMaxSize := config.FilesMaxSize.GetString()
			config.FilesMaxSize.Set(tt.configMaxSize)
			defer config.FilesMaxSize.Set(oldMaxSize)

			// Setup Echo instance with updated config
			e, err := setupTestEnv()
			require.NoError(t, err)

			// Initialize test file fixtures
			files.InitTestFileFixtures(t)

			// Create multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("files", "test.pdf")
			require.NoError(t, err)

			// Write dummy data of specified size
			_, err = io.CopyN(part, bytes.NewReader(make([]byte, tt.fileSize)), tt.fileSize)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/1/attachments", body)
			req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())

			// Add JWT token to request header for authentication
			token, err := auth.NewUserJWTAuthtoken(&testuser1, false)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()

			// Execute request
			e.ServeHTTP(rec, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// If we expect an error, verify the error response includes code and message
			if tt.expectedStatus == http.StatusRequestEntityTooLarge {
				assert.Contains(t, rec.Body.String(), "4013") // Error code
				assert.Contains(t, rec.Body.String(), "uploaded file exceeds")
			}
		})
	}
}
