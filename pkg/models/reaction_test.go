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
)

func TestReaction_ReadAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}

		r := &Reaction{
			EntityID:         1,
			EntityKindString: "tasks",
		}

		reactions, _, _, err := r.ReadAll(s, u, "", 0, 0)
		require.NoError(t, err)
		assert.IsType(t, ReactionMap{}, reactions)

		reactionMap := reactions.(ReactionMap)
		assert.Len(t, reactionMap["👋"], 1)
		assert.Equal(t, int64(1), reactionMap["👋"][0].ID)
	})
	t.Run("invalid entity", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		r := &Reaction{
			EntityID:         1,
			EntityKindString: "loremipsum",
		}

		_, _, _, err := r.ReadAll(s, u, "", 0, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidReactionEntityKind{Kind: "loremipsum"})
	})
	t.Run("no access to task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}

		r := &Reaction{
			EntityID:         34,
			EntityKindString: "tasks",
		}

		_, _, _, err := r.ReadAll(s, u, "", 0, 0)
		require.Error(t, err)
	})
	t.Run("nonexistant task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		r := &Reaction{
			EntityID:         9999999,
			EntityKindString: "tasks",
		}

		_, _, _, err := r.ReadAll(s, u, "", 0, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTaskDoesNotExist{ID: r.EntityID})
	})
	t.Run("no access to comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}

		r := &Reaction{
			EntityID:         18,
			EntityKindString: "comments",
		}

		_, _, _, err := r.ReadAll(s, u, "", 0, 0)
		require.Error(t, err)
	})
	t.Run("nonexistant comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		r := &Reaction{
			EntityID:         9999999,
			EntityKindString: "comments",
		}

		_, _, _, err := r.ReadAll(s, u, "", 0, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTaskCommentDoesNotExist{ID: r.EntityID})
	})
}

func TestReaction_Create(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		r := &Reaction{
			EntityID:         1,
			EntityKindString: "tasks",
			Value:            "🦙",
		}

		can, err := r.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)

		err = r.Create(s, u)
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "reactions", map[string]interface{}{
			"entity_id":   r.EntityID,
			"entity_kind": ReactionKindTask,
			"user_id":     u.ID,
			"value":       r.Value,
		}, false)
	})
	t.Run("no permission to access task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		r := &Reaction{
			EntityID:         34,
			EntityKindString: "tasks",
			Value:            "🦙",
		}

		can, err := r.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("no permission to access comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		r := &Reaction{
			EntityID:         18,
			EntityKindString: "comments",
			Value:            "🦙",
		}

		can, err := r.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestReaction_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}

		r := &Reaction{
			EntityID:         1,
			EntityKindString: "tasks",
			Value:            "👋",
		}

		can, err := r.CanDelete(s, u)
		require.NoError(t, err)
		assert.True(t, can)

		err = r.Delete(s, u)
		require.NoError(t, err)

		db.AssertMissing(t, "reactions", map[string]interface{}{
			"entity_id":   r.EntityID,
			"entity_kind": ReactionKindTask,
			"value":       "👋",
		})
	})
}
