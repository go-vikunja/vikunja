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
	"reflect"
	"testing"
)

func TestTeam_Create(t *testing.T) {
	doer := &user.User{
		ID:       1,
		Username: "user1",
	}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{
			Name:        "Testteam293",
			Description: "Lorem Ispum",
		}
		err := team.Create(doer)
		assert.NoError(t, err)
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{}
		err := team.Create(doer)
		assert.Error(t, err)
		assert.True(t, IsErrTeamNameCannotBeEmpty(err))
	})
}

func TestTeam_ReadOne(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{ID: 1}
		err := team.ReadOne()
		assert.NoError(t, err)
		assert.Equal(t, "testteam1", team.Name)
		assert.Equal(t, "Lorem Ipsum", team.Description)
		assert.Equal(t, int64(1), team.CreatedBy.ID)
		assert.Equal(t, int64(1), team.CreatedByID)
	})
	t.Run("invalid id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{ID: -1}
		err := team.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{ID: 99999}
		err := team.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
}

func TestTeam_ReadAll(t *testing.T) {
	doer := &user.User{ID: 1}
	t.Run("normal", func(t *testing.T) {
		team := &Team{}
		ts, _, _, err := team.ReadAll(doer, "", 1, 50)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(ts).Kind(), reflect.Slice)
		s := reflect.ValueOf(ts)
		assert.Equal(t, 8, s.Len())
	})
}

func TestTeam_Update(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{
			ID:   1,
			Name: "SomethingNew",
		}
		err := team.Update()
		assert.NoError(t, err)
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{
			ID:   1,
			Name: "",
		}
		err := team.Update()
		assert.Error(t, err)
		assert.True(t, IsErrTeamNameCannotBeEmpty(err))
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{
			ID:   9999,
			Name: "SomethingNew",
		}
		err := team.Update()
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
	})
}

func TestTeam_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		team := &Team{
			ID: 1,
		}
		err := team.Delete()
		assert.NoError(t, err)
	})
}

func TestIsErrInvalidRight(t *testing.T) {
	assert.NoError(t, RightAdmin.isValid())
	assert.NoError(t, RightRead.isValid())
	assert.NoError(t, RightWrite.isValid())

	// Check invalid
	var tr Right = 938
	err := tr.isValid()
	assert.Error(t, err)
	assert.True(t, IsErrInvalidRight(err))
}
