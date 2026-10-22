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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"

	"github.com/stretchr/testify/require"
)

// Guards foreign and unrelated view positioning (GHSA-w39f-h553-h2mx).
func TestTaskPositionCanUpdateValidatesView(t *testing.T) {
	setup := func(t *testing.T) (s *xorm.Session, task *Task, view, otherView *ProjectView, writer *user.User) {
		t.Helper()
		db.LoadAndAssertFixtures(t)
		s = db.NewSession()

		owner := &user.User{ID: 1}
		writer = &user.User{ID: 2}

		proj := &Project{Title: "position-validation project"}
		require.NoError(t, proj.Create(s, owner))
		_, err := s.Insert(&ProjectUser{UserID: writer.ID, ProjectID: proj.ID, Permission: PermissionWrite})
		require.NoError(t, err)

		task = &Task{Title: "position-validation task", ProjectID: proj.ID}
		require.NoError(t, task.Create(s, owner))

		view = &ProjectView{Title: "position-validation view", ProjectID: proj.ID, ViewKind: ProjectViewKindList, Position: 1}
		_, err = s.Insert(view)
		require.NoError(t, err)

		otherProj := &Project{Title: "position-validation other project"}
		require.NoError(t, otherProj.Create(s, owner))
		otherView = &ProjectView{Title: "position-validation other view", ProjectID: otherProj.ID, ViewKind: ProjectViewKindList, Position: 1}
		_, err = s.Insert(otherView)
		require.NoError(t, err)

		require.NoError(t, s.Commit())
		return
	}

	t.Run("valid write to a view of the task's project", func(t *testing.T) {
		s, task, view, _, writer := setup(t)
		defer s.Close()

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 123}
		can, err := tp.CanUpdate(s, writer)
		require.NoError(t, err)
		require.True(t, can)

		require.NoError(t, tp.Update(s, writer))
		db.AssertExists(t, "task_positions", map[string]interface{}{
			"task_id":         task.ID,
			"project_view_id": view.ID,
			"position":        123,
		}, false)
	})

	t.Run("view of a foreign project is denied and writes no row", func(t *testing.T) {
		s, task, _, otherView, writer := setup(t)
		defer s.Close()

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: otherView.ID, Position: 123}
		can, err := tp.CanUpdate(s, writer)
		require.NoError(t, err)
		require.False(t, can, "a view of a project the caller cannot access must be rejected")

		db.AssertMissing(t, "task_positions", map[string]interface{}{
			"task_id":         task.ID,
			"project_view_id": otherView.ID,
		})
	})

	t.Run("saved filter: matching task of the owner is allowed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner := &user.User{ID: 1}
		proj := &Project{Title: "position-validation filter project"}
		require.NoError(t, proj.Create(s, owner))
		task := &Task{Title: "position-validation filter task", ProjectID: proj.ID}
		require.NoError(t, task.Create(s, owner))

		sf := &SavedFilter{
			Title:   "position-validation filter",
			Filters: &TaskCollection{Filter: "title like \"position-validation filter task\""},
		}
		require.NoError(t, sf.Create(s, owner))
		require.NoError(t, sf.Update(s, owner))

		view := &ProjectView{}
		exists, err := s.Where("project_id = ?", getProjectIDFromSavedFilterID(sf.ID)).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 42}
		can, err := tp.CanUpdate(s, owner)
		require.NoError(t, err)
		require.True(t, can, "the owner may position a task which matches the filter")
	})

	t.Run("saved filter: non-owner is denied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner := &user.User{ID: 1}
		writer := &user.User{ID: 2}
		proj := &Project{Title: "position-validation foreign filter project"}
		require.NoError(t, proj.Create(s, owner))
		_, err := s.Insert(&ProjectUser{UserID: writer.ID, ProjectID: proj.ID, Permission: PermissionWrite})
		require.NoError(t, err)
		task := &Task{Title: "position-validation foreign filter task", ProjectID: proj.ID}
		require.NoError(t, task.Create(s, owner))

		sf := &SavedFilter{
			Title:   "position-validation foreign filter",
			Filters: &TaskCollection{Filter: "title like \"position-validation foreign filter task\""},
		}
		require.NoError(t, sf.Create(s, owner))
		require.NoError(t, sf.Update(s, owner))

		view := &ProjectView{}
		exists, err := s.Where("project_id = ?", getProjectIDFromSavedFilterID(sf.ID)).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 42}
		can, err := tp.CanUpdate(s, writer)
		require.NoError(t, err)
		require.False(t, can, "a saved-filter view of another user must be rejected")
	})

	t.Run("saved filter: nonmatching task is denied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner := &user.User{ID: 1}
		proj := &Project{Title: "position-validation mismatch project"}
		require.NoError(t, proj.Create(s, owner))
		task := &Task{Title: "position-validation mismatch task", ProjectID: proj.ID}
		require.NoError(t, task.Create(s, owner))

		sf := &SavedFilter{
			Title:   "position-validation mismatch filter",
			Filters: &TaskCollection{Filter: "title like \"something else entirely\""},
		}
		require.NoError(t, sf.Create(s, owner))
		require.NoError(t, sf.Update(s, owner))

		view := &ProjectView{}
		exists, err := s.Where("project_id = ?", getProjectIDFromSavedFilterID(sf.ID)).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 42}
		can, err := tp.CanUpdate(s, owner)
		require.NoError(t, err)
		require.False(t, can, "a task which does not match the filter must be rejected")
	})

	t.Run("saved filter: relative dates resolve in the owner's timezone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner := &user.User{ID: 1}
		// UTC+14 makes this match only when the owner's timezone is used.
		_, err := s.Cols("timezone").Where("id = ?", owner.ID).Update(&user.User{Timezone: "Pacific/Kiritimati"})
		require.NoError(t, err)

		proj := &Project{Title: "position-validation tz project"}
		require.NoError(t, proj.Create(s, owner))
		task := &Task{
			Title:     "position-validation tz task",
			ProjectID: proj.ID,
			DueDate:   time.Date(2026, 8, 29, 23, 30, 0, 0, time.UTC),
		}
		require.NoError(t, task.Create(s, owner))

		sf := &SavedFilter{
			Title:   "position-validation tz filter",
			Filters: &TaskCollection{Filter: "due_date > 2026-08-30"},
		}
		require.NoError(t, sf.Create(s, owner))
		require.NoError(t, sf.Update(s, owner))

		view := &ProjectView{}
		exists, err := s.Where("project_id = ?", getProjectIDFromSavedFilterID(sf.ID)).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 42}
		can, err := tp.CanUpdate(s, owner)
		require.NoError(t, err)
		require.True(t, can,
			"the filter must be evaluated with the owner's timezone, matching the view the user actually sees")
	})

	t.Run("saved filter: owner lookup errors are returned", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		owner := &user.User{ID: 1}
		proj := &Project{Title: "position-validation missing owner"}
		require.NoError(t, proj.Create(s, owner))
		task := &Task{Title: "position-validation missing owner task", ProjectID: proj.ID}
		require.NoError(t, task.Create(s, owner))

		sf := &SavedFilter{
			Title:   "position-validation missing owner filter",
			Filters: &TaskCollection{Filter: "title like \"position-validation missing owner task\""},
		}
		require.NoError(t, sf.Create(s, owner))
		require.NoError(t, sf.Update(s, owner))

		view := &ProjectView{}
		exists, err := s.Where("project_id = ?", getProjectIDFromSavedFilterID(sf.ID)).Get(view)
		require.NoError(t, err)
		require.True(t, exists)

		_, err = s.ID(owner.ID).Delete(&user.User{})
		require.NoError(t, err)

		tp := &TaskPosition{TaskID: task.ID, ProjectViewID: view.ID, Position: 42}
		can, err := tp.CanUpdate(s, owner)
		require.Error(t, err)
		require.False(t, can)
	})
}
