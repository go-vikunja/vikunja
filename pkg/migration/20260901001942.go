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

type ProjectTaskCounter20260901001942 struct {
	ProjectID int64 `xorm:"bigint not null pk"`
	LastIndex int64 `xorm:"bigint not null default 0"`
}

func (ProjectTaskCounter20260901001942) TableName() string { return "project_task_counters" }

type TaskIndexAlias20260901001942 struct {
	ProjectID int64 `xorm:"bigint not null unique(task_index_alias)"`
	Index     int64 `xorm:"bigint not null unique(task_index_alias)"`
	TaskID    int64 `xorm:"bigint not null index"`
}

func (TaskIndexAlias20260901001942) TableName() string { return "task_index_aliases" }

type projectID20260901001942 struct {
	ID int64 `xorm:"bigint"`
}

func (projectID20260901001942) TableName() string { return "projects" }

type taskProjectIndex20260901001942 struct {
	ProjectID int64 `xorm:"bigint"`
	Index     int64 `xorm:"bigint 'index'"`
}

func (taskProjectIndex20260901001942) TableName() string { return "tasks" }

type taskIndexRows20260901001942 interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close() error
}

func collectTaskIndexHighWaterMarks20260901001942(rows taskIndexRows20260901001942, projectCount int) (lastIndexes map[int64]int64, err error) {
	defer func() {
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}()

	lastIndexes = make(map[int64]int64, projectCount)
	task := &taskProjectIndex20260901001942{}
	for rows.Next() {
		if err = rows.Scan(task); err != nil {
			return nil, err
		}
		if _, exists := lastIndexes[task.ProjectID]; !exists {
			lastIndexes[task.ProjectID] = task.Index
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return lastIndexes, nil
}

func addTaskIndexState20260901001942(tx *xorm.Engine) error {
	if err := tx.Sync( //nolint:forbidigo // both tables are new
		ProjectTaskCounter20260901001942{},
		TaskIndexAlias20260901001942{},
	); err != nil {
		return err
	}

	s := tx.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}

	projects := []*projectID20260901001942{}
	if err := s.Find(&projects); err != nil {
		_ = s.Rollback()
		return err
	}

	rows, err := s.Asc("project_id").Desc("index").Rows(&taskProjectIndex20260901001942{})
	if err != nil {
		_ = s.Rollback()
		return err
	}
	lastIndexes, err := collectTaskIndexHighWaterMarks20260901001942(rows, len(projects))
	if err != nil {
		_ = s.Rollback()
		return err
	}

	counters := make([]*ProjectTaskCounter20260901001942, 0, len(projects))
	for _, project := range projects {
		counters = append(counters, &ProjectTaskCounter20260901001942{
			ProjectID: project.ID,
			LastIndex: lastIndexes[project.ID],
		})
	}

	if len(counters) > 0 {
		if _, err := s.Insert(counters); err != nil {
			_ = s.Rollback()
			return err
		}
	}
	return s.Commit()
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260901001942",
		Description: "Add monotonic task index counters and historical aliases",
		Migrate:     addTaskIndexState20260901001942,
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(TaskIndexAlias20260901001942{}, ProjectTaskCounter20260901001942{})
		},
	})
}
