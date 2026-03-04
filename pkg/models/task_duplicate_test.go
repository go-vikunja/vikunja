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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskDuplicate(t *testing.T) {
	t.Run("basic duplicate", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}

		td := &TaskDuplicate{
			TaskID: 1,
		}
		can, err := td.CanCreate(s, u)
		require.NoError(t, err)
		assert.True(t, can)
		err = td.Create(s, u)
		require.NoError(t, err)
		assert.NotEqual(t, int64(0), td.Task.ID)
		assert.NotEqual(t, int64(1), td.Task.ID)
		assert.Equal(t, "task #1", td.Task.Title)

		// Verify labels were copied
		labelCount, err := s.Where("task_id = ?", td.Task.ID).Count(&LabelTask{})
		require.NoError(t, err)
		originalLabelCount, err := s.Where("task_id = ?", int64(1)).Count(&LabelTask{})
		require.NoError(t, err)
		assert.Equal(t, originalLabelCount, labelCount)

		// Verify assignees were copied
		assigneeCount, err := s.Where("task_id = ?", td.Task.ID).Count(&TaskAssginee{})
		require.NoError(t, err)
		originalAssigneeCount, err := s.Where("task_id = ?", int64(1)).Count(&TaskAssginee{})
		require.NoError(t, err)
		assert.Equal(t, originalAssigneeCount, assigneeCount)

		// Verify a "copiedfrom" relation was created
		relationCount, err := s.
			Where("task_id = ? AND other_task_id = ? AND relation_kind = ?",
				td.Task.ID, int64(1), RelationKindCopiedFrom).
			Count(&TaskRelation{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), relationCount)
	})

	t.Run("no permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 2 does not have write access to task 1's project (project 1)
		u := &user.User{ID: 2}

		td := &TaskDuplicate{
			TaskID: 1,
		}
		can, err := td.CanCreate(s, u)
		require.NoError(t, err)
		assert.False(t, can)
	})
}
