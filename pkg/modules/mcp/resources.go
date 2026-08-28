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

// The SDK caches getServer's *mcp.Server per session and has no tools/list
// filter callback, so each session gets a server carrying only the tools the
// requesting token's scopes allow. Dispatch re-checks scopes per call anyway.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var registerResourcesOnce sync.Once

// RegisterResources runs at most once per process so repeat calls don't trip the duplicate-name guard.
func RegisterResources() {
	registerResourcesOnce.Do(func() {
		for _, r := range allResources() {
			if err := Register(r); err != nil {
				panic(fmt.Errorf("mcp: failed to register %s resource: %w", r.Name, err))
			}
		}
	})
}

func allResources() []Resource {
	return []Resource{
		{
			Name:        "projects",
			Description: "Vikunja projects (containers for tasks)",
			Model:       func() handler.CObject { return &models.Project{} },
			Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		},
		{
			Name:        "tasks",
			Description: "Vikunja tasks (work items inside a project)",
			Model:       func() handler.CObject { return &models.Task{} },
			// models.Task.ReadAll is a stub; TaskCollection is the real filter engine.
			Models: map[Op]func() handler.CObject{
				OpReadAll: func() handler.CObject { return &models.TaskCollection{} },
			},
			Ops: OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
			// "s" duplicates the reserved search arg, view-scoped listing is polymorphic and stays REST-only, index is server-assigned.
			Exclude: []string{"s", "project_view_id", "index"},
			// Omitting project_id lists tasks across every project the caller can see.
			OptionalFields: []string{"project_id"},
		},
		{
			Name:        "labels",
			Description: "Vikunja labels (reusable tags attachable to tasks)",
			Model:       func() handler.CObject { return &models.Label{} },
			Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		},
		{
			Name:        "teams",
			Description: "Vikunja teams (groups of users that can share projects)",
			Model:       func() handler.CObject { return &models.Team{} },
			Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		},
		{
			Name:        "tasks_comments",
			Description: "Comments attached to a Vikunja task",
			Model:       func() handler.CObject { return &models.TaskComment{} },
			Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
			// Live check so toggling comments doesn't need a restart, matching the REST routes.
			Gate: config.ServiceEnableTaskComments.GetBool,
		},
		{
			// The REST layer has no per-assignee read_one or update endpoint.
			Name:        "tasks_assignees",
			Description: "Users assigned to a Vikunja task",
			Model:       func() handler.CObject { return &models.TaskAssginee{} },
			Ops:         OpCreate | OpReadAll | OpDelete,
		},

		// Deliberately absent from the catalog: api tokens (self-escalation),
		// webhooks and link shares (outbound requests / public exposure),
		// buckets, task positions and saved filters (scopes don't map onto (group, op)).
		{
			Name:        "tasks_labels",
			Description: "Labels attached to a Vikunja task; create adds a label, delete removes it",
			Model:       func() handler.CObject { return &models.LabelTask{} },
			Ops:         OpCreate | OpReadAll | OpDelete,
			Tier:        TierCatalog,
		},
		{
			Name:        "tasks_relations",
			Description: "Relations between Vikunja tasks (subtask, parenttask, blocking, related, …)",
			Model:       func() handler.CObject { return &models.TaskRelation{} },
			Ops:         OpCreate | OpDelete,
			Tier:        TierCatalog,
		},
		{
			Name:           "teams_members",
			Description:    "Members of a Vikunja team, addressed by team id and username",
			Model:          func() handler.CObject { return &models.TeamMember{} },
			Ops:            OpCreate | OpDelete,
			Tier:           TierCatalog,
			IdentityFields: []string{"username"},
		},
		{
			Name:           "projects_users",
			Description:    "Users a Vikunja project is shared with, addressed by project id and username",
			Model:          func() handler.CObject { return &models.ProjectUser{} },
			Ops:            OpCreate | OpReadAll | OpUpdate | OpDelete,
			Tier:           TierCatalog,
			IdentityFields: []string{"username"},
		},
		{
			Name:           "projects_teams",
			Description:    "Teams a Vikunja project is shared with, addressed by project id and team id",
			Model:          func() handler.CObject { return &models.TeamProject{} },
			Ops:            OpCreate | OpReadAll | OpUpdate | OpDelete,
			Tier:           TierCatalog,
			IdentityFields: []string{"team_id"},
		},
		{
			Name:           "projects_views",
			Description:    "Views of a Vikunja project (list, gantt, table, kanban)",
			Model:          func() handler.CObject { return &models.ProjectView{} },
			Ops:            OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
			Tier:           TierCatalog,
			IdentityFields: []string{"id", "project_id"},
		},
	}
}

// A nil token yields a server with no tools; the entry handler already rejects unauthenticated requests.
func installToolsForToken(srv *mcp.Server, token *models.APIToken) {
	for _, r := range snapshotResources() {
		if r.Tier != TierTyped || !r.enabled() {
			continue
		}
		for _, op := range AllOps() {
			if r.Ops&op == 0 || !tokenAuthorizes(token, r.Name, op) {
				continue
			}
			name := r.Name + "_" + op.Permission()
			srv.AddTool(&mcp.Tool{
				Name:        name,
				Description: r.toolDescription(op),
				InputSchema: r.spec(op).schema,
			}, rawToolHandler(name))
		}
	}
	installUsersSearchTool(srv, token)
	installCatalogTools(srv)
}

// Domain failures surface as IsError tool results, not JSON-RPC protocol errors.
// Results go out as text content too, for clients that ignore structuredContent.
func rawToolHandler(toolName string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := Dispatch(ctx, toolName, req.Params.Arguments)
		if err != nil {
			//nolint:nilerr // IsError tool result, not a JSON-RPC protocol error
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: toolErrorText(err)}},
			}, nil
		}
		body, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return nil, fmt.Errorf("mcp: marshal %s result: %w", toolName, marshalErr)
		}
		res := &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
		}
		if len(body) > 0 && body[0] == '{' {
			res.StructuredContent = result
		}
		return res, nil
	}
}

// govalidator echoes the rejected value back in its message, so a field entry
// can be as large as the payload that failed; a tool result goes straight into
// the model's context window.
const maxInvalidFieldRunes = 200

// ValidationHTTPError keeps the offending field names out of Error(), but a tool result is plain text.
func toolErrorText(err error) string {
	var invalid models.ValidationHTTPError
	if !errors.As(err, &invalid) || len(invalid.InvalidFields) == 0 {
		return err.Error()
	}
	fields := make([]string, len(invalid.InvalidFields))
	for i, f := range invalid.InvalidFields {
		fields[i] = truncateRunes(f, maxInvalidFieldRunes)
	}
	return err.Error() + ": " + strings.Join(fields, "; ")
}

func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "…"
}
