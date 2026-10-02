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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-fprf-r6rv-xg99: default_project_id must be 0 or a project the user can write to.
func TestUserSettingsDefaultProject(t *testing.T) {
	for _, api := range []struct {
		name   string
		method string
		path   string
	}{
		{"v1", http.MethodPost, "/api/v1/user/settings/general"},
		{"v2", http.MethodPut, "/api/v2/user/settings/general"},
	} {
		t.Run(api.name, func(t *testing.T) {
			set := func(t *testing.T, projectID int64, stored ...int64) int {
				e, err := setupTestEnv()
				require.NoError(t, err)
				if len(stored) > 0 {
					s := db.NewSession()
					defer s.Close()
					_, err = s.Exec("UPDATE users SET default_project_id = ? WHERE id = 1", stored[0])
					require.NoError(t, err)
					require.NoError(t, s.Commit())
				}
				body := `{"overdue_tasks_reminders_time":"09:00","default_project_id":` + strconv.FormatInt(projectID, 10) + `}`
				return apiTokenReq(e, api.method, api.path, userJWT(t, 1), body).Code
			}

			for _, c := range []struct {
				name      string
				projectID int64
			}{
				{"none", 0},
				{"own project", 1},
				{"write share", 10},
			} {
				t.Run("allows "+c.name, func(t *testing.T) {
					assert.Equal(t, http.StatusOK, set(t, c.projectID))
					db.AssertExists(t, "users", map[string]interface{}{"id": 1, "default_project_id": c.projectID}, false)
				})
			}

			for _, c := range []struct {
				name      string
				projectID int64
			}{
				{"foreign project", 20},
				{"read share", 9},
				{"nonexistent project", 9999999},
				{"archived project", 22},
			} {
				t.Run("rejects "+c.name, func(t *testing.T) {
					assert.Equal(t, http.StatusBadRequest, set(t, c.projectID))
					db.AssertMissing(t, "users", map[string]interface{}{"id": 1, "default_project_id": c.projectID})
				})
			}

			t.Run("keeps an unchanged default the user lost access to", func(t *testing.T) {
				assert.Equal(t, http.StatusOK, set(t, 20, 20))
			})
		})
	}
}
