//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTeamMember_Create(t *testing.T) {

	// Dummy team member
	dummyteammember := TeamMember{
		TeamID:   1,
		Username: "user3",
	}

	// Doer
	doer, err := GetUserByID(1)
	assert.NoError(t, err)

	// Insert a new team member
	allowed, _ := dummyteammember.CanCreate(doer)
	assert.True(t, allowed)
	err = dummyteammember.Create(doer)
	assert.NoError(t, err)

	// Check he's in there
	team := Team{ID: 1}
	err = team.ReadOne()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(team.Members))

	// Try inserting a user twice
	err = dummyteammember.Create(doer)
	assert.Error(t, err)
	assert.True(t, IsErrUserIsMemberOfTeam(err))

	// Delete it
	allowed, _ = dummyteammember.CanDelete(doer)
	assert.True(t, allowed)
	err = dummyteammember.Delete()
	assert.NoError(t, err)

	// Delete the other one
	tm := TeamMember{TeamID: 1, Username: "user2"}
	err = tm.Delete()
	assert.NoError(t, err)

	// Try deleting the last one
	tm = TeamMember{TeamID: 1, Username: "user1"}
	err = tm.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrCannotDeleteLastTeamMember(err))

	// Try inserting a user which does not exist
	dummyteammember.Username = "user9484"
	err = dummyteammember.Create(doer)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Try adding a user to a team which does not exist
	tm = TeamMember{TeamID: 94824, Username: "user1"}
	err = tm.Create(doer)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))
}
