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

package services

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKanbanService_CanCreate(t *testing.T) {
	t.Run("CanUpdateProject_CanCreateBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		ks := NewKanbanService(s.Engine())
		bucket := &models.Bucket{
			ProjectViewID: 4, // Belongs to project 1
			ProjectID:     1,
			Title:         "New Bucket",
		}

		canCreate, err := ks.CanCreate(s, bucket, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("CannotUpdateProject_CannotCreateBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		ks := NewKanbanService(s.Engine())
		bucket := &models.Bucket{
			ProjectViewID: 4,
			ProjectID:     1,
			Title:         "New Bucket",
		}

		canCreate, err := ks.CanCreate(s, bucket, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("NonExistentProjectView", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ks := NewKanbanService(s.Engine())
		bucket := &models.Bucket{
			ProjectViewID: 9999,
			ProjectID:     1,
			Title:         "New Bucket",
		}

		canCreate, err := ks.CanCreate(s, bucket, u)

		assert.Error(t, err)
		assert.False(t, canCreate)
	})
}

func TestKanbanService_CanUpdate(t *testing.T) {
	t.Run("CanUpdateProject_CanUpdateBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1, bucket 1 belongs to project 1
		ks := NewKanbanService(s.Engine())

		canUpdate, err := ks.CanUpdate(s, 1, 1, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("CannotUpdateProject_CannotUpdateBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		ks := NewKanbanService(s.Engine())

		canUpdate, err := ks.CanUpdate(s, 1, 1, u)

		require.NoError(t, err)
		assert.False(t, canUpdate)
	})

	t.Run("NonExistentBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ks := NewKanbanService(s.Engine())

		canUpdate, err := ks.CanUpdate(s, 9999, 1, u)

		assert.Error(t, err)
		assert.True(t, models.IsErrBucketDoesNotExist(err))
		assert.False(t, canUpdate)
	})
}

func TestKanbanService_CanDelete(t *testing.T) {
	t.Run("CanUpdateProject_CanDeleteBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		ks := NewKanbanService(s.Engine())

		canDelete, err := ks.CanDelete(s, 1, 1, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("CannotUpdateProject_CannotDeleteBucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		ks := NewKanbanService(s.Engine())

		canDelete, err := ks.CanDelete(s, 1, 1, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}

func TestKanbanService_CanUpdateTaskBucket(t *testing.T) {
	t.Run("CanUpdateProject_CanMoveTask", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Owner of project 1
		ks := NewKanbanService(s.Engine())

		canUpdate, err := ks.CanUpdateTaskBucket(s, 1, 1, 4, u)

		require.NoError(t, err)
		assert.True(t, canUpdate)
	})

	t.Run("CannotUpdateProject_CannotMoveTask", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		ks := NewKanbanService(s.Engine())

		canUpdate, err := ks.CanUpdateTaskBucket(s, 1, 1, 4, u)

		require.NoError(t, err)
		assert.False(t, canUpdate)
	})
}
