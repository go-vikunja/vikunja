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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

type tasksStub20261021192508 struct {
	ID          int64 `xorm:"autoincr not null unique pk"`
	ProjectID   int64 `xorm:"bigint not null"`
	CreatedByID int64 `xorm:"bigint not null"`
}

func (tasksStub20261021192508) TableName() string { return "tasks" }

type taskAssigneesStub20261021192508 struct {
	ID     int64 `xorm:"autoincr not null unique pk"`
	TaskID int64 `xorm:"bigint not null"`
	UserID int64 `xorm:"bigint not null"`
}

func (taskAssigneesStub20261021192508) TableName() string { return "task_assignees" }

type usersStub20261021192508 struct {
	ID int64 `xorm:"autoincr not null unique pk"`
}

func (usersStub20261021192508) TableName() string { return "users" }

type subscriptionsStub20261021192508 struct {
	ID         int64     `xorm:"autoincr not null unique pk"`
	EntityType int       `xorm:"not null unique(entity_user)"`
	EntityID   int64     `xorm:"bigint not null unique(entity_user)"`
	UserID     int64     `xorm:"bigint not null unique(entity_user)"`
	Muted      bool      `xorm:"not null default false"`
	Created    time.Time `xorm:"created not null"`
}

func (subscriptionsStub20261021192508) TableName() string { return "subscriptions" }

func TestSubscribeCreatorsAndAssigneesToTasks20261021192508(t *testing.T) {
	const subscriptionEntityProject = 2

	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []interface{}{
		tasksStub20261021192508{},
		taskAssigneesStub20261021192508{},
		usersStub20261021192508{},
		subscriptionsStub20261021192508{},
	}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	_, err = x.Insert(
		&usersStub20261021192508{ID: 1},
		&usersStub20261021192508{ID: 2},
		&tasksStub20261021192508{ID: 1, ProjectID: 1, CreatedByID: 1},
		&tasksStub20261021192508{ID: 2, ProjectID: 1, CreatedByID: 1},
		&tasksStub20261021192508{ID: 3, ProjectID: 1, CreatedByID: 2},
		&tasksStub20261021192508{ID: 4, ProjectID: 3, CreatedByID: 2},
		&tasksStub20261021192508{ID: 5, ProjectID: 3, CreatedByID: -1},
		&tasksStub20261021192508{ID: 6, ProjectID: 3, CreatedByID: 1},
		&taskAssigneesStub20261021192508{TaskID: 3, UserID: 1},
		// task 6: assignee who is also the creator gets one row
		&taskAssigneesStub20261021192508{TaskID: 6, UserID: 1},
		// task 4: assignee who is also the creator and muted the project
		&taskAssigneesStub20261021192508{TaskID: 4, UserID: 2},
		// task 99 does not exist
		&taskAssigneesStub20261021192508{TaskID: 99, UserID: 1},
		// user 3 does not exist
		&taskAssigneesStub20261021192508{TaskID: 1, UserID: 3},
		// task 2: creator opted out
		&subscriptionsStub20261021192508{EntityType: subscriptionEntityTask20261021192508, EntityID: 2, UserID: 1, Muted: true},
		// task 3: creator's project subscription doesn't prevent the task row
		&subscriptionsStub20261021192508{EntityType: subscriptionEntityProject, EntityID: 1, UserID: 2},
		// task 4: neither does a muted one
		&subscriptionsStub20261021192508{EntityType: subscriptionEntityProject, EntityID: 3, UserID: 2, Muted: true},
		// task 6: someone else's subscription doesn't count
		&subscriptionsStub20261021192508{EntityType: subscriptionEntityTask20261021192508, EntityID: 6, UserID: 2},
	)
	require.NoError(t, err)

	require.NoError(t, subscribeCreatorsAndAssigneesToTasks20261021192508(x))

	subs := []*subscriptionsStub20261021192508{}
	require.NoError(t, x.Where(builder.Eq{"entity_type": subscriptionEntityTask20261021192508}).OrderBy("entity_id, user_id").Find(&subs))

	type row struct {
		EntityID, UserID int64
		Muted            bool
	}
	got := make([]row, 0, len(subs))
	for _, s := range subs {
		got = append(got, row{s.EntityID, s.UserID, s.Muted})
	}
	assert.Equal(t, []row{
		{1, 1, false},
		{2, 1, true},
		{3, 1, false},
		{3, 2, false},
		{4, 2, false},
		{6, 1, false},
		{6, 2, false},
	}, got)
}
