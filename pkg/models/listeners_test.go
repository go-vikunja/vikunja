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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/audit"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// The CalDAV ctag derives from the project's updated timestamp, so the
// listener must bump it on sub-entity changes or clients never refetch.
func TestHandleTaskUpdateLastUpdated(t *testing.T) {
	t.Run("bumps task and project updated times", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		taskBefore, err := GetTaskByIDSimple(s, 1)
		require.NoError(t, err)
		projectBefore, err := GetProjectSimpleByID(s, 1)
		require.NoError(t, err)
		_ = s.Close()

		events.TestListener(t, &TaskRelationCreatedEvent{
			Task: &taskBefore,
			Doer: &user.User{ID: 1},
		}, &HandleTaskUpdateLastUpdated{})

		s2 := db.NewSession()
		defer s2.Close()
		taskAfter, err := GetTaskByIDSimple(s2, 1)
		require.NoError(t, err)
		projectAfter, err := GetProjectSimpleByID(s2, 1)
		require.NoError(t, err)

		require.True(t, taskAfter.Updated.After(taskBefore.Updated), "task updated time must advance")
		require.True(t, projectAfter.Updated.After(projectBefore.Updated), "project updated time must advance")
	})

	t.Run("does not fail for a nonexistent task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		events.TestListener(t, &TaskRelationCreatedEvent{
			Task: &Task{ID: 99999},
			Doer: &user.User{ID: 1},
		}, &HandleTaskUpdateLastUpdated{})
	})
}

// Access-filtering only excludes owners without access to the task's project, not disabled
// ones: recalculating positions for a disabled owner's filter must not fail it for everyone.
func TestUpdateTaskInSavedFilterViews_InactiveFilterOwner(t *testing.T) {
	// Positions are crowded enough to force a recalculation when a task is added.
	createCrowdedFilterView := func(t *testing.T, s *xorm.Session, filterID, viewID, ownerID int64) *ProjectView {
		view, _ := createKanbanFilterView(t, s, filterID, viewID, ownerID, "done = false")
		_, err := s.Insert(&TaskPosition{TaskID: 2, ProjectViewID: view.ID, Position: MinPositionSpacing / 2})
		require.NoError(t, err)
		return view
	}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	disabledOwnerView := createCrowdedFilterView(t, s, 9998, 9998, 17)
	activeOwnerView := createCrowdedFilterView(t, s, 9999, 9999, 1)
	// Filters of users without access are skipped anyway, so share project 1 with the disabled owner.
	_, err := s.Insert(&ProjectUser{UserID: 17, ProjectID: 1, Permission: PermissionRead})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	_ = s.Close()

	events.TestListener(t, &TaskUpdatedEvent{
		Task: &Task{ID: 1, ProjectID: 1},
		Doer: &user.User{ID: 1},
	}, &UpdateTaskInSavedFilterViews{})

	s2 := db.NewSession()
	defer s2.Close()

	// The filter of the active owner is still maintained,
	db.AssertExists(t, "task_buckets", map[string]interface{}{
		"task_id":         1,
		"project_view_id": activeOwnerView.ID,
	}, false)

	// the one of the disabled owner is skipped.
	db.AssertMissing(t, "task_buckets", map[string]interface{}{
		"task_id":         1,
		"project_view_id": disabledOwnerView.ID,
	})
}

func TestUpdateTasksBatchInSavedFilterViews(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()

	view, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "done = false")
	require.NoError(t, s.Commit())
	_ = s.Close()

	events.TestListener(t, &TasksBatchCreatedEvent{
		Tasks: []*Task{
			{ID: 1, ProjectID: 1, Index: 1},
			{ID: 3, ProjectID: 1, Index: 3},
		},
		Doer: &user.User{ID: 1},
	}, &UpdateTasksBatchInSavedFilterViews{})

	for _, taskID := range []int64{1, 3} {
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         taskID,
			"project_view_id": view.ID,
		}, false)
		db.AssertExists(t, "task_positions", map[string]interface{}{
			"task_id":         taskID,
			"project_view_id": view.ID,
		}, false)
	}

	s2 := db.NewSession()
	defer s2.Close()
	positions := []*TaskPosition{}
	require.NoError(t, s2.Where("project_view_id = ?", view.ID).In("task_id", 1, 3).Find(&positions))
	require.Len(t, positions, 2)

	positionByTask := map[int64]float64{}
	for _, p := range positions {
		positionByTask[p.TaskID] = p.Position
	}
	assert.NotZero(t, positionByTask[1])
	assert.NotZero(t, positionByTask[3])
	// Each task lands on top of the view, so later batch members get smaller positions.
	assert.Less(t, positionByTask[3], positionByTask[1])
}

func newKanbanView(t *testing.T, s *xorm.Session, viewID, projectID, ownerID int64) (*ProjectView, *Bucket) {
	view := &ProjectView{
		ID:                      viewID,
		ProjectID:               projectID,
		Title:                   "kanban",
		ViewKind:                ProjectViewKindKanban,
		BucketConfigurationMode: BucketConfigurationModeManual,
	}
	_, err := s.Insert(view)
	require.NoError(t, err)

	bucket := &Bucket{ProjectViewID: view.ID, Title: "backlog", CreatedByID: ownerID}
	_, err = s.Insert(bucket)
	require.NoError(t, err)

	return view, bucket
}

func createKanbanFilterView(t *testing.T, s *xorm.Session, filterID, viewID, ownerID int64, filter string) (*ProjectView, *Bucket) {
	_, err := s.Insert(&SavedFilter{
		ID:      filterID,
		Title:   "filter",
		OwnerID: ownerID,
		Filters: &TaskCollection{Filter: filter},
	})
	require.NoError(t, err)

	return newKanbanView(t, s, viewID, getProjectIDFromSavedFilterID(filterID), ownerID)
}

// A filter can only contain tasks its owner can see, so filters of users without
// access to the task's project must neither be evaluated nor receive rows.
// Task 1 lives on project 1 (owner user 1), task 34 on project 20 (owner user 13); neither is shared.
func TestUpdateTasksInSavedFilterViews_OnlyFiltersOfUsersWithAccess(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	user1View, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "done = false")
	user13View, _ := createKanbanFilterView(t, s, 9998, 9998, 13, "done = false")
	require.NoError(t, s.Commit())
	_ = s.Close()

	events.TestListener(t, &TasksBatchCreatedEvent{
		Tasks: []*Task{
			{ID: 1, ProjectID: 1, Index: 1},
			{ID: 34, ProjectID: 20, Index: 20},
		},
		Doer: &user.User{ID: 1},
	}, &UpdateTasksBatchInSavedFilterViews{})

	for _, table := range []string{"task_buckets", "task_positions"} {
		db.AssertExists(t, table, map[string]interface{}{"task_id": 1, "project_view_id": user1View.ID}, false)
		db.AssertExists(t, table, map[string]interface{}{"task_id": 34, "project_view_id": user13View.ID}, false)
		db.AssertMissing(t, table, map[string]interface{}{"task_id": 34, "project_view_id": user1View.ID})
		db.AssertMissing(t, table, map[string]interface{}{"task_id": 1, "project_view_id": user13View.ID})
	}
}

// Access via share, team or ancestor ownership must feed the filter like direct ownership.
func TestUpdateTasksInSavedFilterViews_AccessViaShareOrParent(t *testing.T) {
	assertHasRows := func(t *testing.T, taskID, viewID int64) {
		for _, table := range []string{"task_buckets", "task_positions"} {
			db.AssertExists(t, table, map[string]interface{}{"task_id": taskID, "project_view_id": viewID}, false)
		}
	}
	assertNoRows := func(t *testing.T, taskID, viewID int64) {
		for _, table := range []string{"task_buckets", "task_positions"} {
			db.AssertMissing(t, table, map[string]interface{}{"task_id": taskID, "project_view_id": viewID})
		}
	}

	// Project 9 is owned by user 6 and shared read-only with user 1 only via users_projects id 3 (no team path); user 16 has no access.
	t.Run("direct project share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		sharedView, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "done = false")
		noAccessView, _ := createKanbanFilterView(t, s, 9998, 9998, 16, "done = false")
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &Task{ID: 18, ProjectID: 9},
			Doer: &user.User{ID: 1},
		}, &UpdateTaskInSavedFilterViews{})

		assertHasRows(t, 18, sharedView.ID)
		assertNoRows(t, 18, noAccessView.ID)
	})

	// Project 6 is owned by user 6 and shared read-only with team 2 (team_projects id 2), whose only member is user 1.
	t.Run("team project share", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		teamView, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "done = false")
		noAccessView, _ := createKanbanFilterView(t, s, 9998, 9998, 16, "done = false")
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &Task{ID: 15, ProjectID: 6},
			Doer: &user.User{ID: 1},
		}, &UpdateTaskInSavedFilterViews{})

		assertHasRows(t, 15, teamView.ID)
		assertNoRows(t, 15, noAccessView.ID)
	})

	// Fresh parent/child pair so ancestor ownership is the only access path (project 19's
	// owner, user 6, has a direct admin share on it too, which would otherwise mask this).
	t.Run("parent project owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		const (
			childProjectID int64 = 9991
			taskID         int64 = 9990
			parentOwnerID  int64 = 16
			childOwnerID   int64 = 13
			noAccessID     int64 = 2
		)
		parentProjectID := int64(9990)
		_, err := s.Insert(&Project{ID: parentProjectID, Title: "ancestor parent", Identifier: "ANCPARENT", OwnerID: parentOwnerID})
		require.NoError(t, err)
		_, err = s.Insert(&Project{ID: childProjectID, Title: "ancestor child", Identifier: "ANCCHILD", OwnerID: childOwnerID, ParentProjectID: &parentProjectID})
		require.NoError(t, err)
		_, err = s.Insert(&Task{ID: taskID, Title: "child project task", ProjectID: childProjectID, Index: 1, CreatedByID: childOwnerID})
		require.NoError(t, err)

		parentOwnerView, _ := createKanbanFilterView(t, s, 9999, 9999, parentOwnerID, "done = false")
		noAccessView, _ := createKanbanFilterView(t, s, 9998, 9998, noAccessID, "done = false")
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &Task{ID: taskID, ProjectID: childProjectID},
			Doer: &user.User{ID: 1},
		}, &UpdateTaskInSavedFilterViews{})

		assertHasRows(t, taskID, parentOwnerView.ID)
		assertNoRows(t, taskID, noAccessView.ID)
	})
}

// Tasks not matching the filter expression get no rows even though the owner can see them.
func TestUpdateTasksInSavedFilterViews_TaskNotInFilter(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	view, _ := createKanbanFilterView(t, s, 9999, 9999, 1, "done = false")
	require.NoError(t, s.Commit())
	_ = s.Close()

	// Task 2 is done.
	events.TestListener(t, &TaskUpdatedEvent{
		Task: &Task{ID: 2, ProjectID: 1},
		Doer: &user.User{ID: 1},
	}, &UpdateTaskInSavedFilterViews{})

	db.AssertMissing(t, "task_buckets", map[string]interface{}{"task_id": 2, "project_view_id": view.ID})
	db.AssertMissing(t, "task_positions", map[string]interface{}{"task_id": 2, "project_view_id": view.ID})
}

// Covers multi-view fan-out and the per-view bucket_id LEFT JOIN match, neither tested above.
func TestUpdateTasksInSavedFilterViews_BucketIDFilterAndMultipleViews(t *testing.T) {
	t.Run("multiple views of one filter", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		// createKanbanFilterView already inserts the filter, so the second view only adds a ProjectView + Bucket.
		view1, bucket1 := createKanbanFilterView(t, s, 9990, 9990, 1, "done = false")
		view2, bucket2 := newKanbanView(t, s, 9991, view1.ProjectID, 1)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &Task{ID: 1, ProjectID: 1},
			Doer: &user.User{ID: 1},
		}, &UpdateTaskInSavedFilterViews{})

		for _, vb := range []struct {
			view   *ProjectView
			bucket *Bucket
		}{{view1, bucket1}, {view2, bucket2}} {
			// Each view has its own default bucket; the task must land in that view's own bucket, not the other view's.
			db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": vb.view.ID, "bucket_id": vb.bucket.ID}, false)
			db.AssertExists(t, "task_positions", map[string]interface{}{"task_id": 1, "project_view_id": vb.view.ID}, false)
		}
	})

	t.Run("filter on bucket_id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		const filterID, viewID int64 = 9992, 9992
		view, _ := newKanbanView(t, s, viewID, getProjectIDFromSavedFilterID(filterID), 1)
		bucket := &Bucket{ProjectViewID: view.ID, Title: "target", CreatedByID: 1}
		_, err := s.Insert(bucket)
		require.NoError(t, err)

		_, err = s.Insert(&SavedFilter{
			ID:      filterID,
			Title:   "bucket filter",
			OwnerID: 1,
			Filters: &TaskCollection{Filter: fmt.Sprintf("done = false && bucket_id = %d", bucket.ID)},
		})
		require.NoError(t, err)

		plainView, _ := createKanbanFilterView(t, s, 9993, 9993, 1, "done = false")
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &Task{ID: 1, ProjectID: 1},
			Doer: &user.User{ID: 1},
		}, &UpdateTaskInSavedFilterViews{})

		// The plain done=false filter is unaffected by the bucket_id filter's join path.
		db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": plainView.ID}, false)
		db.AssertExists(t, "task_positions", map[string]interface{}{"task_id": 1, "project_view_id": plainView.ID}, false)

		// Task 1 has no task_buckets row in this view yet, so the LEFT JOIN yields no match.
		db.AssertMissing(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": view.ID})
		db.AssertMissing(t, "task_positions", map[string]interface{}{"task_id": 1, "project_view_id": view.ID})
	})

	t.Run("filter on bucket_id, task already in bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		const filterID, viewID int64 = 9994, 9994
		view, _ := newKanbanView(t, s, viewID, getProjectIDFromSavedFilterID(filterID), 1)
		bucket := &Bucket{ProjectViewID: view.ID, Title: "target", CreatedByID: 1}
		_, err := s.Insert(bucket)
		require.NoError(t, err)

		_, err = s.Insert(&SavedFilter{
			ID:      filterID,
			Title:   "bucket filter",
			OwnerID: 1,
			Filters: &TaskCollection{Filter: fmt.Sprintf("done = false && bucket_id = %d", bucket.ID)},
		})
		require.NoError(t, err)

		// Task 1 is already in the matching bucket for this view, so the join now matches it.
		_, err = s.Insert(&TaskBucket{TaskID: 1, ProjectViewID: view.ID, BucketID: bucket.ID})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskUpdatedEvent{
			Task: &Task{ID: 1, ProjectID: 1},
			Doer: &user.User{ID: 1},
		}, &UpdateTaskInSavedFilterViews{})

		// The freshly computed position is written even though the bucket already existed.
		db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": view.ID, "bucket_id": bucket.ID}, false)
		db.AssertExists(t, "task_positions", map[string]interface{}{"task_id": 1, "project_view_id": view.ID}, false)

		s3 := db.NewSession()
		defer s3.Close()
		position := &TaskPosition{}
		has, err := s3.Where("task_id = ? AND project_view_id = ?", 1, view.ID).Get(position)
		require.NoError(t, err)
		require.True(t, has)
		assert.NotZero(t, position.Position)
	})
}

// A filter that fails to parse is logged and skipped, not treated as a listener error
// that would abort processing for every other filter.
func TestUpdateTasksInSavedFilterViews_InvalidFilterIsSkipped(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()

	const brokenFilterID, brokenViewID int64 = 9995, 9995
	// bogus_field has no corresponding Task struct field, so parsing rejects it with ErrInvalidTaskField.
	brokenView, _ := createKanbanFilterView(t, s, brokenFilterID, brokenViewID, 1, "bogus_field = 1")

	healthyView, _ := createKanbanFilterView(t, s, 9996, 9996, 1, "done = false")
	require.NoError(t, s.Commit())
	_ = s.Close()

	events.TestListener(t, &TaskUpdatedEvent{
		Task: &Task{ID: 1, ProjectID: 1},
		Doer: &user.User{ID: 1},
	}, &UpdateTaskInSavedFilterViews{})

	db.AssertExists(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": healthyView.ID}, false)
	db.AssertExists(t, "task_positions", map[string]interface{}{"task_id": 1, "project_view_id": healthyView.ID}, false)

	db.AssertMissing(t, "task_buckets", map[string]interface{}{"task_id": 1, "project_view_id": brokenView.ID})
	db.AssertMissing(t, "task_positions", map[string]interface{}{"task_id": 1, "project_view_id": brokenView.ID})
}

// Subscriptions survive losing access to the entity they point at: nothing
// purges them when a share is removed, and access can change with no
// revocation event at all. Every notification listener must therefore re-check
// read permission before delivering.
//
// Task 32 lives on project 3, which user 2 can read and user 6 cannot.
func TestSubscriberNotifications_SkipUsersWithoutReadAccess(t *testing.T) {
	const (
		taskID      int64 = 32
		projectID   int64 = 3
		withAccess  int64 = 2
		lostAccess  int64 = 6
		doerID      int64 = 1
		assigneeID  int64 = 3
		childProjID int64 = 9990
	)

	subscribeBoth := func(t *testing.T, s *xorm.Session, entityType SubscriptionEntityType, entityID int64) {
		for _, userID := range []int64{withAccess, lostAccess} {
			_, err := s.Insert(&Subscription{
				UserID:     userID,
				EntityType: entityType,
				EntityID:   entityID,
			})
			require.NoError(t, err)
		}
	}

	assertOnlySubscriberWithAccessNotified := func(t *testing.T, notificationName string) {
		db.AssertExists(t, "notifications", map[string]interface{}{
			"notifiable_id": withAccess,
			"name":          notificationName,
		}, false)
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"notifiable_id": lostAccess,
			"name":          notificationName,
		})
	}

	t.Run("task comment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeBoth(t, s, SubscriptionEntityTask, taskID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)

		comment := &TaskComment{Comment: "secret", TaskID: taskID, AuthorID: doerID}
		_, err = s.Insert(comment)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    &user.User{ID: doerID},
			Comment: comment,
		}, &SendTaskCommentNotification{})

		assertOnlySubscriberWithAccessNotified(t, (&TaskCommentNotification{}).Name())
	})

	t.Run("task assigned", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeBoth(t, s, SubscriptionEntityTask, taskID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)
		assignee, err := user.GetUserByID(s, assigneeID)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskAssigneeCreatedEvent{
			Task:     &task,
			Assignee: assignee,
			Doer:     &user.User{ID: doerID},
		}, &SendTaskAssignedNotification{})

		assertOnlySubscriberWithAccessNotified(t, (&TaskAssignedNotification{}).Name())
	})

	t.Run("task deleted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeBoth(t, s, SubscriptionEntityTask, taskID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskDeletedEvent{
			Task: &task,
			Doer: &user.User{ID: doerID},
		}, &SendTaskDeletedNotification{})

		assertOnlySubscriberWithAccessNotified(t, (&TaskDeletedNotification{}).Name())
	})

	t.Run("task created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeBoth(t, s, SubscriptionEntityTask, taskID)

		task, err := GetTaskByIDSimple(s, taskID)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &TaskCreatedEvent{
			Task: &task,
			Doer: &user.User{ID: doerID},
		}, &SendTaskCreatedNotification{})

		assertOnlySubscriberWithAccessNotified(t, (&TaskCreatedNotification{}).Name())
	})

	// Subscribers are inherited from the parent project, but the notification
	// discloses the newly created child, so the child is what gets checked.
	t.Run("project created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		subscribeBoth(t, s, SubscriptionEntityProject, projectID)

		parentID := projectID
		child := &Project{
			ID:              childProjID,
			Title:           "child",
			Identifier:      "CHILD",
			OwnerID:         doerID,
			ParentProjectID: &parentID,
		}
		_, err := s.Insert(child)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		events.TestListener(t, &ProjectCreatedEvent{
			Project: child,
			Doer:    &user.User{ID: doerID},
		}, &SendProjectCreatedNotification{})

		assertOnlySubscriberWithAccessNotified(t, (&ProjectCreatedNotification{}).Name())
	})
}

// The listener runs after the deleting transaction committed, so the task is already
// soft-deleted by the time it looks up who to notify.
func TestSendTaskDeletedNotification(t *testing.T) {
	// Task 32 belongs to project 3, which user 2 can read and user 6 cannot.
	const (
		taskID       = 32
		projectID    = 3
		subscriberID = 2
		outsiderID   = 6
	)

	doer := &user.User{ID: 3}
	notificationName := (&TaskDeletedNotification{}).Name()

	deleteSubscribedTask := func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()

		for _, userID := range []int64{subscriberID, outsiderID} {
			_, err := s.Insert(&Subscription{
				EntityType: SubscriptionEntityTask,
				EntityID:   taskID,
				UserID:     userID,
			})
			require.NoError(t, err)
		}

		task := &Task{ID: taskID}
		require.NoError(t, task.ReadOne(s, doer))
		require.Equal(t, int64(projectID), task.ProjectID)

		canRead, _, err := (&Project{ID: projectID}).CanRead(s, &user.User{ID: subscriberID})
		require.NoError(t, err)
		require.True(t, canRead)
		canRead, _, err = (&Project{ID: projectID}).CanRead(s, &user.User{ID: outsiderID})
		require.NoError(t, err)
		require.False(t, canRead)

		require.NoError(t, task.Delete(s, doer))
		require.NoError(t, s.Commit())
		_ = s.Close()

		s2 := db.NewSession()
		softDeleted, err := s2.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", taskID).Exist(&Task{})
		require.NoError(t, err)
		require.True(t, softDeleted, "the task must be soft-deleted before the event fires")
		_ = s2.Close()

		events.TestListener(t, &TaskDeletedEvent{Task: task, Doer: doer}, &SendTaskDeletedNotification{})
	}

	t.Run("notifies a subscriber who can read the project", func(t *testing.T) {
		deleteSubscribedTask(t)

		db.AssertExists(t, "notifications", map[string]interface{}{
			"notifiable_id": subscriberID,
			"name":          notificationName,
		}, false)
	})
	t.Run("does not notify a subscriber who cannot read the project", func(t *testing.T) {
		deleteSubscribedTask(t)

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"notifiable_id": outsiderID,
			"name":          notificationName,
		})
	})
}

// The listener registry is global and watermill rejects duplicate handler
// names, so register once per process (relevant for -count > 1).
var registerAuditEventsOnce sync.Once

// A full personal data export is the most sensitive self-service action there
// is. Driving it through the real router instead of calling the listener
// directly is what makes this a regression test of the registration itself.
func TestAuditUserDataExportRequested(t *testing.T) {
	logfile := filepath.Join(t.TempDir(), "audit.log")
	config.AuditLogfile.Set(logfile)
	require.NoError(t, audit.Init())
	t.Cleanup(audit.Close)

	license.SetForTests([]license.Feature{license.FeatureAuditLogs})
	t.Cleanup(license.ResetForTests)

	registerAuditEventsOnce.Do(registerEventsForAuditLogging)

	events.Unfake()
	t.Cleanup(events.Fake)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	ready, err := events.InitEventsForTesting(ctx)
	require.NoError(t, err)
	<-ready

	require.NoError(t, events.Dispatch(&UserDataExportRequestedEvent{User: &user.User{ID: 42}}))

	var content []byte
	require.Eventually(t, func() bool {
		c, err := os.ReadFile(logfile)
		if err != nil {
			return false
		}
		content = bytes.TrimSpace(c)
		return len(content) > 0
	}, 5*time.Second, 10*time.Millisecond, "expected an audit entry for the export request")

	var entry audit.Entry
	require.NoError(t, json.Unmarshal(content, &entry))
	assert.Equal(t, audit.ActionUserDataExportRequested, entry.Action)
	assert.Equal(t, audit.UserActor(42), entry.Actor)
	assert.Equal(t, audit.UserTarget(42), entry.Target)
	assert.Equal(t, audit.OutcomeSuccess, entry.Outcome)
}

func TestWebhookDeliveryListenerSkipsErrorReporting(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// httptest binds to loopback, which the SSRF-safe client blocks by default.
	previousAllowNonRoutable := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
	previousClient := webhookClient
	webhookClient = nil
	t.Cleanup(func() {
		config.OutgoingRequestsAllowNonRoutableIPs.Set(previousAllowNonRoutable)
		webhookClient = previousClient
	})

	s := db.NewSession()
	webhook := &Webhook{
		TargetURL:   ts.URL,
		Events:      []string{"task.updated"},
		ProjectID:   1,
		CreatedByID: 1,
	}
	_, err := s.Insert(webhook)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	_ = s.Close()

	payload, err := json.Marshal(&WebhookDeliveryEvent{
		WebhookID: webhook.ID,
		Payload:   &WebhookPayload{EventName: "task.updated"},
	})
	require.NoError(t, err)

	msg := message.NewMessage("test", payload)
	err = (&WebhookDeliveryListener{}).Handle(msg)

	require.Error(t, err)
	assert.Equal(t, "true", msg.Metadata.Get(events.MetadataSkipErrorReporting))
}
