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

type migrationStatusBefore20260830162731 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	UserID       int64     `xorm:"bigint not null"`
	MigratorName string    `xorm:"varchar(255)"`
	StartedAt    time.Time `xorm:"not null"`
	FinishedAt   time.Time `xorm:"null"`
}

func (migrationStatusBefore20260830162731) TableName() string {
	return "migration_status"
}

func TestAddActiveUserClaim20260830162731(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := migrationStatusBefore20260830162731{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	finishedAt := time.Now().Truncate(time.Second)
	existing := &migrationStatusBefore20260830162731{
		ID:           1,
		UserID:       42,
		MigratorName: "todoist",
		StartedAt:    finishedAt.Add(-time.Minute),
		FinishedAt:   finishedAt,
	}
	_, err = x.Insert(existing)
	require.NoError(t, err)

	before := migrationStatusTable20260830162731(t, x)
	require.NotEmpty(t, before.Indexes)
	require.NoError(t, addActiveUserClaim20260830162731(x))

	after := migrationStatusTable20260830162731(t, x)
	require.NotNil(t, after.GetColumn("active_user_id"))
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		require.Equal(t, index.Type, preserved.Type)
		require.Equal(t, index.Cols, preserved.Cols)
	}

	got := &migrationActiveUserClaim20260830162731{}
	found, err := x.ID(existing.ID).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, existing.UserID, got.UserID)
	require.Equal(t, existing.MigratorName, got.MigratorName)
	require.WithinDuration(t, existing.StartedAt, got.StartedAt, time.Second)
	require.WithinDuration(t, existing.FinishedAt, got.FinishedAt, time.Second)
	require.Nil(t, got.ActiveUserID)

	activeUserID := int64(42)
	_, err = x.Insert(&migrationActiveUserClaim20260830162731{
		ID:           2,
		UserID:       activeUserID,
		MigratorName: "trello",
		StartedAt:    time.Now(),
		ActiveUserID: &activeUserID,
	})
	require.NoError(t, err)
	_, err = x.Insert(&migrationActiveUserClaim20260830162731{
		ID:           3,
		UserID:       activeUserID,
		MigratorName: "microsoft-todo",
		StartedAt:    time.Now(),
		ActiveUserID: &activeUserID,
	})
	require.Error(t, err, "the unique index should reject a second active migration for the user")
}

func migrationStatusTable20260830162731(t *testing.T, x *xorm.Engine) *schemas.Table {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == "migration_status" {
			return table
		}
	}
	t.Fatal("migration_status table not found")
	return nil
}
