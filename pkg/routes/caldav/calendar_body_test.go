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

package caldav

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	caldav2 "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unparseableCalendar = "BEGIN:VCALENDAR\r\nnot a calendar"

func calendarWithTask(uid, extraLine string) string {
	return "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//test//EN\r\nBEGIN:VTODO\r\nUID:" + uid +
		"\r\nSUMMARY:Calendar body test\r\n" + extraLine + "END:VTODO\r\nEND:VCALENDAR\r\n"
}

func tooManyParameters() string {
	var b strings.Builder
	b.WriteString("X-TEST")
	for i := range caldav2.MaxPropertyParameters + 1 {
		b.WriteString(";P" + strconv.Itoa(i) + "=v")
	}
	b.WriteString(":v\r\n")
	return b.String()
}

func callCaldavHandler(t *testing.T, handler func(*echo.Context) error, u *user.User, method, path, body string, params echo.PathValues) (*httptest.ResponseRecorder, error) {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Depth", "1")
	c := echo.New().NewContext(req, rec)
	c.Set("userBasicAuth", u)
	if params != nil {
		c.SetPathValues(params)
	}

	return rec, handler(c)
}

func projectParams(projectID int64) echo.PathValues {
	return echo.PathValues{{Name: "project", Value: strconv.FormatInt(projectID, 10)}}
}

func taskParams(projectID int64, uid string) echo.PathValues {
	return echo.PathValues{
		{Name: "project", Value: strconv.FormatInt(projectID, 10)},
		{Name: "task", Value: uid + ".ics"},
	}
}

func requireHTTPStatus(t *testing.T, want int, rec *httptest.ResponseRecorder, err error) {
	t.Helper()

	if err != nil {
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok, "unexpected error %v", err)
		assert.Equal(t, want, httpErr.Code)
		return
	}
	assert.Equal(t, want, rec.Code, rec.Body.String())
}

func taskWithUIDExists(t *testing.T, uid string) bool {
	t.Helper()

	s := db.NewSession()
	defer s.Close()
	exists, err := s.Where("uid = ?", uid).Exist(&models.Task{})
	require.NoError(t, err)
	return exists
}

func TestCalendarBodyParsing(t *testing.T) {
	t.Run("collection PUT without access does not parse the body", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Project 1 belongs to another user.
		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, http.MethodPut, projectPath(1), unparseableCalendar, projectParams(1))
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("collection requests other than PUT do not parse the body", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, "PROPFIND", projectPath(36), unparseableCalendar, projectParams(36))
		require.NoError(t, err)
		assert.Equal(t, http.StatusMultiStatus, rec.Code)
	})

	t.Run("collection PUT with access still creates the task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, http.MethodPut, projectPath(36), calendarWithTask("uid-collection-put", ""), projectParams(36))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		assert.True(t, taskWithUIDExists(t, "uid-collection-put"))
	})

	t.Run("collection PUT with an unparseable body is invalid", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, http.MethodPut, projectPath(36), unparseableCalendar, projectParams(36))
		require.Error(t, err)
		assert.True(t, models.IsErrInvalidData(err), "got %v", err)
	})

	t.Run("task PUT with too many property parameters is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		const uid = "uid-too-many-params"
		rec, err := callCaldavHandler(t, TaskHandler, caldavFilterUser, http.MethodPut, projectPath(36)+uid+".ics", calendarWithTask(uid, tooManyParameters()), taskParams(36, uid))
		requireHTTPStatus(t, http.StatusBadRequest, rec, err)
		assert.False(t, taskWithUIDExists(t, uid))
	})

	t.Run("task PUT with an oversized body is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		const uid = "uid-too-large"
		body := calendarWithTask(uid, "DESCRIPTION:"+strings.Repeat("a", caldav2.MaxVCalendarSize)+"\r\n")
		rec, err := callCaldavHandler(t, TaskHandler, caldavFilterUser, http.MethodPut, projectPath(36)+uid+".ics", body, taskParams(36, uid))
		requireHTTPStatus(t, http.StatusRequestEntityTooLarge, rec, err)
		assert.False(t, taskWithUIDExists(t, uid))
	})

	t.Run("collection PUT with too many property parameters is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, http.MethodPut, projectPath(36), calendarWithTask("uid-collection-params", tooManyParameters()), projectParams(36))
		requireHTTPStatus(t, http.StatusBadRequest, rec, err)
		assert.False(t, taskWithUIDExists(t, "uid-collection-params"))
	})
}

func TestRejectedBodiesOnInaccessibleProjects(t *testing.T) {
	// Project 1 belongs to another user, so every rejected body must look like a missing project.
	const uid = "uid-inaccessible"
	tests := []struct {
		name    string
		handler func(*echo.Context) error
		method  string
		path    string
		body    string
		params  echo.PathValues
		// TaskHandler returns the error; ProjectHandler writes the 404 itself
		returnsErr bool
	}{
		{
			name:    "collection PUT with an oversized body",
			handler: ProjectHandler,
			method:  http.MethodPut,
			path:    projectPath(1),
			body:    calendarWithTask(uid, "DESCRIPTION:"+strings.Repeat("a", caldav2.MaxVCalendarSize)+"\r\n"),
			params:  projectParams(1),
		},
		{
			name:       "task PUT with too many property parameters",
			handler:    TaskHandler,
			method:     http.MethodPut,
			path:       projectPath(1) + uid + ".ics",
			body:       calendarWithTask(uid, tooManyParameters()),
			params:     taskParams(1, uid),
			returnsErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			rec, err := callCaldavHandler(t, tt.handler, caldavFilterUser, tt.method, tt.path, tt.body, tt.params)
			if tt.returnsErr {
				require.Error(t, err)
				assert.True(t, models.IsErrProjectDoesNotExist(err), "got %v", err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Equal(t, "Project not found", rec.Body.String())
		})
	}
}
