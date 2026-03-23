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

func TestAddRelatedTasksToTasks_FiltersInaccessibleProjects(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	// Task 1 is in project 1 (owned by user 1).
	// The fixture adds a "related" relation from task 1 -> task 41.
	// Task 41 is in project 36 (owned by user 15, not shared with user 1).
	// User 1 should NOT see task 41 in the related tasks.

	taskMap := map[int64]*Task{
		1: {
			ID:           1,
			ProjectID:    1,
			RelatedTasks: make(RelatedTaskMap),
		},
	}
	taskIDs := []int64{1}

	err := addRelatedTasksToTasks(s, taskIDs, taskMap, u)
	require.NoError(t, err)

	// Task 29 is in project 1 (same project, user 1 has access) — should be present
	foundTask29 := false
	// Task 41 is in project 36 (user 1 has no access) — must NOT be present
	foundTask41 := false

	for _, relatedTasks := range taskMap[1].RelatedTasks {
		for _, rt := range relatedTasks {
			if rt.ID == 29 {
				foundTask29 = true
			}
			if rt.ID == 41 {
				foundTask41 = true
			}
		}
	}

	assert.True(t, foundTask29, "Task 29 (same project) should be visible as a related task")
	assert.False(t, foundTask41, "Task 41 (different project, no access) should NOT be visible as a related task")
}
