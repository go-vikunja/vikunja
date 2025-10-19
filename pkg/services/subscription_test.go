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

func TestSubscriptionService_Create(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "task",
			EntityType: models.SubscriptionEntityTask,
			EntityID:   1,
		}

		err := service.Create(s, sub, u)
		require.NoError(t, err)

		db.AssertExists(t, "subscriptions", map[string]interface{}{
			"entity_type": 3,
			"entity_id":   1,
			"user_id":     u.ID,
		}, false)
	})

	t.Run("already exists", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "task",
			EntityType: models.SubscriptionEntityTask,
			EntityID:   2,
		}

		err := service.Create(s, sub, u)
		require.Error(t, err)
		assert.True(t, models.IsErrSubscriptionAlreadyExists(err))
	})

	t.Run("forbidden for link shares", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{}
		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "task",
			EntityType: models.SubscriptionEntityTask,
			EntityID:   1,
		}

		err := service.Create(s, sub, linkShare)
		require.Error(t, err)
		t.Logf("Error type: %T, value: %v", err, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})

	t.Run("nonexisting project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "project",
			EntityType: models.SubscriptionEntityProject,
			EntityID:   99999999,
		}

		err := service.Create(s, sub, u)
		require.Error(t, err)
		assert.True(t, models.IsErrProjectDoesNotExist(err))
	})

	t.Run("nonexisting task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "task",
			EntityType: models.SubscriptionEntityTask,
			EntityID:   99999999,
		}

		err := service.Create(s, sub, u)
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})

	t.Run("no permissions to see project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "project",
			EntityType: models.SubscriptionEntityProject,
			EntityID:   20,
		}

		err := service.Create(s, sub, u)
		require.Error(t, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})

	t.Run("no permissions to see task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "task",
			EntityType: models.SubscriptionEntityTask,
			EntityID:   14,
		}

		err := service.Create(s, sub, u)
		require.Error(t, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})

	t.Run("invalid entity type", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub := &models.Subscription{
			Entity:     "unknown",
			EntityType: models.SubscriptionEntityUnknown,
			EntityID:   1,
		}

		err := service.Create(s, sub, u)
		require.Error(t, err)
		assert.True(t, models.IsErrUnknownSubscriptionEntityType(err))
	})
}

func TestSubscriptionService_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		service := NewSubscriptionService(s.Engine())

		err := service.Delete(s, models.SubscriptionEntityTask, 2, u)
		require.NoError(t, err)

		db.AssertMissing(t, "subscriptions", map[string]interface{}{
			"entity_type": 3,
			"entity_id":   2,
			"user_id":     u.ID,
		})
	})

	t.Run("forbidden for link shares", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		linkShare := &models.LinkSharing{}
		service := NewSubscriptionService(s.Engine())

		err := service.Delete(s, models.SubscriptionEntityTask, 2, linkShare)
		require.Error(t, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})

	t.Run("not owner of the subscription", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 2}
		service := NewSubscriptionService(s.Engine())

		err := service.Delete(s, models.SubscriptionEntityTask, 2, u)
		require.Error(t, err)
		assert.True(t, models.IsErrGenericForbidden(err))
	})

	t.Run("invalid entity type", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		service := NewSubscriptionService(s.Engine())

		err := service.Delete(s, models.SubscriptionEntityUnknown, 2, u)
		require.Error(t, err)
		assert.True(t, models.IsErrUnknownSubscriptionEntityType(err))
	})
}

func TestSubscriptionService_GetForUser(t *testing.T) {
	u := &user.User{ID: 6}

	t.Run("test each individually", func(t *testing.T) {
		t.Run("project", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			service := NewSubscriptionService(s.Engine())
			sub, err := service.GetForUser(s, models.SubscriptionEntityProject, 12, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(3), sub.ID)
		})

		t.Run("task", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			service := NewSubscriptionService(s.Engine())
			sub, err := service.GetForUser(s, models.SubscriptionEntityTask, 22, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(4), sub.ID)
		})
	})

	t.Run("inherited", func(t *testing.T) {
		t.Run("project from parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			service := NewSubscriptionService(s.Engine())
			// Project 25 belongs to project 12 where user 6 has subscribed to
			sub, err := service.GetForUser(s, models.SubscriptionEntityProject, 25, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(12), sub.EntityID)
			assert.Equal(t, int64(3), sub.ID)
		})

		t.Run("project from parent's parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			service := NewSubscriptionService(s.Engine())
			// Project 26 belongs to project 25 which belongs to project 12 where user 6 has subscribed to
			sub, err := service.GetForUser(s, models.SubscriptionEntityProject, 26, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(12), sub.EntityID)
			assert.Equal(t, int64(3), sub.ID)
		})

		t.Run("task from parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			service := NewSubscriptionService(s.Engine())
			// Task 39 belongs to project 25 which belongs to project 12 where the user has subscribed
			sub, err := service.GetForUser(s, models.SubscriptionEntityTask, 39, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
		})

		t.Run("task from project", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			service := NewSubscriptionService(s.Engine())
			// Task 21 belongs to project 32 which the user has subscribed to
			sub, err := service.GetForUser(s, models.SubscriptionEntityTask, 21, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(8), sub.ID)
		})
	})

	t.Run("invalid type", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		_, err := service.GetForUser(s, 2342, 21, u)
		require.Error(t, err)
		assert.True(t, models.IsErrUnknownSubscriptionEntityType(err))
	})

	t.Run("double subscription should be returned once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		sub, err := service.GetForUser(s, models.SubscriptionEntityTask, 18, u)
		require.NoError(t, err)
		assert.Equal(t, int64(9), sub.ID)
	})
}

func TestSubscriptionService_NoCrossUserProjectInheritance(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	user1 := &user.User{ID: 1}
	user2 := &user.User{ID: 2}

	service := NewSubscriptionService(s.Engine())
	sub := &models.Subscription{
		Entity:     "project",
		EntityType: models.SubscriptionEntityProject,
		EntityID:   3,
	}

	err := service.Create(s, sub, user1)
	require.NoError(t, err)

	result, err := service.GetForUser(s, models.SubscriptionEntityTask, 32, user2)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestSubscriptionService_GetForEntities(t *testing.T) {
	t.Run("multiple projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntities(s, models.SubscriptionEntityProject, []int64{12, 32})
		require.NoError(t, err)
		assert.NotNil(t, subs)

		// Project 12 has subscription ID 3 for user 6
		assert.Len(t, subs[12], 1)
		assert.Equal(t, int64(3), subs[12][0].ID)

		// Project 32 has subscription ID 8 for user 6
		assert.Len(t, subs[32], 1)
		assert.Equal(t, int64(8), subs[32][0].ID)
	})

	t.Run("multiple tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntities(s, models.SubscriptionEntityTask, []int64{2, 22})
		require.NoError(t, err)
		assert.NotNil(t, subs)

		// Task 2 has subscription ID 1 for user 1
		assert.Len(t, subs[2], 1)
		assert.Equal(t, int64(1), subs[2][0].ID)

		// Task 22 has subscription ID 4 for user 6
		assert.Len(t, subs[22], 1)
		assert.Equal(t, int64(4), subs[22][0].ID)
	})
}

func TestSubscriptionService_GetForEntitiesAndUser(t *testing.T) {
	u := &user.User{ID: 6}

	t.Run("filter by user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntitiesAndUser(s, models.SubscriptionEntityProject, []int64{12}, u)
		require.NoError(t, err)
		assert.NotNil(t, subs)
		assert.Len(t, subs[12], 1)
		assert.Equal(t, int64(6), subs[12][0].UserID)
	})

	t.Run("no subscription for user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		u2 := &user.User{ID: 2}
		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntitiesAndUser(s, models.SubscriptionEntityProject, []int64{12}, u2)
		require.NoError(t, err)
		assert.NotNil(t, subs)
		// User 2 has no subscription to project 12
		assert.Len(t, subs[12], 0)
	})
}

func TestSubscriptionService_GetForEntity(t *testing.T) {
	t.Run("single project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntity(s, models.SubscriptionEntityProject, 12)
		require.NoError(t, err)
		assert.NotNil(t, subs)
		assert.Len(t, subs, 1)
		assert.Equal(t, int64(3), subs[0].ID)
	})

	t.Run("single task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntity(s, models.SubscriptionEntityTask, 22)
		require.NoError(t, err)
		assert.NotNil(t, subs)
		assert.Len(t, subs, 1)
		assert.Equal(t, int64(4), subs[0].ID)
	})

	t.Run("entity with no subscriptions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		defer s.Close()

		service := NewSubscriptionService(s.Engine())
		subs, err := service.GetForEntity(s, models.SubscriptionEntityTask, 1)
		require.NoError(t, err)
		// Should return nil or empty array if no subscriptions
		assert.True(t, subs == nil || len(subs) == 0)
	})
}
