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
	"testing"

	"code.vikunja.io/api/pkg/db"
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

	view := &BotUser{ID: bot.ID}
	canRead, _, err := view.CanRead(s, other)
	require.NoError(t, err)
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

	upd := &BotUser{ID: bot.ID, Status: user.StatusDisabled, User: user.User{Name: "Renamed"}}
	require.NoError(t, upd.Update(s, owner))
	assert.Equal(t, user.StatusDisabled, upd.Status)
	assert.Equal(t, "Renamed", upd.Name)
}

func TestBotUser_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	bot := &BotUser{User: user.User{Username: "bot-delete"}}
	require.NoError(t, bot.Create(s, owner))

	del := &BotUser{ID: bot.ID}
	require.NoError(t, del.Delete(s, owner))
}
