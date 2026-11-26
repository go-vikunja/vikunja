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

type apiTokens20261126080432 struct {
	OAuthClientID string `xorm:"text null 'oauth_client_id'"`
	OAuthResource string `xorm:"text null 'oauth_resource'"`
}

func (apiTokens20261126080432) TableName() string {
	return "api_tokens"
}

type oauthCodes20261126080432 struct {
	Scope    string `xorm:"varchar(255) null"`
	Resource string `xorm:"text null"`
}

func (oauthCodes20261126080432) TableName() string {
	return "oauth_codes"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261126080432",
		Description: "Add OAuth grant columns to api_tokens and oauth_codes",
		Migrate: func(tx *xorm.Engine) error {
			return partialSync(tx, apiTokens20261126080432{}, oauthCodes20261126080432{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
