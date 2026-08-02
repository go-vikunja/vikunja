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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectPermissions_PermissionUpdate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Project owner (User 1)
	owner := &user.User{ID: 1, Username: "user1"}
	// Shared user (User 2)
	sharedUser := &user.User{ID: 2, Username: "user2"}

	// Create project owned by User 1
	project := &Project{
		Title: "Project for PermissionUpdate Test",
	}
	err := project.Create(s, owner)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	// Share project with User 2 with PermissionUpdate (1)
	s = db.NewSession()
	defer s.Close()
	pu := &ProjectUser{
		ProjectID:  project.ID,
		Username:   sharedUser.Username,
		Permission: PermissionUpdate,
	}
	err = pu.Create(s, owner)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	t.Run("user with PermissionUpdate can read project", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		canRead, maxPerm, err := project.CanRead(s, sharedUser)
		require.NoError(t, err)
		assert.True(t, canRead, "User with PermissionUpdate should be able to read project")
		assert.Equal(t, int(PermissionUpdate), maxPerm, "Max permission should be PermissionUpdate")
	})

	t.Run("user with PermissionUpdate cannot write or admin project", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		canWrite, err := project.CanWrite(s, sharedUser)
		require.NoError(t, err)
		assert.False(t, canWrite, "User with PermissionUpdate cannot write project metadata")

		isAdmin, err := project.IsAdmin(s, sharedUser)
		require.NoError(t, err)
		assert.False(t, isAdmin, "User with PermissionUpdate is not admin")
	})

	t.Run("user with PermissionUpdate can toggle task completion", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		// Create a task in the project by owner
		task := &Task{
			ProjectID: project.ID,
			Title:     "Task for Completion Test",
		}
		err := task.Create(s, owner)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s = db.NewSession()
		defer s.Close()

		// User 2 toggles completion
		now := time.Now()
		taskUpdate := &Task{
			ID:        task.ID,
			ProjectID: project.ID,
			Done:      true,
			DoneAt:    now,
		}

		canUpdate, err := taskUpdate.CanUpdate(s, sharedUser)
		require.NoError(t, err)
		assert.True(t, canUpdate, "User with PermissionUpdate can toggle task completion")
	})

	t.Run("user with PermissionUpdate cannot edit task title", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		// Create a task in the project by owner
		task := &Task{
			ProjectID: project.ID,
			Title:     "Original Task Title",
		}
		err := task.Create(s, owner)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s = db.NewSession()
		defer s.Close()

		// User 2 attempts to change title
		taskUpdate := &Task{
			ID:        task.ID,
			ProjectID: project.ID,
			Title:     "Modified Task Title",
		}

		canUpdate, err := taskUpdate.CanUpdate(s, sharedUser)
		assert.False(t, canUpdate, "User with PermissionUpdate cannot edit task title")
		assert.ErrorIs(t, err, ErrGenericForbidden{})
	})

	t.Run("user with PermissionUpdate can create task comments", func(t *testing.T) {
		s = db.NewSession()
		defer s.Close()

		task := &Task{
			ProjectID: project.ID,
			Title:     "Task for Comment Test",
		}
		err := task.Create(s, owner)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s = db.NewSession()
		defer s.Close()

		comment := &TaskComment{
			TaskID:  task.ID,
			Comment: "Comment by PermissionUpdate user",
		}
		canComment, err := comment.CanCreate(s, sharedUser)
		require.NoError(t, err)
		assert.True(t, canComment, "User with PermissionUpdate can post task comments")
	})
}
