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
	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type projectViewBucketSort20260718131307Old struct {
	ID              int64  `xorm:"autoincr not null unique pk"`
	BucketSortBy    string `xorm:"varchar(50) null default null"`
	BucketSortOrder string `xorm:"varchar(4) null default null"`
}

func (projectViewBucketSort20260718131307Old) TableName() string {
	return "project_views"
}

type projectViewBucketSort20260718131307New struct {
	ID              int64    `xorm:"autoincr not null unique pk"`
	BucketSortBy    []string `xorm:"json null default null"`
	BucketSortOrder []string `xorm:"json null default null"`
}

func (projectViewBucketSort20260718131307New) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260718131307",
		Description: "convert bucket_sort_by/bucket_sort_order from single values to ordered lists",
		Migrate: func(tx *xorm.Engine) error {
			// bucket_sort_by/bucket_sort_order started out as VARCHAR(50)/VARCHAR(4),
			// which is far too small to hold a JSON-encoded list. SQLite has no
			// enforced column length, so it doesn't need widening.
			switch db.Type() {
			case schemas.MYSQL:
				if _, err := tx.Exec("ALTER TABLE project_views MODIFY COLUMN bucket_sort_by TEXT NULL"); err != nil {
					return err
				}
				if _, err := tx.Exec("ALTER TABLE project_views MODIFY COLUMN bucket_sort_order TEXT NULL"); err != nil {
					return err
				}
			case schemas.POSTGRES:
				if _, err := tx.Exec("ALTER TABLE project_views ALTER COLUMN bucket_sort_by TYPE TEXT"); err != nil {
					return err
				}
				if _, err := tx.Exec("ALTER TABLE project_views ALTER COLUMN bucket_sort_order TYPE TEXT"); err != nil {
					return err
				}
			}

			oldViews := []*projectViewBucketSort20260718131307Old{}
			err := tx.
				Where("(bucket_sort_by IS NOT NULL AND bucket_sort_by != '') OR (bucket_sort_order IS NOT NULL AND bucket_sort_order != '')").
				Find(&oldViews)
			if err != nil {
				return err
			}

			for _, view := range oldViews {
				newView := &projectViewBucketSort20260718131307New{
					ID: view.ID,
				}
				if view.BucketSortBy != "" {
					newView.BucketSortBy = []string{view.BucketSortBy}
				}
				if view.BucketSortOrder != "" {
					newView.BucketSortOrder = []string{view.BucketSortOrder}
				}

				if _, err := tx.
					Where("id = ?", view.ID).
					Cols("bucket_sort_by", "bucket_sort_order").
					Update(newView); err != nil {
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
