package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList_Create(t *testing.T) {
	// Create test database
	assert.NoError(t, PrepareTestDatabase())

	// Get our doer
	doer, _, err := GetUserByID(1)
	assert.NoError(t, err)

	// Dummy list for testing
	dummylist := List{
		Title:       "test",
		Description: "Lorem Ipsum",
		NamespaceID: 1,
	}

	// Check if the user can create
	assert.True(t, dummylist.CanCreate(&doer))

	// Create it
	err = dummylist.Create(&doer)
	assert.NoError(t, err)

	// Get the list
	newdummy := List{ID: dummylist.ID}
	err = newdummy.ReadOne()
	assert.NoError(t, err)
	assert.Equal(t, dummylist.Title, newdummy.Title)
	assert.Equal(t, dummylist.Description, newdummy.Description)
	assert.Equal(t, dummylist.OwnerID, doer.ID)

	// Check if the user can see it
	assert.True(t, dummylist.CanRead(&doer))

	// Try updating a list
	assert.True(t, dummylist.CanUpdate(&doer))
	dummylist.Description = "Lorem Ipsum dolor sit amet."
	err = dummylist.Update()
	assert.NoError(t, err)

	// Delete it
	assert.True(t, dummylist.CanDelete(&doer))

	err = dummylist.Delete()
	assert.NoError(t, err)

	// Try updating a nonexistant list
	err = dummylist.Update()
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Delete a nonexistant list
	err = dummylist.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Check failing with no title
	list2 := List{}
	err = list2.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrListTitleCannotBeEmpty(err))

	// Check creation with a nonexistant namespace
	list3 := List{
		Title:       "test",
		Description: "Lorem Ipsum",
		NamespaceID: 876694,
	}

	err = list3.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Try creating with a nonexistant owner
	nUser := &User{ID: 9482385}
	err = dummylist.Create(nUser)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))
}
