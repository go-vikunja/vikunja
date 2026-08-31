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
	"sort"
	"sync"
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
		assert.Equal(t, []int64{10, 11, 5, 12, 13}, indexes)

		counter := &ProjectTaskCounter{}
		has, err := s.ID(4).Get(counter)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(13), counter.LastIndex)
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
		assert.Equal(t, []int64{100, 101, 102, 103}, indexes)
	})

	t.Run("missing counter is an invariant violation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := setNewTaskIndexes(s, 999, []*Task{{}})
		require.ErrorContains(t, err, "task index counter")
	})

	t.Run("rollback releases the reservation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		require.NoError(t, setNewTaskIndexes(s, 4, []*Task{{}, {}}))
		require.NoError(t, s.Rollback())
		require.NoError(t, s.Close())

		s = db.NewSession()
		defer s.Close()
		tasks := []*Task{{}}
		require.NoError(t, setNewTaskIndexes(s, 4, tasks))
		assert.Equal(t, int64(1), tasks[0].Index)
	})

	t.Run("concurrent reservations are unique", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		oldMaxOpen := x.DB().Stats().MaxOpenConnections
		x.SetMaxOpenConns(1)
		defer x.SetMaxOpenConns(oldMaxOpen)

		start := make(chan struct{})
		indexes := make(chan int64, 2)
		errs := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		for range 2 {
			go func() {
				defer wg.Done()
				<-start
				s := db.NewSession()
				defer s.Close()
				tasks := []*Task{{}}
				if err := setNewTaskIndexes(s, 4, tasks); err != nil {
					errs <- err
					return
				}
				if err := s.Commit(); err != nil {
					errs <- err
					return
				}
				indexes <- tasks[0].Index
			}()
		}
		close(start)
		wg.Wait()
		close(indexes)
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}
		got := make([]int64, 0, 2)
		for index := range indexes {
			got = append(got, index)
		}
		sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
		assert.Equal(t, []int64{1, 2}, got)
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
	t.Run("failed create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		usr := &user.User{ID: 1}
		s := db.NewSession()

		bucket := &Bucket{Title: "full batch", ProjectViewID: 4, CreatedByID: 1, Limit: 1}
		_, err := s.Insert(bucket)
		require.NoError(t, err)
		batch := &BulkTaskCreation{
			ProjectID: 1,
			Tasks: []*Task{
				{Title: "fits", BucketID: bucket.ID},
				{Title: "overflows", BucketID: bucket.ID},
			},
		}
		require.Error(t, batch.Create(s, usr))
		require.NoError(t, s.Rollback())
		require.NoError(t, s.Close())

		s = db.NewSession()
		defer s.Close()
		next := &Task{Title: "after failed batch", ProjectID: 1}
		require.NoError(t, next.Create(s, usr))
		assert.Equal(t, int64(35), next.Index)
	})

	t.Run("failed move", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		move := &Task{ID: 12, ProjectID: 2, CoverImageAttachmentID: 999}
		require.Error(t, move.Update(s, &user.User{ID: 1}))
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
	})
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

	t.Run("alias conflict aborts the move", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&TaskIndexAlias{ProjectID: 1, Index: 12, TaskID: 1})
		require.NoError(t, err)
		require.Error(t, (&Task{ID: 12, ProjectID: 2}).Update(s, &user.User{ID: 1}))

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
