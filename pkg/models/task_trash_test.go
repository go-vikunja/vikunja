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
	"maps"
	"slices"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func trashTaskForTest(t *testing.T, taskID int64, a web.Auth) {
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, (&Task{ID: taskID}).Delete(s, a))
	require.NoError(t, s.Commit())
}

func setTaskColumnsForTest(t *testing.T, taskID int64, cols map[string]any) {
	s := db.NewAutocommitSession()
	defer s.Close()
	_, err := s.Table("tasks").Where("id = ?", taskID).Update(cols)
	require.NoError(t, err)
}

func restoreTaskForTest(t *testing.T, taskID int64, a web.Auth) (*Task, error) {
	s := db.NewSession()
	defer s.Close()

	task := &Task{ID: taskID}
	can, err := task.CanRestore(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}
	if err := task.Restore(s, a); err != nil {
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return task, nil
}

func deletePermanentlyForTest(t *testing.T, taskID int64, a web.Auth) error {
	s := db.NewSession()
	defer s.Close()
	task := &Task{ID: taskID}
	can, err := task.CanDeletePermanently(s, a)
	if err != nil {
		return err
	}
	if !can {
		return ErrGenericForbidden{}
	}
	require.NoError(t, task.DeletePermanently(s, a))
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return nil
}

func purgedTaskDoers(t *testing.T) map[int64]*user.User {
	t.Helper()
	doers := map[int64]*user.User{}
	for _, e := range events.GetDispatchedEvents((&TaskPurgedEvent{}).Name()) {
		purged, ok := e.(*TaskPurgedEvent)
		require.True(t, ok)
		doers[purged.Task.ID] = purged.Doer
	}
	return doers
}

func TestTask_Restore(t *testing.T) {
	u1 := &user.User{ID: 1}

	t.Run("restores into the default bucket of every view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		task, err := restoreTaskForTest(t, 51, u1)
		require.NoError(t, err)
		events.AssertDispatched(t, &TaskRestoredEvent{})

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":            51,
			"deleted_at":    nil,
			"deleted_by_id": nil,
		}, false)
		assert.True(t, task.Updated.After(time.Date(2018, 12, 1, 1, 12, 4, 0, time.UTC)))

		// View 4 is project 1's manual kanban with default bucket 1
		db.AssertExists(t, "task_positions", map[string]interface{}{
			"task_id":         51,
			"project_view_id": 1,
		}, false)
		db.AssertExists(t, "task_positions", map[string]interface{}{
			"task_id":         51,
			"project_view_id": 4,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         51,
			"project_view_id": 4,
			"bucket_id":       1,
		}, false)

		s := db.NewSession()
		defer s.Close()
		read := &Task{ID: 51}
		require.NoError(t, read.ReadOne(s, u1))
	})
	t.Run("done task restores into the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 51, map[string]any{"done": true})

		_, err := restoreTaskForTest(t, 51, u1)
		require.NoError(t, err)

		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         51,
			"project_view_id": 4,
			"bucket_id":       3,
		}, false)
	})
	t.Run("round trip through delete keeps the deleter cleared", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		trashTaskForTest(t, 1, u1)

		_, err := restoreTaskForTest(t, 1, u1)
		require.NoError(t, err)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":            1,
			"deleted_at":    nil,
			"deleted_by_id": nil,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         1,
			"project_view_id": 4,
		}, false)
	})
	t.Run("write link share can restore in its project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		share := &LinkSharing{
			ID:         2,
			ProjectID:  2,
			Permission: PermissionWrite,
		}
		trashTaskForTest(t, 13, share)

		_, err := restoreTaskForTest(t, 13, share)
		require.NoError(t, err)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":         13,
			"deleted_at": nil,
		}, false)
	})
	t.Run("read-only link share cannot restore", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := restoreTaskForTest(t, 51, &LinkSharing{
			ID:         1,
			ProjectID:  1,
			Permission: PermissionRead,
		})
		require.ErrorIs(t, err, ErrGenericForbidden{})
	})
	t.Run("unrelated user cannot tell the task is trashed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := restoreTaskForTest(t, 51, &user.User{ID: 2})
		assert.True(t, IsErrTaskDoesNotExist(err))
		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id":         51,
			"deleted_at": nil,
		})
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 51}, false)
	})
	t.Run("reader cannot restore", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 32, map[string]any{"deleted_at": time.Now()})

		_, err := restoreTaskForTest(t, 32, u1)
		require.ErrorIs(t, err, ErrGenericForbidden{})
	})
	t.Run("refused in an archived project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 36, map[string]any{"deleted_at": time.Now()})

		_, err := restoreTaskForTest(t, 36, u1)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("second restore of the same loaded task is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		defer events.CleanupPending(s)

		first := &Task{ID: 51}
		second := &Task{ID: 51}
		_, err := first.CanRestore(s, u1)
		require.NoError(t, err)
		_, err = second.CanRestore(s, u1)
		require.NoError(t, err)

		require.NoError(t, first.Restore(s, u1))
		err = second.Restore(s, u1)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
	t.Run("reappears in saved filter views", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		view, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "done = false")
		require.NoError(t, s.Commit())
		_ = s.Close()
		events.ClearDispatchedEvents()

		_, err := restoreTaskForTest(t, 51, u1)
		require.NoError(t, err)
		dispatched := events.GetDispatchedEvents((&TaskRestoredEvent{}).Name())
		require.Len(t, dispatched, 1)
		inFilterView := map[string]interface{}{
			"task_id":         51,
			"project_view_id": view.ID,
		}
		db.AssertMissing(t, "task_buckets", inFilterView)

		events.TestListener(t, dispatched[0], &UpdateTaskInSavedFilterViews{})
		db.AssertExists(t, "task_buckets", inFilterView, false)
	})
	t.Run("live task is not in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := restoreTaskForTest(t, 1, u1)
		require.Error(t, err)
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
	t.Run("refused while its project is in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 53, map[string]any{"deleted_at": time.Now()})

		_, err := restoreTaskForTest(t, 53, u1)
		var inTrash ErrProjectIsInTrash
		require.ErrorAs(t, err, &inTrash)
		assert.Equal(t, ErrCodeProjectIsInTrash, inTrash.HTTPError().Code)
		assert.Equal(t, int64(45), inTrash.ID)
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 53, "deleted_at": nil})
	})
}

func TestTaskTrash_InstanceAdminInArchivedProject(t *testing.T) {
	setup := func(t *testing.T) *user.User {
		db.LoadAndAssertFixtures(t)
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		t.Cleanup(license.ResetForTests)
		s := db.NewAutocommitSession()
		defer s.Close()
		_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)
		setTaskColumnsForTest(t, 36, map[string]any{"deleted_at": time.Now()})
		return &user.User{
			ID:      2,
			IsAdmin: true,
		}
	}

	t.Run("restore is refused", func(t *testing.T) {
		admin := setup(t)

		_, err := restoreTaskForTest(t, 36, admin)
		assert.True(t, IsErrProjectIsArchived(err), "got %v", err)
	})
	t.Run("permanent delete is refused", func(t *testing.T) {
		admin := setup(t)

		err := deletePermanentlyForTest(t, 36, admin)
		assert.True(t, IsErrProjectIsArchived(err), "got %v", err)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 36}, false)
	})
}

func trashedTaskIDs(tasks []*TrashedTask) []int64 {
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func TestGetTrashedTasks(t *testing.T) {
	u1 := &user.User{ID: 1}

	setup := func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		trashTaskForTest(t, 1, u1)
		// Task 36 is in archived project 22, task 32 in project 3 where user 1 only reads
		setTaskColumnsForTest(t, 36, map[string]any{"deleted_at": time.Now()})
		setTaskColumnsForTest(t, 32, map[string]any{"deleted_at": time.Now()})
	}

	t.Run("writable live projects only, newest first", func(t *testing.T) {
		setup(t)
		s := db.NewSession()
		defer s.Close()

		tasks, total, err := GetTrashedTasks(s, u1, 1, 50)
		require.NoError(t, err)
		assert.Equal(t, []int64{
			1,
			51,
		}, trashedTaskIDs(tasks))
		assert.Equal(t, int64(2), total)

		require.NotNil(t, tasks[0].DeletedBy)
		assert.Equal(t, int64(1), tasks[0].DeletedBy.ID)
		assert.False(t, tasks[0].DeletedAt.IsZero())
		// Task 51 was deleted before deleted_by_id existed
		assert.Nil(t, tasks[1].DeletedBy)
		assert.NotEmpty(t, tasks[1].Identifier)
	})
	t.Run("paginated", func(t *testing.T) {
		setup(t)
		s := db.NewSession()
		defer s.Close()

		tasks, total, err := GetTrashedTasks(s, u1, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, []int64{51}, trashedTaskIDs(tasks))
		assert.Equal(t, int64(2), total)
	})
	t.Run("owner of another project sees its trash", func(t *testing.T) {
		setup(t)
		s := db.NewSession()
		defer s.Close()

		tasks, _, err := GetTrashedTasks(s, &user.User{ID: 3}, 1, 50)
		require.NoError(t, err)
		assert.Contains(t, trashedTaskIDs(tasks), int64(32))
		assert.NotContains(t, trashedTaskIDs(tasks), int64(51))
	})
	t.Run("link shares have no trash", func(t *testing.T) {
		setup(t)
		s := db.NewSession()
		defer s.Close()

		share := &LinkSharing{
			ID:         2,
			ProjectID:  2,
			Permission: PermissionWrite,
		}
		_, _, err := GetTrashedTasks(s, share, 1, 50)
		require.ErrorIs(t, err, ErrGenericForbidden{})
	})
}

func TestTask_DeletePermanently(t *testing.T) {
	t.Run("owner purges a trashed task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		require.NoError(t, deletePermanentlyForTest(t, 51, &user.User{ID: 1}))
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 51})

		doers := purgedTaskDoers(t)
		require.Len(t, doers, 1)
		require.NotNil(t, doers[51])
		assert.Equal(t, int64(1), doers[51].ID)
	})
	t.Run("admin shares purge", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		u1 := &user.User{ID: 1}
		// User 1 is admin of project 11 by user share and of project 8 by team share
		trashTaskForTest(t, 20, u1)
		trashTaskForTest(t, 17, u1)

		require.NoError(t, deletePermanentlyForTest(t, 20, u1))
		require.NoError(t, deletePermanentlyForTest(t, 17, u1))
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 20})
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 17})
	})
	t.Run("writer can restore but not purge", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		u1 := &user.User{ID: 1}
		// User 1 can write to project 10 by user share and to project 7 by team share
		for _, taskID := range []int64{
			19,
			16,
		} {
			trashTaskForTest(t, taskID, u1)
			require.ErrorIs(t, deletePermanentlyForTest(t, taskID, u1), ErrGenericForbidden{})
			db.AssertExists(t, "tasks", map[string]interface{}{"id": taskID}, false)

			_, err := restoreTaskForTest(t, taskID, u1)
			require.NoError(t, err)
		}
	})
	t.Run("unrelated user cannot tell the task is trashed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		assert.True(t, IsErrTaskDoesNotExist(deletePermanentlyForTest(t, 51, &user.User{ID: 2})))
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 51}, false)
	})
	t.Run("reader is denied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 32, map[string]any{"deleted_at": time.Now()})
		require.ErrorIs(t, deletePermanentlyForTest(t, 32, &user.User{ID: 1}), ErrGenericForbidden{})
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 32}, false)
	})
	t.Run("refused in an archived project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setTaskColumnsForTest(t, 36, map[string]any{"deleted_at": time.Now()})
		assert.True(t, IsErrProjectIsArchived(deletePermanentlyForTest(t, 36, &user.User{ID: 1})))
		db.AssertExists(t, "tasks", map[string]interface{}{"id": 36}, false)
	})
	t.Run("write link share is denied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		share := &LinkSharing{
			ID:         2,
			ProjectID:  2,
			Permission: PermissionWrite,
		}
		trashTaskForTest(t, 13, share)
		require.ErrorIs(t, deletePermanentlyForTest(t, 13, share), ErrGenericForbidden{})
	})
	t.Run("live task is not in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		err := deletePermanentlyForTest(t, 1, &user.User{ID: 1})
		assert.True(t, IsErrTaskDoesNotExist(err))
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":         1,
			"deleted_at": nil,
		}, false)
	})
}

func TestEmptyTaskTrash(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	u1 := &user.User{ID: 1}
	// Deleted by someone else, still purged
	trashTaskForTest(t, 1, u1)
	setTaskColumnsForTest(t, 1, map[string]any{"deleted_by_id": 2})
	setTaskColumnsForTest(t, 36, map[string]any{"deleted_at": time.Now()})
	setTaskColumnsForTest(t, 32, map[string]any{"deleted_at": time.Now()})
	trashTaskForTest(t, 19, u1)
	trashTaskForTest(t, 20, u1)

	events.ClearDispatchedEvents()

	s := db.NewSession()
	defer s.Close()
	require.NoError(t, EmptyTaskTrash(s, u1))
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)

	db.AssertMissing(t, "tasks", map[string]interface{}{"id": 1})
	db.AssertMissing(t, "tasks", map[string]interface{}{"id": 51})
	db.AssertMissing(t, "tasks", map[string]interface{}{"id": 20})
	// Archived, read-only and write-only projects are out of scope
	db.AssertExists(t, "tasks", map[string]interface{}{"id": 36}, false)
	db.AssertExists(t, "tasks", map[string]interface{}{"id": 32}, false)
	db.AssertExists(t, "tasks", map[string]interface{}{"id": 19}, false)

	doers := purgedTaskDoers(t)
	assert.ElementsMatch(t, []int64{
		1,
		20,
		51,
	}, slices.Collect(maps.Keys(doers)))
	for id, doer := range doers {
		require.NotNil(t, doer, "task %d", id)
		assert.Equal(t, int64(1), doer.ID, "task %d", id)
	}
}

func TestTask_CanRead_Trashed(t *testing.T) {
	t.Run("reader learns the task is in the trash", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		trashTaskForTest(t, 1, &user.User{ID: 1})
		s := db.NewSession()
		defer s.Close()

		_, _, err := (&Task{ID: 1}).CanRead(s, &user.User{ID: 1})
		var inTrash ErrTaskIsInTrash
		require.ErrorAs(t, err, &inTrash)
		assert.Equal(t, int64(1), inTrash.ProjectID)
		assert.False(t, inTrash.DeletedAt.IsZero())
		require.NotNil(t, inTrash.DeletedBy)
		assert.Equal(t, int64(1), inTrash.DeletedBy.ID)
	})
	t.Run("unknown deleter", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, _, err := (&Task{ID: 51}).CanRead(s, &user.User{ID: 1})
		var inTrash ErrTaskIsInTrash
		require.ErrorAs(t, err, &inTrash)
		assert.Nil(t, inTrash.DeletedBy)
	})
	t.Run("no access hides that the task exists", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, _, err := (&Task{ID: 51}).CanRead(s, &user.User{ID: 2})
		assert.True(t, IsErrTaskDoesNotExist(err))
	})
}

func TestTask_WritesOnTrashedTask(t *testing.T) {
	writes := []struct {
		name  string
		check func(s *xorm.Session, a web.Auth, taskID int64) (bool, error)
	}{
		{
			name: "update",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&Task{ID: taskID}).CanUpdate(s, a)
			},
		},
		{
			name: "delete",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&Task{ID: taskID}).CanDelete(s, a)
			},
		},
		{
			name: "comment",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&TaskComment{TaskID: taskID}).CanCreate(s, a)
			},
		},
		{
			name: "attachment",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&TaskAttachment{TaskID: taskID}).CanCreate(s, a)
			},
		},
		{
			name: "assignee",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&TaskAssginee{
					TaskID: taskID,
					UserID: 1,
				}).CanCreate(s, a)
			},
		},
		{
			name: "labels",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&LabelTaskBulk{TaskID: taskID}).CanCreate(s, a)
			},
		},
		{
			name: "relation",
			check: func(s *xorm.Session, a web.Auth, taskID int64) (bool, error) {
				return (&TaskRelation{
					TaskID:       taskID,
					OtherTaskID:  2,
					RelationKind: RelationKindRelated,
				}).CanCreate(s, a)
			},
		},
	}

	for _, w := range writes {
		t.Run(w.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			// User 1 only reads project 3
			setTaskColumnsForTest(t, 32, map[string]any{"deleted_at": time.Now()})
			s := db.NewSession()
			defer s.Close()

			_, err := w.check(s, &user.User{ID: 1}, 51)
			require.ErrorAs(t, err, &ErrTaskIsInTrash{}, "writer")
			_, err = w.check(s, &user.User{ID: 1}, 32)
			require.ErrorAs(t, err, &ErrTaskIsInTrash{}, "reader")
			_, err = w.check(s, &user.User{ID: 2}, 51)
			assert.True(t, IsErrTaskDoesNotExist(err), "stranger: %v", err)
		})
	}
}
