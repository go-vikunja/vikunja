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

package migration

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type apiTokensBefore20260906180128 struct {
	ID             int64               `xorm:"bigint autoincr not null unique pk"`
	Title          string              `xorm:"not null"`
	TokenSalt      string              `xorm:"not null"`
	TokenHash      string              `xorm:"not null unique"`
	TokenLastEight string              `xorm:"not null index varchar(8)"`
	Permissions    map[string][]string `xorm:"json not null"`
	ExpiresAt      time.Time           `xorm:"not null"`
	Created        time.Time           `xorm:"created not null"`
	OwnerID        int64               `xorm:"bigint not null"`
}

func (apiTokensBefore20260906180128) TableName() string {
	return "api_tokens"
}

type apiTokensAfter20260906180128 struct {
	ID             int64               `xorm:"bigint autoincr not null unique pk"`
	Title          string              `xorm:"not null"`
	TokenSalt      string              `xorm:"null"`
	TokenHash      string              `xorm:"null unique"`
	TokenLastEight string              `xorm:"null index varchar(8)"`
	TokenSha256    *string             `xorm:"varchar(64) null unique"`
	Permissions    map[string][]string `xorm:"json not null"`
	ExpiresAt      time.Time           `xorm:"not null"`
	Created        time.Time           `xorm:"created not null"`
	OwnerID        int64               `xorm:"bigint not null"`
}

func (apiTokensAfter20260906180128) TableName() string {
	return "api_tokens"
}

func apiTokensIndex20260906180128(t *testing.T, x *xorm.Engine, name string) *schemas.Index {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name != "api_tokens" {
			continue
		}
		for _, index := range table.Indexes {
			// Dialects strip the "UQE_<table>_" prefix off model-style index names; XName puts it back.
			if index.XName("api_tokens") == name {
				return index
			}
		}
		return nil
	}
	t.Fatal("api_tokens table not found")
	return nil
}

func TestAPITokenSha25620260906180128(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := apiTokensBefore20260906180128{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table)) //nolint:forbidigo // test-local table

	expires := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	_, err = x.Insert([]*apiTokensBefore20260906180128{
		{
			ID:             1,
			Title:          "first",
			TokenSalt:      "salt1",
			TokenHash:      "hash1",
			TokenLastEight: "abcdefgh",
			Permissions:    map[string][]string{"tasks": {"read_all"}},
			ExpiresAt:      expires,
			OwnerID:        1,
		},
		{
			ID:             2,
			Title:          "second",
			TokenSalt:      "salt2",
			TokenHash:      "hash2",
			TokenLastEight: "12345678",
			Permissions:    map[string][]string{"projects": {"read_all"}},
			ExpiresAt:      expires,
			OwnerID:        2,
		},
	})
	require.NoError(t, err)

	before := []*apiTokensBefore20260906180128{}
	require.NoError(t, x.OrderBy("id").Find(&before))
	require.Len(t, before, 2)

	requireLegacyRowsIntact := func(t *testing.T) {
		t.Helper()
		after := []*apiTokensAfter20260906180128{}
		require.NoError(t, x.In("id", []int64{1, 2}).OrderBy("id").Find(&after))
		require.Len(t, after, 2)

		for i, old := range before {
			migrated := after[i]
			require.Equal(t, old.ID, migrated.ID)
			require.Equal(t, old.Title, migrated.Title)
			require.Equal(t, old.TokenSalt, migrated.TokenSalt)
			require.Equal(t, old.TokenHash, migrated.TokenHash)
			require.Equal(t, old.TokenLastEight, migrated.TokenLastEight)
			require.Equal(t, old.Permissions, migrated.Permissions)
			require.Equal(t, old.ExpiresAt.Unix(), migrated.ExpiresAt.Unix())
			require.Equal(t, old.Created.Unix(), migrated.Created.Unix())
			require.Equal(t, old.OwnerID, migrated.OwnerID)
			require.Nil(t, migrated.TokenSha256)
		}
	}

	require.NoError(t, addAPITokenSha25620260906180128(x))
	requireLegacyRowsIntact(t)

	index := apiTokensIndex20260906180128(t, x, "UQE_api_tokens_token_sha256")
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)

	require.NoError(t, addAPITokenSha25620260906180128(x))
	requireLegacyRowsIntact(t)

	insertNew := func(id int64, title, sha string) error {
		sess := x.NewSession()
		defer sess.Close()
		_, err := sess.Nullable("token_salt", "token_hash", "token_last_eight").
			Insert(&apiTokensAfter20260906180128{
				ID:          id,
				Title:       title,
				TokenSha256: &sha,
				Permissions: map[string][]string{"tasks": {"read_all"}},
				ExpiresAt:   expires,
				OwnerID:     1,
			})
		return err
	}

	require.NoError(t, insertNew(3, "sha only", "sha3"))
	require.NoError(t, insertNew(4, "sha only again", "sha4"))

	// A re-run after the upgrade must keep hashes of tokens created since — they have no legacy hash to fall back on.
	require.NoError(t, addAPITokenSha25620260906180128(x))
	requireLegacyRowsIntact(t)

	newRows := []*apiTokensAfter20260906180128{}
	require.NoError(t, x.In("id", []int64{3, 4}).OrderBy("id").Find(&newRows))
	require.Len(t, newRows, 2)
	require.NotNil(t, newRows[0].TokenSha256)
	require.Equal(t, "sha3", *newRows[0].TokenSha256)
	require.NotNil(t, newRows[1].TokenSha256)
	require.Equal(t, "sha4", *newRows[1].TokenSha256)

	require.Error(t, insertNew(5, "duplicate sha", "sha4"))
}
