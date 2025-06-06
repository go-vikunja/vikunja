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
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestParseFilter(t *testing.T) {
	t.Run("boolean filter true", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("done = true", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "done", result[0].field)
		assert.Equal(t, taskFilterComparatorEquals, result[0].comparator)
		assert.Equal(t, true, result[0].value)
	})
	t.Run("boolean filter false", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("done = false", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "done", result[0].field)
		assert.Equal(t, taskFilterComparatorEquals, result[0].comparator)
		assert.Equal(t, false, result[0].value)
	})
	t.Run("numeric one", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("project_id = 1", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "project_id", result[0].field)
		assert.Equal(t, int64(1), result[0].value)
	})
	t.Run("numeric long", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("project_id = 4234", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "project_id", result[0].field)
		assert.Equal(t, int64(4234), result[0].value)
	})
	t.Run("in", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("project_id in 1,2,3", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "project_id", result[0].field)
		assert.Equal(t, taskFilterComparatorIn, result[0].comparator)
		require.Len(t, result[0].value, 3)
		assert.Equal(t, int64(1), result[0].value.([]interface{})[0])
		assert.Equal(t, int64(2), result[0].value.([]interface{})[1])
		assert.Equal(t, int64(3), result[0].value.([]interface{})[2])
	})
	t.Run("not in", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("project_id not in 1,2,3", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "project_id", result[0].field)
		assert.Equal(t, taskFilterComparatorNotIn, result[0].comparator)
		require.Len(t, result[0].value, 3)
		assert.Equal(t, int64(1), result[0].value.([]interface{})[0])
		assert.Equal(t, int64(2), result[0].value.([]interface{})[1])
		assert.Equal(t, int64(3), result[0].value.([]interface{})[2])
	})
	t.Run("use project for project_id", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("project in 1,2,3", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "project_id", result[0].field)
		require.Len(t, result[0].value, 3)
		assert.Equal(t, int64(1), result[0].value.([]interface{})[0])
		assert.Equal(t, int64(2), result[0].value.([]interface{})[1])
		assert.Equal(t, int64(3), result[0].value.([]interface{})[2])
	})
	t.Run("project in with spaces", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("project in 1, 2, 3", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "project_id", result[0].field)
		require.Len(t, result[0].value, 3)
		assert.Equal(t, int64(1), result[0].value.([]interface{})[0])
		assert.Equal(t, int64(2), result[0].value.([]interface{})[1])
		assert.Equal(t, int64(3), result[0].value.([]interface{})[2])
	})
	t.Run("date math strings", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("due_date < now+30d", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "due_date", result[0].field)
		in30Days := time.Now().Add(time.Hour * 24 * 30)
		assert.Equal(t, in30Days.Unix(), result[0].value.(time.Time).Unix())
	})
	t.Run("date math strings with quotes", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("due_date < 'now+30d'", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "due_date", result[0].field)
		in30Days := time.Now().Add(time.Hour * 24 * 30)
		assert.Equal(t, in30Days.Unix(), result[0].value.(time.Time).Unix())
	})
	t.Run("string values with single quotes", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("title = 'foo bar'", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "title", result[0].field)
		assert.Equal(t, "foo bar", result[0].value)
	})
	t.Run("string values with double quotes", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString(`title = "foo bar"`, "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "title", result[0].field)
		assert.Equal(t, "foo bar", result[0].value)
	})
	t.Run("string values without quotes", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("title = foo bar", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "title", result[0].field)
		assert.Equal(t, "foo bar", result[0].value)
	})
	t.Run("string values with single quote in them", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("title = foo's bar", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "title", result[0].field)
		assert.Equal(t, "foo's bar", result[0].value)
	})
	t.Run("string values with souble quote in them", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString(`title = foo"s bar`, "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "title", result[0].field)
		assert.Equal(t, `foo"s bar`, result[0].value)
	})
	t.Run("like query", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("title like foo bar", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "title", result[0].field)
		assert.Equal(t, "foo bar", result[0].value)
	})
	t.Run("invalid field", func(t *testing.T) {
		_, err := getTaskFiltersFromFilterString("invalid = foo", "UTC")

		require.Error(t, err)
	})
	t.Run("multiple filters with AND", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("done = false && priority > 3", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "done", result[0].field)
		assert.Equal(t, false, result[0].value)
		assert.Equal(t, "priority", result[1].field)
		assert.Equal(t, taskFilterComparatorGreater, result[1].comparator)
		assert.Equal(t, int64(3), result[1].value)
	})
	t.Run("multiple filters with OR", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("due_date < now || percent_done = 100", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "due_date", result[0].field)
		assert.Equal(t, taskFilterComparatorLess, result[0].comparator)
		assert.Equal(t, "percent_done", result[1].field)
		assert.InEpsilon(t, float64(100), result[1].value, 0)
	})
	t.Run("complex query with parentheses", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("(priority >= 4 && due_date < now+7d) || (done = false && assignees in 'John,Jane')", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Len(t, result[0].value, 2)
		require.Len(t, result[1].value, 2)

		firstSet := result[0].value.([]*taskFilter)
		assert.Equal(t, "priority", firstSet[0].field)
		assert.Equal(t, taskFilterComparatorGreateEquals, firstSet[0].comparator)
		assert.Equal(t, "due_date", firstSet[1].field)
		assert.Equal(t, taskFilterComparatorLess, firstSet[1].comparator)

		secondSet := result[1].value.([]*taskFilter)
		assert.Equal(t, "done", secondSet[0].field)
		assert.Equal(t, taskFilterComparatorEquals, secondSet[0].comparator)
		assert.Equal(t, "assignees", secondSet[1].field)
		assert.Equal(t, taskFilterComparatorIn, secondSet[1].comparator)
	})
	t.Run("not equals comparator", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("priority != 3", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "priority", result[0].field)
		assert.Equal(t, taskFilterComparatorNotEquals, result[0].comparator)
		assert.Equal(t, int64(3), result[0].value)
	})
	t.Run("less than or equal comparator", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("percent_done <= 50", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "percent_done", result[0].field)
		assert.Equal(t, taskFilterComparatorLessEquals, result[0].comparator)
		assert.InEpsilon(t, float64(50), result[0].value, 0)
	})
	t.Run("date field with exact date", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("start_date = 2023-06-15", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "start_date", result[0].field)
		assert.Equal(t, taskFilterComparatorEquals, result[0].comparator)
		expectedDate, err := time.ParseInLocation("2006-01-02", "2023-06-15", time.UTC)
		require.NoError(t, err)
		resultTime := result[0].value.(time.Time)
		assert.Equal(t, expectedDate.Format(time.RFC3339), resultTime.Format(time.RFC3339))
	})
	t.Run("in query with multiple values", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("priority in 1,3,5", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "priority", result[0].field)
		assert.Equal(t, taskFilterComparatorIn, result[0].comparator)
		require.Len(t, result[0].value, 3)
		assert.Equal(t, int64(1), result[0].value.([]interface{})[0])
		assert.Equal(t, int64(3), result[0].value.([]interface{})[1])
		assert.Equal(t, int64(5), result[0].value.([]interface{})[2])
	})
	t.Run("done_at field with relative time", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("done_at > now-7d", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "done_at", result[0].field)
		assert.Equal(t, taskFilterComparatorGreater, result[0].comparator)
		sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
		assert.Equal(t, sevenDaysAgo.Unix(), result[0].value.(time.Time).Unix())
	})
	t.Run("date filter with 0000-01-01", func(t *testing.T) {
		result, err := getTaskFiltersFromFilterString("due_date > 0000-01-01", "UTC")

		require.NoError(t, err)
		require.Len(t, result, 1)
		date := result[0].value.(time.Time)
		if db.GetDialect() == builder.MYSQL {
			assert.Equal(t, 1, date.Year())
		} else {
			assert.Equal(t, 0, date.Year())
		}
	})
}
