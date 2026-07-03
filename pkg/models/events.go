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
	"code.vikunja.io/api/pkg/user"
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

// TaskUpdatedEvent represents an event where a task has been updated
type TaskUpdatedEvent struct {
	Task *Task      `json:"task"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for TaskUpdatedEvent
func (t *TaskUpdatedEvent) Name() string {
	return "task.updated"
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

// TaskPositionsRecalculatedEvent represents a TaskPositionsRecalculatedEvent event
type TaskPositionsRecalculatedEvent struct {
	NewTaskPositions []*TaskPosition
}

// Name defines the name for TaskPositionsRecalculatedEvent
func (t *TaskPositionsRecalculatedEvent) Name() string {
	return "task.positions.recalculated"
}

// TaskReminderFiredEvent represents an event where a task reminder has fired
type TaskReminderFiredEvent struct {
	Task     *Task         `json:"task"`
	User     *user.User    `json:"user"`
	Project  *Project      `json:"project"`
	Reminder *TaskReminder `json:"reminder"`
}

// Name defines the name for TaskReminderFiredEvent
func (t *TaskReminderFiredEvent) Name() string {
	return "task.reminder.fired"
}

// TaskOverdueEvent represents an event where a task is overdue
type TaskOverdueEvent struct {
	Task    *Task      `json:"task"`
	User    *user.User `json:"user"`
	Project *Project   `json:"project"`
}

// Name defines the name for TaskOverdueEvent
func (t *TaskOverdueEvent) Name() string {
	return "task.overdue"
}

// TasksOverdueEvent represents an event where multiple tasks are overdue for a user
type TasksOverdueEvent struct {
	Tasks    []*Task            `json:"tasks"`
	User     *user.User         `json:"user"`
	Projects map[int64]*Project `json:"projects"`
}

// Name defines the name for TasksOverdueEvent
func (t *TasksOverdueEvent) Name() string {
	return "tasks.overdue"
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
	Project *Project   `json:"project"`
	Doer    *user.User `json:"doer"`
}

// Name defines the name for ProjectUpdatedEvent
func (p *ProjectUpdatedEvent) Name() string {
	return "project.updated"
}

// ProjectDeletedEvent represents an event where a project has been deleted
type ProjectDeletedEvent struct {
	Project *Project   `json:"project"`
	Doer    *user.User `json:"doer"`
}

// Name defines the name for ProjectDeletedEvent
func (p *ProjectDeletedEvent) Name() string {
	return "project.deleted"
}

////////////////////
// Sharing Events //
////////////////////

// ProjectSharedWithUserEvent represents an event where a project has been shared with a user
type ProjectSharedWithUserEvent struct {
	Project *Project   `json:"project"`
	User    *user.User `json:"user"`
	Doer    *user.User `json:"doer"`
}

// Name defines the name for ProjectSharedWithUserEvent
func (p *ProjectSharedWithUserEvent) Name() string {
	return "project.shared.user"
}

// ProjectSharedWithTeamEvent represents an event where a project has been shared with a team
type ProjectSharedWithTeamEvent struct {
	Project *Project   `json:"project"`
	Team    *Team      `json:"team"`
	Doer    *user.User `json:"doer"`
}

// Name defines the name for ProjectSharedWithTeamEvent
func (p *ProjectSharedWithTeamEvent) Name() string {
	return "project.shared.team"
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

// TeamMemberRemovedEvent defines an event where a user is removed from a team
type TeamMemberRemovedEvent struct {
	Team   *Team      `json:"team"`
	Member *user.User `json:"member"`
	Doer   *user.User `json:"doer"`
}

// Name defines the name for TeamMemberRemovedEvent
func (t *TeamMemberRemovedEvent) Name() string {
	return "team.member.removed"
}

// TeamCreatedEvent represents a TeamCreatedEvent event
type TeamCreatedEvent struct {
	Team *Team      `json:"team"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for TeamCreatedEvent
func (t *TeamCreatedEvent) Name() string {
	return "team.created"
}

// TeamDeletedEvent represents a TeamDeletedEvent event
type TeamDeletedEvent struct {
	Team *Team      `json:"team"`
	Doer *user.User `json:"doer"`
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

/////////////////////
// Webhook Events  //
/////////////////////

// WebhookDeliveryEvent is an internal event used to fan out a single
// webhook delivery. One of these is dispatched per matching webhook by
// WebhookListener; the WebhookDeliveryListener performs the actual HTTP
// call. This event is intentionally not exposed via RegisterEventForWebhook
// — users cannot subscribe to it.
type WebhookDeliveryEvent struct {
	// WebhookID is the id of the webhook row to deliver to. The delivery
	// listener loads the webhook at delivery time so secrets are never
	// embedded in the message bus.
	WebhookID int64 `json:"webhook_id"`
	// Payload is the fully prepared webhook payload, including the already
	// expanded event.Data map. Build-once semantics: retries replay the
	// same payload rather than rebuilding it.
	Payload *WebhookPayload `json:"payload"`
}

// Name defines the name for WebhookDeliveryEvent
func (w *WebhookDeliveryEvent) Name() string {
	return "webhook.delivery"
}

// TimeEntryCreatedEvent represents a time entry being created
type TimeEntryCreatedEvent struct {
	TimeEntry *TimeEntry `json:"time_entry"`
	Doer      *user.User `json:"doer"`
}

// Name defines the name for TimeEntryCreatedEvent
func (e *TimeEntryCreatedEvent) Name() string {
	return "time-entry.created"
}

// TimeEntryUpdatedEvent represents a time entry being updated (including a timer being stopped)
type TimeEntryUpdatedEvent struct {
	TimeEntry *TimeEntry `json:"time_entry"`
	Doer      *user.User `json:"doer"`
}

// Name defines the name for TimeEntryUpdatedEvent
func (e *TimeEntryUpdatedEvent) Name() string {
	return "time-entry.updated"
}

// TimeEntryDeletedEvent represents a time entry being deleted
type TimeEntryDeletedEvent struct {
	TimeEntry *TimeEntry `json:"time_entry"`
	Doer      *user.User `json:"doer"`
}

// Name defines the name for TimeEntryDeletedEvent
func (e *TimeEntryDeletedEvent) Name() string {
	return "time-entry.deleted"
}

////////////////////
// API Token Events

// API token events carry IDs only: the freshly created token struct holds the
// raw token string, which must never end up in a message payload (the poison
// queue logs payloads on handler failure).

// APITokenIssuedEvent represents an API token being created
type APITokenIssuedEvent struct {
	TokenID int64 `json:"token_id"`
	DoerID  int64 `json:"doer_id"`
	OwnerID int64 `json:"owner_id"`
}

// Name defines the name for APITokenIssuedEvent
func (e *APITokenIssuedEvent) Name() string {
	return "api-token.issued"
}

// APITokenRevokedEvent represents an API token being deleted
type APITokenRevokedEvent struct {
	TokenID int64 `json:"token_id"`
	DoerID  int64 `json:"doer_id"`
}

// Name defines the name for APITokenRevokedEvent
func (e *APITokenRevokedEvent) Name() string {
	return "api-token.revoked"
}

// APITokenUsedEvent represents an API token authenticating a request
type APITokenUsedEvent struct {
	TokenID int64 `json:"token_id"`
	OwnerID int64 `json:"owner_id"`
}

// Name defines the name for APITokenUsedEvent
func (e *APITokenUsedEvent) Name() string {
	return "api-token.used"
}

//////////////////
// Admin Events

// Admin events cover mutations performed through the instance-admin API. They
// exist for audit logging and are deliberately not registered as webhook
// events — they are instance-level, not project-scoped.

// AdminUserCreatedEvent represents a user being provisioned through the admin API
type AdminUserCreatedEvent struct {
	User *user.User `json:"user"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for AdminUserCreatedEvent
func (e *AdminUserCreatedEvent) Name() string {
	return "admin.user.created"
}

// AdminUserAdminGrantedEvent represents a user being promoted to instance admin
type AdminUserAdminGrantedEvent struct {
	User *user.User `json:"user"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for AdminUserAdminGrantedEvent
func (e *AdminUserAdminGrantedEvent) Name() string {
	return "admin.user.admin.granted"
}

// AdminUserAdminRevokedEvent represents a user's instance-admin flag being revoked
type AdminUserAdminRevokedEvent struct {
	User *user.User `json:"user"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for AdminUserAdminRevokedEvent
func (e *AdminUserAdminRevokedEvent) Name() string {
	return "admin.user.admin.revoked"
}

// AdminUserStatusChangedEvent represents a user's account status being changed by an admin
type AdminUserStatusChangedEvent struct {
	User      *user.User  `json:"user"`
	Doer      *user.User  `json:"doer"`
	OldStatus user.Status `json:"old_status"`
	NewStatus user.Status `json:"new_status"`
}

// Name defines the name for AdminUserStatusChangedEvent
func (e *AdminUserStatusChangedEvent) Name() string {
	return "admin.user.status.changed"
}

// AdminUserPasswordSetEvent represents an admin setting a user's password.
// It carries no password material.
type AdminUserPasswordSetEvent struct {
	User *user.User `json:"user"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for AdminUserPasswordSetEvent
func (e *AdminUserPasswordSetEvent) Name() string {
	return "admin.user.password.set"
}

// AdminUserPasswordResetSentEvent represents an admin triggering the
// password-reset email for a user. It carries no reset token.
type AdminUserPasswordResetSentEvent struct {
	User *user.User `json:"user"`
	Doer *user.User `json:"doer"`
}

// Name defines the name for AdminUserPasswordResetSentEvent
func (e *AdminUserPasswordResetSentEvent) Name() string {
	return "admin.user.password_reset.sent"
}

// AdminUserDeletedEvent represents a user being deleted through the admin API
type AdminUserDeletedEvent struct {
	User *user.User `json:"user"`
	Doer *user.User `json:"doer"`
	// Mode is "now" for immediate deletion or "scheduled" for the
	// email-confirmation self-deletion flow.
	Mode string `json:"mode"`
}

// Name defines the name for AdminUserDeletedEvent
func (e *AdminUserDeletedEvent) Name() string {
	return "admin.user.deleted"
}

// AdminProjectOwnerChangedEvent represents an admin reassigning a project's owner
type AdminProjectOwnerChangedEvent struct {
	Project    *Project   `json:"project"`
	Doer       *user.User `json:"doer"`
	OldOwnerID int64      `json:"old_owner_id"`
	NewOwnerID int64      `json:"new_owner_id"`
}

// Name defines the name for AdminProjectOwnerChangedEvent
func (e *AdminProjectOwnerChangedEvent) Name() string {
	return "admin.project.owner.changed"
}
