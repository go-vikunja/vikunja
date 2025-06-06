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

type savedFilters20210403145503 struct {
	IsFavorite bool `xorm:"default false" json:"is_favorite"`
}

func (savedFilters20210403145503) TableName() string {
	return "saved_filters"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210403145503",
		Description: "",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(savedFilters20210403145503{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
