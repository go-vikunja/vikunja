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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

type usersBefore20261001040000 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	Username string `xorm:"varchar(250) not null unique"`
	Email    string `xorm:"varchar(250) null index"`
}

func (usersBefore20261001040000) TableName() string {
	return "users"
}

type usersAfter20261001040000 struct {
	ID                 int64  `xorm:"bigint autoincr not null unique pk"`
	Username           string `xorm:"varchar(250) not null unique"`
	Email              string `xorm:"varchar(250) null index"`
	MustChangePassword bool   `xorm:"bool not null default false"`
}

func (usersAfter20261001040000) TableName() string {
	return "users"
}

func TestAddUsersMustChangePassword20261001040000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := usersBefore20261001040000{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	_, err = x.Insert(&usersBefore20261001040000{ID: 1, Username: "existing", Email: "existing@example.com"})
	require.NoError(t, err)

	tableOf := func() *schemas.Table {
		tables, err := x.DBMetas()
		require.NoError(t, err)
		for _, tbl := range tables {
			if tbl.Name == "users" {
				return tbl
			}
		}
		t.Fatal("users table not found")
		return nil
	}

	before := tableOf()
	require.NotEmpty(t, before.Indexes)

	require.NoError(t, addUsersMustChangePassword20261001040000(x))

	after := tableOf()
	require.NotNil(t, after.GetColumn("must_change_password"), "migration did not add the column")
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		assert.Equal(t, index.Type, preserved.Type)
		assert.Equal(t, index.Cols, preserved.Cols)
	}

	got := &usersAfter20261001040000{}
	found, err := x.ID(1).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "existing", got.Username)
	assert.False(t, got.MustChangePassword, "existing rows default to false")

	_, err = x.Insert(&usersAfter20261001040000{ID: 2, Username: "existing"})
	require.Error(t, err, "the unique username index should still reject duplicates")
}
