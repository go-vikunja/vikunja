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

// DataExportRequestEvent represents a DataExportRequestEvent event
type DataExportRequestEvent struct {
	User *user.User
}

// Name defines the name for DataExportRequestEvent
func (t *DataExportRequestEvent) Name() string {
	return "user.export.request"
}

/////////////////
// Task Events //
/////////////////

// TaskCreatedEvent represents an event where a task has been created
type TaskCreatedEvent struct {
	Task *Task
	Doer *user.User
}

// Name defines the name for TaskCreatedEvent
func (t *TaskCreatedEvent) Name() string {
	return "task.created"
}

// TaskUpdatedEvent represents an event where a task has been updated
type TaskUpdatedEvent struct {
	Task *Task
	Doer *user.User
}

// Name defines the name for TaskUpdatedEvent
func (t *TaskUpdatedEvent) Name() string {
	return "task.updated"
}

// TaskDeletedEvent represents a TaskDeletedEvent event
type TaskDeletedEvent struct {
	Task *Task
	Doer *user.User
}

// Name defines the name for TaskDeletedEvent
func (t *TaskDeletedEvent) Name() string {
	return "task.deleted"
}

// TaskAssigneeCreatedEvent represents an event where a task has been assigned to a user
type TaskAssigneeCreatedEvent struct {
	Task     *Task
	Assignee *user.User
	Doer     *user.User
}

// Name defines the name for TaskAssigneeCreatedEvent
func (t *TaskAssigneeCreatedEvent) Name() string {
	return "task.assignee.created"
}

// TaskAssigneeDeletedEvent represents a TaskAssigneeDeletedEvent event
type TaskAssigneeDeletedEvent struct {
	Task     *Task
	Assignee *user.User
	Doer     *user.User
}

// Name defines the name for TaskAssigneeDeletedEvent
func (t *TaskAssigneeDeletedEvent) Name() string {
	return "task.assignee.deleted"
}

// TaskCommentCreatedEvent represents an event where a task comment has been created
type TaskCommentCreatedEvent struct {
	Task    *Task
	Comment *TaskComment
	Doer    *user.User
}

// Name defines the name for TaskCommentCreatedEvent
func (t *TaskCommentCreatedEvent) Name() string {
	return "task.comment.created"
}

// TaskCommentUpdatedEvent represents a TaskCommentUpdatedEvent event
type TaskCommentUpdatedEvent struct {
	Task    *Task
	Comment *TaskComment
	Doer    *user.User
}

// Name defines the name for TaskCommentUpdatedEvent
func (t *TaskCommentUpdatedEvent) Name() string {
	return "task.comment.edited"
}

// TaskCommentDeletedEvent represents a TaskCommentDeletedEvent event
type TaskCommentDeletedEvent struct {
	Task    *Task
	Comment *TaskComment
	Doer    *user.User
}

// Name defines the name for TaskCommentDeletedEvent
func (t *TaskCommentDeletedEvent) Name() string {
	return "task.comment.deleted"
}

// TaskAttachmentCreatedEvent represents a TaskAttachmentCreatedEvent event
type TaskAttachmentCreatedEvent struct {
	Task       *Task
	Attachment *TaskAttachment
	Doer       *user.User
}

// Name defines the name for TaskAttachmentCreatedEvent
func (t *TaskAttachmentCreatedEvent) Name() string {
	return "task.attachment.created"
}

// TaskAttachmentDeletedEvent represents a TaskAttachmentDeletedEvent event
type TaskAttachmentDeletedEvent struct {
	Task       *Task
	Attachment *TaskAttachment
	Doer       *user.User
}

// Name defines the name for TaskAttachmentDeletedEvent
func (t *TaskAttachmentDeletedEvent) Name() string {
	return "task.attachment.deleted"
}

// TaskRelationCreatedEvent represents a TaskRelationCreatedEvent event
type TaskRelationCreatedEvent struct {
	Task     *Task
	Relation *TaskRelation
	Doer     *user.User
}

// Name defines the name for TaskRelationCreatedEvent
func (t *TaskRelationCreatedEvent) Name() string {
	return "task.relation.created"
}

// TaskRelationDeletedEvent represents a TaskRelationDeletedEvent event
type TaskRelationDeletedEvent struct {
	Task     *Task
	Relation *TaskRelation
	Doer     *user.User
}

// Name defines the name for TaskRelationDeletedEvent
func (t *TaskRelationDeletedEvent) Name() string {
	return "task.relation.deleted"
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
// Project Events //
/////////////////

// ProjectCreatedEvent represents an event where a project has been created
type ProjectCreatedEvent struct {
	Project *Project
	Doer    *user.User
}

// Name defines the name for ProjectCreatedEvent
func (l *ProjectCreatedEvent) Name() string {
	return "project.created"
}

// ProjectUpdatedEvent represents an event where a project has been updated
type ProjectUpdatedEvent struct {
	Project *Project
	Doer    web.Auth
}

// Name defines the name for ProjectUpdatedEvent
func (l *ProjectUpdatedEvent) Name() string {
	return "project.updated"
}

// ProjectDeletedEvent represents an event where a project has been deleted
type ProjectDeletedEvent struct {
	Project *Project
	Doer    web.Auth
}

// Name defines the name for ProjectDeletedEvent
func (t *ProjectDeletedEvent) Name() string {
	return "project.deleted"
}

////////////////////
// Sharing Events //
////////////////////

// ProjectSharedWithUserEvent represents an event where a project has been shared with a user
type ProjectSharedWithUserEvent struct {
	Project *Project
	User    *user.User
	Doer    web.Auth
}

// Name defines the name for ProjectSharedWithUserEvent
func (l *ProjectSharedWithUserEvent) Name() string {
	return "project.shared.user"
}

// ProjectSharedWithTeamEvent represents an event where a project has been shared with a team
type ProjectSharedWithTeamEvent struct {
	Project *Project
	Team    *Team
	Doer    web.Auth
}

// Name defines the name for ProjectSharedWithTeamEvent
func (l *ProjectSharedWithTeamEvent) Name() string {
	return "project.shared.team"
}

/////////////////
// Team Events //
/////////////////

// TeamMemberAddedEvent defines an event where a user is added to a team
type TeamMemberAddedEvent struct {
	Team   *Team
	Member *user.User
	Doer   *user.User
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

// UserDataExportRequestedEvent represents a UserDataExportRequestedEvent event
type UserDataExportRequestedEvent struct {
	User *user.User
}

// Name defines the name for UserDataExportRequestedEvent
func (t *UserDataExportRequestedEvent) Name() string {
	return "user.export.requested"
}
