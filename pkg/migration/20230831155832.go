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

type apiTokens20230831155832 struct {
	ID             int64               `xorm:"bigint autoincr not null unique pk" json:"id" param:"token"`
	Title          string              `xorm:"not null" json:"title"`
	TokenSalt      string              `xorm:"not null" json:"-"`
	TokenHash      string              `xorm:"not null unique" json:"-"`
	TokenLastEight string              `xorm:"not null index varchar(8)" json:"-"`
	Permissions    map[string][]string `xorm:"json not null" json:"permissions"`
	ExpiresAt      time.Time           `xorm:"not null" json:"expires_at"`
	OwnerID        int64               `xorm:"bigint not null" json:"-"`
	Created        time.Time           `xorm:"created not null" json:"created"`
}

func (apiTokens20230831155832) TableName() string {
	return "api_tokens"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20230831155832",
		Description: "",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(apiTokens20230831155832{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
