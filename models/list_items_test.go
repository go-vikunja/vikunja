package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListTask_Create(t *testing.T) {
	//assert.NoError(t, PrepareTestDatabase())

	// Fake list task
	listtask := ListTask{
		Text:        "Lorem",
		Description: "Lorem Ipsum BACKERY",
		ListID:      1,
	}

	// Add one point to a list
	doer, _, err := GetUserByID(1)
	assert.NoError(t, err)

	assert.True(t, listtask.CanCreate(&doer))

	err = listtask.Create(&doer)
	assert.NoError(t, err)

	// Update it
	listtask.Text = "Test34"
	assert.True(t, listtask.CanUpdate(&doer))
	err = listtask.Update()
	assert.NoError(t, err)

	// Check if it was updated
	li, err := GetListTaskByID(listtask.ID)
	assert.NoError(t, err)
	assert.Equal(t, li.Text, "Test34")

	// Delete the task
	assert.True(t, listtask.CanDelete(&doer))
	err = listtask.Delete()
	assert.NoError(t, err)

	// Delete a nonexistant task
	listtask.ID = 0
	err = listtask.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrListTaskDoesNotExist(err))

	// Try adding a list task with an empty text
	listtask.Text = ""
	err = listtask.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrListTaskCannotBeEmpty(err))

	// Try adding one to a nonexistant list
	listtask.ListID = 99993939
	listtask.Text = "Lorem Ipsum"
	err = listtask.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Try updating a nonexistant task
	listtask.ID = 94829352
	err = listtask.Update()
	assert.Error(t, err)
	assert.True(t, IsErrListTaskDoesNotExist(err))

	// Try inserting an task with a nonexistant user
	nUser := &User{ID: 9482385}
	listtask.ListID = 1
	err = listtask.Create(nUser)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))
}
