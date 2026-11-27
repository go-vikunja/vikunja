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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// A pseudo collection has no permissions: the task's project decides, creating is refused.
func TestAccessPerMethodFavorites(t *testing.T) {
	const favorites = "/dav/projects/-1"
	const task = favorites + "/uid-caldav-test.ics"
	putHeaders := map[string]string{"Content-Type": "text/calendar; charset=utf-8"}

	favorite := func(t *testing.T) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		_, err := s.Insert(&models.Favorite{
			EntityID: 40,
			UserID:   testuser15.ID,
			Kind:     models.FavoriteKindTask,
		})
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
