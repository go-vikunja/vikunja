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
	"time"

	"code.vikunja.io/api/pkg/models"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type savedFilters20200906184746 struct {
	ID          int64                  `xorm:"autoincr not null unique pk" json:"id"`
	Filters     *models.TaskCollection `xorm:"JSON not null" json:"filters"`
	Title       string                 `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	Description string                 `xorm:"longtext null" json:"description"`
	OwnerID     int64                  `xorm:"int(11) not null INDEX" json:"-"`
	Created     time.Time              `xorm:"created not null" json:"created"`
	Updated     time.Time              `xorm:"updated not null" json:"updated"`
}

func (savedFilters20200906184746) TableName() string {
	return "saved_filters"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200906184746",
		Description: "Add the saved filters column",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(savedFilters20200906184746{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
