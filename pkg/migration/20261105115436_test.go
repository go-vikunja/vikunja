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

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
	"src.techknowlogick.com/xormigrate"
)

type project20261105115436 struct {
	ID                 int64  `xorm:"bigint autoincr not null unique pk"`
	BackgroundFileID   int64  `xorm:"null"`
	BackgroundBlurHash string `xorm:"varchar(50) null"`
}

func (project20261105115436) TableName() string { return "projects" }

type taskAttachment20261105115436 struct {
	ID     int64 `xorm:"bigint autoincr not null unique pk"`
	FileID int64 `xorm:"bigint not null"`
}

func (taskAttachment20261105115436) TableName() string { return "task_attachments" }

type user20261105115436 struct {
	ID           int64 `xorm:"bigint autoincr not null unique pk"`
	AvatarFileID int64 `xorm:"null"`
	ExportFileID int64 `xorm:"bigint null"`
}

func (user20261105115436) TableName() string { return "users" }

type migrationStatus20261105115436 struct {
	ID           int64  `xorm:"bigint autoincr not null unique pk"`
	UploadFileID *int64 `xorm:"bigint null"`
}

func (migrationStatus20261105115436) TableName() string { return "migration_status" }

func TestClearSharedBackgrounds20261105115436(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []any{
		project20261105115436{},
		taskAttachment20261105115436{},
		user20261105115436{},
		migrationStatus20261105115436{},
	}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	upload := int64(14)
	_, err = x.Insert(
		&taskAttachment20261105115436{ID: 1, FileID: 11},
		&user20261105115436{ID: 1, AvatarFileID: 12, ExportFileID: 13},
		&migrationStatus20261105115436{ID: 1, UploadFileID: &upload},
		&migrationStatus20261105115436{ID: 2},
		&project20261105115436{ID: 1, BackgroundFileID: 11, BackgroundBlurHash: "a"},
		&project20261105115436{ID: 2, BackgroundFileID: 12, BackgroundBlurHash: "b"},
		&project20261105115436{ID: 3, BackgroundFileID: 13, BackgroundBlurHash: "c"},
		&project20261105115436{ID: 4, BackgroundFileID: 14, BackgroundBlurHash: "d"},
		&project20261105115436{ID: 5, BackgroundFileID: 15, BackgroundBlurHash: "e"},
		&project20261105115436{ID: 6, BackgroundFileID: 15, BackgroundBlurHash: "f"},
		&project20261105115436{ID: 7, BackgroundFileID: 16, BackgroundBlurHash: "g"},
		&project20261105115436{ID: 8},
	)
	require.NoError(t, err)

	idx := slices.IndexFunc(migrations, func(m *xormigrate.Migration) bool { return m.ID == "20261105115436" })
	require.NotEqual(t, -1, idx)
	require.NoError(t, migrations[idx].Migrate(x))

	projects := []*project20261105115436{}
	require.NoError(t, x.OrderBy("id").Find(&projects))
	got := map[int64]int64{}
	for _, p := range projects {
		got[p.ID] = p.BackgroundFileID
		if p.BackgroundFileID == 0 {
			require.Emptyf(t, p.BackgroundBlurHash, "project %d kept its blur hash", p.ID)
		}
	}
	require.Equal(t, map[int64]int64{
		1: 0,  // task attachment
		2: 0,  // avatar
		3: 0,  // data export
		4: 0,  // pending import upload
		5: 15, // oldest of two projects sharing a file
		6: 0,
		7: 16, // own file
		8: 0,
	}, got)
}
