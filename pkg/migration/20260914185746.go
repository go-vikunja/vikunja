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

type migrationStatusUploadFile20260914185746 struct {
	UploadFileID *int64 `xorm:"bigint null"`
}

func (migrationStatusUploadFile20260914185746) TableName() string {
	return "migration_status"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260914185746",
		Description: "Add upload_file_id to migration_status so a queued file import can be run by any instance",
		Migrate: func(tx *xorm.Engine) error {
			if err := partialSync(tx, migrationStatusUploadFile20260914185746{}); err != nil {
				return fmt.Errorf("could not add migration_status.upload_file_id: %w", err)
			}
			return nil
		},
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
