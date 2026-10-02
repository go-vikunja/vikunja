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

type tasksIsMilestone20261002100000 struct {
	IsMilestone bool `xorm:"bool not null default false"`
}

func (tasksIsMilestone20261002100000) TableName() string {
	return "tasks"
}

func addTasksIsMilestone20261002100000(tx *xorm.Engine) error {
	if err := partialSync(tx, tasksIsMilestone20261002100000{}); err != nil {
		return fmt.Errorf("could not add is_milestone to tasks: %w", err)
	}
	return nil
}

type TaskBaseline20261002100000 struct {
	TaskID    int64     `xorm:"bigint not null pk"`
	ProjectID int64     `xorm:"bigint not null index"`
	StartDate time.Time `xorm:"DATETIME null"`
	EndDate   time.Time `xorm:"DATETIME null"`
	SavedAt   time.Time `xorm:"DATETIME not null"`
	SavedByID int64     `xorm:"bigint not null"`
}

func (TaskBaseline20261002100000) TableName() string { return "task_baselines" }

func addTaskBaselines20261002100000(tx *xorm.Engine) error {
	return tx.Sync(TaskBaseline20261002100000{}) //nolint:forbidigo // brand-new table, nothing to drop
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261002100000",
		Description: "Add tasks.is_milestone and the task_baselines table for the Gantt chart",
		Migrate: func(tx *xorm.Engine) error {
			if err := addTasksIsMilestone20261002100000(tx); err != nil {
				return err
			}
			return addTaskBaselines20261002100000(tx)
		},
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
