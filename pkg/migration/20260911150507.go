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

type userInviteLink20260911150507 struct {
	ID               int64      `xorm:"bigint autoincr not null unique pk"`
	Name             string     `xorm:"varchar(250) not null"`
	TokenHash        string     `xorm:"varchar(64) not null unique"`
	MaxUses          *int64     `xorm:"bigint null"`
	Uses             int64      `xorm:"bigint not null default 0"`
	ExpiresAt        *time.Time `xorm:"datetime null"`
	SkipEmailConfirm bool       `xorm:"not null default false"`
	CreatedByID      int64      `xorm:"bigint not null index"`
	Created          time.Time  `xorm:"created not null"`
	Updated          time.Time  `xorm:"updated not null"`
}

func (userInviteLink20260911150507) TableName() string { return "user_invite_links" }

type userInviteLinkTeam20260911150507 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	InviteLinkID int64     `xorm:"bigint not null unique(invite_team)"`
	TeamID       int64     `xorm:"bigint not null index unique(invite_team)"`
	Created      time.Time `xorm:"created not null"`
}

func (userInviteLinkTeam20260911150507) TableName() string { return "user_invite_link_teams" }
func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID: "20260911150507", Description: "Add user invite links and team assignments",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(userInviteLink20260911150507{}, userInviteLinkTeam20260911150507{}) //nolint:forbidigo // brand-new tables, nothing to drop
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(userInviteLinkTeam20260911150507{}, userInviteLink20260911150507{})
		},
	})
}
