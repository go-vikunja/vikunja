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
	"testing"

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserExportDownload(t *testing.T) {
	t.Run("no export file", func(t *testing.T) {
		// Use testuser15 which has no export file (ExportFileID = 0)
		body := `{"password": "12345678"}`
		_, err := newTestRequestWithUser(t, http.MethodPost, apiv1.DownloadUserDataExport, &testuser15, body, nil, nil)
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "No user data export found")
	})

	t.Run("export file metadata exists but physical file does not exist", func(t *testing.T) {
		// Use testuser1 which has export_file_id = 1, and file metadata exists but physical file doesn't exist
		body := `{"password": "12345678"}`
		_, err := newTestRequestWithUser(t, http.MethodPost, apiv1.DownloadUserDataExport, &testuser1, body, nil, nil)
		require.Error(t, err)
		// This should return 404 when the physical file doesn't exist
		assert.Equal(t, http.StatusNotFound, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "No user data export found")
	})
}
