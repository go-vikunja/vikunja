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

	"github.com/stretchr/testify/require"
)

type userFor20260908112000 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	Username string `xorm:"varchar(250) not null unique"`
}

func (userFor20260908112000) TableName() string {
	return "users"
}

func TestMigrateAtUsernames20260908112000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := userFor20260908112000{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table)) //nolint:forbidigo // test-local table

	insertUser := func(id int64, username string) {
		sess := x.NewSession()
		defer sess.Close()
		_, err := sess.Insert(&userFor20260908112000{
			ID:       id,
			Username: username,
		})
		require.NoError(t, err)
	}

	insertUser(1, "@alice")
	insertUser(2, "bob")
	insertUser(3, "@bob")
	insertUser(4, "@bob_1")
	insertUser(5, "@")
	insertUser(6, "charlie")

	require.NoError(t, migrateAtUsernames20260908112000(x))

	getUser := func(id int64) *userFor20260908112000 {
		sess := x.NewSession()
		defer sess.Close()
		u := &userFor20260908112000{ID: id}
		has, err := sess.Get(u)
		require.NoError(t, err)
		require.True(t, has)
		return u
	}

	require.Equal(t, "alice", getUser(1).Username)
	require.Equal(t, "bob", getUser(2).Username)
	require.Equal(t, "bob_1", getUser(3).Username)   // @bob -> bob exists (id 2) -> bob_1
	require.Equal(t, "bob_1_1", getUser(4).Username) // @bob_1 -> bob_1 exists (id 3) -> bob_1_1
	require.Equal(t, "user_5", getUser(5).Username)  // @ -> user_5
	require.Equal(t, "charlie", getUser(6).Username) // unaffected

	// Re-run should be a no-op
	require.NoError(t, migrateAtUsernames20260908112000(x))
	require.Equal(t, "alice", getUser(1).Username)
	require.Equal(t, "bob", getUser(2).Username)
}
