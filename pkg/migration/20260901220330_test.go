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

	"github.com/stretchr/testify/require"
)

func TestAddTaskIndexAliases20260901220330(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, x.DropTables(TaskIndexAlias20260901220330{}))
	})
	require.NoError(t, x.DropTables(TaskIndexAlias20260901220330{}))

	require.NoError(t, addTaskIndexAliases20260901220330(x))

	_, err = x.Insert(&TaskIndexAlias20260901220330{ProjectID: 1, Index: 3, TaskID: 1})
	require.NoError(t, err)

	_, err = x.Insert(&TaskIndexAlias20260901220330{ProjectID: 1, Index: 3, TaskID: 2})
	require.Error(t, err)
}
