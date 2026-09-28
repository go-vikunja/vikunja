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

type labels20260928140648 struct {
	ID    int64  `xorm:"bigint autoincr not null unique pk"`
	Title string `xorm:"varchar(250) not null"`
}

func (labels20260928140648) TableName() string {
	return "labels"
}

type labelTasks20260928140648 struct {
	ID      int64 `xorm:"bigint autoincr not null unique pk"`
	TaskID  int64 `xorm:"bigint not null"`
	LabelID int64 `xorm:"bigint not null"`
}

func (labelTasks20260928140648) TableName() string {
	return "label_tasks"
}

func TestDeleteOrphanedLabelTasks20260928140648(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(labels20260928140648{}, labelTasks20260928140648{}))

	_, err = x.Insert(&labels20260928140648{ID: 1, Title: "still here"})
	require.NoError(t, err)
	_, err = x.Insert(
		&labelTasks20260928140648{ID: 1, TaskID: 1, LabelID: 1},
		&labelTasks20260928140648{ID: 2, TaskID: 1, LabelID: 42},
	)
	require.NoError(t, err)

	require.NoError(t, deleteOrphanedLabelTasks20260928140648(x))

	var remaining []*labelTasks20260928140648
	require.NoError(t, x.Find(&remaining))
	require.Len(t, remaining, 1)
	require.Equal(t, int64(1), remaining[0].LabelID)

	// Idempotent once there is nothing left to clean.
	require.NoError(t, deleteOrphanedLabelTasks20260928140648(x))
}
