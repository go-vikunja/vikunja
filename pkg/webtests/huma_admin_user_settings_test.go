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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumaAdminUserSettings(t *testing.T) {
	const path = "/api/v2/admin/users/1/settings/general"

	setup := func(t *testing.T) (*echo.Echo, string) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		t.Cleanup(license.ResetForTests)
		return e, humaTokenFor(t, promoteToAdmin(t, 2))
	}

	t.Run("replaces settings and audits the change", func(t *testing.T) {
		e, token := setup(t)
		events.ClearDispatchedEvents()
		res := humaRequest(t, e, http.MethodPut, path,
			`{"discoverable_by_name":true,"week_start":3,"overdue_tasks_reminders_time":"10:00"}`, token, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		db.AssertExists(t, "users", map[string]any{"id": 1, "discoverable_by_name": true, "week_start": 3}, false)
		dispatched := events.GetDispatchedEvents((&models.AdminUserSettingsUpdatedEvent{}).Name())
		require.Len(t, dispatched, 1)
		evt := dispatched[0].(*models.AdminUserSettingsUpdatedEvent)
		assert.Equal(t, int64(1), evt.User.ID)
		assert.Equal(t, int64(2), evt.Doer.ID)
		// PUT replaces everything, so unsent fields fall back to their zero value.
		assert.Equal(t, []string{
			"discoverable_by_name",
			"email_reminders_enabled",
			"overdue_tasks_reminders_enabled",
			"overdue_tasks_reminders_time",
			"timezone",
			"week_start",
		}, evt.ChangedFields)
	})

	t.Run("merge-patch changes only the sent fields", func(t *testing.T) {
		e, token := setup(t)
		res := humaRequest(t, e, http.MethodPut, path,
			`{"week_start":3,"overdue_tasks_reminders_time":"10:00","frontend_settings":{"color_schema":"dark"}}`, token, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		res = humaRequest(t, e, http.MethodPatch, path,
			`{"discoverable_by_email":true,"frontend_settings":{"default_view":"kanban"}}`, token, "application/merge-patch+json")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		db.AssertExists(t, "users", map[string]any{"id": 1, "discoverable_by_email": true, "week_start": 3}, false)
		var body struct {
			FrontendSettings map[string]any `json:"frontend_settings"`
			User             struct {
				ID       int64   `json:"id"`
				Username string  `json:"username"`
				Email    *string `json:"email"`
			} `json:"user"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		assert.Equal(t, "dark", body.FrontendSettings["color_schema"])
		assert.Equal(t, "kanban", body.FrontendSettings["default_view"])
		assert.Equal(t, int64(1), body.User.ID)
		assert.Equal(t, "user1", body.User.Username)
		assert.Nil(t, body.User.Email)
	})

	t.Run("default project is checked against the target user", func(t *testing.T) {
		e, token := setup(t)
		// Project 1 belongs to user 1; project 20 is foreign to them.
		res := humaRequest(t, e, http.MethodPut, path,
			`{"overdue_tasks_reminders_time":"09:00","default_project_id":1}`, token, "")
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		res = humaRequest(t, e, http.MethodPut, path,
			`{"overdue_tasks_reminders_time":"09:00","default_project_id":20}`, token, "")
		assert.Equal(t, http.StatusBadRequest, res.Code, res.Body.String())
		db.AssertMissing(t, "users", map[string]any{"id": 1, "default_project_id": 20})
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		e, token := setup(t)
		res := humaRequest(t, e, http.MethodGet, "/api/v2/admin/users/9999999/settings/general", "", token, "")
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}
