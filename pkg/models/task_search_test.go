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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKanbanViewBucketFiltering(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	view, err := GetProjectViewByID(s, 4)
	require.NoError(t, err)

	project, err := GetProjectSimpleByID(s, view.ProjectID)
	require.NoError(t, err)

	buckets, err := GetTasksInBucketsForView(s, view, []*Project{project}, &taskSearchOptions{}, &user.User{ID: 1})
	require.NoError(t, err)

	taskBuckets := map[int64][]int64{}
	for _, b := range buckets {
		for _, tsk := range b.Tasks {
			taskBuckets[tsk.ID] = append(taskBuckets[tsk.ID], b.ID)
		}
	}

	for tid, bs := range taskBuckets {
		assert.Lenf(t, bs, 1, "task %d appears in multiple buckets: %v", tid, bs)
	}

	for _, id := range []int64{40, 41, 42, 43, 44, 45, 46} {
		assert.NotContains(t, taskBuckets, id)
	}
}

// TestTaskSearchRelevanceRanking verifies that a multi-word search ranks the task
// matching all words above tasks matching only some. The ranking is BM25-based and
// therefore only enforced on ParadeDB; on other databases we only assert that the
// matching tasks are returned (no order guarantee), keeping the test green across
// the whole CI database matrix.
func TestTaskSearchRelevanceRanking(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	usr := &user.User{ID: 1}

	allWords := &Task{Title: "Backup server migration", ProjectID: 1}
	require.NoError(t, allWords.Create(s, usr))
	oneWordA := &Task{Title: "Backup of old files", ProjectID: 1}
	require.NoError(t, oneWordA.Create(s, usr))
	oneWordB := &Task{Title: "server room booking", ProjectID: 1}
	require.NoError(t, oneWordB.Create(s, usr))

	assertRelevanceRanked := func(t *testing.T, tc *TaskCollection) {
		got, _, _, err := tc.ReadAll(s, usr, "backup server", 0, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		gotIDs := make([]int64, len(gotTasks))
		for i, tsk := range gotTasks {
			gotIDs[i] = tsk.ID
		}

		require.Contains(t, gotIDs, allWords.ID, "the task matching all words should be returned")

		if db.ParadeDBAvailable() {
			require.NotEmpty(t, gotTasks)
			assert.Equal(t, allWords.ID, gotTasks[0].ID, "task matching all query words should rank first by BM25 relevance")
		}
	}

	// Without a view: plain "tasks.*, pdb.score(tasks.id)" select.
	t.Run("no view", func(t *testing.T) {
		assertRelevanceRanked(t, &TaskCollection{ProjectID: 1})
	})

	// With a view: exercises the task_positions LEFT JOIN, which adds
	// task_positions.position to the DISTINCT select alongside pdb.score(tasks.id).
	t.Run("list view", func(t *testing.T) {
		assertRelevanceRanked(t, &TaskCollection{ProjectID: 1, ProjectViewID: 1})
	})
}

// TestTaskSearchRelevanceRankingNumericIndex covers a numeric search (e.g. "#42"):
// it matches both a task by its per-project index and tasks whose title/description
// contain that number via fuzzy text search. On ParadeDB the exact-index task must
// rank first, then the text matches by relevance. This is the case that combines an
// `index = N` equality with the ParadeDB ||| operators; scoring such a mixed boolean
// group is an unsupported query shape, so it is run as two arms (index, then text).
func TestTaskSearchRelevanceRankingNumericIndex(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	usr := &user.User{ID: 1}

	// The exact-index task: its index is what we search for. Its title deliberately
	// does not contain the number, so it can only be found by the index match.
	exactIndex := &Task{Title: "Quarterly planning offsite", ProjectID: 1}
	require.NoError(t, exactIndex.Create(s, usr))
	require.NotZero(t, exactIndex.Index)

	indexStr := strconv.FormatInt(exactIndex.Index, 10)
	search := "#" + indexStr

	// Text matches: their titles contain the searched number so the fuzzy text arm
	// returns them, but they are not the exact-index task.
	textA := &Task{Title: "Review ticket " + search + " backlog", ProjectID: 1}
	require.NoError(t, textA.Create(s, usr))
	textB := &Task{Title: "Notes about " + search, ProjectID: 1}
	require.NoError(t, textB.Create(s, usr))

	assertIndexFirst := func(t *testing.T, tc *TaskCollection) {
		got, _, _, err := tc.ReadAll(s, usr, search, 0, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		gotIDs := make([]int64, len(gotTasks))
		for i, tsk := range gotTasks {
			gotIDs[i] = tsk.ID
		}

		require.Contains(t, gotIDs, exactIndex.ID, "the exact-index task should be returned")

		if db.ParadeDBAvailable() {
			require.NotEmpty(t, gotTasks)
			assert.Equal(t, exactIndex.ID, gotTasks[0].ID, "the exact-index match should rank first")
			assert.Contains(t, gotIDs, textA.ID, "text matches should also be returned, ranked after the index match")
		}
	}

	t.Run("no view", func(t *testing.T) {
		assertIndexFirst(t, &TaskCollection{ProjectID: 1})
	})

	t.Run("list view", func(t *testing.T) {
		assertIndexFirst(t, &TaskCollection{ProjectID: 1, ProjectViewID: 1})
	})
}
