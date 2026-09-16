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

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type duplicateSubscription20260829190000 struct {
	EntityType int64
	EntityID   int64
	UserID     int64
	KeepID     int64
}

func dedupeSubscriptions20260829190000(tx *xorm.Engine) error {
	// Duplicates are always identical rows — muted ships in this same release and
	// defaults to false — so keeping the lowest id loses nothing.
	duplicates := []*duplicateSubscription20260829190000{}
	err := tx.Table("subscriptions").
		Select("entity_type, entity_id, user_id, MIN(id) AS keep_id").
		GroupBy("entity_type, entity_id, user_id").
		Having("COUNT(*) > 1").
		Find(&duplicates)
	if err != nil {
		return fmt.Errorf("could not look up duplicate subscriptions: %w", err)
	}

	for _, d := range duplicates {
		_, err := tx.Table("subscriptions").
			Where("entity_type = ? AND entity_id = ? AND user_id = ? AND id > ?", d.EntityType, d.EntityID, d.UserID, d.KeepID).
			Delete()
		if err != nil {
			return fmt.Errorf("could not delete duplicate subscriptions of user %d for entity %d: %w", d.UserID, d.EntityID, err)
		}
	}

	return nil
}

func addUniqueSubscriptionIndex20260829190000(tx *xorm.Engine) error {
	if err := dedupeSubscriptions20260829190000(tx); err != nil {
		return err
	}

	// Name must match what xorm derives from the unique(entity_user) tag on the
	// Subscription struct, otherwise a later sync creates a second copy.
	query := "CREATE UNIQUE INDEX IF NOT EXISTS UQE_subscriptions_entity_user ON subscriptions (entity_type, entity_id, user_id)"
	if db.Type() == schemas.MYSQL {
		// MySQL lacks IF NOT EXISTS on CREATE INDEX, so the duplicate-index error is tolerated below instead.
		query = "CREATE UNIQUE INDEX UQE_subscriptions_entity_user ON subscriptions (entity_type, entity_id, user_id)"
	}

	_, err := tx.Exec(query)
	if err != nil && !strings.Contains(err.Error(), "Duplicate key name") {
		return fmt.Errorf("could not create unique index on subscriptions: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260829190000",
		Description: "Add unique index on subscriptions (entity_type, entity_id, user_id) and remove duplicate rows",
		Migrate:     addUniqueSubscriptionIndex20260829190000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
