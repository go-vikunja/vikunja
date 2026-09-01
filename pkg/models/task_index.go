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

import (
	"math"

	"xorm.io/xorm"
)

type ProjectTaskCounter struct {
	ProjectID int64 `xorm:"bigint not null pk"`
	LastIndex int64 `xorm:"bigint not null default 0"`
}

func (*ProjectTaskCounter) TableName() string { return "project_task_counters" }

// reserveTaskIndexes atomically claims count indexes and returns the new high water mark.
func reserveTaskIndexes(s *xorm.Session, projectID, count int64) (lastIndex int64, err error) {
	affected, err := s.ID(projectID).
		Where("last_index <= ?", math.MaxInt64-count).
		Incr("last_index", count).
		Update(&ProjectTaskCounter{})
	if err != nil {
		return 0, err
	}

	counter := &ProjectTaskCounter{}
	has, err := s.ID(projectID).Get(counter)
	if err != nil {
		return 0, err
	}
	if has {
		if affected == 0 {
			return 0, ErrTaskIndexExhausted{ProjectID: projectID}
		}
		return counter.LastIndex, nil
	}

	// Projects inserted outside CreateProject (fixtures, test seeding) have no counter row.
	highestTask := &Task{}
	_, err = s.Unscoped().
		Where("project_id = ?", projectID).
		Desc("index").
		Cols("index").
		Get(highestTask)
	if err != nil {
		return 0, err
	}

	// Aliases keep indexes of moved-away tasks reserved, so they can outrank every remaining task.
	highestAlias := &TaskIndexAlias{}
	_, err = s.Where("project_id = ?", projectID).
		Desc("index").
		Cols("index").
		Get(highestAlias)
	if err != nil {
		return 0, err
	}

	highest := max(highestTask.Index, highestAlias.Index)
	if highest > math.MaxInt64-count {
		return 0, ErrTaskIndexExhausted{ProjectID: projectID}
	}

	lastIndex = highest + count
	_, err = s.Insert(&ProjectTaskCounter{ProjectID: projectID, LastIndex: lastIndex})
	if err != nil {
		return 0, err
	}

	return lastIndex, nil
}

type TaskIndexAlias struct {
	ProjectID int64 `xorm:"bigint not null unique(task_index_alias)"`
	Index     int64 `xorm:"bigint not null unique(task_index_alias)"`
	TaskID    int64 `xorm:"bigint not null index"`
}

func (*TaskIndexAlias) TableName() string { return "task_index_aliases" }

// GetTaskIDByIndexAlias resolves a retired task address.
func GetTaskIDByIndexAlias(s *xorm.Session, projectID, index int64) (int64, error) {
	alias := &TaskIndexAlias{}
	has, err := s.Where("project_id = ? AND `index` = ?", projectID, index).Get(alias)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, ErrTaskDoesNotExist{}
	}
	return alias.TaskID, nil
}
