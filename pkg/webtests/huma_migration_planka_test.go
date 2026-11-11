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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/migration"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	"code.vikunja.io/api/pkg/modules/migration/planka"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePlankaServer accepts the api key "good" and rejects everything else.
func fakePlankaServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"item":{"id":"1","name":"Me"}}`))
	}))
	t.Cleanup(srv.Close)
	// httptest listens on 127.0.0.1
	config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(false) })
	return srv
}

func TestHumaMigrationPlanka(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)
	srv := fakePlankaServer(t)

	t.Run("status - never migrated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/planka/status", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`, "body: %s", rec.Body.String())
	})

	t.Run("no auth url route", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/planka/auth", "", token, "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("migrate rejects missing url", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/planka/migrate", `{"token":"good"}`, token, "")
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, planka.ErrCodeInvalidConfig), "body: %s", rec.Body.String())
	})

	t.Run("migrate rejects missing credentials", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/planka/migrate", `{"url":"`+srv.URL+`","username":"me"}`, token, "")
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, planka.ErrCodeInvalidConfig), "body: %s", rec.Body.String())
	})

	t.Run("migrate rejects bad credentials", func(t *testing.T) {
		events.ClearDispatchedEvents()
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/planka/migrate", `{"url":"`+srv.URL+`","token":"very-secret-token-value"}`, token, "")
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"code":%d`, planka.ErrCodeInvalidCredentials), "body: %s", rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "very-secret-token-value", "the token must not be echoed back")
	})

	t.Run("migrate kicks off the migration", func(t *testing.T) {
		events.ClearDispatchedEvents()
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/planka/migrate", `{"url":"`+srv.URL+`","token":"good"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"Migration was started successfully."`)
		events.AssertDispatched(t, &migrationHandler.MigrationRequestedEvent{})
	})
}

func TestHumaMigrationPlanka_AlreadyRunning(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)
	srv := fakePlankaServer(t)

	s := db.NewSession()
	_, err := s.Insert(&migration.Status{
		UserID:       testuser1.ID,
		MigratorName: "planka",
		StartedAt:    time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	_ = s.Close()

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/planka/migrate", `{"url":"`+srv.URL+`","token":"good"}`, token, "")
	assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
}

func TestHumaMigrationPlanka_Unauthenticated(t *testing.T) {
	e := setupMigrationTestEnv(t)

	t.Run("status", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/planka/status", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("migrate", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/planka/migrate", `{"url":"http://x","token":"x"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

func TestHumaMigrationPlanka_ListedInInfo(t *testing.T) {
	e := setupMigrationTestEnv(t)
	rec := humaRequest(t, e, http.MethodGet, "/api/v2/info", "", "", "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"planka"`)
}
