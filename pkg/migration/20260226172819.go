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

type oauthCodes20260226172819 struct {
	ID                  int64  `xorm:"autoincr not null unique pk"`
	UserID              int64  `xorm:"bigint not null"`
	Code                string `xorm:"varchar(128) not null unique index"`
	ExpiresAt           string `xorm:"not null"`
	ClientID            string `xorm:"varchar(255) not null"`
	RedirectURI         string `xorm:"text not null"`
	CodeChallenge       string `xorm:"varchar(128) not null"`
	CodeChallengeMethod string `xorm:"varchar(10) not null"`
	Created             string `xorm:"created not null"`
}

func (oauthCodes20260226172819) TableName() string {
	return "oauth_codes"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260226172819",
		Description: "add oauth_codes table for OAuth 2.0 authorization codes",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(oauthCodes20260226172819{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(oauthCodes20260226172819{})
		},
	})
}
