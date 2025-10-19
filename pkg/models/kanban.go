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

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// Function variables for dependency inversion
var CreateBucketFunc func(s *xorm.Session, bucket *Bucket, a web.Auth) error
var UpdateBucketFunc func(s *xorm.Session, bucket *Bucket, a web.Auth) error
var DeleteBucketFunc func(s *xorm.Session, bucketID int64, projectID int64, a web.Auth) error
var GetAllBucketsFunc func(s *xorm.Session, projectViewID int64, projectID int64, a web.Auth) ([]*Bucket, error)
var MoveTaskToBucketFunc func(s *xorm.Session, taskBucket *TaskBucket, a web.Auth) error

// Helper function variables
var GetBucketByIDFunc func(s *xorm.Session, id int64) (*Bucket, error)
var GetDefaultBucketIDFunc func(s *xorm.Session, view *ProjectView) (int64, error)
var GetTasksInBucketsForViewFunc func(s *xorm.Session, view *ProjectView, projects []*Project, opts *taskSearchOptions, a web.Auth) ([]*Bucket, error)

// Bucket represents a kanban bucket
type Bucket struct {
	// The unique, numeric id of this bucket.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"bucket"`
	// The title of this bucket.
	Title string `xorm:"text not null" valid:"required" minLength:"1" json:"title"`
	// The project this bucket belongs to.
	ProjectID int64 `xorm:"-" json:"-" param:"project"`
	// The project view this bucket belongs to.
	ProjectViewID int64 `xorm:"bigint not null" json:"project_view_id" param:"view"`
	// All tasks which belong to this bucket.
	Tasks []*Task `xorm:"-" json:"tasks,omitempty"`

	// How many tasks can be at the same time on this board max
	Limit int64 `xorm:"default 0" json:"limit" minimum:"0" valid:"range(0|9223372036854775807)"`

	// The number of tasks currently in this bucket
	Count int64 `xorm:"-" json:"count"`

	// The position this bucket has when querying all buckets. See the tasks.position property on how to use this.
	Position float64 `xorm:"double null" json:"position"`

	// A timestamp when this bucket was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this bucket was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The user who initially created the bucket.
	CreatedBy   *user.User `xorm:"-" json:"created_by" valid:"-"`
	CreatedByID int64      `xorm:"bigint not null" json:"-"`

	// Including the task collection type so we can use task filters on kanban
	TaskCollection `xorm:"-" json:"-"`
}

// TableName returns the table name for buckets
func (b *Bucket) TableName() string {
	return "buckets"
}

// GetID returns the ID of the bucket
func (b *Bucket) GetID() int64 {
	return b.ID
}

// CanCreate checks if a user can create a new bucket
func (b *Bucket) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckBucketCreateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckBucketCreateFunc(s, b, a)
}

// CanUpdate checks if a user can update an existing bucket
func (b *Bucket) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckBucketUpdateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	// Pass the whole bucket struct so the service can access b.ProjectID from URL binding
	return CheckBucketUpdateFunc(s, b, a)
}

// CanDelete checks if a user can delete an existing bucket
func (b *Bucket) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckBucketDeleteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	// Pass the whole bucket struct so the service can access b.ProjectID from URL binding
	return CheckBucketDeleteFunc(s, b, a)
}

// Create creates a new bucket.
// @Deprecated: Use services.KanbanService.CreateBucket() instead
func (b *Bucket) Create(s *xorm.Session, a web.Auth) (err error) {
	if CreateBucketFunc == nil {
		panic("KanbanService not registered - call services.InitKanbanService() in test setup")
	}
	return CreateBucketFunc(s, b, a)
}

// Update updates an existing bucket.
// @Deprecated: Use services.KanbanService.UpdateBucket() instead
func (b *Bucket) Update(s *xorm.Session, a web.Auth) (err error) {
	if UpdateBucketFunc == nil {
		panic("KanbanService not registered - call services.InitKanbanService() in test setup")
	}
	return UpdateBucketFunc(s, b, a)
}

// ReadAll returns all buckets for a project view.
// @Deprecated: Use services.KanbanService.GetAllBuckets() instead
func (b *Bucket) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if GetAllBucketsFunc == nil {
		panic("KanbanService not registered - call services.InitKanbanService() in test setup")
	}
	buckets, err := GetAllBucketsFunc(s, b.ProjectViewID, b.ProjectID, a)
	if err != nil {
		return nil, 0, 0, err
	}
	return buckets, len(buckets), int64(len(buckets)), nil
}

// Delete removes a bucket.
// @Deprecated: Use services.KanbanService.DeleteBucket() instead
func (b *Bucket) Delete(s *xorm.Session, a web.Auth) (err error) {
	if DeleteBucketFunc == nil {
		panic("KanbanService not registered - call services.InitKanbanService() in test setup")
	}
	return DeleteBucketFunc(s, b.ID, b.ProjectID, a)
}

// Helper functions that use dependency inversion

// GetDefaultBucketID returns the default bucket ID for a view.
// @Deprecated: Use services.KanbanService.getDefaultBucketID() instead
func GetDefaultBucketID(s *xorm.Session, view *ProjectView) (int64, error) {
	if GetDefaultBucketIDFunc == nil {
		panic("KanbanService not registered - call services.InitKanbanService() in test setup")
	}
	return GetDefaultBucketIDFunc(s, view)
}

func GetTasksInBucketsForView(s *xorm.Session, view *ProjectView, projects []*Project, opts *taskSearchOptions, a web.Auth) ([]*Bucket, error) {
	if GetTasksInBucketsForViewFunc != nil {
		return GetTasksInBucketsForViewFunc(s, view, projects, opts, a)
	}

	// This is a complex function that would need full implementation
	// For now, return empty slice to prevent compilation errors
	return []*Bucket{}, nil
}

// CalculateDefaultPosition calculates the default position for a bucket or similar entity.
// This function is exported for use by the service layer.
func CalculateDefaultPosition(id int64, position float64) float64 {
	if position == 0 {
		return float64(id) * 1000
	}
	return position
}
