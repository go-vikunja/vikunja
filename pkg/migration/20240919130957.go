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

type projectView20240919130957BucketConfiguration struct {
	Title  string `json:"title"`
	Filter string `json:"filter"`
}

type projectView20240919130957Lowercase struct {
	ID                  int64                                           `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	BucketConfiguration []*projectView20240919130957BucketConfiguration `xorm:"json" json:"bucket_configuration"`
}

func (projectView20240919130957Lowercase) TableName() string {
	return "project_views"
}

type projectView20240919130957TitleCase struct {
	ID                  int64 `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	BucketConfiguration []*struct {
		Title  string `json:"Title"`
		Filter string `json:"Filter"`
	} `xorm:"json" json:"bucket_configuration"`
}

func (projectView20240919130957TitleCase) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240919130957",
		Description: "",
		Migrate: func(tx *xorm.Engine) (err error) {
			oldViews := []*projectView20240919130957TitleCase{}
			// 1 is manual
			err = tx.Where("bucket_configuration_mode != 1 AND view_kind = 3").Find(&oldViews)
			if err != nil {
				return
			}

			if len(oldViews) == 0 {
				return nil
			}

			for _, view := range oldViews {
				newView := &projectView20240919130957Lowercase{
					ID:                  view.ID,
					BucketConfiguration: make([]*projectView20240919130957BucketConfiguration, 0),
				}

				for _, bc := range view.BucketConfiguration {
					newView.BucketConfiguration = append(newView.BucketConfiguration, &projectView20240919130957BucketConfiguration{
						Filter: bc.Filter,
						Title:  bc.Title,
					})
				}
				_, err = tx.
					Where("id = ?", view.ID).
					Cols("id", "bucket_configuration").
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
