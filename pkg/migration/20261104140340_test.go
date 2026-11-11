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
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type taskAttachmentBefore20261104140340 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID      int64     `xorm:"bigint not null"`
	FileID      int64     `xorm:"bigint not null"`
	CreatedByID int64     `xorm:"bigint not null"`
	Created     time.Time `xorm:"created"`
}

func (taskAttachmentBefore20261104140340) TableName() string {
	return "task_attachments"
}

func TestTaskAttachmentFileIDIndex20261104140340(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := taskAttachmentBefore20261104140340{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	existing := &taskAttachmentBefore20261104140340{ID: 1, TaskID: 2, FileID: 3, CreatedByID: 4}
	_, err = x.Insert(existing)
	require.NoError(t, err)

	before := taskAttachmentsTable20261104140340(t, x)
	require.NoError(t, partialSync(x, taskAttachmentFileIDIndex20261104140340{}))
	after := taskAttachmentsTable20261104140340(t, x)

	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		require.Equal(t, index.Type, preserved.Type)
		require.Equal(t, index.Cols, preserved.Cols)
	}

	hasFileIDIndex := false
	for _, index := range after.Indexes {
		if slices.Equal(index.Cols, []string{"file_id"}) {
			hasFileIDIndex = true
		}
	}
	require.True(t, hasFileIDIndex, "no index on file_id")

	got := &taskAttachmentBefore20261104140340{}
	found, err := x.ID(existing.ID).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, existing.FileID, got.FileID)
}

func taskAttachmentsTable20261104140340(t *testing.T, x *xorm.Engine) *schemas.Table {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == "task_attachments" {
			return table
		}
	}
	t.Fatal("task_attachments table not found")
	return nil
}
