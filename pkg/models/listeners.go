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
	"github.com/ThreeDotsLabs/watermill/message"
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

// Hanlde is executed when the event IncreaseTaskCounter listens on is fired
func (s *IncreaseTaskCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.IncrBy(metrics.TaskCountKey, 1)
}

// DecreaseTaskCounter  represents a listener
type DecreaseTaskCounter struct {
}

// Name defines the name for the DecreaseTaskCounter listener
func (s *DecreaseTaskCounter) Name() string {
	return "task.counter.decrease"
}

// Hanlde is executed when the event DecreaseTaskCounter listens on is fired
func (s *DecreaseTaskCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.DecrBy(metrics.TaskCountKey, 1)
}

// SendTaskCommentNotification  represents a listener
type SendTaskCommentNotification struct {
}

// Name defines the name for the SendTaskCommentNotification listener
func (s *SendTaskCommentNotification) Name() string {
	return "send.task.comment.notification"
}

// Handle is executed when the event SendTaskCommentNotification listens on is fired
func (s *SendTaskCommentNotification) Handle(payload message.Payload) (err error) {
	event := &TaskCommentCreatedEvent{}
	err = json.Unmarshal(payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	subscribers, err := getSubscribersForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending task comment notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
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

// SendTaskAssignedNotification  represents a listener
type SendTaskAssignedNotification struct {
}

// Name defines the name for the SendTaskAssignedNotification listener
func (s *SendTaskAssignedNotification) Name() string {
	return "send.task.assigned.notification"
}

// Handle is executed when the event SendTaskAssignedNotification listens on is fired
func (s *SendTaskAssignedNotification) Handle(payload message.Payload) (err error) {
	event := &TaskAssigneeCreatedEvent{}
	err = json.Unmarshal(payload, event)
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

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		n := &TaskAssignedNotification{
			Doer:     event.Doer,
			Task:     event.Task,
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
	return "send.task.deleted.notification"
}

// Handle is executed when the event SendTaskDeletedNotification listens on is fired
func (s *SendTaskDeletedNotification) Handle(payload message.Payload) (err error) {
	event := &TaskDeletedEvent{}
	err = json.Unmarshal(payload, event)
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
	return "subscribe.assignee.to.task"
}

// Handle is executed when the event SubscribeAssigneeToTask listens on is fired
func (s *SubscribeAssigneeToTask) Handle(payload message.Payload) (err error) {
	event := &TaskAssigneeCreatedEvent{}
	err = json.Unmarshal(payload, event)
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

///////
// List Event Listeners

type IncreaseListCounter struct {
}

func (s *IncreaseListCounter) Name() string {
	return "list.counter.increase"
}

func (s *IncreaseListCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.IncrBy(metrics.ListCountKey, 1)
}

type DecreaseListCounter struct {
}

func (s *DecreaseListCounter) Name() string {
	return "list.counter.decrease"
}

func (s *DecreaseListCounter) Handle(payload message.Payload) (err error) {
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
func (s *SendListCreatedNotification) Handle(payload message.Payload) (err error) {
	event := &ListCreatedEvent{}
	err = json.Unmarshal(payload, event)
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
func (s *IncreaseNamespaceCounter) Handle(payload message.Payload) (err error) {
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
func (s *DecreaseNamespaceCounter) Handle(payload message.Payload) (err error) {
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
func (s *IncreaseTeamCounter) Handle(payload message.Payload) (err error) {
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
func (s *DecreaseTeamCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.DecrBy(metrics.TeamCountKey, 1)
}

// SendTeamMemberAddedNotification  represents a listener
type SendTeamMemberAddedNotification struct {
}

// Name defines the name for the SendTeamMemberAddedNotification listener
func (s *SendTeamMemberAddedNotification) Name() string {
	return "send.team.member.added.notification"
}

// Handle is executed when the event SendTeamMemberAddedNotification listens on is fired
func (s *SendTeamMemberAddedNotification) Handle(payload message.Payload) (err error) {
	event := &TeamMemberAddedEvent{}
	err = json.Unmarshal(payload, event)
	if err != nil {
		return err
	}

	return notifications.Notify(event.Member, &TeamMemberAddedNotification{
		Member: event.Member,
		Doer:   event.Doer,
		Team:   event.Team,
	})
}
