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
	"code.vikunja.io/api/pkg/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsCountFromDatabase verifies that each metric key counts the right table
// straight from the database. This guards the count key -> table name mapping; the
// caching/expiry/invalidation behaviour itself is covered by the keyvalue RememberFor
// tests.
func TestMetricsCountFromDatabase(t *testing.T) {
	// Projects and tasks exclude trashed fixtures.
	cases := map[string]struct {
		table    string
		expected int64
	}{
		metrics.UserCountKey: {
			table:    "users",
			expected: 25,
		},
		metrics.ProjectCountKey: {
			table:    "projects",
			expected: 44,
		},
		metrics.TaskCountKey: {
			table:    "tasks",
			expected: 53,
		},
		metrics.TeamCountKey: {
			table:    "teams",
			expected: 13,
		},
		metrics.FilesCountKey: {
			table:    "files",
			expected: 1,
		},
		metrics.AttachmentsCountKey: {
			table:    "task_attachments",
			expected: 4,
		},
	}

	db.LoadAndAssertFixtures(t)

	for key, tc := range cases {
		t.Run(tc.table, func(t *testing.T) {
			// Drop any value cached by a previous test so we recompute from the DB.
			require.NoError(t, metrics.InvalidateCount(key))

			count, err := metrics.GetCount(key)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, count)
		})
	}
}
