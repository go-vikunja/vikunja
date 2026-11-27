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
	propfind, get, report, options, putUpdate, putCreate, del, delCollection int
	allow                                                                    string
}

var (
	readOnlyAccess = accessExpectation{
		propfind:      207,
		get:           200,
		report:        207,
		options:       200,
		allow:         allowReadOnly,
		putUpdate:     403,
		putCreate:     403,
		del:           403,
		delCollection: 405,
	}
	readWriteAccess = accessExpectation{
		propfind:      207,
		get:           200,
		report:        207,
		options:       200,
		allow:         allowReadWrite,
		putUpdate:     204,
		putCreate:     201,
		del:           204,
		delCollection: 405,
	}
	noAccess = accessExpectation{
		propfind:      404,
		get:           404,
		report:        404,
		options:       404,
		putUpdate:     404,
		putCreate:     404,
		del:           404,
		delCollection: 404,
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

// Writes run last so the reads see the fixture state.
func assertAccess(t *testing.T, e *echo.Echo, username, project, taskUID, ownerUsername string, want accessExpectation) {
	t.Helper()
	auth := basicAuthHeader(username, fixturePassword)
	collection := "/dav/projects/" + project
	task := collection + "/" + taskUID + ".ics"
	as := map[string]string{"Authorization": auth}

	propfind := caldavRequest(t, e, "PROPFIND", collection, PropfindCalendarCollectionProperties, map[string]string{
		"Authorization": auth,
		"Depth":         "0",
	})
	assert.Equal(t, want.propfind, propfind.Code, "PROPFIND %s", collection)

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		assert.Equal(t, want.get, caldavRequest(t, e, method, task, "", as).Code, "%s %s", method, task)
		assert.Equal(t, want.get, caldavRequest(t, e, method, collection, "", as).Code, "%s %s", method, collection)
	}
	assert.Equal(t, want.report, caldavRequest(t, e, "REPORT", collection, ReportCalendarQuery, as).Code, "REPORT %s", collection)

	for _, path := range []string{task, collection} {
		options := caldavRequest(t, e, http.MethodOptions, path, "", as)
		assert.Equal(t, want.options, options.Code, "OPTIONS %s", path)
		assert.Equal(t, want.allow, options.Header().Get("Allow"), "OPTIONS %s Allow", path)
	}

	assert.Equal(t, want.delCollection, caldavRequest(t, e, http.MethodDelete, collection, "", as).Code, "DELETE %s", collection)

	ownerAs := map[string]string{"Authorization": basicAuthHeader(ownerUsername, fixturePassword)}
	ownerGet := func() *httptest.ResponseRecorder {
		return caldavRequest(t, e, http.MethodGet, task, "", ownerAs)
	}

	updated := "Updated by " + username
	update := caldavPUTWith(t, e, task, NewVTodo(taskUID, updated).Build(), as)
	assert.Equal(t, want.putUpdate, update.Code, "PUT update %s", task)
	if want.putUpdate == http.StatusNoContent {
		assert.Contains(t, ownerGet().Body.String(), "SUMMARY:"+updated, "PUT must change %s", task)
	}
	viaCollection := "Updated via collection by " + username
	update = caldavPUTWith(t, e, collection, NewVTodo(taskUID, viaCollection).Build(), as)
	assert.Equal(t, want.putUpdate, update.Code, "PUT update via %s", collection)
	if want.putUpdate == http.StatusNoContent {
		assert.Contains(t, ownerGet().Body.String(), "SUMMARY:"+viaCollection, "PUT via %s must change %s", collection, task)
	} else {
		body := ownerGet().Body.String()
		assert.NotContains(t, body, updated, "denied PUT must not change %s", task)
		assert.NotContains(t, body, viaCollection, "denied PUT must not change %s", task)
	}

	newUID := taskUID + "-new-" + username
	newTask := collection + "/" + newUID + ".ics"
	create := caldavPUTWith(t, e, newTask, NewVTodo(newUID, "Created").Build(), as)
	assert.Equal(t, want.putCreate, create.Code, "PUT create in %s", collection)
	newViaCollectionUID := newUID + "-c"
	newViaCollection := collection + "/" + newViaCollectionUID + ".ics"
	create = caldavPUTWith(t, e, collection, NewVTodo(newViaCollectionUID, "Created").Build(), as)
	assert.Equal(t, want.putCreate, create.Code, "PUT create via %s", collection)
	if want.putCreate == http.StatusCreated {
		assert.Equal(t, http.StatusOK, caldavRequest(t, e, http.MethodGet, newTask, "", ownerAs).Code, "created %s", newTask)
		assert.Equal(t, http.StatusOK, caldavRequest(t, e, http.MethodGet, newViaCollection, "", ownerAs).Code, "created %s", newViaCollection)
	} else {
		assert.False(t, taskUIDExists(t, newUID), "denied PUT must not create %s", newTask)
		assert.False(t, taskUIDExists(t, newViaCollectionUID), "denied PUT must not create %s", newViaCollection)
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
		{
			name:    "team read",
			project: "6",
			access:  readOnlyAccess,
		},
		{
			name:    "team write",
			project: "7",
			access:  readWriteAccess,
		},
		{
			name:    "team admin",
			project: "8",
			access:  readWriteAccess,
		},
		{
			name:    "user share read",
			project: "9",
			access:  readOnlyAccess,
		},
		{
			name:    "user share write",
			project: "10",
			access:  readWriteAccess,
		},
		{
			name:    "user share admin",
			project: "11",
			access:  readWriteAccess,
		},
		{
			name:    "read share inherited from parent",
			project: "12",
			access:  readOnlyAccess,
		},
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

	t.Run("archived own project", func(t *testing.T) {
		e := setupTestEnv(t)
		archive(t, 36)
		assertAccess(t, e, testuser15.Username, "36", "uid-caldav-test", testuser15.Username, readOnlyAccess)
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
		{
			name:     "archived own project",
			project:  "22",
			username: testuser1.Username,
			want: collectionExpectation{
				propfind:  207,
				report:    207,
				options:   200,
				putCreate: 403,
				allow:     allowReadOnly,
			},
		},
		{
			name:     "archived project, unrelated user",
			project:  "22",
			username: "user13",
			want: collectionExpectation{
				propfind:  404,
				report:    404,
				options:   404,
				putCreate: 404,
			},
		},
		{
			name:     "own saved filter",
			project:  "-2",
			username: testuser1.Username,
			want: collectionExpectation{
				propfind:  207,
				report:    207,
				options:   200,
				putCreate: 403,
				allow:     allowReadWrite,
			},
		},
		{
			name:     "saved filter, unrelated user",
			project:  "-2",
			username: "user13",
			want: collectionExpectation{
				propfind:  404,
				report:    404,
				options:   404,
				putCreate: 404,
			},
		},
	}

	e := setupTestEnv(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth := basicAuthHeader(tc.username, fixturePassword)
			collection := "/dav/projects/" + tc.project

			propfind := caldavRequest(t, e, "PROPFIND", collection, PropfindCalendarCollectionProperties, map[string]string{
				"Authorization": auth,
				"Depth":         "0",
			})
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

	for _, path := range []string{
		"/dav/principals/user1/",
		"/dav/principals/user15/sub/",
	} {
		rec := caldavPROPFIND(t, e, path, "0", PropfindDiscoveryProps)
		assert.Equal(t, http.StatusNotFound, rec.Code, "PROPFIND %s", path)
	}
}

// Neither URL names a collection and a task, so neither may fall back to a lookup by UID alone.
func TestPinMalformedTaskURLs(t *testing.T) {
	e := setupTestEnv(t)

	// Without it a lookup of the empty UID finds nothing and the empty-UID case passes by accident.
	s := db.NewSession()
	_, err := s.Insert(&models.Task{Title: "Task without uid", ProjectID: 36, Index: 999, CreatedByID: testuser15.ID})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	for _, path := range []string{"/dav/projects//uid-caldav-test.ics", "/dav/projects/36/.ics"} {
		t.Run(path, func(t *testing.T) {
			for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions} {
				rec := caldavRequest(t, e, method, path, "", nil)
				assert.Equal(t, http.StatusNotFound, rec.Code, "%s. Body:\n%s", method, rec.Body.String())
			}

			rec := caldavPROPFIND(t, e, path, "0", PropfindResourceProperties)
			assert.Equal(t, http.StatusNotFound, rec.Code, "PROPFIND. Body:\n%s", rec.Body.String())

			rec = caldavPUT(t, e, path, NewVTodo("uid-caldav-test", "Overwritten").Build())
			assert.Equal(t, http.StatusNotFound, rec.Code, "PUT. Body:\n%s", rec.Body.String())
		})
	}

	assert.Equal(t, http.StatusOK, caldavGET(t, e, pinTaskHref).Code, "the task must survive")
	assert.NotContains(t, caldavGET(t, e, pinTaskHref).Body.String(), "Overwritten")
}

// tasks.uid has no unique index: the same UID in another readable project must not hide
// the task living in the project the URL names.
func TestPinSameUIDInAnotherProject(t *testing.T) {
	e := setupTestEnv(t)

	s := db.NewSession()
	_, err := s.Insert(&models.Task{Title: "Duplicate UID task", UID: "uid-caldav-test", ProjectID: 38, Index: 999, CreatedByID: testuser15.ID})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	const duplicate = "/dav/projects/38/uid-caldav-test.ics"
	rec := caldavGET(t, e, duplicate)
	require.Equal(t, http.StatusOK, rec.Code, "Body:\n%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "SUMMARY:Duplicate UID task")

	rec = caldavPUT(t, e, duplicate, NewVTodo("uid-caldav-test", "Duplicate updated").Build())
	require.Equal(t, http.StatusNoContent, rec.Code, "Body:\n%s", rec.Body.String())
	assert.Contains(t, caldavGET(t, e, duplicate).Body.String(), "SUMMARY:Duplicate updated")

	original := caldavGET(t, e, pinTaskHref).Body.String()
	assert.Contains(t, original, "SUMMARY:Title Caldav Test")

	require.Equal(t, http.StatusNoContent, caldavDELETE(t, e, duplicate).Code)
	assert.Equal(t, http.StatusNotFound, caldavGET(t, e, duplicate).Code)
	assert.Equal(t, http.StatusOK, caldavGET(t, e, pinTaskHref).Code)
}
