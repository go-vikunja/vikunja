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

func deleteOrphanedLabelTasks20260928140648(tx *xorm.Engine) error {
	if _, err := tx.Exec("DELETE FROM label_tasks WHERE label_id NOT IN (SELECT id FROM labels)"); err != nil {
		return fmt.Errorf("could not delete orphaned label_tasks: %w", err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260928140648",
		Description: "Remove label_tasks rows whose label was deleted; label deletion did not clean them up.",
		Migrate:     deleteOrphanedLabelTasks20260928140648,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
