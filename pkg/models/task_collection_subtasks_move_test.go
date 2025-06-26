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

func TestTaskCollection_SubtaskRemainsAfterMove(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	c := &TaskCollection{
		ProjectID: 1,
		Expand:    []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	res, _, _, err := c.ReadAll(s, u, "", 0, 50)
	require.NoError(t, err)
	tasks, ok := res.([]*Task)
	require.True(t, ok)

	found := false
	for _, tsk := range tasks {
		if tsk.ID == 29 {
			found = true
			break
		}
	}
	assert.True(t, found, "subtask should be returned before moving")

	subtask := &Task{ID: 29, ProjectID: 7}
	err = subtask.Update(s, u)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	s2 := db.NewSession()
	defer s2.Close()
	c = &TaskCollection{
		ProjectID: 7,
		Expand:    []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	res, _, _, err = c.ReadAll(s2, u, "", 0, 50)
	require.NoError(t, err)
	tasks, ok = res.([]*Task)
	require.True(t, ok)

	found = false
	for _, tsk := range tasks {
		if tsk.ID == 29 {
			found = true
			break
		}
	}
	assert.True(t, found, "subtask should be returned after moving to another project")
}
