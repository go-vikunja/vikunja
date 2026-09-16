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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Guards user-directory access by link shares (GHSA-vfxw-3x8p-2vjr).
func TestHumaUserSearchLinkShareRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	share := &models.LinkSharing{ID: 1, Hash: "test", ProjectID: 1, Permission: models.PermissionRead}
	token, err := auth.NewLinkShareJWTAuthtoken(share)
	require.NoError(t, err)

	for path, desc := range map[string]string{
		"/api/v2/users?q=user2":           "global user search",
		"/api/v2/projects/1/users/search": "project user search",
		"/api/v2/projects/1/users":        "project user listing",
		"/api/v2/projects/1/teams":        "project team listing",
	} {
		t.Run(desc, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, path, "", token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	}
}
