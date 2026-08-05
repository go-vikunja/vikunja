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

// The listener evaluates every kanban filter view in the instance, so a filter owned by
// a disabled user used to fail the handler for everyone: recalculating positions looks
// up the owner, which errors out with "Account is disabled".
func TestUpdateTaskInSavedFilterViews_InactiveFilterOwner(t *testing.T) {
	// Positions are crowded enough to force a recalculation when a task is added.
	createCrowdedFilterView := func(t *testing.T, s *xorm.Session, filterID, viewID, ownerID int64) *ProjectView {
		_, err := s.Insert(&SavedFilter{
			ID:      filterID,
			Title:   "filter",
			OwnerID: ownerID,
			Filters: &TaskCollection{Filter: "done = false"},
		})
		require.NoError(t, err)

		view := &ProjectView{
			ID:                      viewID,
			ProjectID:               getProjectIDFromSavedFilterID(filterID),
			Title:                   "kanban",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
		}
		_, err = s.Insert(view)
		require.NoError(t, err)

		_, err = s.Insert(&Bucket{ProjectViewID: view.ID, Title: "backlog", CreatedByID: ownerID})
		require.NoError(t, err)

		_, err = s.Insert(&TaskPosition{TaskID: 2, ProjectViewID: view.ID, Position: MinPositionSpacing / 2})
		require.NoError(t, err)

		return view
	}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	disabledOwnerView := createCrowdedFilterView(t, s, 9998, 9998, 17)
	activeOwnerView := createCrowdedFilterView(t, s, 9999, 9999, 1)
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

	_, err := s.Insert(&SavedFilter{
		ID:      9999,
		Title:   "filter",
		OwnerID: 1,
		Filters: &TaskCollection{Filter: "done = false"},
	})
	require.NoError(t, err)

	view := &ProjectView{
		ID:                      9999,
		ProjectID:               getProjectIDFromSavedFilterID(9999),
		Title:                   "kanban",
		ViewKind:                ProjectViewKindKanban,
		BucketConfigurationMode: BucketConfigurationModeManual,
	}
	_, err = s.Insert(view)
	require.NoError(t, err)

	_, err = s.Insert(&Bucket{ProjectViewID: view.ID, Title: "backlog", CreatedByID: 1})
	require.NoError(t, err)
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
