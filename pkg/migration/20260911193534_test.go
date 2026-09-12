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

type migrationStatusBefore20260911193534 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	UserID       int64     `xorm:"bigint not null"`
	MigratorName string    `xorm:"varchar(255)"`
	StartedAt    time.Time `xorm:"not null"`
	FinishedAt   time.Time `xorm:"null"`
	ActiveUserID *int64    `xorm:"bigint null unique"`
}

func (migrationStatusBefore20260911193534) TableName() string {
	return "migration_status"
}

// The migration struct only declares the added column, so reads need the full row.
type migrationStatusHeartbeatRow20260911193534 struct {
	ID           int64      `xorm:"bigint autoincr not null unique pk"`
	UserID       int64      `xorm:"bigint not null"`
	MigratorName string     `xorm:"varchar(255)"`
	StartedAt    time.Time  `xorm:"not null"`
	FinishedAt   time.Time  `xorm:"null"`
	HeartbeatAt  *time.Time `xorm:"null"`
	ActiveUserID *int64     `xorm:"bigint null unique"`
}

func (migrationStatusHeartbeatRow20260911193534) TableName() string {
	return "migration_status"
}

func TestAddMigrationStatusHeartbeat20260911193534(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := migrationStatusBefore20260911193534{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table))

	activeUserID := int64(42)
	startedAt := time.Now().Truncate(time.Second)
	existing := &migrationStatusBefore20260911193534{
		ID:           1,
		UserID:       activeUserID,
		MigratorName: "todoist",
		StartedAt:    startedAt,
		ActiveUserID: &activeUserID,
	}
	_, err = x.Insert(existing)
	require.NoError(t, err)

	before := migrationStatusTable20260830162731(t, x)
	require.NotEmpty(t, before.Indexes)
	require.NoError(t, addMigrationStatusHeartbeat20260911193534(x))

	after := migrationStatusTable20260830162731(t, x)
	require.NotNil(t, after.GetColumn("heartbeat_at"))
	for _, column := range before.ColumnsSeq() {
		require.NotNilf(t, after.GetColumn(column), "migration dropped column %s", column)
	}
	for name, index := range before.Indexes {
		preserved, found := after.Indexes[name]
		require.Truef(t, found, "migration dropped index %s", name)
		require.Equal(t, index.Type, preserved.Type)
		require.Equal(t, index.Cols, preserved.Cols)
	}

	got := &migrationStatusHeartbeatRow20260911193534{}
	found, err := x.ID(existing.ID).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	require.WithinDuration(t, existing.StartedAt, got.StartedAt, time.Second)
	require.Equal(t, existing.MigratorName, got.MigratorName)
	require.NotNil(t, got.ActiveUserID)
	require.Equal(t, activeUserID, *got.ActiveUserID)
	require.Nil(t, got.HeartbeatAt, "existing claims must fall back to started_at")

	beatAt := time.Now().Truncate(time.Second)
	_, err = x.ID(existing.ID).Cols("heartbeat_at").Update(&migrationStatusHeartbeatRow20260911193534{HeartbeatAt: &beatAt})
	require.NoError(t, err)

	found, err = x.ID(existing.ID).Get(got)
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, got.HeartbeatAt)
	require.WithinDuration(t, beatAt, *got.HeartbeatAt, time.Second)
}
