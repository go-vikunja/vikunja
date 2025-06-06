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

type linkShares20210411113105 struct {
	Password    string `xorm:"text null"`
	SharingType int    `xorm:"bigint INDEX not null default 0"`
}

func (linkShares20210411113105) TableName() string {
	return "link_shares"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210411113105",
		Description: "Add password field to link shares",
		Migrate: func(tx *xorm.Engine) error {

			// Make all existing share links type 1 (no password)
			if _, err := tx.Update(&linkShares20210411113105{SharingType: 1}); err != nil {
				return err
			}

			return tx.Sync2(linkShares20210411113105{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
