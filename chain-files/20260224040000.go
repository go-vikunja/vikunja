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

type taskChains20260224040000 struct {
	ID          int64     `xorm:"autoincr not null unique pk"`
	Title       string    `xorm:"varchar(250) not null"`
	Description string    `xorm:"longtext null"`
	OwnerID     int64     `xorm:"bigint not null INDEX"`
	Created     time.Time `xorm:"created not null"`
	Updated     time.Time `xorm:"updated not null"`
}

func (taskChains20260224040000) TableName() string {
	return "task_chains"
}

type taskChainSteps20260224040000 struct {
	ID           int64  `xorm:"autoincr not null unique pk"`
	ChainID      int64  `xorm:"bigint not null INDEX"`
	Sequence     int    `xorm:"int not null default 0"`
	Title        string `xorm:"varchar(250) not null"`
	Description  string `xorm:"longtext null"`
	OffsetDays   int    `xorm:"int not null default 0"`
	DurationDays int    `xorm:"int not null default 1"`
	Priority     int64  `xorm:"bigint null"`
	HexColor     string `xorm:"varchar(6) null"`
	LabelIDs     string `xorm:"json null"`
}

func (taskChainSteps20260224040000) TableName() string {
	return "task_chain_steps"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260224040000",
		Description: "Add task chains and chain steps tables",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync(taskChains20260224040000{})
			if err != nil {
				return err
			}
			return tx.Sync(taskChainSteps20260224040000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
