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
// data lifecycle events as JSONL, with optional forwarding to stdout, syslog
// or webhook sinks.
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
// The local file is the source of truth: a failed file write is returned to
// the router for retry, while forwarder failures are only logged so a dead
// sink cannot poison-queue every event. Tamper evidence comes from filesystem
// permissions (the file is created 0600) plus shipping entries to an external
// sink, not from hash chains or signatures. Rotation is size-based with
// age-based cleanup of rotated files; retention is the operator's concern.
package audit
