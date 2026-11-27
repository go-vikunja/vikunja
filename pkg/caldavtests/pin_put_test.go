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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		{
			name:   "no headers",
			status: http.StatusCreated,
		},
		{
			name:    "If-Match etag",
			headers: map[string]string{"If-Match": `"x"`},
			status:  http.StatusPreconditionFailed,
		},
		{
			name:    "If-Match star",
			headers: map[string]string{"If-Match": "*"},
			status:  http.StatusPreconditionFailed,
		},
		{
			name:    "If-None-Match star",
			headers: map[string]string{"If-None-Match": "*"},
			status:  http.StatusCreated,
		},
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
		name    string
		headers func(etag string) map[string]string
		status  int
	}{
		{
			name:    "no headers",
			headers: func(string) map[string]string { return nil },
			status:  http.StatusNoContent,
		},
		{
			name:    "matching If-Match",
			headers: func(etag string) map[string]string { return map[string]string{"If-Match": etag} },
			status:  http.StatusNoContent,
		},
		{
			name:    "If-Match star",
			headers: func(string) map[string]string { return map[string]string{"If-Match": "*"} },
			status:  http.StatusNoContent,
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
			updated := tc.status == http.StatusNoContent
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
		for _, path := range []string{
			"/dav/projects/36",
			"/dav/projects/36/",
		} {
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

		require.Equal(t, http.StatusNoContent, rec.Code, "Body:\n%s", rec.Body.String())
		assertETagHeader(t, rec, true)
		assertSummary(t, e, pinTaskHref, "Overwritten via collection")
	})

	t.Run("existing UID with the stored ETag in If-Match", func(t *testing.T) {
		e := setupTestEnv(t)
		etag := caldavGET(t, e, pinTaskHref).Header().Get("ETag")
		require.NotEmpty(t, etag)

		// The body's DTSTAMP has nothing to do with the stored task's ETag.
		body := NewVTodo("uid-caldav-test", "Matched via collection").DtStamp(time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)).Build()
		rec := caldavPUTWith(t, e, "/dav/projects/36", body, map[string]string{"If-Match": etag})

		require.Equal(t, http.StatusNoContent, rec.Code, "Body:\n%s", rec.Body.String())
		assertSummary(t, e, pinTaskHref, "Matched via collection")
	})

	t.Run("body without UID creates a task", func(t *testing.T) {
		e := setupTestEnv(t)

		s := db.NewSession()
		uidless := &models.Task{
			Title:       "Task without uid",
			ProjectID:   36,
			Index:       999,
			CreatedByID: testuser15.ID,
		}
		_, err := s.Insert(uidless)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		body := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//test//EN\r\nBEGIN:VTODO\r\nSUMMARY:Created without uid\r\nEND:VTODO\r\nEND:VCALENDAR\r\n"
		rec := caldavPUT(t, e, "/dav/projects/36", body)

		require.Equal(t, http.StatusCreated, rec.Code, "Body:\n%s", rec.Body.String())
		assertETagHeader(t, rec, true)

		s = db.NewSession()
		defer s.Close()
		stored, err := models.GetTaskByIDSimple(s, uidless.ID)
		require.NoError(t, err)
		assert.Equal(t, "Task without uid", stored.Title)
		created := &models.Task{}
		has, err := s.Where("title = ? AND project_id = ?", "Created without uid", 36).Get(created)
		require.NoError(t, err)
		require.True(t, has)
		assert.NotEmpty(t, created.UID)
	})

	t.Run("home set", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPUT(t, e, "/dav/projects", NewVTodo("pin-put-on-home", "Via home").Build())

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Equal(t, allowHomeSet, rec.Header().Get("Allow"))
		assertETagHeader(t, rec, false)
	})
}
