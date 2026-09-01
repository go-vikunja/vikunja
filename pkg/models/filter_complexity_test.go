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
	"fmt"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Guards GHSA-xxc3-xpmc-vmvr before fexpr parses attacker-controlled input.
func TestFilterComplexityLimits(t *testing.T) {
	parens := func(depth int) string {
		return strings.Repeat("(", depth) + "done = true" + strings.Repeat(")", depth)
	}

	t.Run("task filter: depth 100 passes", func(t *testing.T) {
		filters, err := getTaskFiltersFromFilterString(parens(100), "UTC")
		require.NoError(t, err)
		require.Len(t, filters, 1)
	})

	t.Run("task filter: depth 101 is rejected with a small byte size", func(t *testing.T) {
		filter := parens(101)
		require.Less(t, len(filter), 4096, "the depth check must reject independently of the byte cap")
		_, err := getTaskFiltersFromFilterString(filter, "UTC")
		require.Error(t, err)
		var tooComplex *ErrFilterTooComplex
		require.ErrorAs(t, err, &tooComplex)
		assert.Equal(t, fmt.Sprintf("it exceeds the %d-level nesting limit", maxFilterDepth), tooComplex.Reason)
	})

	t.Run("time entry filter: depth 101 is rejected", func(t *testing.T) {
		_, err := timeEntryFilterCond(parens(101), "UTC")
		require.Error(t, err)
		assert.True(t, IsErrFilterTooComplex(err))
	})

	t.Run("time entry filter: normal nested control works", func(t *testing.T) {
		cond, err := timeEntryFilterCond("(task_id = 1 && start_time = 1544500000) || task_id = 2", "UTC")
		require.NoError(t, err)
		require.NotNil(t, cond)
	})

	t.Run("oversized input is rejected without being echoed", func(t *testing.T) {
		filter := strings.Repeat("done = false && ", 1300) // ~20 KiB
		_, err := getTaskFiltersFromFilterString(filter, "UTC")
		require.Error(t, err)
		var tooComplex *ErrFilterTooComplex
		require.ErrorAs(t, err, &tooComplex)
		assert.Equal(t, fmt.Sprintf("it exceeds the %d-byte limit", maxFilterBytes), tooComplex.Reason)
		assert.NotContains(t, err.Error(), "done = false && done = false")
	})

	t.Run("oversized raw input is rejected before preprocessing", func(t *testing.T) {
		oversized := func(clause string) string {
			repetitions := maxFilterBytes/len(clause+" && ") + 1
			filter := strings.Repeat(clause+" && ", repetitions) + clause
			require.Greater(t, len(filter), maxFilterBytes)
			require.Less(t, len(preprocessFilterString(filter)), maxFilterBytes)
			return filter
		}

		_, err := getTaskFiltersFromFilterString(oversized("project not in 1"), "UTC")
		require.Error(t, err)
		assert.True(t, IsErrFilterTooComplex(err))

		_, err = timeEntryFilterCond(oversized("project_id not in 1"), "UTC")
		require.Error(t, err)
		assert.True(t, IsErrFilterTooComplex(err))
	})

	t.Run("preprocessing expansion is bounded", func(t *testing.T) {
		filter := "title = value" + strings.Repeat("'", 9000) + "suffix"
		require.Less(t, len(filter), maxFilterBytes)
		require.Greater(t, len(preprocessFilterString(filter)), maxFilterBytes)

		_, err := getTaskFiltersFromFilterString(filter, "UTC")
		require.Error(t, err)
		assert.True(t, IsErrFilterTooComplex(err))

		_, err = timeEntryFilterCond(filter, "UTC")
		require.Error(t, err)
		assert.True(t, IsErrFilterTooComplex(err))
	})

	t.Run("parens inside quotes do not count toward depth", func(t *testing.T) {
		quoted := "'" + strings.Repeat("(", 200) + "'"
		filters, err := getTaskFiltersFromFilterString("title = "+quoted+" && done = false", "UTC")
		require.NoError(t, err)
		require.Len(t, filters, 2)
	})

	t.Run("escaped quotes keep the quoted run intact", func(t *testing.T) {
		_, err := getTaskFiltersFromFilterString("title = 'a \\'' "+strings.Repeat("( ", 150)+"done = false"+strings.Repeat(" )", 150), "UTC")
		require.Error(t, err)
		assert.True(t, IsErrFilterTooComplex(err), "depth after a correctly terminated escaped-quote run must still be tracked")
	})

	t.Run("unmatched closing parenthesis is rejected", func(t *testing.T) {
		_, err := getTaskFiltersFromFilterString("done = true)", "UTC")
		require.Error(t, err)
		assert.True(t, IsErrInvalidFilterExpression(err))
	})
}

func TestErrFilterTooComplexHTTPErrorCode(t *testing.T) {
	httpErr := (&ErrFilterTooComplex{Reason: "test"}).HTTPError()
	assert.Equal(t, 4033, httpErr.Code)
	assert.NotEqual(t, ErrCodeInvalidTaskRepeatInterval, httpErr.Code)
}

func TestIsErrInvalidFilterMatchesParseErrors(t *testing.T) {
	_, err := getTaskFiltersFromFilterString("done = true &&", "UTC")
	require.Error(t, err)
	assert.True(t, IsErrInvalidFilterExpression(err), "the checker must match what the parse paths actually return")
	assert.True(t, isErrInvalidFilter(err))

	assert.True(t, IsErrFilterTooComplex(&ErrFilterTooComplex{Reason: "test"}))
	assert.True(t, isErrInvalidFilter(&ErrFilterTooComplex{Reason: "test"}))
}

func TestMatchTasksToViewsOfFilterSkipsUnparsableFilter(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	filter := &SavedFilter{
		ID:      1,
		OwnerID: 1,
		Filters: &TaskCollection{Filter: "done = "},
	}
	view := &ProjectView{ID: 999}
	task := &Task{ID: 1, ProjectID: 1}
	accessByProject := map[int64]map[int64]bool{1: {1: true}}

	matched, err := matchTasksToViewsOfFilter(s, []*Task{task}, filter, []*ProjectView{view}, accessByProject, "UTC")
	require.NoError(t, err, "an unparsable saved filter must be skipped, not fail the fetch")
	assert.Empty(t, matched)
}
