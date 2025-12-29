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
	"encoding/json"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Flexible bucket configuration format that can handle both string and object filters
type bucketConfigurationCatchup struct {
	Title  string          `json:"title"`
	Filter json.RawMessage `json:"filter"`
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

			// Find all filter-mode views - we'll check individual buckets in code
			err = tx.Where("bucket_configuration_mode = 2").
				Find(&oldViews)
			if err != nil {
				return
			}

			// Transform each view's bucket_configuration
			for _, view := range oldViews {
				newView := &projectViewBucketsCatchupNew{
					ID: view.ID,
				}

				needsUpdate := false

				// Convert each bucket configuration from old to new format
				for _, configuration := range view.BucketConfiguration {
					newConfig := &bucketConfigurationCatchupNew{
						Title: configuration.Title,
					}

					// Check if filter is a string (old format) or object (already converted)
					if len(configuration.Filter) > 0 {
						switch configuration.Filter[0] {
						case '"':
							// It's a JSON string - extract and wrap in object
							var filterString string
							if err := json.Unmarshal(configuration.Filter, &filterString); err != nil {
								return err
							}
							newConfig.Filter = &taskCollection20241118123644{
								Filter: filterString,
							}
							needsUpdate = true
						case '{':
							// It's already an object - preserve it
							var existingFilter taskCollection20241118123644
							if err := json.Unmarshal(configuration.Filter, &existingFilter); err != nil {
								return err
							}
							newConfig.Filter = &existingFilter
						}
					}

					newView.BucketConfiguration = append(newView.BucketConfiguration, newConfig)
				}

				// Only update if we actually found string filters to convert
				if !needsUpdate {
					continue
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
