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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

/////////////////
// Task Events //
/////////////////

// TaskCreatedEvent represents an event where a task has been created
type TaskCreatedEvent struct {
	Task *Task
	Doer web.Auth
}

// Name defines the name for TaskCreatedEvent
func (t *TaskCreatedEvent) Name() string {
	return "task.created"
}

// TaskUpdatedEvent represents an event where a task has been updated
type TaskUpdatedEvent struct {
	Task *Task
	Doer web.Auth
}

// Name defines the name for TaskUpdatedEvent
func (t *TaskUpdatedEvent) Name() string {
	return "task.updated"
}

// TaskDeletedEvent represents a TaskDeletedEvent event
type TaskDeletedEvent struct {
	Task *Task
	Doer web.Auth
}

// Name defines the name for TaskDeletedEvent
func (t *TaskDeletedEvent) Name() string {
	return "task.deleted"
}

// TaskAssigneeCreatedEvent represents an event where a task has been assigned to a user
type TaskAssigneeCreatedEvent struct {
	Task     *Task
	Assignee *user.User
	Doer     web.Auth
}

// Name defines the name for TaskAssigneeCreatedEvent
func (t *TaskAssigneeCreatedEvent) Name() string {
	return "task.assignee.created"
}

// TaskCommentCreatedEvent represents an event where a task comment has been created
type TaskCommentCreatedEvent struct {
	Task    *Task
	Comment *TaskComment
	Doer    web.Auth
}

// Name defines the name for TaskCommentCreatedEvent
func (t *TaskCommentCreatedEvent) Name() string {
	return "task.comment.created"
}

//////////////////////
// Namespace Events //
//////////////////////

// NamespaceCreatedEvent represents an event where a namespace has been created
type NamespaceCreatedEvent struct {
	Namespace *Namespace
	Doer      web.Auth
}

// Name defines the name for NamespaceCreatedEvent
func (n *NamespaceCreatedEvent) Name() string {
	return "namespace.created"
}

// NamespaceUpdatedEvent represents an event where a namespace has been updated
type NamespaceUpdatedEvent struct {
	Namespace *Namespace
	Doer      web.Auth
}

// Name defines the name for NamespaceUpdatedEvent
func (n *NamespaceUpdatedEvent) Name() string {
	return "namespace.updated"
}

// NamespaceDeletedEvent represents a NamespaceDeletedEvent event
type NamespaceDeletedEvent struct {
	Namespace *Namespace
	Doer      web.Auth
}

// TopicName defines the name for NamespaceDeletedEvent
func (t *NamespaceDeletedEvent) Name() string {
	return "namespace.deleted"
}

/////////////////
// List Events //
/////////////////

// ListCreatedEvent represents an event where a list has been created
type ListCreatedEvent struct {
	List *List
	Doer web.Auth
}

// Name defines the name for ListCreatedEvent
func (l *ListCreatedEvent) Name() string {
	return "list.created"
}

// ListUpdatedEvent represents an event where a list has been updated
type ListUpdatedEvent struct {
	List *List
	Doer web.Auth
}

// Name defines the name for ListUpdatedEvent
func (l *ListUpdatedEvent) Name() string {
	return "list.updated"
}

// ListDeletedEvent represents an event where a list has been deleted
type ListDeletedEvent struct {
	List *List
	Doer web.Auth
}

// Name defines the name for ListDeletedEvent
func (t *ListDeletedEvent) Name() string {
	return "list.deleted"
}

////////////////////
// Sharing Events //
////////////////////

// ListSharedWithUserEvent represents an event where a list has been shared with a user
type ListSharedWithUserEvent struct {
	List *List
	User *user.User
	Doer web.Auth
}

// Name defines the name for ListSharedWithUserEvent
func (l *ListSharedWithUserEvent) Name() string {
	return "list.shared.user"
}

// ListSharedWithTeamEvent represents an event where a list has been shared with a team
type ListSharedWithTeamEvent struct {
	List *List
	Team *Team
	Doer web.Auth
}

// Name defines the name for ListSharedWithTeamEvent
func (l *ListSharedWithTeamEvent) Name() string {
	return "list.shared.team"
}

// NamespaceSharedWithUserEvent represents an event where a namespace has been shared with a user
type NamespaceSharedWithUserEvent struct {
	Namespace *Namespace
	User      *user.User
	Doer      web.Auth
}

// Name defines the name for NamespaceSharedWithUserEvent
func (n *NamespaceSharedWithUserEvent) Name() string {
	return "namespace.shared.user"
}

// NamespaceSharedWithTeamEvent represents an event where a namespace has been shared with a team
type NamespaceSharedWithTeamEvent struct {
	Namespace *Namespace
	Team      *Team
	Doer      web.Auth
}

// Name defines the name for NamespaceSharedWithTeamEvent
func (n *NamespaceSharedWithTeamEvent) Name() string {
	return "namespace.shared.team"
}

/////////////////
// Team Events //
/////////////////

// TeamMemberAddedEvent defines an event where a user is added to a team
type TeamMemberAddedEvent struct {
	Team   *Team
	Member *user.User
	Doer   web.Auth
}

// Name defines the name for TeamMemberAddedEvent
func (t *TeamMemberAddedEvent) Name() string {
	return "team.member.added"
}

// TeamCreatedEvent represents a TeamCreatedEvent event
type TeamCreatedEvent struct {
	Team *Team
	Doer web.Auth
}

// Name defines the name for TeamCreatedEvent
func (t *TeamCreatedEvent) Name() string {
	return "team.created"
}

// TeamDeletedEvent represents a TeamDeletedEvent event
type TeamDeletedEvent struct {
	Team *Team
	Doer web.Auth
}

// Name defines the name for TeamDeletedEvent
func (t *TeamDeletedEvent) Name() string {
	return "team.deleted"
}
