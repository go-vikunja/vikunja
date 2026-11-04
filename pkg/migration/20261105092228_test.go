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
	"slices"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type userBefore20261105092228 struct {
	ID           int64  `xorm:"bigint autoincr not null unique pk"`
	Username     string `xorm:"varchar(250) not null unique"`
	AvatarFileID int64  `xorm:"null"`
	ExportFileID int64  `xorm:"bigint null"`
}

func (userBefore20261105092228) TableName() string { return "users" }

type projectBefore20261105092228 struct {
	ID               int64 `xorm:"bigint autoincr not null unique pk"`
	OwnerID          int64 `xorm:"bigint index not null"`
	BackgroundFileID int64 `xorm:"null"`
}

func (projectBefore20261105092228) TableName() string { return "projects" }

type migrationStatusBefore20261105092228 struct {
	ID           int64  `xorm:"bigint autoincr not null unique pk"`
	UserID       int64  `xorm:"bigint not null"`
	UploadFileID *int64 `xorm:"bigint null"`
}

func (migrationStatusBefore20261105092228) TableName() string { return "migration_status" }

func TestFileReferenceIndexes20261105092228(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []any{
		userBefore20261105092228{},
		projectBefore20261105092228{},
		migrationStatusBefore20261105092228{},
	}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	fileID := int64(3)
	_, err = x.Insert(
		&userBefore20261105092228{ID: 1, Username: "a", AvatarFileID: 1, ExportFileID: 2},
		&projectBefore20261105092228{ID: 1, OwnerID: 1, BackgroundFileID: 4},
		&migrationStatusBefore20261105092228{ID: 1, UserID: 1, UploadFileID: &fileID},
	)
	require.NoError(t, err)

	before := tableMetas20261105092228(t, x)
	idx := slices.IndexFunc(migrations, func(m *xormigrate.Migration) bool { return m.ID == "20261105092228" })
	require.NotEqual(t, -1, idx)
	require.NoError(t, migrations[idx].Migrate(x))
	after := tableMetas20261105092228(t, x)

	for name, table := range before {
		for _, column := range table.ColumnsSeq() {
			require.NotNilf(t, after[name].GetColumn(column), "migration dropped column %s.%s", name, column)
		}
		for indexName, index := range table.Indexes {
			preserved, found := after[name].Indexes[indexName]
			require.Truef(t, found, "migration dropped index %s on %s", indexName, name)
			require.Equal(t, index.Type, preserved.Type)
			require.Equal(t, index.Cols, preserved.Cols)
		}
	}

	for table, column := range map[string]string{
		"users":            "avatar_file_id",
		"projects":         "background_file_id",
		"migration_status": "upload_file_id",
	} {
		require.Truef(t, hasIndexOn20261105092228(after[table], column), "no index on %s.%s", table, column)
	}
	require.True(t, hasIndexOn20261105092228(after["users"], "export_file_id"), "no index on users.export_file_id")

	got := &userBefore20261105092228{}
	found, err := x.ID(1).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(2), got.ExportFileID)
}

func hasIndexOn20261105092228(table *schemas.Table, column string) bool {
	for _, index := range table.Indexes {
		if slices.Equal(index.Cols, []string{column}) {
			return true
		}
	}
	return false
}

func tableMetas20261105092228(t *testing.T, x *xorm.Engine) map[string]*schemas.Table {
	t.Helper()
	metas, err := x.DBMetas()
	require.NoError(t, err)
	tables := map[string]*schemas.Table{}
	for _, table := range metas {
		tables[table.Name] = table
	}
	for _, name := range []string{"users", "projects", "migration_status"} {
		require.Containsf(t, tables, name, "%s table not found", name)
	}
	return tables
}
