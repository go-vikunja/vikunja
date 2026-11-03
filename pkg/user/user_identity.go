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

package user

import (
	"time"

	"xorm.io/xorm"
)

// Matched on Issuer, not the provider key, so key renames don't break links.
type Identity struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	UserID  int64     `xorm:"bigint not null unique(user_issuer)" json:"-"`
	Issuer  string    `xorm:"varchar(250) not null collate utf8mb4_bin unique(user_issuer) unique(issuer_subject)" json:"issuer"`
	Subject string    `xorm:"varchar(255) not null collate utf8mb4_bin unique(issuer_subject)" json:"-"`
	Created time.Time `xorm:"created not null" json:"created"`
}

func (*Identity) TableName() string {
	return "user_identities"
}

func GetIdentityByIssuerAndSubject(s *xorm.Session, issuer, subject string) (identity *Identity, found bool, err error) {
	identity = &Identity{}
	found, err = s.Where("issuer = ? AND subject = ?", issuer, subject).Get(identity)
	if err != nil || !found {
		return nil, false, err
	}
	// utf8mb4_bin is PAD SPACE on MySQL/MariaDB and ignores trailing spaces.
	if identity.Issuer != issuer || identity.Subject != subject {
		return nil, false, nil
	}
	return identity, true, nil
}

func GetIdentitiesByUserID(s *xorm.Session, userID int64) (identities []*Identity, err error) {
	identities = []*Identity{}
	err = s.Where("user_id = ?", userID).OrderBy("id ASC").Find(&identities)
	return identities, err
}
