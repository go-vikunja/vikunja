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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// Permission Delegation Function Variables
//
// These function variables are set by the service layer during initialization
// to avoid circular import dependencies. Models use these to delegate permission
// checks to the service layer without directly importing the services package.
//
// Pattern:
// - Service layer sets these function variables during initialization
// - Model permission methods call these functions instead of implementing logic
// - Service layer implements the actual permission checking logic
//
// Migration Status:
// These variables will be populated incrementally as permission methods are migrated:
// - T-PERM-006: Project permissions
// - T-PERM-007: Task permissions
// - T-PERM-008: Label & Kanban permissions
// - T-PERM-009: Link Share & Subscription permissions
// - T-PERM-010: Task Relations permissions
// - T-PERM-011: Project Relations permissions
// - T-PERM-012: Misc permissions (API tokens, teams, webhooks, etc.)

var (
	// Project permission delegation
	// These will be set in T-PERM-006
	CheckProjectReadFunc    func(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error)
	CheckProjectWriteFunc   func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	CheckProjectUpdateFunc  func(s *xorm.Session, projectID int64, project *Project, a web.Auth) (bool, error)
	CheckProjectDeleteFunc  func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	CheckProjectCreateFunc  func(s *xorm.Session, project *Project, a web.Auth) (bool, error)
	CheckProjectIsAdminFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)

	// Task permission delegation
	// These will be set in T-PERM-007
	CheckTaskReadFunc   func(s *xorm.Session, taskID int64, a web.Auth) (bool, int, error)
	CheckTaskWriteFunc  func(s *xorm.Session, taskID int64, a web.Auth) (bool, error)
	CheckTaskUpdateFunc func(s *xorm.Session, taskID int64, task *Task, a web.Auth) (bool, error)
	CheckTaskDeleteFunc func(s *xorm.Session, taskID int64, a web.Auth) (bool, error)
	CheckTaskCreateFunc func(s *xorm.Session, task *Task, a web.Auth) (bool, error)

	// Label permission delegation
	// These will be set in T-PERM-008
	CheckLabelReadFunc   func(s *xorm.Session, labelID int64, a web.Auth) (bool, int, error)
	CheckLabelWriteFunc  func(s *xorm.Session, labelID int64, a web.Auth) (bool, error)
	CheckLabelUpdateFunc func(s *xorm.Session, labelID int64, a web.Auth) (bool, error)
	CheckLabelDeleteFunc func(s *xorm.Session, labelID int64, a web.Auth) (bool, error)
	CheckLabelCreateFunc func(s *xorm.Session, label *Label, a web.Auth) (bool, error)

	// Kanban/Bucket permission delegation
	// These will be set in T-PERM-008
	// Note: Update/Delete pass the whole bucket struct to access ProjectID from URL binding (T-PERM-016B fix)
	CheckBucketReadFunc   func(s *xorm.Session, bucketID int64, a web.Auth) (bool, int, error)
	CheckBucketWriteFunc  func(s *xorm.Session, bucketID int64, a web.Auth) (bool, error)
	CheckBucketUpdateFunc func(s *xorm.Session, bucket *Bucket, a web.Auth) (bool, error)
	CheckBucketDeleteFunc func(s *xorm.Session, bucket *Bucket, a web.Auth) (bool, error)
	CheckBucketCreateFunc func(s *xorm.Session, bucket *Bucket, a web.Auth) (bool, error)

	// LinkSharing permission delegation
	// These will be set in T-PERM-009
	CheckLinkShareReadFunc   func(s *xorm.Session, shareID int64, a web.Auth) (bool, int, error)
	CheckLinkShareWriteFunc  func(s *xorm.Session, shareID int64, a web.Auth) (bool, error)
	CheckLinkShareUpdateFunc func(s *xorm.Session, shareID int64, a web.Auth) (bool, error)
	CheckLinkShareDeleteFunc func(s *xorm.Session, shareID int64, a web.Auth) (bool, error)
	CheckLinkShareCreateFunc func(s *xorm.Session, share *LinkSharing, a web.Auth) (bool, error)

	// Subscription permission delegation
	// These will be set in T-PERM-009
	CheckSubscriptionCreateFunc func(s *xorm.Session, subscription *Subscription, a web.Auth) (bool, error)
	CheckSubscriptionDeleteFunc func(s *xorm.Session, subscription *Subscription, a web.Auth) (bool, error)

	// TaskComment permission delegation
	// These will be set in T-PERM-010
	CheckTaskCommentReadFunc   func(s *xorm.Session, commentID int64, a web.Auth) (bool, int, error)
	CheckTaskCommentWriteFunc  func(s *xorm.Session, commentID int64, a web.Auth) (bool, error)
	CheckTaskCommentUpdateFunc func(s *xorm.Session, comment *TaskComment, a web.Auth) (bool, error)
	CheckTaskCommentDeleteFunc func(s *xorm.Session, comment *TaskComment, a web.Auth) (bool, error)
	CheckTaskCommentCreateFunc func(s *xorm.Session, comment *TaskComment, a web.Auth) (bool, error)

	// TaskAttachment permission delegation
	// These will be set in T-PERM-010
	CheckTaskAttachmentReadFunc   func(s *xorm.Session, attachmentID int64, a web.Auth) (bool, int, error)
	CheckTaskAttachmentDeleteFunc func(s *xorm.Session, attachmentID int64, a web.Auth) (bool, error)
	CheckTaskAttachmentCreateFunc func(s *xorm.Session, attachment *TaskAttachment, a web.Auth) (bool, error)

	// TaskRelation permission delegation
	// These will be set in T-PERM-010
	CheckTaskRelationCreateFunc func(s *xorm.Session, relation *TaskRelation, a web.Auth) (bool, error)
	CheckTaskRelationDeleteFunc func(s *xorm.Session, relation *TaskRelation, a web.Auth) (bool, error)

	// TaskAssignee permission delegation
	// These will be set in T-PERM-010
	CheckTaskAssigneeCreateFunc func(s *xorm.Session, assignee *TaskAssginee, a web.Auth) (bool, error)
	CheckTaskAssigneeDeleteFunc func(s *xorm.Session, assignee *TaskAssginee, a web.Auth) (bool, error)

	// LabelTask permission delegation
	// These will be set in T-PERM-010
	CheckLabelTaskCreateFunc func(s *xorm.Session, labelTask *LabelTask, a web.Auth) (bool, error)
	CheckLabelTaskDeleteFunc func(s *xorm.Session, labelTask *LabelTask, a web.Auth) (bool, error)

	// ProjectTeam permission delegation
	// These will be set in T-PERM-011
	CheckProjectTeamReadFunc   func(s *xorm.Session, projectID int64, teamID int64, a web.Auth) (bool, error)
	CheckProjectTeamWriteFunc  func(s *xorm.Session, projectID int64, teamID int64, a web.Auth) (bool, error)
	CheckProjectTeamUpdateFunc func(s *xorm.Session, projectID int64, teamID int64, a web.Auth) (bool, error)
	CheckProjectTeamDeleteFunc func(s *xorm.Session, projectID int64, teamID int64, a web.Auth) (bool, error)
	CheckProjectTeamCreateFunc func(s *xorm.Session, teamProject *TeamProject, a web.Auth) (bool, error)

	// ProjectUser permission delegation
	// These will be set in T-PERM-011
	CheckProjectUserReadFunc   func(s *xorm.Session, projectID int64, userID int64, a web.Auth) (bool, error)
	CheckProjectUserWriteFunc  func(s *xorm.Session, projectID int64, userID int64, a web.Auth) (bool, error)
	CheckProjectUserUpdateFunc func(s *xorm.Session, projectID int64, userID int64, a web.Auth) (bool, error)
	CheckProjectUserDeleteFunc func(s *xorm.Session, projectID int64, userID int64, a web.Auth) (bool, error)
	CheckProjectUserCreateFunc func(s *xorm.Session, projectUser *ProjectUser, a web.Auth) (bool, error)

	// ProjectView permission delegation
	// These will be set in T-PERM-011
	CheckProjectViewReadFunc   func(s *xorm.Session, viewID int64, a web.Auth) (bool, int, error)
	CheckProjectViewWriteFunc  func(s *xorm.Session, viewID int64, a web.Auth) (bool, error)
	CheckProjectViewUpdateFunc func(s *xorm.Session, viewID int64, a web.Auth) (bool, error)
	CheckProjectViewDeleteFunc func(s *xorm.Session, viewID int64, a web.Auth) (bool, error)
	CheckProjectViewCreateFunc func(s *xorm.Session, view *ProjectView, a web.Auth) (bool, error)

	// Miscellaneous entity permissions (T-PERM-012)

	// APIToken permission delegation
	CheckAPITokenDeleteFunc func(s *xorm.Session, tokenID int64, a web.Auth) (bool, error)

	// Reaction permission delegation
	CheckReactionCreateFunc func(s *xorm.Session, reaction *Reaction, a web.Auth) (bool, error)
	CheckReactionDeleteFunc func(s *xorm.Session, reaction *Reaction, a web.Auth) (bool, error)

	// SavedFilter permission delegation
	CheckSavedFilterReadFunc   func(s *xorm.Session, filterID int64, a web.Auth) (bool, int, error)
	CheckSavedFilterWriteFunc  func(s *xorm.Session, filterID int64, a web.Auth) (bool, error)
	CheckSavedFilterUpdateFunc func(s *xorm.Session, filterID int64, a web.Auth) (bool, error)
	CheckSavedFilterDeleteFunc func(s *xorm.Session, filterID int64, a web.Auth) (bool, error)
	CheckSavedFilterCreateFunc func(s *xorm.Session, filter *SavedFilter, a web.Auth) (bool, error)

	// Team permission delegation
	CheckTeamReadFunc   func(s *xorm.Session, teamID int64, a web.Auth) (bool, int, error)
	CheckTeamWriteFunc  func(s *xorm.Session, teamID int64, a web.Auth) (bool, error)
	CheckTeamUpdateFunc func(s *xorm.Session, teamID int64, a web.Auth) (bool, error)
	CheckTeamDeleteFunc func(s *xorm.Session, teamID int64, a web.Auth) (bool, error)
	CheckTeamCreateFunc func(s *xorm.Session, team *Team, a web.Auth) (bool, error)

	// TeamMember permission delegation
	CheckTeamMemberCreateFunc func(s *xorm.Session, member *TeamMember, a web.Auth) (bool, error)
	CheckTeamMemberDeleteFunc func(s *xorm.Session, member *TeamMember, a web.Auth) (bool, error)

	// Webhook permission delegation
	CheckWebhookReadFunc   func(s *xorm.Session, webhookID int64, a web.Auth) (bool, error)
	CheckWebhookUpdateFunc func(s *xorm.Session, webhookID int64, a web.Auth) (bool, error)
	CheckWebhookDeleteFunc func(s *xorm.Session, webhookID int64, a web.Auth) (bool, error)
	CheckWebhookCreateFunc func(s *xorm.Session, webhook *Webhook, a web.Auth) (bool, error)

	// BulkTask permission delegation
	CheckBulkTaskUpdateFunc func(s *xorm.Session, taskIDs []int64, a web.Auth) (bool, error)

	// ProjectDuplicate permission delegation
	CheckProjectDuplicateCreateFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)

	// TaskPosition permission delegation
	CheckTaskPositionUpdateFunc func(s *xorm.Session, taskID int64, a web.Auth) (bool, error)

	// ProjectView permission delegation (T-PERM-011)
	ProjectViewCanReadFunc   func(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error)
	ProjectViewCanCreateFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	ProjectViewCanUpdateFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	ProjectViewCanDeleteFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
)
