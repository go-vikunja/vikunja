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
	"xorm.io/xorm"
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

func TestTaskRelation_SubtaskOrder(t *testing.T) {
	u := &user.User{ID: 1}
	subtaskIDs := []int64{12, 11, 10, 9, 8}

	// Relations are created in an order that does not match the subtask ids, so an
	// accidental sort by task id can't pass these.
	seed := func(t *testing.T, s *xorm.Session) {
		for _, subtaskID := range subtaskIDs {
			rel := TaskRelation{
				TaskID:       2,
				OtherTaskID:  subtaskID,
				RelationKind: RelationKindSubtask,
			}
			require.NoError(t, rel.Create(s, u))
		}
	}

	readSubtaskIDs := func(t *testing.T, s *xorm.Session, view *ProjectView) (ids []int64) {
		taskMap := map[int64]*Task{
			2: {ID: 2, RelatedTasks: map[RelationKind][]*Task{}},
		}
		require.NoError(t, addRelatedTasksToTasks(s, []int64{2}, taskMap, u, view))
		for _, subtask := range taskMap[2].RelatedTasks[RelationKindSubtask] {
			ids = append(ids, subtask.ID)
		}
		return ids
	}

	t.Run("sorted by their position in the view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		seed(t, s)

		// Deliberately the reverse of the relation creation order
		for i, subtaskID := range subtaskIDs {
			_, err := s.Insert(&TaskPosition{
				TaskID:        subtaskID,
				ProjectViewID: 1,
				Position:      float64(len(subtaskIDs)-i) * 100,
			})
			require.NoError(t, err)
		}

		view, err := GetProjectViewByID(s, 1)
		require.NoError(t, err)

		reversed := make([]int64, 0, len(subtaskIDs))
		for i := len(subtaskIDs) - 1; i >= 0; i-- {
			reversed = append(reversed, subtaskIDs[i])
		}
		assert.Equal(t, reversed, readSubtaskIDs(t, s, view))
	})

	t.Run("falls back to relation order without a view", func(t *testing.T) {
		// The task detail page, CalDAV and exports read tasks without a view, so there
		// are no positions to sort by. Relation id keeps that stable.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		seed(t, s)

		assert.Equal(t, subtaskIDs, readSubtaskIDs(t, s, nil))
	})

	t.Run("sorts subtasks without a position in the view last", func(t *testing.T) {
		// Subtasks living in another project have no position row for this view.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		seed(t, s)

		_, err := s.Insert(&TaskPosition{TaskID: 9, ProjectViewID: 1, Position: 100})
		require.NoError(t, err)

		view, err := GetProjectViewByID(s, 1)
		require.NoError(t, err)

		// 9 is positioned so it comes first, the rest keep their relation order
		assert.Equal(t, []int64{9, 12, 11, 10, 8}, readSubtaskIDs(t, s, view))
	})
}
