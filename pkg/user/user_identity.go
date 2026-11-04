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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/notifications"

	"xorm.io/xorm"
)

// Matched on Issuer, not the provider key, so key renames don't break links.
type Identity struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"The unique, numeric id of this linked identity."`
	UserID  int64     `xorm:"bigint not null unique(user_issuer)" json:"-"`
	Issuer  string    `xorm:"varchar(250) not null collate utf8mb4_bin unique(user_issuer) unique(issuer_subject)" json:"issuer" readOnly:"true" doc:"The issuer URL from the provider's ID token."`
	Subject string    `xorm:"varchar(255) not null collate utf8mb4_bin unique(issuer_subject)" json:"-"`
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"When the identity was linked."`
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

func HasIdentityForIssuer(s *xorm.Session, userID int64, issuer string) (bool, error) {
	return s.
		Where("user_id = ? AND issuer = ?", userID, issuer).
		Exist(&Identity{})
}

// CreateIdentity mails u about the new login method once s commits, so a
// stolen session can't quietly add one.
func CreateIdentity(s *xorm.Session, u *User, issuer, subject, providerName string) (*Identity, error) {
	identity := &Identity{
		UserID:  u.ID,
		Issuer:  issuer,
		Subject: subject,
	}
	if _, err := s.Insert(identity); err != nil {
		return nil, fmt.Errorf("inserting user identity: %w", err)
	}

	events.OnCommit(s, func() error {
		return notifications.Notify(u, &IdentityLinkedNotification{
			User:         u,
			ProviderName: providerName,
			Linked:       identity.Created,
		})
	})
	return identity, nil
}

// Only deletes userID's own identity.
func DeleteIdentity(s *xorm.Session, userID, identityID int64) (deleted bool, err error) {
	n, err := s.
		Where("id = ? AND user_id = ?", identityID, userID).
		Delete(&Identity{})
	return n > 0, err
}

// OIDCUserExists reports whether issuer and subject belong to a user created through OpenID.
func OIDCUserExists(s *xorm.Session, issuer, subject string) (bool, error) {
	return s.
		Where("issuer = ? AND subject = ?", issuer, subject).
		Exist(&User{})
}
