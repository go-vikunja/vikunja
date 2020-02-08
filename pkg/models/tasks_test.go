// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTask_Create(t *testing.T) {
	usr := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	// We only test creating a task here, the rights are all well tested in the integration tests.

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Text:        "Lorem",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(usr)
		assert.NoError(t, err)
		// Assert getting a uid
		assert.NotEmpty(t, task.UID)
		// Assert getting a new index
		assert.NotEmpty(t, task.Index)
		assert.Equal(t, int64(18), task.Index)

	})
	t.Run("empty text", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Text:        "",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(usr)
		assert.Error(t, err)
		assert.True(t, IsErrTaskCannotBeEmpty(err))
	})
	t.Run("nonexistant list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			Text:        "Test",
			Description: "Lorem Ipsum Dolor",
			ListID:      9999999,
		}
		err := task.Create(usr)
		assert.Error(t, err)
		assert.True(t, IsErrListDoesNotExist(err))
	})
	t.Run("noneixtant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		nUser := &user.User{ID: 99999999}
		task := &Task{
			Text:        "Test",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Create(nUser)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
}

func TestTask_Update(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID:          1,
			Text:        "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Update()
		assert.NoError(t, err)
	})
	t.Run("nonexistant task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID:          9999999,
			Text:        "test10000",
			Description: "Lorem Ipsum Dolor",
			ListID:      1,
		}
		err := task.Update()
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}

func TestTask_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{
			ID: 1,
		}
		err := task.Delete()
		assert.NoError(t, err)
	})
}

func TestUpdateDone(t *testing.T) {
	t.Run("marking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldTask := &Task{Done: false}
		newTask := &Task{Done: true}
		updateDone(oldTask, newTask)
		assert.NotEqual(t, timeutil.TimeStamp(0), oldTask.DoneAt)
	})
	t.Run("unmarking a task as done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldTask := &Task{Done: true}
		newTask := &Task{Done: false}
		updateDone(oldTask, newTask)
		assert.Equal(t, timeutil.TimeStamp(0), oldTask.DoneAt)
	})
}

func TestTask_ReadOne(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{ID: 1}
		err := task.ReadOne()
		assert.NoError(t, err)
		assert.Equal(t, "task #1", task.Text)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		task := &Task{ID: 99999}
		err := task.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}
