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
	"xorm.io/xorm/schemas"
)

// Old bucket configuration format (filter as string)
type bucketConfigurationCatchup struct {
	Title  string `json:"title"`
	Filter string `json:"filter"`
}

// New bucket configuration format (filter as object)
type bucketConfigurationCatchupNew struct {
	Title  string                        `json:"title"`
	Filter *taskCollection20241118123644 `json:"filter"`
}

// Old format project view
type projectViewBucketsCatchup struct {
	ID                  int64                         `xorm:"autoincr not null unique pk"`
	BucketConfiguration []*bucketConfigurationCatchup `xorm:"json"`
}

func (projectViewBucketsCatchup) TableName() string {
	return "project_views"
}

// New format project view
type projectViewBucketsCatchupNew struct {
	ID                  int64                            `xorm:"autoincr not null unique pk"`
	BucketConfiguration []*bucketConfigurationCatchupNew `xorm:"json"`
}

func (projectViewBucketsCatchupNew) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20251001113831",
		Description: "catch up on bucket_configuration filter format conversions",
		Migrate: func(tx *xorm.Engine) (err error) {
			oldViews := []*projectViewBucketsCatchup{}

			// Find views with bucket_configuration in old string format
			// Only check views with bucket_configuration_mode = 2 (filter mode)
			// Pattern: bucket_configuration contains "filter":"<string>" but not "filter":{"filter":
			if tx.Dialect().URI().DBType == schemas.POSTGRES {
				err = tx.Where("bucket_configuration_mode = 2 AND bucket_configuration::text like '%\"filter\":\"%'").
					And("bucket_configuration::text not like '%\"filter\":{\"filter\":%'").
					Find(&oldViews)
			} else {
				err = tx.Where("bucket_configuration_mode = 2 AND bucket_configuration like '%\"filter\":\"%'").
					And("bucket_configuration not like '%\"filter\":{\"filter\":%'").
					Find(&oldViews)
			}

			if err != nil {
				return
			}

			// Transform each view's bucket_configuration
			for _, view := range oldViews {
				newView := &projectViewBucketsCatchupNew{
					ID: view.ID,
				}

				// Convert each bucket configuration from old to new format
				for _, configuration := range view.BucketConfiguration {
					newView.BucketConfiguration = append(newView.BucketConfiguration,
						&bucketConfigurationCatchupNew{
							Title: configuration.Title,
							Filter: &taskCollection20241118123644{
								Filter: configuration.Filter, // Wrap string in object
							},
						})
				}

				// Update only the bucket_configuration column
				_, err = tx.Where("id = ?", view.ID).
					Cols("bucket_configuration").
					Update(newView)
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
