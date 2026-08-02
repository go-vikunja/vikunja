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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// MaxTasksPerBulkCreation is the maximum number of tasks one bulk creation request may contain.
const MaxTasksPerBulkCreation = 100

// BulkTaskCreation is a helper struct to create multiple tasks in one project at once.
type BulkTaskCreation struct {
	ProjectID int64 `json:"-"`
	// The maxItems tag cannot reference MaxTasksPerBulkCreation; TestBulkTaskCreation_TaskLimitTag pins them together.
	Tasks []*Task `json:"tasks" minItems:"1" maxItems:"100" doc:"The tasks to create. Each entry accepts the same fields as the single-task create endpoint. Creation is atomic: if one task is invalid, none are created."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// CanCreate checks if a user may create tasks in the project — one write check covers the whole batch since all tasks land in the same project.
func (btc *BulkTaskCreation) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	p := &Project{ID: btc.ProjectID}
	return p.CanWrite(s, a)
}

// Create creates all tasks of the batch in one transaction — an invalid task rolls back everything and the error names its index.
func (btc *BulkTaskCreation) Create(s *xorm.Session, a web.Auth) (err error) {
	if len(btc.Tasks) == 0 || len(btc.Tasks) > MaxTasksPerBulkCreation {
		return ErrInvalidBulkTaskCreationCount{Count: len(btc.Tasks)}
	}

	for i, t := range btc.Tasks {
		if t == nil {
			return ErrInvalidTaskInBulkCreation{Index: i, Err: ErrTaskCannotBeEmpty{}}
		}
		t.Position = 0
	}

	return createTasks(s, btc.ProjectID, btc.Tasks, a, true, true)
}
