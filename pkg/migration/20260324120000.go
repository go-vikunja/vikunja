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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type LicenseStatus20260324120000 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	InstanceID  string    `xorm:"varchar(36) not null"`
	LicenseKey  string    `xorm:"text not null"`
	Response    string    `xorm:"text not null"`
	ValidatedAt time.Time `xorm:"datetime null"`
	Created     time.Time `xorm:"created not null"`
	Updated     time.Time `xorm:"updated not null"`
}

func (LicenseStatus20260324120000) TableName() string {
	return "license_status"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260324120000",
		Description: "Add license_status table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(LicenseStatus20260324120000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(LicenseStatus20260324120000{})
		},
	})
}
