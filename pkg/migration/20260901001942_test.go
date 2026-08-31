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
)

type projectBefore20260901001942 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`
}

func (projectBefore20260901001942) TableName() string { return "projects" }

type taskBefore20260901001942 struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64     `xorm:"bigint not null"`
	Index     int64     `xorm:"bigint not null default 0 'index'"`
	DeletedAt time.Time `xorm:"deleted"`
}

func (taskBefore20260901001942) TableName() string { return "tasks" }

func TestAddTaskIndexState20260901001942(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []interface{}{
		projectBefore20260901001942{},
		taskBefore20260901001942{},
		ProjectTaskCounter20260901001942{},
		TaskIndexAlias20260901001942{},
	}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(projectBefore20260901001942{}, taskBefore20260901001942{}))

	_, err = x.Insert(
		&projectBefore20260901001942{ID: 1},
		&projectBefore20260901001942{ID: 2},
		&projectBefore20260901001942{ID: 3},
	)
	require.NoError(t, err)
	_, err = x.Insert(
		&taskBefore20260901001942{ID: 1, ProjectID: 1, Index: 3},
		&taskBefore20260901001942{ID: 2, ProjectID: 1, Index: 8, DeletedAt: time.Now()},
		&taskBefore20260901001942{ID: 3, ProjectID: 2, Index: 2},
	)
	require.NoError(t, err)

	require.NoError(t, addTaskIndexState20260901001942(x))

	counters := []*ProjectTaskCounter20260901001942{}
	require.NoError(t, x.OrderBy("project_id").Find(&counters))
	require.Equal(t, []*ProjectTaskCounter20260901001942{
		{ProjectID: 1, LastIndex: 8},
		{ProjectID: 2, LastIndex: 2},
		{ProjectID: 3, LastIndex: 0},
	}, counters)

	aliasCount, err := x.Count(&TaskIndexAlias20260901001942{})
	require.NoError(t, err)
	assert.Zero(t, aliasCount)

	_, err = x.Insert(&ProjectTaskCounter20260901001942{ProjectID: 1, LastIndex: 9})
	require.Error(t, err)
	_, err = x.Insert(&TaskIndexAlias20260901001942{ProjectID: 1, Index: 3, TaskID: 1})
	require.NoError(t, err)
	_, err = x.Insert(&TaskIndexAlias20260901001942{ProjectID: 1, Index: 3, TaskID: 2})
	require.Error(t, err)
}
