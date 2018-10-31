package models

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestTeamList(t *testing.T) {
	// Dummy relation
	tl := TeamList{
		TeamID: 1,
		ListID: 1,
		Right:  TeamRightAdmin,
	}

	// Dummyuser
	u, err := GetUserByID(1)
	assert.NoError(t, err)

	// Check normal creation
	assert.True(t, tl.CanCreate(&u))
	err = tl.Create(&u)
	assert.NoError(t, err)

	// Check again
	err = tl.Create(&u)
	assert.Error(t, err)
	assert.True(t, IsErrTeamAlreadyHasAccess(err))

	// Check with wrong rights
	tl2 := tl
	tl2.Right = TeamRightUnknown
	err = tl2.Create(&u)
	assert.Error(t, err)
	assert.True(t, IsErrInvalidTeamRight(err))

	// Check with inexistant team
	tl3 := tl
	tl3.TeamID = 3253
	err = tl3.Create(&u)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Check with inexistant list
	tl4 := tl
	tl4.ListID = 3252
	err = tl4.Create(&u)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Test Read all
	teams, err := tl.ReadAll(&u)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
	s := reflect.ValueOf(teams)
	assert.Equal(t, s.Len(), 1)

	// Test Read all for nonexistant list
	_, err = tl4.ReadAll(&u)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Test Read all for a list where the user is owner of the namespace this list belongs to
	tl5 := tl
	tl5.ListID = 2
	_, err = tl5.ReadAll(&u)
	assert.NoError(t, err)

	// Test read all for a list where the user not has access
	tl6 := tl
	tl6.ListID = 3
	_, err = tl6.ReadAll(&u)
	assert.Error(t, err)
	assert.True(t, IsErrNeedToHaveListReadAccess(err))

	// Delete
	assert.True(t, tl.CanDelete(&u))
	err = tl.Delete()
	assert.NoError(t, err)

	// Delete a nonexistant team
	err = tl3.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Delete with a nonexistant list
	err = tl4.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotHaveAccessToList(err))

}
