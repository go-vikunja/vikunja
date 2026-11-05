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
	"xorm.io/builder"
	"xorm.io/xorm"
)

func TestTaskAssignees_ArchivedProject(t *testing.T) {
	u := &user.User{ID: 1}
	cases := []struct {
		name        string
		taskID      int64
		assigneeIDs []int64
		write       func(*xorm.Session) error
	}{
		{
			name:   "delete",
			taskID: 30,
			assigneeIDs: []int64{
				1,
				2,
			},
			write: func(s *xorm.Session) error {
				return (&TaskAssginee{
					TaskID: 30,
					UserID: 1,
				}).Delete(s, u)
			},
		},
		{
			name:   "add",
			taskID: 1,
			write: func(s *xorm.Session) error {
				return (&TaskAssginee{
					TaskID: 1,
					UserID: 1,
				}).Create(s, u)
			},
		},
		{
			name:   "bulk replace",
			taskID: 30,
			assigneeIDs: []int64{
				1,
				2,
			},
			write: func(s *xorm.Session) error {
				return (&BulkAssignees{TaskID: 30}).Create(s, u)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			_, err := s.ID(1).Cols("is_archived").Update(&Project{IsArchived: true})
			require.NoError(t, err)

			err = tc.write(s)
			require.ErrorIs(t, err, ErrProjectIsArchived{ProjectID: 1})
			require.NoError(t, s.Commit())

			db.AssertCount(t, "task_assignees", builder.Eq{"task_id": tc.taskID}, int64(len(tc.assigneeIDs)))
			for _, assigneeID := range tc.assigneeIDs {
				db.AssertExists(t, "task_assignees", map[string]interface{}{
					"task_id": tc.taskID,
					"user_id": assigneeID,
				}, false)
			}
		})
	}
}
