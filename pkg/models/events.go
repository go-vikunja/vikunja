// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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
	Task *Task      `json:"task"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for TaskCreatedEvent
func (t *TaskCreatedEvent) Name() string {
	return "task.created"
}

func (t *TaskCreatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskUpdatedEvent represents an event where a task has been updated
type TaskUpdatedEvent struct {
	Task *Task      `json:"task"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for TaskUpdatedEvent
func (t *TaskUpdatedEvent) Name() string {
	return "task.updated"
}

func (t *TaskUpdatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskDeletedEvent represents a TaskDeletedEvent event
type TaskDeletedEvent struct {
	Task *Task      `json:"task"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for TaskDeletedEvent
func (t *TaskDeletedEvent) Name() string {
	return "task.deleted"
}

func (t *TaskDeletedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskAssigneeCreatedEvent represents an event where a task has been assigned to a user
type TaskAssigneeCreatedEvent struct {
	Task     *Task      `json:"task"`
	Assignee *user.User `json:"assignee"`
	Doer     *user.User `json:"doer"`
}

// Name defines the name for TaskAssigneeCreatedEvent
func (t *TaskAssigneeCreatedEvent) Name() string {
	return "task.assignee.created"
}

func (t *TaskAssigneeCreatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskAssigneeDeletedEvent represents a TaskAssigneeDeletedEvent event
type TaskAssigneeDeletedEvent struct {
	Task     *Task      `json:"task"`
	Assignee *user.User `json:"assignee"`
	Doer     *user.User `json:"doer"`
}

// Name defines the name for TaskAssigneeDeletedEvent
func (t *TaskAssigneeDeletedEvent) Name() string {
	return "task.assignee.deleted"
}

func (t *TaskAssigneeDeletedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskCommentCreatedEvent represents an event where a task comment has been created
type TaskCommentCreatedEvent struct {
	Task    *Task        `json:"task"`
	Comment *TaskComment `json:"comment"`
	Doer    *user.User   `json:"doer"`
}

// Name defines the name for TaskCommentCreatedEvent
func (t *TaskCommentCreatedEvent) Name() string {
	return "task.comment.created"
}

func (t *TaskCommentCreatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskCommentUpdatedEvent represents a TaskCommentUpdatedEvent event
type TaskCommentUpdatedEvent struct {
	Task    *Task        `json:"task"`
	Comment *TaskComment `json:"comment"`
	Doer    *user.User   `json:"doer"`
}

// Name defines the name for TaskCommentUpdatedEvent
func (t *TaskCommentUpdatedEvent) Name() string {
	return "task.comment.edited"
}

func (t *TaskCommentUpdatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskCommentDeletedEvent represents a TaskCommentDeletedEvent event
type TaskCommentDeletedEvent struct {
	Task    *Task        `json:"task"`
	Comment *TaskComment `json:"comment"`
	Doer    *user.User   `json:"doer"`
}

// Name defines the name for TaskCommentDeletedEvent
func (t *TaskCommentDeletedEvent) Name() string {
	return "task.comment.deleted"
}

func (t *TaskCommentDeletedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskAttachmentCreatedEvent represents a TaskAttachmentCreatedEvent event
type TaskAttachmentCreatedEvent struct {
	Task       *Task           `json:"task"`
	Attachment *TaskAttachment `json:"attachment"`
	Doer       *user.User      `json:"doer"`
}

// Name defines the name for TaskAttachmentCreatedEvent
func (t *TaskAttachmentCreatedEvent) Name() string {
	return "task.attachment.created"
}

func (t *TaskAttachmentCreatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskAttachmentDeletedEvent represents a TaskAttachmentDeletedEvent event
type TaskAttachmentDeletedEvent struct {
	Task       *Task           `json:"task"`
	Attachment *TaskAttachment `json:"attachment"`
	Doer       *user.User      `json:"doer"`
}

// Name defines the name for TaskAttachmentDeletedEvent
func (t *TaskAttachmentDeletedEvent) Name() string {
	return "task.attachment.deleted"
}

func (t *TaskAttachmentDeletedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskRelationCreatedEvent represents a TaskRelationCreatedEvent event
type TaskRelationCreatedEvent struct {
	Task     *Task         `json:"task"`
	Relation *TaskRelation `json:"relation"`
	Doer     *user.User    `json:"doer"`
}

// Name defines the name for TaskRelationCreatedEvent
func (t *TaskRelationCreatedEvent) Name() string {
	return "task.relation.created"
}

func (t *TaskRelationCreatedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

// TaskRelationDeletedEvent represents a TaskRelationDeletedEvent event
type TaskRelationDeletedEvent struct {
	Task     *Task         `json:"task"`
	Relation *TaskRelation `json:"relation"`
	Doer     *user.User    `json:"doer"`
}

// Name defines the name for TaskRelationDeletedEvent
func (t *TaskRelationDeletedEvent) Name() string {
	return "task.relation.deleted"
}

func (t *TaskRelationDeletedEvent) ProjectID() int64 {
	return t.Task.ProjectID
}

////////////////////
// Project Events //
////////////////////

// ProjectCreatedEvent represents an event where a project has been created
type ProjectCreatedEvent struct {
	Project *Project   `json:"project"`
	Doer    *user.User `json:"doer"`
}

// Name defines the name for ProjectCreatedEvent
func (l *ProjectCreatedEvent) Name() string {
	return "project.created"
}

// ProjectUpdatedEvent represents an event where a project has been updated
type ProjectUpdatedEvent struct {
	Project *Project `json:"project"`
	Doer    web.Auth `json:"doer"`
}

// Name defines the name for ProjectUpdatedEvent
func (p *ProjectUpdatedEvent) Name() string {
	return "project.updated"
}

func (p *ProjectUpdatedEvent) ProjectID() int64 {
	return p.Project.ID
}

// ProjectDeletedEvent represents an event where a project has been deleted
type ProjectDeletedEvent struct {
	Project *Project `json:"project"`
	Doer    web.Auth `json:"doer"`
}

// Name defines the name for ProjectDeletedEvent
func (p *ProjectDeletedEvent) Name() string {
	return "project.deleted"
}

func (p *ProjectDeletedEvent) ProjectID() int64 {
	return p.Project.ID
}

////////////////////
// Sharing Events //
////////////////////

// ProjectSharedWithUserEvent represents an event where a project has been shared with a user
type ProjectSharedWithUserEvent struct {
	Project *Project   `json:"project"`
	User    *user.User `json:"user"`
	Doer    web.Auth   `json:"doer"`
}

// Name defines the name for ProjectSharedWithUserEvent
func (p *ProjectSharedWithUserEvent) Name() string {
	return "project.shared.user"
}

func (p *ProjectSharedWithUserEvent) ProjectID() int64 {
	return p.Project.ID
}

// ProjectSharedWithTeamEvent represents an event where a project has been shared with a team
type ProjectSharedWithTeamEvent struct {
	Project *Project `json:"project"`
	Team    *Team    `json:"team"`
	Doer    web.Auth `json:"doer"`
}

// Name defines the name for ProjectSharedWithTeamEvent
func (p *ProjectSharedWithTeamEvent) Name() string {
	return "project.shared.team"
}

func (p *ProjectSharedWithTeamEvent) ProjectID() int64 {
	return p.Project.ID
}

/////////////////
// Team Events //
/////////////////

// TeamMemberAddedEvent defines an event where a user is added to a team
type TeamMemberAddedEvent struct {
	Team   *Team      `json:"team"`
	Member *user.User `json:"member"`
	Doer   *user.User `json:"doer"`
}

// Name defines the name for TeamMemberAddedEvent
func (t *TeamMemberAddedEvent) Name() string {
	return "team.member.added"
}

// TeamCreatedEvent represents a TeamCreatedEvent event
type TeamCreatedEvent struct {
	Team *Team    `json:"team"`
	Doer web.Auth `json:"doer"`
}

// Name defines the name for TeamCreatedEvent
func (t *TeamCreatedEvent) Name() string {
	return "team.created"
}

// TeamDeletedEvent represents a TeamDeletedEvent event
type TeamDeletedEvent struct {
	Team *Team    `json:"team"`
	Doer web.Auth `json:"doer"`
}

// Name defines the name for TeamDeletedEvent
func (t *TeamDeletedEvent) Name() string {
	return "team.deleted"
}

// UserDataExportRequestedEvent represents a UserDataExportRequestedEvent event
type UserDataExportRequestedEvent struct {
	User *user.User `json:"user"`
}

// Name defines the name for UserDataExportRequestedEvent
func (t *UserDataExportRequestedEvent) Name() string {
	return "user.export.requested"
}
