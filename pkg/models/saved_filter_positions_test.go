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

func TestSavedFilterUpdateInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "posfilter",
		Filters: &TaskCollection{Filter: "id = 1"},
	}

	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	err = sf.Update(s, u)
	require.NoError(t, err)

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	tp := &TaskPosition{}
	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, 1).Get(tp)
	require.NoError(t, err)
	require.True(t, exists)
	assert.NotZero(t, tp.Position)
}

func TestCronInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "cronfilter",
		Filters: &TaskCollection{Filter: "due_date > '2018-01-01T00:00:00'"},
	}

	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	task := &Task{}
	exists, err = s.Where("id = ?", 5).Get(task)
	require.NoError(t, err)
	require.True(t, exists)

	tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 0}
	_, err = s.Insert(tp)
	require.NoError(t, err)

	_, err = calculateNewPositionForTask(s, u, task, view)
	require.NoError(t, err)

	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, task.ID).Get(tp)
	require.NoError(t, err)
	require.True(t, exists)
	assert.NotZero(t, tp.Position)
}
