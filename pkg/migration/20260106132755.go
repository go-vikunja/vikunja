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

type wikiPages20260106132755 struct {
	ID        int64  `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64  `xorm:"bigint not null INDEX"`
	ParentID  *int64 `xorm:"bigint null INDEX"`
	Title     string `xorm:"varchar(250) not null"`
	Content   string `xorm:"longtext null"`
	Path      string `xorm:"varchar(500) not null INDEX"`
	IsFolder  bool   `xorm:"bool default false"`
	Position  float64 `xorm:"double not null"`
	CreatedByID int64  `xorm:"bigint not null"`
	
	Created int64 `xorm:"created not null"`
	Updated int64 `xorm:"updated not null"`
}

func (wikiPages20260106132755) TableName() string {
	return "wiki_pages"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260106132755",
		Description: "Add wiki pages support to projects",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(wikiPages20260106132755{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(wikiPages20260106132755{})
		},
	})
}
