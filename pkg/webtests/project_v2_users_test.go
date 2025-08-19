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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectUsers(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	rec, err := th.Request(t, "GET", "/api/v2/projects/3/users", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var users []*models.UserWithPermission
	err = json.Unmarshal(rec.Body.Bytes(), &users)
	require.NoError(t, err)

	assert.Len(t, users, 2)
	assert.Equal(t, int64(1), users[0].ID)
	assert.Equal(t, models.PermissionAdmin, users[0].Permission)
	assert.Equal(t, int64(2), users[1].ID)
	assert.Equal(t, models.PermissionRead, users[1].Permission)
}

func TestAddProjectUser(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	body := `{"username": "user3", "permission": 1}`
	rec, err := th.Request(t, "POST", "/api/v2/projects/1/users", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var pu models.ProjectUser
	err = json.Unmarshal(rec.Body.Bytes(), &pu)
	require.NoError(t, err)

	// The user id is not in the response, so we can't check it.
	// assert.Equal(t, int64(3), pu.UserID)
	assert.Equal(t, models.PermissionWrite, pu.Permission)
}

func TestUpdateProjectUser(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	body := `{"permission": 2}`
	rec, err := th.Request(t, "PUT", "/api/v2/projects/3/users/2", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pu models.ProjectUser
	err = json.Unmarshal(rec.Body.Bytes(), &pu)
	require.NoError(t, err)

	// The user id is not in the response, so we can't check it.
	// assert.Equal(t, int64(2), pu.UserID)
	assert.Equal(t, models.PermissionAdmin, pu.Permission)
}

func TestRemoveProjectUser(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	rec, err := th.Request(t, "DELETE", "/api/v2/projects/3/users/2", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
