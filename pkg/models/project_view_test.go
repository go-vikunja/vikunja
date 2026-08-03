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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectView_Update(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("switch list view to kanban seeds default buckets", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:                      1,
			ProjectID:               1,
			Title:                   "List",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeNone,
		}
		err := view.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Equal(t, BucketConfigurationModeManual, view.BucketConfigurationMode)

		s2 := db.NewSession()
		defer s2.Close()
		buckets := []*Bucket{}
		err = s2.Where("project_view_id = ?", view.ID).OrderBy("position asc").Find(&buckets)
		require.NoError(t, err)
		require.Len(t, buckets, 3)

		assert.Equal(t, "To-Do", buckets[0].Title)
		assert.Equal(t, "Doing", buckets[1].Title)
		assert.Equal(t, "Done", buckets[2].Title)

		assert.Equal(t, buckets[0].ID, view.DefaultBucketID)
		assert.Equal(t, buckets[2].ID, view.DoneBucketID)
		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                        1,
			"view_kind":                 ProjectViewKindKanban,
			"bucket_configuration_mode": BucketConfigurationModeManual,
			"default_bucket_id":         buckets[0].ID,
			"done_bucket_id":            buckets[2].ID,
		}, false)

		taskCount, err := s2.Where("project_id = ?", view.ProjectID).Count(&Task{})
		require.NoError(t, err)
		assert.Positive(t, taskCount)

		taskBuckets := []*TaskBucket{}
		err = s2.Where("project_view_id = ?", view.ID).Find(&taskBuckets)
		require.NoError(t, err)
		assert.Len(t, taskBuckets, int(taskCount))
		for _, tb := range taskBuckets {
			assert.Equal(t, buckets[0].ID, tb.BucketID)
		}
	})

	t.Run("default bucket of another view is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		foreignBucket := &Bucket{}
		exists, err := s.Where("project_view_id = ?", 8).Get(foreignBucket)
		require.NoError(t, err)
		require.True(t, exists)

		view := &ProjectView{
			ID:                      4,
			ProjectID:               1,
			Title:                   "Kanban",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
			DefaultBucketID:         foreignBucket.ID,
		}
		err = view.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBucketDoesNotBelongToProject(err))
	})

	t.Run("done bucket of another view is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		foreignBucket := &Bucket{}
		exists, err := s.Where("project_view_id = ?", 8).Get(foreignBucket)
		require.NoError(t, err)
		require.True(t, exists)

		view := &ProjectView{
			ID:                      4,
			ProjectID:               1,
			Title:                   "Kanban",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
			DoneBucketID:            foreignBucket.ID,
		}
		err = view.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBucketDoesNotBelongToProject(err))
	})

	t.Run("unchanged stale bucket ids are reset instead of rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		foreignBucket := &Bucket{}
		exists, err := s.Where("project_view_id = ?", 8).Get(foreignBucket)
		require.NoError(t, err)
		require.True(t, exists)

		_, err = s.ID(4).
			Cols("default_bucket_id", "done_bucket_id").
			Update(&ProjectView{DefaultBucketID: foreignBucket.ID, DoneBucketID: foreignBucket.ID})
		require.NoError(t, err)

		view := &ProjectView{
			ID:                      4,
			ProjectID:               1,
			Title:                   "Kanban",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
			DefaultBucketID:         foreignBucket.ID,
			DoneBucketID:            foreignBucket.ID,
		}
		err = view.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Zero(t, view.DefaultBucketID)
		assert.Zero(t, view.DoneBucketID)
		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                4,
			"default_bucket_id": 0,
			"done_bucket_id":    0,
		}, false)
	})

	t.Run("echoed id of a deleted bucket is reset instead of rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		doneBucket := &Bucket{ID: 3, ProjectViewID: 4, ProjectID: 1}
		require.NoError(t, doneBucket.Delete(s, u))

		view := &ProjectView{
			ID:                      4,
			ProjectID:               1,
			Title:                   "Kanban renamed",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
			DefaultBucketID:         1,
			DoneBucketID:            3,
		}
		require.NoError(t, view.Update(s, u))
		require.NoError(t, s.Commit())

		assert.Zero(t, view.DoneBucketID)
		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                4,
			"title":             "Kanban renamed",
			"default_bucket_id": 1,
			"done_bucket_id":    0,
		}, false)
	})

	t.Run("omitted bucket configuration mode keeps the stored one", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:                      1,
			ProjectID:               1,
			Title:                   "List",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeFilter,
			BucketConfiguration: []*ProjectViewBucketConfiguration{
				{Title: "Open", Filter: &TaskCollection{Filter: "done = false"}},
			},
		}
		require.NoError(t, view.Update(s, u))

		view.BucketConfigurationMode = BucketConfigurationModeNone
		require.NoError(t, view.Update(s, u))
		require.NoError(t, s.Commit())

		assert.Equal(t, BucketConfigurationModeFilter, view.BucketConfigurationMode)
		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                        1,
			"bucket_configuration_mode": BucketConfigurationModeFilter,
		}, false)

		s2 := db.NewSession()
		defer s2.Close()
		bucketCount, err := s2.Where("project_view_id = ?", view.ID).Count(&Bucket{})
		require.NoError(t, err)
		assert.Zero(t, bucketCount)
	})

	t.Run("partial update keeps the stored bucket configuration", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:                      1,
			ProjectID:               1,
			Title:                   "List",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeFilter,
			BucketConfiguration: []*ProjectViewBucketConfiguration{
				{Title: "Open", Filter: &TaskCollection{Filter: "done = false"}},
				{Title: "Done", Filter: &TaskCollection{Filter: "done = true"}},
			},
		}
		require.NoError(t, view.Update(s, u))

		partial := &ProjectView{
			ID:        1,
			ProjectID: 1,
			Title:     "Renamed",
			ViewKind:  ProjectViewKindKanban,
		}
		require.NoError(t, partial.Update(s, u))
		require.NoError(t, s.Commit())

		assert.Equal(t, BucketConfigurationModeFilter, partial.BucketConfigurationMode)
		require.Len(t, partial.BucketConfiguration, 2)

		s2 := db.NewSession()
		defer s2.Close()
		stored, err := GetProjectViewByIDAndProject(s2, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, BucketConfigurationModeFilter, stored.BucketConfigurationMode)
		require.Len(t, stored.BucketConfiguration, 2)
		assert.Equal(t, "Open", stored.BucketConfiguration[0].Title)
		assert.Equal(t, "done = true", stored.BucketConfiguration[1].Filter.Filter)
	})

	t.Run("updating an already manual kanban view does not seed again", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		bucketsBefore, err := s.Where("project_view_id = ?", 4).Count(&Bucket{})
		require.NoError(t, err)
		taskBucketsBefore, err := s.Where("project_view_id = ?", 4).Count(&TaskBucket{})
		require.NoError(t, err)

		view := &ProjectView{
			ID:                      4,
			ProjectID:               1,
			Title:                   "Kanban renamed",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeManual,
			DefaultBucketID:         1,
			DoneBucketID:            3,
		}
		require.NoError(t, view.Update(s, u))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		bucketsAfter, err := s2.Where("project_view_id = ?", 4).Count(&Bucket{})
		require.NoError(t, err)
		assert.Equal(t, bucketsBefore, bucketsAfter)

		taskBucketsAfter, err := s2.Where("project_view_id = ?", 4).Count(&TaskBucket{})
		require.NoError(t, err)
		assert.Equal(t, taskBucketsBefore, taskBucketsAfter)
	})

	t.Run("saved filter view backfills tasks and positions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		sf := &SavedFilter{
			Title:   "backfill filter",
			Filters: &TaskCollection{Filter: "done = false"},
		}
		require.NoError(t, sf.Create(s, u))

		view := &ProjectView{}
		exists, err := s.
			Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).
			Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		_, err = s.Where("project_view_id = ?", view.ID).Delete(&TaskBucket{})
		require.NoError(t, err)
		_, err = s.Where("project_view_id = ?", view.ID).Delete(&TaskPosition{})
		require.NoError(t, err)

		view.ViewKind = ProjectViewKindList
		require.NoError(t, view.Update(s, u))

		view.ViewKind = ProjectViewKindKanban
		view.BucketConfigurationMode = BucketConfigurationModeNone
		require.NoError(t, view.Update(s, u))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		taskBucketCount, err := s2.Where("project_view_id = ?", view.ID).Count(&TaskBucket{})
		require.NoError(t, err)
		assert.Positive(t, taskBucketCount)

		positionCount, err := s2.Where("project_view_id = ? AND position != 0", view.ID).Count(&TaskPosition{})
		require.NoError(t, err)
		assert.Positive(t, positionCount)
	})

	t.Run("switch list view to kanban with filter mode does not seed buckets", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:                      1,
			ProjectID:               1,
			Title:                   "List",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeFilter,
			BucketConfiguration: []*ProjectViewBucketConfiguration{
				{Title: "Open", Filter: &TaskCollection{Filter: "done = false"}},
			},
		}
		err := view.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		bucketCount, err := s2.Where("project_view_id = ?", view.ID).Count(&Bucket{})
		require.NoError(t, err)
		assert.Zero(t, bucketCount)
	})

	t.Run("invalid bucket configuration filter is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:                      1,
			ProjectID:               1,
			Title:                   "List",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeFilter,
			BucketConfiguration: []*ProjectViewBucketConfiguration{
				{Title: "Broken", Filter: &TaskCollection{Filter: "nonexistingfield = true"}},
			},
		}
		err := view.Update(s, u)
		require.ErrorContains(t, err, "nonexistingfield")
	})

	t.Run("switch kanban view to list resets bucket configuration mode", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:                      4,
			ProjectID:               1,
			Title:                   "Kanban",
			ViewKind:                ProjectViewKindList,
			BucketConfigurationMode: BucketConfigurationModeManual,
		}
		err := view.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                        4,
			"view_kind":                 ProjectViewKindList,
			"bucket_configuration_mode": BucketConfigurationModeNone,
		}, false)
	})

	t.Run("kanban round-trip does not duplicate buckets", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:        4,
			ProjectID: 1,
			Title:     "Kanban",
			ViewKind:  ProjectViewKindList,
		}
		err := view.Update(s, u)
		require.NoError(t, err)

		taskWhileList := &Task{
			Title:     "task created while the view was a list",
			ProjectID: 1,
		}
		err = taskWhileList.Create(s, u)
		require.NoError(t, err)

		view.ViewKind = ProjectViewKindKanban
		view.BucketConfigurationMode = BucketConfigurationModeNone
		err = view.Update(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		bucketCount, err := s2.Where("project_view_id = ?", view.ID).Count(&Bucket{})
		require.NoError(t, err)
		// View 4 has 3 buckets in the fixtures, they survive the round-trip
		assert.Equal(t, int64(3), bucketCount)

		assert.NotZero(t, view.DefaultBucketID)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":         taskWhileList.ID,
			"project_view_id": view.ID,
			"bucket_id":       view.DefaultBucketID,
		}, false)
	})

	t.Run("kanban round-trip restores both bucket ids", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ID:        4,
			ProjectID: 1,
			Title:     "Kanban",
			ViewKind:  ProjectViewKindList,
		}
		require.NoError(t, view.Update(s, u))

		view.ViewKind = ProjectViewKindKanban
		view.BucketConfigurationMode = BucketConfigurationModeNone
		view.DefaultBucketID = 0
		view.DoneBucketID = 0
		require.NoError(t, view.Update(s, u))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                4,
			"default_bucket_id": 1,
			"done_bucket_id":    3,
		}, false)
	})
}

func TestProjectView_InsertTaskBuckets(t *testing.T) {
	u := &user.User{ID: 1}

	// Tasks 2 and 47 both belong to project 1, but only task 2 is placed in
	// view 4 by the fixtures.
	t.Run("already placed task does not conflict", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := insertTaskBuckets(s, 4, 1, []int64{2, 47})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()

		count, err := s2.Where("task_id = ? AND project_view_id = ?", 2, 4).Count(&TaskBucket{})
		require.NoError(t, err)
		assert.EqualValues(t, 1, count)

		existing := &TaskBucket{}
		_, err = s2.Where("task_id = ? AND project_view_id = ?", 2, 4).Get(existing)
		require.NoError(t, err)
		assert.EqualValues(t, 3, existing.BucketID, "the existing placement must win")

		added := &TaskBucket{}
		exists, err := s2.Where("task_id = ? AND project_view_id = ?", 47, 4).Get(added)
		require.NoError(t, err)
		require.True(t, exists)
		assert.EqualValues(t, 1, added.BucketID)
	})

	t.Run("switch to kanban with a task already placed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&TaskBucket{TaskID: 1, BucketID: 1, ProjectViewID: 1})
		require.NoError(t, err)

		view := &ProjectView{
			ID:                      1,
			ProjectID:               1,
			Title:                   "List",
			ViewKind:                ProjectViewKindKanban,
			BucketConfigurationMode: BucketConfigurationModeNone,
		}
		require.NoError(t, view.Update(s, u))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()

		count, err := s2.Where("task_id = ? AND project_view_id = ?", 1, 1).Count(&TaskBucket{})
		require.NoError(t, err)
		assert.EqualValues(t, 1, count)
	})
}

func TestProjectView_SavedFilterWithBucketIDFilter(t *testing.T) {
	u := &user.User{ID: 1}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "bucket filter",
		Filters: &TaskCollection{Filter: "bucket_id = 1"},
	}
	require.NoError(t, sf.Create(s, u))

	sf.Title = "bucket filter renamed"
	require.NoError(t, sf.Update(s, u))
	require.NoError(t, s.Commit())
}

func TestProjectView_Create(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("creating a kanban view without a mode seeds default buckets", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ProjectID: 1,
			Title:     "New Kanban",
			ViewKind:  ProjectViewKindKanban,
		}
		err := view.Create(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Equal(t, BucketConfigurationModeManual, view.BucketConfigurationMode)

		s2 := db.NewSession()
		defer s2.Close()
		buckets := []*Bucket{}
		err = s2.Where("project_view_id = ?", view.ID).OrderBy("position asc").Find(&buckets)
		require.NoError(t, err)
		require.Len(t, buckets, 3)
		assert.Equal(t, buckets[0].ID, view.DefaultBucketID)
		assert.Equal(t, buckets[2].ID, view.DoneBucketID)
	})

	t.Run("bucket ids of another view are discarded", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		view := &ProjectView{
			ProjectID:       1,
			Title:           "New List",
			ViewKind:        ProjectViewKindList,
			DefaultBucketID: 1,
			DoneBucketID:    3,
		}
		err := view.Create(s, u)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Zero(t, view.DefaultBucketID)
		assert.Zero(t, view.DoneBucketID)
		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":                view.ID,
			"default_bucket_id": 0,
			"done_bucket_id":    0,
		}, false)
	})
}
