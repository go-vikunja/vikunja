package models

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestList_ReadAll(t *testing.T) {
	// Create test database
	//assert.NoError(t, PrepareTestDatabase())

	// Get all lists for our namespace
	lists, err := GetListsByNamespaceID(1)
	assert.NoError(t, err)
	assert.Equal(t, len(lists), 2)

	// Get all lists our user has access to
	user, err := GetUserByID(1)
	assert.NoError(t, err)

	lists2 := List{}
	lists3, err := lists2.ReadAll(&user)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(lists3).Kind(), reflect.Slice)
	s := reflect.ValueOf(lists3)
	assert.Equal(t, s.Len(), 1)

	// Try getting lists for a nonexistant user
	_, err = lists2.ReadAll(&User{ID: 984234})
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))
}
