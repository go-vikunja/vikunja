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
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func TestSubscriptionGetTypeFromString(t *testing.T) {
	t.Run("project", func(t *testing.T) {
		entityType := getEntityTypeFromString("project")
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityProject), entityType)
	})
	t.Run("task", func(t *testing.T) {
		entityType := getEntityTypeFromString("task")
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityTask), entityType)
	})
	t.Run("invalid", func(t *testing.T) {
		entityType := getEntityTypeFromString("someomejghsd")
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityUnknown), entityType)
	})
}

func TestSubscription_Create(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sb := &Subscription{
			Entity:   "task",
			EntityID: 1,
		}

		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)

		err = sb.Create(s, u)
		require.NoError(t, err)

		require.NoError(t, s.Commit())
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

		sb := &Subscription{
			Entity:   "task",
			EntityID: 2,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)

		err = sb.Create(s, u)
		require.Error(t, err)
		terr := &ErrSubscriptionAlreadyExists{}
		assert.ErrorAs(t, err, &terr)
	})
	t.Run("forbidden for link shares", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		linkShare := &LinkSharing{}

		sb := &Subscription{
			Entity:   "task",
			EntityID: 1,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, linkShare)
		require.Error(t, err)
		assert.False(t, can)
	})
	t.Run("nonexisting project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sb := &Subscription{
			Entity:   "project",
			EntityID: 99999999,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, u)
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
		assert.False(t, can)
	})
	t.Run("noneixsting task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sb := &Subscription{
			Entity:   "task",
			EntityID: 99999999,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
		assert.False(t, can)
	})
	t.Run("no permissions to see project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sb := &Subscription{
			Entity:   "project",
			EntityID: 20,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("no permissions to see task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sb := &Subscription{
			Entity:   "task",
			EntityID: 14,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("existing subscription for (entity_id, entity_type, user_id) ", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sb := &Subscription{
			Entity:   "task",
			EntityID: 2,
			UserID:   u.ID,
		}

		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)

		err = sb.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrSubscriptionAlreadyExists(err))
	})

	// TODO: Add tests to test triggering of notifications for subscribed things
}

func TestSubscription_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		sb := &Subscription{
			Entity:   "task",
			EntityID: 2,
			UserID:   u.ID,
		}

		can, err := sb.CanDelete(s, u)
		require.NoError(t, err)
		assert.True(t, can)

		err = sb.Delete(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
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

		linkShare := &LinkSharing{}

		sb := &Subscription{
			Entity:   "task",
			EntityID: 1,
			UserID:   1,
		}

		can, err := sb.CanDelete(s, linkShare)
		require.Error(t, err)
		assert.False(t, can)
	})
	t.Run("not owner of the subscription", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 2}
		sb := &Subscription{
			Entity:   "task",
			EntityID: 2,
			UserID:   u.ID,
		}

		can, err := sb.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestSubscriptionGet(t *testing.T) {
	u := &user.User{ID: 6}

	t.Run("test each individually", func(t *testing.T) {
		t.Run("project", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			sub, err := GetSubscriptionForUser(s, SubscriptionEntityProject, 12, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(3), sub.ID)
		})
		t.Run("task", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 22, u)
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

			// Project 25 belongs to project 12 where user 6 has subscribed to
			sub, err := GetSubscriptionForUser(s, SubscriptionEntityProject, 25, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(12), sub.EntityID)
			assert.Equal(t, int64(3), sub.ID)
		})
		t.Run("project from parent's parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Project 26 belongs to project 25 which belongs to project 12 where user 6 has subscribed to
			sub, err := GetSubscriptionForUser(s, SubscriptionEntityProject, 26, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(12), sub.EntityID)
			assert.Equal(t, int64(3), sub.ID)
		})
		t.Run("task from parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Task 39 belongs to project 25 which belongs to project 12 where the user has subscribed
			sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 39, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			// assert.Equal(t, int64(2), sub.ID) TODO
		})
		t.Run("task from project", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Task 21 belongs to project 32 which the user has subscribed to
			sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 21, u)
			require.NoError(t, err)
			assert.NotNil(t, sub)
			assert.Equal(t, int64(8), sub.ID)
		})
	})
	t.Run("invalid type", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetSubscriptionForUser(s, 2342, 21, u)
		require.Error(t, err)
		assert.True(t, IsErrUnknownSubscriptionEntityType(err))
	})
	t.Run("double subscription should be returned once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 18, u)
		require.NoError(t, err)
		assert.Equal(t, int64(9), sub.ID)
	})
	t.Run("soft-deleted task resolves no subscription", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Task 51 is soft-deleted; the raw CTE must not resolve its subscription (id 11)
		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 51, &user.User{ID: 1})
		require.NoError(t, err)
		assert.Nil(t, sub)
	})
}

func TestGetSubscriptionsForEntitySkipsUsersWithoutReadAccess(t *testing.T) {
	const (
		taskID     int64 = 32
		projectID  int64 = 3
		withAccess int64 = 2
		lostAccess int64 = 6
	)

	subscribeBoth := func(t *testing.T, s *xorm.Session, entityType SubscriptionEntityType, entityID int64) {
		for _, userID := range []int64{withAccess, lostAccess} {
			_, err := s.Insert(&Subscription{
				UserID:     userID,
				EntityType: entityType,
				EntityID:   entityID,
			})
			require.NoError(t, err)
		}
	}

	subscriberIDs := func(subs []*SubscriptionWithUser) (ids []int64) {
		for _, sub := range subs {
			ids = append(ids, sub.UserID)
		}
		return ids
	}

	t.Run("task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		subscribeBoth(t, s, SubscriptionEntityTask, taskID)

		subs, err := GetSubscriptionsForEntity(s, SubscriptionEntityTask, taskID)
		require.NoError(t, err)
		assert.Equal(t, []int64{withAccess}, subscriberIDs(subs))
	})
	t.Run("project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		subscribeBoth(t, s, SubscriptionEntityProject, projectID)

		subs, err := GetSubscriptionsForEntity(s, SubscriptionEntityProject, projectID)
		require.NoError(t, err)
		assert.Equal(t, []int64{withAccess}, subscriberIDs(subs))
	})
	t.Run("user only lookup is not filtered", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		subscribeBoth(t, s, SubscriptionEntityTask, taskID)
		subscribeBoth(t, s, SubscriptionEntityProject, projectID)

		subs, err := GetSubscriptionsForEntitiesAndUser(s, SubscriptionEntityTask, []int64{taskID}, &user.User{ID: lostAccess})
		require.NoError(t, err)
		assert.Equal(t, []int64{lostAccess}, subscriberIDs(subs[taskID]))

		subs, err = GetSubscriptionsForEntitiesAndUser(s, SubscriptionEntityProject, []int64{projectID}, &user.User{ID: lostAccess})
		require.NoError(t, err)
		assert.Equal(t, []int64{lostAccess}, subscriberIDs(subs[projectID]))
	})
}

func TestSubscription_NoCrossUserProjectInheritance(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	user1 := &user.User{ID: 1}
	user2 := &user.User{ID: 2}

	sb := &Subscription{
		Entity:   "project",
		EntityID: 3,
	}
	can, err := sb.CanCreate(s, user1)
	require.NoError(t, err)
	require.True(t, can)
	require.NoError(t, sb.Create(s, user1))

	sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 32, user2)
	require.NoError(t, err)
	assert.Nil(t, sub)
}

func TestSubscription_Mute(t *testing.T) {
	// User 6 is subscribed to project 32 (subscription 8), task 21 belongs to it
	u := &user.User{ID: 6}

	mute := func(t *testing.T, s *xorm.Session, a web.Auth, entity string, entityID int64) {
		sb := &Subscription{
			Entity:   entity,
			EntityID: entityID,
		}

		can, err := sb.CanDelete(s, a)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, sb.Delete(s, a))
	}

	t.Run("unsubscribing from a task subscribed through its project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 21, u)
		require.NoError(t, err)
		assert.Nil(t, sub)

		// The project subscription itself must stay untouched
		sub, err = GetSubscriptionForUser(s, SubscriptionEntityProject, 32, u)
		require.NoError(t, err)
		require.NotNil(t, sub)
		assert.Equal(t, int64(8), sub.ID)
	})
	t.Run("muted task is not notified", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 3 owns project 3, user 2 has it shared with them
		owner := &user.User{ID: 3}
		other := &user.User{ID: 2}
		for _, a := range []web.Auth{owner, other} {
			sb := &Subscription{Entity: "project", EntityID: 3}
			can, err := sb.CanCreate(s, a)
			require.NoError(t, err)
			require.True(t, can)
			require.NoError(t, sb.Create(s, a))
		}

		mute(t, s, other, "task", 32)

		subs, err := GetSubscriptionsForEntity(s, SubscriptionEntityTask, 32)
		require.NoError(t, err)
		require.Len(t, subs, 1)
		assert.Equal(t, owner.ID, subs[0].UserID)
	})
	t.Run("muting a project also mutes its children", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 25 is a child of project 12 which the user is subscribed to, project 26 is
		// a child of 25
		mute(t, s, u, "project", 25)

		for _, projectID := range []int64{25, 26} {
			sub, err := GetSubscriptionForUser(s, SubscriptionEntityProject, projectID, u)
			require.NoError(t, err)
			assert.Nil(t, sub)
		}

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityProject, 12, u)
		require.NoError(t, err)
		require.NotNil(t, sub)
		assert.Equal(t, int64(3), sub.ID)
	})
	t.Run("cannot unsubscribe twice", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)

		sb := &Subscription{
			Entity:   "task",
			EntityID: 21,
		}
		can, err := sb.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("needs read access to the entity", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 6 lost access to project 3, task 32 belongs to it
		_, err := s.Insert(&Subscription{
			UserID:     u.ID,
			EntityType: SubscriptionEntityProject,
			EntityID:   3,
		})
		require.NoError(t, err)

		inherited, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 32, u)
		require.NoError(t, err)
		require.NotNil(t, inherited)

		sb := &Subscription{
			Entity:   "task",
			EntityID: 32,
		}
		can, err := sb.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("removing an own subscription after losing access writes no opt-out", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 6 lost access to project 3, task 32 belongs to it
		for _, sub := range []*Subscription{
			{EntityType: SubscriptionEntityProject, EntityID: 3},
			{EntityType: SubscriptionEntityTask, EntityID: 32},
		} {
			sub.UserID = u.ID
			_, err := s.Insert(sub)
			require.NoError(t, err)
		}

		sb := &Subscription{
			Entity:   "task",
			EntityID: 32,
		}
		can, err := sb.CanDelete(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, sb.Delete(s, u))

		require.NoError(t, s.Commit())
		db.AssertMissing(t, "subscriptions", map[string]interface{}{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   32,
			"user_id":     u.ID,
		})
	})
	t.Run("subscribing again lifts the mute", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)

		sb := &Subscription{
			Entity:   "task",
			EntityID: 21,
		}
		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, sb.Create(s, u))

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 21, u)
		require.NoError(t, err)
		require.NotNil(t, sub)
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityTask), sub.EntityType)

		require.NoError(t, s.Commit())
		db.AssertExists(t, "subscriptions", map[string]interface{}{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   21,
			"user_id":     u.ID,
			"muted":       false,
		}, false)
		db.AssertCount(t, "subscriptions", builder.Eq{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   21,
			"user_id":     u.ID,
		}, 1)
	})
	t.Run("subscribing again refreshes created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)

		backdated := time.Now().Add(-42 * time.Hour).UTC()
		_, err := s.
			Where("entity_id = ? AND entity_type = ? AND user_id = ?", 21, SubscriptionEntityTask, u.ID).
			Cols("created").
			Update(&Subscription{Created: backdated})
		require.NoError(t, err)

		sb := &Subscription{
			Entity:   "task",
			EntityID: 21,
		}
		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, sb.Create(s, u))

		own, err := getOwnSubscription(s, SubscriptionEntityTask, 21, u.ID)
		require.NoError(t, err)
		require.NotNil(t, own)
		assert.True(t, own.Created.After(backdated.Add(time.Hour)), "created stayed at %s", own.Created)
	})
	t.Run("unsubscribing again after lifting the mute", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)

		sb := &Subscription{
			Entity:   "task",
			EntityID: 21,
		}
		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, sb.Create(s, u))

		mute(t, s, u, "task", 21)

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 21, u)
		require.NoError(t, err)
		assert.Nil(t, sub)

		require.NoError(t, s.Commit())
		db.AssertCount(t, "subscriptions", builder.Eq{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   21,
			"user_id":     u.ID,
		}, 1)
	})
	t.Run("unsubscribing from a task with an own subscription and a subscribed project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Subscription 9 is on task 18, subscription 10 on project 9 which task 18 belongs to
		mute(t, s, u, "task", 18)

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 18, u)
		require.NoError(t, err)
		assert.Nil(t, sub)

		require.NoError(t, s.Commit())
		db.AssertCount(t, "subscriptions", builder.Eq{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   18,
			"user_id":     u.ID,
		}, 1)
		db.AssertExists(t, "subscriptions", map[string]interface{}{
			"entity_type": SubscriptionEntityTask,
			"entity_id":   18,
			"user_id":     u.ID,
			"muted":       true,
		}, false)
	})
	t.Run("mute outlives the subscription it was made against", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)
		mute(t, s, u, "project", 32)

		sb := &Subscription{
			Entity:   "project",
			EntityID: 32,
		}
		can, err := sb.CanCreate(s, u)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, sb.Create(s, u))

		// A mute is user intent, so subscribing to the project again leaves the task muted
		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 21, u)
		require.NoError(t, err)
		assert.Nil(t, sub)
	})
	t.Run("being assigned does not lift the mute", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mute(t, s, u, "task", 21)

		ta := &TaskAssginee{TaskID: 21, UserID: u.ID}
		require.NoError(t, ta.Create(s, u))

		sub, err := GetSubscriptionForUser(s, SubscriptionEntityTask, 21, u)
		require.NoError(t, err)
		assert.Nil(t, sub)
	})
}
