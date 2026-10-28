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

// Unreadable trashed tasks are ErrTaskDoesNotExist so their existence doesn't leak.
func getReadableTrashedTask(s *xorm.Session, a web.Auth, taskID int64) (*Task, error) {
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

	canRead, _, err := (&Project{ID: task.ProjectID}).CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrTaskDoesNotExist{ID: taskID}
	}
	return task, nil
}

// Readers of the project get ErrTaskIsInTrash; everyone else keeps notFound.
func trashedTaskReadError(s *xorm.Session, a web.Auth, taskID int64, notFound error) error {
	trashed, err := getReadableTrashedTask(s, a, taskID)
	if IsErrTaskDoesNotExist(err) {
		return notFound
	}
	if err != nil {
		return err
	}

	inTrash := ErrTaskIsInTrash{
		ID:        trashed.ID,
		ProjectID: trashed.ProjectID,
		DeletedAt: trashed.DeletedAt,
	}
	if trashed.DeletedByID != 0 {
		deleters, err := getUsersOrLinkSharesFromIDs(s, []int64{trashed.DeletedByID})
		if err != nil {
			return err
		}
		inTrash.DeletedBy = deleters[trashed.DeletedByID]
	}
	return inTrash
}

// GetTaskByIDSimple, but a trashed task is ErrTaskIsInTrash for readers of its project.
func getTaskByIDForAuth(s *xorm.Session, a web.Auth, taskID int64) (Task, error) {
	task, err := GetTaskByIDSimple(s, taskID)
	if IsErrTaskDoesNotExist(err) {
		return Task{}, trashedTaskReadError(s, a, taskID, err)
	}
	return task, err
}

// Restore brings a trashed task back into its default or done bucket, positioned like a new task.
func (t *Task) Restore(s *xorm.Session, a web.Auth) error {
	restored, err := s.Table("tasks").
		Where(builder.And(builder.Eq{"id": t.ID}, builder.NotNull{"deleted_at"})).
		Update(map[string]any{
			"deleted_at":    nil,
			"deleted_by_id": nil,
		})
	if err != nil {
		return err
	}
	// A concurrent restore got there first
	if restored != 1 {
		return ErrTaskDoesNotExist{ID: t.ID}
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

// DeletePermanently purges a trashed task.
func (t *Task) DeletePermanently(s *xorm.Session, a web.Auth) error {
	return purgeTrashedTask(s, t, doerFromAuth(s, a))
}

func purgeTrashedTask(s *xorm.Session, t *Task, doer *user.User) error {
	if err := hardDeleteTask(s, t); err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TaskPurgedEvent{
		Task: t,
		Doer: doer,
	})
	return nil
}

func trashedTasksCond(s *xorm.Session, a web.Auth, minPermission Permission) (builder.Cond, *projectAccess, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, nil, ErrGenericForbidden{}
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, nil, err
	}
	access, err := getProjectAccessForUser(s, u.ID)
	if err != nil {
		return nil, nil, err
	}

	return builder.And(
		builder.NotNull{"tasks.deleted_at"},
		access.cond("tasks.project_id", minPermission),
		builder.In("tasks.project_id", builder.Select("id").From("projects").Where(builder.Eq{"is_archived": false})),
	), access, nil
}

// TrashedTask is a task in the caller's trash.
type TrashedTask struct {
	Task
	CanDeletePermanently bool `json:"can_delete_permanently" readOnly:"true" doc:"Whether the requesting user may permanently delete this task, which needs admin access to its project. Write access is enough to see and restore it."`
}

// GetTrashedTasks returns the trash of the auth, newest first.
func GetTrashedTasks(s *xorm.Session, a web.Auth, page, perPage int) (trashed []*TrashedTask, total int64, err error) {
	cond, access, err := trashedTasksCond(s, a, PermissionWrite)
	if err != nil {
		return nil, 0, err
	}

	total, err = s.Unscoped().Where(cond).Count(&Task{})
	if err != nil {
		return nil, 0, err
	}

	tasks := []*Task{}
	err = s.Unscoped().
		Where(cond).
		OrderBy("tasks.deleted_at DESC, tasks.id DESC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&tasks)
	if err != nil {
		return nil, 0, err
	}

	taskMap := make(map[int64]*Task, len(tasks))
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	if err = addMoreInfoToTasks(s, taskMap, a, nil, nil); err != nil {
		return nil, 0, err
	}

	trashed = make([]*TrashedTask, 0, len(tasks))
	for _, task := range tasks {
		permission, _ := access.permission(task.ProjectID)
		trashed = append(trashed, &TrashedTask{
			Task:                 *task,
			CanDeletePermanently: permission >= PermissionAdmin,
		})
	}
	return trashed, total, nil
}

// EmptyTaskTrash purges the auth's trash in projects it administers, including tasks others deleted.
func EmptyTaskTrash(s *xorm.Session, a web.Auth) error {
	cond, _, err := trashedTasksCond(s, a, PermissionAdmin)
	if err != nil {
		return err
	}

	tasks := []*Task{}
	if err = s.Unscoped().Where(cond).Find(&tasks); err != nil {
		return err
	}
	doer := doerFromAuth(s, a)
	for _, task := range tasks {
		if err = purgeTrashedTask(s, task, doer); err != nil {
			return err
		}
	}
	return nil
}
