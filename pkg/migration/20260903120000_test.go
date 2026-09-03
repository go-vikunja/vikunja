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
	"xorm.io/xorm"
)

type tasksStub20260903120000 struct {
	ID        int64     `xorm:"autoincr not null unique pk"`
	ProjectID int64     `xorm:"bigint not null"`
	Done      bool      `xorm:"null index(done_due_date)"`
	DueDate   time.Time `xorm:"DATETIME null index(done_due_date) 'due_date'"`
}

func (tasksStub20260903120000) TableName() string {
	return "tasks"
}

func tasksIndexNames20260903120000(t *testing.T, x *xorm.Engine) []string {
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name != "tasks" {
			continue
		}
		names := make([]string, 0, len(table.Indexes))
		for _, index := range table.Indexes {
			names = append(names, index.XName(table.Name))
		}
		return names
	}
	t.Fatal("tasks table not found")
	return nil
}

func TestSwapTasksDueDateIndex20260903120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []interface{}{tasksStub20260903120000{}}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))
	require.Contains(t, tasksIndexNames20260903120000(t, x), "IDX_tasks_done_due_date")

	require.NoError(t, swapTasksDueDateIndex20260903120000(x))

	names := tasksIndexNames20260903120000(t, x)
	assert.Contains(t, names, "IDX_tasks_project_done_due_date")
	assert.NotContains(t, names, "IDX_tasks_done_due_date")

	// Installs that already lost the old index must not abort the migration.
	require.NoError(t, swapTasksDueDateIndex20260903120000(x))
}
