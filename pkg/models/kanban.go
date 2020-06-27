// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"time"
)

// Bucket represents a kanban bucket
type Bucket struct {
	// The unique, numeric id of this bucket.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"bucket"`
	// The title of this bucket.
	Title string `xorm:"text not null" valid:"required" minLength:"1" json:"title"`
	// The list this bucket belongs to.
	ListID int64 `xorm:"int(11) not null" json:"list_id" param:"list"`
	// All tasks which belong to this bucket.
	Tasks []*Task `xorm:"-" json:"tasks"`

	// A timestamp when this bucket was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this bucket was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The user who initially created the bucket.
	CreatedBy   *user.User `xorm:"-" json:"created_by" valid:"-"`
	CreatedByID int64      `xorm:"int(11) not null" json:"-"`

	web.Rights   `xorm:"-" json:"-"`
	web.CRUDable `xorm:"-" json:"-"`
}

// TableName returns the table name for this bucket.
func (b *Bucket) TableName() string {
	return "buckets"
}

func getBucketByID(id int64) (b *Bucket, err error) {
	b = &Bucket{}
	exists, err := x.Where("id = ?", id).Get(b)
	if err != nil {
		return
	}
	if !exists {
		return b, ErrBucketDoesNotExist{BucketID: id}
	}
	return
}

func getDefaultBucket(listID int64) (bucket *Bucket, err error) {
	bucket = &Bucket{}
	_, err = x.
		Where("list_id = ?", listID).
		OrderBy("id asc").
		Get(bucket)
	return
}

// ReadAll returns all buckets with their tasks for a certain list
// @Summary Get all kanban buckets of a list
// @Description Returns all kanban buckets with belong to a list including their tasks.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List Id"
// @Success 200 {array} models.Bucket "The buckets with their tasks"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /lists/{id}/buckets [get]
func (b *Bucket) ReadAll(auth web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	// Note: I'm ignoring pagination for now since I've yet to figure out a way on how to make it work
	// I'll probably just don't do it and instead make individual tasks archivable.

	// Get all buckets for this list
	buckets := []*Bucket{}
	err = x.Where("list_id = ?", b.ListID).Find(&buckets)
	if err != nil {
		return
	}

	// Make a map from the bucket slice with their id as key so that we can use it to put the tasks in their buckets
	bucketMap := make(map[int64]*Bucket, len(buckets))
	userIDs := make([]int64, 0, len(buckets))
	for _, bb := range buckets {
		bucketMap[bb.ID] = bb
		userIDs = append(userIDs, bb.CreatedByID)
	}

	// Get all users
	users := make(map[int64]*user.User)
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	for _, bb := range buckets {
		bb.CreatedBy = users[bb.CreatedByID]
	}

	// Get all tasks for this list
	opts := &taskOptions{
		sortby: []*sortParam{
			{
				sortBy:  taskPropertyPosition,
				orderBy: orderAscending,
			},
		},
	}
	tasks, _, _, err := getTasksForLists([]*List{{ID: b.ListID}}, opts)
	if err != nil {
		return
	}

	// Put all tasks in their buckets
	// All tasks which are not associated to any bucket will have bucket id 0 which is the nil value for int64
	// Since we created a bucked with that id at the beginning, all tasks should be in there.
	for _, task := range tasks {
		// Check if the bucket exists in the map to prevent nil pointer panics
		if _, exists := bucketMap[task.BucketID]; !exists {
			log.Debugf("Tried to put task %d into bucket %d which does not exist in list %d", task.ID, task.BucketID, b.ListID)
			continue
		}
		bucketMap[task.BucketID].Tasks = append(bucketMap[task.BucketID].Tasks, task)
	}

	return buckets, len(buckets), int64(len(buckets)), nil
}

// Create creates a new bucket
// @Summary Create a new bucket
// @Description Creates a new kanban bucket on a list.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List Id"
// @Param bucket body models.Bucket true "The bucket object"
// @Success 200 {object} models.Bucket "The created bucket object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid bucket object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/buckets [put]
func (b *Bucket) Create(a web.Auth) (err error) {
	b.CreatedByID = a.GetID()

	_, err = x.Insert(b)
	return
}

// Update Updates an existing bucket
// @Summary Update an existing bucket
// @Description Updates an existing kanban bucket.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param listID path int true "List Id"
// @Param bucketID path int true "Bucket Id"
// @Param bucket body models.Bucket true "The bucket object"
// @Success 200 {object} models.Bucket "The created bucket object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid bucket object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The bucket does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/buckets/{bucketID} [post]
func (b *Bucket) Update() (err error) {
	_, err = x.Where("id = ?", b.ID).Update(b)
	return
}

// Delete removes a bucket, but no tasks
// @Summary Deletes an existing bucket
// @Description Deletes an existing kanban bucket and dissociates all of its task. It does not delete any tasks. You cannot delete the last bucket on a list.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param listID path int true "List Id"
// @Param bucketID path int true "Bucket Id"
// @Success 200 {object} models.Message "Successfully deleted."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The bucket does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/buckets/{bucketID} [delete]
func (b *Bucket) Delete() (err error) {
	// Prevent removing the last bucket
	total, err := x.Where("list_id = ?", b.ListID).Count(&Bucket{})
	if err != nil {
		return
	}
	if total <= 1 {
		return ErrCannotRemoveLastBucket{
			BucketID: b.ID,
			ListID:   b.ListID,
		}
	}

	// Remove the bucket itself
	_, err = x.Where("id = ?", b.ID).Delete(&Bucket{})
	if err != nil {
		return
	}

	// Get the default bucket
	defaultBucket, err := getDefaultBucket(b.ListID)
	if err != nil {
		return
	}

	// Remove all associations of tasks to that bucket
	_, err = x.Where("bucket_id = ?", b.ID).Cols("bucket_id").Update(&Task{BucketID: defaultBucket.ID})
	if err != nil {
		return
	}
	return
}
