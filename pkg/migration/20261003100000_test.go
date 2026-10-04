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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRisks20261003100000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	drop := func() {
		require.NoError(t, x.DropTables(Risk20261003100000{}, RiskStatusHistory20261003100000{}))
	}
	t.Cleanup(drop)
	drop()

	require.NoError(t, addRisks20261003100000(x))

	// The defaults of a minimal row.
	_, err = x.Insert(&Risk20261003100000{ProjectID: 1, Title: "A risk", Probability: 3, Impact: 3, Status: "open", CreatedByID: 1})
	require.NoError(t, err)
	got := &Risk20261003100000{}
	has, err := x.Where("title = ?", "A risk").Get(got)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, "open", got.Status)

	_, err = x.Insert(&RiskStatusHistory20261003100000{RiskID: got.ID, ToStatus: "open", ChangedByID: 1})
	require.NoError(t, err)

	// The columns lists filter on are indexed.
	tables, err := x.DBMetas()
	require.NoError(t, err)
	indexed := map[string]bool{}
	for _, table := range tables {
		if table.Name != "risks" {
			continue
		}
		for _, index := range table.Indexes {
			for _, col := range index.Cols {
				indexed[col] = true
			}
		}
	}
	for _, col := range []string{"project_id", "owner_id", "status", "due_date"} {
		assert.Truef(t, indexed[col], "risks.%s should be indexed", col)
	}
}
