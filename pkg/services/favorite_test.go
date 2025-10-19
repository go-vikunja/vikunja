// Vikunja is a to-do list application to facilitate your life.
// Copyright 2p018-present Vikunja and contributors. All rights reserved.
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

func TestFavoriteService_AddToFavorite(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should add entity to favorites", func(t *testing.T) {
		err := fs.AddToFavorite(s, 100, u, models.FavoriteKindTask)
		require.NoError(t, err)

		// Verify it was added
		isFav, err := fs.IsFavorite(s, 100, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, isFav)
	})

	t.Run("should handle nil auth", func(t *testing.T) {
		err := fs.AddToFavorite(s, 101, nil, models.FavoriteKindTask)
		require.NoError(t, err)

		// Verify it was not added
		exists, err := s.Where("entity_id = ? AND kind = ?", 101, models.FavoriteKindTask).Exist(&models.Favorite{})
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestFavoriteService_RemoveFromFavorite(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should remove entity from favorites", func(t *testing.T) {
		// First add a favorite
		err := fs.AddToFavorite(s, 200, u, models.FavoriteKindProject)
		require.NoError(t, err)

		// Then remove it
		err = fs.RemoveFromFavorite(s, 200, u, models.FavoriteKindProject)
		require.NoError(t, err)

		// Verify it was removed
		isFav, err := fs.IsFavorite(s, 200, u, models.FavoriteKindProject)
		require.NoError(t, err)
		assert.False(t, isFav)
	})

	t.Run("should handle removing non-existent favorite", func(t *testing.T) {
		err := fs.RemoveFromFavorite(s, 999, u, models.FavoriteKindTask)
		require.NoError(t, err)
	})

	t.Run("should handle nil auth", func(t *testing.T) {
		err := fs.RemoveFromFavorite(s, 201, nil, models.FavoriteKindTask)
		require.NoError(t, err)
	})
}

func TestFavoriteService_IsFavorite(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should return true for existing favorite", func(t *testing.T) {
		// Add a favorite
		err := fs.AddToFavorite(s, 300, u, models.FavoriteKindTask)
		require.NoError(t, err)

		// Check if it's a favorite
		isFav, err := fs.IsFavorite(s, 300, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, isFav)
	})

	t.Run("should return false for non-existent favorite", func(t *testing.T) {
		isFav, err := fs.IsFavorite(s, 999, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.False(t, isFav)
	})

	t.Run("should handle nil auth", func(t *testing.T) {
		isFav, err := fs.IsFavorite(s, 301, nil, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.False(t, isFav)
	})
}

func TestFavoriteService_GetFavoritesMap(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should return map of favorites", func(t *testing.T) {
		// Add multiple favorites
		err := fs.AddToFavorite(s, 400, u, models.FavoriteKindProject)
		require.NoError(t, err)
		err = fs.AddToFavorite(s, 401, u, models.FavoriteKindProject)
		require.NoError(t, err)

		// Get favorites map
		entityIDs := []int64{400, 401, 402}
		favMap, err := fs.GetFavoritesMap(s, entityIDs, u, models.FavoriteKindProject)
		require.NoError(t, err)

		assert.True(t, favMap[400])
		assert.True(t, favMap[401])
		assert.False(t, favMap[402])
	})

	t.Run("should return empty map for empty entity list", func(t *testing.T) {
		favMap, err := fs.GetFavoritesMap(s, []int64{}, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.Empty(t, favMap)
	})

	t.Run("should handle nil auth", func(t *testing.T) {
		favMap, err := fs.GetFavoritesMap(s, []int64{400, 401}, nil, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.Empty(t, favMap)
	})
}

func TestFavoriteService_GetForUserByType(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)

	t.Run("should get all favorites for a user and type", func(t *testing.T) {
		u := &user.User{ID: 1}

		// Note: The fixtures may not be loading properly for the favorites table,
		// so we manually insert a test favorite for now
		testFavorite := &models.Favorite{
			EntityID: 24,
			UserID:   1,
			Kind:     models.FavoriteKindProject,
		}
		_, err := s.Insert(testFavorite)
		assert.NoError(t, err)

		favorites, err := fs.GetForUserByType(s, u, models.FavoriteKindProject)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(favorites), 1) // At least the one we inserted

		// Check that our inserted favorite is in the results
		found := false
		for _, fav := range favorites {
			if fav.EntityID == 24 {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("should return an empty slice when there are no favorites", func(t *testing.T) {
		u := &user.User{ID: 99} // User with no favorites
		favorites, err := fs.GetForUserByType(s, u, models.FavoriteKindProject)
		assert.NoError(t, err)
		assert.Len(t, favorites, 0)
	})
}

func TestFavoriteService_DuplicateFavorites(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should handle duplicate favorite gracefully", func(t *testing.T) {
		// Add favorite first time
		err := fs.AddToFavorite(s, 500, u, models.FavoriteKindTask)
		require.NoError(t, err)

		// Try to add same favorite again
		err = fs.AddToFavorite(s, 500, u, models.FavoriteKindTask)
		// Should either succeed (idempotent) or fail gracefully
		// The important part is checking that only one entry exists

		// Verify only one entry exists in database
		count, err := s.Where("entity_id = ? AND user_id = ? AND kind = ?",
			500, u.ID, models.FavoriteKindTask).Count(&models.Favorite{})
		require.NoError(t, err)
		// Due to composite primary key, duplicate insert should fail, so we should have exactly 1
		assert.Equal(t, int64(1), count, "Should have exactly one favorite entry")

		// Verify it's still marked as favorite
		isFav, err := fs.IsFavorite(s, 500, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, isFav)
	})
}

func TestFavoriteService_MultipleUsers(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u1 := &user.User{ID: 1}
	u2 := &user.User{ID: 2}

	t.Run("should allow different users to favorite same entity", func(t *testing.T) {
		entityID := int64(600)

		// User 1 favorites entity
		err := fs.AddToFavorite(s, entityID, u1, models.FavoriteKindProject)
		require.NoError(t, err)

		// User 2 favorites same entity
		err = fs.AddToFavorite(s, entityID, u2, models.FavoriteKindProject)
		require.NoError(t, err)

		// Verify both users have it as favorite
		isFavU1, err := fs.IsFavorite(s, entityID, u1, models.FavoriteKindProject)
		require.NoError(t, err)
		assert.True(t, isFavU1, "Entity should be favorite for user 1")

		isFavU2, err := fs.IsFavorite(s, entityID, u2, models.FavoriteKindProject)
		require.NoError(t, err)
		assert.True(t, isFavU2, "Entity should be favorite for user 2")

		// Verify two separate entries exist
		count, err := s.Where("entity_id = ? AND kind = ?", entityID, models.FavoriteKindProject).
			Count(&models.Favorite{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count, "Should have two favorite entries (one per user)")
	})

	t.Run("should not affect other users when one removes favorite", func(t *testing.T) {
		entityID := int64(601)

		// Both users favorite the entity
		err := fs.AddToFavorite(s, entityID, u1, models.FavoriteKindTask)
		require.NoError(t, err)
		err = fs.AddToFavorite(s, entityID, u2, models.FavoriteKindTask)
		require.NoError(t, err)

		// User 1 removes favorite
		err = fs.RemoveFromFavorite(s, entityID, u1, models.FavoriteKindTask)
		require.NoError(t, err)

		// User 1 should not have it as favorite
		isFavU1, err := fs.IsFavorite(s, entityID, u1, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.False(t, isFavU1, "Entity should not be favorite for user 1")

		// User 2 should still have it as favorite
		isFavU2, err := fs.IsFavorite(s, entityID, u2, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, isFavU2, "Entity should still be favorite for user 2")
	})
}

func TestFavoriteService_KindIsolation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should isolate favorites by kind", func(t *testing.T) {
		entityID := int64(700)

		// Add entity as Task favorite
		err := fs.AddToFavorite(s, entityID, u, models.FavoriteKindTask)
		require.NoError(t, err)

		// Verify it's a favorite for Task kind
		isFavTask, err := fs.IsFavorite(s, entityID, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, isFavTask, "Entity should be favorite for Task kind")

		// Verify it's NOT a favorite for Project kind
		isFavProject, err := fs.IsFavorite(s, entityID, u, models.FavoriteKindProject)
		require.NoError(t, err)
		assert.False(t, isFavProject, "Entity should not be favorite for Project kind")
	})

	t.Run("should allow same entity to be favorite for different kinds", func(t *testing.T) {
		entityID := int64(701)

		// Add as both Task and Project favorite
		err := fs.AddToFavorite(s, entityID, u, models.FavoriteKindTask)
		require.NoError(t, err)
		err = fs.AddToFavorite(s, entityID, u, models.FavoriteKindProject)
		require.NoError(t, err)

		// Verify both kinds show as favorite
		isFavTask, err := fs.IsFavorite(s, entityID, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, isFavTask)

		isFavProject, err := fs.IsFavorite(s, entityID, u, models.FavoriteKindProject)
		require.NoError(t, err)
		assert.True(t, isFavProject)

		// Verify two separate entries exist
		count, err := s.Where("entity_id = ? AND user_id = ?", entityID, u.ID).
			Count(&models.Favorite{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count, "Should have two entries (one per kind)")
	})

	t.Run("should respect kind in GetFavoritesMap", func(t *testing.T) {
		entityIDs := []int64{702, 703, 704}

		// Add 702 and 704 as Task favorites
		err := fs.AddToFavorite(s, 702, u, models.FavoriteKindTask)
		require.NoError(t, err)
		err = fs.AddToFavorite(s, 704, u, models.FavoriteKindTask)
		require.NoError(t, err)

		// Add 703 as Project favorite (but not Task)
		err = fs.AddToFavorite(s, 703, u, models.FavoriteKindProject)
		require.NoError(t, err)

		// Get favorites map for Task kind
		favMapTask, err := fs.GetFavoritesMap(s, entityIDs, u, models.FavoriteKindTask)
		require.NoError(t, err)
		assert.True(t, favMapTask[702], "702 should be Task favorite")
		assert.False(t, favMapTask[703], "703 should not be Task favorite")
		assert.True(t, favMapTask[704], "704 should be Task favorite")

		// Get favorites map for Project kind
		favMapProject, err := fs.GetFavoritesMap(s, entityIDs, u, models.FavoriteKindProject)
		require.NoError(t, err)
		assert.False(t, favMapProject[702], "702 should not be Project favorite")
		assert.True(t, favMapProject[703], "703 should be Project favorite")
		assert.False(t, favMapProject[704], "704 should not be Project favorite")
	})
}

func TestFavoriteService_GetFavoritesMap_PartialMatches(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should handle mixed favorite status in GetFavoritesMap", func(t *testing.T) {
		// Add favorites for entities 800, 802, 804 (even numbers only)
		evenIDs := []int64{800, 802, 804}
		for _, id := range evenIDs {
			err := fs.AddToFavorite(s, id, u, models.FavoriteKindTask)
			require.NoError(t, err)
		}

		// Query for both favorited and non-favorited entities
		queryIDs := []int64{800, 801, 802, 803, 804}
		favMap, err := fs.GetFavoritesMap(s, queryIDs, u, models.FavoriteKindTask)
		require.NoError(t, err)

		// Verify correct status for each
		assert.True(t, favMap[800], "800 should be favorite")
		assert.False(t, favMap[801], "801 should not be favorite")
		assert.True(t, favMap[802], "802 should be favorite")
		assert.False(t, favMap[803], "803 should not be favorite")
		assert.True(t, favMap[804], "804 should be favorite")

		// Verify map only contains true entries (false is implicit)
		trueCount := 0
		for _, isFav := range favMap {
			if isFav {
				trueCount++
			}
		}
		assert.Equal(t, 3, trueCount, "Should have exactly 3 true entries")
	})

	t.Run("should return all false when no entities are favorited", func(t *testing.T) {
		queryIDs := []int64{900, 901, 902}
		favMap, err := fs.GetFavoritesMap(s, queryIDs, u, models.FavoriteKindProject)
		require.NoError(t, err)

		for _, id := range queryIDs {
			assert.False(t, favMap[id], "Entity %d should not be favorite", id)
		}
	})
}
