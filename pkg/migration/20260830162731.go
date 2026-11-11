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
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// Completed claims use NULL because every supported database allows repeated NULLs in unique indexes.
type migrationActiveUserClaim20260830162731 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	UserID       int64     `xorm:"bigint not null"`
	MigratorName string    `xorm:"varchar(255)"`
	StartedAt    time.Time `xorm:"not null"`
	FinishedAt   time.Time `xorm:"null"`
	ActiveUserID *int64    `xorm:"bigint null unique"`
}

func (migrationActiveUserClaim20260830162731) TableName() string {
	return "migration_status"
}

func addActiveUserClaim20260830162731(tx *xorm.Engine) error {
	if err := partialSync(tx, migrationActiveUserClaim20260830162731{}); err != nil {
		return err
	}

	// partialSync skips unique constraints; xorm's derived name prevents a later sync from duplicating the index.
	query := "CREATE UNIQUE INDEX IF NOT EXISTS UQE_migration_status_active_user_id ON migration_status (active_user_id)"
	if db.Type() == schemas.MYSQL {
		// MySQL lacks CREATE INDEX IF NOT EXISTS, so tolerate its duplicate-index error below.
		query = "CREATE UNIQUE INDEX UQE_migration_status_active_user_id ON migration_status (active_user_id)"
	}

	_, err := tx.Exec(query)
	if err != nil && !strings.Contains(err.Error(), "Duplicate key name") {
		return fmt.Errorf("could not create unique index on migration_status.active_user_id: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260830162731",
		Description: "Add nullable active_user_id with unique index to migration_status to serialize migrations per user",
		Migrate:     addActiveUserClaim20260830162731,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
