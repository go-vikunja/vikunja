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

// The action catalog: TierCatalog resources don't get first-class tools in
// tools/list — they're reachable through two meta-tools instead, keeping the
// per-session tool list (and the tokens it costs an LLM client) small while
// still exposing the long tail of CRUD resources.
//
//   - find_action lists the catalog actions the requesting token's scopes
//     authorise. Without arguments it returns a cheap name+description
//     index; naming an action or resource returns full input schemas.
//   - do_action invokes one action by name. It funnels into the same
//     Dispatch path as the typed tools, so schema validation and the
//     per-call scope re-check apply identically.

import (
	"context"
	"encoding/json"
	"fmt"

	"code.vikunja.io/api/pkg/models"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolFindAction = "find_action"
	toolDoAction   = "do_action"
)

// actionInfo is one find_action result entry.
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

func findActionSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"action":   {Type: "string", Description: "Return the full input schema for this single action (e.g. tasks_labels_create)."},
			"resource": {Type: "string", Description: "Return the full input schemas for every action of this resource (e.g. tasks_labels)."},
		},
		AdditionalProperties: falseSchema(),
	}
}

func doActionSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"action":    {Type: "string", Description: "The action to invoke, as returned by find_action (e.g. tasks_labels_create)."},
			"arguments": {Type: "object", Description: "The action's arguments, matching the input_schema find_action returned for it."},
		},
		Required:             []string{"action"},
		AdditionalProperties: falseSchema(),
	}
}

// installCatalogTools registers the two meta-tools. They're always present
// for an mcp:access token; a token with no catalog scopes just gets an
// empty find_action result, and do_action re-checks scopes per call.
func installCatalogTools(srv *mcp.Server, token *models.APIToken) {
	srv.AddTool(&mcp.Tool{
		Name: toolFindAction,
		Description: "Discover additional Vikunja actions beyond the tools listed here: sharing projects with users or teams, task labels and relations (subtasks), team members, project views and more. " +
			"Returns the actions your token authorises; pass action or resource to get full input schemas. Invoke them with do_action.",
		InputSchema: findActionSchema(),
	}, findActionHandler(token))

	srv.AddTool(&mcp.Tool{
		Name:        toolDoAction,
		Description: "Invoke an action discovered via find_action. Arguments must match the action's input_schema.",
		InputSchema: doActionSchema(),
	}, doActionHandler)
}

// catalogActions returns the catalog entries the token authorises,
// optionally filtered to one action or resource, with schemas attached when
// the filter is specific enough to keep the payload small.
func catalogActions(token *models.APIToken, action, resource string) []actionInfo {
	withSchemas := action != "" || resource != ""
	out := []actionInfo{}
	for _, r := range snapshotResources() {
		if r.Tier != TierCatalog || !r.enabled() {
			continue
		}
		if resource != "" && r.Name != resource {
			continue
		}
		for _, op := range AllOps() {
			if r.Ops&op == 0 || !tokenAuthorizes(token, r.Name, op) {
				continue
			}
			name := r.Name + "_" + op.ToolSuffix()
			if action != "" && name != action {
				continue
			}
			info := actionInfo{Name: name, Description: r.toolDescription(op)}
			if withSchemas {
				info.InputSchema = r.spec(op).schema
			}
			out = append(out, info)
		}
	}
	return out
}

func findActionHandler(token *models.APIToken) mcp.ToolHandler {
	return func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args findActionArgs
		if len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				//nolint:nilerr // IsError tool result, not a JSON-RPC protocol error
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: "invalid arguments: " + err.Error()}},
				}, nil
			}
		}

		result := map[string]any{"actions": catalogActions(token, args.Action, args.Resource)}
		body, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("mcp: marshal find_action result: %w", err)
		}
		return &mcp.CallToolResult{
			Content:           []mcp.Content{&mcp.TextContent{Text: string(body)}},
			StructuredContent: result,
		}, nil
	}
}

func doActionHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args doActionArgs
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil || args.Action == "" {
		//nolint:nilerr // IsError tool result, not a JSON-RPC protocol error
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "do_action requires an \"action\" name; discover actions with find_action"}},
		}, nil
	}
	// Dispatch validates the arguments and re-checks the token's scope, so
	// do_action can't reach anything a direct tool call couldn't.
	return rawToolHandler(args.Action)(ctx, &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: args.Action, Arguments: args.Arguments},
	})
}
