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
	"code.vikunja.io/api/pkg/models"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type linkSharing20190818210133 struct {
	ID          int64              `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Hash        string             `xorm:"varchar(40) not null unique" json:"hash" param:"hash"`
	ListID      int64              `xorm:"int(11) not null" json:"list_id"`
	Right       models.Permission  `xorm:"int(11) INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`
	SharingType models.SharingType `xorm:"int(11) INDEX not null default 0" json:"sharing_type" valid:"length(0|2)" maximum:"2" default:"0"`
	SharedByID  int64              `xorm:"int(11) INDEX not null"`
	Created     int64              `xorm:"created not null" json:"created"`
	Updated     int64              `xorm:"updated not null" json:"updated"`
}

// TableName holds the table name for this share
func (linkSharing20190818210133) TableName() string {
	return "link_sharing"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20190818210133",
		Description: "Add link sharing table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(linkSharing20190818210133{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(linkSharing20190818210133{})
		},
	})
}
