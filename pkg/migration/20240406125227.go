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
	"code.vikunja.io/api/pkg/log"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type taskBucket20240406125227 struct {
	BucketID      int64 `xorm:"bigint not null index"`
	TaskID        int64 `xorm:"bigint not null index"`
	ProjectViewID int64 `xorm:"bigint not null index"`
}

func (taskBucket20240406125227) TableName() string {
	return "task_buckets"
}

type bucket20240406125227 struct {
	ID            int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"bucket"`
	ProjectViewID int64 `xorm:"bigint not null" json:"project_view_id" param:"view"`
}

func (bucket20240406125227) TableName() string {
	return "buckets"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240406125227",
		Description: "Add correct project_view_id to task_buckets",
		Migrate: func(tx *xorm.Engine) error {

			buckets := make(map[int64]*bucket20240406125227)

			err := tx.Find(&buckets)
			if err != nil {
				return err
			}

			tbs := []*taskBucket20240406125227{}
			err = tx.Where("project_view_id = 0").Find(&tbs)
			if err != nil {
				return err
			}

			if len(tbs) == 0 {
				return nil
			}

			for _, tb := range tbs {
				bucket, exists := buckets[tb.BucketID]
				if !exists {
					log.Debugf("Bucket %d does not exist but has task_buckets relation", tb.BucketID)
					continue
				}
				tb.ProjectViewID = bucket.ProjectViewID
				_, err = tx.
					Where("task_id = ? AND bucket_id = ?", tb.TaskID, tb.BucketID).
					Cols("project_view_id").
					Update(tb)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
