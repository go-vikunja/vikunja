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

func TestReactionsService_Create(t *testing.T) {
	t.Run("create reaction successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)
		u := &user.User{ID: 1}

		reaction := &models.Reaction{
			EntityID:   1,
			EntityKind: models.ReactionKindTask,
			Value:      "🦙",
		}

		err := service.Create(s, reaction, u)
		require.NoError(t, err)
		assert.NotZero(t, reaction.ID)
		assert.Equal(t, int64(1), reaction.UserID)

		// Verify in database
		db.AssertExists(t, "reactions", map[string]interface{}{
			"entity_id":   1,
			"entity_kind": models.ReactionKindTask,
			"value":       "🦙",
			"user_id":     1,
		}, false)
	})

	t.Run("create duplicate reaction is idempotent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)
		u := &user.User{ID: 1}

		// Create first reaction
		reaction := &models.Reaction{
			EntityID:   1,
			EntityKind: models.ReactionKindTask,
			Value:      "👍",
		}

		err := service.Create(s, reaction, u)
		require.NoError(t, err)

		// Try to create same reaction again
		reaction2 := &models.Reaction{
			EntityID:   1,
			EntityKind: models.ReactionKindTask,
			Value:      "👍",
		}

		err = service.Create(s, reaction2, u)
		require.NoError(t, err) // Should not error, just be idempotent
	})

	t.Run("create comment reaction", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)
		u := &user.User{ID: 1}

		reaction := &models.Reaction{
			EntityID:   1,
			EntityKind: models.ReactionKindComment,
			Value:      "❤️",
		}

		err := service.Create(s, reaction, u)
		require.NoError(t, err)
		assert.NotZero(t, reaction.ID)

		// Verify in database
		db.AssertExists(t, "reactions", map[string]interface{}{
			"entity_id":   1,
			"entity_kind": models.ReactionKindComment,
			"value":       "❤️",
			"user_id":     1,
		}, false)
	})
}

func TestReactionsService_Delete(t *testing.T) {
	t.Run("delete own reaction", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		// Delete reaction (fixture has reaction ID 1: user 1, task 1, 👋)
		err := service.Delete(s, 1, 1, "👋", models.ReactionKindTask)
		require.NoError(t, err)

		// Verify deletion
		db.AssertMissing(t, "reactions", map[string]interface{}{
			"entity_id":   1,
			"entity_kind": models.ReactionKindTask,
			"value":       "👋",
			"user_id":     1,
		})
	})

	t.Run("delete nonexistent reaction", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		// Try to delete reaction that doesn't exist
		err := service.Delete(s, 1, 1, "🦙", models.ReactionKindTask)
		require.NoError(t, err) // Should not error
	})

	t.Run("cannot delete another user's reaction", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		// User 2 tries to delete user 1's reaction
		err := service.Delete(s, 1, 2, "👋", models.ReactionKindTask)
		require.NoError(t, err) // Deletion succeeds but nothing deleted

		// Verify reaction still exists
		db.AssertExists(t, "reactions", map[string]interface{}{
			"entity_id":   1,
			"entity_kind": models.ReactionKindTask,
			"value":       "👋",
			"user_id":     1,
		}, false)
	})
}

func TestReactionsService_GetAll(t *testing.T) {
	t.Run("get all reactions for task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		reactions, err := service.GetAll(s, 1, models.ReactionKindTask)
		require.NoError(t, err)
		assert.NotNil(t, reactions)

		// Fixture has one reaction for task 1
		assert.Len(t, reactions["👋"], 1)
		assert.Equal(t, int64(1), reactions["👋"][0].ID)
	})

	t.Run("get reactions for task with no reactions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		reactions, err := service.GetAll(s, 2, models.ReactionKindTask)
		require.NoError(t, err)
		assert.Empty(t, reactions)
	})

	t.Run("get comment reactions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		reactions, err := service.GetAll(s, 1, models.ReactionKindComment)
		require.NoError(t, err)
		// Assuming fixtures have comment reactions
		assert.NotNil(t, reactions)
	})
}

func TestReactionsService_AddReactionsToTasks(t *testing.T) {
	t.Run("add reactions to multiple tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		taskMap := map[int64]*models.Task{
			1: {ID: 1},
			2: {ID: 2},
		}

		err := service.AddReactionsToTasks(s, []int64{1, 2}, taskMap)
		require.NoError(t, err)

		// Task 1 has reactions
		assert.NotNil(t, taskMap[1].Reactions)
		assert.Len(t, taskMap[1].Reactions["👋"], 1)

		// Task 2 has no reactions (should be nil, not empty map)
		assert.Nil(t, taskMap[2].Reactions)
	})

	t.Run("add reactions to empty task list", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewReactionsService(testEngine)

		taskMap := map[int64]*models.Task{}

		err := service.AddReactionsToTasks(s, []int64{}, taskMap)
		require.NoError(t, err)
	})
}
