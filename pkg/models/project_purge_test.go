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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func projectPurgeMarkerForTest(t *testing.T, s *xorm.Session, projectID int64) time.Time {
	t.Helper()
	marker := struct {
		LastPurgedAt time.Time `xorm:"null"`
	}{}
	found, err := s.Table("projects").Where("id = ?", projectID).Get(&marker)
	require.NoError(t, err)
	require.True(t, found)
	return marker.LastPurgedAt
}

func TestTaskPurgeSyncMarker(t *testing.T) {
	u := &user.User{ID: 1}
	t.Run("advances even when another purge has a newer deletion timestamp", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		future := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
		s := db.NewSession()
		defer s.Close()
		_, err := s.Table("projects").Where("id = ?", 1).Update(map[string]any{
			"last_purged_at": future,
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		require.NoError(t, deletePermanentlyForTest(t, 51, u))
		s = db.NewSession()
		defer s.Close()
		assert.True(t, projectPurgeMarkerForTest(t, s, 1).Equal(future.Add(time.Second)))
		assert.True(t, projectPurgeMarkerForTest(t, s, 2).IsZero())
	})

	t.Run("rolls back together with the purged task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		task := &Task{ID: 51}
		can, err := task.CanDeletePermanently(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, task.DeletePermanently(s, u))
		assert.False(t, projectPurgeMarkerForTest(t, s, 1).IsZero())
		require.NoError(t, s.Rollback())
		s.Close()

		s = db.NewSession()
		defer s.Close()
		assert.True(t, projectPurgeMarkerForTest(t, s, 1).IsZero())
		db.AssertExists(t, "tasks", map[string]any{
			"id": 51,
		}, false)
	})

	t.Run("retention cleanup keeps the marker unchanged", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		deleteExpiredTasks(time.Now())
		s := db.NewSession()
		defer s.Close()
		assert.True(t, projectPurgeMarkerForTest(t, s, 1).IsZero())
	})
}
