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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projectWithTask(task models.Task) []*models.ProjectWithTasksAndBuckets {
	return []*models.ProjectWithTasksAndBuckets{
		{
			Project: models.Project{Title: "p"},
			Tasks: []*models.TaskWithComments{
				{Task: task},
			},
		},
	}
}

func TestCheckItemLimit(t *testing.T) {
	config.MigrationMaxItems.Set(10)
	t.Cleanup(func() { config.MigrationMaxItems.Set(100000) })
	var limitErr *ErrImportItemLimitExceeded

	t.Run("counts nested related tasks", func(t *testing.T) {
		nested := &models.Task{
			Title: "nested",
			RelatedTasks: models.RelatedTaskMap{
				models.RelationKindSubtask: {
					{Title: "a"},
					{Title: "b"},
				},
			},
		}
		root := models.Task{
			Title: "root",
			RelatedTasks: models.RelatedTaskMap{
				models.RelationKindSubtask: {
					nested,
					{Title: "c"},
				},
			},
		}
		// 2 (project, task) + 2*2 (relation, task) for root + 2*2 for nested = 10
		require.NoError(t, CheckItemLimit(projectWithTask(root)))

		nested.RelatedTasks[models.RelationKindSubtask] = append(nested.RelatedTasks[models.RelationKindSubtask], &models.Task{Title: "d"})
		require.ErrorAs(t, CheckItemLimit(projectWithTask(root)), &limitErr)
	})

	t.Run("does not count references to queued tasks as new tasks", func(t *testing.T) {
		str := projectWithTask(models.Task{
			ID:    1,
			Title: "root",
			RelatedTasks: models.RelatedTaskMap{
				models.RelationKindRelated: {
					{ID: 2},
					{ID: 2},
					{ID: 2},
				},
			},
		})
		str[0].Tasks = append(str[0].Tasks, &models.TaskWithComments{Task: models.Task{ID: 2, Title: "other"}})
		// 3 (project, 2 tasks) + 3 relations
		require.NoError(t, CheckItemLimit(str))
	})

	t.Run("counts labels", func(t *testing.T) {
		labels := make([]*models.Label, 9)
		for i := range labels {
			labels[i] = &models.Label{Title: "l"}
		}
		require.ErrorAs(t, CheckItemLimit(projectWithTask(models.Task{
			Title:  "root",
			Labels: labels,
		})), &limitErr)
	})

	t.Run("counts assignees", func(t *testing.T) {
		assignees := make([]*user.User, 9)
		for i := range assignees {
			assignees[i] = &user.User{}
		}
		// project, task, 8 assignees
		require.NoError(t, CheckItemLimit(projectWithTask(models.Task{Title: "root", Assignees: assignees[:8]})))
		require.ErrorAs(t, CheckItemLimit(projectWithTask(models.Task{Title: "root", Assignees: assignees})), &limitErr)
	})

	t.Run("counts reminders of related tasks", func(t *testing.T) {
		reminders := make([]*models.TaskReminder, 7)
		for i := range reminders {
			reminders[i] = &models.TaskReminder{}
		}
		related := &models.Task{Title: "related", Reminders: reminders}
		root := models.Task{
			Title:        "root",
			RelatedTasks: models.RelatedTaskMap{models.RelationKindSubtask: {related}},
		}
		// project, task, relation, related task, 6 reminders
		related.Reminders = reminders[:6]
		require.NoError(t, CheckItemLimit(projectWithTask(root)))
		related.Reminders = reminders
		require.ErrorAs(t, CheckItemLimit(projectWithTask(root)), &limitErr)
	})

	t.Run("terminates on a related task cycle", func(t *testing.T) {
		a := &models.Task{Title: "a"}
		b := &models.Task{
			Title: "b",
			RelatedTasks: models.RelatedTaskMap{
				models.RelationKindRelated: {a},
			},
		}
		a.RelatedTasks = models.RelatedTaskMap{
			models.RelationKindRelated: {b},
		}
		assert.ErrorAs(t, CheckItemLimit(projectWithTask(*a)), &limitErr)
	})
}
