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

// TestTaskCollection_ReadAll has been moved to pkg/services/task_test.go as TestTaskService_GetAllByProjectWithDetails
// This was done to follow the refactoring guide: complex integration tests belong in the services layer,
// while model tests should be simple unit tests.

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

func TestTaskSearchWithExpandSubtasks(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	project, err := GetProjectSimpleByID(s, 36)
	require.NoError(t, err)

	opts := &taskSearchOptions{
		search: "Caldav",
		expand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	tasks, _, _, err := getRawTasksForProjects(s, []*Project{project}, &user.User{ID: 15}, opts)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)
}

func TestTaskCollection_SubtaskWithMultipleParentsNoDuplicates(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 15}

	// Use existing tasks from fixtures:
	// - Task 41: Parent task in project 36 (already exists)
	// - Task 42: Another parent task in project 36 (already exists)
	// - Task 43: Subtask in project 36 (already a subtask of task 41)

	// Add a second parent relationship: task 43 -> task 42
	// This will make task 43 have multiple parents (task 41 and task 42)
	relation := &TaskRelation{
		TaskID:       43, // subtask
		OtherTaskID:  42, // second parent
		RelationKind: RelationKindParenttask,
		CreatedByID:  15,
	}
	_, err := s.Insert(relation)
	require.NoError(t, err)

	// Create inverse relation: task 42 -> task 43
	inverseRelation := &TaskRelation{
		TaskID:       42, // second parent
		OtherTaskID:  43, // subtask
		RelationKind: RelationKindSubtask,
		CreatedByID:  15,
	}
	_, err = s.Insert(inverseRelation)
	require.NoError(t, err)

	// Test Project 36 - should include tasks 41, 42, and 43, but task 43 should only appear once
	c := &TaskCollection{
		ProjectID: 36,
		Expand:    []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	res, _, _, err := c.ReadAll(s, u, "", 0, 50)
	require.NoError(t, err)
	tasks, ok := res.([]*Task)
	require.True(t, ok)

	// Count how many times task 43 (the subtask) appears
	subtaskCount := 0
	for _, task := range tasks {
		if task.ID == 43 {
			subtaskCount++
		}
	}

	// The subtask should appear exactly once (as a subtask, not as a standalone task)
	assert.Equal(t, 1, subtaskCount, "Subtask should appear exactly once in Project 36")

	// Verify that both parent tasks are present
	foundParent1 := false
	foundParent2 := false
	for _, task := range tasks {
		if task.ID == 41 {
			foundParent1 = true
		}
		if task.ID == 42 {
			foundParent2 = true
		}
	}
	assert.True(t, foundParent1, "Parent task 41 should be present")
	assert.True(t, foundParent2, "Parent task 42 should be present")
}
