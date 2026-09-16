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
	"encoding/json"
	"fmt"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type notificationsProjectID20260728130118 struct {
	ProjectID int64 `xorm:"bigint index not null default 0"`
}

func (notificationsProjectID20260728130118) TableName() string {
	return "notifications"
}

type notificationToBackfill20260728130118 struct {
	ID           int64  `xorm:"bigint autoincr not null unique pk"`
	Notification string `xorm:"text not null"`
}

func (notificationToBackfill20260728130118) TableName() string {
	return "notifications"
}

// No deleted column on purpose: task.deleted's task is soft-deleted, and a
// deleted tag would make xorm filter it away.
type taskProject20260728130118 struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64 `xorm:"bigint not null"`
}

func (taskProject20260728130118) TableName() string {
	return "tasks"
}

// Hardcoded rather than read from the notification registry: a migration is
// frozen in time, the registry grows.
var projectScopedNotifications20260728130118 = []string{
	"task.reminder",
	"task.comment",
	"task.assigned",
	"task.deleted",
	"task.mentioned",
	"project.created",
}

// Copy of notifications.ProjectIDUnresolved, so this migration does not move with it.
const notificationProjectUnresolved20260728130118 = -1

const notificationBackfillBatch20260728130118 = 500

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260728130118",
		Description: "Add notifications.project_id and backfill it from the stored payloads",
		Migrate: func(tx *xorm.Engine) error {
			if err := partialSync(tx, notificationsProjectID20260728130118{}); err != nil {
				return err
			}
			return backfillNotificationProjectIDs20260728130118(tx)
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

func backfillNotificationProjectIDs20260728130118(tx *xorm.Engine) error {
	lastID := int64(0)
	for {
		rows := []*notificationToBackfill20260728130118{}
		err := tx.
			Where(builder.Gt{"id": lastID}).
			In("name", projectScopedNotifications20260728130118).
			OrderBy("id").
			Limit(notificationBackfillBatch20260728130118).
			Find(&rows)
		if err != nil {
			return fmt.Errorf("could not read notifications to backfill: %w", err)
		}
		if len(rows) == 0 {
			return nil
		}

		if err := backfillNotificationBatch20260728130118(tx, rows); err != nil {
			return err
		}

		lastID = rows[len(rows)-1].ID
		if len(rows) < notificationBackfillBatch20260728130118 {
			return nil
		}
	}
}

func backfillNotificationBatch20260728130118(tx *xorm.Engine, rows []*notificationToBackfill20260728130118) error {
	projectByNotification := make(map[int64]int64, len(rows))
	notificationsByTask := map[int64][]int64{}

	for _, row := range rows {
		projectID, taskID := notificationScope20260728130118(row.Notification)
		switch {
		case projectID > 0:
			projectByNotification[row.ID] = projectID
		case taskID > 0:
			notificationsByTask[taskID] = append(notificationsByTask[taskID], row.ID)
		default:
			projectByNotification[row.ID] = notificationProjectUnresolved20260728130118
		}
	}

	if err := resolveTaskProjects20260728130118(tx, notificationsByTask, projectByNotification); err != nil {
		return err
	}

	notificationsByProject := map[int64][]int64{}
	for notificationID, projectID := range projectByNotification {
		notificationsByProject[projectID] = append(notificationsByProject[projectID], notificationID)
	}
	for projectID, notificationIDs := range notificationsByProject {
		_, err := tx.
			In("id", notificationIDs).
			Cols("project_id").
			Update(&notificationsProjectID20260728130118{ProjectID: projectID})
		if err != nil {
			return fmt.Errorf("could not backfill the project of %d notifications: %w", len(notificationIDs), err)
		}
	}

	return nil
}

func resolveTaskProjects20260728130118(tx *xorm.Engine, notificationsByTask map[int64][]int64, projectByNotification map[int64]int64) error {
	if len(notificationsByTask) == 0 {
		return nil
	}

	taskIDs := make([]int64, 0, len(notificationsByTask))
	for taskID := range notificationsByTask {
		taskIDs = append(taskIDs, taskID)
	}

	tasks := []*taskProject20260728130118{}
	if err := tx.In("id", taskIDs).Find(&tasks); err != nil {
		return fmt.Errorf("could not look up the projects of %d tasks: %w", len(taskIDs), err)
	}

	projectByTask := make(map[int64]int64, len(tasks))
	for _, task := range tasks {
		projectByTask[task.ID] = task.ProjectID
	}

	for taskID, notificationIDs := range notificationsByTask {
		projectID, has := projectByTask[taskID]
		if !has || projectID <= 0 {
			projectID = notificationProjectUnresolved20260728130118
		}
		for _, notificationID := range notificationIDs {
			projectByNotification[notificationID] = projectID
		}
	}

	return nil
}

// Parsed generically because a migration must not depend on the notification
// structs in pkg/models: they keep changing, this file must not.
func notificationScope20260728130118(payload string) (projectID, taskID int64) {
	decoder := json.NewDecoder(strings.NewReader(payload))
	// Without this a large id would round-trip through float64.
	decoder.UseNumber()

	var content map[string]interface{}
	if err := decoder.Decode(&content); err != nil {
		return 0, 0
	}

	if project, is := content["project"].(map[string]interface{}); is {
		if id := notificationPayloadInt20260728130118(project["id"]); id > 0 {
			return id, 0
		}
	}

	task, is := content["task"].(map[string]interface{})
	if !is {
		return 0, 0
	}
	if id := notificationPayloadInt20260728130118(task["project_id"]); id > 0 {
		return id, 0
	}
	return 0, notificationPayloadInt20260728130118(task["id"])
}

func notificationPayloadInt20260728130118(value interface{}) int64 {
	number, is := value.(json.Number)
	if !is {
		return 0
	}
	parsed, err := number.Int64()
	if err != nil {
		return 0
	}
	return parsed
}
