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

type unsplashPhoto20200524221534 struct {
	ID         int64  `xorm:"autoincr not null unique pk"`
	FileID     int64  `xorm:"not null"`
	UnsplashID string `xorm:"varchar(50)"`
	Author     string `xorm:"text"`
	AuthorName string `xorm:"text"`
}

func (u unsplashPhoto20200524221534) TableName() string {
	return "unsplash_photos"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200524221534",
		Description: "Create unsplash photo table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(unsplashPhoto20200524221534{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(unsplashPhoto20200524221534{})
		},
	})
}
