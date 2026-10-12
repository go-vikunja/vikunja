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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func deleteOrphanedReactions20261012090000(tx *xorm.Engine) error {
	if _, err := tx.Exec("DELETE FROM reactions WHERE user_id NOT IN (SELECT id FROM users)"); err != nil {
		return fmt.Errorf("could not delete orphaned reactions: %w", err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261012090000",
		Description: "Remove reactions whose user does not exist; old user deletions and link shares left them behind.",
		Migrate:     deleteOrphanedReactions20261012090000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
