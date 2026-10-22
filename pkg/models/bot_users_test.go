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
	"context"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBotUser_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner, err := user.GetUserByID(s, 1)
		require.NoError(t, err)

		bot := &BotUser{User: user.User{Username: "bot-model-success"}}
		require.NoError(t, bot.Create(s, owner))
		assert.True(t, bot.IsBot())
		assert.Equal(t, owner.ID, bot.BotOwnerID)
	})
	t.Run("bot cannot create bot", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		botOwner := &user.User{ID: 555, BotOwnerID: 1}
		bot := &BotUser{User: user.User{Username: "bot-child"}}
		err := bot.Create(s, botOwner)
		require.Error(t, err)
		assert.True(t, user.IsErrBotNotOwned(err))
	})
}

func TestBotUser_ReadAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	bot := &BotUser{User: user.User{Username: "bot-readall"}}
	require.NoError(t, bot.Create(s, owner))

	list := &BotUser{}
	result, _, _, err := list.ReadAll(s, owner, "", 1, 50)
	require.NoError(t, err)
	bots, ok := result.([]*BotUser)
	require.True(t, ok)
	found := false
	for _, u := range bots {
		if u.Username == "bot-readall" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestBotUser_CanRead_NotOwned(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	other, err := user.GetUserByID(s, 2)
	require.NoError(t, err)

	bot := &BotUser{User: user.User{Username: "bot-notowned"}}
	require.NoError(t, bot.Create(s, owner))

	view := &BotUser{User: user.User{ID: bot.ID}}
	canRead, _, err := view.CanRead(s, other)
	require.True(t, user.IsErrUserDoesNotExist(err))
	assert.False(t, canRead)
}

func TestBotUser_Update_Status(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	bot := &BotUser{User: user.User{Username: "bot-update"}}
	require.NoError(t, bot.Create(s, owner))

	upd := &BotUser{Status: new(user.StatusDisabled), User: user.User{ID: bot.ID, Name: "Renamed"}}
	require.NoError(t, upd.Update(s, owner))
	assert.Equal(t, user.StatusDisabled, *upd.Status)
	assert.Equal(t, "Renamed", upd.Name)
}

func TestBotUser_Update_OmittedStatusKeepsLocked(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner := &user.User{ID: 1}
	bot := &BotUser{User: user.User{Username: "bot-locked-rename"}}
	require.NoError(t, bot.Create(s, owner))
	_, err := s.ID(bot.ID).Cols("status").Update(&user.User{Status: user.StatusAccountLocked})
	require.NoError(t, err)

	upd := &BotUser{User: user.User{ID: bot.ID, Name: "Renamed"}}
	require.NoError(t, upd.Update(s, owner))
	assert.Equal(t, user.StatusAccountLocked, *upd.Status)
	assert.Equal(t, "Renamed", upd.Name)
}

func TestBotUser_Update_StatusEvent(t *testing.T) {
	owner := &user.User{ID: 1}

	t.Run("status change dispatches bot.status.changed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		bot := &BotUser{User: user.User{Username: "bot-event"}}
		require.NoError(t, bot.Create(s, owner))
		upd := &BotUser{Status: new(user.StatusDisabled), User: user.User{ID: bot.ID}}
		require.NoError(t, upd.Update(s, owner))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		evt := singleDispatchedEvent[*BotStatusChangedEvent](t)
		assert.Equal(t, owner.ID, evt.Doer.ID)
		assert.Equal(t, bot.ID, evt.Bot.ID)
		assert.Equal(t, user.StatusActive, evt.OldStatus)
		assert.Equal(t, user.StatusDisabled, evt.NewStatus)
	})

	t.Run("unchanged status dispatches nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		s := db.NewSession()
		defer s.Close()

		bot := &BotUser{User: user.User{Username: "bot-no-event"}}
		require.NoError(t, bot.Create(s, owner))
		upd := &BotUser{Status: new(user.StatusActive), User: user.User{ID: bot.ID, Name: "Renamed"}}
		require.NoError(t, upd.Update(s, owner))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)

		assert.Zero(t, events.CountDispatchedEvents((&BotStatusChangedEvent{}).Name()))
	})
}

func TestBotUser_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	bot := &BotUser{User: user.User{Username: "bot-delete"}}
	require.NoError(t, bot.Create(s, owner))

	del := &BotUser{User: user.User{ID: bot.ID}}
	require.NoError(t, del.Delete(s, owner))
}

func TestBotUser_ManageDisabled(t *testing.T) {
	for _, status := range []user.Status{user.StatusDisabled, user.StatusAccountLocked} {
		t.Run(status.String(), func(t *testing.T) {
			testManageInactiveBot(t, status)
		})
	}
}

func testManageInactiveBot(t *testing.T, status user.Status) {
	for _, operation := range []string{"read", "update", "delete", "tokens"} {
		t.Run(operation, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			owner := &user.User{ID: 1}
			other := &user.User{ID: 2}
			bot := &BotUser{User: user.User{Username: "bot-disabled"}}
			require.NoError(t, bot.Create(s, owner))
			_, err := s.ID(bot.ID).Cols("status").Update(&user.User{Status: status})
			require.NoError(t, err)
			_, err = user.GetUserByID(s, bot.ID)
			require.True(t, user.IsErrUserStatusError(err))
			target := &BotUser{User: user.User{ID: bot.ID}}
			switch operation {
			case "read":
				allowed, _, err := target.CanRead(s, owner)
				require.NoError(t, err)
				require.True(t, allowed)
				_, _, err = target.CanRead(s, other)
				require.True(t, user.IsErrUserDoesNotExist(err))
				require.NoError(t, target.ReadOne(s, owner))
				assert.Equal(t, status, *target.Status)
			case "update":
				allowed, err := target.CanUpdate(s, owner)
				require.NoError(t, err)
				require.True(t, allowed)
				_, err = target.CanUpdate(s, other)
				require.True(t, user.IsErrUserDoesNotExist(err))
				target.Status = new(user.StatusActive)
				require.NoError(t, target.Update(s, owner))
				assert.Equal(t, user.StatusActive, *target.Status)
				reloaded, err := user.GetUserByID(s, bot.ID)
				require.NoError(t, err)
				assert.Equal(t, user.StatusActive, reloaded.Status)
			case "delete":
				allowed, err := target.CanDelete(s, owner)
				require.NoError(t, err)
				require.True(t, allowed)
				_, err = target.CanDelete(s, other)
				require.True(t, user.IsErrUserDoesNotExist(err))
				require.NoError(t, target.Delete(s, owner))
				_, err = user.GetUserByID(s, bot.ID)
				require.True(t, user.IsErrUserDoesNotExist(err))
			case "tokens":
				token := &APIToken{
					OwnerID: bot.ID,
					Title:   "Disabled bot token",
				}
				require.NoError(t, token.Create(s, owner))
				result, count, total, err := token.ReadAll(s, owner, "", 1, 50)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
				assert.Equal(t, int64(1), total)
				assert.Equal(t, bot.ID, result.([]*APIToken)[0].OwnerID)
				_, _, _, err = token.ReadAll(s, other, "", 1, 50)
				require.True(t, user.IsErrUserDoesNotExist(err))
				allowed, err := token.CanDelete(s, owner)
				require.NoError(t, err)
				require.True(t, allowed)
				allowed, err = token.CanDelete(s, other)
				require.NoError(t, err)
				require.False(t, allowed)
			}
		})
	}
}
