// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"bufio"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/utils"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
)

// ReminderDueNotification represents a ReminderDueNotification notification
type ReminderDueNotification struct {
	User *user.User `json:"user"`
	Task *Task      `json:"task"`
}

// ToMail returns the mail notification for ReminderDueNotification
func (n *ReminderDueNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		To(n.User.Email).
		Subject(`Reminder for "`+n.Task.Title+`"`).
		Greeting("Hi "+n.User.GetName()+",").
		Line(`This is a friendly reminder of the task "`+n.Task.Title+`".`).
		Action("Open Task", config.ServiceFrontendurl.GetString()+"tasks/"+strconv.FormatInt(n.Task.ID, 10)).
		Line("Have a nice day!")
}

// ToDB returns the ReminderDueNotification notification in a format which can be saved in the db
func (n *ReminderDueNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *ReminderDueNotification) Name() string {
	return ""
}

// TaskCommentNotification represents a TaskCommentNotification notification
type TaskCommentNotification struct {
	Doer      *user.User   `json:"doer"`
	Task      *Task        `json:"task"`
	Comment   *TaskComment `json:"comment"`
	Mentioned bool         `json:"mentioned"`
}

func (n *TaskCommentNotification) SubjectID() int64 {
	return n.Comment.ID
}

// ToMail returns the mail notification for TaskCommentNotification
func (n *TaskCommentNotification) ToMail() *notifications.Mail {

	mail := notifications.NewMail().
		From(n.Doer.GetNameAndFromEmail())

	subject := "Re: " + n.Task.Title
	if n.Mentioned {
		subject = n.Doer.GetName() + ` mentioned you in a comment in "` + n.Task.Title + `"`
		mail.Line("**" + n.Doer.GetName() + "** mentioned you in a comment:")
	}

	mail.Subject(subject)

	lines := bufio.NewScanner(strings.NewReader(n.Comment.Comment))
	for lines.Scan() {
		mail.Line(lines.Text())
	}

	return mail.
		Action("View Task", n.Task.GetFrontendURL())
}

// ToDB returns the TaskCommentNotification notification in a format which can be saved in the db
func (n *TaskCommentNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TaskCommentNotification) Name() string {
	return "task.comment"
}

// TaskAssignedNotification represents a TaskAssignedNotification notification
type TaskAssignedNotification struct {
	Doer     *user.User `json:"doer"`
	Task     *Task      `json:"task"`
	Assignee *user.User `json:"assignee"`
}

// ToMail returns the mail notification for TaskAssignedNotification
func (n *TaskAssignedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject(n.Task.Title+"("+n.Task.GetFullIdentifier()+")"+" has been assigned to "+n.Assignee.GetName()).
		Line(n.Doer.GetName()+" has assigned this task to "+n.Assignee.GetName()+".").
		Action("View Task", n.Task.GetFrontendURL())
}

// ToDB returns the TaskAssignedNotification notification in a format which can be saved in the db
func (n *TaskAssignedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TaskAssignedNotification) Name() string {
	return "task.assigned"
}

// TaskDeletedNotification represents a TaskDeletedNotification notification
type TaskDeletedNotification struct {
	Doer *user.User `json:"doer"`
	Task *Task      `json:"task"`
}

// ToMail returns the mail notification for TaskDeletedNotification
func (n *TaskDeletedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject(n.Task.Title + "(" + n.Task.GetFullIdentifier() + ")" + " has been delete").
		Line(n.Doer.GetName() + " has deleted the task " + n.Task.Title + "(" + n.Task.GetFullIdentifier() + ")")
}

// ToDB returns the TaskDeletedNotification notification in a format which can be saved in the db
func (n *TaskDeletedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TaskDeletedNotification) Name() string {
	return "task.deleted"
}

// ListCreatedNotification represents a ListCreatedNotification notification
type ListCreatedNotification struct {
	Doer *user.User `json:"doer"`
	List *List      `json:"list"`
}

// ToMail returns the mail notification for ListCreatedNotification
func (n *ListCreatedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject(n.Doer.GetName()+` created the list "`+n.List.Title+`"`).
		Line(n.Doer.GetName()+` created the list "`+n.List.Title+`"`).
		Action("View List", config.ServiceFrontendurl.GetString()+"lists/")
}

// ToDB returns the ListCreatedNotification notification in a format which can be saved in the db
func (n *ListCreatedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *ListCreatedNotification) Name() string {
	return "list.created"
}

// TeamMemberAddedNotification represents a TeamMemberAddedNotification notification
type TeamMemberAddedNotification struct {
	Member *user.User `json:"member"`
	Doer   *user.User `json:"doer"`
	Team   *Team      `json:"team"`
}

// ToMail returns the mail notification for TeamMemberAddedNotification
func (n *TeamMemberAddedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject(n.Doer.GetName()+" added you to the "+n.Team.Name+" team in Vikunja").
		From(n.Doer.GetNameAndFromEmail()).
		Greeting("Hi "+n.Member.GetName()+",").
		Line(n.Doer.GetName()+" has just added you to the "+n.Team.Name+" team in Vikunja.").
		Action("View Team", config.ServiceFrontendurl.GetString()+"teams/"+strconv.FormatInt(n.Team.ID, 10)+"/edit")
}

// ToDB returns the TeamMemberAddedNotification notification in a format which can be saved in the db
func (n *TeamMemberAddedNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *TeamMemberAddedNotification) Name() string {
	return "team.member.added"
}

// UndoneTaskOverdueNotification represents a UndoneTaskOverdueNotification notification
type UndoneTaskOverdueNotification struct {
	User *user.User
	Task *Task
}

// ToMail returns the mail notification for UndoneTaskOverdueNotification
func (n *UndoneTaskOverdueNotification) ToMail() *notifications.Mail {
	until := time.Until(n.Task.DueDate).Round(1*time.Hour) * -1
	return notifications.NewMail().
		Subject(`Task "`+n.Task.Title+`" is overdue`).
		Greeting("Hi "+n.User.GetName()+",").
		Line(`This is a friendly reminder of the task "`+n.Task.Title+`" which is overdue since `+utils.HumanizeDuration(until)+` and not yet done.`).
		Action("Open Task", config.ServiceFrontendurl.GetString()+"tasks/"+strconv.FormatInt(n.Task.ID, 10)).
		Line("Have a nice day!")
}

// ToDB returns the UndoneTaskOverdueNotification notification in a format which can be saved in the db
func (n *UndoneTaskOverdueNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *UndoneTaskOverdueNotification) Name() string {
	return "task.undone.overdue"
}

// UndoneTasksOverdueNotification represents a UndoneTasksOverdueNotification notification
type UndoneTasksOverdueNotification struct {
	User  *user.User
	Tasks []*Task
}

// ToMail returns the mail notification for UndoneTasksOverdueNotification
func (n *UndoneTasksOverdueNotification) ToMail() *notifications.Mail {

	overdueLine := ""
	for _, task := range n.Tasks {
		until := time.Until(task.DueDate).Round(1*time.Hour) * -1
		overdueLine += `* [` + task.Title + `](` + config.ServiceFrontendurl.GetString() + "tasks/" + strconv.FormatInt(task.ID, 10) + `), overdue since ` + utils.HumanizeDuration(until) + "\n"
	}

	return notifications.NewMail().
		Subject(`Your overdue tasks`).
		Greeting("Hi "+n.User.GetName()+",").
		Line("You have the following overdue tasks:").
		Line(overdueLine).
		Action("Open Vikunja", config.ServiceFrontendurl.GetString()).
		Line("Have a nice day!")
}

// ToDB returns the UndoneTasksOverdueNotification notification in a format which can be saved in the db
func (n *UndoneTasksOverdueNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *UndoneTasksOverdueNotification) Name() string {
	return "task.undone.overdue"
}

// UserMentionedInTaskNotification represents a UserMentionedInTaskNotification notification
type UserMentionedInTaskNotification struct {
	Doer  *user.User `json:"doer"`
	Task  *Task      `json:"task"`
	IsNew bool       `json:"is_new"`
}

func (n *UserMentionedInTaskNotification) SubjectID() int64 {
	return n.Task.ID
}

// ToMail returns the mail notification for UserMentionedInTaskNotification
func (n *UserMentionedInTaskNotification) ToMail() *notifications.Mail {
	subject := n.Doer.GetName() + ` mentioned you in a new task "` + n.Task.Title + `"`
	if n.IsNew {
		subject = n.Doer.GetName() + ` mentioned you in a task "` + n.Task.Title + `"`
	}

	mail := notifications.NewMail().
		From(n.Doer.GetNameAndFromEmail()).
		Subject(subject).
		Line("**" + n.Doer.GetName() + "** mentioned you in a task:")

	lines := bufio.NewScanner(strings.NewReader(n.Task.Description))
	for lines.Scan() {
		mail.Line(lines.Text())
	}

	return mail.
		Action("View Task", n.Task.GetFrontendURL())
}

// ToDB returns the UserMentionedInTaskNotification notification in a format which can be saved in the db
func (n *UserMentionedInTaskNotification) ToDB() interface{} {
	return n
}

// Name returns the name of the notification
func (n *UserMentionedInTaskNotification) Name() string {
	return "task.mentioned"
}

// DataExportReadyNotification represents a DataExportReadyNotification notification
type DataExportReadyNotification struct {
	User *user.User `json:"user"`
}

// ToMail returns the mail notification for DataExportReadyNotification
func (n *DataExportReadyNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("Your Vikunja Data Export is ready").
		Greeting("Hi "+n.User.GetName()+",").
		Line("Your Vikunja Data Export is ready for you to download. Click the button below to download it:").
		Action("Download", config.ServiceFrontendurl.GetString()+"user/export/download").
		Line("The download will be available for the next 7 days.").
		Line("Have a nice day!")
}

// ToDB returns the DataExportReadyNotification notification in a format which can be saved in the db
func (n *DataExportReadyNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *DataExportReadyNotification) Name() string {
	return "data.export.ready"
}
