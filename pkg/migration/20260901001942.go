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
	"errors"
	"slices"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type ProjectTaskCounter20260901001942 struct {
	ProjectID int64 `xorm:"bigint not null pk"`
	LastIndex int64 `xorm:"bigint not null default 0"`
}

func (ProjectTaskCounter20260901001942) TableName() string { return "project_task_counters" }

type projectID20260901001942 struct {
	ID int64 `xorm:"bigint"`
}

func (projectID20260901001942) TableName() string { return "projects" }

type taskProjectIndex20260901001942 struct {
	ProjectID int64 `xorm:"bigint"`
	Index     int64 `xorm:"bigint 'index'"`
}

func (taskProjectIndex20260901001942) TableName() string { return "tasks" }

// SQLite allows 32766 bind parameters per statement, Postgres 65535, so one
// insert for every project blows up on instances with more than ~16k projects.
const counterBackfillBatch20260901001942 = 500

func addProjectTaskCounters20260901001942(tx *xorm.Engine) error {
	if err := tx.Sync(ProjectTaskCounter20260901001942{}); err != nil { //nolint:forbidigo // brand-new table, nothing to drop
		return err
	}

	s := tx.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}
	fail := func(err error) error {
		_ = s.Rollback()
		return err
	}

	projects := []*projectID20260901001942{}
	if err := s.Find(&projects); err != nil {
		return fail(err)
	}

	rows, err := s.Asc("project_id").Desc("index").Rows(&taskProjectIndex20260901001942{})
	if err != nil {
		return fail(err)
	}

	lastIndexes := make(map[int64]int64, len(projects))
	task := &taskProjectIndex20260901001942{}
	for rows.Next() {
		if err := rows.Scan(task); err != nil {
			return fail(errors.Join(err, rows.Close()))
		}
		// Rows are ordered by descending index, so the first row of a project is its high water mark.
		if _, exists := lastIndexes[task.ProjectID]; !exists {
			lastIndexes[task.ProjectID] = task.Index
		}
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return fail(err)
	}

	counters := make([]*ProjectTaskCounter20260901001942, 0, len(projects))
	for _, project := range projects {
		counters = append(counters, &ProjectTaskCounter20260901001942{
			ProjectID: project.ID,
			LastIndex: lastIndexes[project.ID],
		})
	}

	for chunk := range slices.Chunk(counters, counterBackfillBatch20260901001942) {
		if _, err := s.Insert(chunk); err != nil {
			return fail(err)
		}
	}

	return s.Commit()
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260901001942",
		Description: "Add project task counters and backfill them with the highest used task index per project",
		Migrate:     addProjectTaskCounters20260901001942,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
