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
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type projectsBefore20261104142611 struct {
	ID      int64  `xorm:"bigint autoincr not null unique pk"`
	OwnerID int64  `xorm:"bigint INDEX not null"`
	Title   string `xorm:"varchar(250) not null"`
}

func (projectsBefore20261104142611) TableName() string {
	return "projects"
}

func projectTable20261104142611(t *testing.T, x *xorm.Engine) *schemas.Table {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == "projects" {
			return table
		}
	}
	t.Fatal("projects table not found")
	return nil
}

func TestAddProjectPurgeMarker20261104142611(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	table := projectsBefore20261104142611{}
	require.NoError(t, x.DropTables(table))
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.Sync2(table))
	existing := projectsBefore20261104142611{
		ID:      1,
		OwnerID: 42,
		Title:   "Existing project",
	}
	_, err = x.Insert(&existing)
	require.NoError(t, err)
	before := projectTable20261104142611(t, x)
	require.NotEmpty(t, before.Indexes)
	for _, m := range migrations {
		if m.ID == "20261104142611" {
			require.NoError(t, m.Migrate(x))
			break
		}
	}
	after := projectTable20261104142611(t, x)
	require.NotNil(t, after.GetColumn("last_purged_at"))
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		require.Equal(t, index.Type, preserved.Type)
		require.Equal(t, index.Cols, preserved.Cols)
	}
	got := projectsBefore20261104142611{}
	found, err := x.ID(existing.ID).Get(&got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, existing, got)
}
