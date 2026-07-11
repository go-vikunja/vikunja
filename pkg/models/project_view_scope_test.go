// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectsForProjectScope(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	current := []*Project{{ID: 27}}

	t.Run("current project only", func(t *testing.T) {
		projects, err := getProjectsForProjectScope(s, &Project{ID: 27, TaskScope: ProjectTaskScopeCurrent}, current)
		require.NoError(t, err)
		assert.Equal(t, []int64{27}, projectIDs(projects))
	})

	t.Run("all descendants", func(t *testing.T) {
		projects, err := getProjectsForProjectScope(s, &Project{ID: 27, TaskScope: ProjectTaskScopeAll}, current)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int64{27, 12, 25, 26}, projectIDs(projects))
	})

	t.Run("selected descendants ignores unrelated projects", func(t *testing.T) {
		projects, err := getProjectsForProjectScope(s, &Project{
			ID:                 27,
			TaskScope:          ProjectTaskScopeSelected,
			IncludedProjectIDs: []int64{25, 1},
		}, current)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int64{27, 25}, projectIDs(projects))
	})
}

func TestProjectIncludedProjectIDsPersistence(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	project := &Project{ID: 27, IncludedProjectIDs: []int64{12, 21}}
	_, err := s.ID(project.ID).Cols("included_project_ids").Update(project)
	require.NoError(t, err)

	stored, err := GetProjectSimpleByID(s, project.ID)
	require.NoError(t, err)
	assert.Equal(t, project.IncludedProjectIDs, stored.IncludedProjectIDs)
}

func projectIDs(projects []*Project) []int64 {
	ids := make([]int64, 0, len(projects))
	for _, project := range projects {
		ids = append(ids, project.ID)
	}
	return ids
}
