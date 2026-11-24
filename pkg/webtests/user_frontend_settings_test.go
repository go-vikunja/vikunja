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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// frontend_settings must be no value or a JSON object up to models.MaxFrontendSettingsSize.
func TestUserSettingsFrontendSettings(t *testing.T) {
	for _, api := range []struct {
		name          string
		method        string
		path          string
		invalidStatus int
	}{
		{"v1", http.MethodPost, "/api/v1/user/settings/general", http.StatusPreconditionFailed},
		{"v2", http.MethodPut, "/api/v2/user/settings/general", http.StatusUnprocessableEntity},
	} {
		t.Run(api.name, func(t *testing.T) {
			set := func(t *testing.T, frontendSettings string) (int, string) {
				e, err := setupTestEnv()
				require.NoError(t, err)
				body := `{"overdue_tasks_reminders_time":"09:00","frontend_settings":` + frontendSettings + `}`
				rec := apiTokenReq(e, api.method, api.path, userJWT(t, 1), body)
				return rec.Code, rec.Body.String()
			}

			for _, c := range []struct {
				name  string
				value string
			}{
				{"no value", `null`},
				{"an object", `{"color_schema":"dark","nested":{"a":1}}`},
				{"an object at the limit", `{"a":"` + strings.Repeat("a", models.MaxFrontendSettingsSize-8) + `"}`},
			} {
				t.Run("allows "+c.name, func(t *testing.T) {
					code, body := set(t, c.value)
					assert.Equal(t, http.StatusOK, code, "body: %s", body)
				})
			}

			for _, c := range []struct {
				name  string
				value string
			}{
				{"a string", `"dark"`},
				{"an array", `[1,2]`},
				{"a number", `42`},
				{"a boolean", `true`},
				{"an object over the limit", `{"a":"` + strings.Repeat("a", models.MaxFrontendSettingsSize-7) + `"}`},
			} {
				t.Run("rejects "+c.name, func(t *testing.T) {
					code, body := set(t, c.value)
					assert.Equal(t, api.invalidStatus, code, "body: %s", body)
					assert.Contains(t, body, "frontend_settings")
					db.AssertExists(t, "users", map[string]interface{}{"id": 1, "frontend_settings": nil}, false)
				})
			}
		})
	}
}
