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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type notificationID20260728131931 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`
}

func (notificationID20260728131931) TableName() string {
	return "notifications"
}

// Hardcoded rather than read from the notification registry: a migration is
// frozen in time, the registry grows.
var knownNotifications20260728131931 = []string{
	"api_token.expiring.day",
	"api_token.expiring.week",
	"project.created",
	"task.assigned",
	"task.comment",
	"task.deleted",
	"task.mentioned",
	"task.reminder",
	"team.member.added",
}

const unknownNotificationBatch20260728131931 = 500

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260728131931",
		Description: "Delete stored notifications whose type no longer exists",
		Migrate:     deleteUnknownNotifications20260728131931,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

// A row whose name no type answers to can be neither rendered nor
// permission-checked, so keeping it only preserves the leak.
func deleteUnknownNotifications20260728131931(tx *xorm.Engine) error {
	lastID := int64(0)
	for {
		rows := []*notificationID20260728131931{}
		err := tx.
			Where(builder.Gt{"id": lastID}).
			NotIn("name", knownNotifications20260728131931).
			OrderBy("id").
			Limit(unknownNotificationBatch20260728131931).
			Find(&rows)
		if err != nil {
			return fmt.Errorf("could not read notifications with an unknown name: %w", err)
		}
		if len(rows) == 0 {
			return nil
		}

		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}

		if _, err := tx.In("id", ids).Delete(&notificationID20260728131931{}); err != nil {
			return fmt.Errorf("could not delete %d notifications with an unknown name: %w", len(ids), err)
		}

		lastID = ids[len(ids)-1]
		if len(rows) < unknownNotificationBatch20260728131931 {
			return nil
		}
	}
}
