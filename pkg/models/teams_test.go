// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
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
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestTeam_Create(t *testing.T) {
	//Dummyteam
	dummyteam := Team{
		Name:        "Testteam293",
		Description: "Lorem Ispum",
	}

	// Doer
	doer, err := GetUserByID(1)
	assert.NoError(t, err)

	// Insert it
	allowed, _ := dummyteam.CanCreate(doer)
	assert.True(t, allowed)
	err = dummyteam.Create(doer)
	assert.NoError(t, err)

	// Check if it was inserted and we're admin
	tm := Team{ID: dummyteam.ID}
	err = tm.ReadOne()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tm.Members))
	assert.Equal(t, doer.ID, tm.Members[0].User.ID)
	assert.True(t, tm.Members[0].Admin)
	allowed, _ = dummyteam.CanRead(doer)
	assert.True(t, allowed)

	// Try getting a team with an ID < 0
	_, err = GetTeamByID(-1)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Get all teams the user is part of
	ts, _, _, err := tm.ReadAll(doer, "", 1, 50)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(ts).Kind(), reflect.Slice)
	s := reflect.ValueOf(ts)
	assert.Equal(t, 9, s.Len())

	// Check inserting it with an empty name
	dummyteam.Name = ""
	err = dummyteam.Create(doer)
	assert.Error(t, err)
	assert.True(t, IsErrTeamNameCannotBeEmpty(err))

	// update it (still no name, should fail)
	allowed, _ = dummyteam.CanUpdate(doer)
	assert.True(t, allowed)
	err = dummyteam.Update()
	assert.Error(t, err)
	assert.True(t, IsErrTeamNameCannotBeEmpty(err))

	// Update it, this time with a name
	dummyteam.Name = "Lorem"
	err = dummyteam.Update()
	assert.NoError(t, err)

	// Delete it
	allowed, err = dummyteam.CanDelete(doer)
	assert.NoError(t, err)
	assert.True(t, allowed)
	err = dummyteam.Delete()
	assert.NoError(t, err)

	// Try deleting a (now) nonexistant team
	allowed, err = dummyteam.CanDelete(doer)
	assert.False(t, allowed)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Try updating the (now) nonexistant team
	err = dummyteam.Update()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))
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
