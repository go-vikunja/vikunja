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

type toolTier int

const (
	catalogTool toolTier = iota
	typedTool
)

// Anything not listed stays off MCP; credentials, account, webhooks, shares, file transfer and admin routes are deliberately absent.
var exposedOperations = map[string]toolTier{
	"buckets-create":                  catalogTool,
	"buckets-delete":                  catalogTool,
	"buckets-list":                    catalogTool,
	"buckets-update":                  catalogTool,
	"filters-create":                  catalogTool,
	"filters-delete":                  catalogTool,
	"filters-read":                    catalogTool,
	"filters-update":                  catalogTool,
	"labels-create":                   typedTool,
	"labels-delete":                   typedTool,
	"labels-list":                     typedTool,
	"labels-read":                     typedTool,
	"labels-update":                   typedTool,
	"notifications-delete-all":        catalogTool,
	"notifications-list":              catalogTool,
	"notifications-mark-all-read":     catalogTool,
	"notifications-mark-read":         catalogTool,
	"project-tasks-list":              catalogTool,
	"project-teams-create":            catalogTool,
	"project-teams-delete":            catalogTool,
	"project-teams-list":              catalogTool,
	"project-teams-update":            catalogTool,
	"project-time-entries-list":       catalogTool,
	"project-users-create":            catalogTool,
	"project-users-delete":            catalogTool,
	"project-users-list":              catalogTool,
	"project-users-update":            catalogTool,
	"project-view-buckets-tasks-list": catalogTool,
	"project-view-tasks-list":         catalogTool,
	"project-views-create":            catalogTool,
	"project-views-delete":            catalogTool,
	"project-views-list":              catalogTool,
	"project-views-read":              catalogTool,
	"project-views-update":            catalogTool,
	"projects-create":                 typedTool,
	"projects-delete":                 typedTool,
	"projects-duplicate":              catalogTool,
	"projects-list":                   typedTool,
	"projects-read":                   typedTool,
	"projects-update":                 typedTool,
	"projects-users-search":           catalogTool,
	"reactions-create":                catalogTool,
	"reactions-delete":                catalogTool,
	"reactions-list":                  catalogTool,
	"task-assignees-bulk":             catalogTool,
	"task-assignees-create":           typedTool,
	"task-assignees-delete":           typedTool,
	"task-assignees-list":             typedTool,
	"task-attachments-delete":         catalogTool,
	"task-attachments-list":           catalogTool,
	"task-bucket-update":              catalogTool,
	"task-comments-create":            typedTool,
	"task-comments-delete":            typedTool,
	"task-comments-list":              typedTool,
	"task-comments-read":              typedTool,
	"task-comments-update":            typedTool,
	"task-labels-bulk-replace":        catalogTool,
	"task-labels-create":              catalogTool,
	"task-labels-delete":              catalogTool,
	"task-labels-list":                catalogTool,
	"task-time-entries-list":          catalogTool,
	"tasks-bulk-create":               catalogTool,
	"tasks-bulk-update":               catalogTool,
	"tasks-create":                    typedTool,
	"tasks-delete":                    typedTool,
	"tasks-duplicate":                 catalogTool,
	"tasks-list":                      typedTool,
	"tasks-mark-read":                 catalogTool,
	"tasks-position-update":           catalogTool,
	"tasks-read":                      typedTool,
	"tasks-read-by-index":             catalogTool,
	"tasks-relations-create":          catalogTool,
	"tasks-relations-delete":          catalogTool,
	"tasks-update":                    typedTool,
	"teams-create":                    catalogTool,
	"teams-delete":                    catalogTool,
	"teams-list":                      catalogTool,
	"teams-members-add":               catalogTool,
	"teams-members-remove":            catalogTool,
	"teams-members-toggle-admin":      catalogTool,
	"teams-read":                      catalogTool,
	"teams-update":                    catalogTool,
	"time-entries-create":             catalogTool,
	"time-entries-delete":             catalogTool,
	"time-entries-list":               catalogTool,
	"time-entries-read":               catalogTool,
	"time-entries-timer-stop":         catalogTool,
	"time-entries-update":             catalogTool,
	"users-search":                    typedTool,
}

// ExposedToolNames are the tool names the allow-list can produce. Exported so tests can check it against the live v2 routes.
func ExposedToolNames() []string {
	names := make([]string, 0, len(exposedOperations))
	for id := range exposedOperations {
		names = append(names, toolNameFor(id))
	}
	return names
}

func exposure(operationID string, op *huma.Operation) (typed bool, ok bool) {
	tier, exposed := exposedOperations[operationID]
	if !exposed {
		return false, false
	}
	if op.RequestBody != nil {
		if _, schema := bodyMedia(op); schema == nil {
			return false, false
		}
	}
	return tier == typedTool, true
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
