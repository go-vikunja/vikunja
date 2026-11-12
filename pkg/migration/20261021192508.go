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
	"fmt"
	"slices"
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type subscriptions20261021192508 struct {
	ID         int64     `xorm:"autoincr not null unique pk"`
	EntityType int       `xorm:"not null"`
	EntityID   int64     `xorm:"bigint not null"`
	UserID     int64     `xorm:"bigint not null"`
	Created    time.Time `xorm:"created not null"`
}

func (subscriptions20261021192508) TableName() string {
	return "subscriptions"
}

type subscriptionKey20261021192508 struct {
	ID     int64
	UserID int64
}

type tasks20261021192508 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	CreatedByID int64 `xorm:"bigint not null"`
}

func (tasks20261021192508) TableName() string {
	return "tasks"
}

type taskAssignees20261021192508 struct {
	TaskID int64 `xorm:"bigint INDEX not null"`
	UserID int64 `xorm:"bigint INDEX not null"`
}

func (taskAssignees20261021192508) TableName() string {
	return "task_assignees"
}

type users20261021192508 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`
}

func (users20261021192508) TableName() string {
	return "users"
}

const subscriptionEntityTask20261021192508 = 3

// 4 params per row stays under SQLite's 999 variable limit before 3.32.
const subscriptionBackfillBatch20261021192508 = 200

// An existing task row, muted or not, is the user's choice.
func subscribeCreatorsAndAssigneesToTasks20261021192508(tx *xorm.Engine) error {
	allTasks := []*tasks20261021192508{}
	if err := tx.Cols("id", "created_by_id").Find(&allTasks); err != nil {
		return fmt.Errorf("could not get tasks: %w", err)
	}
	taskExists := make(map[int64]bool, len(allTasks))
	for _, t := range allTasks {
		taskExists[t.ID] = true
	}

	assignees := []*taskAssignees20261021192508{}
	if err := tx.Cols("task_id", "user_id").Find(&assignees); err != nil {
		return fmt.Errorf("could not get task assignees: %w", err)
	}

	allUsers := []*users20261021192508{}
	if err := tx.Cols("id").Find(&allUsers); err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}
	userExists := make(map[int64]bool, len(allUsers))
	for _, u := range allUsers {
		userExists[u.ID] = true
	}

	taskSubs := []*subscriptions20261021192508{}
	err := tx.
		Cols("entity_id", "user_id").
		Where(builder.Eq{"entity_type": subscriptionEntityTask20261021192508}).
		Find(&taskSubs)
	if err != nil {
		return fmt.Errorf("could not get task subscriptions: %w", err)
	}
	taskSubscribed := make(map[subscriptionKey20261021192508]bool, len(taskSubs))
	for _, s := range taskSubs {
		taskSubscribed[subscriptionKey20261021192508{
			ID:     s.EntityID,
			UserID: s.UserID,
		}] = true
	}

	subs := []*subscriptions20261021192508{}
	add := func(taskID, userID int64) {
		taskKey := subscriptionKey20261021192508{
			ID:     taskID,
			UserID: userID,
		}
		if !userExists[userID] || taskSubscribed[taskKey] {
			return
		}
		taskSubscribed[taskKey] = true
		subs = append(subs, &subscriptions20261021192508{
			EntityType: subscriptionEntityTask20261021192508,
			EntityID:   taskID,
			UserID:     userID,
		})
	}
	for _, t := range allTasks {
		add(t.ID, t.CreatedByID)
	}
	for _, a := range assignees {
		if taskExists[a.TaskID] {
			add(a.TaskID, a.UserID)
		}
	}

	sess := tx.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return fmt.Errorf("could not start transaction to subscribe task creators and assignees: %w", err)
	}

	for batch := range slices.Chunk(subs, subscriptionBackfillBatch20261021192508) {
		if _, err := sess.Insert(&batch); err != nil {
			return fmt.Errorf("could not subscribe task creators and assignees: %w", err)
		}
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("could not commit task creator and assignee subscriptions: %w", err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261021192508",
		Description: "Subscribe creators and assignees to their existing tasks",
		Migrate:     subscribeCreatorsAndAssigneesToTasks20261021192508,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
