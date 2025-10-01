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

package services

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskService_Create_SetsServiceLayerBehavior(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	u := &user.User{ID: 1}

	dueDate := time.Date(2025, time.January, 10, 12, 0, 0, 0, time.UTC)

	task := &models.Task{
		Title:       "Service layer task",
		Description: "Created via TaskService",
		ProjectID:   1,
		DueDate:     dueDate,
		Assignees: []*user.User{
			{ID: 1},
		},
		Reminders: []*models.TaskReminder{
			{
				RelativeTo:     models.ReminderRelationDueDate,
				RelativePeriod: -1800,
			},
			{
				Reminder: dueDate.Add(30 * time.Minute),
			},
		},
		IsFavorite: true,
	}

	createdTask, err := ts.Create(s, task, u)
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	assert.Equal(t, u.ID, createdTask.CreatedByID)
	require.NotNil(t, createdTask.CreatedBy)
	assert.Equal(t, u.ID, createdTask.CreatedBy.ID)
	assert.Equal(t, "user1", createdTask.CreatedBy.Username)

	assert.NotEmpty(t, createdTask.UID)
	assert.NotEmpty(t, createdTask.Identifier)
	assert.True(t, createdTask.ID > 0)
	assert.NotZero(t, createdTask.Index)

	require.NotNil(t, createdTask.Assignees)
	assert.Equal(t, u.ID, createdTask.Assignees[0].ID)

	require.Len(t, createdTask.Reminders, 2)
	assert.True(t, createdTask.Reminders[0].Reminder.Before(createdTask.Reminders[1].Reminder))

	require.NoError(t, s.Commit())

	db.AssertExists(t, "favorites", map[string]interface{}{
		"entity_id": createdTask.ID,
		"user_id":   u.ID,
		"kind":      models.FavoriteKindTask,
	}, false)
}

func TestTaskService_Create_PermissionDenied(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	unauthorized := &user.User{ID: 2}

	task := &models.Task{
		Title:     "Should fail",
		ProjectID: 1,
	}

	_, err := ts.Create(s, task, unauthorized)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestTaskService_CreateWithoutPermissionCheck_AllowsBypass(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(testEngine)
	unauthorized := &user.User{ID: 2}

	task := &models.Task{
		Title:     "Bypass permissions",
		ProjectID: 1,
	}

	createdTask, err := ts.CreateWithoutPermissionCheck(s, task, unauthorized)
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	assert.Equal(t, unauthorized.ID, createdTask.CreatedByID)
	require.NotNil(t, createdTask.CreatedBy)
	assert.Equal(t, unauthorized.ID, createdTask.CreatedBy.ID)

	require.NoError(t, s.Commit())
}
