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
	"strconv"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

type notificationsBefore20260728130118 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	NotifiableID int64     `xorm:"bigint not null"`
	Notification string    `xorm:"text not null"`
	Name         string    `xorm:"varchar(250) index not null"`
	SubjectID    int64     `xorm:"bigint null"`
	Created      time.Time `xorm:"created not null"`
}

func (notificationsBefore20260728130118) TableName() string {
	return "notifications"
}

type notificationsAfter20260728130118 struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64 `xorm:"bigint index not null default 0"`
}

func (notificationsAfter20260728130118) TableName() string {
	return "notifications"
}

type tasksFor20260728130118 struct {
	ID        int64      `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64      `xorm:"bigint not null"`
	Deleted   *time.Time `xorm:"datetime null deleted"`
}

func (tasksFor20260728130118) TableName() string {
	return "tasks"
}

func TestBackfillNotificationProjectIDs20260728130118(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	// x is the process-global test engine; these tables would leak into other tests.
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(notificationsBefore20260728130118{}, tasksFor20260728130118{}))
	})
	require.NoError(t, x.DropTables(notificationsBefore20260728130118{}, tasksFor20260728130118{}))
	require.NoError(t, x.Sync2(notificationsBefore20260728130118{}, tasksFor20260728130118{}))

	deletedAt := time.Now()
	_, err = x.Insert(
		&tasksFor20260728130118{ID: 10, ProjectID: 42},
		&tasksFor20260728130118{ID: 11, ProjectID: 43, Deleted: &deletedAt},
	)
	require.NoError(t, err)

	embedded := insertNotification20260728130118(t, x, "task.comment", `{"task":{"id":10,"project_id":42},"project":{"id":7,"title":"t"}}`)
	fromTask := insertNotification20260728130118(t, x, "task.assigned", `{"task":{"id":10,"project_id":42}}`)
	taskLookup := insertNotification20260728130118(t, x, "task.deleted", `{"task":{"id":10}}`)
	softDeletedTask := insertNotification20260728130118(t, x, "task.deleted", `{"task":{"id":11}}`)
	unknownTask := insertNotification20260728130118(t, x, "task.deleted", `{"task":{"id":999}}`)
	unparseable := insertNotification20260728130118(t, x, "task.comment", `{"task": "not an object"}`)
	accountScoped := insertNotification20260728130118(t, x, "team.member.added", `{"team":{"id":1}}`)
	unknownName := insertNotification20260728130118(t, x, "legacy.notification", `{"whatever":true}`)

	require.NoError(t, partialSync(x, notificationsProjectID20260728130118{}))
	require.NoError(t, backfillNotificationProjectIDs20260728130118(x))

	require.EqualValues(t, 7, projectIDOfNotification20260728130118(t, x, embedded),
		"the embedded project wins over the task's, since that is the data the payload carries")
	require.EqualValues(t, 42, projectIDOfNotification20260728130118(t, x, fromTask))
	require.EqualValues(t, 42, projectIDOfNotification20260728130118(t, x, taskLookup))
	require.EqualValues(t, 43, projectIDOfNotification20260728130118(t, x, softDeletedTask),
		"task.deleted's task is by definition soft-deleted, so the lookup must still find it")
	require.EqualValues(t, -1, projectIDOfNotification20260728130118(t, x, unknownTask))
	require.EqualValues(t, -1, projectIDOfNotification20260728130118(t, x, unparseable),
		"a project-scoped row that cannot be resolved must fail closed, not become account-scoped")
	require.EqualValues(t, 0, projectIDOfNotification20260728130118(t, x, accountScoped))
	require.EqualValues(t, 0, projectIDOfNotification20260728130118(t, x, unknownName),
		"a name this version does not know stays account-scoped rather than disappearing")

	require.NoError(t, backfillNotificationProjectIDs20260728130118(x))
	require.EqualValues(t, 7, projectIDOfNotification20260728130118(t, x, embedded))
	require.EqualValues(t, 43, projectIDOfNotification20260728130118(t, x, softDeletedTask))
}

func TestBackfillNotificationProjectIDsBatches20260728130118(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, x.DropTables(notificationsBefore20260728130118{}, tasksFor20260728130118{}))
	})
	require.NoError(t, x.DropTables(notificationsBefore20260728130118{}, tasksFor20260728130118{}))
	require.NoError(t, x.Sync2(notificationsBefore20260728130118{}, tasksFor20260728130118{}))

	total := notificationBackfillBatch20260728130118*2 + 3
	ids := make([]int64, 0, total)
	for i := 0; i < total; i++ {
		projectID := i%9 + 1
		ids = append(ids, insertNotification20260728130118(t, x, "task.comment",
			`{"project":{"id":`+strconv.Itoa(projectID)+`}}`))
	}

	require.NoError(t, partialSync(x, notificationsProjectID20260728130118{}))
	require.NoError(t, backfillNotificationProjectIDs20260728130118(x))

	for i, id := range ids {
		require.EqualValues(t, i%9+1, projectIDOfNotification20260728130118(t, x, id))
	}
}

func insertNotification20260728130118(t *testing.T, x *xorm.Engine, name, payload string) int64 {
	t.Helper()
	n := &notificationsBefore20260728130118{NotifiableID: 1, Name: name, Notification: payload}
	_, err := x.Insert(n)
	require.NoError(t, err)
	return n.ID
}

func projectIDOfNotification20260728130118(t *testing.T, x *xorm.Engine, id int64) int64 {
	t.Helper()
	n := &notificationsAfter20260728130118{}
	has, err := x.ID(id).Get(n)
	require.NoError(t, err)
	require.True(t, has)
	return n.ProjectID
}
