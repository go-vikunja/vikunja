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

	"github.com/golang-jwt/jwt/v5"
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
			ProjectID:  1,
			Permission: PermissionRead,
		}
		err := share.Create(s, doer)

		require.NoError(t, err)
		assert.NotEmpty(t, share.Hash)
		assert.NotEmpty(t, share.ID)
		assert.Equal(t, SharingTypeWithoutPassword, share.SharingType)
		require.NoError(t, s.Commit())
		db.AssertExists(t, "link_shares", map[string]interface{}{
			"id": share.ID,
		}, false)
	})
	t.Run("invalid permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID:  1,
			Permission: Permission(123),
		}
		err := share.Create(s, doer)

		require.Error(t, err)
		assert.True(t, IsErrInvalidPermission(err))
	})
	t.Run("password should be hashed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ProjectID:  1,
			Permission: PermissionRead,
			Password:   "somePassword",
		}
		err := share.Create(s, doer)

		require.NoError(t, err)
		assert.NotEmpty(t, share.Hash)
		assert.NotEmpty(t, share.ID)
		assert.Empty(t, share.Password)
		require.NoError(t, s.Commit())
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
	t.Run("should forbid read-only users from listing link shares", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 has only read access to project 3
		share := &LinkSharing{
			ProjectID: 3,
		}
		_, _, _, err := share.ReadAll(s, doer, "", 1, -1)
		require.Error(t, err)
		assert.True(t, IsErrGenericForbidden(err))
	})
	t.Run("should forbid write users from listing link shares", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 has write access to project 10
		share := &LinkSharing{
			ProjectID: 10,
		}
		_, _, _, err := share.ReadAll(s, doer, "", 1, -1)
		require.Error(t, err)
		assert.True(t, IsErrGenericForbidden(err))
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

func TestLinkSharing_toUser(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		share := &LinkSharing{
			ID:      1,
			Name:    "",
			Created: time.Now(),
			Updated: time.Now(),
		}

		user := share.toUser()

		assert.Equal(t, "link-share-1", user.Username)
		assert.Equal(t, "Link Share", user.Name)
		assert.Equal(t, int64(-1), user.ID)
	})

	t.Run("name provided", func(t *testing.T) {
		share := &LinkSharing{
			ID:      2,
			Name:    "My Test Share",
			Created: time.Now(),
			Updated: time.Now(),
		}

		user := share.toUser()

		assert.Equal(t, "link-share-2", user.Username)
		assert.Equal(t, "My Test Share (Link Share)", user.Name)
		assert.Equal(t, int64(-2), user.ID)
	})
}

func TestGetLinkShareFromClaims(t *testing.T) {
	// Mirrors NewLinkShareJWTAuthtoken, including the legacy `permission`
	// and `sharedByID` claims so the tests below can prove they're ignored.
	buildClaims := func(id int64, hash string, projectID int64, permission Permission, sharedByID int64) jwt.MapClaims {
		return jwt.MapClaims{
			"type":       float64(2), // AuthTypeLinkShare
			"id":         float64(id),
			"hash":       hash,
			"project_id": float64(projectID),
			"permission": float64(permission),
			"sharedByID": float64(sharedByID),
		}
	}

	t.Run("valid share returns DB values", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbShare, err := GetLinkShareByID(s, 1)
		require.NoError(t, err)

		claims := buildClaims(dbShare.ID, dbShare.Hash, dbShare.ProjectID, dbShare.Permission, dbShare.SharedByID)

		got, err := GetLinkShareFromClaims(s, claims)
		require.NoError(t, err)
		assert.Equal(t, dbShare.ID, got.ID)
		assert.Equal(t, dbShare.Hash, got.Hash)
		assert.Equal(t, dbShare.ProjectID, got.ProjectID)
		assert.Equal(t, dbShare.Permission, got.Permission)
		assert.Equal(t, dbShare.SharedByID, got.SharedByID)
	})

	t.Run("deleted share is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbShare, err := GetLinkShareByID(s, 1)
		require.NoError(t, err)
		claims := buildClaims(dbShare.ID, dbShare.Hash, dbShare.ProjectID, dbShare.Permission, dbShare.SharedByID)

		_, err = s.Where("id = ?", dbShare.ID).Delete(&LinkSharing{})
		require.NoError(t, err)

		_, err = GetLinkShareFromClaims(s, claims)
		require.Error(t, err)
		assert.True(t, IsErrLinkShareTokenInvalid(err),
			"expected ErrLinkShareTokenInvalid for deleted share, got %T: %v", err, err)
	})

	t.Run("permission downgrade takes effect immediately", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbShare, err := GetLinkShareByID(s, 3)
		require.NoError(t, err)
		require.Equal(t, PermissionAdmin, dbShare.Permission,
			"fixture precondition: share id=3 must start as admin")

		// Capture claims while the share is still admin, then downgrade.
		claims := buildClaims(dbShare.ID, dbShare.Hash, dbShare.ProjectID, PermissionAdmin, dbShare.SharedByID)

		_, err = s.Where("id = ?", dbShare.ID).Cols("permission").Update(&LinkSharing{Permission: PermissionRead})
		require.NoError(t, err)

		got, err := GetLinkShareFromClaims(s, claims)
		require.NoError(t, err)
		assert.Equal(t, PermissionRead, got.Permission,
			"permission must come from DB, not from the (stale) JWT claim")
	})

	t.Run("hash mismatch is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbShare, err := GetLinkShareByID(s, 1)
		require.NoError(t, err)

		claims := buildClaims(dbShare.ID, "not-the-real-hash", dbShare.ProjectID, dbShare.Permission, dbShare.SharedByID)

		_, err = GetLinkShareFromClaims(s, claims)
		require.Error(t, err)
		assert.True(t, IsErrLinkShareTokenInvalid(err))
	})

	t.Run("sharedByID comes from DB not from claim", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		dbShare, err := GetLinkShareByID(s, 1)
		require.NoError(t, err)

		// Bogus sharedByID in the claim must be ignored in favor of the DB value.
		claims := buildClaims(dbShare.ID, dbShare.Hash, dbShare.ProjectID, dbShare.Permission, 9999999)

		got, err := GetLinkShareFromClaims(s, claims)
		require.NoError(t, err)
		assert.Equal(t, dbShare.SharedByID, got.SharedByID)
	})

	t.Run("missing id claim is rejected", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		claims := jwt.MapClaims{
			"hash": "whatever",
		}
		_, err := GetLinkShareFromClaims(s, claims)
		require.Error(t, err)
		assert.True(t, IsErrLinkShareTokenInvalid(err))
	})

	t.Run("missing hash claim is rejected", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		claims := jwt.MapClaims{
			"id": float64(1),
		}
		_, err := GetLinkShareFromClaims(s, claims)
		require.Error(t, err)
		assert.True(t, IsErrLinkShareTokenInvalid(err))
	})
}
