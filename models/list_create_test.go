package models

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestList_Create(t *testing.T) {
	// Create test database
	assert.NoError(t, PrepareTestDatabase())

	// Get our doer
	doer, _, err := GetUserByID(1)
	assert.NoError(t, err)

	// Dummy list for testing
	dummylist := List{
		Title: "test",
		Description: "Lorem Ipsum",
	}

	// Create it
	err = dummylist.Create(&doer)
	assert.NoError(t, err)

	// Get the list
	newdummy := List{ID:dummylist.ID}
	err = newdummy.ReadOne()
	assert.NoError(t, err)
	assert.Equal(t, dummylist.Title, newdummy.Title)
	assert.Equal(t, dummylist.Description, newdummy.Description)
	assert.Equal(t, dummylist.OwnerID, doer.ID)

	// Check failing with no title
	list2 := List{}
	err = list2.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrListTitleCannotBeEmpty(err))
}
