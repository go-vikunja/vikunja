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

type taskRelationsBefore20261103165926 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID       int64     `xorm:"bigint not null"`
	OtherTaskID  int64     `xorm:"bigint not null"`
	RelationKind string    `xorm:"varchar(50) not null"`
	CreatedByID  int64     `xorm:"bigint not null"`
	Created      time.Time `xorm:"created not null"`
}

func (taskRelationsBefore20261103165926) TableName() string {
	return "task_relations"
}

func taskRelationsTable20261103165926(t *testing.T, x *xorm.Engine) *schemas.Table {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == "task_relations" {
			return table
		}
	}
	t.Fatal("task_relations table not found")
	return nil
}

func TestAddTaskRelationsOtherTaskKindIndex20261103165926(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := taskRelationsBefore20261103165926{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	_, err = x.Insert(&taskRelationsBefore20261103165926{TaskID: 1, OtherTaskID: 2, RelationKind: "subtask", CreatedByID: 1})
	require.NoError(t, err)

	before := taskRelationsTable20261103165926(t, x)
	require.NoError(t, addTaskRelationsOtherTaskKindIndex20261103165926(x))
	after := taskRelationsTable20261103165926(t, x)

	index, found := after.Indexes["other_task_kind"]
	require.True(t, found, "index other_task_kind missing")
	require.Equal(t, []string{"other_task_id", "relation_kind"}, index.Cols)

	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		require.Equal(t, index.Cols, preserved.Cols)
	}

	count, err := x.Count(&taskRelationsBefore20261103165926{})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

// DBs converted from sqlite carry relation_kind as TEXT, which mysql refuses to
// index without a key length (#4374).
func TestAddTaskRelationsOtherTaskKindIndexSkipsMysqlTextColumn20261103165926(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	if x.Dialect().URI().DBType != schemas.MYSQL {
		t.Skip("mysql-specific key length restriction")
	}

	table := taskRelationsBefore20261103165926{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	_, err = x.Exec("ALTER TABLE task_relations MODIFY relation_kind TEXT NOT NULL")
	require.NoError(t, err)

	require.NoError(t, addTaskRelationsOtherTaskKindIndex20261103165926(x))

	_, found := taskRelationsTable20261103165926(t, x).Indexes["other_task_kind"]
	require.False(t, found)
}
