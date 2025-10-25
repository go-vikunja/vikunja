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
		assert.Greater(t, len(token.Token), 8)
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
func createMockContext(_ string, path string, token *models.APIToken) echo.Context {
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

// T066 Tests: Label-Task Relations Permission Registration
func TestLabelTaskPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	labelTaskRoutes, hasLabelTasks := v1Routes["tasks_labels"]
	require.True(t, hasLabelTasks, "Should have tasks_labels routes registered in v1")

	// Verify all expected permissions exist
	assert.NotNil(t, labelTaskRoutes["read_all"], "Should have read_all permission for v1_tasks_labels")
	assert.NotNil(t, labelTaskRoutes["create"], "Should have create permission for v1_tasks_labels")
	assert.NotNil(t, labelTaskRoutes["delete"], "Should have delete permission for v1_tasks_labels")
	assert.NotNil(t, labelTaskRoutes["update"], "Should have update permission for v1_tasks_labels")

	bulkRoutes, hasBulk := v1Routes["tasks_labels_bulk"]
	require.True(t, hasBulk, "Should have tasks_labels_bulk group registered in v1")
	assert.NotNil(t, bulkRoutes["update"], "Should have update permission registered for bulk operations")

	// Verify the route details are correct
	if labelTaskRoutes["read_all"] != nil {
		assert.Equal(t, "GET", labelTaskRoutes["read_all"].Method)
		assert.Contains(t, labelTaskRoutes["read_all"].Path, "/tasks/:projecttask/labels")
	}

	if labelTaskRoutes["create"] != nil {
		assert.Equal(t, "PUT", labelTaskRoutes["create"].Method)
		assert.Contains(t, labelTaskRoutes["create"].Path, "/tasks/:projecttask/labels")
	}

	if labelTaskRoutes["update"] != nil {
		assert.Equal(t, "POST", labelTaskRoutes["update"].Method)
		assert.Contains(t, labelTaskRoutes["update"].Path, "/tasks/:projecttask/labels/bulk")
	}

	if labelTaskRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", labelTaskRoutes["delete"].Method)
		assert.Contains(t, labelTaskRoutes["delete"].Path, "/tasks/:projecttask/labels/:label")
	}

	if bulkRoutes["update"] != nil {
		assert.Equal(t, "POST", bulkRoutes["update"].Method)
		assert.Contains(t, bulkRoutes["update"].Path, "/tasks/:projecttask/labels/bulk")
	}
}

func TestAPITokenCanAddLabelToTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_labels": []string{"create"},
	})

	// Verify token can access the add label route
	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/labels")
	assert.True(t, canAccess, "Token with create permission should be able to add labels to tasks")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:projecttask/labels/:label")
	assert.False(t, cannotDelete, "Token with only create permission should not be able to delete labels from tasks")
}

func TestAPITokenCanRemoveLabelFromTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_labels": []string{"delete"},
	})

	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:projecttask/labels/:label")
	assert.True(t, canAccess, "Token with delete permission should be able to remove labels from tasks")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/labels")
	assert.False(t, cannotCreate, "Token with only delete permission should not be able to add labels to tasks")
}

func TestAPITokenCanGetTaskLabels(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_labels": []string{"read_all"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/tasks/:projecttask/labels")
	assert.True(t, canAccess, "Token with read_all permission should be able to get task labels")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/labels")
	assert.False(t, cannotCreate, "Token with only read_all permission should not be able to add labels")
}

func TestAPITokenCanBulkUpdateTaskLabels(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_labels": []string{"update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:projecttask/labels/bulk")
	assert.True(t, canAccess, "Token with update permission should be able to bulk update task labels")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/labels")
	assert.False(t, cannotCreate, "Token with only update permission should not be able to add single labels")
}

// T067 Tests: Task Assignee Permission Registration
func TestTaskAssigneePermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	assigneeRoutes, hasAssignees := v1Routes["tasks_assignees"]
	require.True(t, hasAssignees, "Should have tasks_assignees routes registered in v1")

	// Verify all expected permissions exist
	assert.NotNil(t, assigneeRoutes["read_all"], "Should have read_all permission for v1_tasks_assignees")
	assert.NotNil(t, assigneeRoutes["create"], "Should have create permission for v1_tasks_assignees")
	assert.NotNil(t, assigneeRoutes["delete"], "Should have delete permission for v1_tasks_assignees")
	assert.NotNil(t, assigneeRoutes["update"], "Should have update permission for v1_tasks_assignees")

	bulkRoutes, hasBulk := v1Routes["tasks_assignees_bulk"]
	require.True(t, hasBulk, "Should have tasks_assignees_bulk group registered in v1")
	assert.NotNil(t, bulkRoutes["update"], "Should have update permission registered for bulk operations")

	// Verify the route details are correct
	if assigneeRoutes["read_all"] != nil {
		assert.Equal(t, "GET", assigneeRoutes["read_all"].Method)
		assert.Contains(t, assigneeRoutes["read_all"].Path, "/tasks/:projecttask/assignees")
	}

	if assigneeRoutes["create"] != nil {
		assert.Equal(t, "PUT", assigneeRoutes["create"].Method)
		assert.Contains(t, assigneeRoutes["create"].Path, "/tasks/:projecttask/assignees")
	}

	if assigneeRoutes["update"] != nil {
		assert.Equal(t, "POST", assigneeRoutes["update"].Method)
		assert.Contains(t, assigneeRoutes["update"].Path, "/tasks/:projecttask/assignees/bulk")
	}

	if assigneeRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", assigneeRoutes["delete"].Method)
		assert.Contains(t, assigneeRoutes["delete"].Path, "/tasks/:projecttask/assignees/:user")
	}

	if bulkRoutes["update"] != nil {
		assert.Equal(t, "POST", bulkRoutes["update"].Method)
		assert.Contains(t, bulkRoutes["update"].Path, "/tasks/:projecttask/assignees/bulk")
	}
}

func TestAPITokenCanAddAssigneeToTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_assignees": []string{"create"},
	})

	// Verify token can access the add assignee route
	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/assignees")
	assert.True(t, canAccess, "Token with create permission should be able to add assignees to tasks")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:projecttask/assignees/:user")
	assert.False(t, cannotDelete, "Token with only create permission should not be able to remove assignees from tasks")
}

func TestAPITokenCanRemoveAssigneeFromTask(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_assignees": []string{"delete"},
	})

	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:projecttask/assignees/:user")
	assert.True(t, canAccess, "Token with delete permission should be able to remove assignees from tasks")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/assignees")
	assert.False(t, cannotCreate, "Token with only delete permission should not be able to add assignees to tasks")
}

func TestAPITokenCanGetTaskAssignees(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_assignees": []string{"read_all"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/tasks/:projecttask/assignees")
	assert.True(t, canAccess, "Token with read_all permission should be able to get task assignees")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/assignees")
	assert.False(t, cannotCreate, "Token with only read_all permission should not be able to add assignees")
}

func TestAPITokenCanBulkUpdateTaskAssignees(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_assignees": []string{"update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:projecttask/assignees/bulk")
	assert.True(t, canAccess, "Token with update permission should be able to bulk update task assignees")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:projecttask/assignees")
	assert.False(t, cannotCreate, "Token with only update permission should not be able to add single assignees")
}

// T068 Tests: Task Relation Permission Registration
func TestTaskRelationPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	relationRoutes, hasRelations := v1Routes["tasks_relations"]
	require.True(t, hasRelations, "Should have tasks_relations routes registered in v1")

	// Verify all expected permissions exist
	assert.NotNil(t, relationRoutes["create"], "Should have create permission for v1_tasks_relations")
	assert.NotNil(t, relationRoutes["delete"], "Should have delete permission for v1_tasks_relations")

	// Verify the route details are correct
	if relationRoutes["create"] != nil {
		assert.Equal(t, "PUT", relationRoutes["create"].Method)
		assert.Contains(t, relationRoutes["create"].Path, "/tasks/:task/relations")
	}

	if relationRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", relationRoutes["delete"].Method)
		assert.Contains(t, relationRoutes["delete"].Path, "/tasks/:task/relations/:relationKind/:otherTask")
	}
}

func TestAPITokenCanCreateTaskRelation(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_relations": []string{"create"},
	})

	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:task/relations")
	assert.True(t, canAccess, "Token with create permission should be able to create task relations")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:task/relations/:relationKind/:otherTask")
	assert.False(t, cannotDelete, "Token with only create permission should not be able to delete task relations")
}

func TestAPITokenCanDeleteTaskRelation(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_relations": []string{"delete"},
	})

	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/tasks/:task/relations/:relationKind/:otherTask")
	assert.True(t, canAccess, "Token with delete permission should be able to delete task relations")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/tasks/:task/relations")
	assert.False(t, cannotCreate, "Token with only delete permission should not be able to create task relations")
}

// T069 Tests: Task Position Permission Registration
func TestTaskPositionPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	positionRoutes, hasPosition := v1Routes["tasks_position"]
	require.True(t, hasPosition, "Should have tasks_position routes registered in v1")

	// Verify all expected permissions exist
	assert.NotNil(t, positionRoutes["update"], "Should have update permission for v1_tasks_position")

	// Verify the route details are correct
	if positionRoutes["update"] != nil {
		assert.Equal(t, "POST", positionRoutes["update"].Method)
		assert.Contains(t, positionRoutes["update"].Path, "/tasks/:task/position")
	}
}

func TestAPITokenCanUpdateTaskPosition(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks_position": []string{"update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:task/position")
	assert.True(t, canAccess, "Token with update permission should be able to update task position")
}

// T070 Tests: Bulk Task Permission Registration
func TestBulkTaskPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	bulkRoutes, hasBulk := v1Routes["tasks_bulk"]
	require.True(t, hasBulk, "Should have tasks_bulk routes registered in v1")

	// Verify all expected permissions exist (uses 'bulk_update' to avoid conflicts with single-task update)
	assert.NotNil(t, bulkRoutes["bulk_update"], "Should have bulk_update permission for v1_tasks_bulk")

	// Verify the route details are correct
	if bulkRoutes["bulk_update"] != nil {
		assert.Equal(t, "POST", bulkRoutes["bulk_update"].Method)
		assert.Contains(t, bulkRoutes["bulk_update"].Path, "/tasks/bulk")
	}
}

func TestAPITokenCanBulkUpdateTasks(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Bulk task route uses 'bulk_update' scope to avoid conflicting with single-task update
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks": []string{"bulk_update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/tasks/bulk")
	assert.True(t, canAccess, "Token with bulk_update permission should be able to bulk update tasks")
}

// T071 Tests: Kanban/Bucket Permission Registration
func TestKanbanPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	kanbanRoutes, hasKanban := v1Routes["projects_views_buckets"]
	require.True(t, hasKanban, "Should have projects_views_buckets routes registered in v1")

	// Verify bucket management permissions exist
	assert.NotNil(t, kanbanRoutes["read_all"], "Should have read_all permission for v1_projects_views_buckets")
	assert.NotNil(t, kanbanRoutes["create"], "Should have create permission for v1_projects_views_buckets")
	assert.NotNil(t, kanbanRoutes["update"], "Should have update permission for v1_projects_views_buckets")
	assert.NotNil(t, kanbanRoutes["delete"], "Should have delete permission for v1_projects_views_buckets")

	// Verify the route details are correct
	if kanbanRoutes["read_all"] != nil {
		assert.Equal(t, "GET", kanbanRoutes["read_all"].Method)
		assert.Contains(t, kanbanRoutes["read_all"].Path, "/projects/:project/views/:view/buckets")
	}

	if kanbanRoutes["create"] != nil {
		assert.Equal(t, "PUT", kanbanRoutes["create"].Method)
		assert.Contains(t, kanbanRoutes["create"].Path, "/projects/:project/views/:view/buckets")
	}

	if kanbanRoutes["update"] != nil {
		assert.Equal(t, "POST", kanbanRoutes["update"].Method)
		assert.Contains(t, kanbanRoutes["update"].Path, "/projects/:project/views/:view/buckets/:bucket")
	}

	if kanbanRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", kanbanRoutes["delete"].Method)
		assert.Contains(t, kanbanRoutes["delete"].Path, "/projects/:project/views/:view/buckets/:bucket")
	}

	// Check for move_task in the projects_views_buckets_tasks group (separate group due to /tasks suffix)
	tasksRoutes, hasTasks := v1Routes["projects_views_buckets_tasks"]
	require.True(t, hasTasks, "Should have projects_views_buckets_tasks routes registered in v1")
	assert.NotNil(t, tasksRoutes["move_task"], "Should have move_task permission for v1_projects_views_buckets_tasks")

	if tasksRoutes["move_task"] != nil {
		assert.Equal(t, "POST", tasksRoutes["move_task"].Method)
		assert.Contains(t, tasksRoutes["move_task"].Path, "/projects/:project/views/:view/buckets/:bucket/tasks")
	}
}

func TestAPITokenCanGetBuckets(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views_buckets": []string{"read_all"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/projects/:project/views/:view/buckets")
	assert.True(t, canAccess, "Token with read_all permission should be able to get buckets")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/views/:view/buckets")
	assert.False(t, cannotCreate, "Token with only read_all permission should not be able to create buckets")
}

func TestAPITokenCanCreateBucket(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views_buckets": []string{"create"},
	})

	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/views/:view/buckets")
	assert.True(t, canAccess, "Token with create permission should be able to create buckets")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/projects/:project/views/:view/buckets/:bucket")
	assert.False(t, cannotDelete, "Token with only create permission should not be able to delete buckets")
}

func TestAPITokenCanUpdateBucket(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views_buckets": []string{"update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/projects/:project/views/:view/buckets/:bucket")
	assert.True(t, canAccess, "Token with update permission should be able to update buckets")
}

func TestAPITokenCanDeleteBucket(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views_buckets": []string{"delete"},
	})

	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/projects/:project/views/:view/buckets/:bucket")
	assert.True(t, canAccess, "Token with delete permission should be able to delete buckets")
}

func TestAPITokenCanMoveTaskToBucket(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views_buckets": []string{"move_task"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/projects/:project/views/:view/buckets/:bucket/tasks")
	assert.True(t, canAccess, "Token with move_task permission should be able to move tasks to buckets")
}

// T072 Tests: Project View Permission Registration
func TestProjectViewPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	viewRoutes, hasViews := v1Routes["projects_views"]
	require.True(t, hasViews, "Should have projects_views routes registered in v1")

	// Verify all expected permissions exist
	assert.NotNil(t, viewRoutes["read_all"], "Should have read_all permission for v1_projects_views")
	assert.NotNil(t, viewRoutes["read_one"], "Should have read_one permission for v1_projects_views")
	assert.NotNil(t, viewRoutes["create"], "Should have create permission for v1_projects_views")
	assert.NotNil(t, viewRoutes["update"], "Should have update permission for v1_projects_views")
	assert.NotNil(t, viewRoutes["delete"], "Should have delete permission for v1_projects_views")

	// Verify the route details are correct
	if viewRoutes["read_all"] != nil {
		assert.Equal(t, "GET", viewRoutes["read_all"].Method)
		assert.Contains(t, viewRoutes["read_all"].Path, "/projects/:project/views")
		assert.NotContains(t, viewRoutes["read_all"].Path, "/projects/:project/views/:view")
	}

	if viewRoutes["read_one"] != nil {
		assert.Equal(t, "GET", viewRoutes["read_one"].Method)
		assert.Contains(t, viewRoutes["read_one"].Path, "/projects/:project/views/:view")
	}

	if viewRoutes["create"] != nil {
		assert.Equal(t, "PUT", viewRoutes["create"].Method)
		assert.Contains(t, viewRoutes["create"].Path, "/projects/:project/views")
	}

	if viewRoutes["update"] != nil {
		assert.Equal(t, "POST", viewRoutes["update"].Method)
		assert.Contains(t, viewRoutes["update"].Path, "/projects/:project/views/:view")
	}

	if viewRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", viewRoutes["delete"].Method)
		assert.Contains(t, viewRoutes["delete"].Path, "/projects/:project/views/:view")
	}
}

func TestAPITokenCanGetAllProjectViews(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views": []string{"read_all"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/projects/:project/views")
	assert.True(t, canAccess, "Token with read_all permission should be able to get all project views")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/views")
	assert.False(t, cannotCreate, "Token with only read_all permission should not be able to create project views")
}

func TestAPITokenCanGetOneProjectView(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views": []string{"read_one"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/projects/:project/views/:view")
	assert.True(t, canAccess, "Token with read_one permission should be able to get a single project view")
}

func TestAPITokenCanCreateProjectView(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views": []string{"create"},
	})

	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/views")
	assert.True(t, canAccess, "Token with create permission should be able to create project views")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/projects/:project/views/:view")
	assert.False(t, cannotDelete, "Token with only create permission should not be able to delete project views")
}

func TestAPITokenCanUpdateProjectView(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views": []string{"update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/projects/:project/views/:view")
	assert.True(t, canAccess, "Token with update permission should be able to update project views")
}

func TestAPITokenCanDeleteProjectView(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_views": []string{"delete"},
	})

	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/projects/:project/views/:view")
	assert.True(t, canAccess, "Token with delete permission should be able to delete project views")
}

// T073 Tests: Saved Filter Permission Registration
func TestSavedFilterPermissionRegistration(t *testing.T) {
	routes := models.GetAPITokenRoutes()

	v1Routes, hasV1 := routes["v1"]
	require.True(t, hasV1, "Should have v1 routes registered")

	filterRoutes, hasFilters := v1Routes["filters"]
	require.True(t, hasFilters, "Should have filters routes registered in v1")

	// Verify all expected permissions exist
	assert.NotNil(t, filterRoutes["read_all"], "Should have read_all permission for v1_filters")
	assert.NotNil(t, filterRoutes["read_one"], "Should have read_one permission for v1_filters")
	assert.NotNil(t, filterRoutes["create"], "Should have create permission for v1_filters")
	assert.NotNil(t, filterRoutes["update"], "Should have update permission for v1_filters")
	assert.NotNil(t, filterRoutes["delete"], "Should have delete permission for v1_filters")

	// Verify the route details are correct
	if filterRoutes["read_all"] != nil {
		assert.Equal(t, "GET", filterRoutes["read_all"].Method)
		assert.Equal(t, "/api/v1/filters", filterRoutes["read_all"].Path)
	}

	if filterRoutes["read_one"] != nil {
		assert.Equal(t, "GET", filterRoutes["read_one"].Method)
		assert.Contains(t, filterRoutes["read_one"].Path, "/filters/:filter")
	}

	if filterRoutes["create"] != nil {
		assert.Equal(t, "PUT", filterRoutes["create"].Method)
		assert.Equal(t, "/api/v1/filters", filterRoutes["create"].Path)
	}

	if filterRoutes["update"] != nil {
		assert.Equal(t, "POST", filterRoutes["update"].Method)
		assert.Contains(t, filterRoutes["update"].Path, "/filters/:filter")
	}

	if filterRoutes["delete"] != nil {
		assert.Equal(t, "DELETE", filterRoutes["delete"].Method)
		assert.Contains(t, filterRoutes["delete"].Path, "/filters/:filter")
	}
}

func TestAPITokenCanGetAllSavedFilters(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_filters": []string{"read_all"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/filters")
	assert.True(t, canAccess, "Token with read_all permission should be able to get all saved filters")

	cannotCreate := canTokenAccessRoute(token, "PUT", "/api/v1/filters")
	assert.False(t, cannotCreate, "Token with only read_all permission should not be able to create saved filters")
}

func TestAPITokenCanGetOneSavedFilter(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_filters": []string{"read_one"},
	})

	canAccess := canTokenAccessRoute(token, "GET", "/api/v1/filters/:filter")
	assert.True(t, canAccess, "Token with read_one permission should be able to get a single saved filter")
}

func TestAPITokenCanCreateSavedFilter(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_filters": []string{"create"},
	})

	canAccess := canTokenAccessRoute(token, "PUT", "/api/v1/filters")
	assert.True(t, canAccess, "Token with create permission should be able to create saved filters")

	cannotDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/filters/:filter")
	assert.False(t, cannotDelete, "Token with only create permission should not be able to delete saved filters")
}

func TestAPITokenCanUpdateSavedFilter(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_filters": []string{"update"},
	})

	canAccess := canTokenAccessRoute(token, "POST", "/api/v1/filters/:filter")
	assert.True(t, canAccess, "Token with update permission should be able to update saved filters")
}

func TestAPITokenCanDeleteSavedFilter(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_filters": []string{"delete"},
	})

	canAccess := canTokenAccessRoute(token, "DELETE", "/api/v1/filters/:filter")
	assert.True(t, canAccess, "Token with delete permission should be able to delete saved filters")
}

// ===== Token Level Tests =====

func TestAPITokenService_CreateWithTokenLevel(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("create standard token by default", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title: "Test Standard Token",
			APIPermissions: models.APIPermissions{
				"v1_tasks": []string{"read_one"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.NoError(t, err)
		assert.Equal(t, models.APITokenLevelStandard, token.TokenLevel, "Token should default to standard level")
	})

	t.Run("create admin token explicitly", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title:      "Test Admin Token",
			TokenLevel: models.APITokenLevelAdmin,
			APIPermissions: models.APIPermissions{
				"v1_projects_webhooks": []string{"read_all"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.NoError(t, err)
		assert.Equal(t, models.APITokenLevelAdmin, token.TokenLevel, "Token should be admin level")
	})

	t.Run("reject invalid token level", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		service := NewAPITokenService(testEngine)
		u := &user.User{ID: 1}

		token := &models.APIToken{
			Title:      "Invalid Level Token",
			TokenLevel: models.APITokenLevel("superadmin"), // Invalid level
			APIPermissions: models.APIPermissions{
				"v1_tasks": []string{"read_one"},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := service.Create(s, token, u)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token_level must be either 'standard' or 'admin'")
	})
}

func TestWebhookPermissionRegistration(t *testing.T) {
	registerTestAPIRoutes()

	routes := models.GetAPITokenRoutes()
	assert.NotNil(t, routes)
	assert.Contains(t, routes, "v1")

	webhooksGroup := routes["v1"]["projects_webhooks"]
	assert.NotNil(t, webhooksGroup, "Webhooks permission group should be registered")

	// Check all webhook permissions are registered
	assert.NotNil(t, webhooksGroup["read_all"], "read_all permission should exist")
	assert.NotNil(t, webhooksGroup["create"], "create permission should exist")
	assert.NotNil(t, webhooksGroup["update"], "update permission should exist")
	assert.NotNil(t, webhooksGroup["delete"], "delete permission should exist")

	// Verify admin-only flags
	assert.True(t, webhooksGroup["read_all"].AdminOnly, "Webhook read_all should be admin-only")
	assert.True(t, webhooksGroup["create"].AdminOnly, "Webhook create should be admin-only")
	assert.True(t, webhooksGroup["update"].AdminOnly, "Webhook update should be admin-only")
	assert.True(t, webhooksGroup["delete"].AdminOnly, "Webhook delete should be admin-only")
}

func TestTeamPermissionRegistration(t *testing.T) {
	registerTestAPIRoutes()

	routes := models.GetAPITokenRoutes()
	assert.NotNil(t, routes)
	assert.Contains(t, routes, "v1")

	// Debug: Print what's in the groups
	t.Logf("teams group keys: %v", getKeys(routes["v1"]["teams"]))
	t.Logf("teams_members group keys: %v", getKeys(routes["v1"]["teams_members"]))
	t.Logf("teams_members_admin group keys: %v", getKeys(routes["v1"]["teams_members_admin"]))

	teamsGroup := routes["v1"]["teams"]
	assert.NotNil(t, teamsGroup, "Teams permission group should be registered")

	// Check all team permissions are registered
	assert.NotNil(t, teamsGroup["read_all"], "read_all permission should exist")
	assert.NotNil(t, teamsGroup["read_one"], "read_one permission should exist")
	assert.NotNil(t, teamsGroup["create"], "create permission should exist")
	assert.NotNil(t, teamsGroup["update"], "update permission should exist")
	assert.NotNil(t, teamsGroup["delete"], "delete permission should exist")
	assert.NotNil(t, teamsGroup["add_member"], "add_member permission should exist")
	assert.NotNil(t, teamsGroup["remove_member"], "remove_member permission should exist")
	assert.NotNil(t, teamsGroup["update_member"], "update_member permission should exist")

	// Verify admin-only flags
	assert.True(t, teamsGroup["read_all"].AdminOnly, "Team read_all should be admin-only")
	assert.True(t, teamsGroup["create"].AdminOnly, "Team create should be admin-only")
	assert.True(t, teamsGroup["add_member"].AdminOnly, "Team add_member should be admin-only")
}

func getKeys(m models.APITokenRoute) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestStandardTokenCannotAccessWebhooks(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create a standard token with webhook permissions
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_webhooks": []string{"read_all", "create", "update", "delete"},
	})
	// Explicitly set to standard level (should be default anyway)
	token.TokenLevel = models.APITokenLevelStandard

	// Standard tokens should be denied even with correct permissions
	canReadAll := canTokenAccessRoute(token, "GET", "/api/v1/projects/:project/webhooks")
	assert.False(t, canReadAll, "Standard token should not access webhook read_all even with permission")

	canCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/webhooks")
	assert.False(t, canCreate, "Standard token should not access webhook create even with permission")

	canUpdate := canTokenAccessRoute(token, "POST", "/api/v1/projects/:project/webhooks/:webhook")
	assert.False(t, canUpdate, "Standard token should not access webhook update even with permission")

	canDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/projects/:project/webhooks/:webhook")
	assert.False(t, canDelete, "Standard token should not access webhook delete even with permission")
}

func TestAdminTokenCanAccessWebhooks(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create an admin token with webhook permissions
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_projects_webhooks": []string{"read_all", "create", "update", "delete"},
	})
	token.TokenLevel = models.APITokenLevelAdmin

	// Admin tokens should be allowed
	canReadAll := canTokenAccessRoute(token, "GET", "/api/v1/projects/:project/webhooks")
	assert.True(t, canReadAll, "Admin token with permission should access webhook read_all")

	canCreate := canTokenAccessRoute(token, "PUT", "/api/v1/projects/:project/webhooks")
	assert.True(t, canCreate, "Admin token with permission should access webhook create")

	canUpdate := canTokenAccessRoute(token, "POST", "/api/v1/projects/:project/webhooks/:webhook")
	assert.True(t, canUpdate, "Admin token with permission should access webhook update")

	canDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/projects/:project/webhooks/:webhook")
	assert.True(t, canDelete, "Admin token with permission should access webhook delete")
}

func TestStandardTokenCannotAccessTeams(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create a standard token with team permissions
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_teams": []string{"read_all", "create", "add_member"},
	})
	token.TokenLevel = models.APITokenLevelStandard

	// Standard tokens should be denied even with correct permissions
	canReadAll := canTokenAccessRoute(token, "GET", "/api/v1/teams")
	assert.False(t, canReadAll, "Standard token should not access team read_all even with permission")

	canCreate := canTokenAccessRoute(token, "PUT", "/api/v1/teams")
	assert.False(t, canCreate, "Standard token should not access team create even with permission")

	canAddMember := canTokenAccessRoute(token, "PUT", "/api/v1/teams/:team/members")
	assert.False(t, canAddMember, "Standard token should not access team add_member even with permission")
}

func TestAdminTokenCanAccessTeams(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create an admin token with team permissions
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_teams": []string{"read_all", "read_one", "create", "update", "delete", "add_member", "remove_member", "update_member"},
	})
	token.TokenLevel = models.APITokenLevelAdmin

	// Admin tokens should be allowed
	canReadAll := canTokenAccessRoute(token, "GET", "/api/v1/teams")
	assert.True(t, canReadAll, "Admin token with permission should access team read_all")

	canReadOne := canTokenAccessRoute(token, "GET", "/api/v1/teams/:team")
	assert.True(t, canReadOne, "Admin token with permission should access team read_one")

	canCreate := canTokenAccessRoute(token, "PUT", "/api/v1/teams")
	assert.True(t, canCreate, "Admin token with permission should access team create")

	canUpdate := canTokenAccessRoute(token, "POST", "/api/v1/teams/:team")
	assert.True(t, canUpdate, "Admin token with permission should access team update")

	canDelete := canTokenAccessRoute(token, "DELETE", "/api/v1/teams/:team")
	assert.True(t, canDelete, "Admin token with permission should access team delete")

	canAddMember := canTokenAccessRoute(token, "PUT", "/api/v1/teams/:team/members")
	assert.True(t, canAddMember, "Admin token with permission should access team add_member")

	canRemoveMember := canTokenAccessRoute(token, "DELETE", "/api/v1/teams/:team/members/:user")
	assert.True(t, canRemoveMember, "Admin token with permission should access team remove_member")

	canUpdateMember := canTokenAccessRoute(token, "POST", "/api/v1/teams/:team/members/:user/admin")
	assert.True(t, canUpdateMember, "Admin token with permission should access team update_member")
}

func TestAdminTokenWithoutPermissionsDenied(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Create an admin token WITHOUT webhook/team permissions
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks": []string{"read_one"}, // Only task permissions
	})
	token.TokenLevel = models.APITokenLevelAdmin

	// Admin token should still be denied without the right permissions
	canAccessWebhook := canTokenAccessRoute(token, "GET", "/api/v1/projects/:project/webhooks")
	assert.False(t, canAccessWebhook, "Admin token without webhook permission should be denied")

	canAccessTeam := canTokenAccessRoute(token, "GET", "/api/v1/teams")
	assert.False(t, canAccessTeam, "Admin token without team permission should be denied")
}

func TestStandardTokenCanAccessNonAdminRoutes(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	// Standard tokens should still work for non-admin routes
	token := createTokenWithPermissions(t, s, models.APIPermissions{
		"v1_tasks":    []string{"read_one", "update"},
		"v1_projects": []string{"read_all"},
	})
	token.TokenLevel = models.APITokenLevelStandard

	// Verify standard routes still work
	canReadTask := canTokenAccessRoute(token, "GET", "/api/v1/tasks/:taskid")
	assert.True(t, canReadTask, "Standard token should access non-admin task routes")

	canUpdateTask := canTokenAccessRoute(token, "POST", "/api/v1/tasks/:taskid")
	assert.True(t, canUpdateTask, "Standard token should access non-admin task routes")

	canReadProjects := canTokenAccessRoute(token, "GET", "/api/v1/projects")
	assert.True(t, canReadProjects, "Standard token should access non-admin project routes")
}
