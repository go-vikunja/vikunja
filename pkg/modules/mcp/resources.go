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

// resources.go owns the central list of MCP-exposed resources. Each entry
// is a pure declaration — name (== API-token scope group), model
// constructor, supported ops, exposure tier — and everything else (input
// schemas, argument application, dispatch) is derived from the model's
// struct tags at registration time. See schema.go for the derivation rules
// and registry.go for the Resource fields.
//
// Scope-filtered tools/list: the SDK calls the getServer factory in
// NewStreamableHTTPHandler exactly once per session (at the initialize
// request) and caches the returned *mcp.Server for the lifetime of that
// session. There is no filter callback in mcp.ServerOptions, so we build a
// per-session *mcp.Server that only registers the tools the requesting
// token's APIPermissions allows. tools/list then naturally returns the
// allowed subset. The dispatcher additionally re-checks scopes on every
// tools/call as a defence-in-depth measure (the same session could in
// principle be reused across requests carrying different tokens).

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var registerResourcesOnce sync.Once

// RegisterResources populates the package-level registry with every
// MCP-exposed resource. It runs at most once per process; subsequent calls
// are no-ops so tests that pre-populate the registry or call this twice
// don't crash on the duplicate-name guard.
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
			// Listing goes through TaskCollection — the REST layer's filter
			// engine — whose filter/sort_by/order_by fields surface via
			// their query tags. models.Task.ReadAll is a no-op stub.
			Models: map[Op]func() handler.CObject{
				OpReadAll: func() handler.CObject { return &models.TaskCollection{} },
			},
			Ops: OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
			// "s" duplicates the reserved search argument; view-scoped
			// listing is polymorphic (buckets vs tasks) and stays REST-only;
			// index is server-assigned despite its readOnly+param tags.
			Exclude: []string{"s", "project_view_id", "index"},
			// Omitting project_id lists tasks across every project the
			// caller can see.
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
			// Live config check so toggling comments doesn't need a restart;
			// the REST routes are gated on the same setting.
			Gate: config.ServiceEnableTaskComments.GetBool,
		},
		{
			// Only the three ops the REST layer supports (PUT/GET-all/DELETE)
			// — there is no per-assignee read_one or update endpoint.
			Name:        "tasks_assignees",
			Description: "Users assigned to a Vikunja task",
			Model:       func() handler.CObject { return &models.TaskAssginee{} },
			Ops:         OpCreate | OpReadAll | OpDelete,
		},

		// Catalog tier — reachable via find_action / do_action only. Ops
		// mirror each resource's REST surface. Deliberately absent: api
		// tokens (self-escalation), webhooks (server-side outbound
		// requests), link shares (public exposure), buckets and task
		// positions (their v1 token scopes don't map onto (group, op)
		// permissions), saved filters (nested filter object).
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

// installToolsForToken registers one tool per (typed resource, op) pair the
// given token's APIPermissions authorise, plus the catalog meta-tools.
// Called from newServer (mcp.go) at session-init time. A nil token (which
// should never happen in production because the entry handler rejects
// unauthenticated requests) yields a server with no tools — defensive, the
// dispatcher would also reject the call.
func installToolsForToken(srv *mcp.Server, token *models.APIToken) {
	for _, r := range snapshotResources() {
		if r.Tier != TierTyped || !r.enabled() {
			continue
		}
		for _, op := range AllOps() {
			if r.Ops&op == 0 || !tokenAuthorizes(token, r.Name, op) {
				continue
			}
			name := r.Name + "_" + op.ToolSuffix()
			srv.AddTool(&mcp.Tool{
				Name:        name,
				Description: r.toolDescription(op),
				InputSchema: r.spec(op).schema,
			}, rawToolHandler(name))
		}
	}
	installCatalogTools(srv, token)
}

// rawToolHandler adapts Dispatch to the SDK's low-level ToolHandler. Domain
// failures (permission denials, missing rows, validation errors) surface as
// IsError tool results per the SDK convention, not as JSON-RPC protocol
// errors. On success the result is serialised once and returned as both
// text content (for clients that ignore structuredContent) and — when it is
// a JSON object, the only shape the field permits — structured output.
func rawToolHandler(toolName string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := Dispatch(ctx, toolName, req.Params.Arguments)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
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
