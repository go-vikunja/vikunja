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

	"code.vikunja.io/api/pkg/log"
	"xorm.io/xorm"
)

// TaskCaldavDeletion records a task that was deleted so CalDAV sync-collection
// clients can receive a 404 for that task UID during the next delta sync.
type TaskCaldavDeletion struct {
	ID        int64     `xorm:"bigint autoincr not null pk"`
	UID       string    `xorm:"varchar(250) not null index"`
	ProjectID int64     `xorm:"bigint not null index"`
	DeletedAt time.Time `xorm:"timestampz not null index"`
}

// TableName returns the table name for TaskCaldavDeletion.
func (TaskCaldavDeletion) TableName() string {
	return "task_caldav_deletions"
}

// RecordCaldavTaskDeletion inserts a deletion record for the given task.
// Only records tasks that have a UID (i.e. were ever synced via CalDAV).
func RecordCaldavTaskDeletion(s *xorm.Session, task *Task) error {
	if task.UID == "" {
		return nil
	}
	_, err := s.Insert(&TaskCaldavDeletion{
		UID:       task.UID,
		ProjectID: task.ProjectID,
		DeletedAt: time.Now().UTC(),
	})
	return err
}

// GetCaldavDeletionsSince returns all tasks deleted in the given project after the given time.
func GetCaldavDeletionsSince(s *xorm.Session, projectID int64, since time.Time) ([]*TaskCaldavDeletion, error) {
	var deletions []*TaskCaldavDeletion
	err := s.
		Where("project_id = ? AND deleted_at > ?", projectID, since.UTC()).
		Find(&deletions)
	log.Debugf("[CALDAV sync-collection] GetCaldavDeletionsSince project=%d since=%v → %d results", projectID, since.UTC(), len(deletions))
	return deletions, err
}

// CleanupOldCaldavDeletions deletes records older than the given time.
func CleanupOldCaldavDeletions(s *xorm.Session, olderThan time.Time) error {
	_, err := s.Where("deleted_at < ?", olderThan.UTC()).Delete(&TaskCaldavDeletion{})
	return err
}
