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

	"github.com/stretchr/testify/assert"
)

// Pure structure tests - no DB required
// CRUD and permission tests moved to pkg/services/subscription_test.go

func TestSubscriptionGetTypeFromString(t *testing.T) {
	t.Run("project", func(t *testing.T) {
		entityType := getEntityTypeFromString("project")
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityProject), entityType)
	})
	t.Run("task", func(t *testing.T) {
		entityType := getEntityTypeFromString("task")
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityTask), entityType)
	})
	t.Run("invalid", func(t *testing.T) {
		entityType := getEntityTypeFromString("someomejghsd")
		assert.Equal(t, SubscriptionEntityType(SubscriptionEntityUnknown), entityType)
	})
}
