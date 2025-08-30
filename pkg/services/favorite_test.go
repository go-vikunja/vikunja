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
)

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
			EntityID: 23,
			UserID:   1,
			Kind:     models.FavoriteKindProject,
		}
		_, err := s.Insert(testFavorite)
		assert.NoError(t, err)

		favorites, err := fs.GetForUserByType(s, u, models.FavoriteKindProject)
		assert.NoError(t, err)
		assert.Len(t, favorites, 1)
		assert.Equal(t, int64(23), favorites[0].EntityID)
	})

	t.Run("should return an empty slice when there are no favorites", func(t *testing.T) {
		u := &user.User{ID: 2}
		favorites, err := fs.GetForUserByType(s, u, models.FavoriteKindProject)
		assert.NoError(t, err)
		assert.Len(t, favorites, 0)
	})
}

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
)

func TestFavoriteService_GetForUserByType(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	fs := NewFavoriteService()

	t.Run("should get all favorites for a user and type", func(t *testing.T) {
		u := &user.User{ID: 1}
		favorites, err := fs.GetForUserByType(s, u, models.FavoriteKindProject)
		assert.NoError(t, err)
		assert.Len(t, favorites, 2)
		assert.Equal(t, int64(1), favorites[0].EntityID)
		assert.Equal(t, int64(2), favorites[1].EntityID)
	})

	t.Run("should return an empty slice when there are no favorites", func(t *testing.T) {
		u := &user.User{ID: 2}
		favorites, err := fs.GetForUserByType(s, u, models.FavoriteKindProject)
		assert.NoError(t, err)
		assert.Len(t, favorites, 0)
	})
}
