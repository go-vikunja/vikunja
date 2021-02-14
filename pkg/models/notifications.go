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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
)

// ReminderDueNotification represents a ReminderDueNotification notification
type ReminderDueNotification struct {
	User *user.User
	Task *Task
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

// TaskCommentNotification represents a TaskCommentNotification notification
type TaskCommentNotification struct {
	Doer    *user.User
	Task    *Task
	Comment *TaskComment
}

// ToMail returns the mail notification for TaskCommentNotification
func (n *TaskCommentNotification) ToMail() *notifications.Mail {

	mail := notifications.NewMail().
		From(n.Doer.GetName() + " via Vikunja <" + config.MailerFromEmail.GetString() + ">").
		Subject("Re: " + n.Task.Title)

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

// TaskAssignedNotification represents a TaskAssignedNotification notification
type TaskAssignedNotification struct {
	Doer     *user.User
	Task     *Task
	Assignee *user.User
}

// ToMail returns the mail notification for TaskAssignedNotification
func (n *TaskAssignedNotification) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject(n.Task.Title+"("+n.Task.GetFullIdentifier()+")"+" has been assigned to "+n.Assignee.GetName()).
		Line(n.Doer.GetName()+" has assigned this task to "+n.Assignee.GetName()).
		Action("View Task", n.Task.GetFrontendURL())
}

// ToDB returns the TaskAssignedNotification notification in a format which can be saved in the db
func (n *TaskAssignedNotification) ToDB() interface{} {
	return n
}

// TaskDeletedNotification represents a TaskDeletedNotification notification
type TaskDeletedNotification struct {
	Doer *user.User
	Task *Task
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

// ListCreatedNotification represents a ListCreatedNotification notification
type ListCreatedNotification struct {
	Doer *user.User
	List *List
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
	return nil
}
