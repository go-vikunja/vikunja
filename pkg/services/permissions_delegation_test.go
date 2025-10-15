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
	"github.com/stretchr/testify/assert"
)

func TestInitPermissionService(t *testing.T) {
	t.Run("InitPermissionService executes without error", func(t *testing.T) {
		// This test verifies that InitPermissionService can be called
		// Currently it's a placeholder, but should not panic or error
		InitPermissionService()
		// Success if we reach here
	})

	t.Run("Permission delegation function variables exist", func(t *testing.T) {
		// Verify that the delegation function variables are defined
		// Some are initialized (Project after T-PERM-006), others are nil until their tasks complete

		// Project permissions (T-PERM-006 - COMPLETE)
		assert.NotNil(t, models.CheckProjectReadFunc, "Project permissions should be initialized after T-PERM-006")
		assert.NotNil(t, models.CheckProjectWriteFunc)
		assert.NotNil(t, models.CheckProjectUpdateFunc)
		assert.NotNil(t, models.CheckProjectDeleteFunc)
		assert.NotNil(t, models.CheckProjectCreateFunc)
		assert.NotNil(t, models.CheckProjectIsAdminFunc)

		// Task permissions (T-PERM-007 - COMPLETE)
		assert.NotNil(t, models.CheckTaskReadFunc, "Task permissions should be initialized after T-PERM-007")
		assert.NotNil(t, models.CheckTaskWriteFunc)
		assert.NotNil(t, models.CheckTaskUpdateFunc)
		assert.NotNil(t, models.CheckTaskDeleteFunc)
		assert.NotNil(t, models.CheckTaskCreateFunc)

		// Label permissions (T-PERM-008 - COMPLETE)
		assert.NotNil(t, models.CheckLabelReadFunc, "Label permissions should be initialized after T-PERM-008")
		assert.NotNil(t, models.CheckLabelUpdateFunc)
		assert.NotNil(t, models.CheckLabelDeleteFunc)
		assert.NotNil(t, models.CheckLabelCreateFunc)

		// Bucket permissions (T-PERM-008 - COMPLETE)
		assert.NotNil(t, models.CheckBucketUpdateFunc, "Bucket permissions should be initialized after T-PERM-008")
		assert.NotNil(t, models.CheckBucketDeleteFunc)
		assert.NotNil(t, models.CheckBucketCreateFunc)

		// LinkShare permissions (T-PERM-009 - COMPLETE)
		assert.NotNil(t, models.CheckLinkShareReadFunc, "LinkShare permissions should be initialized after T-PERM-009")
		assert.NotNil(t, models.CheckLinkShareUpdateFunc)
		assert.NotNil(t, models.CheckLinkShareDeleteFunc)
		assert.NotNil(t, models.CheckLinkShareCreateFunc)

		// Subscription permissions (T-PERM-009 - COMPLETE)
		assert.NotNil(t, models.CheckSubscriptionCreateFunc, "Subscription permissions should be initialized after T-PERM-009")
		assert.NotNil(t, models.CheckSubscriptionDeleteFunc)

		// LabelTask permissions (T-PERM-008 - COMPLETE)
		assert.NotNil(t, models.CheckLabelTaskCreateFunc)
		assert.NotNil(t, models.CheckLabelTaskDeleteFunc)

		// ProjectTeam permissions
		assert.Nil(t, models.CheckProjectTeamReadFunc)
		assert.Nil(t, models.CheckProjectTeamWriteFunc)
		assert.Nil(t, models.CheckProjectTeamUpdateFunc)
		assert.Nil(t, models.CheckProjectTeamDeleteFunc)
		assert.Nil(t, models.CheckProjectTeamCreateFunc)

		// ProjectUser permissions
		assert.Nil(t, models.CheckProjectUserReadFunc)
		assert.Nil(t, models.CheckProjectUserWriteFunc)
		assert.Nil(t, models.CheckProjectUserUpdateFunc)
		assert.Nil(t, models.CheckProjectUserDeleteFunc)
		assert.Nil(t, models.CheckProjectUserCreateFunc)

		// ProjectView permissions
		assert.Nil(t, models.CheckProjectViewReadFunc)
		assert.Nil(t, models.CheckProjectViewWriteFunc)
		assert.Nil(t, models.CheckProjectViewUpdateFunc)
		assert.Nil(t, models.CheckProjectViewDeleteFunc)
		assert.Nil(t, models.CheckProjectViewCreateFunc)

		// Misc permissions
		assert.Nil(t, models.CheckAPITokenDeleteFunc)
		assert.Nil(t, models.CheckReactionCreateFunc)
		assert.Nil(t, models.CheckReactionDeleteFunc)
		assert.Nil(t, models.CheckSavedFilterReadFunc)
		assert.Nil(t, models.CheckSavedFilterWriteFunc)
		assert.Nil(t, models.CheckSavedFilterUpdateFunc)
		assert.Nil(t, models.CheckSavedFilterDeleteFunc)
		assert.Nil(t, models.CheckSavedFilterCreateFunc)
		assert.Nil(t, models.CheckTeamReadFunc)
		assert.Nil(t, models.CheckTeamWriteFunc)
		assert.NotNil(t, models.CheckTeamUpdateFunc) // Wired up in T-PERM-016B follow-up
		assert.NotNil(t, models.CheckTeamDeleteFunc) // Wired up in T-PERM-016B follow-up
		assert.Nil(t, models.CheckTeamCreateFunc)
		assert.Nil(t, models.CheckTeamMemberCreateFunc)
		assert.Nil(t, models.CheckTeamMemberDeleteFunc)
		assert.Nil(t, models.CheckWebhookReadFunc)
		assert.Nil(t, models.CheckWebhookUpdateFunc)
		assert.Nil(t, models.CheckWebhookDeleteFunc)
		assert.Nil(t, models.CheckWebhookCreateFunc)
		assert.Nil(t, models.CheckBulkTaskUpdateFunc)
		assert.Nil(t, models.CheckProjectDuplicateCreateFunc)
		assert.Nil(t, models.CheckTaskPositionUpdateFunc)
	})
}
