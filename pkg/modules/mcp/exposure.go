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

package mcp

import (
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type Tier uint8

const (
	TierTyped Tier = iota
	TierCatalog
)

func toolNameFor(operationID string) string { return strings.ReplaceAll(operationID, "-", "_") }

var typedTools = map[string]bool{
	"projects_list": true, "projects_read": true, "projects_create": true, "projects_update": true, "projects_delete": true,
	"tasks_list": true, "tasks_read": true, "tasks_create": true, "tasks_update": true, "tasks_delete": true,
	"labels_list": true, "labels_read": true, "labels_create": true, "labels_update": true, "labels_delete": true,
	"task_comments_list": true, "task_comments_read": true, "task_comments_create": true, "task_comments_update": true, "task_comments_delete": true,
	"task_assignees_list": true, "task_assignees_create": true, "task_assignees_delete": true, "users_search": true,
}

// Exclude credentials, account management, outbound requests, public shares, and file transfers.
var deniedOperationPrefixes = []string{
	"admin-", "auth-", "oauth-", "token-", "tokens-", "caldav-tokens-", "sessions-", "totp-",
	"user-", "bots-", "webhooks-", "shares-", "migration-", "backgrounds-", "projects-background-",
	"avatar-", "testing-", "health", "info", "notifications-atom-feed", "task-attachments-download",
}

func exposure(operationID string, op *huma.Operation) (Tier, bool) {
	for _, p := range deniedOperationPrefixes {
		if strings.HasPrefix(operationID, p) {
			return 0, false
		}
	}
	if op.RequestBody != nil {
		if _, schema := bodyMedia(op); schema == nil {
			return 0, false
		}
	}
	if typedTools[toolNameFor(operationID)] {
		return TierTyped, true
	}
	return TierCatalog, true
}
func bodyMedia(op *huma.Operation) (contentType string, schema *huma.Schema) {
	if op.RequestBody == nil {
		return "", nil
	}
	for _, ct := range []string{"application/json", "application/merge-patch+json"} {
		if m := op.RequestBody.Content[ct]; m != nil && m.Schema != nil {
			return ct, m.Schema
		}
	}
	return "", nil
}
