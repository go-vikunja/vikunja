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
	"context"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// Fixture projects 45 and 46 (child of 45) are in user 1's trash.

func trashProjectForTest(t *testing.T, projectID int64, a web.Auth) {
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, (&Project{ID: projectID}).Delete(s, a))
	require.NoError(t, s.Commit())
}

func restoreProjectForTest(t *testing.T, projectID int64, a web.Auth) error {
	s := db.NewSession()
	defer s.Close()
	project := &Project{ID: projectID}
	can, err := project.CanRestore(s, a)
	if err != nil {
		return err
	}
	if !can {
		return ErrGenericForbidden{}
	}
	if err := project.Restore(s, a); err != nil {
		return err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return nil
}

func TestProject_Restore(t *testing.T) {
	u1 := &user.User{ID: 1}

	t.Run("restores the project with its sub-projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		require.NoError(t, restoreProjectForTest(t, 45, u1))
		events.AssertDispatched(t, &ProjectRestoredEvent{})

		db.AssertExists(t, "projects", map[string]interface{}{"id": 45, "deleted_at": nil, "deleted_by_id": nil}, false)
		db.AssertExists(t, "projects", map[string]interface{}{"id": 46, "deleted_at": nil}, false)
		// Bumped for CalDAV
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 53, "updated": "2018-12-01 01:12:04"})

		s := db.NewSession()
		defer s.Close()
		can, _, err := (&Task{ID: 54}).CanRead(s, u1)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("brings back sub-projects trashed separately before", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		u6 := &user.User{ID: 6}
		// 27 -> 12 -> 25 -> 26
		trashProjectForTest(t, 26, u6)
		trashProjectForTest(t, 12, u6)

		require.NoError(t, restoreProjectForTest(t, 12, u6))
		for _, id := range []int64{12, 25, 26} {
			db.AssertExists(t, "projects", map[string]interface{}{"id": id, "deleted_at": nil}, false)
		}
	})
	t.Run("leaves separately trashed tasks in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		trashProjectForTest(t, 1, u1)
		require.NoError(t, restoreProjectForTest(t, 1, u1))
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 51, "deleted_at": nil})
	})
	t.Run("refused while the parent is in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		err := restoreProjectForTest(t, 46, u1)
		require.ErrorAs(t, err, &ErrParentProjectIsInTrash{})
	})
	t.Run("refused while the parent is archived", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// 21 is a child of archived 22
		s := db.NewSession()
		_, err := s.Table("projects").Where("id = ?", 21).Update(map[string]any{"deleted_at": time.Now()})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		err = restoreProjectForTest(t, 21, u1)
		assert.True(t, IsErrParentProjectIsArchived(err))
	})
	t.Run("admin share may restore, read share may not", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// User 1 is admin of project 11 and reads project 9, both owned by others
		trashProjectForTest(t, 11, &user.User{ID: 6})
		trashProjectForTest(t, 9, &user.User{ID: 6})

		require.NoError(t, restoreProjectForTest(t, 11, u1))
		require.ErrorIs(t, restoreProjectForTest(t, 9, u1), ErrGenericForbidden{})
	})
	t.Run("unrelated user is denied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		require.ErrorIs(t, restoreProjectForTest(t, 45, &user.User{ID: 2}), ErrGenericForbidden{})
	})
	t.Run("live project is not in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		assert.True(t, IsErrProjectDoesNotExist(restoreProjectForTest(t, 1, u1)))
	})
}

func TestProject_ReadTrashed(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, _, err := (&Project{ID: 45}).CanRead(s, &user.User{ID: 1})
	var inTrash ErrProjectIsInTrash
	require.ErrorAs(t, err, &inTrash)
	require.NotNil(t, inTrash.DeletedBy)
	assert.Equal(t, int64(1), inTrash.DeletedBy.ID)
	assert.True(t, inTrash.CanRestore)

	_, _, err = (&Project{ID: 45}).CanRead(s, &user.User{ID: 2})
	assert.True(t, IsErrProjectDoesNotExist(err))

	// A task in a trashed project points at the project
	_, _, err = (&Task{ID: 53}).CanRead(s, &user.User{ID: 1})
	require.ErrorAs(t, err, &inTrash)
}

func TestProject_TrashHidesContents(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u1 := &user.User{ID: 1}

	projects, _, _, err := (&Project{}).ReadAll(s, u1, "", 1, 500)
	require.NoError(t, err)
	for _, p := range projects.([]*Project) {
		assert.NotContains(t, []int64{45, 46}, p.ID)
	}

	tasks, _, _, err := (&TaskCollection{}).ReadAll(s, u1, "", 1, 500)
	require.NoError(t, err)
	for _, task := range tasks.([]*Task) {
		assert.NotContains(t, []int64{53, 54}, task.ID)
	}
}

func TestGetTrashedProjects(t *testing.T) {
	t.Run("topmost trashed projects with their sub-project count", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		projects, total, err := GetTrashedProjects(s, &user.User{ID: 1}, 1, 50)
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, int64(45), projects[0].ID)
		assert.Equal(t, int64(1), projects[0].TrashedSubProjects)
		require.NotNil(t, projects[0].DeletedBy)
		assert.Equal(t, int64(1), projects[0].DeletedBy.ID)
	})
	t.Run("separately trashed sub-projects fold into their ancestor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		u6 := &user.User{ID: 6}
		trashProjectForTest(t, 26, u6)
		trashProjectForTest(t, 12, u6)
		s := db.NewSession()
		defer s.Close()

		projects, _, err := GetTrashedProjects(s, u6, 1, 50)
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, int64(12), projects[0].ID)
		assert.Equal(t, int64(2), projects[0].TrashedSubProjects)
	})
	t.Run("other users see nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		projects, _, err := GetTrashedProjects(s, &user.User{ID: 2}, 1, 50)
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
	t.Run("link shares have no trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, _, err := GetTrashedProjects(s, &LinkSharing{ID: 3, ProjectID: 3, Permission: PermissionAdmin}, 1, 50)
		require.ErrorIs(t, err, ErrGenericForbidden{})
	})
}

func TestProject_DeletePermanently(t *testing.T) {
	purge := func(t *testing.T, projectID int64, a web.Auth) error {
		s := db.NewSession()
		defer s.Close()
		project := &Project{ID: projectID}
		can, err := project.CanDeletePermanently(s, a)
		if err != nil {
			return err
		}
		if !can {
			return ErrGenericForbidden{}
		}
		require.NoError(t, project.DeletePermanently(s))
		return s.Commit()
	}

	t.Run("admin purges a trashed project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		require.NoError(t, purge(t, 45, &user.User{ID: 1}))
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 45})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 46})
	})
	t.Run("unrelated user is denied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		require.ErrorIs(t, purge(t, 45, &user.User{ID: 2}), ErrGenericForbidden{})
		db.AssertExists(t, "projects", map[string]interface{}{"id": 45}, false)
	})
	t.Run("live project is not in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		assert.True(t, IsErrProjectDoesNotExist(purge(t, 1, &user.User{ID: 1})))
	})
}

func TestEmptyProjectTrash(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	purged, err := EmptyProjectTrash(s, &user.User{ID: 1})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	assert.Equal(t, 1, purged)
	db.AssertMissing(t, "projects", map[string]interface{}{"id": 45})
	db.AssertMissing(t, "projects", map[string]interface{}{"id": 46})
}

func TestDeleteExpiredProjects(t *testing.T) {
	trashedAt := time.Date(2018, 12, 3, 10, 0, 0, 0, time.UTC)

	t.Run("older than the retention period", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		deleteExpiredProjects(trashedAt.Add(TrashRetention + time.Hour))
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 45})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 46})
		db.AssertExists(t, "projects", map[string]interface{}{"id": 1}, false)
	})
	t.Run("within the retention period", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		deleteExpiredProjects(trashedAt.Add(TrashRetention - time.Hour))
		db.AssertExists(t, "projects", map[string]interface{}{"id": 45}, false)
	})
}

func TestGetProjectCount(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	tasks, err := s.Where("project_id = ?", 1).Count(&Task{})
	require.NoError(t, err)

	count, err := GetProjectCount(s, &user.User{ID: 1}, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count.SubProjects)
	assert.Equal(t, tasks, count.Tasks)
	// Entries 1, 2 and 4, all by user 1
	assert.Equal(t, int64(3), count.TimeEntries)
	assert.Equal(t, int64(1), count.TimeEntryAuthors)

	count, err = GetProjectCount(s, &user.User{ID: 6}, 27)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count.SubProjects)

	_, err = GetProjectCount(s, &user.User{ID: 2}, 1)
	require.ErrorIs(t, err, ErrGenericForbidden{})
}

func TestTimeEntry_TrashedProject(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	u3 := &user.User{ID: 3}

	result, _, _, err := (&TimeEntry{}).ReadAll(s, u3, "", 1, 50)
	require.NoError(t, err)
	inTrash := map[int64]bool{}
	for _, e := range result.([]*TimeEntry) {
		inTrash[e.ID] = e.InTrash
	}
	assert.True(t, inTrash[5])
	assert.True(t, inTrash[6])

	// The project owner doesn't see other people's entries in it anymore
	result, _, _, err = (&TimeEntry{}).ReadAll(s, &user.User{ID: 1}, "", 1, 50)
	require.NoError(t, err)
	for _, e := range result.([]*TimeEntry) {
		assert.NotContains(t, []int64{5, 6}, e.ID)
	}

	read := &TimeEntry{ID: 6}
	can, _, err := read.CanRead(s, u3)
	require.NoError(t, err)
	assert.True(t, can)
	assert.True(t, read.InTrash)

	can, err = (&TimeEntry{ID: 5}).CanDelete(s, u3)
	require.NoError(t, err)
	assert.False(t, can)

	db.AssertCount(t, "time_entries", builder.Eq{"project_id": 45}, 1)
}

func TestBuildOverview_SkipsTrashedProjects(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	live, err := s.Table("projects").Where("deleted_at IS NULL").Count()
	require.NoError(t, err)
	all, err := s.Table("projects").Count()
	require.NoError(t, err)
	require.Greater(t, all, live)

	overview, err := BuildOverview(s)
	require.NoError(t, err)
	assert.Equal(t, live, overview.Projects)
}

func TestCleanupAfterProjectAccessLoss_KeepsTrashedProjectsWithAccess(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, err := s.Insert(&TaskAssginee{TaskID: 54, UserID: 1})
	require.NoError(t, err)

	// User 1 still owns trashed project 45 and its child 46
	require.NoError(t, cleanupAfterProjectAccessLossForUsers(s, []int64{1}, []int64{45}))
	require.NoError(t, s.Commit())

	db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": 54, "user_id": 1}, false)
}
