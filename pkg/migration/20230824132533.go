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
	"xorm.io/xorm"
)

type buckets20230824132533 struct {
	ID           int64     `xorm:"int(11) autoincr not null unique pk" json:"id" param:"bucket"`
	Title        string    `xorm:"text not null" valid:"required" minLength:"1" json:"title"`
	ProjectID    int64     `xorm:"int(11) not null" json:"project_id" param:"project"`
	IsDoneBucket bool      `xorm:"BOOL" json:"is_done_bucket"`
	Position     float64   `xorm:"double null" json:"position"`
	Created      time.Time `xorm:"created not null" json:"created"`
	Updated      time.Time `xorm:"updated not null" json:"updated"`
	CreatedByID  int64     `xorm:"int(11) not null" json:"-"`
}

func (buckets20230824132533) TableName() string {
	return "buckets"
}

type project20230824132533 struct {
	ID      int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"project"`
	OwnerID int64 `xorm:"int(11) INDEX not null" json:"-"`
}

func (project *project20230824132533) TableName() string {
	return "projects"
}

type task20230824132533 struct {
	ID        int64     `xorm:"int(11) autoincr not null unique pk" json:"id" param:"projecttask"`
	ProjectID int64     `xorm:"int(11) INDEX not null" json:"project_id" param:"project"`
	Updated   time.Time `xorm:"updated not null" json:"updated"`
	BucketID  int64     `xorm:"int(11) null" json:"bucket_id"`
}

func (task *task20230824132533) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20230824132533",
		Description: "",
		Migrate: func(tx *xorm.Engine) (err error) {
			projects := []*project20230824132533{}
			err = tx.
				Join("LEFT", "buckets", "buckets.project_id = projects.id").
				Where("buckets.id is null").
				Find(&projects)
			if err != nil {
				return
			}

			// This map contains all buckets with their project ids as key
			buckets := make(map[int64]*buckets20230824132533, len(projects))
			for _, project := range projects {

				buckets[project.ID] = &buckets20230824132533{
					ProjectID:   project.ID,
					Title:       "Backlog",
					CreatedByID: project.OwnerID,
				}

				_, err = tx.Insert(buckets[project.ID])
				if err != nil {
					return
				}

				// We can put all tasks from that project in the new bucket because we know
				// it is the only bucket in the project
				_, err = tx.Where("project_id = ?", project.ID).
					Cols("bucket_id").
					Update(&task20230824132533{BucketID: buckets[project.ID].ID})
				if err != nil {
					return
				}
			}

			return
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
