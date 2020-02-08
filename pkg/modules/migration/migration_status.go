// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/api/pkg/user"
)

// Status represents this migration status
type Status struct {
	ID           int64              `xorm:"int(11) autoincr not null unique pk" json:"id"`
	UserID       int64              `xorm:"int(11) not null" json:"-"`
	MigratorName string             `xorm:"varchar(255)" json:"migrator_name"`
	Created      timeutil.TimeStamp `xorm:"created not null 'created_unix'" json:"time_unix"`
}

// TableName holds the table name for the migration status table
func (s *Status) TableName() string {
	return "migration_status"
}

// SetMigrationStatus sets the migration status for a user
func SetMigrationStatus(m Migrator, u *user.User) (err error) {
	status := &Status{
		UserID:       u.ID,
		MigratorName: m.Name(),
	}
	_, err = x.Insert(status)
	return
}

// GetMigrationStatus returns the migration status for a migration and a user
func GetMigrationStatus(m Migrator, u *user.User) (status *Status, err error) {
	status = &Status{}
	_, err = x.Where("user_id = ? and migrator_name = ?", u.ID, m.Name()).Desc("id").Get(status)
	return
}
