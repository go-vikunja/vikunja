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

func TestTaskIndexStateFixtures(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	counters := []*ProjectTaskCounter{}
	require.NoError(t, s.OrderBy("project_id").Find(&counters))
	require.Len(t, counters, 44)
	assert.Equal(t, &ProjectTaskCounter{ProjectID: 1, LastIndex: 34}, counters[0])
	assert.Equal(t, &ProjectTaskCounter{ProjectID: 4, LastIndex: 0}, counters[3])

	aliases, err := s.Count(&TaskIndexAlias{})
	require.NoError(t, err)
	assert.Zero(t, aliases)

	_, err = s.Insert(&TaskIndexAlias{ProjectID: 1, Index: 99, TaskID: 1})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	db.LoadAndAssertFixtures(t)
	s = db.NewSession()
	defer s.Close()
	aliases, err = s.Count(&TaskIndexAlias{})
	require.NoError(t, err)
	assert.Zero(t, aliases)
}

func TestTaskIndexStateConstraints(t *testing.T) {
	t.Run("one counter per project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&ProjectTaskCounter{ProjectID: 1, LastIndex: 999})
		require.Error(t, err)
	})

	t.Run("one alias per historical address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&TaskIndexAlias{ProjectID: 1, Index: 99, TaskID: 1})
		require.NoError(t, err)
		_, err = s.Insert(&TaskIndexAlias{ProjectID: 1, Index: 99, TaskID: 2})
		require.Error(t, err)
	})
}

func TestProjectCreateInitializesTaskIndexCounter(t *testing.T) {
	t.Run("commit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		project := &Project{Title: "counter project"}
		require.NoError(t, project.Create(s, &user.User{ID: 1}))

		counter := &ProjectTaskCounter{}
		has, err := s.ID(project.ID).Get(counter)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(0), counter.LastIndex)

		count, err := s.Where("project_id = ?", project.ID).Count(&ProjectTaskCounter{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		require.NoError(t, s.Commit())
	})

	t.Run("rollback", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		project := &Project{Title: "rolled back counter project"}
		require.NoError(t, project.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Rollback())
		require.NoError(t, s.Close())

		check := db.NewSession()
		defer check.Close()
		projectCount, err := check.ID(project.ID).Count(&Project{})
		require.NoError(t, err)
		assert.Zero(t, projectCount)
		counterCount, err := check.ID(project.ID).Count(&ProjectTaskCounter{})
		require.NoError(t, err)
		assert.Zero(t, counterCount)
	})
}
