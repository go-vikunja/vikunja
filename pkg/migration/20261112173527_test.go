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
	"xorm.io/xorm/schemas"
)

type bucketBefore20261112173527 struct {
	ID            int64  `xorm:"bigint autoincr not null unique pk"`
	Title         string `xorm:"text not null"`
	ProjectViewID int64  `xorm:"bigint not null"`
}

func (bucketBefore20261112173527) TableName() string { return "buckets" }

type labelBefore20261112173527 struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk"`
	Title       string `xorm:"varchar(250) not null"`
	CreatedByID int64  `xorm:"bigint not null"`
}

func (labelBefore20261112173527) TableName() string { return "labels" }

func TestLabelAndBucketLookupIndexes20261112173527(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	tables := []any{bucketBefore20261112173527{}, labelBefore20261112173527{}}
	t.Cleanup(func() { require.NoError(t, x.DropTables(tables...)) })
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))
	bucket := &bucketBefore20261112173527{ID: 1, Title: "bucket", ProjectViewID: 42}
	label := &labelBefore20261112173527{ID: 1, Title: "label", CreatedByID: 37}
	_, err = x.Insert(bucket, label)
	require.NoError(t, err)

	readTables := func() map[string]*schemas.Table {
		metas, err := x.DBMetas()
		require.NoError(t, err)
		result := map[string]*schemas.Table{}
		for _, table := range metas {
			if table.Name == "buckets" || table.Name == "labels" {
				result[table.Name] = table
			}
		}
		require.Len(t, result, 2)
		return result
	}
	before := readTables()
	idx := slices.IndexFunc(migrations, func(m *xormigrate.Migration) bool { return m.ID == "20261112173527" })
	require.NotEqual(t, -1, idx)
	require.NoError(t, migrations[idx].Migrate(x))
	after := readTables()
	for name, table := range before {
		for _, column := range table.ColumnsSeq() {
			require.NotNilf(t, after[name].GetColumn(column), "migration dropped column %s.%s", name, column)
		}
		require.NotEmpty(t, table.Indexes)
		for indexName, index := range table.Indexes {
			preserved, found := after[name].Indexes[indexName]
			require.Truef(t, found, "migration dropped index %s on %s", indexName, name)
			require.Equal(t, index.Type, preserved.Type)
			require.Equal(t, index.Cols, preserved.Cols)
		}
	}
	for table, column := range map[string]string{"buckets": "project_view_id", "labels": "created_by_id"} {
		found := false
		for _, index := range after[table].Indexes {
			found = found || (index.Type == schemas.IndexType && slices.Equal(index.Cols, []string{column}))
		}
		require.Truef(t, found, "no index on %s.%s", table, column)
	}
	gotBucket := &bucketBefore20261112173527{}
	found, err := x.ID(1).Get(gotBucket)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, bucket, gotBucket)
	gotLabel := &labelBefore20261112173527{}
	found, err = x.ID(1).Get(gotLabel)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, label, gotLabel)
}
