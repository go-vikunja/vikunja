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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type accessExpectation struct {
	propfind, get, report, options, putUpdate, putCreate, del int
	allow                                                     string
}

var (
	readOnlyAccess = accessExpectation{
		propfind: 207, get: 200, report: 207, options: 200, allow: allowReadOnly,
		putUpdate: 403, putCreate: 403, del: 403,
	}
	readWriteAccess = accessExpectation{
		propfind: 207, get: 200, report: 207, options: 200, allow: allowReadWrite,
		putUpdate: 201, putCreate: 201, del: 204,
	}
	noAccess = accessExpectation{
		propfind: 404, get: 404, report: 404, options: 404,
		putUpdate: 404, putCreate: 404, del: 404,
	}
)

func taskUIDExists(t *testing.T, uid string) bool {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	exists, err := s.Where("uid = ?", uid).Exist(&models.Task{})
	require.NoError(t, err)
	return exists
}

// assertAccess runs every verb as username against an existing task, writes last so the
// reads see the fixture state, and checks that denied writes left the task alone.
func assertAccess(t *testing.T, e *echo.Echo, username, project, taskUID, ownerUsername string, want accessExpectation) {
	t.Helper()
	auth := basicAuthHeader(username, fixturePassword)
	collection := "/dav/projects/" + project
	task := collection + "/" + taskUID + ".ics"
	as := map[string]string{"Authorization": auth}

	propfind := caldavRequest(t, e, "PROPFIND", collection, PropfindCalendarCollectionProperties, map[string]string{"Authorization": auth, "Depth": "0"})
	assert.Equal(t, want.propfind, propfind.Code, "PROPFIND %s", collection)

	assert.Equal(t, want.get, caldavRequest(t, e, http.MethodGet, task, "", as).Code, "GET %s", task)
	assert.Equal(t, want.report, caldavRequest(t, e, "REPORT", collection, ReportCalendarQuery, as).Code, "REPORT %s", collection)

	options := caldavRequest(t, e, http.MethodOptions, collection, "", as)
	assert.Equal(t, want.options, options.Code, "OPTIONS %s", collection)
	assert.Equal(t, want.allow, options.Header().Get("Allow"), "OPTIONS %s Allow", collection)

	update := caldavPUTWith(t, e, task, NewVTodo(taskUID, "Updated by "+username).Build(), as)
	assert.Equal(t, want.putUpdate, update.Code, "PUT update %s", task)

	newUID := taskUID + "-new-" + username
	newTask := collection + "/" + newUID + ".ics"
	create := caldavPUTWith(t, e, newTask, NewVTodo(newUID, "Created").Build(), as)
	assert.Equal(t, want.putCreate, create.Code, "PUT create in %s", collection)

	ownerAs := map[string]string{"Authorization": basicAuthHeader(ownerUsername, fixturePassword)}
	ownerGet := func() *httptest.ResponseRecorder {
		return caldavRequest(t, e, http.MethodGet, task, "", ownerAs)
	}
	if want.putUpdate != http.StatusCreated {
		assert.NotContains(t, ownerGet().Body.String(), "Updated by "+username, "denied PUT must not change %s", task)
	} else {
		assert.Contains(t, ownerGet().Body.String(), "SUMMARY:Updated by "+username, "PUT must change %s", task)
	}
	if want.putCreate == http.StatusCreated {
		assert.Equal(t, http.StatusOK, caldavRequest(t, e, http.MethodGet, newTask, "", ownerAs).Code, "created %s", newTask)
	} else {
		assert.False(t, taskUIDExists(t, newUID), "denied PUT must not create %s", newTask)
	}

	assert.Equal(t, want.del, caldavRequest(t, e, http.MethodDelete, task, "", as).Code, "DELETE %s", task)
	if want.del == http.StatusNoContent {
		assert.Equal(t, http.StatusNotFound, ownerGet().Code, "deleted %s", task)
	} else {
		assert.Equal(t, http.StatusOK, ownerGet().Code, "denied DELETE must keep %s", task)
	}
}

func TestPinAccessPaths(t *testing.T) {
	// user6 owns projects 6-12; user1 reaches them through the path named, user13 has no access.
	tests := []struct {
		name    string
		project string
		access  accessExpectation
	}{
		{name: "team read", project: "6", access: readOnlyAccess},
		{name: "team write", project: "7", access: readWriteAccess},
		{name: "team admin", project: "8", access: readWriteAccess},
		{name: "user share read", project: "9", access: readOnlyAccess},
		{name: "user share write", project: "10", access: readWriteAccess},
		{name: "user share admin", project: "11", access: readWriteAccess},
		{name: "read share inherited from parent", project: "12", access: readOnlyAccess},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := setupTestEnv(t)
			uid := "pin-access-" + tc.project
			created := caldavPUTWith(t, e, "/dav/projects/"+tc.project+"/"+uid+".ics", NewVTodo(uid, "Owned by user6").Build(), map[string]string{
				"Authorization": basicAuthHeader("user6", fixturePassword),
			})
			require.Equal(t, http.StatusCreated, created.Code)

			t.Run("unrelated user", func(t *testing.T) {
				assertAccess(t, e, "user13", tc.project, uid, "user6", noAccess)
			})
			t.Run("shared user", func(t *testing.T) {
				assertAccess(t, e, testuser1.Username, tc.project, uid, "user6", tc.access)
			})
		})
	}

	t.Run("owner", func(t *testing.T) {
		e := setupTestEnv(t)

		t.Run("unrelated user", func(t *testing.T) {
			assertAccess(t, e, testuser1.Username, "36", "uid-caldav-test", testuser15.Username, noAccess)
		})
		t.Run("owner", func(t *testing.T) {
			assertAccess(t, e, testuser15.Username, "36", "uid-caldav-test", testuser15.Username, readWriteAccess)
		})
	})
}

func TestPinAccessCollectionOnly(t *testing.T) {
	type collectionExpectation struct {
		propfind, report, options, putCreate int
		allow                                string
	}

	tests := []struct {
		name     string
		project  string
		username string
		want     collectionExpectation
	}{
		// Project 22 is user1's own, archived.
		{name: "archived own project", project: "22", username: testuser1.Username, want: collectionExpectation{207, 207, 200, 403, allowReadOnly}},
		{name: "archived project, unrelated user", project: "22", username: "user13", want: collectionExpectation{404, 404, 404, 404, ""}},
		// Saved filter 1 (pseudo project -2) belongs to user1.
		{name: "own saved filter", project: "-2", username: testuser1.Username, want: collectionExpectation{207, 207, 200, 403, allowReadWrite}},
		{name: "saved filter, unrelated user", project: "-2", username: "user13", want: collectionExpectation{404, 404, 404, 404, ""}},
	}

	// Every write here is denied, so the cases share one fixture load.
	e := setupTestEnv(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth := basicAuthHeader(tc.username, fixturePassword)
			collection := "/dav/projects/" + tc.project

			propfind := caldavRequest(t, e, "PROPFIND", collection, PropfindCalendarCollectionProperties, map[string]string{"Authorization": auth, "Depth": "0"})
			assert.Equal(t, tc.want.propfind, propfind.Code, "PROPFIND")
			report := caldavRequest(t, e, "REPORT", collection, ReportCalendarQuery, map[string]string{"Authorization": auth})
			assert.Equal(t, tc.want.report, report.Code, "REPORT")
			options := caldavRequest(t, e, http.MethodOptions, collection, "", map[string]string{"Authorization": auth})
			assert.Equal(t, tc.want.options, options.Code, "OPTIONS")
			assert.Equal(t, tc.want.allow, options.Header().Get("Allow"), "Allow")
			create := caldavPUTWith(t, e, collection+"/pin-access-create.ics", NewVTodo("pin-access-create", "Created").Build(), map[string]string{"Authorization": auth})
			assert.Equal(t, tc.want.putCreate, create.Code, "PUT create")
			assert.False(t, taskUIDExists(t, "pin-access-create"), "denied PUT must not create a task")
		})
	}
}

func TestPinAccessGETCollection(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavRequest(t, e, http.MethodGet, "/dav/projects/36", "", map[string]string{
		"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
	})

	assert.Equal(t, http.StatusNotFound, rec.Code, "Body:\n%s", rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "BEGIN:VCALENDAR")
}

func TestPinAccessForeignPrincipal(t *testing.T) {
	e := setupTestEnv(t)

	for _, path := range []string{"/dav/principals/user1/", "/dav/principals/user15/sub/"} {
		rec := caldavPROPFIND(t, e, path, "0", PropfindDiscoveryProps)
		assert.Equal(t, http.StatusNotFound, rec.Code, "PROPFIND %s", path)
	}
}
