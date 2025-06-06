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

func TestLinkSharing_Create(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID: 1,
			Right:     RightRead,
		}
		err := share.Create(s, doer)

		require.NoError(t, err)
		assert.NotEmpty(t, share.Hash)
		assert.NotEmpty(t, share.ID)
		assert.Equal(t, SharingTypeWithoutPassword, share.SharingType)
		db.AssertExists(t, "link_shares", map[string]interface{}{
			"id": share.ID,
		}, false)
	})
	t.Run("invalid right", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID: 1,
			Right:     Right(123),
		}
		err := share.Create(s, doer)

		require.Error(t, err)
		assert.True(t, IsErrInvalidRight(err))
	})
	t.Run("password should be hashed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID: 1,
			Right:     RightRead,
			Password:  "somePassword",
		}
		err := share.Create(s, doer)

		require.NoError(t, err)
		assert.NotEmpty(t, share.Hash)
		assert.NotEmpty(t, share.ID)
		assert.Empty(t, share.Password)
		db.AssertExists(t, "link_shares", map[string]interface{}{
			"id":           share.ID,
			"sharing_type": SharingTypeWithPassword,
		}, false)
	})
}

func TestLinkSharing_ReadAll(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("all no password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID: 1,
		}
		all, _, _, err := share.ReadAll(s, doer, "", 1, -1)
		shares := all.([]*LinkSharing)

		require.NoError(t, err)
		assert.Len(t, shares, 2)
		for _, sharing := range shares {
			assert.Empty(t, sharing.Password)
		}
	})
	t.Run("search", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID: 1,
		}
		all, _, _, err := share.ReadAll(s, doer, "wITHPASS", 1, -1)
		shares := all.([]*LinkSharing)

		require.NoError(t, err)
		assert.Len(t, shares, 1)
		assert.Equal(t, int64(4), shares[0].ID)
	})
}

func TestLinkSharing_ReadOne(t *testing.T) {
	doer := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ID: 1,
		}
		err := share.ReadOne(s, doer)

		require.NoError(t, err)
		assert.NotEmpty(t, share.Hash)
		assert.Equal(t, SharingTypeWithoutPassword, share.SharingType)
	})
	t.Run("with password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ID: 4,
		}
		err := share.ReadOne(s, doer)

		require.NoError(t, err)
		assert.NotEmpty(t, share.Hash)
		assert.Equal(t, SharingTypeWithPassword, share.SharingType)
		assert.Empty(t, share.Password)
	})
}
