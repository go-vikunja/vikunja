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
	"time"

	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TaskBaseline is the planned start and end of a task at the moment a baseline was saved, to compare
// the plan with what happened. A project has one baseline; saving a new one replaces it.
type TaskBaseline struct {
	TaskID    int64     `xorm:"bigint not null pk" json:"task_id"`
	ProjectID int64     `xorm:"bigint not null index" json:"-"`
	StartDate time.Time `xorm:"DATETIME null" json:"start_date"`
	EndDate   time.Time `xorm:"DATETIME null" json:"end_date"`
	SavedAt   time.Time `xorm:"DATETIME not null" json:"saved_at"`
	SavedByID int64     `xorm:"bigint not null" json:"-"`
}

// TableName returns the table name of TaskBaseline.
func (*TaskBaseline) TableName() string { return "task_baselines" }

const baselineInsertBatch = 200

func loadBaselineProject(s *xorm.Session, projectID int64) (*Project, error) {
	if projectID <= 0 {
		return nil, ErrInvalidData{Message: "a baseline belongs to a real project"}
	}
	return GetProjectSimpleByID(s, projectID)
}

// SaveProjectBaseline records the current start and end of every dated, not deleted task of the
// project and replaces the previous baseline. It needs write access. It returns the number of tasks
// in the new baseline. It does not commit.
func SaveProjectBaseline(s *xorm.Session, a web.Auth, projectID int64) (int, error) {
	project, err := loadBaselineProject(s, projectID)
	if err != nil {
		return 0, err
	}
	can, err := project.CanWrite(s, a)
	if err != nil {
		return 0, err
	}
	if !can {
		return 0, ErrGenericForbidden{}
	}

	doer, err := GetUserOrLinkShareUser(s, a)
	if err != nil {
		return 0, err
	}

	var tasks []*Task
	err = s.Where(builder.And(builder.Eq{"project_id": projectID}, taskNotDeletedCond("tasks"))).Find(&tasks)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	rows := make([]*TaskBaseline, 0, len(tasks))
	for _, t := range tasks {
		if t.StartDate.IsZero() && t.EndDate.IsZero() {
			continue
		}
		rows = append(rows, &TaskBaseline{
			TaskID:    t.ID,
			ProjectID: projectID,
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
			SavedAt:   now,
			SavedByID: doer.ID,
		})
	}

	_, err = s.Where("project_id = ?", projectID).Delete(&TaskBaseline{})
	if err != nil {
		return 0, err
	}
	// A task that was moved here from another project may still carry that project's baseline row.
	// The task id is the primary key, so it has to go before this project's row can be written.
	for start := 0; start < len(rows); start += baselineInsertBatch {
		stop := min(start+baselineInsertBatch, len(rows))
		ids := make([]int64, 0, stop-start)
		for _, row := range rows[start:stop] {
			ids = append(ids, row.TaskID)
		}
		_, err = s.In("task_id", ids).Delete(&TaskBaseline{})
		if err != nil {
			return 0, err
		}
	}
	for start := 0; start < len(rows); start += baselineInsertBatch {
		stop := min(start+baselineInsertBatch, len(rows))
		_, err = s.Insert(rows[start:stop])
		if err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

// ClearProjectBaseline removes the baseline of the project. It needs write access. It does not commit.
func ClearProjectBaseline(s *xorm.Session, a web.Auth, projectID int64) error {
	project, err := loadBaselineProject(s, projectID)
	if err != nil {
		return err
	}
	can, err := project.CanWrite(s, a)
	if err != nil {
		return err
	}
	if !can {
		return ErrGenericForbidden{}
	}
	_, err = s.Where("project_id = ?", projectID).Delete(&TaskBaseline{})
	return err
}

// GetProjectBaseline returns the baseline of the project, empty if none was saved. It needs read
// access. The time is when it was saved, zero without a baseline.
func GetProjectBaseline(s *xorm.Session, a web.Auth, projectID int64) ([]*TaskBaseline, time.Time, error) {
	project, err := loadBaselineProject(s, projectID)
	if err != nil {
		return nil, time.Time{}, err
	}
	can, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, time.Time{}, err
	}
	if !can {
		return nil, time.Time{}, ErrGenericForbidden{}
	}

	// Only tasks that still are in this project and not deleted: a baseline row outlives a task that
	// moved to another project or was deleted.
	rows := []*TaskBaseline{}
	err = s.
		Where("project_id = ? AND task_id IN (SELECT id FROM tasks WHERE project_id = ? AND deleted_at IS NULL)", projectID, projectID).
		OrderBy("task_id ASC").
		Find(&rows)
	if err != nil {
		return nil, time.Time{}, err
	}
	var savedAt time.Time
	if len(rows) > 0 {
		savedAt = rows[0].SavedAt
	}
	return rows, savedAt, nil
}
