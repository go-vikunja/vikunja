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

type users20260609191533 struct {
	ProFeatureOverrides map[string]bool `xorm:"json null"`
}

func (users20260609191533) TableName() string {
	return "users"
}

// Mirrors models.ProFeatureInstanceDefault.
type proFeatureInstanceDefaults20260609191533 struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk"`
	Feature string    `xorm:"varchar(50) not null unique"`
	Enabled bool      `xorm:"not null"`
	Created time.Time `xorm:"created not null"`
	Updated time.Time `xorm:"updated not null"`
}

func (proFeatureInstanceDefaults20260609191533) TableName() string {
	return "pro_feature_instance_defaults"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260609191533",
		Description: "Add per-user pro feature toggles",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync(users20260609191533{}); err != nil {
				return err
			}
			return tx.Sync(proFeatureInstanceDefaults20260609191533{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
