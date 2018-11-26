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

func TestList_Create(t *testing.T) {
	// Create test database
	//assert.NoError(t, PrepareTestDatabase())

	// Get our doer
	doer, err := GetUserByID(1)
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
