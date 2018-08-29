package models

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"reflect"
)

func TestTeamNamespace(t *testing.T) {
	// Dummy team <-> namespace relation
	tn := TeamNamespace{
		TeamID: 1,
		NamespaceID: 1,
		Right: TeamRightAdmin,
	}

	dummyuser, _, err := GetUserByID(1)
	assert.NoError(t, err)

	// Test normal creation
	assert.True(t, tn.CanCreate(&dummyuser))
	err = tn.Create(&dummyuser)
	assert.NoError(t, err)

	// Test again (should fail)
	err = tn.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrTeamAlreadyHasAccess(err))

	// Test with invalid team right
	tn2 := tn
	tn2.Right = TeamRightUnknown
	err = tn2.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrInvalidTeamRight(err))

	// Check with inexistant team
	tn3 := tn
	tn3.TeamID = 324
	err = tn3.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Check with a namespace which does not exist
	tn4 := tn
	tn4.NamespaceID = 423
	err = tn4.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Check readall
	teams, err := tn.ReadAll(&dummyuser)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
	s := reflect.ValueOf(teams)
	assert.Equal(t, s.Len(), 1)

	// Check readall for a nonexistant namespace
	_, err = tn4.ReadAll(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Check with no right to read the namespace
	nouser := &User{ID: 393}
	_, err = tn.ReadAll(nouser)
	assert.Error(t, err)
	assert.True(t, IsErrNeedToHaveNamespaceReadAccess(err))

	// Delete it
	assert.True(t, tn.CanDelete(&dummyuser))
	err = tn.Delete()
	assert.NoError(t, err)

	// Try deleting with a nonexisting team
	err = tn3.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Try deleting with a nonexistant namespace
	err = tn4.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

}