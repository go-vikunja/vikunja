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

type webhooks20230913202615 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id" param:"webhook"`
	TargetURL   string    `xorm:"not null" valid:"minstringlength(1)" minLength:"1" json:"target_url"`
	Events      []string  `xorm:"JSON not null" valid:"minstringlength(1)" minLength:"1" json:"event"`
	ProjectID   int64     `xorm:"bigint not null index" json:"project_id" param:"project"`
	Secret      string    `xorm:"null" json:"secret"`
	CreatedByID int64     `xorm:"bigint not null" json:"-"`
	Created     time.Time `xorm:"created not null" json:"created"`
	Updated     time.Time `xorm:"updated not null" json:"updated"`
}

func (webhooks20230913202615) TableName() string {
	return "webhooks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20230913202615",
		Description: "",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(webhooks20230913202615{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
