// Vikunja is a todo-list application to facilitate your life.
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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestTaskComment_Create(t *testing.T) {
	u := &user.User{ID: 1}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{
			Comment: "test",
			TaskID:  1,
		}
		err := tc.Create(u)
		assert.NoError(t, err)
		assert.Equal(t, "test", tc.Comment)
		assert.Equal(t, int64(1), tc.Author.ID)
		db.AssertExists(t, "task_comments", map[string]interface{}{
			"id":        tc.ID,
			"author_id": u.ID,
			"comment":   "test",
			"task_id":   1,
		}, false)
	})
	t.Run("nonexisting task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{
			Comment: "test",
			TaskID:  99999,
		}
		err := tc.Create(u)
		assert.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}

func TestTaskComment_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{ID: 1}
		err := tc.Delete()
		assert.NoError(t, err)
		db.AssertMissing(t, "task_comments", map[string]interface{}{
			"id": 1,
		})
	})
	t.Run("nonexisting comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{ID: 9999}
		err := tc.Delete()
		assert.Error(t, err)
		assert.True(t, IsErrTaskCommentDoesNotExist(err))
	})
}

func TestTaskComment_Update(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{
			ID:      1,
			Comment: "testing",
		}
		err := tc.Update()
		assert.NoError(t, err)
		db.AssertExists(t, "task_comments", map[string]interface{}{
			"id":      1,
			"comment": "testing",
		}, false)
	})
	t.Run("nonexisting comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{
			ID: 9999,
		}
		err := tc.Update()
		assert.Error(t, err)
		assert.True(t, IsErrTaskCommentDoesNotExist(err))
	})
}

func TestTaskComment_ReadOne(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{ID: 1}
		err := tc.ReadOne()
		assert.NoError(t, err)
		assert.Equal(t, "Lorem Ipsum Dolor Sit Amet", tc.Comment)
		assert.NotEmpty(t, tc.Author.ID)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{ID: 9999}
		err := tc.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrTaskCommentDoesNotExist(err))
	})
}

func TestTaskComment_ReadAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{TaskID: 1}
		u := &user.User{ID: 1}
		result, resultCount, total, err := tc.ReadAll(u, "", 0, -1)
		resultComment := result.([]*TaskComment)
		assert.NoError(t, err)
		assert.Equal(t, 1, resultCount)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, int64(1), resultComment[0].ID)
		assert.Equal(t, "Lorem Ipsum Dolor Sit Amet", resultComment[0].Comment)
		assert.NotEmpty(t, resultComment[0].Author.ID)
	})
	t.Run("no access to task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		tc := &TaskComment{TaskID: 14}
		u := &user.User{ID: 1}
		_, _, _, err := tc.ReadAll(u, "", 0, -1)
		assert.Error(t, err)
		assert.True(t, IsErrGenericForbidden(err))
	})
}
