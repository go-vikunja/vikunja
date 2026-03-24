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

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// HardDelete permanently removes a task and all its related data.
func (t *Task) HardDelete(s *xorm.Session, a web.Auth) (err error) {
	fullTask := &Task{ID: t.ID}
	err = getTaskByIDSimpleIncludingTrashed(s, fullTask)
	if err != nil {
		return err
	}

	if _, err = s.Where("task_id = ?", t.ID).Delete(&TaskAssginee{}); err != nil {
		return err
	}
	err = removeFromFavorite(s, t.ID, a, FavoriteKindTask)
	if err != nil {
		return
	}
	_, err = s.Where("task_id = ?", t.ID).Delete(&LabelTask{})
	if err != nil {
		return
	}
	attachments, err := getTaskAttachmentsByTaskIDs(s, []int64{t.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		err = attachment.Delete(s, a)
		if err != nil && !IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskComment{})
	if err != nil {
		return
	}
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskUnreadStatus{})
	if err != nil {
		return err
	}
	_, err = s.Where("task_id = ? OR other_task_id = ?", t.ID, t.ID).Delete(&TaskRelation{})
	if err != nil {
		return
	}
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskReminder{})
	if err != nil {
		return
	}
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskPosition{})
	if err != nil {
		return
	}
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskBucket{})
	if err != nil {
		return
	}
	_, err = s.ID(t.ID).Delete(Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	events.DispatchOnCommit(s, &TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})

	return updateProjectLastUpdated(s, &Project{ID: fullTask.ProjectID})
}

func getTaskByIDSimpleIncludingTrashed(s *xorm.Session, t *Task) error {
	exists, err := s.ID(t.ID).Get(t)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskDoesNotExist{t.ID}
	}
	return nil
}

// CanDoTrashOperation checks if the user has permission to operate on a trashed task.
func CanDoTrashOperation(s *xorm.Session, t *Task, a web.Auth) (bool, error) {
	task := &Task{ID: t.ID}
	err := getTaskByIDSimpleIncludingTrashed(s, task)
	if err != nil {
		return false, err
	}
	if task.DeletedAt == nil {
		return false, &ErrTaskIsNotTrashed{TaskID: t.ID}
	}
	p := &Project{ID: task.ProjectID}
	return p.CanWrite(s, a)
}

// Restore restores a trashed task by clearing the deleted_at timestamp.
func (t *Task) Restore(s *xorm.Session, a web.Auth) (err error) {
	task := &Task{ID: t.ID}
	err = getTaskByIDSimpleIncludingTrashed(s, task)
	if err != nil {
		return err
	}
	if task.DeletedAt == nil {
		return &ErrTaskIsNotTrashed{TaskID: t.ID}
	}
	_, err = s.ID(t.ID).Cols("deleted_at").Update(&Task{})
	if err != nil {
		return err
	}
	doer, _ := user.GetFromAuth(a)
	events.DispatchOnCommit(s, &TaskRestoredEvent{
		Task: task,
		Doer: doer,
	})
	return updateProjectLastUpdated(s, &Project{ID: task.ProjectID})
}

// GetTrashedTasks returns all trashed tasks the user has read access to.
func GetTrashedTasks(s *xorm.Session, a web.Auth, projectID int64, page int, perPage int) (tasks []*Task, totalCount int64, err error) {
	query := s.Where("tasks.deleted_at IS NOT NULL").
		And(accessibleProjectIDsSubquery(a, "`tasks`.`project_id`"))

	if projectID > 0 {
		query = query.And("tasks.project_id = ?", projectID)
	}

	totalCount, err = query.Count(&Task{})
	if err != nil {
		return nil, 0, err
	}

	query = s.Where("tasks.deleted_at IS NOT NULL").
		And(accessibleProjectIDsSubquery(a, "`tasks`.`project_id`"))

	if projectID > 0 {
		query = query.And("tasks.project_id = ?", projectID)
	}

	if perPage == 0 {
		perPage = 50
	}
	if page == 0 {
		page = 1
	}

	tasks = []*Task{}
	err = query.
		OrderBy("tasks.deleted_at DESC").
		Limit(perPage, (page-1)*perPage).
		Find(&tasks)
	return
}

// EmptyTrash permanently deletes all trashed tasks the user has delete access to.
func EmptyTrash(s *xorm.Session, a web.Auth) (count int64, err error) {
	var tasks []*Task
	err = s.Where("tasks.deleted_at IS NOT NULL").
		And(accessibleProjectIDsSubquery(a, "`tasks`.`project_id`")).
		Find(&tasks)
	if err != nil {
		return 0, err
	}

	for _, task := range tasks {
		canDo, err := CanDoTrashOperation(s, task, a)
		if err != nil && !IsErrTaskIsNotTrashed(err) {
			return count, err
		}
		if !canDo {
			continue
		}

		err = task.HardDelete(s, a)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// RegisterTrashPurgeJob registers a daily cron job that permanently deletes
// tasks that have been in the trash for more than 30 days.
func RegisterTrashPurgeJob() {
	err := cron.Schedule("0 2 * * *", func() {
		s := db.NewSession()
		defer s.Close()

		cutoff := time.Now().Add(-30 * 24 * time.Hour)
		var tasks []*Task
		err := s.Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).Find(&tasks)
		if err != nil {
			log.Errorf("Trash purge: failed to find expired tasks: %s", err)
			return
		}
		if len(tasks) == 0 {
			return
		}

		log.Debugf("Trash purge: permanently deleting %d expired tasks", len(tasks))
		cronAuth := &LinkSharing{}
		for _, task := range tasks {
			err = task.HardDelete(s, cronAuth)
			if err != nil {
				log.Errorf("Trash purge: failed to hard-delete task %d: %s", task.ID, err)
				continue
			}
		}
		if err = s.Commit(); err != nil {
			log.Errorf("Trash purge: failed to commit: %s", err)
		}
	})
	if err != nil {
		log.Errorf("Failed to register trash purge cron job: %s", err)
	}
}

