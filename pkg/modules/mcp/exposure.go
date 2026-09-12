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

func toolNameFor(operationID string) string { return strings.ReplaceAll(operationID, "-", "_") }

var typedTools = map[string]bool{
	"labels_create":         true,
	"labels_delete":         true,
	"labels_list":           true,
	"labels_read":           true,
	"labels_update":         true,
	"projects_create":       true,
	"projects_delete":       true,
	"projects_list":         true,
	"projects_read":         true,
	"projects_update":       true,
	"task_assignees_create": true,
	"task_assignees_delete": true,
	"task_assignees_list":   true,
	"task_comments_create":  true,
	"task_comments_delete":  true,
	"task_comments_list":    true,
	"task_comments_read":    true,
	"task_comments_update":  true,
	"tasks_create":          true,
	"tasks_delete":          true,
	"tasks_list":            true,
	"tasks_read":            true,
	"tasks_update":          true,
	"users_search":          true,
}

// Exclude credentials, account management, outbound requests, public shares, and attachment downloads; uploads drop out with the other multipart bodies, while attachment metadata and deletion stay behind the attachments scope.
var deniedOperationPrefixes = []string{
	"admin-",
	"auth-",
	"oauth-",
	"token-",
	"tokens-",
	"caldav-tokens-",
	"sessions-",
	"totp-",
	"user-",
	"bots-",
	"webhooks-",
	"shares-",
	"migration-",
	"backgrounds-",
	"projects-background-",
	"avatar-",
	"testing-",
	"health",
	"info",
	"notifications-atom-feed",
	"task-attachments-download",
}

func exposure(operationID string, op *huma.Operation) (typed bool, ok bool) {
	for _, p := range deniedOperationPrefixes {
		if strings.HasPrefix(operationID, p) {
			return false, false
		}
	}
	if op.RequestBody != nil {
		if _, schema := bodyMedia(op); schema == nil {
			return false, false
		}
	}
	return typedTools[toolNameFor(operationID)], true
}
func bodyMedia(op *huma.Operation) (contentType string, schema *huma.Schema) {
	if op.RequestBody == nil {
		return "", nil
	}
	for _, ct := range []string{
		"application/json",
		"application/merge-patch+json",
	} {
		if m := op.RequestBody.Content[ct]; m != nil && m.Schema != nil {
			return ct, m.Schema
		}
	}
	return "", nil
}
