// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistedNotificationsHaveWebPushPayloads(t *testing.T) {
	doer := &user.User{ID: 1, Username: "doer", Name: "Doer"}
	target := &user.User{ID: 2, Username: "target", Name: "Target"}
	task := &Task{ID: 42, Title: "Prepare report", Identifier: "PROJ-42"}
	project := &Project{ID: 7, Title: "Planning"}
	team := &Team{ID: 9, Name: "Operators"}
	token := &APIToken{ID: 11, Title: "Automation"}

	tests := []struct {
		name         string
		notification notifications.WebPushable
		url          string
	}{
		{"task reminder", &ReminderDueNotification{User: target, Task: task, Project: project}, "/tasks/42"},
		{"task comment", &TaskCommentNotification{Doer: doer, Task: task, Comment: &TaskComment{ID: 3}, Project: project}, "/tasks/42"},
		{"task assigned", &TaskAssignedNotification{Doer: doer, Task: task, Assignee: target, Target: target, Project: project}, "/tasks/42"},
		{"task deleted", &TaskDeletedNotification{Doer: doer, Task: task}, "/"},
		{"project created", &ProjectCreatedNotification{Doer: doer, Project: project}, "/projects/7"},
		{"team member added", &TeamMemberAddedNotification{Member: target, Doer: doer, Team: team}, "/teams/9/edit"},
		{"task mention", &UserMentionedInTaskNotification{Doer: doer, Task: task, Project: project}, "/tasks/42"},
		{"api token week", &APITokenExpiringWeekNotification{User: target, Token: token}, "/user/settings/api-tokens"},
		{"api token day", &APITokenExpiringDayNotification{User: target, Token: token}, "/user/settings/api-tokens"},
		{"single overdue", &UndoneTaskOverdueNotification{User: target, Task: task, Project: project}, "/tasks/42"},
		{"overdue summary", &UndoneTasksOverdueNotification{User: target, Tasks: map[int64]*Task{task.ID: task}, Projects: map[int64]*Project{project.ID: project}}, "/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := test.notification.ToWebPush("en")
			require.NotNil(t, message)
			assert.Equal(t, "Vikunja", message.Title)
			assert.NotEmpty(t, message.Body)
			assert.Equal(t, test.url, message.URL)
		})
	}
}
