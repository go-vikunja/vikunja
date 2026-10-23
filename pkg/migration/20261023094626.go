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
	"fmt"
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type webhooks20261023094626 struct {
	ID                int64     `xorm:"bigint autoincr not null unique pk"`
	TargetURL         string    `xorm:"not null"`
	Events            []string  `xorm:"JSON not null"`
	ProjectID         int64     `xorm:"bigint null index"`
	UserID            int64     `xorm:"bigint null index"`
	Secret            string    `xorm:"null"`
	BasicAuthUser     string    `xorm:"null"`
	BasicAuthPassword string    `xorm:"null"`
	CreatedByID       int64     `xorm:"bigint not null"`
	Created           time.Time `xorm:"created not null"`
	Updated           time.Time `xorm:"updated not null"`
}

func (webhooks20261023094626) TableName() string {
	return "webhooks"
}

// Fresh v0.22.0 installs never got a webhooks table; v1.1.0's partial Sync
// then created one holding only the basic auth columns (#4214).
func repairWebhooksTable20261023094626(tx *xorm.Engine) error {
	exists, err := tx.IsTableExist("webhooks")
	if err != nil {
		return err
	}
	if exists {
		hasEvents, err := columnExists(tx, "webhooks", "events")
		if err != nil {
			return err
		}
		if hasEvents {
			return nil
		}

		// Inserts into the stub always failed, so any rows here were added by hand.
		empty, err := tx.IsTableEmpty("webhooks")
		if err != nil {
			return err
		}
		if !empty {
			return fmt.Errorf("the webhooks table is missing the events column but contains rows; back them up, drop the webhooks table and restart Vikunja to recreate it")
		}

		if err := tx.DropTables("webhooks"); err != nil {
			return err
		}
	}

	return tx.Sync2(webhooks20261023094626{}) //nolint:forbidigo // table was just dropped or never existed
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261023094626",
		Description: "Recreate the webhooks table on installs where it is missing its core columns",
		Migrate:     repairWebhooksTable20261023094626,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
