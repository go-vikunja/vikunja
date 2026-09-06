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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usersBefore20260906221630 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	Username string `xorm:"varchar(250) not null unique"`
	IsAdmin  bool   `xorm:"not null default false"`
}

func (usersBefore20260906221630) TableName() string {
	return "users"
}

type usersAfter20260906221630 struct {
	ID            int64  `xorm:"bigint autoincr not null unique pk"`
	Username      string `xorm:"varchar(250) not null unique"`
	IsAdmin       bool   `xorm:"not null default false"`
	IsInstanceBot bool   `xorm:"not null default false"`
}

func (usersAfter20260906221630) TableName() string {
	return "users"
}

func TestAddIsInstanceBot20260906221630(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := usersBefore20260906221630{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table)) //nolint:forbidigo // test-local table

	_, err = x.Insert(&usersBefore20260906221630{ID: 1, Username: "user1", IsAdmin: true})
	require.NoError(t, err)

	require.NoError(t, partialSync(x, users20260906221630{}))

	after := []*usersAfter20260906221630{}
	require.NoError(t, x.Find(&after))
	require.Len(t, after, 1)
	assert.False(t, after[0].IsInstanceBot, "existing rows default to a non-bot")
	assert.True(t, after[0].IsAdmin)

	// Idempotent: re-running on an upgraded schema must not fail.
	require.NoError(t, partialSync(x, users20260906221630{}))
}
