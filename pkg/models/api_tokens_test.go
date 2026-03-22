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

func TestAPIToken_ReadAll(t *testing.T) {
	u := &user.User{ID: 1}
	token := &APIToken{}
	s := db.NewSession()
	defer s.Close()
	db.LoadAndAssertFixtures(t)

	// Checking if the user only sees their own tokens

	result, count, total, err := token.ReadAll(s, u, "", 1, 50)
	require.NoError(t, err)
	tokens, is := result.([]*APIToken)
	assert.Truef(t, is, "tokens are not of type []*APIToken")
	assert.Len(t, tokens, 4)
	assert.Len(t, tokens, count)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, int64(1), tokens[0].ID)
	assert.Equal(t, int64(2), tokens[1].ID)
	assert.Equal(t, int64(4), tokens[2].ID)
	assert.Equal(t, int64(5), tokens[3].ID)
}

func TestAPIToken_CanDelete(t *testing.T) {
	t.Run("own token", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{ID: 1}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := token.CanDelete(s, u)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("noneixsting token", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{ID: 999}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := token.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("token of another user", func(t *testing.T) {
		u := &user.User{ID: 2}
		token := &APIToken{ID: 1}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		can, err := token.CanDelete(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestAPIToken_Create(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.NoError(t, err)
	})
	t.Run("with project scope", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{
			ProjectID: 1,
		}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ProjectID)
		assert.False(t, token.IncludeSubProjects)
	})
	t.Run("with project scope and sub-projects", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{
			ProjectID:          1,
			IncludeSubProjects: true,
		}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ProjectID)
		assert.True(t, token.IncludeSubProjects)
	})
	t.Run("with nonexistent project", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{
			ProjectID: 999999,
		}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
	})
	t.Run("with project user has no access to", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{
			ProjectID: 2, // owned by user 3
		}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
	})
	t.Run("include sub-projects without project ID is ignored", func(t *testing.T) {
		u := &user.User{ID: 1}
		token := &APIToken{
			IncludeSubProjects: true,
		}
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		err := token.Create(s, u)
		require.NoError(t, err)
		assert.False(t, token.IncludeSubProjects)
	})
}

func TestGetProjectIDsForToken(t *testing.T) {
	t.Run("no project scope", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		ids, err := GetProjectIDsForToken(s, &APIToken{ProjectID: 0})
		require.NoError(t, err)
		assert.Nil(t, ids)
	})
	t.Run("single project", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		ids, err := GetProjectIDsForToken(s, &APIToken{ProjectID: 1, IncludeSubProjects: false})
		require.NoError(t, err)
		assert.Equal(t, []int64{1}, ids)
	})
	t.Run("with sub-projects", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		// Project 22 has child project 21
		ids, err := GetProjectIDsForToken(s, &APIToken{ProjectID: 22, IncludeSubProjects: true})
		require.NoError(t, err)
		assert.Contains(t, ids, int64(22))
		assert.Contains(t, ids, int64(21))
	})
}

func TestProjectScopeContains(t *testing.T) {
	t.Run("nil scope allows all", func(t *testing.T) {
		assert.True(t, ProjectScopeContains(nil, 1))
		assert.True(t, ProjectScopeContains(nil, 999))
	})
	t.Run("scope contains project", func(t *testing.T) {
		assert.True(t, ProjectScopeContains([]int64{1, 2, 3}, 2))
	})
	t.Run("scope does not contain project", func(t *testing.T) {
		assert.False(t, ProjectScopeContains([]int64{1, 2, 3}, 5))
	})
}

func TestProjectScopedAuth(t *testing.T) {
	t.Run("GetID delegates to inner auth", func(t *testing.T) {
		u := &user.User{ID: 42}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1}}
		assert.Equal(t, int64(42), scoped.GetID())
	})
	t.Run("GetProjectScope returns scope", func(t *testing.T) {
		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1, 2}}
		scope := GetProjectScope(scoped)
		assert.Equal(t, []int64{1, 2}, scope)
	})
	t.Run("GetProjectScope returns nil for regular auth", func(t *testing.T) {
		u := &user.User{ID: 1}
		scope := GetProjectScope(u)
		assert.Nil(t, scope)
	})
	t.Run("UnwrapAuth returns inner auth", func(t *testing.T) {
		u := &user.User{ID: 1}
		scoped := &ProjectScopedAuth{Auth: u, ProjectIDs: []int64{1}}
		assert.Equal(t, u, scoped.UnwrapAuth())
	})
}

func TestAPIToken_GetTokenFromTokenString(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		token, err := GetTokenFromTokenString(s, "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e") // Token 1

		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
	})
	t.Run("invalid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()
		db.LoadAndAssertFixtures(t)

		_, err := GetTokenFromTokenString(s, "tk_loremipsum")

		require.Error(t, err)
		assert.True(t, IsErrAPITokenInvalid(err))
	})
}
