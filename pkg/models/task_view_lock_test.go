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
	"strings"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"
	"xorm.io/xorm/schemas"
)

type viewLockReachedKey struct{}

type viewLockReached struct {
	viewID int64
	once   sync.Once
	ready  chan struct{}
}

type viewLockReachedHook struct{}

func (viewLockReachedHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if reached, ok := c.Ctx.Value(viewLockReachedKey{}).(*viewLockReached); ok &&
		strings.Contains(c.SQL, "project_views") && strings.Contains(c.SQL, "FOR UPDATE") &&
		len(c.Args) == 1 && c.Args[0] == reached.viewID {
		reached.once.Do(func() { close(reached.ready) })
	}
	return c.Ctx, nil
}

func (viewLockReachedHook) AfterProcess(c *contexts.ContextHook) error {
	return c.Err
}

func TestTaskWritesLockViewsBeforeTasks(t *testing.T) {
	if db.Type() != schemas.MYSQL && db.Type() != schemas.POSTGRES {
		t.Skip("Requires database row locks")
	}
	u := &user.User{ID: 1}
	cases := []struct {
		name       string
		taskID     int64
		viewID     int64
		nextViewID int64
		setup      func(*testing.T)
		write      func(*xorm.Session) error
		verify     func(*testing.T)
	}{
		{
			name:   "position",
			taskID: 1,
			viewID: 1,
			write: func(s *xorm.Session) error {
				return (&TaskPosition{
					TaskID:        1,
					ProjectViewID: 1,
					Position:      12345,
				}).Update(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "task_positions", map[string]interface{}{
					"task_id":         1,
					"project_view_id": 1,
					"position":        12345,
				}, false)
			},
		},
		{
			name:   "bucket",
			taskID: 1,
			viewID: 4,
			write: func(s *xorm.Session) error {
				return (&TaskBucket{
					TaskID:        1,
					ProjectViewID: 4,
					ProjectID:     1,
					BucketID:      3,
				}).Update(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "tasks", map[string]interface{}{
					"id":   1,
					"done": true,
				}, false)
			},
		},
		{
			name:   "done",
			taskID: 1,
			viewID: 1,
			write: func(s *xorm.Session) error {
				return (&Task{
					ID:    1,
					Title: "task #1",
					Done:  true,
				}).Update(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "tasks", map[string]interface{}{
					"id":   1,
					"done": true,
				}, false)
			},
		},
		{
			name:   "move",
			taskID: 1,
			viewID: 1,
			write: func(s *xorm.Session) error {
				return (&Task{
					ID:        1,
					Title:     "task #1",
					ProjectID: 2,
				}).Update(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "tasks", map[string]interface{}{
					"id":         1,
					"project_id": 2,
				}, false)
			},
		},
		{
			name:   "delete",
			taskID: 1,
			viewID: 1,
			write:  func(s *xorm.Session) error { return (&Task{ID: 1}).Delete(s, u) },
			verify: func(t *testing.T) {
				db.AssertMissing(t, "task_positions", map[string]interface{}{"task_id": 1})
			},
		},
		{
			name:   "duplicate",
			taskID: 1,
			viewID: 1,
			setup:  files.InitTestFileFixtures,
			write:  func(s *xorm.Session) error { return (&TaskDuplicate{TaskID: 1}).Create(s, u) },
			verify: func(t *testing.T) {
				db.AssertExists(t, "task_relations", map[string]interface{}{
					"other_task_id": 1,
					"relation_kind": RelationKindCopiedFrom,
				}, false)
			},
		},
		{
			name:   "create",
			viewID: 1,
			write: func(s *xorm.Session) error {
				return (&Task{
					Title:     "created without deadlock",
					ProjectID: 1,
				}).Create(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "tasks", map[string]interface{}{
					"title":      "created without deadlock",
					"project_id": 1,
				}, false)
			},
		},
		{
			name:   "saved filter bucket",
			taskID: 1,
			viewID: 9999,
			setup: func(t *testing.T) {
				s := db.NewSession()
				defer s.Close()
				view, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "project = 1")
				_, err := s.Insert(&Bucket{
					ID:            99999,
					ProjectViewID: view.ID,
					Title:         "done",
				})
				require.NoError(t, err)
				_, err = s.ID(view.ID).Cols("done_bucket_id").Update(&ProjectView{DoneBucketID: 99999})
				require.NoError(t, err)
				require.NoError(t, s.Commit())
			},
			write: func(s *xorm.Session) error {
				return (&TaskBucket{
					TaskID:        1,
					ProjectViewID: 9999,
					ProjectID:     getProjectIDFromSavedFilterID(9999),
					BucketID:      99999,
				}).Update(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "tasks", map[string]interface{}{
					"id":   1,
					"done": true,
				}, false)
				db.AssertExists(t, "task_buckets", map[string]interface{}{
					"task_id":         1,
					"project_view_id": 9999,
					"bucket_id":       99999,
				}, false)
			},
		},
		{
			name:       "project tree purge",
			taskID:     100000,
			viewID:     1,
			nextViewID: 100000,
			setup: func(t *testing.T) {
				s := db.NewSession()
				defer s.Close()
				insertTestProject(t, s, &Project{
					ID:      100000,
					OwnerID: 1,
					Title:   "parent",
				})
				_, err := s.ID(1).Cols("parent_project_id").Update(&Project{ParentProjectID: new(int64(100000))})
				require.NoError(t, err)
				require.NoError(t, RebuildProjectAncestors(s))
				_, err = s.Insert(&ProjectView{
					ID:        100000,
					ProjectID: 100000,
					Title:     "parent view",
				})
				require.NoError(t, err)
				_, err = s.Insert(&Task{
					ID:        100000,
					ProjectID: 100000,
					Title:     "parent task",
				})
				require.NoError(t, err)
				require.NoError(t, s.Commit())
			},
			write: func(s *xorm.Session) error {
				return hardDeleteProject(s, &Project{ID: 100000}, nil)
			},
			verify: func(t *testing.T) {
				db.AssertMissing(t, "tasks", map[string]interface{}{"project_id": 1})
				db.AssertMissing(t, "tasks", map[string]interface{}{"project_id": 100000})
			},
		},
		{
			name:   "restore",
			taskID: 51,
			viewID: 1,
			write: func(s *xorm.Session) error {
				task := &Task{ID: 51}
				can, err := task.CanRestore(s, u)
				if err != nil || !can {
					return err
				}
				return task.Restore(s, u)
			},
			verify: func(t *testing.T) {
				db.AssertExists(t, "tasks", map[string]interface{}{
					"id":         51,
					"deleted_at": nil,
				}, false)
			},
		},
	}
	engineSession := db.NewSession()
	engineSession.Engine().AddHook(viewLockReachedHook{})
	_ = engineSession.Close()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			if tc.setup != nil {
				tc.setup(t)
			}
			holder := db.NewSession()
			defer holder.Close()
			require.NoError(t, lockPositionsForViewUpdate(holder, tc.viewID))

			reached := &viewLockReached{
				viewID: tc.viewID,
				ready:  make(chan struct{}),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			writer := db.NewSession().Context(context.WithValue(ctx, viewLockReachedKey{}, reached))
			done := make(chan struct{})
			defer func() {
				cancel()
				_ = holder.Rollback()
				<-done
				events.CleanupPending(writer)
				_ = writer.Close()
			}()
			result := make(chan error, 1)
			go func() {
				defer close(done)
				err := tc.write(writer)
				if err == nil {
					err = writer.Commit()
				} else {
					_ = writer.Rollback()
				}
				result <- err
			}()
			select {
			case <-reached.ready:
			case err := <-result:
				t.Fatalf("Writer finished before waiting for the view: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("Writer did not reach the view lock")
			}
			var lockErr error
			switch {
			case tc.nextViewID != 0:
				lockErr = lockPositionsForViewUpdate(holder, tc.nextViewID)
			case tc.taskID == 0:
				_, lockErr = reserveTaskIndexes(holder, 1, 1)
			default:
				_, lockErr = lockingSession(holder).Unscoped().ID(tc.taskID).Get(&Task{})
			}
			_ = holder.Commit()
			select {
			case err := <-result:
				require.NoError(t, lockErr, "View holder must acquire its next lock without a deadlock")
				require.NoError(t, err, "Task writer must finish without a deadlock")
			case <-time.After(10 * time.Second):
				t.Fatal("Writer did not finish after releasing the view")
			}
			tc.verify(t)
		})
	}
}

func TestBulkTaskWriteRejectsMoveAfterViewDiscovery(t *testing.T) {
	if db.Type() != schemas.MYSQL && db.Type() != schemas.POSTGRES {
		t.Skip("Requires database row locks")
	}
	db.LoadAndAssertFixtures(t)
	u := &user.User{ID: 1}
	setup := db.NewSession()
	_, err := setup.ID(2).Cols("owner_id").Update(&Project{OwnerID: 1})
	require.NoError(t, err)
	require.NoError(t, setup.Commit())
	_ = setup.Close()
	mover := db.NewSession()
	defer mover.Close()
	defer events.CleanupPending(mover)
	require.NoError(t, lockPositionsForViewUpdate(mover, 1))

	reached := &viewLockReached{
		viewID: 1,
		ready:  make(chan struct{}),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	writer := db.NewSession().Context(context.WithValue(ctx, viewLockReachedKey{}, reached))
	writer.Engine().AddHook(viewLockReachedHook{})
	done := make(chan struct{})
	defer func() {
		cancel()
		_ = mover.Rollback()
		<-done
		events.CleanupPending(writer)
		_ = writer.Close()
	}()
	result := make(chan error, 1)
	go func() {
		defer close(done)
		_, err := updateTasks(writer, u, &Task{Done: true}, []int64{
			1,
			3,
		}, []string{"done"})
		_ = writer.Rollback()
		result <- err
	}()
	select {
	case <-reached.ready:
	case err := <-result:
		t.Fatalf("Writer finished before waiting for the view: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("Writer did not reach the view lock")
	}

	require.NoError(t, (&Task{
		ID:        1,
		Title:     "moved",
		ProjectID: 2,
	}).Update(mover, u))
	require.NoError(t, mover.Commit())
	select {
	case err := <-result:
		require.ErrorIs(t, err, ErrGenericForbidden{})
	case <-time.After(10 * time.Second):
		t.Fatal("Writer did not finish after the move")
	}
	db.AssertExists(t, "tasks", map[string]interface{}{
		"id":         1,
		"project_id": 2,
		"done":       false,
	}, false)
	db.AssertExists(t, "tasks", map[string]interface{}{
		"id":   3,
		"done": false,
	}, false)
}
