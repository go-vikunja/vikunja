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

type taskCollection20241118123644 struct {
	Filter string `query:"filter" json:"filter"`
}

type projectViews20241118123644New struct {
	ID     int64                         `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	Filter *taskCollection20241118123644 `xorm:"json null default null" query:"filter" json:"filter"`
}

func (*projectViews20241118123644New) TableName() string {
	return "project_views"
}

type projectViews20241118123644 struct {
	ID     int64  `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	Filter string `xorm:"json null default null" query:"filter" json:"filter"`
}

func (*projectViews20241118123644) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20241118123644",
		Description: "change filter format",
		Migrate: func(tx *xorm.Engine) (err error) {
			oldViews := []*projectViews20241118123644{}

			err = tx.Where("filter != '' AND filter IS NOT NULL").Find(&oldViews)
			if err != nil {
				return
			}

			err = tx.Sync(projectViews20241118123644New{})
			if err != nil {
				return
			}

			for _, view := range oldViews {
				newView := &projectViews20241118123644New{
					ID: view.ID,
					Filter: &taskCollection20241118123644{
						Filter: view.Filter,
					},
				}

				_, err = tx.Where("id = ?", view.ID).
					Cols("filter").
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
