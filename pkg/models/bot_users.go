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

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// BotUser is a thin wrapper around user.User that implements CRUDable + Permissions
// for bot-management endpoints. Ownership lives on users.bot_owner_id, so there is
// no separate table.
type BotUser struct {
	// Status shadows user.User.Status so it is included in JSON responses
	// (the original has json:"-"). nil on update keeps the current status.
	Status *user.Status `xorm:"-" json:"status" nullable:"false" valid:"bot_status" doc:"The bot's status: 0=active, 2=disabled. Set to 2 to disable the bot, 0 to re-enable it."`

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
	b.Status = &b.User.Status
	return nil
}

// ReadAll returns all bots owned by the calling user.
func (b *BotUser) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result any, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

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
		bot.Status = &bot.User.Status
	}
	return bots, len(bots), total, nil
}

// ReadOne returns a single bot user.
func (b *BotUser) ReadOne(s *xorm.Session, a web.Auth) error {
	u, err := getOwnedBotFromAuth(s, b.ID, a)
	if err != nil {
		return err
	}
	b.User = *u
	b.Status = &b.User.Status
	return nil
}

// Update allows a narrow set of fields to be changed on an owned bot.
func (b *BotUser) Update(s *xorm.Session, a web.Auth) error {
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	existing, err := getOwnedBot(s, b.ID, doer)
	if err != nil {
		return err
	}

	cols := []string{"name"}
	existing.Name = b.Name

	oldStatus := existing.Status
	if b.Status != nil && *b.Status != oldStatus {
		existing.Status = *b.Status
		cols = append(cols, "status")
	}
	if b.Username != "" && b.Username != existing.Username {
		if !strings.HasPrefix(b.Username, "bot-") {
			return &user.ErrBotUsernameMustHavePrefix{Username: b.Username}
		}
		existing.Username = b.Username
		cols = append(cols, "username")
	}

	if _, err = s.ID(existing.ID).Cols(cols...).Update(existing); err != nil {
		return err
	}
	if existing.Status != oldStatus {
		events.DispatchOnCommit(s, &BotStatusChangedEvent{
			Bot:       existing,
			Doer:      doer,
			OldStatus: oldStatus,
			NewStatus: existing.Status,
		})
	}
	b.User = *existing
	b.Status = &b.User.Status
	return nil
}

// Delete completely removes the bot user and all associated data.
func (b *BotUser) Delete(s *xorm.Session, a web.Auth) error {
	existing, err := getOwnedBotFromAuth(s, b.ID, a)
	if err != nil {
		return err
	}
	return DeleteUser(s, existing)
}

// getOwnedBot loads a bot in any status, so owners can manage disabled or
// locked bots. Missing, non-bot and foreign ids are indistinguishable.
func getOwnedBot(s *xorm.Session, id int64, owner *user.User) (*user.User, error) {
	bot, err := loadAdminTargetUser(s, id)
	if err != nil {
		return nil, err
	}
	if !bot.IsBotOwnedBy(owner) {
		return nil, user.ErrUserDoesNotExist{UserID: id}
	}
	return bot, nil
}

func getOwnedBotFromAuth(s *xorm.Session, id int64, a web.Auth) (*user.User, error) {
	owner, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}
	return getOwnedBot(s, id, owner)
}
