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

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

type webhookBefore20261104201606 struct {
	ID        int64  `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64  `xorm:"bigint null index"`
	UserID    int64  `xorm:"bigint null index"`
	TargetURL string `xorm:"not null"`
}

func (webhookBefore20261104201606) TableName() string { return "webhooks" }

type webhookAfter20261104201606 struct {
	ID         int64
	ProjectID  int64
	UserID     int64
	TargetURL  string
	IsDisabled bool
}

func (webhookAfter20261104201606) TableName() string { return "webhooks" }

func TestWebhookDisabledMigration20261104201606(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	table := webhookBefore20261104201606{}
	require.NoError(t, x.DropTables(table))
	t.Cleanup(func() { require.NoError(t, x.DropTables(table)) })
	require.NoError(t, x.Sync2(table))
	_, err = x.Insert(&webhookBefore20261104201606{ID: 1, ProjectID: 42, TargetURL: "https://example.com/existing"})
	require.NoError(t, err)
	before, err := x.DBMetas()
	require.NoError(t, err)
	for _, migration := range migrations {
		if migration.ID == "20261104201606" {
			require.NoError(t, migration.Migrate(x))
		}
	}
	after, err := x.DBMetas()
	require.NoError(t, err)
	for _, oldTable := range before {
		if oldTable.Name != "webhooks" {
			continue
		}
		require.NotEmpty(t, oldTable.Indexes)
		for _, newTable := range after {
			if newTable.Name != "webhooks" {
				continue
			}
			require.NotNil(t, newTable.GetColumn("is_disabled"))
			for _, column := range oldTable.ColumnsSeq() {
				require.NotNil(t, newTable.GetColumn(column))
			}
			require.Equal(t, oldTable.Indexes, newTable.Indexes)
		}
	}
	stored := &webhookAfter20261104201606{}
	found, err := x.ID(1).Get(stored)
	require.NoError(t, err)
	require.True(t, found)
	require.False(t, stored.IsDisabled)
	require.Equal(t, int64(42), stored.ProjectID)
	require.Equal(t, "https://example.com/existing", stored.TargetURL)
}
