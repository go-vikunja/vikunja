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

type migrationStatusError20260911193552 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	UserID       int64     `xorm:"bigint not null"`
	MigratorName string    `xorm:"varchar(255)"`
	StartedAt    time.Time `xorm:"not null"`
	FinishedAt   time.Time `xorm:"null"`
	ActiveUserID *int64    `xorm:"bigint null unique"`
	ErrorMessage string    `xorm:"text null"`
}

func (migrationStatusError20260911193552) TableName() string {
	return "migration_status"
}

func addMigrationStatusError20260911193552(tx *xorm.Engine) error {
	if err := partialSync(tx, migrationStatusError20260911193552{}); err != nil {
		return fmt.Errorf("could not add error_message to migration_status: %w", err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260911193552",
		Description: "Add error_message to migration_status so a finished migration says whether it succeeded",
		Migrate:     addMigrationStatusError20260911193552,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
