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

type user20261112082349 struct {
	ID                        int64  `xorm:"bigint autoincr not null unique pk"`
	OverdueTasksRemindersTime string `xorm:"varchar(5) not null default '09:00'"`
}

func (user20261112082349) TableName() string { return "users" }

func TestPadOverdueReminderTimes20261112082349(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, x.DropTables(user20261112082349{}))
	})
	require.NoError(t, x.DropTables(user20261112082349{}))
	require.NoError(t, x.Sync2(user20261112082349{}))

	_, err = x.Insert(
		&user20261112082349{ID: 1, OverdueTasksRemindersTime: "9:00"},
		&user20261112082349{ID: 2, OverdueTasksRemindersTime: "9:30"},
		&user20261112082349{ID: 3, OverdueTasksRemindersTime: "18:15"},
		&user20261112082349{ID: 4, OverdueTasksRemindersTime: "xx:yy"},
	)
	require.NoError(t, err)

	idx := slices.IndexFunc(migrations, func(m *xormigrate.Migration) bool { return m.ID == "20261112082349" })
	require.NotEqual(t, -1, idx)
	require.NoError(t, migrations[idx].Migrate(x))

	users := []*user20261112082349{}
	require.NoError(t, x.OrderBy("id").Find(&users))
	got := map[int64]string{}
	for _, u := range users {
		got[u.ID] = u.OverdueTasksRemindersTime
	}
	require.Equal(t, map[int64]string{
		1: "09:00",
		2: "09:30",
		3: "18:15",
		4: "xx:yy",
	}, got)
}
