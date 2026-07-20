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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type usersPartial20260720120000 struct {
	BotOwnerID int64 `xorm:"bigint null index"`
}

func (usersPartial20260720120000) TableName() string {
	return "users"
}

func usersIndexOnUsername20260720120000(t *testing.T, x *xorm.Engine) *schemas.Index {
	t.Helper()
	indexes := usersIndexesOnUsername20260720120000(t, x)
	if len(indexes) == 0 {
		return nil
	}
	return indexes[0]
}

func TestRecreateMissingIndexes20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))
	require.NotNil(t, usersIndexOnUsername20260720120000(t, x))

	// A partial-struct Sync makes xorm drop every index the struct doesn't declare.
	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	_, err = x.Insert(&user.User{Username: "dup20260720120000"})
	require.NoError(t, err)
	_, err = x.Insert(&user.User{Username: "dup20260720120000"})
	require.NoError(t, err)

	err = recreateMissingIndexes20260720120000(x)
	require.ErrorContains(t, err, "users")
	require.ErrorContains(t, err, "username")

	_, err = x.Exec("DELETE FROM users WHERE username = ?", "dup20260720120000")
	require.NoError(t, err)

	require.NoError(t, recreateMissingIndexes20260720120000(x))
	index := usersIndexOnUsername20260720120000(t, x)
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)

	// Idempotent on a healthy schema.
	require.NoError(t, recreateMissingIndexes20260720120000(x))
}

func usersIndexesOnUsername20260720120000(t *testing.T, x *xorm.Engine) []*schemas.Index {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name != "users" {
			continue
		}
		indexes := make([]*schemas.Index, 0)
		for _, index := range table.Indexes {
			if len(index.Cols) == 1 && index.Cols[0] == "username" {
				indexes = append(indexes, index)
			}
		}
		return indexes
	}
	t.Fatal("users table not found")
	return nil
}

// An equivalent index under a different name (as pgloader produces) must not be
// duplicated by recreating the model's UQE_users_username index.
func TestRecreateMissingIndexesKeepsDifferentlyNamedIndex20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))

	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	_, err = x.Exec("CREATE UNIQUE INDEX some_nonmodel_name ON users (username)")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; leaving this index around leaks into later tests.
		_, err := x.Exec("DROP INDEX some_nonmodel_name")
		require.NoError(t, err)
	})

	require.NoError(t, recreateMissingIndexes20260720120000(x))

	indexes := usersIndexesOnUsername20260720120000(t, x)
	require.Len(t, indexes, 1)
	require.Equal(t, "some_nonmodel_name", indexes[0].Name)
}

func TestPartialSyncKeepsIndexes20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))
	require.NoError(t, recreateMissingIndexes20260720120000(x))

	require.NoError(t, partialSync(x, usersPartial20260720120000{}))

	index := usersIndexOnUsername20260720120000(t, x)
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)
}
