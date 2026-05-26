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
	"sync"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// resources.go owns the central list of MCP-exposed resources. Each entry
// declares: the resource name (matches the API-token scope group), the
// model's EmptyStruct, the set of supported ops, and the per-op input
// wrappers from inputs.go.
//
// RegisterResources is idempotent and safe to call multiple times — the
// registry's duplicate check is converted to a no-op so the function works
// both at production startup (via newServer) and in repeated test setups
// that reset the registry between cases.
//
// installTools walks the registry and registers a typed *mcp.Tool on the
// given server for every (resource, op) pair. The per-op wrapper type is
// hard-coded into a generic addTool helper so the SDK can reflect the input
// schema at registration time — there is no way to feed reflect.Type into
// the AddTool generics at runtime.

var registerResourcesOnce sync.Once

// RegisterResources populates the package-level registry with every
// MCP-exposed resource. It runs at most once per process; subsequent calls
// are no-ops so tests that pre-populate the registry or call this twice
// don't crash on the duplicate-name guard.
func RegisterResources() {
	registerResourcesOnce.Do(func() {
		if err := registerProjects(); err != nil {
			panic(fmt.Errorf("mcp: failed to register projects resource: %w", err))
		}
	})
}

func registerProjects() error {
	return Register(Resource{
		Name:        "projects",
		Description: "Vikunja projects (containers for tasks)",
		EmptyStruct: func() handler.CObject { return &models.Project{} },
		Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &ProjectCreateInput{},
			OpReadOne: &ReadOneInput{},
			OpReadAll: &ReadAllInput{},
			OpUpdate:  &ProjectUpdateInput{},
			OpDelete:  &DeleteInput{},
		},
	})
}

// installTools walks the registry and binds each enabled (resource, op)
// pair to a tool on the given server. Per-op wrapper types are known at
// compile time, so a per-resource installer is the cleanest way to keep the
// SDK's compile-time type parameter happy while the registry stays
// data-driven elsewhere.
//
// Called from newServer (mcp.go); every fresh MCP session gets the full
// tool set. Per-token scope filtering is layered on top in Task 6.
func installTools(srv *mcp.Server) {
	installProjectsTools(srv)
}

func installProjectsTools(srv *mcp.Server) {
	r, ok := lookupResource("projects")
	if !ok {
		// Defensive: RegisterResources must run before installTools.
		// A missing resource means programmer error, not a runtime
		// condition the caller can recover from.
		panic("mcp: projects resource not registered")
	}

	if r.Ops&OpCreate != 0 {
		addTool[*ProjectCreateInput](srv, r, OpCreate, "Create a new project")
	}
	if r.Ops&OpReadOne != 0 {
		addTool[*ReadOneInput](srv, r, OpReadOne, "Fetch a single project by id")
	}
	if r.Ops&OpReadAll != 0 {
		addTool[*ReadAllInput](srv, r, OpReadAll, "List the projects the caller has access to")
	}
	if r.Ops&OpUpdate != 0 {
		addTool[*ProjectUpdateInput](srv, r, OpUpdate, "Update an existing project")
	}
	if r.Ops&OpDelete != 0 {
		addTool[*DeleteInput](srv, r, OpDelete, "Delete a project by id")
	}
}

// addTool registers one MCP tool on the given server. The In type
// parameter must be a pointer-to-struct that implements inputAdapter (and
// optionally readAllInput); the SDK reflects it at registration time to
// build the input schema.
//
// The handler:
//
//  1. Calls DispatchTyped with the already-unmarshalled wrapper. The SDK
//     has already validated the input against the schema by the time the
//     handler runs (see ToolHandlerFor in the SDK docs), so there is no
//     reason to re-marshal and re-unmarshal.
//  2. Maps any error from the dispatcher to an IsError tool result per the
//     SDK's convention that domain failures (permission denials, missing
//     records, validation errors) surface as tool results, not JSON-RPC
//     protocol errors. ToolHandlerFor would do this automatically if we
//     returned the error, but we also want to populate Content with the
//     text explicitly so clients see a sensible message.
//  3. On success, returns the dispatcher's result as the structured Output;
//     the SDK populates Content with the JSON marshalling automatically.
func addTool[In inputAdapter](srv *mcp.Server, r *Resource, op Op, description string) {
	name := r.Name + "_" + op.ToolSuffix()
	tool := &mcp.Tool{
		Name:        name,
		Description: description,
	}
	// Domain-layer failures (permission denials, missing rows, validation
	// errors) surface as IsError tool results per the SDK convention, not as
	// protocol-level errors. The handler intentionally returns a nil error
	// alongside an IsError result; the nolint:nilerr below silences the
	// linter, which can't tell that this is the correct contract for
	// ToolHandlerFor.
	handler := func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		result, err := DispatchTyped(ctx, name, in)
		if err != nil {
			res := &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}
			//nolint:nilerr // IsError tool result, not a JSON-RPC protocol error
			return res, nil, nil
		}
		// Serialise the result manually so Content carries a stable JSON
		// shape; the SDK would do the same automatically when Content is
		// nil, but doing it here keeps the contract explicit and lets us
		// return the same payload as both unstructured text (for clients
		// that ignore structuredContent) and structured output.
		body, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return nil, nil, fmt.Errorf("mcp: marshal %s result: %w", name, marshalErr)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
		}, result, nil
	}
	mcp.AddTool(srv, tool, handler)
}
