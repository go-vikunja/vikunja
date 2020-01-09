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
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestList_CreateOrUpdate(t *testing.T) {
	user := &User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	t.Run("create", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			initFixtures(t)
			list := List{
				Title:       "test",
				Description: "Lorem Ipsum",
				NamespaceID: 1,
			}
			err := list.Create(user)
			assert.NoError(t, err)
		})
		t.Run("nonexistant namespace", func(t *testing.T) {
			initFixtures(t)
			list := List{
				Title:       "test",
				Description: "Lorem Ipsum",
				NamespaceID: 999999,
			}

			err := list.Create(user)
			assert.Error(t, err)
			assert.True(t, IsErrNamespaceDoesNotExist(err))
		})
		t.Run("nonexistant owner", func(t *testing.T) {
			initFixtures(t)
			user := &User{ID: 9482385}
			list := List{
				Title:       "test",
				Description: "Lorem Ipsum",
				NamespaceID: 1,
			}
			err := list.Create(user)
			assert.Error(t, err)
			assert.True(t, IsErrUserDoesNotExist(err))
		})
		t.Run("existing identifier", func(t *testing.T) {
			initFixtures(t)
			list := List{
				Title:       "test",
				Description: "Lorem Ipsum",
				Identifier:  "test1",
				NamespaceID: 1,
			}

			err := list.Create(user)
			assert.Error(t, err)
			assert.True(t, IsErrListIdentifierIsNotUnique(err))
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			initFixtures(t)
			list := List{
				ID:          1,
				Title:       "test",
				Description: "Lorem Ipsum",
				NamespaceID: 1,
			}
			list.Description = "Lorem Ipsum dolor sit amet."
			err := list.Update()
			assert.NoError(t, err)

		})
		t.Run("nonexistant", func(t *testing.T) {
			initFixtures(t)
			list := List{
				ID:    99999999,
				Title: "test",
			}
			err := list.Update()
			assert.Error(t, err)
			assert.True(t, IsErrListDoesNotExist(err))

		})
		t.Run("existing identifier", func(t *testing.T) {
			initFixtures(t)
			list := List{
				Title:       "test",
				Description: "Lorem Ipsum",
				Identifier:  "test1",
				NamespaceID: 1,
			}

			err := list.Create(user)
			assert.Error(t, err)
			assert.True(t, IsErrListIdentifierIsNotUnique(err))
		})
	})
}

func TestList_Delete(t *testing.T) {
	initFixtures(t)
	list := List{
		ID: 1,
	}
	err := list.Delete()
	assert.NoError(t, err)
}

func TestList_ReadAll(t *testing.T) {
	t.Run("all in namespace", func(t *testing.T) {
		initFixtures(t)
		// Get all lists for our namespace
		lists, err := GetListsByNamespaceID(1, &User{})
		assert.NoError(t, err)
		assert.Equal(t, len(lists), 2)
	})
	t.Run("all lists for user", func(t *testing.T) {
		u := &User{ID: 1}
		list := List{}
		lists3, _, _, err := list.ReadAll(u, "", 1, 50)

		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(lists3).Kind(), reflect.Slice)
		s := reflect.ValueOf(lists3)
		assert.Equal(t, 16, s.Len())
	})
	t.Run("lists for nonexistant user", func(t *testing.T) {
		user := &User{ID: 999999}
		list := List{}
		_, _, _, err := list.ReadAll(user, "", 1, 50)
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
}
