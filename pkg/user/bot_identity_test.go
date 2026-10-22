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
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_IsBotOwnedBy(t *testing.T) {
	tests := []struct {
		name string
		u    *User
		a    *User
		want bool
	}{
		{
			name: "bot owned by the caller",
			u:    &User{ID: 23, BotOwnerID: 21},
			a:    &User{ID: 21},
			want: true,
		},
		{
			name: "bot owned by someone else",
			u:    &User{ID: 24, BotOwnerID: 22},
			a:    &User{ID: 21},
			want: false,
		},
		{
			name: "not a bot",
			u:    &User{ID: 21},
			a:    &User{ID: 21},
			want: false,
		},
		{
			name: "owner asked about themselves",
			u:    &User{ID: 21},
			a:    &User{ID: 21, BotOwnerID: 0},
			want: false,
		},
		{
			name: "caller is the bot, subject is the owner",
			u:    &User{ID: 21},
			a:    &User{ID: 23, BotOwnerID: 21},
			want: false,
		},
		{
			name: "sibling bot of the same owner",
			u:    &User{ID: 25, BotOwnerID: 21},
			a:    &User{ID: 23, BotOwnerID: 21},
			want: false,
		},
		{
			// IsBot prevents a zero-ID owner from matching a human.
			name: "human subject, zero-id owner",
			u:    &User{ID: 5},
			a:    &User{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.u.IsBotOwnedBy(tt.a))
		})
	}
}

func TestSameBotIdentityCond(t *testing.T) {
	tests := []struct {
		name string
		// JWT auth carries only an ID, so the identity must resolve from the database.
		auth *User
		want []int64
	}{
		{
			name: "human with two bots",
			auth: &User{ID: 21},
			want: []int64{21, 23, 25},
		},
		{
			name: "bot resolves to its owner, siblings included",
			auth: &User{ID: 23},
			want: []int64{21, 23, 25},
		},
		{
			name: "sibling bot resolves to the same set",
			auth: &User{ID: 25},
			want: []int64{21, 23, 25},
		},
		{
			name: "other owner is isolated",
			auth: &User{ID: 22},
			want: []int64{22, 24},
		},
		{
			name: "other owner's bot is isolated",
			auth: &User{ID: 24},
			want: []int64{22, 24},
		},
		{
			name: "human without bots matches only itself",
			auth: &User{ID: 1},
			want: []int64{1},
		},
		{
			name: "nonexisting user matches nothing",
			auth: &User{ID: 9999},
			want: []int64{},
		},
		{
			name: "zero id matches nothing",
			auth: &User{},
			want: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			users := []*User{}
			require.NoError(t, s.Where(SameBotIdentityCond(tt.auth, "id")).Find(&users))

			ids := make([]int64, 0, len(users))
			for _, u := range users {
				ids = append(ids, u.ID)
			}
			assert.ElementsMatch(t, tt.want, ids)
		})
	}
}
