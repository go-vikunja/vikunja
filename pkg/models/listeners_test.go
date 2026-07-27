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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

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
