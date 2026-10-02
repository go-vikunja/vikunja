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

type TaskExternalRef20261001110000 struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	BotID      int64     `xorm:"bigint not null unique(task_external_ref)"`
	ExternalID string    `xorm:"varchar(250) not null unique(task_external_ref)"`
	TaskID     int64     `xorm:"bigint not null index"`
	Created    time.Time `xorm:"created not null"`
}

func (TaskExternalRef20261001110000) TableName() string { return "task_external_refs" }

func addTaskExternalRefs20261001110000(tx *xorm.Engine) error {
	return tx.Sync(TaskExternalRef20261001110000{}) //nolint:forbidigo // brand-new table, nothing to drop
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261001110000",
		Description: "Add task_external_refs to deduplicate tasks created by integrations",
		Migrate:     addTaskExternalRefs20261001110000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
