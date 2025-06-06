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

type projectView20240329170952 struct {
	ID       int64  `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	Filter   string `xorm:"text null default null" query:"filter" json:"filter"`
	ViewKind int    `xorm:"not null" json:"view_kind"`
}

func (projectView20240329170952) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240329170952",
		Description: "Update default filter for list views to hide completed tasks",
		Migrate: func(tx *xorm.Engine) error {

			// Update the filter for all list views to hide completed tasks unless the filter is already set
			_, err := tx.Where("view_kind = ? AND filter = ?", 0, "").Cols("filter").Update(&projectView20240329170952{Filter: "done = false"})
			if err != nil {
				return err
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
