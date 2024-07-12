// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

type TaskBucket struct {
	BucketID      int64 `xorm:"bigint not null index" json:"bucket_id" param:"bucket"`
	TaskID        int64 `xorm:"bigint not null index" json:"task_id"`
	ProjectViewID int64 `xorm:"bigint not null index" json:"project_view_id" param:"view"`
	ProjectID     int64 `xorm:"-" json:"-" param:"project"`
	TaskDone      bool  `xorm:"-" json:"task_done"`

	web.Rights   `xorm:"-" json:"-"`
	web.CRUDable `xorm:"-" json:"-"`
}

func (b *TaskBucket) TableName() string {
	return "task_buckets"
}

func (b *TaskBucket) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	bucket := Bucket{
		ID:            b.BucketID,
		ProjectID:     b.ProjectID,
		ProjectViewID: b.ProjectViewID,
	}
	return bucket.canDoBucket(s, a)
}

// Update is the handler to update a task bucket
// @Summary Update a task bucket
// @Description Updates a task in a bucket
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view path int true "Project View ID"
// @Param bucket path int true "Bucket ID"
// @Param taskBucket body models.TaskBucket true "The id of the task you want to move into the bucket."
// @Success 200 {object} models.TaskBucket "The updated task bucket."
// @Failure 400 {object} web.HTTPError "Invalid task bucket object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{view}/buckets/{bucket}/tasks [post]
func (b *TaskBucket) Update(s *xorm.Session, a web.Auth) (err error) {

	oldTaskBucket := &TaskBucket{}
	_, err = s.
		Where("task_id = ? AND project_view_id = ?", b.TaskID, b.ProjectViewID).
		Get(oldTaskBucket)
	if err != nil {
		return
	}

	if oldTaskBucket.BucketID == b.BucketID {
		// no need to do anything
		return
	}

	view, err := GetProjectViewByIDAndProject(s, b.ProjectViewID, b.ProjectID)
	if err != nil {
		return err
	}

	bucket, err := getBucketByID(s, b.BucketID)
	if err != nil {
		return err
	}

	// If there is a bucket set, make sure they belong to the same project as the task
	if view.ID != bucket.ProjectViewID {
		return ErrBucketDoesNotBelongToProjectView{
			ProjectViewID: view.ID,
			BucketID:      bucket.ID,
		}
	}

	task, err := GetTaskByIDSimple(s, b.TaskID)
	if err != nil {
		return err
	}

	// Check the bucket limit
	// Only check the bucket limit if the task is being moved between buckets, allow reordering the task within a bucket
	if b.BucketID != 0 && b.BucketID != oldTaskBucket.BucketID {
		err = checkBucketLimit(s, &task, bucket)
		if err != nil {
			return err
		}
	}

	var updateBucket = true

	// mark task done if moved into the done bucket
	var doneChanged bool
	if view.DoneBucketID == b.BucketID {
		doneChanged = true
		task.Done = true
		if task.isRepeating() {
			oldTask := task
			task.Done = false
			updateDone(&oldTask, &task)
			updateBucket = false
			b.BucketID = oldTaskBucket.BucketID
		}
	}

	if oldTaskBucket.BucketID == view.DoneBucketID {
		doneChanged = true
		task.Done = false
	}

	if doneChanged {
		_, err = s.Where("id = ?", task.ID).
			Cols(
				"done",
				"due_date",
				"start_date",
				"end_date",
			).
			Update(task)
		if err != nil {
			return
		}

		err = task.updateReminders(s, &task)
		if err != nil {
			return err
		}
	}

	if updateBucket {
		count, err := s.Where("task_id = ? AND project_view_id = ?", b.TaskID, b.ProjectViewID).
			Cols("bucket_id").
			Update(b)
		if err != nil {
			return err
		}

		if count == 0 {
			_, err = s.Insert(b)
			if err != nil {
				return err
			}
		}
	}

	b.TaskDone = task.Done

	doer, _ := user.GetFromAuth(a)
	return events.Dispatch(&TaskUpdatedEvent{
		Task: &task,
		Doer: doer,
	})
}
