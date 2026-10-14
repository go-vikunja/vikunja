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

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func getTrashedTask(s *xorm.Session, taskID int64) (*Task, error) {
	task := &Task{}
	has, err := s.Unscoped().
		Where(builder.And(builder.Eq{"id": taskID}, builder.NotNull{"deleted_at"})).
		Get(task)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrTaskDoesNotExist{ID: taskID}
	}
	return task, nil
}

// canDoTrashedTask loads the trashed task into t and checks write access to its project.
func (t *Task) canDoTrashedTask(s *xorm.Session, a web.Auth) (bool, error) {
	trashed, err := getTrashedTask(s, t.ID)
	if err != nil {
		return false, err
	}
	*t = *trashed
	return (&Project{ID: t.ProjectID}).CanWrite(s, a)
}

// CanRestore checks whether the auth may restore the trashed task.
func (t *Task) CanRestore(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canDoTrashedTask(s, a)
}

// Restore brings a trashed task back. The task lands in the default (or done)
// bucket at the end of every view; its old position is gone.
func (t *Task) Restore(s *xorm.Session, a web.Auth) error {
	// CanWrite skips this for instance admins
	if err := (&Project{ID: t.ProjectID}).CheckIsArchived(s); err != nil {
		return err
	}

	_, err := s.Table("tasks").
		Where("id = ?", t.ID).
		Update(map[string]any{"deleted_at": nil, "deleted_by_id": nil})
	if err != nil {
		return err
	}
	t.DeletedAt = time.Time{}
	t.DeletedByID = 0

	// CalDAV clients saw the task as deleted and only pick it up again through this
	if err := updateTaskLastUpdated(s, t); err != nil {
		return err
	}

	if err := placeTasksInViews(s, a, t.ProjectID, []*Task{t}, true, nil); err != nil {
		return err
	}

	events.DispatchOnCommit(s, &TaskRestoredEvent{
		Task: t,
		Doer: doerFromAuth(s, a),
	})

	return updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
}
