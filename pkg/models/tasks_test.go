// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTask_Create(t *testing.T) {
	//assert.NoError(t, LoadFixtures())

	// TODO: This test needs refactoring

	// Fake list task
	listtask := Task{
		Text:        "Lorem",
		Description: "Lorem Ipsum BACKERY",
		ListID:      1,
	}

	// Add one point to a list
	doer, err := GetUserByID(1)
	assert.NoError(t, err)

	allowed, _ := listtask.CanCreate(doer)
	assert.True(t, allowed)

	err = listtask.Create(doer)
	assert.NoError(t, err)

	// Update it
	listtask.Text = "Test34"
	allowed, _ = listtask.CanUpdate(doer)
	assert.True(t, allowed)
	err = listtask.Update()
	assert.NoError(t, err)

	// Delete the task
	allowed, _ = listtask.CanDelete(doer)
	assert.True(t, allowed)
	err = listtask.Delete()
	assert.NoError(t, err)

	// Delete a nonexistant task
	listtask.ID = 0
	_, err = listtask.CanDelete(doer) // The check if the task exists happens in CanDelete
	assert.Error(t, err)
	assert.True(t, IsErrTaskDoesNotExist(err))

	// Try adding a list task with an empty text
	listtask.Text = ""
	err = listtask.Create(doer)
	assert.Error(t, err)
	assert.True(t, IsErrTaskCannotBeEmpty(err))

	// Try adding one to a nonexistant list
	listtask.ListID = 99993939
	listtask.Text = "Lorem Ipsum"
	err = listtask.Create(doer)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Try updating a nonexistant task
	listtask.ID = 94829352
	err = listtask.Update()
	assert.Error(t, err)
	assert.True(t, IsErrTaskDoesNotExist(err))

	// Try inserting an task with a nonexistant user
	nUser := &User{ID: 9482385}
	listtask.ListID = 1
	err = listtask.Create(nUser)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))
}

func TestUpdateDone(t *testing.T) {
	t.Run("marking a task as done", func(t *testing.T) {
		oldTask := &Task{Done: false}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask)
		assert.NotEqual(t, int64(0), oldTask.DoneAtUnix)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		updateDone(oldTask, newTask)
		assert.Equal(t, int64(0), oldTask.DoneAtUnix)
	})
}

func TestTask_ReadOne(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		task := &Task{ID: 1}
		err := task.ReadOne()
		assert.NoError(t, err)
		assert.Equal(t, "task #1", task.Text)
	})
	t.Run("nonexisting", func(t *testing.T) {
		task := &Task{ID: 99999}
		err := task.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}
