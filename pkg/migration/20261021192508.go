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
	EntityType int
	EntityID   int64
	UserID     int64
}

type projects20261021192508 struct {
	ID              int64  `xorm:"bigint autoincr not null unique pk"`
	ParentProjectID *int64 `xorm:"bigint INDEX null"`
}

func (projects20261021192508) TableName() string {
	return "projects"
}

type tasks20261021192508 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	ProjectID   int64 `xorm:"bigint INDEX not null"`
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

const (
	subscriptionEntityProject20261021192508 = 2
	subscriptionEntityTask20261021192508    = 3
)

// Any existing task or ancestor-project row, muted or not, is the user's choice.
func subscribeCreatorsAndAssigneesToTasks20261021192508(tx *xorm.Engine) error {
	allProjects := []*projects20261021192508{}
	if err := tx.Cols("id", "parent_project_id").Find(&allProjects); err != nil {
		return fmt.Errorf("could not get projects: %w", err)
	}
	parents := make(map[int64]int64, len(allProjects))
	for _, p := range allProjects {
		parents[p.ID] = 0
		if p.ParentProjectID != nil {
			parents[p.ID] = *p.ParentProjectID
		}
	}

	allTasks := []*tasks20261021192508{}
	if err := tx.Cols("id", "project_id", "created_by_id").OrderBy("id").Find(&allTasks); err != nil {
		return fmt.Errorf("could not get tasks: %w", err)
	}
	taskProjects := make(map[int64]int64, len(allTasks))
	for _, t := range allTasks {
		taskProjects[t.ID] = t.ProjectID
	}

	assignees := []*taskAssignees20261021192508{}
	if err := tx.Cols("task_id", "user_id").OrderBy("task_id, user_id").Find(&assignees); err != nil {
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

	existing := []*subscriptions20261021192508{}
	err := tx.
		Cols("entity_type", "entity_id", "user_id").
		In("entity_type", subscriptionEntityProject20261021192508, subscriptionEntityTask20261021192508).
		Find(&existing)
	if err != nil {
		return fmt.Errorf("could not get subscriptions: %w", err)
	}
	subscribed := make(map[subscriptionKey20261021192508]bool, len(existing))
	for _, s := range existing {
		subscribed[subscriptionKey20261021192508{
			EntityType: s.EntityType,
			EntityID:   s.EntityID,
			UserID:     s.UserID,
		}] = true
	}

	ancestorCache := make(map[int64][]int64)
	ancestorsOf := func(projectID int64) []int64 {
		if cached, ok := ancestorCache[projectID]; ok {
			return cached
		}
		ancestors := []int64{}
		visited := map[int64]bool{}
		for current := projectID; current > 0 && !visited[current]; {
			parent, exists := parents[current]
			if !exists {
				break
			}
			visited[current] = true
			ancestors = append(ancestors, current)
			current = parent
		}
		ancestorCache[projectID] = ancestors
		return ancestors
	}

	isSubscribed := func(taskID, userID int64) bool {
		if subscribed[subscriptionKey20261021192508{
			EntityType: subscriptionEntityTask20261021192508,
			EntityID:   taskID,
			UserID:     userID,
		}] {
			return true
		}
		for _, projectID := range ancestorsOf(taskProjects[taskID]) {
			if subscribed[subscriptionKey20261021192508{
				EntityType: subscriptionEntityProject20261021192508,
				EntityID:   projectID,
				UserID:     userID,
			}] {
				return true
			}
		}
		return false
	}

	subs := []*subscriptions20261021192508{}
	seen := make(map[subscriptionKey20261021192508]bool)
	add := func(taskID, userID int64) {
		key := subscriptionKey20261021192508{
			EntityType: subscriptionEntityTask20261021192508,
			EntityID:   taskID,
			UserID:     userID,
		}
		if seen[key] || !userExists[userID] || isSubscribed(taskID, userID) {
			return
		}
		seen[key] = true
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
		if _, exists := taskProjects[a.TaskID]; exists {
			add(a.TaskID, a.UserID)
		}
	}

	sess := tx.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return fmt.Errorf("could not start transaction to subscribe task creators and assignees: %w", err)
	}

	// 4 params per row stays under SQLite's 999 variable limit before 3.32.
	const batchSize = 200
	for batch := range slices.Chunk(subs, batchSize) {
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
