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

// TaskBucket represents the relation between a task and a kanban bucket.
// A task can only appear once per project view which is ensured by a
// unique index on the combination of task_id and project_view_id.
type TaskBucket struct {
	BucketID int64   `xorm:"bigint not null index" json:"bucket_id" param:"bucket"`
	Bucket   *Bucket `xorm:"-" json:"bucket"`
	// The task which belongs to the bucket. Together with ProjectViewID
	// this field is part of a unique index to prevent duplicates.
	TaskID int64 `xorm:"bigint not null index unique(task_view)" json:"task_id"`
	// The view this bucket belongs to. Combined with TaskID this forms a
	// unique index.
	ProjectViewID int64 `xorm:"bigint not null index unique(task_view)" json:"project_view_id" param:"view"`
	ProjectID     int64 `xorm:"-" json:"-" param:"project"`
	Task          *Task `xorm:"-" json:"task"`

	web.Permissions `xorm:"-" json:"-"`
	web.CRUDable    `xorm:"-" json:"-"`
}

func (b *TaskBucket) TableName() string {
	return "task_buckets"
}

func (b *TaskBucket) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	// DEPRECATED: Use KanbanService.CanUpdateTaskBucket instead
	// This delegation will be removed in T-PERM-014
	if CheckBucketUpdateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	// Load the bucket to get project info for permission check
	bucket := &Bucket{ID: b.BucketID}
	// Note: bucket.ProjectID will be 0 here, but the service will look it up via the view
	return CheckBucketUpdateFunc(s, bucket, a)
}

// upsert inserts or updates a task bucket relation.
// @Deprecated: Use services.KanbanService.upsertTaskBucket() instead - this is only for model backward compatibility
func (b *TaskBucket) upsert(s *xorm.Session) (err error) {
	// This method is called by the Update method below which delegates to service layer
	// The service layer will handle the actual upsert logic
	panic("TaskBucket.upsert() should not be called directly - use services.KanbanService.MoveTaskToBucket() instead")
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
// @Deprecated: Use services.KanbanService.MoveTaskToBucket() instead
func (b *TaskBucket) Update(s *xorm.Session, a web.Auth) (err error) {
	if MoveTaskToBucketFunc == nil {
		panic("KanbanService not registered - call services.InitKanbanService() in test setup")
	}
	return MoveTaskToBucketFunc(s, b, a)
}
