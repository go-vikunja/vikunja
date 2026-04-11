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
	// Status shadows user.User.Status so it is included in JSON responses
	// (the original has json:"-").
	Status user.Status `xorm:"-" json:"status"`

	user.User `xorm:"extends"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// Create creates a new bot user.
func (b *BotUser) Create(s *xorm.Session, a web.Auth) error {
	owner, ok := a.(*user.User)
	if !ok {
		return ErrGenericForbidden{}
	}
	b.ID = 0
	created, err := user.CreateBotUser(s, &b.User, owner)
	if err != nil {
		return err
	}
	b.User = *created
	b.Status = created.Status
	return nil
}

// ReadAll returns all bots owned by the calling user.
func (b *BotUser) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result any, resultCount int, numberOfTotalItems int64, err error) {
	limit, start := getLimitFromPageIndex(page, perPage)
	var bots []*BotUser
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
	for _, bot := range bots {
		bot.Status = bot.User.Status
	}
	return bots, len(bots), total, nil
}

// ReadOne returns a single bot user.
// Ownership is verified in CanRead.
func (b *BotUser) ReadOne(s *xorm.Session, _ web.Auth) error {
	u, err := user.GetUserByID(s, b.ID)
	if err != nil {
		return err
	}
	b.User = *u
	b.Status = u.Status
	return nil
}

// Update allows a narrow set of fields to be changed on an owned bot.
// Ownership is verified in CanUpdate.
func (b *BotUser) Update(s *xorm.Session, _ web.Auth) error {
	existing, err := user.GetUserByID(s, b.ID)
	if err != nil {
		return err
	}

	cols := []string{}

	if b.Name != "" {
		existing.Name = b.Name
		cols = append(cols, "name")
	}
	if b.Status == user.StatusDisabled {
		existing.Status = b.Status
		cols = append(cols, "status")
	} else if b.Status == user.StatusActive && existing.Status != user.StatusActive {
		existing.Status = b.Status
		cols = append(cols, "status")
	}
	if b.Username != "" && b.Username != existing.Username {
		if !strings.HasPrefix(b.Username, "bot-") {
			return &user.ErrBotUsernameMustHavePrefix{Username: b.Username}
		}
		existing.Username = b.Username
		cols = append(cols, "username")
	}

	if len(cols) > 0 {
		_, err = s.ID(existing.ID).Cols(cols...).Update(existing)
	}
	b.User = *existing
	b.Status = existing.Status
	return err
}

// Delete completely removes the bot user and all associated data.
// Ownership is verified in CanDelete.
func (b *BotUser) Delete(s *xorm.Session, _ web.Auth) error {
	existing, err := user.GetUserByID(s, b.ID)
	if err != nil {
		return err
	}
	return DeleteUser(s, existing)
}
