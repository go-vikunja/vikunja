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

type taskUser20261021192508 struct {
	TaskID int64
	UserID int64
}

const (
	subscriptionEntityProject20261021192508 = 2
	subscriptionEntityTask20261021192508    = 3
)

// Reminders now reach only subscribers, but creators were only subscribed
// implicitly since v2.6.0 and assignees since 2021. Any existing row, muted or
// not and on the task or an ancestor project, is the user's own choice and is kept.
func subscribeCreatorsAndAssigneesToTasks20261021192508(tx *xorm.Engine) error {
	pairs := []*taskUser20261021192508{}
	err := tx.SQL(`
WITH RECURSIVE project_ancestors AS (
    SELECT id AS project_id, id AS ancestor_id FROM projects

    UNION ALL

    SELECT pa.project_id, p.parent_project_id
    FROM project_ancestors pa
             JOIN projects p ON p.id = pa.ancestor_id
    WHERE p.parent_project_id > 0
),
task_users AS (
    SELECT id AS task_id, created_by_id AS user_id FROM tasks
    UNION
    SELECT task_id, user_id FROM task_assignees
)
SELECT tu.task_id, tu.user_id
FROM task_users tu
         JOIN tasks t ON t.id = tu.task_id
         JOIN users u ON u.id = tu.user_id
WHERE NOT EXISTS (
    SELECT 1 FROM subscriptions s
    WHERE s.entity_type = ? AND s.entity_id = tu.task_id AND s.user_id = tu.user_id
)
  AND NOT EXISTS (
    SELECT 1 FROM project_ancestors pa
             JOIN subscriptions s ON s.entity_type = ? AND s.entity_id = pa.ancestor_id
    WHERE pa.project_id = t.project_id AND s.user_id = tu.user_id
)`, subscriptionEntityTask20261021192508, subscriptionEntityProject20261021192508).
		Find(&pairs)
	if err != nil {
		return fmt.Errorf("could not find task creators and assignees without a subscription: %w", err)
	}

	sess := tx.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return fmt.Errorf("could not start transaction to subscribe task creators and assignees: %w", err)
	}

	// 4 params per row stays under SQLite's 999 variable limit before 3.32.
	const batchSize = 200
	for start := 0; start < len(pairs); start += batchSize {
		batch := pairs[start:min(start+batchSize, len(pairs))]
		subs := make([]*subscriptions20261021192508, 0, len(batch))
		for _, p := range batch {
			subs = append(subs, &subscriptions20261021192508{
				EntityType: subscriptionEntityTask20261021192508,
				EntityID:   p.TaskID,
				UserID:     p.UserID,
			})
		}
		if _, err := sess.Insert(&subs); err != nil {
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
