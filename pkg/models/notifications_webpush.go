// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package models

import (
	"fmt"
	"strconv"

	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
)

func taskPushMessage(body string, taskID int64) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{Title: "Vikunja", Body: body, URL: "/tasks/" + strconv.FormatInt(taskID, 10)}
}

func (n *ReminderDueNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return taskPushMessage(n.ToTitle(lang), n.Task.ID)
}

func (n *TaskCommentNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return taskPushMessage(n.ToTitle(lang), n.Task.ID)
}

func (n *TaskAssignedNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return taskPushMessage(n.ToTitle(lang), n.Task.ID)
}

func (n *TaskDeletedNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{Title: "Vikunja", Body: n.ToTitle(lang), URL: "/"}
}

func (n *ProjectCreatedNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{Title: "Vikunja", Body: n.ToTitle(lang), URL: fmt.Sprintf("/projects/%d", n.Project.ID)}
}

func (n *TeamMemberAddedNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{Title: "Vikunja", Body: n.ToTitle(lang), URL: fmt.Sprintf("/teams/%d/edit", n.Team.ID)}
}

func (n *UserMentionedInTaskNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return taskPushMessage(n.ToTitle(lang), n.Task.ID)
}

func (n *APITokenExpiringWeekNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{Title: "Vikunja", Body: n.ToTitle(lang), URL: "/user/settings/api-tokens"}
}

func (n *APITokenExpiringDayNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{Title: "Vikunja", Body: n.ToTitle(lang), URL: "/user/settings/api-tokens"}
}

func (n *UndoneTaskOverdueNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return taskPushMessage(i18n.T(lang, "notifications.task.overdue.subject", n.Task.Title, n.Project.Title), n.Task.ID)
}

func (n *UndoneTasksOverdueNotification) ToWebPush(lang string) *notifications.WebPushMessage {
	return &notifications.WebPushMessage{
		Title: "Vikunja",
		Body:  i18n.T(lang, "notifications.web_push.overdue_count", len(n.Tasks)),
		URL:   "/",
	}
}
