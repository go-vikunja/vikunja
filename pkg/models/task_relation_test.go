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

func TestTaskRelation_Create(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  2,
			RelationKind: RelationKindSubtask,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertExists(t, "task_relations", map[string]interface{}{
			"task_id":       1,
			"other_task_id": 2,
			"relation_kind": RelationKindSubtask,
			"created_by_id": 1,
		}, false)
	})
	t.Run("Two Tasks In Different Projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  13,
			RelationKind: RelationKindSubtask,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertExists(t, "task_relations", map[string]interface{}{
			"task_id":       1,
			"other_task_id": 13,
			"relation_kind": RelationKindSubtask,
			"created_by_id": 1,
		}, false)
	})
	t.Run("Already Existing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  29,
			RelationKind: RelationKindSubtask,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrRelationAlreadyExists(err))
	})
	t.Run("Same Task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:      1,
			OtherTaskID: 1,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrRelationTasksCannotBeTheSame(err))
	})
	t.Run("cycle with one subtask", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       29,
			OtherTaskID:  1,
			RelationKind: RelationKindSubtask,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskRelationCycle(err))
	})
	t.Run("cycle with multiple subtasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel1 := TaskRelation{
			TaskID:       1,
			OtherTaskID:  2,
			RelationKind: RelationKindSubtask,
		}
		err := rel1.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel2 := TaskRelation{
			TaskID:       2,
			OtherTaskID:  3,
			RelationKind: RelationKindSubtask,
		}
		err = rel2.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel3 := TaskRelation{
			TaskID:       3,
			OtherTaskID:  4,
			RelationKind: RelationKindSubtask,
		}
		err = rel3.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Cycle happens here
		rel4 := TaskRelation{
			TaskID:       4,
			OtherTaskID:  2,
			RelationKind: RelationKindSubtask,
		}
		err = rel4.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskRelationCycle(err))
	})
	t.Run("cycle with multiple subtasks tasks and relation back to parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel1 := TaskRelation{
			TaskID:       1,
			OtherTaskID:  2,
			RelationKind: RelationKindSubtask,
		}
		err := rel1.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel2 := TaskRelation{
			TaskID:       2,
			OtherTaskID:  3,
			RelationKind: RelationKindSubtask,
		}
		err = rel2.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel3 := TaskRelation{
			TaskID:       3,
			OtherTaskID:  4,
			RelationKind: RelationKindSubtask,
		}
		err = rel3.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Cycle happens here
		rel4 := TaskRelation{
			TaskID:       4,
			OtherTaskID:  1,
			RelationKind: RelationKindSubtask,
		}
		err = rel4.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskRelationCycle(err))
	})
	t.Run("cycle with one parenttask", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  29,
			RelationKind: RelationKindParenttask,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskRelationCycle(err))
	})
	t.Run("cycle with multiple parenttasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel1 := TaskRelation{
			TaskID:       1,
			OtherTaskID:  2,
			RelationKind: RelationKindParenttask,
		}
		err := rel1.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel2 := TaskRelation{
			TaskID:       2,
			OtherTaskID:  3,
			RelationKind: RelationKindParenttask,
		}
		err = rel2.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel3 := TaskRelation{
			TaskID:       3,
			OtherTaskID:  4,
			RelationKind: RelationKindParenttask,
		}
		err = rel3.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Cycle happens here
		rel4 := TaskRelation{
			TaskID:       4,
			OtherTaskID:  2,
			RelationKind: RelationKindParenttask,
		}
		err = rel4.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskRelationCycle(err))
	})
	t.Run("cycle with multiple parenttasks and relation back to parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel1 := TaskRelation{
			TaskID:       1,
			OtherTaskID:  2,
			RelationKind: RelationKindParenttask,
		}
		err := rel1.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel2 := TaskRelation{
			TaskID:       2,
			OtherTaskID:  3,
			RelationKind: RelationKindParenttask,
		}
		err = rel2.Create(s, &user.User{ID: 1})
		require.NoError(t, err)
		rel3 := TaskRelation{
			TaskID:       3,
			OtherTaskID:  4,
			RelationKind: RelationKindParenttask,
		}
		err = rel3.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Cycle happens here
		rel4 := TaskRelation{
			TaskID:       4,
			OtherTaskID:  1,
			RelationKind: RelationKindParenttask,
		}
		err = rel4.Create(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskRelationCycle(err))
	})
}

func TestTaskRelation_Delete(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("Normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  29,
			RelationKind: RelationKindSubtask,
		}
		err := rel.Delete(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertMissing(t, "task_relations", map[string]interface{}{
			"task_id":       1,
			"other_task_id": 29,
			"relation_kind": RelationKindSubtask,
		})
		db.AssertMissing(t, "task_relations", map[string]interface{}{
			"task_id":       29,
			"other_task_id": 1,
			"relation_kind": RelationKindParenttask,
		})
	})
	t.Run("Not existing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       9999,
			OtherTaskID:  3,
			RelationKind: RelationKindSubtask,
		}
		err := rel.Delete(s, u)
		require.Error(t, err)
		assert.True(t, IsErrRelationDoesNotExist(err))
	})
}

func TestTaskRelation_CanCreate(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  2,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("Two tasks on different projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  32,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("No update permissions on base task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       14,
			OtherTaskID:  1,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("No update permissions on base task, but read permissions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       15,
			OtherTaskID:  1,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("No read permissions on other task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  14,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
	t.Run("Nonexisting base task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       999999,
			OtherTaskID:  1,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
		assert.False(t, can)
	})
	t.Run("Nonexisting other task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  999999,
			RelationKind: RelationKindSubtask,
		}
		can, err := rel.CanCreate(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
		assert.False(t, can)
	})
}

// TestTaskBlockingEnforcement tests that blocked tasks cannot be completed if their blockers are not done
func TestTaskBlockingEnforcement(t *testing.T) {
	t.Run("Task with no blockers can be marked complete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := Task{ID: 1, Done: false}
		err := task.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		task.Done = true
		err = task.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		updated := Task{ID: 1}
		err = updated.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, updated.Done)
	})

	t.Run("Task with incomplete blocker cannot be marked complete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create blocking relation: task 3 blocks task 1
		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  3,
			RelationKind: RelationKindBlocked,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Try to mark task 1 as complete while task 3 is not done
		task := Task{ID: 1, Done: true}
		err = task.Update(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskIsBlocked(err))
	})

	t.Run("Task with complete blocker can be marked complete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create blocking relation: task 3 blocks task 1
		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  3,
			RelationKind: RelationKindBlocked,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Mark task 3 as complete (the blocker)
		blocker := Task{ID: 3}
		err = blocker.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		blocker.Done = true
		err = blocker.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Now task 1 should be able to be marked complete
		task := Task{ID: 1}
		err = task.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		task.Done = true
		err = task.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		updated := Task{ID: 1}
		err = updated.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, updated.Done)
	})

	t.Run("Multiple blockers - can only complete when all are done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create two blocking relations: task 3 and task 4 both block task 1
		rel1 := TaskRelation{
			TaskID:       1,
			OtherTaskID:  3,
			RelationKind: RelationKindBlocked,
		}
		err := rel1.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		rel2 := TaskRelation{
			TaskID:       1,
			OtherTaskID:  4,
			RelationKind: RelationKindBlocked,
		}
		err = rel2.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Mark task 3 as complete (one blocker done)
		blocker1 := Task{ID: 3}
		err = blocker1.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		blocker1.Done = true
		err = blocker1.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Try to mark task 1 as complete - should still fail because task 4 is not done
		task := Task{ID: 1, Done: true}
		err = task.Update(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskIsBlocked(err))

		// Mark task 4 as complete (second blocker done)
		blocker2 := Task{ID: 4}
		err = blocker2.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		blocker2.Done = true
		err = blocker2.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Now task 1 should be able to be marked complete
		task = Task{ID: 1}
		err = task.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		task.Done = true
		err = task.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		updated := Task{ID: 1}
		err = updated.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.True(t, updated.Done)
	})

	t.Run("Error includes blocking task information", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create blocking relation
		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  3,
			RelationKind: RelationKindBlocked,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Try to mark task 1 as complete
		task := Task{ID: 1, Done: true}
		err = task.Update(s, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrTaskIsBlocked(err))

		// Error should contain the full list of blocking tasks
		taskErr := err.(ErrTaskIsBlocked)
		require.Len(t, taskErr.BlockingTasks, 1)
		assert.Equal(t, int64(3), taskErr.BlockingTasks[0].ID)
	})

	t.Run("Unmarking as done when blocked does not fail", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create blocking relation: task 3 blocks task 1
		rel := TaskRelation{
			TaskID:       1,
			OtherTaskID:  3,
			RelationKind: RelationKindBlocked,
		}
		err := rel.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Task 1 is not done, try to unmark - should succeed
		task := Task{ID: 1}
		err = task.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, task.Done)

		task.Done = false
		err = task.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Verify still not done
		updated := Task{ID: 1}
		err = updated.ReadOne(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.False(t, updated.Done)
	})

}
