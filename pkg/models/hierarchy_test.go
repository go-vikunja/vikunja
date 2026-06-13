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

func TestGetDescendantProjectIDs(t *testing.T) {
	t.Run("project with no children", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 1 has no children
		descendants, err := getDescendantProjectIDs(s, 1)
		require.NoError(t, err)
		assert.Empty(t, descendants)
	})

	t.Run("project with direct children only", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 29 has children 14 and 19
		descendants, err := getDescendantProjectIDs(s, 29)
		require.NoError(t, err)
		assert.Len(t, descendants, 2)
		assert.Contains(t, descendants, int64(14))
		assert.Contains(t, descendants, int64(19))
	})

	t.Run("project with nested descendants", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Hierarchy: 27 -> 12 -> 25 -> 26
		descendants, err := getDescendantProjectIDs(s, 27)
		require.NoError(t, err)
		assert.Len(t, descendants, 3)
		assert.Contains(t, descendants, int64(12))
		assert.Contains(t, descendants, int64(25))
		assert.Contains(t, descendants, int64(26))
	})

	t.Run("mid-level project returns only its descendants", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 12 has children 25 -> 26
		descendants, err := getDescendantProjectIDs(s, 12)
		require.NoError(t, err)
		assert.Len(t, descendants, 2)
		assert.Contains(t, descendants, int64(25))
		assert.Contains(t, descendants, int64(26))
		// Should NOT contain 27 (parent) or 12 itself
		assert.NotContains(t, descendants, int64(27))
		assert.NotContains(t, descendants, int64(12))
	})
}

func TestGetRootProjectID(t *testing.T) {
	t.Run("root project returns itself", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 27 is a root (no parent)
		rootID, err := getRootProjectID(s, 27)
		require.NoError(t, err)
		assert.Equal(t, int64(27), rootID)
	})

	t.Run("direct child returns parent as root", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 12 has parent 27
		rootID, err := getRootProjectID(s, 12)
		require.NoError(t, err)
		assert.Equal(t, int64(27), rootID)
	})

	t.Run("deeply nested project returns root", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Hierarchy: 27 -> 12 -> 25 -> 26
		// Project 26 should return 27 as root
		rootID, err := getRootProjectID(s, 26)
		require.NoError(t, err)
		assert.Equal(t, int64(27), rootID)
	})

	t.Run("mid-level project returns correct root", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 25 has parent 12, grandparent 27
		rootID, err := getRootProjectID(s, 25)
		require.NoError(t, err)
		assert.Equal(t, int64(27), rootID)
	})
}

func TestGetRootProjectViewID(t *testing.T) {
	t.Run("root project view returns itself", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Get a view from a root project
		view := &ProjectView{}
		exists, err := s.Where("project_id = ?", 1).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		rootViewID, err := getRootProjectViewID(s, view.ID)
		require.NoError(t, err)
		assert.Equal(t, view.ID, rootViewID)
	})

	t.Run("child project view returns matching root view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a hierarchy with views
		// Project 12 has parent 27, both should have views
		childView := &ProjectView{}
		exists, err := s.Where("project_id = ?", 12).Get(childView)
		require.NoError(t, err)
		require.True(t, exists, "child project should have a view")

		parentView := &ProjectView{}
		exists, err = s.Where("project_id = ? AND view_kind = ?", 27, childView.ViewKind).Get(parentView)
		require.NoError(t, err)
		require.True(t, exists, "parent project should have matching view kind")

		rootViewID, err := getRootProjectViewID(s, childView.ID)
		require.NoError(t, err)
		assert.Equal(t, parentView.ID, rootViewID)
	})
}

func TestHierarchicalTaskPositions(t *testing.T) {
	t.Run("position stored at root level when updating in sub-project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		usr := &user.User{ID: 6}

		// Create a parent project
		parentProject := &Project{Title: "Parent"}
		parentProject.OwnerID = usr.ID
		err := parentProject.Create(s, usr)
		require.NoError(t, err)

		// Create a child project
		childProject := &Project{Title: "Child", ParentProjectID: parentProject.ID}
		childProject.OwnerID = usr.ID
		err = childProject.Create(s, usr)
		require.NoError(t, err)

		// Create a task in the child project
		task := &Task{Title: "Test Task", ProjectID: childProject.ID}
		err = task.Create(s, usr)
		require.NoError(t, err)

		// Get the child project's list view
		childView := &ProjectView{}
		_, err = s.Where("project_id = ? AND view_kind = ?", childProject.ID, ProjectViewKindList).Get(childView)
		require.NoError(t, err)

		// Get the parent project's list view
		parentView := &ProjectView{}
		_, err = s.Where("project_id = ? AND view_kind = ?", parentProject.ID, ProjectViewKindList).Get(parentView)
		require.NoError(t, err)

		// Update task position using child view
		tp := &TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: childView.ID,
			Position:      1000,
		}
		err = tp.Update(s, usr)
		require.NoError(t, err)

		// Verify position is stored at parent (root) view level
		storedPosition := &TaskPosition{}
		exists, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, parentView.ID).Get(storedPosition)
		require.NoError(t, err)
		assert.True(t, exists, "position should be stored at root view level")
		assert.InDelta(t, float64(1000), storedPosition.Position, 0.001)

		// Verify position is NOT stored at child view level
		childPosition := &TaskPosition{}
		exists, err = s.Where("task_id = ? AND project_view_id = ?", task.ID, childView.ID).Get(childPosition)
		require.NoError(t, err)
		assert.False(t, exists, "position should NOT be stored at child view level")
	})

	t.Run("reordering in parent view affects sub-project task ordering", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		usr := &user.User{ID: 6}

		// Create hierarchy
		parentProject := &Project{Title: "Parent"}
		parentProject.OwnerID = usr.ID
		err := parentProject.Create(s, usr)
		require.NoError(t, err)

		childProject := &Project{Title: "Child", ParentProjectID: parentProject.ID}
		childProject.OwnerID = usr.ID
		err = childProject.Create(s, usr)
		require.NoError(t, err)

		// Create task in child project
		task := &Task{Title: "Child Task", ProjectID: childProject.ID}
		err = task.Create(s, usr)
		require.NoError(t, err)

		// Get parent view
		parentView := &ProjectView{}
		_, err = s.Where("project_id = ? AND view_kind = ?", parentProject.ID, ProjectViewKindList).Get(parentView)
		require.NoError(t, err)

		// Update position using parent view
		tp := &TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: parentView.ID,
			Position:      500,
		}
		err = tp.Update(s, usr)
		require.NoError(t, err)

		// Get child view
		childView := &ProjectView{}
		_, err = s.Where("project_id = ? AND view_kind = ?", childProject.ID, ProjectViewKindList).Get(childView)
		require.NoError(t, err)

		// When retrieving via child view, should get same position (from root)
		rootViewID, err := getRootProjectViewID(s, childView.ID)
		require.NoError(t, err)
		assert.Equal(t, parentView.ID, rootViewID, "child view should resolve to parent view")

		// Verify the position is accessible
		storedPosition := &TaskPosition{}
		exists, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, rootViewID).Get(storedPosition)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.InDelta(t, float64(500), storedPosition.Position, 0.001)
	})
}

func TestParentProjectShowsSubProjectTasks(t *testing.T) {
	t.Run("parent project view includes tasks from child projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		usr := &user.User{ID: 6}

		// Create hierarchy
		parentProject := &Project{Title: "Parent"}
		parentProject.OwnerID = usr.ID
		err := parentProject.Create(s, usr)
		require.NoError(t, err)

		childProject := &Project{Title: "Child", ParentProjectID: parentProject.ID}
		childProject.OwnerID = usr.ID
		err = childProject.Create(s, usr)
		require.NoError(t, err)

		// Create tasks
		parentTask := &Task{Title: "Parent Task", ProjectID: parentProject.ID}
		err = parentTask.Create(s, usr)
		require.NoError(t, err)

		childTask := &Task{Title: "Child Task", ProjectID: childProject.ID}
		err = childTask.Create(s, usr)
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// Start new session for reading
		s2 := db.NewSession()
		defer s2.Close()

		// Fetch tasks via parent project's task collection
		tc := &TaskCollection{ProjectID: parentProject.ID}
		result, _, _, err := tc.ReadAll(s2, usr, "", 0, -1)
		require.NoError(t, err)

		tasks := result.([]*Task)
		assert.Len(t, tasks, 2, "parent view should show both parent and child tasks")

		taskIDs := make([]int64, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}
		assert.Contains(t, taskIDs, parentTask.ID)
		assert.Contains(t, taskIDs, childTask.ID)
	})

	t.Run("parent project includes tasks from deeply nested projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		usr := &user.User{ID: 6}

		// Create 3-level hierarchy
		rootProject := &Project{Title: "Root"}
		rootProject.OwnerID = usr.ID
		err := rootProject.Create(s, usr)
		require.NoError(t, err)

		midProject := &Project{Title: "Mid", ParentProjectID: rootProject.ID}
		midProject.OwnerID = usr.ID
		err = midProject.Create(s, usr)
		require.NoError(t, err)

		leafProject := &Project{Title: "Leaf", ParentProjectID: midProject.ID}
		leafProject.OwnerID = usr.ID
		err = leafProject.Create(s, usr)
		require.NoError(t, err)

		// Create tasks at each level
		rootTask := &Task{Title: "Root Task", ProjectID: rootProject.ID}
		err = rootTask.Create(s, usr)
		require.NoError(t, err)

		midTask := &Task{Title: "Mid Task", ProjectID: midProject.ID}
		err = midTask.Create(s, usr)
		require.NoError(t, err)

		leafTask := &Task{Title: "Leaf Task", ProjectID: leafProject.ID}
		err = leafTask.Create(s, usr)
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// Fetch tasks via root project
		s2 := db.NewSession()
		defer s2.Close()

		tc := &TaskCollection{ProjectID: rootProject.ID}
		result, _, _, err := tc.ReadAll(s2, usr, "", 0, -1)
		require.NoError(t, err)

		tasks := result.([]*Task)
		assert.Len(t, tasks, 3, "root view should show tasks from all levels")

		taskIDs := make([]int64, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}
		assert.Contains(t, taskIDs, rootTask.ID)
		assert.Contains(t, taskIDs, midTask.ID)
		assert.Contains(t, taskIDs, leafTask.ID)
	})
}
