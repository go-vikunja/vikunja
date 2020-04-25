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

package migration

import (
	"code.vikunja.io/api/pkg/models"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200425182634",
		Description: "Create one bucket for each list",
		Migrate: func(tx *xorm.Engine) (err error) {
			lists := []*models.List{}
			err = tx.Find(&lists)
			if err != nil {
				return
			}

			tasks := []*models.Task{}
			err = tx.Find(&tasks)
			if err != nil {
				return
			}

			// This map contains all buckets with their list ids as key
			buckets := make(map[int64]*models.Bucket, len(lists))
			for _, l := range lists {
				buckets[l.ID] = &models.Bucket{
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
