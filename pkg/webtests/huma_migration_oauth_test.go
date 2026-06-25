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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/migration"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMigrationTestEnv builds a test env with the OAuth migrators enabled so
// their v2 routes are registered (they are gated behind config flags that
// default to false). setupTestEnv resets config to defaults, so the flags must
// be set after it and the router rebuilt.
func setupMigrationTestEnv(t *testing.T) *echo.Echo {
	t.Helper()
	_, err := setupTestEnv()
	require.NoError(t, err)

	// migration.Status is not part of models.GetTables() (pkg/models cannot
	// import pkg/modules/migration without a cycle), so SetupTests never syncs
	// migration_status. Create it here so the status/migrate handlers can query.
	s := db.NewSession()
	require.NoError(t, s.Sync2(&migration.Status{}))
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())

	config.MigrationTodoistEnable.Set(true)
	config.MigrationTrelloEnable.Set(true)
	config.MigrationMicrosoftTodoEnable.Set(true)
	t.Cleanup(func() {
		config.MigrationTodoistEnable.Set(false)
		config.MigrationTrelloEnable.Set(false)
		config.MigrationMicrosoftTodoEnable.Set(false)
	})

	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	return e
}

// TestHumaMigrationOAuth covers the three OAuth migrators' v2 endpoints. There
// is no v1 webtest for these handlers to mirror, so this is the parity baseline.
func TestHumaMigrationOAuth(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	// The generic registration helper wires the same three ops for every
	// migrator, so exercising each name guards against a copy-paste regression.
	for _, name := range []string{"todoist", "trello", "microsoft-todo"} {
		t.Run(name+" auth url", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/"+name+"/auth", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"url":"http`, "auth url must be returned; body: %s", rec.Body.String())
		})

		t.Run(name+" status - never migrated", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/"+name+"/status", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			// A user who never migrated has a zero-value status.
			assert.Contains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`, "body: %s", rec.Body.String())
		})
	}

	t.Run("migrate kicks off the migration", func(t *testing.T) {
		events.ClearDispatchedEvents()
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/todoist/migrate", `{"code":"test-code"}`, token, "")
		// 200, not the wrapper's POST default 201: this queues a job, it does
		// not create a REST resource.
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"Migration was started successfully."`)
		events.AssertDispatched(t, &migrationHandler.MigrationRequestedEvent{})
	})
}

// TestHumaMigrationOAuth_AlreadyRunning ports v1's guard: starting a migration
// while one is already in progress (started, not finished) is refused with 412.
func TestHumaMigrationOAuth_AlreadyRunning(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	s := db.NewSession()
	_, err := s.Insert(&migration.Status{
		UserID:       testuser1.ID,
		MigratorName: "todoist",
		StartedAt:    time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	_ = s.Close()

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/todoist/migrate", `{"code":"test-code"}`, token, "")
	assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
}

// TestHumaMigrationOAuth_Unauthenticated proves all three ops require auth.
func TestHumaMigrationOAuth_Unauthenticated(t *testing.T) {
	e := setupMigrationTestEnv(t)

	t.Run("auth", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/todoist/auth", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("status", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/todoist/status", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
	t.Run("migrate", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/todoist/migrate", `{"code":"x"}`, "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}

// TestHumaMigrationOAuth_Disabled proves a migrator's routes are absent when its
// config flag is off.
func TestHumaMigrationOAuth_Disabled(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	// All migration flags default to false after InitDefaultConfig.

	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/todoist/auth", "", token, "")
	assert.Equal(t, http.StatusNotFound, rec.Code,
		"migration routes must not be registered when the flag is off; body: %s", rec.Body.String())
}
