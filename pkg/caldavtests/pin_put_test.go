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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func caldavPUTWith(t *testing.T, e *echo.Echo, path, vcalendar string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	all := map[string]string{"Content-Type": "text/calendar; charset=utf-8"}
	for k, v := range headers {
		all[k] = v
	}
	return caldavRequest(t, e, http.MethodPut, path, vcalendar, all)
}

func assertETagHeader(t *testing.T, rec *httptest.ResponseRecorder, want bool) {
	t.Helper()
	etag, has := rec.Header()["Etag"]
	if want {
		require.True(t, has, "expected an ETag header")
		assert.NotEmpty(t, etag[0])
		return
	}
	assert.False(t, has, "expected no ETag header, got %v", etag)
}

func assertSummary(t *testing.T, e *echo.Echo, path, summary string) {
	t.Helper()
	rec := caldavGET(t, e, path)
	require.Equal(t, http.StatusOK, rec.Code, "GET %s. Body:\n%s", path, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "SUMMARY:"+summary)
}

func TestPinPUTCreatePreconditions(t *testing.T) {
	const uid = "pin-put-create"
	const path = "/dav/projects/36/" + uid + ".ics"

	tests := []struct {
		name    string
		headers map[string]string
		status  int
	}{
		{name: "no headers", status: http.StatusCreated},
		{name: "If-Match etag", headers: map[string]string{"If-Match": `"x"`}, status: http.StatusPreconditionFailed},
		{name: "If-Match star", headers: map[string]string{"If-Match": "*"}, status: http.StatusPreconditionFailed},
		{name: "If-None-Match star", headers: map[string]string{"If-None-Match": "*"}, status: http.StatusCreated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := setupTestEnv(t)

			rec := caldavPUTWith(t, e, path, NewVTodo(uid, "Created").Build(), tc.headers)

			require.Equal(t, tc.status, rec.Code, "Body:\n%s", rec.Body.String())
			created := tc.status == http.StatusCreated
			assertETagHeader(t, rec, created)
			if created {
				assertSummary(t, e, path, "Created")
			} else {
				assert.Equal(t, http.StatusNotFound, caldavGET(t, e, path).Code)
			}
		})
	}
}

func TestPinPUTUpdatePreconditions(t *testing.T) {
	const uid = "pin-put-update"
	const path = "/dav/projects/36/" + uid + ".ics"

	tests := []struct {
		name string
		// headers gets the ETag of the existing resource.
		headers func(etag string) map[string]string
		status  int
	}{
		{
			name:    "no headers",
			headers: func(string) map[string]string { return nil },
			// 201 rather than 200/204 for an update.
			status: http.StatusCreated,
		},
		{
			name:    "matching If-Match",
			headers: func(etag string) map[string]string { return map[string]string{"If-Match": etag} },
			status:  http.StatusCreated,
		},
		{
			name:    "If-Match star",
			headers: func(string) map[string]string { return map[string]string{"If-Match": "*"} },
			status:  http.StatusCreated,
		},
		{
			name:    "stale If-Match",
			headers: func(string) map[string]string { return map[string]string{"If-Match": `"stale"`} },
			status:  http.StatusPreconditionFailed,
		},
		{
			name:    "If-None-Match star",
			headers: func(string) map[string]string { return map[string]string{"If-None-Match": "*"} },
			status:  http.StatusPreconditionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := setupTestEnv(t)
			require.Equal(t, http.StatusCreated, caldavPUT(t, e, path, NewVTodo(uid, "Original").Build()).Code)
			etag := caldavGET(t, e, path).Header().Get("ETag")
			require.NotEmpty(t, etag)

			rec := caldavPUTWith(t, e, path, NewVTodo(uid, "Updated").Build(), tc.headers(etag))

			require.Equal(t, tc.status, rec.Code, "Body:\n%s", rec.Body.String())
			updated := tc.status == http.StatusCreated
			assertETagHeader(t, rec, updated)
			if updated {
				assertSummary(t, e, path, "Updated")
			} else {
				assertSummary(t, e, path, "Original")
			}
		})
	}
}

// ProjectHandler parses the body of a PUT on a collection URL as a task and hands it on.
func TestPinPUTOnCollection(t *testing.T) {
	t.Run("new UID creates the task", func(t *testing.T) {
		for _, path := range []string{"/dav/projects/36", "/dav/projects/36/"} {
			t.Run(path, func(t *testing.T) {
				e := setupTestEnv(t)
				const uid = "pin-put-on-collection"

				rec := caldavPUT(t, e, path, NewVTodo(uid, "Via collection").Build())

				require.Equal(t, http.StatusCreated, rec.Code, "Body:\n%s", rec.Body.String())
				assertETagHeader(t, rec, true)
				assertSummary(t, e, "/dav/projects/36/"+uid+".ics", "Via collection")
			})
		}
	})

	t.Run("existing UID updates the task", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPUT(t, e, "/dav/projects/36", NewVTodo("uid-caldav-test", "Overwritten via collection").Build())

		require.Equal(t, http.StatusCreated, rec.Code, "Body:\n%s", rec.Body.String())
		assertETagHeader(t, rec, true)
		assertSummary(t, e, pinTaskHref, "Overwritten via collection")
	})

	t.Run("home set", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPUT(t, e, "/dav/projects", NewVTodo("pin-put-on-home", "Via home").Build())

		assert.Equal(t, http.StatusNotFound, rec.Code, "Body:\n%s", rec.Body.String())
		assertETagHeader(t, rec, false)
	})
}
