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

// Package audit persists an audit trail of authentication, authorization and
// data lifecycle events as JSONL.
//
// Events opt in via RegisterEventForAudit, which subscribes one audit
// listener per event on the existing watermill bus; the event→Entry mapping
// is a closure passed at registration. The catalog of audited events lives in
// registerEventsForAuditLogging in pkg/models/listeners.go.
//
// Entries reference actors and targets by opaque ID only — deleting a user
// row orphans their audit references, which satisfies GDPR erasure without
// log redaction.
//
// Audit logging is gated twice: registration on the audit.enabled config key,
// and each write on the licensed audit_logs feature. The license is checked
// per event because it can change at runtime; enabled-but-unlicensed means
// listeners run and write nothing.
//
// Request attribution (IP, user agent, request id) flows from an Echo
// middleware through the request context onto message metadata — see
// pkg/events.RequestMeta. Events dispatched outside a request get
// source type "system" instead.
//
// A failed file write is returned to the router for retry. Tamper evidence
// comes from filesystem permissions (the file is created 0600) plus shipping
// the file to an external system, not from hash chains or signatures.
// Rotation is size-based with age-based cleanup of rotated files; retention
// is the operator's concern.
package audit

import "time"

// Entry is one audit log record. It only references actors and targets by
// opaque ID — no names, emails or content — so GDPR erasure is satisfied by
// deleting the referenced row.
type Entry struct {
	EventID   string    `json:"event_id"` // UUIDv7
	Timestamp time.Time `json:"timestamp"`
	Actor     Actor     `json:"actor"`
	Source    Source    `json:"source"`
	Action    string    `json:"action"`
	// omitzero: actions without a single affected resource (list reads,
	// denied access) have no target.
	Target    Target         `json:"target,omitzero"`
	Outcome   string         `json:"outcome"`
	Reason    string         `json:"reason,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type actorType string
type targetType string

// Actor is the principal which performed the audited action.
type Actor struct {
	Type actorType `json:"type"`
	ID   int64     `json:"id,omitempty"`
}

// Source describes where the action originated from.
type Source struct {
	Type      string `json:"type"`
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// Target is the resource the audited action was performed on.
type Target struct {
	Type targetType `json:"type"`
	ID   int64      `json:"id,omitempty"`
}

// Outcome values for an Entry.
const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
)

// Source types for an Entry.
const (
	SourceHTTP   = "http"
	SourceSystem = "system"
)

// The action catalog. Every audited action is listed here.
const (
	ActionLoginSucceeded  = "auth.login.succeeded"
	ActionLoginFailed     = "auth.login.failed"
	ActionLogout          = "auth.logout"
	ActionAPITokenIssued  = "auth.api_token.issued"  // #nosec G101 -- action identifier, not a credential
	ActionAPITokenRevoked = "auth.api_token.revoked" // #nosec G101
	ActionAPITokenUsed    = "auth.api_token.used"    // #nosec G101

	ActionUserCreated = "user.created"

	ActionTaskCreated           = "task.created"
	ActionTaskUpdated           = "task.updated"
	ActionTaskDeleted           = "task.deleted"
	ActionTaskAssigneeAdded     = "task.assignee.added"
	ActionTaskAssigneeRemoved   = "task.assignee.removed"
	ActionTaskCommentCreated    = "task.comment.created"
	ActionTaskCommentUpdated    = "task.comment.updated"
	ActionTaskCommentDeleted    = "task.comment.deleted"
	ActionTaskAttachmentCreated = "task.attachment.created"
	ActionTaskAttachmentDeleted = "task.attachment.deleted"
	ActionTaskRelationCreated   = "task.relation.created"
	ActionTaskRelationDeleted   = "task.relation.deleted"

	ActionProjectCreated        = "project.created"
	ActionProjectUpdated        = "project.updated"
	ActionProjectDeleted        = "project.deleted"
	ActionProjectSharedWithUser = "project.shared.user"
	ActionProjectSharedWithTeam = "project.shared.team"

	ActionTeamCreated       = "team.created"
	ActionTeamDeleted       = "team.deleted"
	ActionTeamMemberAdded   = "team.member.added"
	ActionTeamMemberRemoved = "team.member.removed"

	ActionAdminUserCreated           = "admin.user.created"
	ActionAdminUserAdminGranted      = "admin.user.admin.granted"
	ActionAdminUserAdminRevoked      = "admin.user.admin.revoked"
	ActionAdminUserStatusChanged     = "admin.user.status.changed"
	ActionAdminUserPasswordSet       = "admin.user.password.set"        // #nosec G101 -- action identifier, not a credential
	ActionAdminUserPasswordResetSent = "admin.user.password_reset.sent" // #nosec G101
	ActionAdminUserDeleted           = "admin.user.deleted"
	ActionAdminProjectOwnerChanged   = "admin.project.owner.changed"
	ActionAdminUsersListed           = "admin.users.listed"
	ActionAdminAccessDenied          = "admin.access.denied"
)

// The type strings are unexported; these constructors are the only way to
// build an Actor or Target, so a mismatched type/ID pair can't be expressed.

func UserActor(id int64) Actor      { return Actor{Type: "user", ID: id} }
func LinkShareActor(id int64) Actor { return Actor{Type: "link_share", ID: id} }
func SystemActor() Actor            { return Actor{Type: "system"} }

// ActorFromDoerID maps a doer ID to an actor. Link shares are disguised as
// users with negative IDs throughout the event payloads.
func ActorFromDoerID(id int64) Actor {
	if id < 0 {
		return LinkShareActor(-id)
	}
	return UserActor(id)
}

func TaskTarget(id int64) Target     { return Target{Type: "task", ID: id} }
func ProjectTarget(id int64) Target  { return Target{Type: "project", ID: id} }
func UserTarget(id int64) Target     { return Target{Type: "user", ID: id} }
func TeamTarget(id int64) Target     { return Target{Type: "team", ID: id} }
func APITokenTarget(id int64) Target { return Target{Type: "api_token", ID: id} }
