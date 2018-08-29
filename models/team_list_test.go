package models

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"reflect"
)

func TestTeamList(t *testing.T) {
	// Dummy relation
	tl := TeamList{
		TeamID: 1,
		ListID: 1,
		Right: TeamRightAdmin,
	}

	// Dummyuser
	user, _, err := GetUserByID(1)
	assert.NoError(t, err)

	// Check normal creation
	assert.True(t, tl.CanCreate(&user))
	err = tl.Create(&user)
	assert.NoError(t, err)

	// Check again
	err = tl.Create(&user)
	assert.Error(t, err)
	assert.True(t, IsErrTeamAlreadyHasAccess(err))

	// Check with wrong rights
	tl2 := tl
	tl2.Right = TeamRightUnknown
	err = tl2.Create(&user)
	assert.Error(t, err)
	assert.True(t, IsErrInvalidTeamRight(err))

	// Check with inexistant team
	tl3 := tl
	tl3.TeamID = 3253
	err = tl3.Create(&user)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Check with inexistant list
	tl4 := tl
	tl4.ListID = 3252
	err = tl4.Create(&user)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Test Read all
	teams, err := tl.ReadAll(&user)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
	s := reflect.ValueOf(teams)
	assert.Equal(t, s.Len(), 1)

	// Test Read all for nonexistant list
	_, err = tl4.ReadAll(&user)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Test Read all for a list where the user does not have access
	tl5 := tl
	tl5.ListID = 2
	_, err = tl5.ReadAll(&user)
	assert.Error(t, err)
	assert.True(t, IsErrNeedToHaveListReadAccess(err))

	// Delete
	assert.True(t, tl.CanDelete(&user))
	err = tl.Delete()
	assert.NoError(t, err)

	// Delete a nonexistant team
	err = tl3.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Delete with a nonexistant list
	err = tl4.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

}
