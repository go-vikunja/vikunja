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
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type usersBefore20260930103000 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	Username string `xorm:"varchar(250) not null unique"`
	Email    string `xorm:"varchar(250) null index"`
}

func (usersBefore20260930103000) TableName() string {
	return "users"
}

// usersAfter20260930103000 reads the migrated rows back; the migration's own struct
// only declares the new columns.
type usersAfter20260930103000 struct {
	ID         int64  `xorm:"bigint autoincr not null unique pk"`
	Username   string `xorm:"varchar(250) not null unique"`
	Email      string `xorm:"varchar(250) null index"`
	ImportID   string `xorm:"varchar(64) null index"`
	JobTitle   string `xorm:"varchar(250) null"`
	Department string `xorm:"varchar(250) null"`
}

func (usersAfter20260930103000) TableName() string {
	return "users"
}

func TestAddUsersImportColumns20260930103000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := usersBefore20260930103000{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	_, err = x.Insert(&usersBefore20260930103000{ID: 1, Username: "existing", Email: "existing@example.com"})
	require.NoError(t, err)

	before := usersTable20260930103000(t, x)
	require.NotEmpty(t, before.Indexes)

	require.NoError(t, addUsersImportColumns20260930103000(x))

	after := usersTable20260930103000(t, x)
	for _, column := range []string{"import_id", "job_title", "department"} {
		require.NotNilf(t, after.GetColumn(column), "migration did not add column %s", column)
	}
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		assert.Equal(t, index.Type, preserved.Type)
		assert.Equal(t, index.Cols, preserved.Cols)
	}

	got := &usersAfter20260930103000{}
	found, err := x.ID(1).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "existing", got.Username)
	assert.Equal(t, "existing@example.com", got.Email)
	assert.Empty(t, got.ImportID)
	assert.Empty(t, got.JobTitle)
	assert.Empty(t, got.Department)

	// The unique username index must survive: a second row with the same username is rejected.
	_, err = x.Insert(&usersAfter20260930103000{ID: 2, Username: "existing"})
	require.Error(t, err, "the unique username index should still reject duplicates")
}

func usersTable20260930103000(t *testing.T, x *xorm.Engine) *schemas.Table {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == "users" {
			return table
		}
	}
	t.Fatal("users table not found")
	return nil
}
