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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type TaskIndexAlias20260901220330 struct {
	ProjectID int64 `xorm:"bigint not null unique(task_index_alias)"`
	Index     int64 `xorm:"bigint not null unique(task_index_alias)"`
	TaskID    int64 `xorm:"bigint not null index"`
}

func (TaskIndexAlias20260901220330) TableName() string { return "task_index_aliases" }

func addTaskIndexAliases20260901220330(tx *xorm.Engine) error {
	return tx.Sync(TaskIndexAlias20260901220330{}) //nolint:forbidigo // brand-new table, nothing to drop
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260901220330",
		Description: "Add task index aliases to keep old task indexes resolvable",
		Migrate:     addTaskIndexAliases20260901220330,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
