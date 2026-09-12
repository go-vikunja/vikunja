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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/models"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolFindAction = "find_action"
	toolDoAction   = "do_action"
)

type actionInfo struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	InputSchema *jsonschema.Schema `json:"input_schema,omitempty"`
}
type findActionArgs struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}
type doActionArgs struct {
	Action    string          `json:"action"`
	Arguments json.RawMessage `json:"arguments"`
}

var findActionSpec = mustResolveSpec(toolFindAction, &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"action": {
			Type:        "string",
			Description: "Return the full input schema for this action (e.g. task_labels_create).",
		},
		"resource": {
			Type:        "string",
			Description: "Return full schemas for actions under this prefix (e.g. task_labels, project_views, teams).",
		},
	},
	AdditionalProperties: falseSchema(),
})
var doActionSpec = mustResolveSpec(toolDoAction, &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"action": {
			Type:        "string",
			Description: "The action returned by find_action (e.g. task_labels_create).",
		},
		"arguments": {
			Type:        "object",
			Description: "Arguments matching the action's input_schema from find_action.",
		},
	},
	Required:             []string{"action"},
	AdditionalProperties: falseSchema(),
})

func installCatalogTools(srv *mcp.Server) {
	srv.AddTool(&mcp.Tool{
		Name:        toolFindAction,
		Description: "Discover additional Vikunja actions: project sharing with users or teams, task labels and relations (subtasks), teams and members, project views, saved filters, time entries and more. Returns actions your token authorizes; pass action or resource for full input schemas. Invoke them with do_action.",
		InputSchema: findActionSpec.schema,
	}, findActionHandler)
	srv.AddTool(&mcp.Tool{
		Name:        toolDoAction,
		Description: "Invoke an action discovered via find_action. Arguments must match its input_schema.",
		InputSchema: doActionSpec.schema,
	}, doActionHandler)
}
func catalogActions(token *models.APIToken, action, resource string) []actionInfo {
	out := []actionInfo{}
	for _, t := range snapshotTools() {
		if t.tier != TierCatalog || !t.authorized(token) {
			continue
		}
		if action != "" && t.name != action {
			continue
		}
		if resource != "" && !strings.HasPrefix(t.name, resource+"_") {
			continue
		}
		info := actionInfo{
			Name:        t.name,
			Description: t.description,
		}
		if action != "" || resource != "" {
			info.InputSchema = t.spec.schema
		}
		out = append(out, info)
	}
	return out
}
func invalidArgsResult(name string, err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("mcp: invalid arguments for %s: %v", name, err)}},
	}
}
func findActionHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args findActionArgs
	if err := decodeToolArgs(findActionSpec, req.Params.Arguments, &args); err != nil {
		//nolint:nilerr // Domain errors use MCP tool results.
		return invalidArgsResult(toolFindAction, err), nil
	}
	result := map[string]any{"actions": catalogActions(TokenFromContext(ctx), args.Action, args.Resource)}
	body, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("mcp: marshal find_action result: %w", err)
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(body)}},
		StructuredContent: result,
	}, nil
}
func doActionHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args doActionArgs
	if err := decodeToolArgs(doActionSpec, req.Params.Arguments, &args); err != nil {
		//nolint:nilerr // Domain errors use MCP tool results.
		return invalidArgsResult(toolDoAction, err), nil
	}
	return rawToolHandler(args.Action)(ctx, &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      args.Action,
			Arguments: args.Arguments,
		},
	})
}
