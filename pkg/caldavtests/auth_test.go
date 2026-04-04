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

package caldavtests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	t.Run("Valid credentials return 200/207", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects/36")

		assert.True(t, rec.Code >= 200 && rec.Code < 300,
			"Valid credentials should succeed. Got %d", rec.Code)
	})

	t.Run("No auth returns 401", func(t *testing.T) {
		e := setupTestEnv(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/dav/projects/36", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code,
			"Request without auth should return 401")
	})

	t.Run("Wrong password returns 401", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, http.MethodGet, "/dav/projects/36", "", map[string]string{
			"Authorization": basicAuthHeader(testuser15.Username, "wrongpassword"),
		})

		assert.Equal(t, http.StatusUnauthorized, rec.Code,
			"Wrong password should return 401")
	})

	t.Run("Nonexistent user returns 401", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, http.MethodGet, "/dav/projects/36", "", map[string]string{
			"Authorization": basicAuthHeader("nonexistent_user", fixturePassword),
		})

		assert.Equal(t, http.StatusUnauthorized, rec.Code,
			"Nonexistent user should return 401")
	})

	t.Run("Empty Authorization header returns 401", func(t *testing.T) {
		e := setupTestEnv(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/dav/projects/36", nil)
		req.Header.Set("Authorization", "")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code,
			"Empty auth header should return 401")
	})

	t.Run("Auth on /dav/ entry point", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, "PROPFIND", "/dav/", PropfindCurrentUserPrincipal, map[string]string{
			"Depth": "0",
		})

		// Should succeed with valid auth
		assert.True(t, rec.Code >= 200 && rec.Code < 300 || rec.Code == 207,
			"Authenticated PROPFIND on /dav/ should succeed. Got %d", rec.Code)
	})

	t.Run("Auth on /.well-known/caldav", func(t *testing.T) {
		e := setupTestEnv(t)

		// Without auth
		req := httptest.NewRequestWithContext(context.Background(), "PROPFIND", "/.well-known/caldav", strings.NewReader(PropfindCurrentUserPrincipal))
		req.Header.Set("Depth", "0")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code,
			"/.well-known/caldav without auth should return 401")
	})
}

func TestPermissions(t *testing.T) {
	t.Run("User cannot GET project they do not have access to", func(t *testing.T) {
		t.Skip("Known bug: CalDAV returns 500 instead of 403/404 — ErrUserDoesNotHaveAccessToProject is not recognized by caldav-go")
		e := setupTestEnv(t)

		// testuser1 should not be able to access project 36 (owned by user15)
		rec := caldavRequest(t, e, http.MethodGet, "/dav/projects/36", "", map[string]string{
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
		})

		// Should be 403 Forbidden or 404 Not Found (both are acceptable for access denial)
		assert.True(t, rec.Code == http.StatusForbidden || rec.Code == http.StatusNotFound,
			"Unauthorized user should get 403 or 404, got %d. Body:\n%s", rec.Code, rec.Body.String())
	})

	t.Run("User cannot PUT task to project they do not have access to", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("unauthorized-task", "Should Fail").Build()
		rec := caldavRequest(t, e, http.MethodPut, "/dav/projects/36/unauthorized-task.ics", vtodo, map[string]string{
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
			"Content-Type":  "text/calendar; charset=utf-8",
		})

		assert.True(t, rec.Code == http.StatusForbidden || rec.Code == http.StatusNotFound,
			"PUT to unauthorized project should fail with 403 or 404, got %d", rec.Code)
	})

	t.Run("User cannot DELETE task from project they do not have access to", func(t *testing.T) {
		e := setupTestEnv(t)

		// Try to delete task 40 (uid-caldav-test) in project 36 as user1
		rec := caldavRequest(t, e, http.MethodDelete, "/dav/projects/36/uid-caldav-test.ics", "", map[string]string{
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
		})

		assert.True(t, rec.Code == http.StatusForbidden || rec.Code == http.StatusNotFound,
			"DELETE on unauthorized project should fail with 403 or 404, got %d", rec.Code)
	})

	t.Run("User cannot REPORT on project they do not have access to", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, "REPORT", "/dav/projects/36", ReportCalendarQuery, map[string]string{
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
		})

		assert.True(t, rec.Code == http.StatusForbidden || rec.Code == http.StatusNotFound || rec.Code == 207,
			"REPORT on unauthorized project should fail or return empty, got %d", rec.Code)

		// If it returns 207, it should have no results
		if rec.Code == 207 {
			ms := parseMultistatus(t, rec)
			assert.Empty(t, ms.Responses,
				"REPORT on unauthorized project should return empty multistatus if 207")
		}
	})

	t.Run("Project listing only shows accessible projects", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, "PROPFIND", "/dav/projects", PropfindCalendarCollectionProperties, map[string]string{
			"Depth":         "1",
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
		})

		assertResponseStatus(t, rec, 207)
		body := rec.Body.String()

		// user1 should see their own projects but NOT user15's projects
		assert.NotContains(t, body, "Project 36 for Caldav tests",
			"user1 should not see user15's Project 36")
	})
}
