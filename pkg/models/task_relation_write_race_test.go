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

	"github.com/stretchr/testify/require"
)

func TestTaskRelation_CreateOtherTaskMovedToForbiddenProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}
	rel := &TaskRelation{
		TaskID:       1,
		OtherTaskID:  2,
		RelationKind: RelationKindRelated,
	}
	can, err := rel.CanCreate(s, u)
	require.NoError(t, err)
	require.True(t, can)

	_, err = s.ID(2).Cols("project_id", "index").Update(&Task{
		ProjectID: 2,
		Index:     9999,
	})
	require.NoError(t, err)

	err = rel.Create(s, u)
	require.ErrorIs(t, err, ErrGenericForbidden{})
	require.NoError(t, s.Commit())

	db.AssertMissing(t, "task_relations", map[string]interface{}{
		"task_id":       1,
		"other_task_id": 2,
		"relation_kind": RelationKindRelated,
	})
	db.AssertMissing(t, "task_relations", map[string]interface{}{
		"task_id":       2,
		"other_task_id": 1,
		"relation_kind": RelationKindRelated,
	})
}

func TestTaskRelation_DeleteOtherTaskMovedToForbiddenProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}
	rel := &TaskRelation{
		TaskID:       1,
		OtherTaskID:  29,
		RelationKind: RelationKindSubtask,
	}
	can, err := rel.CanDelete(s, u)
	require.NoError(t, err)
	require.True(t, can)

	_, err = s.ID(29).Cols("project_id", "index").Update(&Task{
		ProjectID: 2,
		Index:     9999,
	})
	require.NoError(t, err)

	err = rel.Delete(s, u)
	require.ErrorIs(t, err, ErrGenericForbidden{})
	require.NoError(t, s.Commit())

	db.AssertExists(t, "task_relations", map[string]interface{}{
		"id":            1,
		"task_id":       1,
		"other_task_id": 29,
		"relation_kind": RelationKindSubtask,
	}, false)
	db.AssertExists(t, "task_relations", map[string]interface{}{
		"id":            2,
		"task_id":       29,
		"other_task_id": 1,
		"relation_kind": RelationKindParenttask,
	}, false)
}

func TestTaskRelation_ReadOnlyArchivedOtherTask(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(3).Cols("is_archived").Update(&Project{IsArchived: true})
	require.NoError(t, err)

	u := &user.User{ID: 1}
	rel := &TaskRelation{
		TaskID:       33,
		OtherTaskID:  32,
		RelationKind: RelationKindRelated,
	}
	can, err := rel.CanCreate(s, u)
	require.NoError(t, err)
	require.True(t, can)
	require.NoError(t, rel.Create(s, u))
	require.NoError(t, s.Commit())

	db.AssertExists(t, "task_relations", map[string]interface{}{
		"task_id":       33,
		"other_task_id": 32,
		"relation_kind": RelationKindRelated,
	}, false)
	db.AssertExists(t, "task_relations", map[string]interface{}{
		"task_id":       32,
		"other_task_id": 33,
		"relation_kind": RelationKindRelated,
	}, false)

	can, err = rel.CanDelete(s, u)
	require.NoError(t, err)
	require.True(t, can)
	require.NoError(t, rel.Delete(s, u))
	require.NoError(t, s.Commit())

	db.AssertMissing(t, "task_relations", map[string]interface{}{
		"task_id":       33,
		"other_task_id": 32,
		"relation_kind": RelationKindRelated,
	})
	db.AssertMissing(t, "task_relations", map[string]interface{}{
		"task_id":       32,
		"other_task_id": 33,
		"relation_kind": RelationKindRelated,
	})
}
