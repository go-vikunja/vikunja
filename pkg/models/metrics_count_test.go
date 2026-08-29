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
	cases := map[string]string{
		metrics.UserCountKey:        "users",
		metrics.ProjectCountKey:     "projects",
		metrics.TaskCountKey:        "tasks",
		metrics.TeamCountKey:        "teams",
		metrics.FilesCountKey:       "files",
		metrics.AttachmentsCountKey: "task_attachments",
	}

	db.LoadAndAssertFixtures(t)

	s := db.NewSession()
	defer s.Close()

	for key, table := range cases {
		t.Run(table, func(t *testing.T) {
			// Drop any value cached by a previous test so we recompute from the DB.
			require.NoError(t, metrics.InvalidateCount(key))

			query := s.Table(table)
			if key == metrics.TaskCountKey {
				query = query.Where("deleted_at IS NULL")
			}
			expected, err := query.Count()
			require.NoError(t, err)

			count, err := metrics.GetCount(key)
			require.NoError(t, err)
			assert.Equal(t, expected, count)
			assert.Positive(t, count, "fixtures should contain at least one %s", table)
		})
	}
}
