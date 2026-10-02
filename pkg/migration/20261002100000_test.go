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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

type tasksBefore20261002100000 struct {
	ID        int64  `xorm:"bigint autoincr not null unique pk"`
	Title     string `xorm:"varchar(250) not null"`
	ProjectID int64  `xorm:"bigint not null index"`
}

func (tasksBefore20261002100000) TableName() string { return "tasks" }

type tasksAfter20261002100000 struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk"`
	Title       string `xorm:"varchar(250) not null"`
	ProjectID   int64  `xorm:"bigint not null index"`
	IsMilestone bool   `xorm:"bool not null default false"`
}

func (tasksAfter20261002100000) TableName() string { return "tasks" }

func TestAddTasksIsMilestone20261002100000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := tasksBefore20261002100000{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table, TaskBaseline20261002100000{}))
	})
	require.NoError(t, x.DropTables(table, TaskBaseline20261002100000{}))
	require.NoError(t, x.Sync2(table))
	_, err = x.Insert(&tasksBefore20261002100000{ID: 1, Title: "existing", ProjectID: 1})
	require.NoError(t, err)

	tableOf := func() *schemas.Table {
		tables, err := x.DBMetas()
		require.NoError(t, err)
		for _, tbl := range tables {
			if tbl.Name == "tasks" {
				return tbl
			}
		}
		t.Fatal("tasks table not found")
		return nil
	}
	before := tableOf()
	require.NotEmpty(t, before.Indexes)

	require.NoError(t, addTasksIsMilestone20261002100000(x))

	after := tableOf()
	require.NotNil(t, after.GetColumn("is_milestone"))
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		assert.Equal(t, index.Cols, preserved.Cols)
	}

	got := &tasksAfter20261002100000{}
	found, err := x.ID(1).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	assert.False(t, got.IsMilestone, "existing rows default to false")
}

func TestAddTaskBaselines20261002100000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(TaskBaseline20261002100000{}))
	})
	require.NoError(t, x.DropTables(TaskBaseline20261002100000{}))

	require.NoError(t, addTaskBaselines20261002100000(x))

	now := time.Now()
	_, err = x.Insert(&TaskBaseline20261002100000{TaskID: 1, ProjectID: 1, StartDate: now, EndDate: now, SavedAt: now, SavedByID: 1})
	require.NoError(t, err)
	// One baseline row per task.
	_, err = x.Insert(&TaskBaseline20261002100000{TaskID: 1, ProjectID: 1, StartDate: now, EndDate: now, SavedAt: now, SavedByID: 1})
	require.Error(t, err)
}
