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
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/contexts"
)

type taskSelectCounter20260901001942 struct {
	count int
}

type taskIndexRowsStub20260901001942 struct {
	tasks    []taskProjectIndex20260901001942
	index    int
	scanErr  error
	rowsErr  error
	closeErr error
	closed   bool
}

func (r *taskIndexRowsStub20260901001942) Next() bool {
	return r.index < len(r.tasks)
}

func (r *taskIndexRowsStub20260901001942) Scan(beans ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	*(beans[0].(*taskProjectIndex20260901001942)) = r.tasks[r.index]
	r.index++
	return nil
}

func (r *taskIndexRowsStub20260901001942) Err() error {
	return r.rowsErr
}

func (r *taskIndexRowsStub20260901001942) Close() error {
	r.closed = true
	return r.closeErr
}

func (h *taskSelectCounter20260901001942) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return c.Ctx, nil
}

func (h *taskSelectCounter20260901001942) AfterProcess(c *contexts.ContextHook) error {
	sql := strings.ToLower(strings.TrimSpace(c.SQL))
	sql = strings.NewReplacer("`", "", `"`, "").Replace(sql)
	if strings.HasPrefix(sql, "select") && strings.Contains(sql, " from tasks") {
		h.count++
	}
	return nil
}

func TestCollectTaskIndexHighWaterMarks20260901001942(t *testing.T) {
	t.Run("keeps the first ordered index per project", func(t *testing.T) {
		rows := &taskIndexRowsStub20260901001942{tasks: []taskProjectIndex20260901001942{
			{ProjectID: 1, Index: 8},
			{ProjectID: 1, Index: 3},
			{ProjectID: 2, Index: 2},
		}}

		indexes, err := collectTaskIndexHighWaterMarks20260901001942(rows, 2)
		require.NoError(t, err)
		assert.Equal(t, map[int64]int64{1: 8, 2: 2}, indexes)
		assert.True(t, rows.closed)
	})

	for name, rows := range map[string]*taskIndexRowsStub20260901001942{
		"scan error":  {tasks: []taskProjectIndex20260901001942{{ProjectID: 1, Index: 8}}, scanErr: errors.New("scan")},
		"rows error":  {rowsErr: errors.New("rows")},
		"close error": {closeErr: errors.New("close")},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := collectTaskIndexHighWaterMarks20260901001942(rows, 1)
			require.EqualError(t, err, strings.TrimSuffix(name, " error"))
			assert.True(t, rows.closed)
		})
	}
}

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

	taskSelects := &taskSelectCounter20260901001942{}
	x.AddHook(taskSelects)
	require.NoError(t, addTaskIndexState20260901001942(x))
	assert.Equal(t, 1, taskSelects.count)

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
