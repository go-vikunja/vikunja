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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestAPITokenService_Create(t *testing.T) {

	t.Run("create token successfully", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"v1_tasks": []string{"read_one", "update"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.NoError(t, err)
		assert.NotZero(t, token.ID)
		assert.NotEmpty(t, token.Token)
		assert.True(t, len(token.Token) > 8)
		assert.NotEmpty(t, token.TokenSalt)
		assert.NotEmpty(t, token.TokenHash)
		assert.NotEmpty(t, token.TokenLastEight)
		assert.Equal(t, int64(1), token.OwnerID)
		assert.Contains(t, token.Token, models.APITokenPrefix)
	})

	t.Run("create token with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		token := &models.APIToken{
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"v1_tasks": []string{"read_one", "update"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, nil)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("create token with invalid permissions", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"invalid_group": []string{"invalid_permission"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.Error(t, err)
	})

	t.Run("token ID is reset on create", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			ID:    999, // Should be reset
			Title: "Test Token",
			APIPermissions: models.APIPermissions{
				"v1_tasks": []string{"read_one", "update"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.NoError(t, err)
		assert.NotEqual(t, int64(999), token.ID)
	})
}

func TestAPITokenService_GetByID(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ats := NewAPITokenService(testEngine)

	t.Run("Success", func(t *testing.T) {
		token, err := ats.GetByID(s, 1)
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, "test token 1", token.Title)
	})

	t.Run("NotFound", func(t *testing.T) {
		token, err := ats.GetByID(s, 9999)
		assert.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
		assert.Nil(t, token)
	})
}

func TestAPITokenService_Get(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("get own token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token, err := service.Get(s, 1, u)
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, "test token 1", token.Title)
		assert.Equal(t, int64(1), token.OwnerID)
	})

	t.Run("get token of another user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		_, err := service.Get(s, 1, u) // Token 1 belongs to user 1
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("get nonexistent token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		_, err := service.Get(s, 999, u)
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("get with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, err := service.Get(s, 1, nil)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})
}

func TestAPITokenService_GetAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("get all tokens for user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		tokens, count, total, err := service.GetAll(s, u, "", 1, 50)
		require.NoError(t, err)
		assert.Len(t, tokens, 2)
		assert.Equal(t, 2, count)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(1), tokens[0].ID)
		assert.Equal(t, int64(2), tokens[1].ID)
	})

	t.Run("user only sees their own tokens", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		tokens, count, total, err := service.GetAll(s, u, "", 1, 50)
		require.NoError(t, err)
		assert.Len(t, tokens, 1) // User 2 has token 3
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, int64(3), tokens[0].ID) // Should be token 3
	})

	t.Run("search tokens by title", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		tokens, count, total, err := service.GetAll(s, u, "test token 1", 1, 50)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "test token 1", tokens[0].Title)
	})

	t.Run("pagination", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		// Get first page with 1 item per page
		tokens, count, total, err := service.GetAll(s, u, "", 1, 1)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(1), tokens[0].ID)

		// Get second page
		tokens, count, total, err = service.GetAll(s, u, "", 2, 1)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(2), tokens[0].ID)
	})

	t.Run("get all with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, _, _, err := service.GetAll(s, nil, "", 1, 50)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})
}

func TestAPITokenService_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("delete own token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		err := service.Delete(s, 1, u)
		require.NoError(t, err)

		// Verify token is deleted
		_, err = service.Get(s, 1, u)
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("delete token of another user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		err := service.Delete(s, 1, u) // Token 1 belongs to user 1
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("delete nonexistent token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		err := service.Delete(s, 999, u)
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenDoesNotExist(err))
	})

	t.Run("delete with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		err := service.Delete(s, 1, nil)
		require.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})
}

func TestAPITokenService_GetTokenFromTokenString(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("valid token string", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		token, err := service.GetTokenFromTokenString(s, "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e")
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, "test token 1", token.Title)
	})

	t.Run("invalid token string", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, err := service.GetTokenFromTokenString(s, "tk_invalidtoken")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenInvalid(err))
	})

	t.Run("token string with wrong hash", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		// Use correct last 8 chars but wrong prefix
		_, err := service.GetTokenFromTokenString(s, "tk_wronghash0000000000000075f29d2e")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenInvalid(err))
	})
}

func TestAPITokenService_ValidateToken(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("validate valid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		token, u, err := service.ValidateToken(s, "tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e", "/api/v1/tasks", "GET")
		require.NoError(t, err)
		assert.Equal(t, int64(1), token.ID)
		assert.Equal(t, int64(1), u.ID)
		assert.Equal(t, "user1", u.Username)
	})

	t.Run("validate expired token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		// Token 2 is expired (see fixtures)
		_, _, err := service.ValidateToken(s, "tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8", "/api/v1/tasks", "GET")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenExpired(err))
	})

	t.Run("validate invalid token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		_, _, err := service.ValidateToken(s, "tk_invalidtoken", "/api/v1/tasks", "GET")
		require.Error(t, err)
		assert.True(t, models.IsErrAPITokenInvalid(err))
	})
}

func TestAPITokenService_CanDelete(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("can delete own token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		can, err := service.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("cannot delete token of another user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 2}

		can, err := service.CanDelete(s, 1, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("cannot delete nonexistent token", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		can, err := service.CanDelete(s, 999, u)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("cannot delete with nil user", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)

		can, err := service.CanDelete(s, 1, nil)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

func TestAPITokenService_HashToken(t *testing.T) {
	service := NewAPITokenService(testEngine)

	t.Run("consistent hashing", func(t *testing.T) {
		token := "tk_testtoken123"
		salt := "testsalt"

		hash1 := service.hashToken(token, salt)
		hash2 := service.hashToken(token, salt)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})

	t.Run("different salts produce different hashes", func(t *testing.T) {
		token := "tk_testtoken123"
		salt1 := "salt1"
		salt2 := "salt2"

		hash1 := service.hashToken(token, salt1)
		hash2 := service.hashToken(token, salt2)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		token1 := "tk_testtoken1"
		token2 := "tk_testtoken2"
		salt := "testsalt"

		hash1 := service.hashToken(token1, salt)
		hash2 := service.hashToken(token2, salt)

		assert.NotEqual(t, hash1, hash2)
	})
}

// T006: Test helper function to create an API token with specific permissions
// This simplifies test setup by handling the full token creation flow
func createTokenWithPermissions(t *testing.T, s *xorm.Session, permissions models.APIPermissions) *models.APIToken {
	t.Helper()

	service := NewAPITokenService(testEngine)
	u := &user.User{ID: 1}

	token := &models.APIToken{
		Title:          "Test Token",
		APIPermissions: permissions,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	err := service.Create(s, token, u)
	require.NoError(t, err, "Failed to create test token")
	require.NotEmpty(t, token.Token, "Token string should not be empty")

	return token
}

// T007: Test helper function to create a mock Echo context
// This is useful for testing route handlers that require a context
func createMockContext(method, path string, token *models.APIToken) echo.Context {
	// Note: For US1 tests, we may not need full Echo context mocking
	// The registerTestAPIRoutes + models.CanDoAPIRoute pattern is sufficient
	// This helper is provided for future extensibility if needed
	// Implementation would require Echo test framework setup
	return nil // Placeholder - implement if needed for handler testing
}

// Helper function to check if token can access a route
// This wraps the models.CanDoAPIRoute function with a mock context
func canTokenAccessRoute(token *models.APIToken, method, path string) bool {
	e := echo.New()
	req := &http.Request{
		Method: method,
		URL:    &url.URL{Path: path},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// CRITICAL: c.SetPath must use the route pattern, NOT the actual request path
	// Echo's c.Path() returns the registered route pattern (with :param placeholders)
	// while c.Request().URL.Path returns the actual path with values
	c.SetPath(path)

	return models.CanDoAPIRoute(c, token)
}

// T008: Test that v1_tasks route group has all CRUD permissions registered
func TestAPITokenPermissionRegistration(t *testing.T) {
	// Get the registered routes
	routes := models.GetAPITokenRoutes()

	// Test v1 routes exist
	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	// Test tasks routes exist
	taskRoutes, hasTasks := v1Routes["tasks"]
	require.True(t, hasTasks, "Should have tasks routes registered in v1")

	// THESE SHOULD FAIL before fix if routes not properly registered:
	assert.NotNil(t, taskRoutes["create"], "Should have create permission for v1_tasks")
	assert.NotNil(t, taskRoutes["update"], "Should have update permission for v1_tasks")
	assert.NotNil(t, taskRoutes["delete"], "Should have delete permission for v1_tasks")
	assert.NotNil(t, taskRoutes["read_one"], "Should have read_one permission for v1_tasks")

	// Verify the route details are correct
	if taskRoutes["create"] != nil {
		assert.Equal(t, "PUT", taskRoutes["create"].Method)
		assert.Contains(t, taskRoutes["create"].Path, "/projects/:project/tasks")
	}

	if taskRoutes["read_one"] != nil {
		assert.Equal(t, "GET", taskRoutes["read_one"].Method)
		assert.Contains(t, taskRoutes["read_one"].Path, "/tasks/:taskid")
	}

	if taskRoutes["update"] != nil {
		assert.Equal(t, "POST", taskRoutes["update"].Method)
		assert.Contains(t, taskRoutes["update"].Path, "/tasks/:taskid")
	}

	if taskRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", taskRoutes["delete"].Method)
		assert.Contains(t, taskRoutes["delete"].Path, "/tasks/:taskid")
	}
}

// T009: Test that API token with create permission can create tasks
func TestAPITokenCanCreateTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create token with v1_tasks create permission
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks": []string{"create"},
	})

	// Verify token can access the create route (use route pattern with :param)
	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/tasks")
	assert.True(t, canAccess, "Token with create permission should be able to create tasks")

	// Verify token cannot access other routes
	cannotUpdate := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:taskid")
	assert.False(t, cannotUpdate, "Token with only create permission should not be able to update tasks")
}

// T010: Test that API token with update permission can update tasks
func TestAPITokenCanUpdateTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create token with v1_tasks update permission
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks": []string{"update"},
	})

	// Verify token can access the update route (use route pattern with :param)
	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:taskid")
	assert.True(t, canAccess, "Token with update permission should be able to update tasks")

	// Verify token cannot access other routes
	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/tasks")
	assert.False(t, cannotCreate, "Token with only update permission should not be able to create tasks")
}

// T011: Test that API token with delete permission can delete tasks
func TestAPITokenCanDeleteTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create token with v1_tasks delete permission
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks": []string{"delete"},
	})

	// Verify token can access the delete route (use route pattern with :param)
	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:taskid")
	assert.True(t, canAccess, "Token with delete permission should be able to delete tasks")

	// Verify token cannot access other routes
	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/tasks")
	assert.False(t, cannotCreate, "Token with only delete permission should not be able to create tasks")
}

// T012: Test that API token without permission gets denied
func TestAPITokenDeniedWithoutPermission(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create token with only read_one permission (no write permissions)
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks": []string{"read_one"},
	})

	// Verify token can access read route (use route pattern with :param)
	canRead := canTokenAccessRoute(token, "GET", "/api/v1/tasks/:taskid")
	assert.True(t, canRead, "Token with read_one permission should be able to read tasks")

	// Verify token is denied for write operations
	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/tasks")
	assert.False(t, cannotCreate, "Token without create permission should be denied")

	cannotUpdate := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:taskid")
	assert.False(t, cannotUpdate, "Token without update permission should be denied")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:taskid")
	assert.False(t, cannotDelete, "Token without delete permission should be denied")
}

// T036: Test that v2_tasks routes are registered with correct permission scopes
// This test verifies that v2 routes can be used in API token permissions
func TestV2RouteRegistration(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	service := NewAPITokenService(testEngine)
	u := &user.User{ID: 1}

	// Try to create a token with v2_tasks permissions
	// If v2 routes are not registered, this should fail validation
	token := &models.APIToken{
		Title: "Test V2 Token",
		APIPermissions: models.APIPermissions{
			"v2_tasks": []string{"read_all"},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// This will fail if v2_tasks routes are not properly registered
	err := service.Create(s, token, u)
	require.NoError(t, err, "Should be able to create token with v2_tasks permissions if routes are registered")

	// Verify the token was created with v2 permissions
	assert.Contains(t, token.APIPermissions, "v2_tasks")
	assert.Equal(t, []string{"read_all"}, token.APIPermissions["v2_tasks"])
}
