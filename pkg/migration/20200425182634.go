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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type list20200425182634 struct {
	ID      int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"list"`
	OwnerID int64 `xorm:"int(11) INDEX not null" json:"-"`
}

func (l *list20200425182634) TableName() string {
	return "list"
}

type task20200425182634 struct {
	ID       int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	ListID   int64 `xorm:"int(11) INDEX not null" json:"list_id" param:"list"`
	Updated  int64 `xorm:"updated not null" json:"updated"`
	BucketID int64 `xorm:"int(11) null" json:"bucket_id"`
}

func (t *task20200425182634) TableName() string {
	return "tasks"
}

type bucket20200425182634 struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"bucket"`
	Title       string `xorm:"text not null" valid:"required" minLength:"1" json:"title"`
	ListID      int64  `xorm:"int(11) not null" json:"list_id" param:"list"`
	Created     int64  `xorm:"created not null" json:"created"`
	Updated     int64  `xorm:"updated not null" json:"updated"`
	CreatedByID int64  `xorm:"int(11) not null" json:"-"`
}

func (b *bucket20200425182634) TableName() string {
	return "buckets"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200425182634",
		Description: "Create one bucket for each list",
		Migrate: func(tx *xorm.Engine) (err error) {
			lists := []*list20200425182634{}
			err = tx.Find(&lists)
			if err != nil {
				return
			}

			tasks := []*task20200425182634{}
			err = tx.Find(&tasks)
			if err != nil {
				return
			}

			// This map contains all buckets with their list ids as key
			buckets := make(map[int64]*bucket20200425182634, len(lists))
			for _, l := range lists {
				buckets[l.ID] = &bucket20200425182634{
					ListID: l.ID,
					Title:  "New Bucket",
					// The bucket creator is just the same as the list's one
					CreatedByID: l.OwnerID,
				}
				_, err = tx.Insert(buckets[l.ID])
				if err != nil {
					return
				}

				for _, t := range tasks {
					if t.ListID != l.ID {
						continue
					}

					t.BucketID = buckets[l.ID].ID
					_, err = tx.Where("id = ?", t.ID).Update(t)
					if err != nil {
						return
					}
				}
			}

			return
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
