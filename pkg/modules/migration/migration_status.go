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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
)

// Status represents this migration status
type Status struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"The unique, numeric id of this migration status."`
	UserID       int64     `xorm:"bigint not null" json:"-"`
	MigratorName string    `xorm:"varchar(255)" json:"migrator_name" readOnly:"true" doc:"The name of the migrator this status belongs to, e.g. \"todoist\"."`
	StartedAt    time.Time `xorm:"not null" json:"started_at" readOnly:"true" doc:"When the last migration started. Zero value if the user never migrated from this service."`
	FinishedAt   time.Time `xorm:"null" json:"finished_at" readOnly:"true" doc:"When the last migration finished. Zero value while a migration is still running or was never run."`
}

// TableName holds the table name for the migration status table
func (s *Status) TableName() string {
	return "migration_status"
}

// StartMigration sets the migration status for a user
func StartMigration(m MigratorName, u *user.User) (status *Status, err error) {
	s := db.NewSession()
	defer s.Close()

	status = &Status{
		UserID:       u.ID,
		MigratorName: m.Name(),
		StartedAt:    time.Now(),
	}
	_, err = s.Insert(status)
	if err != nil {
		_ = s.Rollback()
		return
	}

	return status, s.Commit()
}

// FinishMigration sets the finished at time and calls it a day
func FinishMigration(status *Status) (err error) {
	s := db.NewSession()
	defer s.Close()

	status.FinishedAt = time.Now()

	_, err = s.Where("id = ?", status.ID).Update(status)
	if err != nil {
		_ = s.Rollback()
		return
	}

	return s.Commit()
}

// GetMigrationStatus returns the migration status for a migration and a user
func GetMigrationStatus(m MigratorName, u *user.User) (status *Status, err error) {
	s := db.NewSession()
	defer s.Close()

	status = &Status{}
	_, err = s.
		Where("user_id = ? and migrator_name = ?", u.ID, m.Name()).
		Desc("id").
		Get(status)
	return
}
