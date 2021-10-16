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
	"encoding/json"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/xorm"
)

// RegisterListeners registers all event listeners
func RegisterListeners() {
	events.RegisterListener((&ListCreatedEvent{}).Name(), &IncreaseListCounter{})
	events.RegisterListener((&ListDeletedEvent{}).Name(), &DecreaseListCounter{})
	events.RegisterListener((&NamespaceCreatedEvent{}).Name(), &IncreaseNamespaceCounter{})
	events.RegisterListener((&NamespaceDeletedEvent{}).Name(), &DecreaseNamespaceCounter{})
	events.RegisterListener((&TaskCreatedEvent{}).Name(), &IncreaseTaskCounter{})
	events.RegisterListener((&TaskDeletedEvent{}).Name(), &DecreaseTaskCounter{})
	events.RegisterListener((&TeamDeletedEvent{}).Name(), &DecreaseTeamCounter{})
	events.RegisterListener((&TeamCreatedEvent{}).Name(), &IncreaseTeamCounter{})
	events.RegisterListener((&TaskCommentCreatedEvent{}).Name(), &SendTaskCommentNotification{})
	events.RegisterListener((&TaskAssigneeCreatedEvent{}).Name(), &SendTaskAssignedNotification{})
	events.RegisterListener((&TaskDeletedEvent{}).Name(), &SendTaskDeletedNotification{})
	events.RegisterListener((&ListCreatedEvent{}).Name(), &SendListCreatedNotification{})
	events.RegisterListener((&TaskAssigneeCreatedEvent{}).Name(), &SubscribeAssigneeToTask{})
	events.RegisterListener((&TeamMemberAddedEvent{}).Name(), &SendTeamMemberAddedNotification{})
	events.RegisterListener((&TaskCommentUpdatedEvent{}).Name(), &HandleTaskCommentEditMentions{})
	events.RegisterListener((&TaskCreatedEvent{}).Name(), &HandleTaskCreateMentions{})
	events.RegisterListener((&TaskUpdatedEvent{}).Name(), &HandleTaskUpdatedMentions{})
	events.RegisterListener((&UserDataExportRequestedEvent{}).Name(), &HandleUserDataExport{})
}

//////
// Task Events

// IncreaseTaskCounter  represents a listener
type IncreaseTaskCounter struct {
}

// Name defines the name for the IncreaseTaskCounter listener
func (s *IncreaseTaskCounter) Name() string {
	return "task.counter.increase"
}

// Handle is executed when the event IncreaseTaskCounter listens on is fired
func (s *IncreaseTaskCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.TaskCountKey, 1)
}

// DecreaseTaskCounter  represents a listener
type DecreaseTaskCounter struct {
}

// Name defines the name for the DecreaseTaskCounter listener
func (s *DecreaseTaskCounter) Name() string {
	return "task.counter.decrease"
}

// Handle is executed when the event DecreaseTaskCounter listens on is fired
func (s *DecreaseTaskCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.TaskCountKey, 1)
}

func notifyMentionedUsers(sess *xorm.Session, task *Task, text string, n notifications.NotificationWithSubject) (users map[int64]*user.User, err error) {
	users, err = FindMentionedUsersInText(sess, text)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	log.Debugf("Processing %d mentioned users for text %d", len(users), n.SubjectID())

	var notified int
	for _, u := range users {
		can, _, err := task.CanRead(sess, u)
		if err != nil {
			return users, err
		}

		if !can {
			continue
		}

		// Don't notify a user if they were already notified
		dbn, err := notifications.GetNotificationsForNameAndUser(sess, u.ID, n.Name(), n.SubjectID())
		if err != nil {
			return users, err
		}

		if len(dbn) > 0 {
			continue
		}

		err = notifications.Notify(u, n)
		if err != nil {
			return users, err
		}
		notified++
	}

	log.Debugf("Notified %d mentioned users for text %d", notified, n.SubjectID())

	return
}

// SendTaskCommentNotification  represents a listener
type SendTaskCommentNotification struct {
}

// Name defines the name for the SendTaskCommentNotification listener
func (s *SendTaskCommentNotification) Name() string {
	return "task.comment.notification.send"
}

// Handle is executed when the event SendTaskCommentNotification listens on is fired
func (s *SendTaskCommentNotification) Handle(msg *message.Message) (err error) {
	event := &TaskCommentCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &TaskCommentNotification{
		Doer:      event.Doer,
		Task:      event.Task,
		Comment:   event.Comment,
		Mentioned: true,
	}
	mentionedUsers, err := notifyMentionedUsers(sess, event.Task, event.Comment.Comment, n)
	if err != nil {
		return err
	}

	subscribers, err := getSubscribersForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending task comment notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		if _, has := mentionedUsers[subscriber.UserID]; has {
			continue
		}

		n := &TaskCommentNotification{
			Doer:    event.Doer,
			Task:    event.Task,
			Comment: event.Comment,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return
}

// HandleTaskCommentEditMentions  represents a listener
type HandleTaskCommentEditMentions struct {
}

// Name defines the name for the HandleTaskCommentEditMentions listener
func (s *HandleTaskCommentEditMentions) Name() string {
	return "handle.task.comment.edit.mentions"
}

// Handle is executed when the event HandleTaskCommentEditMentions listens on is fired
func (s *HandleTaskCommentEditMentions) Handle(msg *message.Message) (err error) {
	event := &TaskCommentUpdatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &TaskCommentNotification{
		Doer:      event.Doer,
		Task:      event.Task,
		Comment:   event.Comment,
		Mentioned: true,
	}
	_, err = notifyMentionedUsers(sess, event.Task, event.Comment.Comment, n)
	return err
}

// SendTaskAssignedNotification  represents a listener
type SendTaskAssignedNotification struct {
}

// Name defines the name for the SendTaskAssignedNotification listener
func (s *SendTaskAssignedNotification) Name() string {
	return "task.assigned.notification.send"
}

// Handle is executed when the event SendTaskAssignedNotification listens on is fired
func (s *SendTaskAssignedNotification) Handle(msg *message.Message) (err error) {
	event := &TaskAssigneeCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	subscribers, err := getSubscribersForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending task assigned notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	task, err := GetTaskByIDSimple(sess, event.Task.ID)
	if err != nil {
		return err
	}

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		n := &TaskAssignedNotification{
			Doer:     event.Doer,
			Task:     &task,
			Assignee: event.Assignee,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return nil
}

// SendTaskDeletedNotification  represents a listener
type SendTaskDeletedNotification struct {
}

// Name defines the name for the SendTaskDeletedNotification listener
func (s *SendTaskDeletedNotification) Name() string {
	return "task.deleted.notification.send"
}

// Handle is executed when the event SendTaskDeletedNotification listens on is fired
func (s *SendTaskDeletedNotification) Handle(msg *message.Message) (err error) {
	event := &TaskDeletedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	subscribers, err := getSubscribersForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending task deleted notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		n := &TaskDeletedNotification{
			Doer: event.Doer,
			Task: event.Task,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return nil
}

type SubscribeAssigneeToTask struct {
}

// Name defines the name for the SubscribeAssigneeToTask listener
func (s *SubscribeAssigneeToTask) Name() string {
	return "task.assignee.subscribe"
}

// Handle is executed when the event SubscribeAssigneeToTask listens on is fired
func (s *SubscribeAssigneeToTask) Handle(msg *message.Message) (err error) {
	event := &TaskAssigneeCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sub := &Subscription{
		UserID:     event.Assignee.ID,
		EntityType: SubscriptionEntityTask,
		EntityID:   event.Task.ID,
	}

	sess := db.NewSession()
	defer sess.Close()

	err = sub.Create(sess, event.Assignee)
	if err != nil && !IsErrSubscriptionAlreadyExists(err) {
		return err
	}

	return sess.Commit()
}

// HandleTaskCreateMentions  represents a listener
type HandleTaskCreateMentions struct {
}

// Name defines the name for the HandleTaskCreateMentions listener
func (s *HandleTaskCreateMentions) Name() string {
	return "task.created.mentions"
}

// Handle is executed when the event HandleTaskCreateMentions listens on is fired
func (s *HandleTaskCreateMentions) Handle(msg *message.Message) (err error) {
	event := &TaskCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &UserMentionedInTaskNotification{
		Task:  event.Task,
		Doer:  event.Doer,
		IsNew: true,
	}
	_, err = notifyMentionedUsers(sess, event.Task, event.Task.Description, n)
	return err
}

// HandleTaskUpdatedMentions  represents a listener
type HandleTaskUpdatedMentions struct {
}

// Name defines the name for the HandleTaskUpdatedMentions listener
func (s *HandleTaskUpdatedMentions) Name() string {
	return "task.updated.mentions"
}

// Handle is executed when the event HandleTaskUpdatedMentions listens on is fired
func (s *HandleTaskUpdatedMentions) Handle(msg *message.Message) (err error) {
	event := &TaskUpdatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &UserMentionedInTaskNotification{
		Task:  event.Task,
		Doer:  event.Doer,
		IsNew: false,
	}
	_, err = notifyMentionedUsers(sess, event.Task, event.Task.Description, n)
	return err

}

///////
// List Event Listeners

type IncreaseListCounter struct {
}

func (s *IncreaseListCounter) Name() string {
	return "list.counter.increase"
}

func (s *IncreaseListCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.ListCountKey, 1)
}

type DecreaseListCounter struct {
}

func (s *DecreaseListCounter) Name() string {
	return "list.counter.decrease"
}

func (s *DecreaseListCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.ListCountKey, 1)
}

// SendListCreatedNotification  represents a listener
type SendListCreatedNotification struct {
}

// Name defines the name for the SendListCreatedNotification listener
func (s *SendListCreatedNotification) Name() string {
	return "send.list.created.notification"
}

// Handle is executed when the event SendListCreatedNotification listens on is fired
func (s *SendListCreatedNotification) Handle(msg *message.Message) (err error) {
	event := &ListCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	subscribers, err := getSubscribersForEntity(sess, SubscriptionEntityList, event.List.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending list created notifications to %d subscribers for list %d", len(subscribers), event.List.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		n := &ListCreatedNotification{
			Doer: event.Doer,
			List: event.List,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return nil
}

//////
// Namespace events

// IncreaseNamespaceCounter  represents a listener
type IncreaseNamespaceCounter struct {
}

// Name defines the name for the IncreaseNamespaceCounter listener
func (s *IncreaseNamespaceCounter) Name() string {
	return "namespace.counter.increase"
}

// Hanlde is executed when the event IncreaseNamespaceCounter listens on is fired
func (s *IncreaseNamespaceCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.NamespaceCountKey, 1)
}

// DecreaseNamespaceCounter  represents a listener
type DecreaseNamespaceCounter struct {
}

// Name defines the name for the DecreaseNamespaceCounter listener
func (s *DecreaseNamespaceCounter) Name() string {
	return "namespace.counter.decrease"
}

// Hanlde is executed when the event DecreaseNamespaceCounter listens on is fired
func (s *DecreaseNamespaceCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.NamespaceCountKey, 1)
}

///////
// Team Events

// IncreaseTeamCounter  represents a listener
type IncreaseTeamCounter struct {
}

// Name defines the name for the IncreaseTeamCounter listener
func (s *IncreaseTeamCounter) Name() string {
	return "team.counter.increase"
}

// Hanlde is executed when the event IncreaseTeamCounter listens on is fired
func (s *IncreaseTeamCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.TeamCountKey, 1)
}

// DecreaseTeamCounter  represents a listener
type DecreaseTeamCounter struct {
}

// Name defines the name for the DecreaseTeamCounter listener
func (s *DecreaseTeamCounter) Name() string {
	return "team.counter.decrease"
}

// Hanlde is executed when the event DecreaseTeamCounter listens on is fired
func (s *DecreaseTeamCounter) Handle(msg *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.TeamCountKey, 1)
}

// SendTeamMemberAddedNotification  represents a listener
type SendTeamMemberAddedNotification struct {
}

// Name defines the name for the SendTeamMemberAddedNotification listener
func (s *SendTeamMemberAddedNotification) Name() string {
	return "team.member.added.notification"
}

// Handle is executed when the event SendTeamMemberAddedNotification listens on is fired
func (s *SendTeamMemberAddedNotification) Handle(msg *message.Message) (err error) {
	event := &TeamMemberAddedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	// Don't notify the user themselves
	if event.Doer.ID == event.Member.ID {
		return nil
	}

	return notifications.Notify(event.Member, &TeamMemberAddedNotification{
		Member: event.Member,
		Doer:   event.Doer,
		Team:   event.Team,
	})
}

// HandleUserDataExport  represents a listener
type HandleUserDataExport struct {
}

// Name defines the name for the HandleUserDataExport listener
func (s *HandleUserDataExport) Name() string {
	return "handle.user.data.export"
}

// Handle is executed when the event HandleUserDataExport listens on is fired
func (s *HandleUserDataExport) Handle(msg *message.Message) (err error) {
	event := &UserDataExportRequestedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	log.Debugf("Starting to export user data for user %d...", event.User.ID)

	sess := db.NewSession()
	defer sess.Close()
	err = sess.Begin()
	if err != nil {
		return
	}

	err = ExportUserData(sess, event.User)
	if err != nil {
		_ = sess.Rollback()
		return
	}

	log.Debugf("Done exporting user data for user %d...", event.User.ID)

	err = sess.Commit()
	return err
}
