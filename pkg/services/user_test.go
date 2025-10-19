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
	"github.com/stretchr/testify/assert"
)

func TestUserService_GetUsersAndProxiesFromIDs(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	registry := NewServiceRegistry(db.GetEngine())
	us := registry.User()

	t.Run("should get users and proxy users from a list of ids", func(t *testing.T) {
		// User with ID 1 exists, Link Share with ID 1 exists and belongs to user 2
		ids := []int64{1, -1}
		users, err := us.GetUsersAndProxiesFromIDs(s, ids)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 2)

		// Check for user 1
		assert.Contains(t, users, int64(1))
		assert.Equal(t, int64(1), users[1].ID)
		assert.Equal(t, "user1", users[1].Username)

		// Check for proxy user from link share 1
		assert.Contains(t, users, int64(-1))
		assert.Equal(t, int64(-1), users[-1].ID)
		assert.Equal(t, "link-share-1", users[-1].Username)
		assert.Equal(t, "Link Share", users[-1].Name)
	})

	t.Run("should return an empty map for an empty list of ids", func(t *testing.T) {
		ids := []int64{}
		users, err := us.GetUsersAndProxiesFromIDs(s, ids)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 0)
	})

	t.Run("should handle non-existing users and link shares gracefully", func(t *testing.T) {
		ids := []int64{999, -999}
		users, err := us.GetUsersAndProxiesFromIDs(s, ids)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 0)
	})
}
