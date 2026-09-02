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
	"math"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectCreateInitializesTaskIndexCounter(t *testing.T) {
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
}

func TestSetNewTaskIndexes(t *testing.T) {
	t.Run("presets only advance the high water mark", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tasks := []*Task{
			{Index: 10},
			{},
			{Index: 5},
			{Index: 10},
			{},
		}
		require.NoError(t, setNewTaskIndexes(s, 4, tasks))

		indexes := make([]int64, len(tasks))
		for i, task := range tasks {
			indexes[i] = task.Index
		}
		assert.Equal(t, []int64{10, 1, 5, 2, 3}, indexes)

		counter := &ProjectTaskCounter{}
		has, err := s.ID(4).Get(counter)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(10), counter.LastIndex)
	})

	t.Run("retired and duplicate presets get new indexes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tasks := []*Task{
			{Index: 100},
			{Index: 1},
			{Index: 100},
			{},
		}
		require.NoError(t, setNewTaskIndexes(s, 1, tasks))

		indexes := make([]int64, len(tasks))
		for i, task := range tasks {
			indexes[i] = task.Index
		}
		assert.Equal(t, []int64{100, 35, 36, 37}, indexes)
	})

	t.Run("missing counter is seeded from existing tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(1).Delete(&ProjectTaskCounter{})
		require.NoError(t, err)

		tasks := []*Task{{}}
		require.NoError(t, setNewTaskIndexes(s, 1, tasks))
		// Task 51 holds index 34 in project 1 and is soft deleted.
		assert.Equal(t, int64(35), tasks[0].Index)

		counter := &ProjectTaskCounter{}
		has, err := s.ID(1).Get(counter)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(35), counter.LastIndex)
	})

	t.Run("missing counter is seeded from retired alias indexes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.ID(1).Delete(&ProjectTaskCounter{})
		require.NoError(t, err)
		_, err = s.Insert(&TaskIndexAlias{ProjectID: 1, Index: 90, TaskID: 1})
		require.NoError(t, err)

		tasks := []*Task{{}}
		require.NoError(t, setNewTaskIndexes(s, 1, tasks))
		assert.Equal(t, int64(91), tasks[0].Index)
	})

	t.Run("presets beyond the bound get a fresh index", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tasks := []*Task{{Index: math.MaxInt32 + 1}}
		require.NoError(t, setNewTaskIndexes(s, 1, tasks))
		assert.Equal(t, int64(35), tasks[0].Index)

		counter := &ProjectTaskCounter{}
		has, err := s.ID(1).Get(counter)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(35), counter.LastIndex)
	})
}

func TestTaskIndexesAreNeverReused(t *testing.T) {
	t.Run("after permanent deletion", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		usr := &user.User{ID: 1}

		s := db.NewSession()
		created := &Task{Title: "to permanently delete", ProjectID: 1}
		require.NoError(t, created.Create(s, usr))
		assert.Equal(t, int64(35), created.Index)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		s = db.NewSession()
		require.NoError(t, hardDeleteTask(s, created))
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		s = db.NewSession()
		defer s.Close()
		next := &Task{Title: "after permanent delete", ProjectID: 1}
		require.NoError(t, next.Create(s, usr))
		assert.Equal(t, int64(36), next.Index)
	})

	t.Run("after moving the highest indexed task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		usr := &user.User{ID: 1}

		s := db.NewSession()
		created := &Task{Title: "move me", ProjectID: 2}
		require.NoError(t, created.Create(s, usr))
		assert.Equal(t, int64(3), created.Index)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		s = db.NewSession()
		moved := &Task{ID: created.ID, ProjectID: 1}
		require.NoError(t, moved.Update(s, usr))
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		s = db.NewSession()
		defer s.Close()
		next := &Task{Title: "after move", ProjectID: 2}
		require.NoError(t, next.Create(s, usr))
		assert.Equal(t, int64(4), next.Index)
	})
}

func TestTaskIndexReservationRollsBackWithTaskChanges(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	move := &Task{ID: 12, ProjectID: 2, CoverImageAttachmentID: 999}
	err := move.Update(s, &user.User{ID: 1})
	var attachmentErr *ErrAttachmentDoesNotBelongToTask
	require.ErrorAs(t, err, &attachmentErr)
	require.NoError(t, s.Rollback())
	require.NoError(t, s.Close())

	s = db.NewSession()
	defer s.Close()
	counter := &ProjectTaskCounter{}
	has, err := s.ID(2).Get(counter)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, int64(2), counter.LastIndex)

	persisted := &Task{ID: 12}
	has, err = s.Get(persisted)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, int64(1), persisted.ProjectID)

	aliases, err := s.Where("task_id = ?", 12).Count(&TaskIndexAlias{})
	require.NoError(t, err)
	assert.Zero(t, aliases)
}

func TestTaskIndexAliases(t *testing.T) {
	t.Run("move records the retired address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		moved := &Task{ID: 12, ProjectID: 2}
		require.NoError(t, moved.Update(s, &user.User{ID: 1}))
		assert.Equal(t, int64(3), moved.Index)

		alias := &TaskIndexAlias{}
		has, err := s.Where("project_id = ? AND `index` = ?", 1, 12).Get(alias)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(12), alias.TaskID)
	})

	t.Run("ordinary update records no alias", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, (&Task{ID: 12, Title: "still here"}).Update(s, &user.User{ID: 1}))
		aliases, err := s.Count(&TaskIndexAlias{})
		require.NoError(t, err)
		assert.Zero(t, aliases)
	})

	t.Run("moving back retains every retired address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		firstMove := &Task{ID: 12, ProjectID: 2}
		require.NoError(t, firstMove.Update(s, &user.User{ID: 1}))
		assert.Equal(t, int64(3), firstMove.Index)

		secondMove := &Task{ID: 12, ProjectID: 1}
		require.NoError(t, secondMove.Update(s, &user.User{ID: 1}))
		assert.Equal(t, int64(35), secondMove.Index)

		aliases := []*TaskIndexAlias{}
		require.NoError(t, s.Where("task_id = ?", 12).OrderBy("project_id").Find(&aliases))
		require.Len(t, aliases, 2)
		assert.Equal(t, &TaskIndexAlias{ProjectID: 1, Index: 12, TaskID: 12}, aliases[0])
		assert.Equal(t, &TaskIndexAlias{ProjectID: 2, Index: 3, TaskID: 12}, aliases[1])
	})

	t.Run("alias records the latest holder", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&TaskIndexAlias{ProjectID: 1, Index: 12, TaskID: 1})
		require.NoError(t, err)

		require.NoError(t, (&Task{ID: 12, ProjectID: 2}).Update(s, &user.User{ID: 1}))

		alias := &TaskIndexAlias{}
		has, err := s.Where("project_id = ? AND `index` = ?", 1, 12).Get(alias)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(12), alias.TaskID)
	})

	t.Run("soft deletion retains aliases and counters", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		moveSession := db.NewSession()
		moved := &Task{ID: 12, ProjectID: 2}
		require.NoError(t, moved.Update(moveSession, &user.User{ID: 1}))
		require.NoError(t, moveSession.Commit())
		require.NoError(t, moveSession.Close())

		deleteSession := db.NewSession()
		require.NoError(t, (&Task{ID: 12}).Delete(deleteSession, &user.User{ID: 3}))
		require.NoError(t, deleteSession.Commit())
		require.NoError(t, deleteSession.Close())

		s := db.NewSession()
		defer s.Close()
		aliases, err := s.Where("task_id = ?", 12).Count(&TaskIndexAlias{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), aliases)
		counter := &ProjectTaskCounter{}
		has, err := s.ID(2).Get(counter)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(3), counter.LastIndex)
	})
}
