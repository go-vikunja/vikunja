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
	"code.vikunja.io/api/pkg/user"
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

// CanDeletePermanently checks whether the auth may purge the trashed task.
// Link shares have no trash, so they can't.
func (t *Task) CanDeletePermanently(s *xorm.Session, a web.Auth) (bool, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return false, nil
	}
	return t.canDoTrashedTask(s, a)
}

// DeletePermanently purges a trashed task right away instead of waiting for the retention period.
func (t *Task) DeletePermanently(s *xorm.Session) error {
	return hardDeleteTask(s, t)
}

// Trashed tasks in non-archived projects the user can write to.
func trashedTasksCond(s *xorm.Session, a web.Auth) (builder.Cond, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, ErrGenericForbidden{}
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}
	access, err := getProjectAccessForUser(s, u.ID)
	if err != nil {
		return nil, err
	}

	writable := []int64{}
	for _, id := range access.sortedIDs {
		if access.permissions[id] >= PermissionWrite {
			writable = append(writable, id)
		}
	}

	return builder.And(
		builder.NotNull{"tasks.deleted_at"},
		builder.In("tasks.project_id", writable),
		builder.In("tasks.project_id", builder.Select("id").From("projects").Where(builder.Eq{"is_archived": false})),
	), nil
}

// GetTrashedTasks returns the trash of the auth, newest first.
func GetTrashedTasks(s *xorm.Session, a web.Auth, page, perPage int) (tasks []*Task, total int64, err error) {
	cond, err := trashedTasksCond(s, a)
	if err != nil {
		return nil, 0, err
	}

	total, err = s.Unscoped().Where(cond).Count(&Task{})
	if err != nil {
		return nil, 0, err
	}

	tasks = []*Task{}
	err = s.Unscoped().
		Where(cond).
		OrderBy("tasks.deleted_at DESC, tasks.id DESC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&tasks)
	if err != nil {
		return nil, 0, err
	}

	taskMap := make(map[int64]*Task, len(tasks))
	deleterIDs := []int64{}
	for _, task := range tasks {
		taskMap[task.ID] = task
		if task.DeletedByID != 0 {
			deleterIDs = append(deleterIDs, task.DeletedByID)
		}
	}
	if err = addMoreInfoToTasks(s, taskMap, a, nil, nil); err != nil {
		return nil, 0, err
	}

	deleters, err := getUsersOrLinkSharesFromIDs(s, deleterIDs)
	if err != nil {
		return nil, 0, err
	}
	for _, task := range tasks {
		task.DeletedBy = deleters[task.DeletedByID]
	}

	return tasks, total, nil
}

// EmptyTaskTrash purges every task in the auth's trash, including ones others deleted.
func EmptyTaskTrash(s *xorm.Session, a web.Auth) (purged int, err error) {
	cond, err := trashedTasksCond(s, a)
	if err != nil {
		return 0, err
	}

	tasks := []*Task{}
	if err = s.Unscoped().Where(cond).Find(&tasks); err != nil {
		return 0, err
	}
	for _, task := range tasks {
		if err = hardDeleteTask(s, task); err != nil {
			return 0, err
		}
	}
	return len(tasks), nil
}
