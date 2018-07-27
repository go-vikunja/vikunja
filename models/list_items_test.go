package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListItem_Create(t *testing.T) {
	//assert.NoError(t, PrepareTestDatabase())

	// Fake list item
	listitem := ListItem{
		Text:        "Lorem",
		Description: "Lorem Ipsum BACKERY",
		ListID:      1,
	}

	// Add one point to a list
	doer, _, err := GetUserByID(1)
	assert.NoError(t, err)

	assert.True(t, listitem.CanCreate(&doer))

	err = listitem.Create(&doer)
	assert.NoError(t, err)

	// Update it
	listitem.Text = "Test34"
	assert.True(t, listitem.CanUpdate(&doer))
	err = listitem.Update()
	assert.NoError(t, err)

	// Check if it was updated
	li, err := GetListItemByID(listitem.ID)
	assert.NoError(t, err)
	assert.Equal(t, li.Text, "Test34")

	// Delete the item
	assert.True(t, listitem.CanDelete(&doer))
	err = listitem.Delete()
	assert.NoError(t, err)

	// Delete a nonexistant item
	listitem.ID = 0
	err = listitem.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrListItemDoesNotExist(err))

	// Try adding a list item with an empty text
	listitem.Text = ""
	err = listitem.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrListItemCannotBeEmpty(err))

	// Try adding one to a nonexistant list
	listitem.ListID = 99993939
	listitem.Text = "Lorem Ipsum"
	err = listitem.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Try updating a nonexistant item
	listitem.ID = 94829352
	err = listitem.Update()
	assert.Error(t, err)
	assert.True(t, IsErrListItemDoesNotExist(err))

	// Try inserting an item with a nonexistant user
	nUser := &User{ID: 9482385}
	listitem.ListID = 1
	err = listitem.Create(nUser)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))
}
