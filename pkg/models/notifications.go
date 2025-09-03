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
	"fmt"
	"sort"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
)

// getDoerAvatarDataURI returns the avatar data URI for a user, for use in email headers.
func getDoerAvatarDataURI(doer *user.User) string {
	provider := avatar.GetProvider(doer)
	dataURI, err := provider.AsDataURI(doer, 20)
	if err != nil {
		return ""
	}
	return dataURI
}

// getThreadID generates a Message-ID format thread ID for a task
func getThreadID(taskID int64) string {
	return fmt.Sprintf("<task-%d@%s>", taskID, mail.GetMailDomain())
}

// ReminderDueNotification represents a ReminderDueNotification notification
type ReminderDueNotification struct {
	User         *user.User    `json:"user,omitempty"`
	Task         *Task         `json:"task"`
	Project      *Project      `json:"project"`
	TaskReminder *TaskReminder `json:"reminder"`
}

// ToMail returns the mail notification for ReminderDueNotification
func (n *ReminderDueNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		IncludeLinkToSettings(lang).
		To(n.User.Email).
		Subject(i18n.T(lang, "notifications.task.reminder.subject", n.Task.Title, n.Project.Title)).
		Greeting(i18n.T(lang, "notifications.greeting", n.User.GetName())).
		Line(i18n.T(lang, "notifications.task.reminder.message", n.Task.Title, n.Project.Title)).
		Action(i18n.T(lang, "notifications.common.actions.open_task"), config.ServicePublicURL.GetString()+"tasks/"+strconv.FormatInt(n.Task.ID, 10)).
		Line(i18n.T(lang, "notifications.common.have_nice_day"))
}

// ToDB returns the ReminderDueNotification notification in a format which can be saved in the db
func (n *ReminderDueNotification) ToDB() interface{} {
	return &ReminderDueNotification{
		Task:    n.Task,
		Project: n.Project,
	}
}

// Name returns the name of the notification
func (n *ReminderDueNotification) Name() string {
	return "task.reminder"
}

// ThreadID returns the thread ID for email threading
func (n *ReminderDueNotification) ThreadID() string {
	return getThreadID(n.Task.ID)
}

// TaskCommentNotification represents a TaskCommentNotification notification
type TaskCommentNotification struct {
	Doer      *user.User   `json:"doer"`
	Task      *Task        `json:"task"`
	Comment   *TaskComment `json:"comment"`
	Mentioned bool         `json:"mentioned"`
	Project   *Project     `json:"project"`
}

func (n *TaskCommentNotification) SubjectID() int64 {
	return n.Comment.ID
}

// ToMail returns the mail notification for TaskCommentNotification
func (n *TaskCommentNotification) ToMail(lang string) *notifications.Mail {
	s := db.NewSession()
	defer s.Close()
	formattedComment := formatMentionsForEmail(s, n.Comment.Comment)

	mail := notifications.NewMail().
		Conversational().
		From(n.Doer.GetNameAndFromEmail()).
		Subject(i18n.T(lang, "notifications.task.comment.subject", n.Task.Title, n.Task.GetFullIdentifier()))

	// Add header line
	action := i18n.T(lang, "notifications.common.actions.left_comment", n.Doer.GetName())
	if n.Mentioned {
		action = i18n.T(lang, "notifications.common.actions.mentioned_you_comment", n.Doer.GetName())
		mail.Subject(i18n.T(lang, "notifications.task.comment.mentioned_subject", n.Doer.GetName(), n.Task.Title, n.Task.GetFullIdentifier()))
	}

	headerLine := notifications.CreateConversationalHeader(
		getDoerAvatarDataURI(n.Doer),
		action,
		n.Task.GetFrontendURL(),
		n.Project.Title,
		n.Task.GetFullIdentifier(),
		n.Task.Title,
	)
	mail.HeaderLine(headerLine)

	// Add the actual comment content wrapped in a div for consistent spacing
	mail.HTML(formattedComment)

	return mail.
		Action(i18n.T(lang, "notifications.common.actions.open_task"), n.Task.GetFrontendURL()).
		IncludeLinkToSettings(lang)
}

// ToDB returns the TaskCommentNotification notification in a format which can be saved in the db
func (n *TaskCommentNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TaskCommentNotification) Name() string {
	return "task.comment"
}

// ThreadID returns the thread ID for email threading
func (n *TaskCommentNotification) ThreadID() string {
	return getThreadID(n.Task.ID)
}

// TaskAssignedNotification represents a TaskAssignedNotification notification
type TaskAssignedNotification struct {
	Doer     *user.User `json:"doer"`
	Task     *Task      `json:"task"`
	Assignee *user.User `json:"assignee"`
	Target   *user.User `json:"-"`
	Project  *Project   `json:"project"`
}

// ToMail returns the mail notification for TaskAssignedNotification
func (n *TaskAssignedNotification) ToMail(lang string) *notifications.Mail {
	if n.Target.ID == n.Assignee.ID {
		// Notification to the assignee
		return notifications.NewMail().
			From(n.Doer.GetNameAndFromEmail()).
			Subject(i18n.T(lang, "notifications.task.assigned.subject_to_assignee", n.Task.Title, n.Task.GetFullIdentifier())).
			Greeting(i18n.T(lang, "notifications.greeting", n.Target.GetName())).
			Line(i18n.T(lang, "notifications.task.assigned.message_to_assignee", n.Doer.GetName(), n.Task.Title)).
			Action(i18n.T(lang, "notifications.common.actions.open_task"), n.Task.GetFrontendURL()).
			IncludeLinkToSettings(lang)
	}

	// Check if the doer assigned the task to themselves
	if n.Doer.ID == n.Assignee.ID {
		return notifications.NewMail().
			From(n.Doer.GetNameAndFromEmail()).
			Subject(i18n.T(lang, "notifications.task.assigned.subject_to_others_self", n.Task.Title, n.Task.GetFullIdentifier(), n.Doer.GetName())).
			Greeting(i18n.T(lang, "notifications.greeting", n.Target.GetName())).
			Line(i18n.T(lang, "notifications.task.assigned.message_to_others_self", n.Doer.GetName())).
			Action(i18n.T(lang, "notifications.common.actions.open_task"), n.Task.GetFrontendURL()).
			IncludeLinkToSettings(lang)
	}

	// Notification to others about assignment
	return notifications.NewMail().
		From(n.Doer.GetNameAndFromEmail()).
		Subject(i18n.T(lang, "notifications.task.assigned.subject_to_others", n.Task.Title, n.Task.GetFullIdentifier(), n.Assignee.GetName())).
		Greeting(i18n.T(lang, "notifications.greeting", n.Target.GetName())).
		Line(i18n.T(lang, "notifications.task.assigned.message_to_others", n.Doer.GetName(), n.Assignee.GetName())).
		Action(i18n.T(lang, "notifications.common.actions.open_task"), n.Task.GetFrontendURL()).
		IncludeLinkToSettings(lang)
}

// ToDB returns the TaskAssignedNotification notification in a format which can be saved in the db
func (n *TaskAssignedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TaskAssignedNotification) Name() string {
	return "task.assigned"
}

// ThreadID returns the thread ID for email threading
func (n *TaskAssignedNotification) ThreadID() string {
	return getThreadID(n.Task.ID)
}

// TaskDeletedNotification represents a TaskDeletedNotification notification
type TaskDeletedNotification struct {
	Doer *user.User `json:"doer"`
	Task *Task      `json:"task"`
}

// ToMail returns the mail notification for TaskDeletedNotification
func (n *TaskDeletedNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.task.deleted.subject", n.Task.Title, n.Task.GetFullIdentifier())).
		Line(i18n.T(lang, "notifications.task.deleted.message", n.Doer.GetName(), n.Task.Title, n.Task.GetFullIdentifier()))
}

// ToDB returns the TaskDeletedNotification notification in a format which can be saved in the db
func (n *TaskDeletedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TaskDeletedNotification) Name() string {
	return "task.deleted"
}

// ThreadID returns the thread ID for email threading
func (n *TaskDeletedNotification) ThreadID() string {
	return getThreadID(n.Task.ID)
}

// ProjectCreatedNotification represents a ProjectCreatedNotification notification
type ProjectCreatedNotification struct {
	Doer    *user.User `json:"doer"`
	Project *Project   `json:"project"`
}

// ToMail returns the mail notification for ProjectCreatedNotification
func (n *ProjectCreatedNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.project.created", n.Doer.GetName(), n.Project.Title)).
		Line(i18n.T(lang, "notifications.project.created", n.Doer.GetName(), n.Project.Title)).
		Action(i18n.T(lang, "notifications.common.actions.open_project"), config.ServicePublicURL.GetString()+"projects/")
}

// ToDB returns the ProjectCreatedNotification notification in a format which can be saved in the db
func (n *ProjectCreatedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *ProjectCreatedNotification) Name() string {
	return "project.created"
}

// TeamMemberAddedNotification represents a TeamMemberAddedNotification notification
type TeamMemberAddedNotification struct {
	Member *user.User `json:"member"`
	Doer   *user.User `json:"doer"`
	Team   *Team      `json:"team"`
}

// ToMail returns the mail notification for TeamMemberAddedNotification
func (n *TeamMemberAddedNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.team.member_added.subject", n.Doer.GetName(), n.Team.Name)).
		From(n.Doer.GetNameAndFromEmail()).
		Greeting(i18n.T(lang, "notifications.greeting", n.Member.GetName())).
		Line(i18n.T(lang, "notifications.team.member_added.message", n.Doer.GetName(), n.Team.Name)).
		Action(i18n.T(lang, "notifications.common.actions.open_team"), config.ServicePublicURL.GetString()+"teams/"+strconv.FormatInt(n.Team.ID, 10)+"/edit")
}

// ToDB returns the TeamMemberAddedNotification notification in a format which can be saved in the db
func (n *TeamMemberAddedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TeamMemberAddedNotification) Name() string {
	return "team.member.added"
}

func getOverdueSinceString(until time.Duration, language string) (overdueSince string) {
	if until == 0 {
		return i18n.T(language, "notifications.task.overdue.overdue_now")
	}

	return i18n.T(language, "notifications.task.overdue.overdue_since", utils.HumanizeDuration(until, language))
}

// UndoneTaskOverdueNotification represents a UndoneTaskOverdueNotification notification
type DailyTasksReminderNotification struct {
	User         *user.User
	OverdueTasks map[int64]*Task
	DueToday     map[int64]*Task
	Projects     map[int64]*Project
}

// ToMail returns the mail notification for DailyTasksReminderNotification
func (n *DailyTasksReminderNotification) ToMail(lang string) *notifications.Mail {

	sortedOverdue := make([]*Task, 0, len(n.OverdueTasks))
	for _, task := range n.OverdueTasks {
		sortedOverdue = append(sortedOverdue, task)
	}
	sort.Slice(sortedOverdue, func(i, j int) bool {
		return sortedOverdue[i].DueDate.Before(sortedOverdue[j].DueDate)
	})

	sortedToday := make([]*Task, 0, len(n.DueToday))
	for _, task := range n.DueToday {
		sortedToday = append(sortedToday, task)
	}
	sort.Slice(sortedToday, func(i, j int) bool {
		return sortedToday[i].DueDate.Before(sortedToday[j].DueDate)
	})

	overdueLine := ""
	for _, task := range sortedOverdue {
		until := time.Until(task.DueDate).Round(1*time.Hour) * -1
		overdueLine += `* [` + task.Title + `](` + config.ServicePublicURL.GetString() + "tasks/" + strconv.FormatInt(task.ID, 10) + `) (` + n.Projects[task.ProjectID].Title + `), ` + i18n.T("notifications.task.overdue.overdue", getOverdueSinceString(until, n.User.Language)) + "\n"
	}

	todayLine := ""
	for _, task := range sortedToday {
		todayLine += `* [` + task.Title + `](` + config.ServicePublicURL.GetString() + "tasks/" + strconv.FormatInt(task.ID, 10) + `) (` + n.Projects[task.ProjectID].Title + `)\n`
	}

	subject := i18n.T(lang, "notifications.task.overdue.multiple_subject")
	if len(n.OverdueTasks) == 0 {
		subject = i18n.T(lang, "notifications.task.reminder.only_due_today_subject")
	}

	m := notifications.NewMail().
		IncludeLinkToSettings(lang).
		Subject(subject).
		Greeting(i18n.T(lang, "notifications.greeting", n.User.GetName()))

	if overdueLine != "" {
		m.Line(i18n.T(lang, "notifications.task.reminder.overdue_intro"))
		m.Line(overdueLine)
	}
	if todayLine != "" {
		m.Line(i18n.T(lang, "notifications.task.reminder.today_intro"))
		m.Line(todayLine)
	}

	m.Action(i18n.T(lang, "notifications.common.actions.open_vikunja"), config.ServicePublicURL.GetString()).
		Line(i18n.T(lang, "notifications.common.have_nice_day"))
	return m
}

// ToDB returns the DailyTasksReminderNotification notification in a format which can be saved in the db
func (n *DailyTasksReminderNotification) ToDB() interface{} { return nil }

// Name returns the name of the notification
func (n *DailyTasksReminderNotification) Name() string { return "task.daily.reminder" }

// UserMentionedInTaskNotification represents a UserMentionedInTaskNotification notification
type UserMentionedInTaskNotification struct {
	Doer    *user.User `json:"doer"`
	Task    *Task      `json:"task"`
	IsNew   bool       `json:"is_new"`
	Project *Project   `json:"project"`
}

func (n *UserMentionedInTaskNotification) SubjectID() int64 {
	return n.Task.ID
}

// ToMail returns the mail notification for UserMentionedInTaskNotification
func (n *UserMentionedInTaskNotification) ToMail(lang string) *notifications.Mail {
	s := db.NewSession()
	defer s.Close()
	formattedDescription := formatMentionsForEmail(s, n.Task.Description)

	var subject string
	if n.IsNew {
		subject = i18n.T(lang, "notifications.task.mentioned.subject_new", n.Doer.GetName(), n.Task.Title, n.Task.GetFullIdentifier())
	} else {
		subject = i18n.T(lang, "notifications.task.mentioned.subject", n.Doer.GetName(), n.Task.Title, n.Task.GetFullIdentifier())
	}

	mail := notifications.NewMail().
		Conversational().
		From(n.Doer.GetNameAndFromEmail()).
		Subject(subject)

	// Add header line
	action := i18n.T(lang, "notifications.common.actions.mentioned_you", n.Doer.GetName())
	if n.IsNew {
		action = i18n.T(lang, "notifications.common.actions.mentioned_you_new_task", n.Doer.GetName())
	}

	headerLine := notifications.CreateConversationalHeader(
		getDoerAvatarDataURI(n.Doer),
		action,
		n.Task.GetFrontendURL(),
		n.Project.Title,
		n.Task.GetFullIdentifier(),
		n.Task.Title,
	)
	mail.HeaderLine(headerLine)

	if formattedDescription != "" {
		mail.HTML(formattedDescription)
	}

	return mail.
		Action(i18n.T(lang, "notifications.common.actions.open_task"), n.Task.GetFrontendURL()).
		IncludeLinkToSettings(lang)
}

// ToDB returns the UserMentionedInTaskNotification notification in a format which can be saved in the db
func (n *UserMentionedInTaskNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *UserMentionedInTaskNotification) Name() string {
	return "task.mentioned"
}

// ThreadID returns the thread ID for email threading
func (n *UserMentionedInTaskNotification) ThreadID() string {
	return getThreadID(n.Task.ID)
}

// DataExportReadyNotification represents a DataExportReadyNotification notification
type DataExportReadyNotification struct {
	User *user.User `json:"user"`
}

// ToMail returns the mail notification for DataExportReadyNotification
func (n *DataExportReadyNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.data_export.ready.subject")).
		Greeting(i18n.T(lang, "notifications.greeting", n.User.GetName())).
		Line(i18n.T(lang, "notifications.data_export.ready.message")).
		Action(i18n.T(lang, "notifications.common.actions.download"), config.ServicePublicURL.GetString()+"user/export/download").
		Line(i18n.T(lang, "notifications.data_export.ready.availability")).
		Line(i18n.T(lang, "notifications.common.have_nice_day"))
}

// ToDB returns the DataExportReadyNotification notification in a format which can be saved in the db
func (n *DataExportReadyNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *DataExportReadyNotification) Name() string {
	return "data.export.ready"
}
