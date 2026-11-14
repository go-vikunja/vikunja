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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resolveFor(t *testing.T, resolve targetResolver, path string, params echo.PathValues) (target, error) {
	t.Helper()
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, path, nil), httptest.NewRecorder())
	if params != nil {
		c.SetPathValues(params)
	}
	return resolve(c, caldavFilterUser)
}

func requireNotFound(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, errorStatus(err), "got %v", err)
}

func TestResolveTarget(t *testing.T) {
	t.Run("entry", func(t *testing.T) {
		for _, path := range []string{"/dav", "/dav/"} {
			got, err := resolveFor(t, resolveEntry, path, nil)
			require.NoError(t, err)
			assert.Equal(t, entryTarget{}, got)
		}
	})

	t.Run("own principal", func(t *testing.T) {
		for _, path := range []string{"/dav/principals/user15", "/dav/principals/user15/", "/.well-known/caldav", "/.well-known/caldav/"} {
			got, err := resolveFor(t, resolvePrincipal, path, nil)
			require.NoError(t, err, path)
			assert.Equal(t, principalTarget{}, got, path)
		}
	})

	t.Run("foreign or nested principal", func(t *testing.T) {
		for _, path := range []string{"/dav/principals/user1/", "/dav/principals/does-not-exist/", "/dav/principals/user15/projects", "/dav/principals/", "/dav/principals/user15x"} {
			_, err := resolveFor(t, resolvePrincipal, path, nil)
			requireNotFound(t, err)
		}
	})

	t.Run("home set", func(t *testing.T) {
		for _, path := range []string{"/dav/projects", "/dav/projects/"} {
			got, err := resolveFor(t, resolveProjects, path, nil)
			require.NoError(t, err)
			assert.Equal(t, homeSetTarget{}, got)
		}
	})

	t.Run("project collection", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		got, err := resolveFor(t, resolveProjects, projectPath(36), projectParams(36))
		require.NoError(t, err)
		require.IsType(t, collectionTarget{}, got)
		assert.Equal(t, int64(36), got.(collectionTarget).project.ID)
	})

	t.Run("favorites collection", func(t *testing.T) {
		got, err := resolveFor(t, resolveProjects, projectPath(models.FavoritesPseudoProjectID), projectParams(models.FavoritesPseudoProjectID))
		require.NoError(t, err)
		require.IsType(t, collectionTarget{}, got)
		assert.Equal(t, int64(models.FavoritesPseudoProjectID), got.(collectionTarget).project.ID)
	})

	t.Run("saved filter collection", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		got, err := resolveFor(t, resolveProjects, projectPath(-2), projectParams(-2))
		require.NoError(t, err)
		require.IsType(t, collectionTarget{}, got)
		project := got.(collectionTarget).project
		assert.Equal(t, int64(-2), project.ID)
		assert.Equal(t, "testfilter1", project.Title)
	})

	t.Run("bad collection ids", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		for _, id := range []string{"abc", "0", "999999", "-999", "9223372036854775808"} {
			params := echo.PathValues{{Name: "project", Value: id}}
			_, err := resolveFor(t, resolveProjects, "/dav/projects/"+id, params)
			requireNotFound(t, err)

			_, err = resolveFor(t, resolveTask, "/dav/projects/"+id+"/uid.ics", append(params, echo.PathValue{Name: "task", Value: "uid.ics"}))
			requireNotFound(t, err)
		}
	})

	t.Run("task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		cases := []struct {
			path, uid string
		}{
			{"/dav/projects/36/uid-caldav-test.ics", "uid-caldav-test"},
			{"/dav/projects/36/uid-caldav-test", "uid-caldav-test"},
			{"/dav/projects/36/a%2Fb%26c.ics", "a/b&c"},
			{"/dav/projects/36/50%25off.ics", "50%off"},
			{"/dav/projects/36/a%2541b.ics", "a%41b"},
			{"/dav/projects/36/t%C3%A4sk.ics", "täsk"},
		}
		for _, tc := range cases {
			// The task param is ignored: routers disagree on whether it arrives decoded.
			got, err := resolveFor(t, resolveTask, tc.path, taskParams(36, "decoy"))
			require.NoError(t, err, tc.path)
			require.IsType(t, taskTarget{}, got)
			task := got.(taskTarget)
			assert.Equal(t, tc.uid, task.uid, tc.path)
			assert.Equal(t, int64(36), task.collection.project.ID)
		}
	})

	t.Run("task without a collection or uid", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := resolveFor(t, resolveTask, "/dav/projects//uid.ics", echo.PathValues{{Name: "project", Value: ""}, {Name: "task", Value: "uid.ics"}})
		requireNotFound(t, err)

		_, err = resolveFor(t, resolveTask, "/dav/projects/36/.ics", taskParams(36, ""))
		requireNotFound(t, err)
	})

	t.Run("task in a pseudo collection", func(t *testing.T) {
		got, err := resolveFor(t, resolveTask, "/dav/projects/-1/uid.ics", taskParams(models.FavoritesPseudoProjectID, "uid"))
		require.NoError(t, err)
		assert.Equal(t, int64(models.FavoritesPseudoProjectID), got.(taskTarget).collection.project.ID)
	})
}

func TestNewStorage(t *testing.T) {
	project := &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}}

	entry := newStorage(caldavFilterUser, entryTarget{})
	assert.Nil(t, entry.project)
	assert.Nil(t, entry.task)
	assert.Equal(t, caldavFilterUser, entry.user)

	principal := newStorage(caldavFilterUser, principalTarget{})
	assert.Nil(t, principal.project)
	assert.Nil(t, principal.task)

	homeSet := newStorage(caldavFilterUser, homeSetTarget{})
	require.NotNil(t, homeSet.project)
	assert.Zero(t, homeSet.project.ID)

	collection := newStorage(caldavFilterUser, collectionTarget{project: project})
	assert.Same(t, project, collection.project)
	assert.Nil(t, collection.task)

	task := newStorage(caldavFilterUser, taskTarget{collection: collectionTarget{project: project}, uid: "a/b"})
	assert.Same(t, project, task.project)
	require.NotNil(t, task.task)
	assert.Equal(t, "a/b", task.task.UID)
}
