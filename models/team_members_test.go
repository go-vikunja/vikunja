package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTeamMember_Create(t *testing.T) {

	// Dummy team member
	dummyteammember := TeamMember{
		TeamID: 1,
		UserID: 3,
	}

	// Doer
	doer, _, err := GetUserByID(1)
	assert.NoError(t, err)

	// Insert a new team member
	assert.True(t, dummyteammember.CanCreate(&doer))
	err = dummyteammember.Create(&doer)
	assert.NoError(t, err)

	// Check he's in there
	team := Team{ID: 1}
	err = team.ReadOne()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(team.Members))

	// Try inserting a user twice
	err = dummyteammember.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrUserIsMemberOfTeam(err))

	// Delete it
	assert.True(t, dummyteammember.CanDelete(&doer))
	err = dummyteammember.Delete()
	assert.NoError(t, err)

	// Delete the other one
	tm := TeamMember{TeamID: 1, UserID: 2}
	err = tm.Delete()
	assert.NoError(t, err)

	// Try deleting the last one
	tm = TeamMember{TeamID: 1, UserID: 1}
	err = tm.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrCannotDeleteLastTeamMember(err))

	// Try inserting a user which does not exist
	dummyteammember.UserID = 9484
	err = dummyteammember.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Try adding a user to a team which does not exist
	tm = TeamMember{TeamID: 94824, UserID: 1}
	err = tm.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))
}
