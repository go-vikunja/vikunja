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

func TestProjectDuplicate(t *testing.T) {

	files.InitTestFileFixtures(t)
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{
		ID: 1,
	}

	l := &ProjectDuplicate{
		ProjectID: 1,
	}
	can, err := l.CanCreate(s, u)
	require.NoError(t, err)
	assert.True(t, can)
	err = l.Create(s, u)
	require.NoError(t, err)

	originalProjectID := l.ProjectID
	duplicatedProjectID := l.Project.ID

	// assert the new project has the same number of buckets as the old one
	numberOfOriginalViews, err := s.Where("project_id = ?", originalProjectID).Count(&ProjectView{})
	require.NoError(t, err)
	numberOfDuplicatedViews, err := s.Where("project_id = ?", duplicatedProjectID).Count(&ProjectView{})
	require.NoError(t, err)
	assert.Equal(t, numberOfOriginalViews, numberOfDuplicatedViews, "duplicated project does not have the same amount of views as the original one")

	// Check that the duplicated project has the same number of tasks as the original
	numberOfOriginalTasks, err := s.Where("project_id = ?", originalProjectID).Count(&Task{})
	require.NoError(t, err)
	numberOfDuplicatedTasks, err := s.Where("project_id = ?", duplicatedProjectID).Count(&Task{})
	require.NoError(t, err)
	assert.Equal(t, numberOfOriginalTasks, numberOfDuplicatedTasks, "duplicated project does not have the same amount of tasks as the original one")

	// Check that each view has the same number of task positions between original and duplicated project
	var originalViews []*ProjectView
	err = s.Where("project_id = ?", originalProjectID).
		OrderBy("position").
		Find(&originalViews)
	require.NoError(t, err)

	var duplicatedViews []*ProjectView
	err = s.Where("project_id = ?", duplicatedProjectID).
		OrderBy("position").
		Find(&duplicatedViews)
	require.NoError(t, err)

	// Create a map of original view positions to compare with duplicated views
	originalViewPositions := make(map[int64]int64)
	for _, view := range originalViews {
		count, err := s.Where("project_view_id = ?", view.ID).Count(&TaskPosition{})
		require.NoError(t, err)
		originalViewPositions[view.ID] = count
	}

	// For each duplicated view, check if it has the same number of task positions as its original counterpart
	for i, view := range duplicatedViews {
		taskPositionsCount, err := s.Where("project_view_id = ?", view.ID).Count(&TaskPosition{})
		require.NoError(t, err)
		assert.Equal(
			t,
			originalViewPositions[originalViews[i].ID],
			taskPositionsCount,
			"duplicated view '%s' does not have the same amount of task positions as the original view", view.Title,
		)
	}

	// Check that each view has the same number of task buckets between original and duplicated project
	originalViewBuckets := make(map[int64]int64)
	for _, view := range originalViews {
		count, err := s.Where("project_view_id = ?", view.ID).Count(&TaskBucket{})
		require.NoError(t, err)
		originalViewBuckets[view.ID] = count
	}

	// For each duplicated view, check if it has the same number of task buckets as its original counterpart
	for i, view := range duplicatedViews {
		taskBucketsCount, err := s.Where("project_view_id = ?", view.ID).Count(&TaskBucket{})
		require.NoError(t, err)
		assert.Equal(
			t,
			originalViewBuckets[originalViews[i].ID],
			taskBucketsCount,
			"duplicated view '%s' does not have the same amount of task buckets as the original view", view.Title,
		)
	}

	// Check that the kanban view in the duplicated project has a different default bucket than the original
	var originalKanbanView *ProjectView
	var duplicatedKanbanView *ProjectView

	for _, view := range originalViews {
		if view.ViewKind == ProjectViewKindKanban {
			originalKanbanView = view
			break
		}
	}

	for _, view := range duplicatedViews {
		if view.ViewKind == ProjectViewKindKanban {
			duplicatedKanbanView = view
			break
		}
	}

	require.NotNil(t, originalKanbanView, "Original project does not have a kanban view")
	require.NotNil(t, duplicatedKanbanView, "Duplicated project does not have a kanban view")

	if originalKanbanView != nil && duplicatedKanbanView != nil {
		assert.NotEqual(t, originalKanbanView.DefaultBucketID, duplicatedKanbanView.DefaultBucketID)
	}
}
