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

func TestListTask_Create(t *testing.T) {
	//assert.NoError(t, LoadFixtures())

	// Fake list task
	listtask := ListTask{
		Text:        "Lorem",
		Description: "Lorem Ipsum BACKERY",
		ListID:      1,
	}

	// Add one point to a list
	doer, err := GetUserByID(1)
	assert.NoError(t, err)

	allowed, _ := listtask.CanCreate(&doer)
	assert.True(t, allowed)

	err = listtask.Create(&doer)
	assert.NoError(t, err)

	// Update it
	listtask.Text = "Test34"
	allowed, _ = listtask.CanUpdate(&doer)
	assert.True(t, allowed)
	err = listtask.Update()
	assert.NoError(t, err)

	// Check if it was updated
	li, err := GetListTaskByID(listtask.ID)
	assert.NoError(t, err)
	assert.Equal(t, li.Text, "Test34")

	// Delete the task
	allowed, _ = listtask.CanDelete(&doer)
	assert.True(t, allowed)
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
