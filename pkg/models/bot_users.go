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

package models

import (
	"strings"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// BotUser is a thin wrapper around user.User that implements CRUDable + Permissions
// for bot-management endpoints. Ownership lives on users.bot_owner_id, so there is
// no separate table.
type BotUser struct {
	// ID shadows user.User.ID so the generic WebHandler can bind the :bot path parameter.
	ID int64 `xorm:"-" json:"id" param:"bot"`

	user.User `xorm:"extends"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName returns the table name for bot users (shared with users).
func (*BotUser) TableName() string { return "users" }

// Create creates a new bot user.
func (b *BotUser) Create(s *xorm.Session, a web.Auth) error {
	owner, ok := a.(*user.User)
	if !ok {
		return ErrGenericForbidden{}
	}
	// Reset embedded ID before insert so xorm does not try to reuse the shadowed value.
	b.User.ID = 0
	created, err := user.CreateBotUser(s, &b.User, owner)
	if err != nil {
		return err
	}
	b.User = *created
	b.ID = created.ID
	return nil
}

// ReadAll returns all bots owned by the calling user.
func (b *BotUser) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	limit, start := getLimitFromPageIndex(page, perPage)
	var bots []*user.User
	q := s.Where("bot_owner_id = ?", a.GetID())
	if search != "" {
		q = q.And("(username LIKE ? OR name LIKE ?)", "%"+search+"%", "%"+search+"%")
	}
	if limit > 0 {
		q = q.Limit(limit, start)
	}
	total, err := q.FindAndCount(&bots)
	if err != nil {
		return nil, 0, 0, err
	}
	return bots, len(bots), total, nil
}

// ReadOne returns a single bot user if owned by the caller.
func (b *BotUser) ReadOne(s *xorm.Session, a web.Auth) error {
	u, err := user.GetUserByID(s, b.ID)
	if err != nil {
		return err
	}
	if u.BotOwnerID != a.GetID() {
		return &user.ErrBotNotOwned{UserID: b.ID}
	}
	b.User = *u
	b.ID = u.ID
	return nil
}

// Update allows a narrow set of fields to be changed on an owned bot.
func (b *BotUser) Update(s *xorm.Session, a web.Auth) error {
	existing, err := user.GetUserByID(s, b.ID)
	if err != nil {
		return err
	}
	if existing.BotOwnerID != a.GetID() {
		return &user.ErrBotNotOwned{UserID: b.ID}
	}
	existing.Name = b.Name
	if b.Status == user.StatusActive || b.Status == user.StatusDisabled {
		existing.Status = b.Status
	}
	if b.Username != "" && b.Username != existing.Username {
		if !strings.HasPrefix(b.Username, "bot-") {
			return &user.ErrBotUsernameMustHavePrefix{Username: b.Username}
		}
		existing.Username = b.Username
	}
	_, err = s.ID(existing.ID).Cols("name", "status", "username").Update(existing)
	b.User = *existing
	b.ID = existing.ID
	return err
}

// Delete completely removes the bot user and all associated data.
func (b *BotUser) Delete(s *xorm.Session, a web.Auth) error {
	existing, err := user.GetUserByID(s, b.ID)
	if err != nil {
		return err
	}
	if existing.BotOwnerID != a.GetID() {
		return &user.ErrBotNotOwned{UserID: b.ID}
	}
	return DeleteUser(s, existing)
}
