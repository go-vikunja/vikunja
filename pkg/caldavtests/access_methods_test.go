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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

type methodExpectation struct {
	getTask, optionsTask, getCollection, optionsCollection, deleteCollection int
	allow                                                                    string
	putUpdate, putUpdateViaCollection, putCreate, putCreateViaCollection     int
	deleteTask                                                               int
}

var (
	methodsReadOnly = methodExpectation{
		getTask: 200, optionsTask: 200, getCollection: 200, optionsCollection: 200, deleteCollection: 405,
		allow:     allowReadOnly,
		putUpdate: 403, putUpdateViaCollection: 403, putCreate: 403, putCreateViaCollection: 403,
		deleteTask: 403,
	}
	methodsReadWrite = methodExpectation{
		getTask: 200, optionsTask: 200, getCollection: 200, optionsCollection: 200, deleteCollection: 405,
		allow:     allowReadWrite,
		putUpdate: 204, putUpdateViaCollection: 204, putCreate: 201, putCreateViaCollection: 201,
		deleteTask: 204,
	}
	methodsNoAccess = methodExpectation{
		getTask: 404, optionsTask: 404, getCollection: 404, optionsCollection: 404, deleteCollection: 404,
		putUpdate: 404, putUpdateViaCollection: 404, putCreate: 404, putCreateViaCollection: 404,
		deleteTask: 404,
	}
)

// assertMethods runs every in-house method as username against an existing task, reads
// first and deleting writes last, and checks that denied writes left the task alone.
func assertMethods(t *testing.T, e *echo.Echo, username, project, taskUID, ownerUsername string, want methodExpectation) {
	t.Helper()
	as := map[string]string{"Authorization": basicAuthHeader(username, fixturePassword)}
	putAs := map[string]string{"Authorization": as["Authorization"], "Content-Type": "text/calendar; charset=utf-8"}
	collection := "/dav/projects/" + project
	task := collection + "/" + taskUID + ".ics"

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		assert.Equal(t, want.getTask, caldavRequest(t, e, method, task, "", as).Code, "%s %s", method, task)
		assert.Equal(t, want.getCollection, caldavRequest(t, e, method, collection, "", as).Code, "%s %s", method, collection)
	}

	for path, status := range map[string]int{task: want.optionsTask, collection: want.optionsCollection} {
		rec := caldavRequest(t, e, http.MethodOptions, path, "", as)
		assert.Equal(t, status, rec.Code, "OPTIONS %s", path)
		assert.Equal(t, want.allow, rec.Header().Get("Allow"), "OPTIONS %s Allow", path)
	}

	assert.Equal(t, want.deleteCollection, caldavRequest(t, e, http.MethodDelete, collection, "", as).Code, "DELETE %s", collection)

	ownerGet := func() string {
		rec := caldavRequest(t, e, http.MethodGet, task, "", map[string]string{"Authorization": basicAuthHeader(ownerUsername, fixturePassword)})
		require.Equal(t, http.StatusOK, rec.Code, "owner GET %s", task)
		return rec.Body.String()
	}

	updated := "Updated by " + username
	assert.Equal(t, want.putUpdate, caldavRequest(t, e, http.MethodPut, task, NewVTodo(taskUID, updated).Build(), putAs).Code, "PUT update %s", task)
	viaCollection := "Updated via collection by " + username
	assert.Equal(t, want.putUpdateViaCollection, caldavRequest(t, e, http.MethodPut, collection, NewVTodo(taskUID, viaCollection).Build(), putAs).Code, "PUT update via %s", collection)
	if want.putUpdate != http.StatusNoContent {
		body := ownerGet()
		assert.NotContains(t, body, updated, "denied PUT must not change %s", task)
		assert.NotContains(t, body, viaCollection, "denied PUT must not change %s", task)
	}

	newUID := taskUID + "-new-" + username
	assert.Equal(t, want.putCreate, caldavRequest(t, e, http.MethodPut, collection+"/"+newUID+".ics", NewVTodo(newUID, "Created").Build(), putAs).Code, "PUT create in %s", collection)
	assert.Equal(t, want.putCreateViaCollection, caldavRequest(t, e, http.MethodPut, collection, NewVTodo(newUID+"-c", "Created").Build(), putAs).Code, "PUT create via %s", collection)

	assert.Equal(t, want.deleteTask, caldavRequest(t, e, http.MethodDelete, task, "", as).Code, "DELETE %s", task)
	if want.deleteTask != http.StatusNoContent {
		ownerGet()
	}
}

func TestAccessPerMethod(t *testing.T) {
	// user6 owns projects 6-12; user1 reaches them through the path named, user13 has no access.
	tests := []struct {
		name    string
		project string
		access  methodExpectation
	}{
		{name: "team read", project: "6", access: methodsReadOnly},
		{name: "team write", project: "7", access: methodsReadWrite},
		{name: "team admin", project: "8", access: methodsReadWrite},
		{name: "user share read", project: "9", access: methodsReadOnly},
		{name: "user share write", project: "10", access: methodsReadWrite},
		{name: "user share admin", project: "11", access: methodsReadWrite},
		{name: "read share inherited from parent", project: "12", access: methodsReadOnly},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := setupTestEnv(t)
			uid := "access-methods-" + tc.project
			created := caldavRequest(t, e, http.MethodPut, "/dav/projects/"+tc.project+"/"+uid+".ics", NewVTodo(uid, "Owned by user6").Build(), map[string]string{
				"Authorization": basicAuthHeader("user6", fixturePassword),
				"Content-Type":  "text/calendar; charset=utf-8",
			})
			require.Equal(t, http.StatusCreated, created.Code)

			t.Run("unrelated user", func(t *testing.T) {
				assertMethods(t, e, "user13", tc.project, uid, "user6", methodsNoAccess)
			})
			t.Run("shared user", func(t *testing.T) {
				assertMethods(t, e, testuser1.Username, tc.project, uid, "user6", tc.access)
			})
		})
	}

	t.Run("owner", func(t *testing.T) {
		e := setupTestEnv(t)
		assertMethods(t, e, testuser15.Username, "36", "uid-caldav-test", testuser15.Username, methodsReadWrite)
	})

	t.Run("archived own project", func(t *testing.T) {
		e := setupTestEnv(t)
		archive(t, 36)
		assertMethods(t, e, testuser15.Username, "36", "uid-caldav-test", testuser15.Username, methodsReadOnly)
	})
}

// A pseudo collection has no permissions of its own: reads and writes of a task it holds
// are decided by the task's project, creating in it is refused.
func TestAccessPerMethodFavorites(t *testing.T) {
	const favorites = "/dav/projects/-1"
	const task = favorites + "/uid-caldav-test.ics"
	putHeaders := map[string]string{"Content-Type": "text/calendar; charset=utf-8"}

	favorite := func(t *testing.T) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		_, err := s.Insert(&models.Favorite{EntityID: 40, UserID: testuser15.ID, Kind: models.FavoriteKindTask})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
	}

	t.Run("a favorite in a writable project", func(t *testing.T) {
		e := setupTestEnv(t)
		favorite(t)

		assert.Equal(t, http.StatusOK, caldavGET(t, e, task).Code)
		options := caldavOPTIONS(t, e, task)
		assert.Equal(t, http.StatusOK, options.Code)
		assert.Equal(t, allowReadWrite, options.Header().Get("Allow"))
		assert.Equal(t, http.StatusNoContent, caldavRequest(t, e, http.MethodPut, task, NewVTodo("uid-caldav-test", "Via favorites").Build(), putHeaders).Code)
		assert.Contains(t, caldavGET(t, e, pinTaskHref).Body.String(), "SUMMARY:Via favorites")
		assert.Equal(t, http.StatusForbidden, caldavRequest(t, e, http.MethodPut, favorites+"/new-favorite.ics", NewVTodo("new-favorite", "New").Build(), putHeaders).Code)
		assert.Equal(t, http.StatusNoContent, caldavDELETE(t, e, task).Code)
		assert.Equal(t, http.StatusNotFound, caldavGET(t, e, pinTaskHref).Code)
	})

	t.Run("a favorite in an archived project", func(t *testing.T) {
		e := setupTestEnv(t)
		favorite(t)
		archive(t, 36)

		assert.Equal(t, http.StatusOK, caldavGET(t, e, task).Code)
		assert.Equal(t, http.StatusForbidden, caldavRequest(t, e, http.MethodPut, task, NewVTodo("uid-caldav-test", "Via favorites").Build(), putHeaders).Code)
		assert.Equal(t, http.StatusForbidden, caldavDELETE(t, e, task).Code)
		assert.Equal(t, http.StatusOK, caldavGET(t, e, pinTaskHref).Code)
	})

	t.Run("a task that is no favorite", func(t *testing.T) {
		e := setupTestEnv(t)

		assert.Equal(t, http.StatusNotFound, caldavGET(t, e, task).Code)
		assert.Equal(t, http.StatusNotFound, caldavRequest(t, e, http.MethodPut, task, NewVTodo("uid-caldav-test", "Via favorites").Build(), putHeaders).Code)
		assert.Equal(t, http.StatusNotFound, caldavDELETE(t, e, task).Code)
		assert.NotContains(t, caldavGET(t, e, pinTaskHref).Body.String(), "Via favorites")
	})
}

func archive(t *testing.T, projectID int64) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where(builder.Eq{"id": projectID}).Cols("is_archived").Update(&models.Project{IsArchived: true})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}
