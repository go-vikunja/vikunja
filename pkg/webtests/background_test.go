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

	bgHandler "code.vikunja.io/api/pkg/modules/background/handler"

	"github.com/stretchr/testify/assert"
)

func TestProjectBackgroundDeletePermission(t *testing.T) {
	t.Run("Read-only user cannot delete project background", func(t *testing.T) {
		// testuser15 has read-only access (permission: 0) to project 35,
		// which has background_file_id: 1.
		// Deleting the background should require write access.
		_, err := newTestRequestWithUser(
			t,
			http.MethodDelete,
			bgHandler.RemoveProjectBackground,
			&testuser15,
			"",
			nil,
			map[string]string{"project": "35"},
		)

		// Should be forbidden for a read-only user
		assert.Error(t, err)
		assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
	})
}
