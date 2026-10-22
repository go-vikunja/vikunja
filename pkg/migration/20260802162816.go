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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type projectView20260802162816 struct {
	ID                      int64 `xorm:"autoincr not null unique pk"`
	ProjectID               int64 `xorm:"not null index"`
	ViewKind                int   `xorm:"not null"`
	BucketConfigurationMode int   `xorm:"default 0"`
	DefaultBucketID         int64 `xorm:"bigint INDEX null"`
	DoneBucketID            int64 `xorm:"bigint INDEX null"`
}

func (projectView20260802162816) TableName() string {
	return "project_views"
}

type bucket20260802162816 struct {
	ID            int64     `xorm:"bigint autoincr not null unique pk"`
	Title         string    `xorm:"text not null"`
	ProjectViewID int64     `xorm:"bigint not null"`
	Position      float64   `xorm:"double null"`
	CreatedByID   int64     `xorm:"bigint not null"`
	Created       time.Time `xorm:"created not null"`
	Updated       time.Time `xorm:"updated not null"`
}

func (bucket20260802162816) TableName() string {
	return "buckets"
}

type taskBucket20260802162816 struct {
	BucketID      int64 `xorm:"bigint not null index"`
	TaskID        int64 `xorm:"bigint not null index unique(task_view)"`
	ProjectViewID int64 `xorm:"bigint not null index unique(task_view)"`
}

func (taskBucket20260802162816) TableName() string {
	return "task_buckets"
}

const taskBucketInsertBatch20260802162816 = 200

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260802162816",
		Description: "Repair kanban views without a bucket configuration mode",
		Migrate:     repairKanbanViews20260802162816,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

func repairKanbanViews20260802162816(tx *xorm.Engine) error {
	brokenViews := []*projectView20260802162816{}
	err := tx.Where(builder.Eq{"view_kind": 3, "bucket_configuration_mode": 0}).Find(&brokenViews)
	if err != nil {
		return err
	}

	for _, view := range brokenViews {
		// 1 = manual bucket configuration
		view.BucketConfigurationMode = 1

		buckets := []*bucket20260802162816{}
		err = tx.Where(builder.Eq{"project_view_id": view.ID}).OrderBy("position, id").Find(&buckets)
		if err != nil {
			return err
		}

		var targetBucket *bucket20260802162816
		if len(buckets) == 0 {
			ownerID, found, err := getOwnerIDForProject20260802162816(tx, view.ProjectID)
			if err != nil {
				return err
			}
			// No owner means the project or saved filter is gone - creating
			// buckets would dangle created_by_id at 0.
			if !found {
				continue
			}

			todo := &bucket20260802162816{Title: "To-Do", ProjectViewID: view.ID, Position: 100, CreatedByID: ownerID}
			doing := &bucket20260802162816{Title: "Doing", ProjectViewID: view.ID, Position: 200, CreatedByID: ownerID}
			done := &bucket20260802162816{Title: "Done", ProjectViewID: view.ID, Position: 300, CreatedByID: ownerID}
			for _, b := range []*bucket20260802162816{todo, doing, done} {
				if _, err := tx.Insert(b); err != nil {
					return err
				}
			}

			view.DefaultBucketID = todo.ID
			view.DoneBucketID = done.ID
			targetBucket = todo

			// Persisted before the task backfill so a crash during it still leaves
			// a re-run finding these buckets via the existing-buckets branch.
			_, err = tx.
				Where(builder.Eq{"id": view.ID}).
				Cols("default_bucket_id", "done_bucket_id").
				Update(view)
			if err != nil {
				return err
			}
		} else {
			targetBucket = buckets[0]
			for _, b := range buckets {
				if b.ID == view.DefaultBucketID {
					targetBucket = b
				}
			}
			// Point the default at where the tasks actually land. DoneBucketID
			// is left alone, designating an arbitrary bucket as done would mark
			// tasks done.
			view.DefaultBucketID = targetBucket.ID
		}

		// Saved filter views (negative project id) get their tasks placed by
		// the filter cron and task listeners instead.
		if view.ProjectID > 0 {
			taskIDs := []int64{}
			err = tx.Table("tasks").
				Where(builder.Eq{"project_id": view.ProjectID}).
				And(builder.IsNull{"deleted_at"}).
				And(builder.NotIn("id", builder.
					Select("task_id").
					From("task_buckets").
					Where(builder.Eq{"project_view_id": view.ID}))).
				Cols("id").
				Find(&taskIDs)
			if err != nil {
				return err
			}

			taskBuckets := []*taskBucket20260802162816{}
			for _, taskID := range taskIDs {
				taskBuckets = append(taskBuckets, &taskBucket20260802162816{
					BucketID:      targetBucket.ID,
					TaskID:        taskID,
					ProjectViewID: view.ID,
				})
			}
			for i := 0; i < len(taskBuckets); i += taskBucketInsertBatch20260802162816 {
				end := min(i+taskBucketInsertBatch20260802162816, len(taskBuckets))
				if _, err := tx.Insert(taskBuckets[i:end]); err != nil {
					return err
				}
			}
		}

		// Must be the last write for this view: the migration runs without an
		// enclosing transaction and the mode flip is what excludes the view
		// from a re-run, so an interrupted run stays resumable.
		_, err = tx.
			Where(builder.Eq{"id": view.ID}).
			Cols("bucket_configuration_mode", "default_bucket_id", "done_bucket_id").
			Update(view)
		if err != nil {
			return err
		}
	}

	return nil
}

func getOwnerIDForProject20260802162816(tx *xorm.Engine, projectID int64) (ownerID int64, found bool, err error) {
	if projectID > 0 {
		found, err = tx.Table("projects").Where(builder.Eq{"id": projectID}).Cols("owner_id").Get(&ownerID)
	} else {
		filterID := projectID*-1 - 1
		found, err = tx.Table("saved_filters").Where(builder.Eq{"id": filterID}).Cols("owner_id").Get(&ownerID)
	}
	return ownerID, found, err
}
