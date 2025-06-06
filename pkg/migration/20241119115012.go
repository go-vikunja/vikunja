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

type projectView20241119115012BucketConfiguration struct {
	Title  string `json:"title"`
	Filter string `json:"filter"`
}

type projectView20241119115012 struct {
	ID                  int64                                           `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	BucketConfiguration []*projectView20241119115012BucketConfiguration `xorm:"json" json:"bucket_configuration"`
}

func (projectView20241119115012) TableName() string {
	return "project_views"
}

type projectView20241119115012BucketConfigurationNew struct {
	Title  string                        `json:"title"`
	Filter *taskCollection20241118123644 `json:"filter"`
}

type projectView20241119115012New struct {
	ID                  int64                                              `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	BucketConfiguration []*projectView20241119115012BucketConfigurationNew `xorm:"json" json:"bucket_configuration"`
}

func (projectView20241119115012New) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20241119115012",
		Description: "change bucket filter format",
		Migrate: func(tx *xorm.Engine) (err error) {
			oldViews := []*projectView20241119115012{}

			err = tx.Where("bucket_configuration_mode = 2").Find(&oldViews)
			if err != nil {
				return
			}

			err = tx.Sync(projectView20241119115012New{})
			if err != nil {
				return
			}

			for _, view := range oldViews {
				newView := &projectView20241119115012New{
					ID: view.ID,
				}

				for _, configuration := range view.BucketConfiguration {
					newView.BucketConfiguration = append(newView.BucketConfiguration, &projectView20241119115012BucketConfigurationNew{
						Title: configuration.Title,
						Filter: &taskCollection20241118123644{
							Filter: configuration.Filter,
						},
					})
				}

				_, err = tx.Where("id = ?", view.ID).Update(newView)
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
