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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectScopeEnforcement(t *testing.T) {
	t.Run("scoped auth can read project in scope", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1}}

		p := &Project{ID: 1}
		can, _, err := p.CanRead(s, scoped)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("scoped auth cannot read project outside scope", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		// User 1 owns project 1, but token is scoped to project 6 (which user 1 also has access to)
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{6}}

		p := &Project{ID: 1}
		can, _, err := p.CanRead(s, scoped)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("scoped auth cannot write to project outside scope", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{6}}

		p := &Project{ID: 1}
		can, err := p.CanWrite(s, scoped)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("scoped auth can write to project in scope", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1}}

		p := &Project{ID: 1}
		can, err := p.CanWrite(s, scoped)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("scoped auth cannot delete project outside scope", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{6}}

		p := &Project{ID: 1}
		can, err := p.IsAdmin(s, scoped)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("unscoped auth can read any project it has access to", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}

		p := &Project{ID: 1}
		can, _, err := p.CanRead(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("scoped auth cannot create top-level projects", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1}}

		p := &Project{Title: "New Top Level"}
		can, err := p.CanCreate(s, scoped)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("scoped auth can create sub-project under scoped project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1}}

		p := &Project{Title: "Sub Project", ParentProjectID: 1}
		can, err := p.CanCreate(s, scoped)
		require.NoError(t, err)
		assert.True(t, can)
	})
}
