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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

type stringBoundsQuoter struct {
	boundingString string
}

func (q stringBoundsQuoter) Quote(s string) string {
	return q.boundingString + s + q.boundingString
}

func TestUrgencyScoreQuery(t *testing.T) {
	t.Parallel()
	const testQuoteBounds = "@"
	for _, tc := range []struct {
		dbType      schemas.DBType
		expectQuery string
		expectErr   string
	}{
		{
			dbType: schemas.POSTGRES,
			expectQuery: `
(
	COALESCE(least(1, (1 << (greatest(0, (- cast(extract(epoch from @tasks.due_date@ - localtimestamp) as int)/86400 - -8) / 2))) / 38.05), 0) * 0.476 +
	COALESCE(@tasks.percent_done@, 0) * 0.476 +
	COALESCE(@tasks.priority@ / 5.0, 0) * 0.048
) as urgency
			`,
		},
		{
			dbType: schemas.MYSQL,
			expectQuery: `
(
	COALESCE(least(1, (1 << (greatest(0, (- cast(unix_timestamp(@tasks.due_date@ - now()) as signed)/86400 - -8) / 2))) / 38.05), 0) * 0.476 +
	COALESCE(@tasks.percent_done@, 0) * 0.476 +
	COALESCE(@tasks.priority@ / 5.0, 0) * 0.048
) as urgency
			`,
		},
		{
			dbType: schemas.SQLITE,
			expectQuery: `
(
	COALESCE(min(1, (1 << (max(0, (- cast(unixepoch(@tasks.due_date@) - unixepoch('now') as int)/86400 - -8) / 2))) / 38.05), 0) * 0.476 +
	COALESCE(@tasks.percent_done@, 0) * 0.476 +
	COALESCE(@tasks.priority@ / 5.0, 0) * 0.048
) as urgency
			`,
		},
		{
			dbType:    "not a real DB",
			expectErr: "unsupported database type: not a real DB",
		},
	} {
		t.Run(string(tc.dbType), func(t *testing.T) {
			t.Parallel()
			query, err := urgencyScoreQuery(defaultWeights(), stringBoundsQuoter{boundingString: testQuoteBounds}, tc.dbType)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tc.expectQuery), strings.TrimSpace(query))
		})
	}
}
