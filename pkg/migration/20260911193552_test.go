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
)

// The migration struct only declares the added column, so reads need the full row.
type migrationStatusErrorRow20260911193552 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	UserID       int64     `xorm:"bigint not null"`
	MigratorName string    `xorm:"varchar(255)"`
	StartedAt    time.Time `xorm:"not null"`
	FinishedAt   time.Time `xorm:"null"`
	ActiveUserID *int64    `xorm:"bigint null unique"`
	ErrorMessage string    `xorm:"text null"`
}

func (migrationStatusErrorRow20260911193552) TableName() string {
	return "migration_status"
}

func TestAddMigrationStatusError20260911193552(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := migrationStatusBefore20260830162731{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))
	require.NoError(t, addActiveUserClaim20260830162731(x))

	startedAt := time.Now().Truncate(time.Second)
	_, err = x.Insert(&migrationActiveUserClaim20260830162731{
		ID:           1,
		UserID:       42,
		MigratorName: "todoist",
		StartedAt:    startedAt,
		FinishedAt:   startedAt.Add(time.Minute),
	})
	require.NoError(t, err)

	before := migrationStatusTable20260830162731(t, x)
	require.NotEmpty(t, before.Indexes)
	require.NoError(t, addMigrationStatusError20260911193552(x))

	after := migrationStatusTable20260830162731(t, x)
	require.NotNil(t, after.GetColumn("error_message"))
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		require.Equal(t, index.Type, preserved.Type)
		require.Equal(t, index.Cols, preserved.Cols)
	}

	got := &migrationStatusErrorRow20260911193552{}
	found, err := x.ID(1).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(42), got.UserID)
	require.Equal(t, "todoist", got.MigratorName)
	require.WithinDuration(t, startedAt, got.StartedAt, time.Second)
	require.Empty(t, got.ErrorMessage, "rows migrated from before the column must not read as failed")

	activeUserID := int64(7)
	_, err = x.Insert(&migrationStatusErrorRow20260911193552{
		ID:           2,
		UserID:       activeUserID,
		MigratorName: "trello",
		StartedAt:    time.Now(),
		ActiveUserID: &activeUserID,
	})
	require.NoError(t, err)
	_, err = x.Insert(&migrationStatusErrorRow20260911193552{
		ID:           3,
		UserID:       activeUserID,
		MigratorName: "csv",
		StartedAt:    time.Now(),
		ActiveUserID: &activeUserID,
	})
	require.Error(t, err, "the unique index on active_user_id must survive the column addition")
}
