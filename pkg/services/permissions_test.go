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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionService_New(t *testing.T) {
	t.Run("NewPermissionService creates service instance", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ps := NewPermissionService(s.Engine())

		require.NotNil(t, ps)
		assert.NotNil(t, ps.DB)
		assert.NotNil(t, ps.Registry) // Registry should be set
	})

	t.Run("Registry provides services", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ps := NewPermissionService(s.Engine())

		// Registry provides all services
		projectService := ps.Registry.Project()
		require.NotNil(t, projectService)

		// Multiple accesses return same instance (singleton)
		projectService2 := ps.Registry.Project()
		assert.Equal(t, projectService, projectService2)
	})

	t.Run("Registry singleton pattern for TaskService", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ps := NewPermissionService(s.Engine())

		// Registry provides task service
		taskService := ps.Registry.Task()
		require.NotNil(t, taskService)

		// Multiple accesses return same instance (singleton)
		taskService2 := ps.Registry.Task()
		assert.Equal(t, taskService, taskService2)
	})

	t.Run("Lazy loading of LabelService", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ps := NewPermissionService(s.Engine())

		// Registry provides label service
		labelService := ps.Registry.Label()
		require.NotNil(t, labelService)

		// Multiple accesses return same instance (singleton)
		labelService2 := ps.Registry.Label()
		assert.Equal(t, labelService, labelService2)
	})
}
