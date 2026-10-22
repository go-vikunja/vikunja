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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Guards stale favorites after project access is revoked (GHSA-jp29-jrxc-92vf).
func TestFavoriteTasksRequireCurrentAccess(t *testing.T) {
	setup := func(t *testing.T) (s *xorm.Session, reader *user.User, proj *Project, task *Task) {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s = db.NewSession()

		owner := &user.User{ID: 1}
		reader = &user.User{ID: 2}

		proj = &Project{Title: "favorites-access project"}
		require.NoError(t, proj.Create(s, owner))
		_, err := s.Insert(&ProjectUser{UserID: reader.ID, ProjectID: proj.ID, Permission: PermissionRead})
		require.NoError(t, err)

		task = &Task{Title: "favorites-access task", ProjectID: proj.ID}
		require.NoError(t, task.Create(s, owner))

		_, err = s.Insert(&Favorite{
			EntityID: task.ID,
			UserID:   reader.ID,
			Kind:     FavoriteKindTask,
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		return
	}

	favoriteIDsFor := func(t *testing.T, s *xorm.Session, reader *user.User) map[int64]bool {
		t.Helper()
		tasks, _, _, err := getRawTasksForProjects(s, []*Project{&FavoritesPseudoProject}, reader, &taskSearchOptions{})
		require.NoError(t, err)
		return taskIDsOf(tasks)
	}

	t.Run("visible while access exists", func(t *testing.T) {
		s, reader, _, task := setup(t)
		defer s.Close()

		assert.Contains(t, favoriteIDsFor(t, s, reader), task.ID)
	})

	t.Run("hidden after access is revoked, favorite row kept", func(t *testing.T) {
		s, reader, proj, task := setup(t)
		defer s.Close()

		_, err := s.Where("user_id = ? AND project_id = ?", reader.ID, proj.ID).Delete(&ProjectUser{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.NotContains(t, favoriteIDsFor(t, s, reader), task.ID,
			"a favorite from a revoked project must not be returned")

		db.AssertExists(t, "favorites", map[string]interface{}{
			"entity_id": task.ID,
			"user_id":   reader.ID,
			"kind":      FavoriteKindTask,
		}, false)
	})

	t.Run("still hidden after the task is edited later", func(t *testing.T) {
		s, reader, proj, task := setup(t)
		defer s.Close()

		_, err := s.Where("user_id = ? AND project_id = ?", reader.ID, proj.ID).Delete(&ProjectUser{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		_, err = s.ID(task.ID).Cols("title").Update(&Task{Title: "favorites-access task edited"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.NotContains(t, favoriteIDsFor(t, s, reader), task.ID)
	})

	t.Run("visible again through team access", func(t *testing.T) {
		s, reader, proj, task := setup(t)
		defer s.Close()

		_, err := s.Where("user_id = ? AND project_id = ?", reader.ID, proj.ID).Delete(&ProjectUser{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		assert.NotContains(t, favoriteIDsFor(t, s, reader), task.ID,
			"must be hidden right after the direct share is revoked")

		team := &Team{Name: "favorites-access team"}
		_, err = s.Insert(team)
		require.NoError(t, err)
		_, err = s.Insert(&TeamMember{TeamID: team.ID, UserID: reader.ID})
		require.NoError(t, err)
		_, err = s.Insert(&TeamProject{TeamID: team.ID, ProjectID: proj.ID, Permission: PermissionRead})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Contains(t, favoriteIDsFor(t, s, reader), task.ID,
			"a favorite must become visible again when access is restored through a team")
	})

	t.Run("visible again when direct access is restored", func(t *testing.T) {
		s, reader, proj, task := setup(t)
		defer s.Close()

		_, err := s.Where("user_id = ? AND project_id = ?", reader.ID, proj.ID).Delete(&ProjectUser{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		_, err = s.Insert(&ProjectUser{UserID: reader.ID, ProjectID: proj.ID, Permission: PermissionRead})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Contains(t, favoriteIDsFor(t, s, reader), task.ID,
			"a favorite must become visible again when access is re-granted")
	})
}
