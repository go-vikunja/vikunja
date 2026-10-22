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

// Rows written before this migration keep a NULL heartbeat and fall back to started_at.
type migrationStatusHeartbeat20260911193534 struct {
	HeartbeatAt *time.Time `xorm:"null"`
}

func (migrationStatusHeartbeat20260911193534) TableName() string {
	return "migration_status"
}

func addMigrationStatusHeartbeat20260911193534(tx *xorm.Engine) error {
	if err := partialSync(tx, migrationStatusHeartbeat20260911193534{}); err != nil {
		return fmt.Errorf("could not add migration_status.heartbeat_at: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260911193534",
		Description: "Add nullable heartbeat_at to migration_status so stale claims are detected from the last beat",
		Migrate:     addMigrationStatusHeartbeat20260911193534,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
