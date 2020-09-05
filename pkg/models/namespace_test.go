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
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNamespace_Create(t *testing.T) {

	// Dummy namespace
	dummynamespace := Namespace{
		Title:       "Test",
		Description: "Lorem Ipsum",
	}

	user1 := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		err := dummynamespace.Create(user1)
		assert.NoError(t, err)
	})
	t.Run("no title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n2 := Namespace{}
		err := n2.Create(user1)
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceNameCannotBeEmpty(err))
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		nUser := &user.User{ID: 9482385}
		dnsp2 := dummynamespace
		err := dnsp2.Create(nUser)
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
}

func TestNamespace_ReadOne(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		n := &Namespace{ID: 1}
		db.LoadAndAssertFixtures(t)
		err := n.ReadOne()
		assert.NoError(t, err)
		assert.Equal(t, n.Title, "testnamespace")
	})
	t.Run("nonexistant", func(t *testing.T) {
		n := &Namespace{ID: 99999}
		db.LoadAndAssertFixtures(t)
		err := n.ReadOne()
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceDoesNotExist(err))
	})
}

func TestNamespace_Update(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n := &Namespace{
			ID:    1,
			Title: "Lorem Ipsum",
		}
		err := n.Update()
		assert.NoError(t, err)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n := &Namespace{
			ID:    99999,
			Title: "Lorem Ipsum",
		}
		err := n.Update()
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceDoesNotExist(err))
	})
	t.Run("nonexisting owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n := &Namespace{
			ID:    1,
			Title: "Lorem Ipsum",
			Owner: &user.User{ID: 99999},
		}
		err := n.Update()
		assert.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
	t.Run("no title", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n := &Namespace{
			ID: 1,
		}
		err := n.Update()
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceNameCannotBeEmpty(err))
	})
}

func TestNamespace_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n := &Namespace{
			ID: 1,
		}
		err := n.Delete()
		assert.NoError(t, err)
	})
	t.Run("nonexisting", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		n := &Namespace{
			ID: 9999,
		}
		err := n.Delete()
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceDoesNotExist(err))
	})
}

func TestNamespace_ReadAll(t *testing.T) {
	user1 := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		n := &Namespace{}
		nn, _, _, err := n.ReadAll(user1, "", 1, -1)
		namespaces := nn.([]*NamespaceWithLists)
		assert.NoError(t, err)
		assert.NotNil(t, namespaces)
		assert.Len(t, namespaces, 10)                // Total of 10 including shared & favorites
		assert.Equal(t, int64(-2), namespaces[0].ID) // The first one should be the one with favorites
		assert.Equal(t, int64(-1), namespaces[1].ID) // The second one should be the one with the shared namespaces
		// Ensure every list and namespace are not archived
		for _, namespace := range namespaces {
			assert.False(t, namespace.IsArchived)
			for _, list := range namespace.Lists {
				assert.False(t, list.IsArchived)
			}
		}
	})
	t.Run("archived", func(t *testing.T) {
		n := &Namespace{
			IsArchived: true,
		}
		nn, _, _, err := n.ReadAll(user1, "", 1, -1)
		namespaces := nn.([]*NamespaceWithLists)
		assert.NoError(t, err)
		assert.NotNil(t, namespaces)
		assert.Len(t, namespaces, 11)                // Total of 11 including shared & favorites, one is archived
		assert.Equal(t, int64(-2), namespaces[0].ID) // The first one should be the one with favorites
		assert.Equal(t, int64(-1), namespaces[1].ID) // The second one should be the one with the shared namespaces
	})
}
