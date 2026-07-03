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
	// Created last on purpose: it has the highest id, so the default id-asc order
	// would place it after the one-word matches. It ranking above them can only
	// come from BM25 relevance, not from the fallback ordering.
	lateAllWords := &Task{Title: "server backup checklist", ProjectID: 1}
	require.NoError(t, lateAllWords.Create(s, usr))

	assertRelevanceRanked := func(t *testing.T, tc *TaskCollection) {
		got, _, _, err := tc.ReadAll(s, usr, "backup server", 1, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		gotIDs := make([]int64, len(gotTasks))
		for i, tsk := range gotTasks {
			gotIDs[i] = tsk.ID
		}

		require.Contains(t, gotIDs, allWords.ID, "the task matching all words should be returned")

		if db.ParadeDBAvailable() {
			// Compare only the tasks created by this test so fixture tasks (present
			// or future) matching the search cannot break the order assertions.
			created := map[int64]bool{allWords.ID: true, oneWordA.ID: true, oneWordB.ID: true, lateAllWords.ID: true}
			pos := map[int64]int{}
			for _, id := range gotIDs {
				if created[id] {
					pos[id] = len(pos)
				}
			}
			require.Len(t, pos, len(created), "all created tasks should match the search")

			for _, allw := range []int64{allWords.ID, lateAllWords.ID} {
				for _, onew := range []int64{oneWordA.ID, oneWordB.ID} {
					assert.Less(t, pos[allw], pos[onew], "tasks matching all query words should rank above one-word matches by BM25 relevance")
				}
			}
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

	// An explicit sort_by must win over relevance: with `id desc` the lowest-id
	// task (allWords) ranks last, the opposite of what BM25 relevance would do.
	// This locks the contract that user-provided sorting disables relevance
	// ranking even on ParadeDB. Only ParadeDB's per-token search matches all
	// three tasks, so the ordering contract is only asserted there (other
	// databases ILIKE the whole phrase and match a different subset).
	t.Run("explicit sort disables relevance ranking", func(t *testing.T) {
		if !db.ParadeDBAvailable() {
			t.Skip("relevance ranking only applies on ParadeDB")
		}

		tc := &TaskCollection{
			ProjectID: 1,
			SortBy:    []string{"id"},
			OrderBy:   []string{"desc"},
		}
		got, _, _, err := tc.ReadAll(s, usr, "backup server", 1, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		created := map[int64]bool{allWords.ID: true, oneWordA.ID: true, oneWordB.ID: true}
		var orderedIDs []int64
		for _, tsk := range gotTasks {
			if created[tsk.ID] {
				orderedIDs = append(orderedIDs, tsk.ID)
			}
		}

		require.Len(t, orderedIDs, len(created), "all created tasks should match the search")
		for i := 1; i < len(orderedIDs); i++ {
			assert.Greater(t, orderedIDs[i-1], orderedIDs[i], "tasks must follow the explicit id-desc sort, not relevance")
		}
		assert.Equal(t, allWords.ID, orderedIDs[len(orderedIDs)-1], "the all-words match (lowest id) ranks last under id-desc, proving relevance was not applied")
	})

	// The all-projects scope appends the Favorites pseudo-project whenever the user
	// has any favorited task (fixtures give user 1 tasks 1 and 15). Both live in
	// projects the user can access, so the favorites arm is redundant, gets dropped
	// and the global search stays relevance-ranked.
	t.Run("global search with in-scope favorites", func(t *testing.T) {
		assertRelevanceRanked(t, &TaskCollection{})
	})

	// Task 13 lives in project 2, which user 1 cannot access: the favorites arm is
	// load-bearing, so the query keeps it and falls back to unranked ordering
	// instead of failing with an unsupported-query-shape error on ParadeDB.
	t.Run("global search with out-of-scope favorite stays unranked", func(t *testing.T) {
		outOfScopeFavorite := &Favorite{EntityID: 13, UserID: usr.ID, Kind: FavoriteKindTask}
		_, err := s.Insert(outOfScopeFavorite)
		require.NoError(t, err)
		defer func() {
			_, err := s.Delete(outOfScopeFavorite)
			require.NoError(t, err)
		}()

		tc := &TaskCollection{}
		got, _, _, err := tc.ReadAll(s, usr, "backup server", 1, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		gotIDs := make([]int64, 0, len(gotTasks))
		for _, tsk := range gotTasks {
			gotIDs = append(gotIDs, tsk.ID)
		}
		require.Contains(t, gotIDs, allWords.ID)
	})

	// sort_by=done,relevance keeps undone tasks first and ranks by relevance within
	// each group — the quick-actions search shape. On databases that cannot score,
	// the relevance param is dropped and the done,id ordering remains.
	t.Run("explicit done and relevance sort", func(t *testing.T) {
		doneAllWords := &Task{Title: "Backup server runbook", ProjectID: 1, Done: true}
		require.NoError(t, doneAllWords.Create(s, usr))

		tc := &TaskCollection{
			ProjectID: 1,
			SortBy:    []string{"done", "relevance"},
		}
		got, _, _, err := tc.ReadAll(s, usr, "backup server", 1, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		pos := map[int64]int{}
		for i, tsk := range gotTasks {
			pos[tsk.ID] = i
		}

		require.Contains(t, pos, allWords.ID)
		require.Contains(t, pos, doneAllWords.ID)
		assert.Less(t, pos[allWords.ID], pos[doneAllWords.ID], "undone tasks must come before done tasks")

		if db.ParadeDBAvailable() {
			// doneAllWords matches all query words: only the done sort taking
			// precedence over its high BM25 score can place it last.
			for _, undone := range []int64{allWords.ID, oneWordA.ID, oneWordB.ID, lateAllWords.ID} {
				require.Contains(t, pos, undone)
				assert.Less(t, pos[undone], pos[doneAllWords.ID], "undone tasks must come before the done all-words match")
			}
			for _, allw := range []int64{allWords.ID, lateAllWords.ID} {
				for _, onew := range []int64{oneWordA.ID, oneWordB.ID} {
					assert.Less(t, pos[allw], pos[onew], "within the undone group, all-words matches must rank above one-word matches")
				}
			}
		}
	})

	// A numeric search cannot be scored; the relevance param must be dropped
	// instead of producing an unsupported-query-shape error.
	t.Run("relevance sort with numeric search falls back", func(t *testing.T) {
		tc := &TaskCollection{
			ProjectID: 1,
			SortBy:    []string{"relevance"},
		}
		got, _, _, err := tc.ReadAll(s, usr, "#1", 1, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)
		assert.NotEmpty(t, gotTasks)
	})
}

// TestTaskSearchMatching pins ParadeDB's matching contract: per-token matching
// across title and description, word-order independence, and fuzzy matching
// within edit distance 1 against token prefixes (fuzzy(1, t)). Other databases
// fall back to a substring match of the whole search string over each field and
// intentionally match less, so this test is ParadeDB-only.
func TestTaskSearchMatching(t *testing.T) {
	if !db.ParadeDBAvailable() {
		t.Skip("requires ParadeDB")
	}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	usr := &user.User{ID: 1}

	newTask := func(title, description string) *Task {
		tsk := &Task{Title: title, Description: description, ProjectID: 1}
		require.NoError(t, tsk.Create(s, usr))
		return tsk
	}

	helloWorld := newTask("hello world", "")
	greenOrange := newTask("what is going on with the green orange two", "and some other data")
	orangeTape := newTask("orange twosided tape and so twoish", "sensible data is here to test")
	meeting := newTask("weekly meeting notes", "discuss the quarterly budget forecast")
	ficus := newTask("untitled chore", "watering the ficus plant")
	k8s := newTask("kubernetes cluster maintenance", "")

	created := []*Task{helloWorld, greenOrange, orangeTape, meeting, ficus, k8s}

	search := func(t *testing.T, query string) map[int64]int {
		tc := &TaskCollection{ProjectID: 1}
		got, _, _, err := tc.ReadAll(s, usr, query, 1, 50)
		require.NoError(t, err)

		gotTasks, is := got.([]*Task)
		require.True(t, is)

		pos := map[int64]int{}
		for i, tsk := range gotTasks {
			pos[tsk.ID] = i
		}
		return pos
	}

	tests := []struct {
		name   string
		search string
		want   []*Task
	}{
		{"prefix match", "Hell", []*Task{helloWorld}},
		// helloWorld matches too: fuzzy and prefix compose, "wo" is one edit
		// from "two" and a prefix of "world".
		{"prefix match across tasks", "two", []*Task{greenOrange, orangeTape, helloWorld}},
		{"all words in one field", "Hello World", []*Task{helloWorld}},
		{"typo one edit away", "Hello orld", []*Task{helloWorld}},
		{"typo in one of two words", "kubernets cluster", []*Task{k8s}},
		{"description-only match", "data", []*Task{greenOrange, orangeTape}},
		{"description-only single task", "ficus", []*Task{ficus}},
		{"terms split between title and description", "meeting budget", []*Task{meeting}},
		{"terms split across tasks' fields", "data orange", []*Task{greenOrange, orangeTape}},
		{"word order reversed", "orange green", []*Task{greenOrange, orangeTape}},
		{"word order as written", "green orange", []*Task{greenOrange, orangeTape}},
		{"no matching token", "xylophone", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := search(t, tt.search)

			want := map[int64]bool{}
			for _, tsk := range tt.want {
				want[tsk.ID] = true
			}

			// Only the created tasks are asserted on, fixture matches are ignored.
			for _, tsk := range created {
				if want[tsk.ID] {
					assert.Contains(t, pos, tsk.ID, "search %q should match task %q", tt.search, tsk.Title)
				} else {
					assert.NotContains(t, pos, tsk.ID, "search %q should not match task %q", tt.search, tsk.Title)
				}
			}
		})
	}

	// Regardless of word order, the task matching both words must rank above
	// the task matching only one.
	for _, query := range []string{"green orange", "orange green"} {
		t.Run("ranking "+query, func(t *testing.T) {
			pos := search(t, query)
			require.Contains(t, pos, greenOrange.ID)
			require.Contains(t, pos, orangeTape.ID)
			assert.Less(t, pos[greenOrange.ID], pos[orangeTape.ID], "the task matching both words should rank first")
		})
	}
}
