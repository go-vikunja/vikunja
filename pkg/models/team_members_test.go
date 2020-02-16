// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTeamMember_Create(t *testing.T) {

	doer := &user.User{
		ID: 1,
	}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tm := &TeamMember{
			TeamID:   1,
			Username: "user3",
		}
		err := tm.Create(doer)
		assert.NoError(t, err)
	})
	t.Run("already existing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tm := &TeamMember{
			TeamID:   1,
			Username: "user1",
		}
		err := tm.Create(doer)
		assert.Error(t, err)
		assert.True(t, IsErrUserIsMemberOfTeam(err))
	})
	t.Run("nonexisting user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tm := &TeamMember{
			TeamID:   1,
			Username: "nonexistinguser",
		}
		err := tm.Create(doer)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
	t.Run("nonexisting team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tm := &TeamMember{
			TeamID:   9999999,
			Username: "user1",
		}
		err := tm.Create(doer)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
}

func TestTeamMember_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tm := &TeamMember{
			TeamID:   1,
			Username: "user1",
		}
		err := tm.Delete()
		assert.NoError(t, err)
	})
}
