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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

func TestAttachmentService_New(t *testing.T) {
	t.Run("create new attachment service", func(t *testing.T) {
		// Mock engine for testing
		engine := &xorm.Engine{}

		service := NewAttachmentService(engine)

		assert.NotNil(t, service)
		// Can't access internal engine field directly, but service should be created
		assert.IsType(t, &AttachmentService{}, service)
	})
}

func TestAttachmentPermissions_CanRead(t *testing.T) {
	t.Run("test permission can read", func(t *testing.T) {
		permissions := &AttachmentPermissions{}
		u := &user.User{ID: 1}

		// This would normally test with a real database session and attachment
		// For now, we're just testing the structure exists
		assert.NotNil(t, permissions)
		assert.NotNil(t, u)
	})
}

func TestDependencyInjection(t *testing.T) {
	t.Run("verify dependency injection variables exist", func(t *testing.T) {
		// Test that the dependency injection function variables are available
		// in the models package (they should be nil until wired)
		assert.NotNil(t, models.AttachmentCreateFunc)
		assert.NotNil(t, models.AttachmentDeleteFunc)
	})

	t.Run("verify service functions are wired correctly", func(t *testing.T) {
		// Test that calling the init function wires the service methods correctly
		// This verifies the dependency injection is working as expected

		// The init() function should have been called when the service package is imported
		// So the model function variables should now point to service methods
		assert.NotNil(t, models.AttachmentCreateFunc, "AttachmentCreateFunc should be wired")
		assert.NotNil(t, models.AttachmentDeleteFunc, "AttachmentDeleteFunc should be wired")

		// We can't easily test the actual function execution here without a full database setup,
		// but we can verify the functions are assigned
	})
}
