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

	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type user20260908112000 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	Username string `xorm:"varchar(250) not null unique"`
}

func (user20260908112000) TableName() string {
	return "users"
}

func migrateAtUsernames20260908112000(tx *xorm.Engine) error {
	sess := tx.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("could not start transaction to migrate @ usernames: %w", err)
	}

	var users []*user20260908112000
	err := sess.
		Where(builder.Like{"username", "@%"}).
		OrderBy("id").
		Find(&users)
	if err != nil {
		return fmt.Errorf("could not fetch users with @ prefix: %w", err)
	}

	for _, u := range users {
		oldUsername := u.Username
		cleaned := strings.TrimLeft(oldUsername, "@")
		if cleaned == "" {
			cleaned = fmt.Sprintf("user_%d", u.ID)
		}

		// Resolve collisions against existing usernames (e.g. @bob vs bob -> bob_1) to satisfy the unique constraint.
		newUsername := cleaned
		suffix := 1
		for {
			exists, err := sess.
				Where(builder.Eq{"username": newUsername}).
				Exist(&user20260908112000{})
			if err != nil {
				return fmt.Errorf("could not check if username %s exists: %w", newUsername, err)
			}
			if !exists {
				break
			}
			newUsername = fmt.Sprintf("%s_%d", cleaned, suffix)
			suffix++
		}

		u.Username = newUsername
		_, err = sess.
			ID(u.ID).
			Cols("username").
			Update(u)
		if err != nil {
			return fmt.Errorf("could not update username from %s to %s for user %d: %w", oldUsername, newUsername, u.ID, err)
		}

		log.Warningf("Migrated username %q to %q for user id %d (usernames starting with @ are no longer permitted)", oldUsername, newUsername, u.ID)
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("could not commit @ username migration: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260908112000",
		Description: "Migrate usernames starting with @ to remove leading @",
		Migrate:     migrateAtUsernames20260908112000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
