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

func TestSubscriptionService_CanCreate(t *testing.T) {
	t.Run("CanReadProject_CanSubscribe", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Has access to project 1
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   1,
		}

		canCreate, err := ss.CanCreate(s, sub, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("CanReadTask_CanSubscribe", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1} // Has access to task 1
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityTask,
			EntityID:   1,
		}

		canCreate, err := ss.CanCreate(s, sub, u)

		require.NoError(t, err)
		assert.True(t, canCreate)
	})

	t.Run("LinkShare_CannotSubscribe", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   1,
		}

		canCreate, err := ss.CanCreate(s, sub, linkShare)

		assert.Error(t, err)
		assert.False(t, canCreate)
	})

	t.Run("NoAccessToProject_CannotSubscribe", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to project 1
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   1,
		}

		canCreate, err := ss.CanCreate(s, sub, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("NoAccessToTask_CannotSubscribe", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 13} // No access to task 1
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityTask,
			EntityID:   1,
		}

		canCreate, err := ss.CanCreate(s, sub, u)

		require.NoError(t, err)
		assert.False(t, canCreate)
	})

	t.Run("UnknownEntityType_Error", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: 999, // Invalid entity type
			EntityID:   1,
		}

		canCreate, err := ss.CanCreate(s, sub, u)

		assert.Error(t, err)
		assert.False(t, canCreate)
	})
}

func TestSubscriptionService_CanDelete(t *testing.T) {
	t.Run("OwnSubscription_CanDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// First create a subscription
		u := &user.User{ID: 1}
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   1,
			UserID:     1,
		}
		_, err := s.Insert(sub)
		require.NoError(t, err)

		ss := NewSubscriptionService(s.Engine())
		canDelete, err := ss.CanDelete(s, sub, u)

		require.NoError(t, err)
		assert.True(t, canDelete)
	})

	t.Run("LinkShare_CannotDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   1,
		}

		canDelete, err := ss.CanDelete(s, sub, linkShare)

		assert.Error(t, err)
		assert.False(t, canDelete)
	})

	t.Run("NonExistentSubscription_CannotDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		ss := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   9999, // Non-existent
			UserID:     1,
		}

		canDelete, err := ss.CanDelete(s, sub, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})

	t.Run("OtherUsersSubscription_CannotDelete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create subscription for user 1
		sub := &models.Subscription{
			EntityType: models.SubscriptionEntityProject,
			EntityID:   1,
			UserID:     1,
		}
		_, err := s.Insert(sub)
		require.NoError(t, err)

		// Try to delete as user 2
		u := &user.User{ID: 2}
		ss := NewSubscriptionService(s.Engine())
		canDelete, err := ss.CanDelete(s, sub, u)

		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}
