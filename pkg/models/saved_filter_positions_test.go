package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSavedFilterUpdateInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "posfilter",
		Filters: &TaskCollection{Filter: "id = 1"},
	}

	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	err = sf.Update(s, u)
	require.NoError(t, err)

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	tp := &TaskPosition{}
	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, 1).Get(tp)
	require.NoError(t, err)
	require.True(t, exists)
	assert.NotZero(t, tp.Position)
}

func TestCronInsertsNonZeroPosition(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	sf := &SavedFilter{
		Title:   "cronfilter",
		Filters: &TaskCollection{Filter: "due_date > '2018-01-01T00:00:00'"},
	}

	u := &user.User{ID: 1}
	err := sf.Create(s, u)
	require.NoError(t, err)

	view := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", getProjectIDFromSavedFilterID(sf.ID), ProjectViewKindKanban).Get(view)
	require.NoError(t, err)
	require.True(t, exists)

	task := &Task{}
	exists, err = s.Where("id = ?", 5).Get(task)
	require.NoError(t, err)
	require.True(t, exists)

	tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 0}
	_, err = s.Insert(tp)
	require.NoError(t, err)

	_, err = calculateNewPositionForTask(s, u, task, view)
	require.NoError(t, err)

	exists, err = s.Where("project_view_id = ? AND task_id = ?", view.ID, task.ID).Get(tp)
	require.NoError(t, err)
	require.True(t, exists)
	assert.NotZero(t, tp.Position)
}
