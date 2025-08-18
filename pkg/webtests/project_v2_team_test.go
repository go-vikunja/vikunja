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

func TestGetProjectTeams(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	rec, err := th.Request(t, "GET", "/api/v2/projects/3/teams", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var teams []*models.TeamWithPermission
	err = json.Unmarshal(rec.Body.Bytes(), &teams)
	require.NoError(t, err)

	assert.Len(t, teams, 1)
	assert.Equal(t, int64(1), teams[0].ID)
	assert.Equal(t, models.PermissionRead, teams[0].Permission)
}

func TestGetProjectTeamsWrongProject(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	_, err := th.Request(t, "GET", "/api/v2/projects/999/teams", nil)
	require.Error(t, err)
}

func TestGetProjectTeamsNoPermission(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	_, err := th.Request(t, "GET", "/api/v2/projects/5/teams", nil)
	require.Error(t, err)
}

func TestAddProjectTeam(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	body := `{"team_id": 2, "permission": 1}`
	rec, err := th.Request(t, "POST", "/api/v2/projects/1/teams", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var tp models.TeamProject
	err = json.Unmarshal(rec.Body.Bytes(), &tp)
	require.NoError(t, err)

	assert.Equal(t, int64(2), tp.TeamID)
	assert.Equal(t, models.PermissionWrite, tp.Permission)
}

func TestUpdateProjectTeam(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	body := `{"permission": 2}`
	rec, err := th.Request(t, "PUT", "/api/v2/projects/3/teams/1", strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var tp models.TeamProject
	err = json.Unmarshal(rec.Body.Bytes(), &tp)
	require.NoError(t, err)

	assert.Equal(t, int64(1), tp.TeamID)
	assert.Equal(t, models.PermissionAdmin, tp.Permission)
}

func TestRemoveProjectTeam(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	rec, err := th.Request(t, "DELETE", "/api/v2/projects/3/teams/1", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
