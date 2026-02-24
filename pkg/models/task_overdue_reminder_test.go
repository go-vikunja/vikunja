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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestGetUndoneOverDueTasks(t *testing.T) {
	t.Run("no undone tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-01-01T01:13:00Z")
		require.NoError(t, err)
		tasks, err := getUndoneOverdueTasks(s, now, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
	t.Run("undone overdue", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-12-01T09:00:00Z")
		require.NoError(t, err)
		uts, err := getUndoneOverdueTasks(s, now, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)
		require.Len(t, uts, 1)
		assert.Len(t, uts[1].tasks, 2)
		// The tasks don't always have the same order, so we only check their presence, not their position.
		var task5Present bool
		var task6Present bool
		for _, t := range uts[1].tasks {
			if t.ID == 5 {
				task5Present = true
			}
			if t.ID == 6 {
				task6Present = true
			}
		}
		assert.Truef(t, task5Present, "expected task 5 to be present but was not")
		assert.Truef(t, task6Present, "expected task 6 to be present but was not")
	})
	t.Run("done overdue", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		now, err := time.Parse(time.RFC3339Nano, "2018-11-01T01:13:00Z")
		require.NoError(t, err)
		tasks, err := getUndoneOverdueTasks(s, now, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
}

func TestGetTaskUsersForTasksPermissionFiltering(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	usersWithAccess, err := getTaskUsersForTasks(s, []int64{35}, nil)
	require.NoError(t, err)

	var hasAssignee bool
	for _, tu := range usersWithAccess {
		if tu.User.ID == 2 {
			hasAssignee = true
			break
		}
	}
	assert.True(t, hasAssignee)
}

// Tests for issue #1581: Assignees not receiving overdue notifications
func TestOverdueTaskNotificationsForAssignees(t *testing.T) {
	t.Run("assignee with direct project share receives overdue notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create an overdue task in project 1 (owned by user 1)
		overdueTime, err := time.Parse(time.RFC3339, "2018-11-30T10:00:00Z")
		require.NoError(t, err)

		task := &Task{
			Title:       "Overdue task assigned to user with direct share",
			Done:        false,
			CreatedByID: 1, // Admin/creator
			ProjectID:   1,
			DueDate:     overdueTime,
		}
		err = task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Give user 2 direct read/write access to project 1
		projectUser := &ProjectUser{
			UserID:     2,
			ProjectID:  1,
			Permission: PermissionWrite, // Read/write access
		}
		_, err = s.Insert(projectUser)
		require.NoError(t, err)

		// Assign task to user 2
		assignee := &TaskAssginee{
			TaskID: task.ID,
			UserID: 2,
		}
		_, err = s.Insert(assignee)
		require.NoError(t, err)

		// Subscribe user 2 (simulating auto-subscribe on assignment)
		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   task.ID,
			UserID:     2,
		}
		_, err = s.Insert(subscription)
		require.NoError(t, err)

		// Get users who should receive overdue notifications
		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers, "should have users for overdue task")

		// Verify both creator (user 1) and assignee (user 2) are included
		var hasCreator bool
		var hasAssignee bool
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == task.ID {
				hasCreator = true
			}
			if tu.User.ID == 2 && tu.Task.ID == task.ID {
				hasAssignee = true
			}
		}

		assert.True(t, hasCreator, "task creator should receive overdue notification")
		assert.True(t, hasAssignee, "assignee with direct project share should receive overdue notification")
	})

	t.Run("assignee with team access receives overdue notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create an overdue task in project 1 (owned by user 1)
		overdueTime, err := time.Parse(time.RFC3339, "2018-11-30T10:00:00Z")
		require.NoError(t, err)

		task := &Task{
			Title:       "Overdue task assigned to user with team access",
			Done:        false,
			CreatedByID: 1, // Admin/creator
			ProjectID:   1,
			DueDate:     overdueTime,
		}
		err = task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Create a team and add user 2 to it
		team := &Team{
			Name: "Test Team for Issue 1581",
		}
		_, err = s.Insert(team)
		require.NoError(t, err)

		teamMember := &TeamMember{
			TeamID: team.ID,
			UserID: 2,
		}
		_, err = s.Insert(teamMember)
		require.NoError(t, err)

		// Share project 1 with the team (read permission)
		teamProject := &TeamProject{
			TeamID:     team.ID,
			ProjectID:  1,
			Permission: PermissionRead,
		}
		_, err = s.Insert(teamProject)
		require.NoError(t, err)

		// Assign task to user 2
		assignee := &TaskAssginee{
			TaskID: task.ID,
			UserID: 2,
		}
		_, err = s.Insert(assignee)
		require.NoError(t, err)

		// Subscribe user 2 (simulating auto-subscribe on assignment)
		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   task.ID,
			UserID:     2,
		}
		_, err = s.Insert(subscription)
		require.NoError(t, err)

		// Get users who should receive overdue notifications
		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers, "should have users for overdue task")

		// Verify both creator (user 1) and assignee (user 2) are included
		var hasCreator bool
		var hasAssignee bool
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == task.ID {
				hasCreator = true
			}
			if tu.User.ID == 2 && tu.Task.ID == task.ID {
				hasAssignee = true
			}
		}

		assert.True(t, hasCreator, "task creator should receive overdue notification")
		assert.True(t, hasAssignee, "assignee with team access should receive overdue notification")
	})

	t.Run("unassigned user who lost team access does not receive notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create an overdue task
		overdueTime, err := time.Parse(time.RFC3339, "2018-11-30T10:00:00Z")
		require.NoError(t, err)

		task := &Task{
			Title:       "Overdue task - user lost access",
			Done:        false,
			CreatedByID: 1,
			ProjectID:   1,
			DueDate:     overdueTime,
		}
		err = task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// User 13 was previously subscribed but no longer has access
		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   task.ID,
			UserID:     13, // User 13 has no access to project 1
		}
		_, err = s.Insert(subscription)
		require.NoError(t, err)

		// Get users who should receive overdue notifications
		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)

		// Verify user 13 is NOT included (no access)
		var hasUser13 bool
		for _, tu := range taskUsers {
			if tu.User.ID == 13 && tu.Task.ID == task.ID {
				hasUser13 = true
			}
		}

		assert.False(t, hasUser13, "user without project access should not receive notification")
	})
}

func TestOverdueTaskNotificationsIncludeSubscribers(t *testing.T) {
	t.Run("subscriber with access receives overdue notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create an overdue task in project 1 (owned by user 1)
		overdueTime, err := time.Parse(time.RFC3339, "2018-11-30T10:00:00Z")
		require.NoError(t, err)

		task := &Task{
			Title:       "Overdue task with subscriber",
			Done:        false,
			CreatedByID: 1,
			ProjectID:   1,
			DueDate:     overdueTime,
		}
		err = task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Subscribe user 2 to this task
		// User 2 needs access to project 1 first - let's give them read access
		projectUser := &ProjectUser{
			UserID:     2,
			ProjectID:  1,
			Permission: PermissionRead,
		}
		_, err = s.Insert(projectUser)
		require.NoError(t, err)

		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   task.ID,
			UserID:     2,
		}
		_, err = s.Insert(subscription)
		require.NoError(t, err)

		// Get users who should receive overdue notifications
		// Use the same condition as the actual overdue reminder code
		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)
		require.NotEmpty(t, taskUsers, "should have users for overdue task")

		// Verify both creator (user 1) and subscriber (user 2) are included
		var hasCreator bool
		var hasSubscriber bool
		for _, tu := range taskUsers {
			if tu.User.ID == 1 && tu.Task.ID == task.ID {
				hasCreator = true
			}
			if tu.User.ID == 2 && tu.Task.ID == task.ID {
				hasSubscriber = true
			}
		}

		assert.True(t, hasCreator, "task creator should receive overdue notification")
		assert.True(t, hasSubscriber, "task subscriber should receive overdue notification")
	})

	t.Run("subscriber without overdue reminders enabled does not receive notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create an overdue task
		overdueTime, err := time.Parse(time.RFC3339, "2018-11-30T10:00:00Z")
		require.NoError(t, err)

		task := &Task{
			Title:       "Overdue task with subscriber who disabled reminders",
			Done:        false,
			CreatedByID: 1,
			ProjectID:   1,
			DueDate:     overdueTime,
		}
		err = task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Give user 2 access to project 1
		projectUser := &ProjectUser{
			UserID:     2,
			ProjectID:  1,
			Permission: PermissionRead,
		}
		_, err = s.Insert(projectUser)
		require.NoError(t, err)

		// Subscribe user 2 to this task but disable their overdue reminders
		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   task.ID,
			UserID:     2,
		}
		_, err = s.Insert(subscription)
		require.NoError(t, err)

		// Disable overdue reminders for user 2
		_, err = s.Exec("UPDATE users SET overdue_tasks_reminders_enabled = false WHERE id = ?", 2)
		require.NoError(t, err)

		// Get users who should receive overdue notifications
		// Use the same condition as the actual overdue reminder code
		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)

		// Verify subscriber (user 2) is NOT included because they disabled reminders
		var hasSubscriber bool
		for _, tu := range taskUsers {
			if tu.User.ID == 2 && tu.Task.ID == task.ID {
				hasSubscriber = true
			}
		}

		assert.False(t, hasSubscriber, "subscriber with overdue reminders disabled should not receive notification")
	})

	t.Run("subscriber without project access does not receive notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create an overdue task
		overdueTime, err := time.Parse(time.RFC3339, "2018-11-30T10:00:00Z")
		require.NoError(t, err)

		task := &Task{
			Title:       "Overdue task with subscriber without access",
			Done:        false,
			CreatedByID: 1,
			ProjectID:   1,
			DueDate:     overdueTime,
		}
		err = task.Create(s, &user.User{ID: 1})
		require.NoError(t, err)

		// Subscribe user 13 who has NO access to project 1
		subscription := &Subscription{
			EntityType: SubscriptionEntityTask,
			EntityID:   task.ID,
			UserID:     13,
		}
		_, err = s.Insert(subscription)
		require.NoError(t, err)

		// Get users who should receive overdue notifications
		// Use the same condition as the actual overdue reminder code
		taskUsers, err := getTaskUsersForTasks(s, []int64{task.ID}, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		require.NoError(t, err)

		// Verify subscriber (user 13) is NOT included because they don't have project access
		var hasSubscriber bool
		for _, tu := range taskUsers {
			if tu.User.ID == 13 && tu.Task.ID == task.ID {
				hasSubscriber = true
			}
		}

		assert.False(t, hasSubscriber, "subscriber without project access should not receive notification")
	})
}
