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

// varchar, not text: MySQL can't put a unique index on text.
// Binary collation: OIDC subjects are case-sensitive, MySQL's default collation is not. Ignored elsewhere.
type userIdentities20261104001459 struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk"`
	UserID  int64     `xorm:"bigint not null unique(user_issuer)"`
	Issuer  string    `xorm:"varchar(250) not null collate utf8mb4_bin unique(user_issuer) unique(issuer_subject)"`
	Subject string    `xorm:"varchar(255) not null collate utf8mb4_bin unique(issuer_subject)"`
	Created time.Time `xorm:"created not null"`
}

func (userIdentities20261104001459) TableName() string {
	return "user_identities"
}

func createUserIdentities20261104001459(tx *xorm.Engine) error {
	return tx.Sync(userIdentities20261104001459{}) //nolint:forbidigo // brand-new table, nothing to drop
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261104001459",
		Description: "Add user_identities table for OIDC identities linked to local users",
		Migrate:     createUserIdentities20261104001459,
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(userIdentities20261104001459{})
		},
	})
}
