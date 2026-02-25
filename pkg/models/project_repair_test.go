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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepairOrphanedProjects(t *testing.T) {
	t.Run("finds and repairs orphaned projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		result, err := RepairOrphanedProjects(s, false)
		require.NoError(t, err)

		assert.Equal(t, 1, result.Repaired)

		// Verify the project was re-parented to top level
		project := &Project{ID: 39}
		has, err := s.Get(project)
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, int64(0), project.ParentProjectID)
	})

	t.Run("dry run does not modify anything", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		result, err := RepairOrphanedProjects(s, true)
		require.NoError(t, err)

		assert.Equal(t, 1, result.Found)
		assert.Equal(t, 0, result.Repaired)

		// Verify the project was NOT changed
		project := &Project{ID: 39}
		has, err := s.Get(project)
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, int64(999999), project.ParentProjectID)
	})

	t.Run("no orphans returns zero counts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// First, repair all orphans
		_, err := RepairOrphanedProjects(s, false)
		require.NoError(t, err)

		// Run again - should find nothing
		result, err := RepairOrphanedProjects(s, false)
		require.NoError(t, err)

		assert.Equal(t, 0, result.Found)
		assert.Equal(t, 0, result.Repaired)
	})
}

