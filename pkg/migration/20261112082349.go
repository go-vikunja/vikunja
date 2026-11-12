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
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261112082349",
		Description: "Zero-pad overdue tasks reminder times to HH:MM",
		Migrate: func(tx *xorm.Engine) error {
			times := []string{}
			if err := tx.Table("users").Distinct("overdue_tasks_reminders_time").Find(&times); err != nil {
				return err
			}
			for _, old := range times {
				parsed, err := time.Parse("15:04", old)
				if err != nil {
					continue
				}
				padded := parsed.Format("15:04")
				if padded == old {
					continue
				}
				_, err = tx.Table("users").
					Where("overdue_tasks_reminders_time = ?", old).
					Update(map[string]any{"overdue_tasks_reminders_time": padded})
				if err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
