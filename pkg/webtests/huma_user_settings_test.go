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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All subtests in a Test* func share one env: setupTestEnv rotates the JWT
// secret per call, so a token must be issued from the same env it's used
// against. Where a subtest mutates the user, later subtests account for it.

func TestHumaUserShow(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/user", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	body := rec.Body.String()
	assert.Contains(t, body, `"id":1`)
	assert.Contains(t, body, `"username":"user1"`)
	// Like v1, /user does not disclose the email (GetUserByID strips it); the
	// json:"email,omitempty" tag then drops the field entirely.
	assert.NotContains(t, body, `"email":""`)
	// Computed account facts v1 returned alongside the user object.
	assert.Contains(t, body, `"auth_provider":"local"`)
	assert.Contains(t, body, `"is_local_user":true`)
	assert.Contains(t, body, `"is_admin":false`)
	// The nested settings use the shared models.UserGeneralSettings shape.
	assert.Contains(t, body, `"settings":`)
	assert.Contains(t, body, `"frontend_settings":`)
	assert.Contains(t, body, `"extra_settings_links":`)

	t.Run("Unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestHumaUserChangePassword(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Wrong old password", func(t *testing.T) {
		// CheckUserCredentials → ErrWrongUsernameOrPassword (403).
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/password",
			`{"old_password":"invalid","new_password":"123456789"}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Empty old password", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/password",
			`{"old_password":"","new_password":"123456789"}`, token, "")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("New password too short", func(t *testing.T) {
		// v2 maps govalidator failures (bcrypt_password) to 422, not v1's 412.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/password",
			`{"old_password":"12345678","new_password":"1234567"}`, token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Normal - run last, it changes the password", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/password",
			`{"old_password":"12345678","new_password":"123456789"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "The password was updated successfully.")
	})
}

func TestHumaUserUpdateEmail(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Wrong password", func(t *testing.T) {
		// CheckUserCredentials → ErrWrongUsernameOrPassword (403).
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/email",
			`{"new_email":"new@example.com","password":"invalid"}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Missing new email", func(t *testing.T) {
		// new_email carries valid:"...,required"; v2 maps the failure to 422.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/email",
			`{"password":"12345678"}`, token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Normal", func(t *testing.T) {
		// The mailer is disabled by default in tests, so the change applies immediately.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/email",
			`{"new_email":"new@example.com","password":"12345678"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "Your email address was updated.")
	})
	t.Run("Mailer enabled", func(t *testing.T) {
		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)
		require.NoError(t, keyvalue.Del("email_confirm_sent_1"))

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/email",
			`{"new_email":"confirm@example.com","password":"12345678"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "We sent you an email with a link to confirm your new email address. Your current address stays active until you confirm.")

		show := humaRequest(t, e, http.MethodGet, "/api/v2/user", "", token, "")
		require.Equal(t, http.StatusOK, show.Code, "body: %s", show.Body.String())
		assert.Contains(t, show.Body.String(), `"pending_email":"confirm@example.com"`)
	})
}

func TestHumaUserPendingEmail(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	config.MailerEnabled.Set(true)
	defer config.MailerEnabled.Set(false)

	t.Run("Resend without pending change", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/email/resend", "", token, "")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Cancel without pending change", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/user/settings/email", "", token, "")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("Pending change is exposed, resend is throttled, cancel works", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/email",
			`{"new_email":"pending@example.com","password":"12345678"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/user", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"pending_email":"pending@example.com"`)

		// A resend immediately after the PUT hits the 1-minute cooldown (error code 1036).
		rec = humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/email/resend", "", token, "")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())

		require.NoError(t, keyvalue.Del("email_confirm_sent_1"))
		rec = humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/email/resend", "", token, "")
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/email/resend", "", token, "")
		assert.Equal(t, http.StatusTooManyRequests, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/user/settings/email", "", token, "")
		assert.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/user", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "pending_email")
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/user/settings/email", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/user/settings/email/resend", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

func TestHumaUserUpdateSettings(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Normal", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/general",
			`{"name":"New Name","week_start":1,"overdue_tasks_reminders_time":"10:00","timezone":"Europe/Berlin"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "The settings were updated successfully.")

		// The change is observable through user-show.
		show := humaRequest(t, e, http.MethodGet, "/api/v2/user", "", token, "")
		require.Equal(t, http.StatusOK, show.Code)
		assert.Contains(t, show.Body.String(), `"name":"New Name"`)
	})
	t.Run("Frontend settings round-trip as arbitrary JSON", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/general",
			`{"overdue_tasks_reminders_time":"09:00","frontend_settings":{"color_schema":"dark","nested":{"a":1}}}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		show := humaRequest(t, e, http.MethodGet, "/api/v2/user", "", token, "")
		require.Equal(t, http.StatusOK, show.Code)
		var resp struct {
			Settings struct {
				FrontendSettings map[string]any `json:"frontend_settings"`
			} `json:"settings"`
		}
		require.NoError(t, json.Unmarshal(show.Body.Bytes(), &resp))
		assert.Equal(t, "dark", resp.Settings.FrontendSettings["color_schema"])
	})
	t.Run("Invalid week_start", func(t *testing.T) {
		// week_start carries valid:"range(0|6)"; out of range maps to 422.
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/general",
			`{"week_start":9,"overdue_tasks_reminders_time":"09:00"}`, token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})
}

func TestHumaUserAvatarProvider(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("Get", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/avatar/provider", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"avatar_provider":`)
	})
	t.Run("Set then get reflects the change", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/avatar/provider",
			`{"avatar_provider":"initials"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"avatar_provider":"initials"`)

		get := humaRequest(t, e, http.MethodGet, "/api/v2/user/settings/avatar/provider", "", token, "")
		require.Equal(t, http.StatusOK, get.Code)
		assert.Contains(t, get.Body.String(), `"avatar_provider":"initials"`)
	})
	t.Run("Invalid provider", func(t *testing.T) {
		// UpdateUser rejects unknown providers with ErrInvalidAvatarProvider (412).
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/avatar/provider",
			`{"avatar_provider":"nonsense"}`, token, "")
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
	})
}

func TestHumaUserTimezones(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/timezones", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var zones []string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &zones))
	assert.NotEmpty(t, zones)
	assert.Contains(t, zones, "Europe/Berlin")
}
