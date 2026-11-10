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
	"net/http/httptest"
	"net/url"
	"strconv"
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

func setupMigrationTestEnv(t *testing.T) *echo.Echo {
	t.Helper()
	_, err := setupTestEnv()
	require.NoError(t, err)

	// SetupTests cannot own migration_status without an import cycle.
	s := db.NewSession()
	require.NoError(t, s.Sync2(&migration.Status{}))
	_, err = s.Where("1 = 1").Delete(&migration.Status{})
	require.NoError(t, err)
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

func TestHumaMigrationOAuth(t *testing.T) {
	e := setupMigrationTestEnv(t)
	token := humaTokenFor(t, &testuser1)

	for _, name := range []string{"todoist", "trello", "microsoft-todo"} {
		t.Run(name+" auth url", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/"+name+"/auth", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"url":"http`, "auth url must be returned; body: %s", rec.Body.String())
		})

		t.Run(name+" status - never migrated", func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/"+name+"/status", "", token, "")
			require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"started_at":"0001-01-01T00:00:00Z"`, "body: %s", rec.Body.String())
		})
	}

	t.Run("migrate kicks off the migration", func(t *testing.T) {
		events.ClearDispatchedEvents()
		state := issuedOAuthState(t, e, "/api/v2", "todoist", token)
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/todoist/migrate", `{"code":"test-code","state":"`+state+`"}`, token, "")
		// 200, not the wrapper's POST default 201: this queues a job, it does
		// not create a REST resource.
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"message":"Migration was started successfully."`)
		events.AssertDispatched(t, &migrationHandler.MigrationRequestedEvent{})
	})
}

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

	state := issuedOAuthState(t, e, "/api/v2", "todoist", token)
	rec := humaRequest(t, e, http.MethodPost, "/api/v2/migration/todoist/migrate", `{"code":"test-code","state":"`+state+`"}`, token, "")
	assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "body: %s", rec.Body.String())
}

func issuedOAuthState(t *testing.T, e *echo.Echo, prefix, name, token string) string {
	t.Helper()
	rec := humaRequest(t, e, http.MethodGet, prefix+"/migration/"+name+"/auth", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var body struct {
		URL string `json:"url"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	authURL, err := url.Parse(body.URL)
	require.NoError(t, err)
	query := authURL.Query()
	if name == "trello" {
		// trello's implicit flow carries the state in the return url
		returnURL, err := url.Parse(query.Get("return_url"))
		require.NoError(t, err)
		query = returnURL.Query()
	}
	state := query.Get("state")
	require.NotEmpty(t, state, "auth url %s carries no state", body.URL)
	return state
}

// GHSA-xgw3-ccqg-f6ww
func TestMigrationOAuth_State(t *testing.T) {
	for _, prefix := range []string{"/api/v1", "/api/v2"} {
		for _, name := range []string{"todoist", "trello", "microsoft-todo"} {
			migrate := func(t *testing.T, e *echo.Echo, token, state string) *httptest.ResponseRecorder {
				body := `{"code":"test-code"}`
				if state != "" {
					body = `{"code":"test-code","state":"` + state + `"}`
				}
				return humaRequest(t, e, http.MethodPost, prefix+"/migration/"+name+"/migrate", body, token, "")
			}
			assertRejected := func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
				assert.Contains(t, rec.Body.String(), strconv.Itoa(migration.ErrCodeInvalidOAuthState), "body: %s", rec.Body.String())
			}

			t.Run(prefix+" "+name+" without state", func(t *testing.T) {
				e := setupMigrationTestEnv(t)
				assertRejected(t, migrate(t, e, humaTokenFor(t, &testuser1), ""))
			})
			t.Run(prefix+" "+name+" with another user's state", func(t *testing.T) {
				e := setupMigrationTestEnv(t)
				state := issuedOAuthState(t, e, prefix, name, humaTokenFor(t, &testuser2))
				assertRejected(t, migrate(t, e, humaTokenFor(t, &testuser1), state))
			})
			t.Run(prefix+" "+name+" with the user's own state, once", func(t *testing.T) {
				e := setupMigrationTestEnv(t)
				token := humaTokenFor(t, &testuser1)
				state := issuedOAuthState(t, e, prefix, name, token)
				rec := migrate(t, e, token, state)
				require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
				assertRejected(t, migrate(t, e, token, state))
			})
		}
	}
}

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

func TestHumaMigrationOAuth_Disabled(t *testing.T) {
	_, err := setupTestEnv()
	require.NoError(t, err)
	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/migration/todoist/auth", "", token, "")
	assert.Equal(t, http.StatusNotFound, rec.Code,
		"migration routes must not be registered when the flag is off; body: %s", rec.Body.String())
}
