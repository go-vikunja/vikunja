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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

func TestUrgencyProperty_NormalizedPropertyScore(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		description string
		property    UrgencyProperty
		filter      *TaskCollection
		dbType      schemas.DBType
		expectQuery string
		expectErr   string
	}{
		{
			description: "invalid property",
			property:    0,
			expectErr:   "unrecognized urgency score property: <err: invalid urgency property enum value: 0>",
		},
		{
			description: "due date",
			property:    UrgencyDueDate,
			dbType:      schemas.SQLITE,
			// Full coverage in [TestUrgencyScoreQuery]
			expectQuery: `min(1, (1 << (max(0, (- cast(unixepoch("tasks.due_date") - unixepoch('now') as int)/86400 - -8) / 2))) / 38.05)`,
		},
		{
			description: "matches filter",
			property:    UrgencyMatchesFilter,
			filter:      &TaskCollection{Filter: `done = false`},
			expectQuery: "CASE WHEN (tasks.`done`=false) THEN 1 ELSE 0 END",
		},
		{
			description: "invalid filter string",
			property:    UrgencyMatchesFilter,
			filter:      &TaskCollection{Filter: "very broken"},
			expectErr:   `could not parse filter string "very broken": Task filter expression 'very broken' is invalid [ExpressionError: expected a sign operator, got "broken" (identifier)]`,
		},
		{
			description: "percent done",
			property:    UrgencyPercentDone,
			expectQuery: `"tasks.percent_done"`,
		},
		{
			description: "priority",
			property:    UrgencyPriority,
			expectQuery: `"tasks.priority" / 5.0`,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			quoter := stringBoundsQuoter{boundingString: `"`}
			query, err := tc.property.normalizedPropertyScore(tc.filter, quoter, tc.dbType)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectQuery, query)
		})
	}
}
