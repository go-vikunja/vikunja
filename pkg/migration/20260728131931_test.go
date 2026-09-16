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
)

type notificationsFor20260728131931 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	NotifiableID int64     `xorm:"bigint not null"`
	Notification string    `xorm:"text not null"`
	Name         string    `xorm:"varchar(250) index not null"`
	SubjectID    int64     `xorm:"bigint null"`
	ProjectID    int64     `xorm:"bigint index not null default 0"`
	Created      time.Time `xorm:"created not null"`
}

func (notificationsFor20260728131931) TableName() string {
	return "notifications"
}

func TestDeleteUnknownNotifications20260728131931(t *testing.T) {
	x := prepareNotificationsTable20260728131931(t)

	listCreated := insertNotification20260728131931(t, x, "list.created",
		`{"doer":{"id":1},"list":{"id":7,"title":"secret","description":"internal"}}`)
	garbage := insertNotification20260728131931(t, x, "¯\\_(ツ)_/¯", `not even json`)
	empty := insertNotification20260728131931(t, x, "", `{}`)

	known := make(map[string]int64, len(knownNotifications20260728131931))
	for _, name := range knownNotifications20260728131931 {
		known[name] = insertNotification20260728131931(t, x, name, `{"task":{"id":1}}`)
	}

	require.NoError(t, deleteUnknownNotifications20260728131931(x))

	require.False(t, notificationExists20260728131931(t, x, listCreated),
		"a list.created row from before the project rename must be gone")
	require.False(t, notificationExists20260728131931(t, x, garbage))
	require.False(t, notificationExists20260728131931(t, x, empty))
	for name, id := range known {
		require.True(t, notificationExists20260728131931(t, x, id),
			"%q is a known name and must survive", name)
	}

	require.NoError(t, deleteUnknownNotifications20260728131931(x))
	require.EqualValues(t, len(known), countNotifications20260728131931(t, x),
		"re-running the migration must not delete anything else")
}

func TestDeleteUnknownNotificationsBatches20260728131931(t *testing.T) {
	x := prepareNotificationsTable20260728131931(t)

	unknown := 0
	for i := 0; i < unknownNotificationBatch20260728131931*2+3; i++ {
		insertNotification20260728131931(t, x, "list.created", `{"list":{"id":1}}`)
		unknown++
		if i%10 == 0 {
			insertNotification20260728131931(t, x, "task.comment", `{"task":{"id":1}}`)
		}
	}

	before := countNotifications20260728131931(t, x)
	require.NoError(t, deleteUnknownNotifications20260728131931(x))
	require.Equal(t, before-int64(unknown), countNotifications20260728131931(t, x))
}

func prepareNotificationsTable20260728131931(t *testing.T) *xorm.Engine {
	t.Helper()
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	// x is the process-global test engine; this table would leak into other tests.
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(notificationsFor20260728131931{}))
	})
	require.NoError(t, x.DropTables(notificationsFor20260728131931{}))
	require.NoError(t, x.Sync2(notificationsFor20260728131931{}))

	return x
}

func insertNotification20260728131931(t *testing.T, x *xorm.Engine, name, payload string) int64 {
	t.Helper()
	n := &notificationsFor20260728131931{NotifiableID: 1, Name: name, Notification: payload}
	_, err := x.Insert(n)
	require.NoError(t, err)
	return n.ID
}

func notificationExists20260728131931(t *testing.T, x *xorm.Engine, id int64) bool {
	t.Helper()
	has, err := x.ID(id).Get(&notificationsFor20260728131931{})
	require.NoError(t, err)
	return has
}

func countNotifications20260728131931(t *testing.T, x *xorm.Engine) int64 {
	t.Helper()
	count, err := x.Table("notifications").Count()
	require.NoError(t, err)
	return count
}
