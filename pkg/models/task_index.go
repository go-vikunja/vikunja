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

package models

type ProjectTaskCounter struct {
	ProjectID int64 `xorm:"bigint not null pk"`
	LastIndex int64 `xorm:"bigint not null default 0"`
}

func (*ProjectTaskCounter) TableName() string { return "project_task_counters" }

type TaskIndexAlias struct {
	ProjectID int64 `xorm:"bigint not null unique(task_index_alias)"`
	Index     int64 `xorm:"bigint not null unique(task_index_alias)"`
	TaskID    int64 `xorm:"bigint not null index"`
}

func (*TaskIndexAlias) TableName() string { return "task_index_aliases" }
