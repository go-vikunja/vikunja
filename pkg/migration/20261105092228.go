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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type userFileIndexes20261105092228 struct {
	AvatarFileID int64 `xorm:"null index"`
	ExportFileID int64 `xorm:"bigint null index"`
}

func (userFileIndexes20261105092228) TableName() string {
	return "users"
}

type projectFileIndexes20261105092228 struct {
	BackgroundFileID int64 `xorm:"null index"`
}

func (projectFileIndexes20261105092228) TableName() string {
	return "projects"
}

type migrationStatusFileIndexes20261105092228 struct {
	UploadFileID *int64 `xorm:"bigint null index"`
}

func (migrationStatusFileIndexes20261105092228) TableName() string {
	return "migration_status"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261105092228",
		Description: "Add indexes on the file id columns of users, projects and migration_status",
		Migrate: func(tx *xorm.Engine) error {
			for _, table := range []any{
				userFileIndexes20261105092228{},
				projectFileIndexes20261105092228{},
				migrationStatusFileIndexes20261105092228{},
			} {
				if err := partialSync(tx, table); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
