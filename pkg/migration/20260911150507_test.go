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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

func TestMigrationUserInviteLinks(t *testing.T) {
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, engine.Sync(userInviteLink20260911150507{}, userInviteLinkTeam20260911150507{}))
	t.Cleanup(func() {
		require.NoError(t, engine.DropTables(userInviteLinkTeam20260911150507{}, userInviteLink20260911150507{}))
	})
	link := &userInviteLink20260911150507{Name: "welcome", TokenHash: "hash", CreatedByID: 1}
	_, err = engine.Insert(link)
	require.NoError(t, err)
	stored := &userInviteLink20260911150507{}
	found, err := engine.ID(link.ID).Get(stored)
	require.NoError(t, err)
	require.True(t, found)
	require.Nil(t, stored.MaxUses)
	require.Nil(t, stored.ExpiresAt)
	_, err = engine.Insert(&userInviteLink20260911150507{Name: "duplicate", TokenHash: "hash", CreatedByID: 1})
	require.Error(t, err)
	_, err = engine.Insert(&userInviteLinkTeam20260911150507{InviteLinkID: link.ID, TeamID: 1})
	require.NoError(t, err)
	_, err = engine.Insert(&userInviteLinkTeam20260911150507{InviteLinkID: link.ID, TeamID: 1})
	require.Error(t, err)
}
