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

type OAuthClient20260401221453 struct {
	ClientID     string    `xorm:"varchar(50) not null unique pk" json:"client_id"`
	ClientName   string    `xorm:"varchar(255) not null" json:"client_name"`
	RedirectURIs string    `xorm:"text not null" json:"redirect_uris"`
	Created      time.Time `xorm:"created not null" json:"created"`
}

func (OAuthClient20260401221453) TableName() string {
	return "oauth_clients"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260401221453",
		Description: "",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(OAuthClient20260401221453{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
