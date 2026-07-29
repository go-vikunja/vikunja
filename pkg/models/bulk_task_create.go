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

// BulkTaskCreate represents a bulk task creation payload.
type BulkTaskCreate struct {
	Tasks []*Task `json:"tasks" minItems:"1" maxItems:"200" doc:"The tasks to create, each with the project to create it in. Slice order is the order the tasks end up in: they are placed above everything the views already contain, keeping the order they were passed in."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// CanCreate checks if the user can create tasks in all involved projects.
func (bt *BulkTaskCreate) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if len(bt.Tasks) == 0 {
		return false, ErrBulkTasksNeedAtLeastOne{}
	}

	checked := map[int64]struct{}{}
	for _, t := range bt.Tasks {
		if _, has := checked[t.ProjectID]; has {
			continue
		}
		checked[t.ProjectID] = struct{}{}

		p := &Project{ID: t.ProjectID}
		can, err := p.CanWrite(s, a)
		if err != nil || !can {
			return false, err
		}
	}

	return true, nil
}

// Create creates multiple tasks at once.
func (bt *BulkTaskCreate) Create(s *xorm.Session, a web.Auth) (err error) {
	if len(bt.Tasks) == 0 {
		return ErrBulkTasksNeedAtLeastOne{}
	}

	tasksByProject := map[int64][]*Task{}
	for _, t := range bt.Tasks {
		tasksByProject[t.ProjectID] = append(tasksByProject[t.ProjectID], t)
	}

	// Shared with the create so the views are fetched once for both
	state := &taskCreateState{}

	return createTasksAtTopOfViews(s, a, tasksByProject, state, func() error {
		return createTasks(s, bt.Tasks, a, createTaskOpts{
			updateAssignees: true,
			setBucket:       true,
			// Positions are set for the whole batch at once - one at a time would place
			// every task at the same spot and leave the order to conflict repair.
			skipPositions: true,
			state:         state,
		})
	})
}
