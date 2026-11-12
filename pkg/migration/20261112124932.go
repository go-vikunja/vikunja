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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type jobs20261112124932 struct {
	ID          int64     `xorm:"bigint autoincr pk"`
	Queue       string    `xorm:"varchar(100) not null index(claim)"`
	Kind        string    `xorm:"varchar(200) not null"`
	Version     int       `xorm:"not null default 1"`
	Payload     string    `xorm:"longtext not null"`
	State       string    `xorm:"varchar(20) not null index(claim)"`
	Attempt     int       `xorm:"not null default 0"`
	MaxAttempts int       `xorm:"not null default 5"`
	RunAt       time.Time `xorm:"datetime not null index(claim)"`
	ClaimedAt   time.Time `xorm:"datetime null"`
	ClaimOwner  string    `xorm:"varchar(100) null"`
	ExpiresAt   time.Time `xorm:"datetime null index"`
	UniqueKey   *string   `xorm:"varchar(64) null unique"`
	LastError   string    `xorm:"text null"`
	FinishedAt  time.Time `xorm:"datetime null index"`
}

func (jobs20261112124932) TableName() string {
	return "jobs"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261112124932",
		Description: "Add durable background jobs",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(jobs20261112124932{}) //nolint:forbidigo // New table.
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables("jobs")
		},
	})
}
