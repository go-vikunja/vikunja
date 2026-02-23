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

type taskTemplates20260223120000 struct {
	ID          int64   `xorm:"autoincr not null unique pk"`
	Title       string  `xorm:"varchar(250) not null"`
	Description string  `xorm:"longtext null"`
	Priority    int64   `xorm:"bigint null"`
	HexColor    string  `xorm:"varchar(6) null"`
	PercentDone float64 `xorm:"DOUBLE null"`
	RepeatAfter int64   `xorm:"bigint null"`
	RepeatMode  int     `xorm:"not null default 0"`
	LabelIDs    string  `xorm:"json null"`
	OwnerID     int64   `xorm:"bigint not null INDEX"`
	Created     time.Time `xorm:"created not null"`
	Updated     time.Time `xorm:"updated not null"`
}

func (taskTemplates20260223120000) TableName() string {
	return "task_templates"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260223120000",
		Description: "Add task templates table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(taskTemplates20260223120000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
