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
	"bytes"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestTaskChildWrites_TaskMovedToForbiddenProject(t *testing.T) {
	u := &user.User{ID: 1}

	cases := []struct {
		name   string
		can    func(s *xorm.Session, a web.Auth) (bool, error)
		write  func(s *xorm.Session) error
		assert func(t *testing.T)
	}{
		{
			name: "comment create",
			can:  (&TaskComment{TaskID: 1}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&TaskComment{
					TaskID:  1,
					Comment: "raced",
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "task_comments", map[string]interface{}{"comment": "raced"})
			},
		},
		{
			name: "comment update",
			can: (&TaskComment{
				ID:     1,
				TaskID: 1,
			}).CanUpdate,
			write: func(s *xorm.Session) error {
				return (&TaskComment{
					ID:      1,
					TaskID:  1,
					Comment: "raced",
				}).Update(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "task_comments", map[string]interface{}{
					"id":      1,
					"comment": "raced",
				})
			},
		},
		{
			name: "comment delete",
			can: (&TaskComment{
				ID:     1,
				TaskID: 1,
			}).CanDelete,
			write: func(s *xorm.Session) error {
				return (&TaskComment{
					ID:     1,
					TaskID: 1,
				}).Delete(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertExists(t, "task_comments", map[string]interface{}{"id": 1}, false)
			},
		},
		{
			name: "attachment create",
			can:  (&TaskAttachment{TaskID: 1}).CanCreate,
			write: func(s *xorm.Session) error {
				content := []byte("raced")
				return (&TaskAttachment{TaskID: 1}).NewAttachment(s, bytes.NewReader(content), "raced", uint64(len(content)), u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "files", map[string]interface{}{"name": "raced"})
			},
		},
		{
			name: "attachment delete",
			can: (&TaskAttachment{
				ID:     1,
				TaskID: 1,
			}).CanDelete,
			write: func(s *xorm.Session) error {
				return (&TaskAttachment{
					ID:     1,
					TaskID: 1,
				}).Delete(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertExists(t, "task_attachments", map[string]interface{}{"id": 1}, false)
			},
		},
		{
			name: "label add",
			can: (&LabelTask{
				TaskID:  1,
				LabelID: 1,
			}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&LabelTask{
					TaskID:  1,
					LabelID: 1,
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "label_tasks", map[string]interface{}{
					"task_id":  1,
					"label_id": 1,
				})
			},
		},
		{
			name: "label remove",
			can: (&LabelTask{
				TaskID:  1,
				LabelID: 4,
			}).CanDelete,
			write: func(s *xorm.Session) error {
				return (&LabelTask{
					TaskID:  1,
					LabelID: 4,
				}).Delete(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertExists(t, "label_tasks", map[string]interface{}{
					"task_id":  1,
					"label_id": 4,
				}, false)
			},
		},
		{
			name: "label bulk",
			can:  (&LabelTaskBulk{TaskID: 1}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&LabelTaskBulk{
					TaskID: 1,
					Labels: []*Label{{ID: 1}},
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertExists(t, "label_tasks", map[string]interface{}{
					"task_id":  1,
					"label_id": 4,
				}, false)
			},
		},
		{
			name: "assignee add",
			can: (&TaskAssginee{
				TaskID: 1,
				UserID: 1,
			}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&TaskAssginee{
					TaskID: 1,
					UserID: 1,
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": 1})
			},
		},
		{
			name: "assignee bulk",
			can:  (&BulkAssignees{TaskID: 1}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&BulkAssignees{
					TaskID:    1,
					Assignees: []*user.User{{ID: 1}},
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "task_assignees", map[string]interface{}{"task_id": 1})
			},
		},
		{
			name: "relation create",
			can: (&TaskRelation{
				TaskID:       1,
				OtherTaskID:  2,
				RelationKind: RelationKindRelated,
			}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&TaskRelation{
					TaskID:       1,
					OtherTaskID:  2,
					RelationKind: RelationKindRelated,
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "task_relations", map[string]interface{}{
					"task_id":       1,
					"other_task_id": 2,
				})
			},
		},
		{
			name: "relation delete",
			can: (&TaskRelation{
				TaskID:       1,
				OtherTaskID:  29,
				RelationKind: RelationKindSubtask,
			}).CanDelete,
			write: func(s *xorm.Session) error {
				return (&TaskRelation{
					TaskID:       1,
					OtherTaskID:  29,
					RelationKind: RelationKindSubtask,
				}).Delete(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertExists(t, "task_relations", map[string]interface{}{"id": 1}, false)
			},
		},
		{
			name: "reaction create",
			can: (&Reaction{
				EntityID:         1,
				EntityKindString: "tasks",
				Value:            "🚀",
			}).CanCreate,
			write: func(s *xorm.Session) error {
				return (&Reaction{
					EntityID:         1,
					EntityKindString: "tasks",
					Value:            "🚀",
				}).Create(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertMissing(t, "reactions", map[string]interface{}{
					"entity_id": 1,
					"value":     "🚀",
				})
			},
		},
		{
			name: "reaction delete",
			can: (&Reaction{
				EntityID:         1,
				EntityKindString: "tasks",
				Value:            "👋",
			}).CanDelete,
			write: func(s *xorm.Session) error {
				return (&Reaction{
					EntityID:         1,
					EntityKindString: "tasks",
					EntityKind:       ReactionKindTask,
					Value:            "👋",
				}).Delete(s, u)
			},
			assert: func(t *testing.T) {
				db.AssertExists(t, "reactions", map[string]interface{}{"id": 1}, false)
			},
		},
		{
			name:  "duplicate",
			can:   (&TaskDuplicate{TaskID: 1}).CanCreate,
			write: func(s *xorm.Session) error { return (&TaskDuplicate{TaskID: 1}).Create(s, u) },
			assert: func(t *testing.T) {
				db.AssertMissing(t, "task_relations", map[string]interface{}{
					"other_task_id": 1,
					"relation_kind": RelationKindCopiedFrom,
				})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			files.InitTestFileFixtures(t)
			s := db.NewSession()
			defer s.Close()

			can, err := tc.can(s, u)
			require.NoError(t, err)
			require.True(t, can)

			_, err = s.ID(1).Cols("project_id", "index").Update(&Task{
				ProjectID: 2,
				Index:     9999,
			})
			require.NoError(t, err)

			err = tc.write(s)
			require.ErrorIs(t, err, ErrGenericForbidden{})
			require.NoError(t, s.Commit())

			tc.assert(t)
		})
	}
}

// User 1 only reads project 3, which holds task 32.
func grantUser1WriteOnProject3(t *testing.T, s *xorm.Session) {
	t.Helper()
	_, err := s.Where("user_id = ? AND project_id = ?", 1, 3).Cols("permission").Update(&ProjectUser{Permission: PermissionWrite})
	require.NoError(t, err)
}
