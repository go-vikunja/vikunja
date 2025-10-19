// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package services

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestWebhookService_CanRead(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ws := NewWebhookService(testEngine)

	t.Run("user with read access can read webhooks", func(t *testing.T) {
		u := &user.User{ID: 1} // Owner of project 1
		can, maxRight, err := ws.CanRead(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
		assert.Greater(t, maxRight, 0)
	})

	t.Run("user without access cannot read webhooks", func(t *testing.T) {
		u := &user.User{ID: 13} // No access to project 1
		can, _, err := ws.CanRead(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

func TestWebhookService_CanCreate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ws := NewWebhookService(testEngine)

	t.Run("project admin can create webhook", func(t *testing.T) {
		u := &user.User{ID: 1} // Owner of project 1
		can, err := ws.CanCreate(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with write permission can create webhook", func(t *testing.T) {
		u := &user.User{ID: 6} // Has write access to project 7
		can, err := ws.CanCreate(s, 7, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with read-only permission cannot create webhook", func(t *testing.T) {
		u := &user.User{ID: 3} // Has read-only access to project 1
		can, err := ws.CanCreate(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("link share cannot create webhook", func(t *testing.T) {
		ls := &models.LinkSharing{ID: 1}
		can, err := ws.CanCreate(s, 1, ls)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

func TestWebhookService_CanUpdate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ws := NewWebhookService(testEngine)

	t.Run("project admin can update webhook", func(t *testing.T) {
		u := &user.User{ID: 1}
		can, err := ws.CanUpdate(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with write permission can update webhook", func(t *testing.T) {
		u := &user.User{ID: 6}
		can, err := ws.CanUpdate(s, 7, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with read-only permission cannot update webhook", func(t *testing.T) {
		u := &user.User{ID: 3}
		can, err := ws.CanUpdate(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}

func TestWebhookService_CanDelete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ws := NewWebhookService(testEngine)

	t.Run("project admin can delete webhook", func(t *testing.T) {
		u := &user.User{ID: 1}
		can, err := ws.CanDelete(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with write permission can delete webhook", func(t *testing.T) {
		u := &user.User{ID: 6}
		can, err := ws.CanDelete(s, 7, u)
		assert.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("user with read-only permission cannot delete webhook", func(t *testing.T) {
		u := &user.User{ID: 3}
		can, err := ws.CanDelete(s, 1, u)
		assert.NoError(t, err)
		assert.False(t, can)
	})
}
