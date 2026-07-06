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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Approach for scope-filtered tools/list (Task 6): the SDK calls the
// getServer factory in NewStreamableHTTPHandler exactly once per session
// (at the initialize request, when no Mcp-Session-Id matches an existing
// session) and caches the returned *mcp.Server for the lifetime of that
// session. There is no filter callback in mcp.ServerOptions, so we build a
// per-session *mcp.Server that only registers the tools the requesting
// token's APIPermissions allows. tools/list then naturally returns the
// allowed subset. The dispatcher additionally re-checks scopes on every
// tools/call as a defence-in-depth measure (the same session could in
// principle be reused across requests carrying different tokens).

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
//
// task_comments is always registered (its model is always available); the
// install-time check in installTaskCommentsToolsForToken gates whether the
// tools actually appear in tools/list per the live ServiceEnableTaskComments
// setting, so toggling the config doesn't require a server restart.
func RegisterResources() {
	registerResourcesOnce.Do(func() {
		registrars := []struct {
			name string
			fn   func() error
		}{
			{"projects", registerProjects},
			{"tasks", registerTasks},
			{"labels", registerLabels},
			{"teams", registerTeams},
			{"tasks_comments", registerTaskComments},
			{"tasks_assignees", registerTaskAssignees},
		}
		for _, r := range registrars {
			if err := r.fn(); err != nil {
				panic(fmt.Errorf("mcp: failed to register %s resource: %w", r.name, err))
			}
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

// registerTasks omits OpReadAll because models.Task.ReadAll is a no-op
// stub (the REST layer routes /tasks to TaskCollection, which is out of
// scope for v1 per the plan). Tools/list will not include tasks_read_all.
func registerTasks() error {
	return Register(Resource{
		Name:        "tasks",
		Description: "Vikunja tasks (work items inside a project)",
		EmptyStruct: func() handler.CObject { return &models.Task{} },
		Ops:         OpCreate | OpReadOne | OpUpdate | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &TaskCreateInput{},
			OpReadOne: &ReadOneInput{},
			OpUpdate:  &TaskUpdateInput{},
			OpDelete:  &DeleteInput{},
		},
	})
}

func registerLabels() error {
	return Register(Resource{
		Name:        "labels",
		Description: "Vikunja labels (reusable tags attachable to tasks)",
		EmptyStruct: func() handler.CObject { return &models.Label{} },
		Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &LabelCreateInput{},
			OpReadOne: &ReadOneInput{},
			OpReadAll: &ReadAllInput{},
			OpUpdate:  &LabelUpdateInput{},
			OpDelete:  &DeleteInput{},
		},
	})
}

func registerTeams() error {
	return Register(Resource{
		Name:        "teams",
		Description: "Vikunja teams (groups of users that can share projects)",
		EmptyStruct: func() handler.CObject { return &models.Team{} },
		Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &TeamCreateInput{},
			OpReadOne: &ReadOneInput{},
			OpReadAll: &ReadAllInput{},
			OpUpdate:  &TeamUpdateInput{},
			OpDelete:  &DeleteInput{},
		},
	})
}

// registerTaskComments uses per-op wrappers (rather than the shared
// ReadOne/Delete/ReadAll wrappers) because every comment operation needs the
// parent task_id supplied as a JSON arg — the REST layer binds it from the
// URL, but MCP has no URL to bind from.
func registerTaskComments() error {
	return Register(Resource{
		Name:        "tasks_comments",
		Description: "Comments attached to a Vikunja task",
		EmptyStruct: func() handler.CObject { return &models.TaskComment{} },
		Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &TaskCommentCreateInput{},
			OpReadOne: &TaskCommentReadOneInput{},
			OpReadAll: &TaskCommentReadAllInput{},
			OpUpdate:  &TaskCommentUpdateInput{},
			OpDelete:  &TaskCommentDeleteInput{},
		},
	})
}

// registerTaskAssignees registers only the three ops the REST layer
// supports for the assignee resource (PUT/GET-all/DELETE) — there is no
// per-assignee read_one or update endpoint in REST, so MCP doesn't expose
// them either.
func registerTaskAssignees() error {
	return Register(Resource{
		Name:        "tasks_assignees",
		Description: "Users assigned to a Vikunja task",
		EmptyStruct: func() handler.CObject { return &models.TaskAssginee{} },
		Ops:         OpCreate | OpReadAll | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &TaskAssigneeCreateInput{},
			OpReadAll: &TaskAssigneeReadAllInput{},
			OpDelete:  &TaskAssigneeDeleteInput{},
		},
	})
}

// installToolsForToken walks every per-resource installer below and binds
// the resource's (resource, op) tools onto the given server, gated by the
// token's APIPermissions. Per-op wrapper types are known at compile time, so
// each resource has its own installer; the registry stays data-driven
// everywhere else.
//
// Called from newServer (mcp.go) at session-init time. A nil token (which
// should never happen in production because the entry handler rejects
// unauthenticated requests) yields a server with no tools — defensive, the
// dispatcher would also reject the call.
func installToolsForToken(srv *mcp.Server, token *models.APIToken) {
	installProjectsToolsForToken(srv, token)
	installTasksToolsForToken(srv, token)
	installLabelsToolsForToken(srv, token)
	installTeamsToolsForToken(srv, token)
	installTaskCommentsToolsForToken(srv, token)
	installTaskAssigneesToolsForToken(srv, token)
}

// resourceOrPanic looks up a registered resource by name; missing resources
// indicate that RegisterResources hasn't run, which is a programmer error.
func resourceOrPanic(name string) *Resource {
	r, ok := lookupResource(name)
	if !ok {
		panic("mcp: " + name + " resource not registered")
	}
	return r
}

func installProjectsToolsForToken(srv *mcp.Server, token *models.APIToken) {
	r := resourceOrPanic("projects")

	if r.Ops&OpCreate != 0 && tokenAuthorizes(token, r.Name, OpCreate) {
		addTool[*ProjectCreateInput](srv, r, OpCreate, "Create a new project")
	}
	if r.Ops&OpReadOne != 0 && tokenAuthorizes(token, r.Name, OpReadOne) {
		addTool[*ReadOneInput](srv, r, OpReadOne, "Fetch a single project by id")
	}
	if r.Ops&OpReadAll != 0 && tokenAuthorizes(token, r.Name, OpReadAll) {
		addTool[*ReadAllInput](srv, r, OpReadAll, "List the projects the caller has access to")
	}
	if r.Ops&OpUpdate != 0 && tokenAuthorizes(token, r.Name, OpUpdate) {
		addTool[*ProjectUpdateInput](srv, r, OpUpdate, "Update an existing project")
	}
	if r.Ops&OpDelete != 0 && tokenAuthorizes(token, r.Name, OpDelete) {
		addTool[*DeleteInput](srv, r, OpDelete, "Delete a project by id")
	}
}

func installTasksToolsForToken(srv *mcp.Server, token *models.APIToken) {
	r := resourceOrPanic("tasks")

	if r.Ops&OpCreate != 0 && tokenAuthorizes(token, r.Name, OpCreate) {
		addTool[*TaskCreateInput](srv, r, OpCreate, "Create a new task inside a project")
	}
	if r.Ops&OpReadOne != 0 && tokenAuthorizes(token, r.Name, OpReadOne) {
		addTool[*ReadOneInput](srv, r, OpReadOne, "Fetch a single task by id")
	}
	if r.Ops&OpUpdate != 0 && tokenAuthorizes(token, r.Name, OpUpdate) {
		addTool[*TaskUpdateInput](srv, r, OpUpdate, "Update an existing task")
	}
	if r.Ops&OpDelete != 0 && tokenAuthorizes(token, r.Name, OpDelete) {
		addTool[*DeleteInput](srv, r, OpDelete, "Delete a task by id")
	}
	// OpReadAll is intentionally not exposed: models.Task.ReadAll is a stub.
	// Listing tasks is handled by TaskCollection at the REST layer, which is
	// out of scope for v1.
}

func installLabelsToolsForToken(srv *mcp.Server, token *models.APIToken) {
	r := resourceOrPanic("labels")

	if r.Ops&OpCreate != 0 && tokenAuthorizes(token, r.Name, OpCreate) {
		addTool[*LabelCreateInput](srv, r, OpCreate, "Create a new label")
	}
	if r.Ops&OpReadOne != 0 && tokenAuthorizes(token, r.Name, OpReadOne) {
		addTool[*ReadOneInput](srv, r, OpReadOne, "Fetch a single label by id")
	}
	if r.Ops&OpReadAll != 0 && tokenAuthorizes(token, r.Name, OpReadAll) {
		addTool[*ReadAllInput](srv, r, OpReadAll, "List labels the caller has access to")
	}
	if r.Ops&OpUpdate != 0 && tokenAuthorizes(token, r.Name, OpUpdate) {
		addTool[*LabelUpdateInput](srv, r, OpUpdate, "Update an existing label")
	}
	if r.Ops&OpDelete != 0 && tokenAuthorizes(token, r.Name, OpDelete) {
		addTool[*DeleteInput](srv, r, OpDelete, "Delete a label by id")
	}
}

func installTeamsToolsForToken(srv *mcp.Server, token *models.APIToken) {
	r := resourceOrPanic("teams")

	if r.Ops&OpCreate != 0 && tokenAuthorizes(token, r.Name, OpCreate) {
		addTool[*TeamCreateInput](srv, r, OpCreate, "Create a new team")
	}
	if r.Ops&OpReadOne != 0 && tokenAuthorizes(token, r.Name, OpReadOne) {
		addTool[*ReadOneInput](srv, r, OpReadOne, "Fetch a single team by id")
	}
	if r.Ops&OpReadAll != 0 && tokenAuthorizes(token, r.Name, OpReadAll) {
		addTool[*ReadAllInput](srv, r, OpReadAll, "List teams the caller belongs to")
	}
	if r.Ops&OpUpdate != 0 && tokenAuthorizes(token, r.Name, OpUpdate) {
		addTool[*TeamUpdateInput](srv, r, OpUpdate, "Update an existing team")
	}
	if r.Ops&OpDelete != 0 && tokenAuthorizes(token, r.Name, OpDelete) {
		addTool[*DeleteInput](srv, r, OpDelete, "Delete a team by id")
	}
}

// installTaskCommentsToolsForToken is gated on the live
// config.ServiceEnableTaskComments setting. When task comments are disabled
// at the service level, the REST routes aren't registered either; mirroring
// that gate here keeps the MCP surface consistent.
func installTaskCommentsToolsForToken(srv *mcp.Server, token *models.APIToken) {
	if !config.ServiceEnableTaskComments.GetBool() {
		return
	}
	r := resourceOrPanic("tasks_comments")

	if r.Ops&OpCreate != 0 && tokenAuthorizes(token, r.Name, OpCreate) {
		addTool[*TaskCommentCreateInput](srv, r, OpCreate, "Create a comment on a task")
	}
	if r.Ops&OpReadOne != 0 && tokenAuthorizes(token, r.Name, OpReadOne) {
		addTool[*TaskCommentReadOneInput](srv, r, OpReadOne, "Fetch a single task comment")
	}
	if r.Ops&OpReadAll != 0 && tokenAuthorizes(token, r.Name, OpReadAll) {
		addTool[*TaskCommentReadAllInput](srv, r, OpReadAll, "List all comments on a task")
	}
	if r.Ops&OpUpdate != 0 && tokenAuthorizes(token, r.Name, OpUpdate) {
		addTool[*TaskCommentUpdateInput](srv, r, OpUpdate, "Update an existing task comment")
	}
	if r.Ops&OpDelete != 0 && tokenAuthorizes(token, r.Name, OpDelete) {
		addTool[*TaskCommentDeleteInput](srv, r, OpDelete, "Delete a task comment")
	}
}

func installTaskAssigneesToolsForToken(srv *mcp.Server, token *models.APIToken) {
	r := resourceOrPanic("tasks_assignees")

	if r.Ops&OpCreate != 0 && tokenAuthorizes(token, r.Name, OpCreate) {
		addTool[*TaskAssigneeCreateInput](srv, r, OpCreate, "Assign a user to a task")
	}
	if r.Ops&OpReadAll != 0 && tokenAuthorizes(token, r.Name, OpReadAll) {
		addTool[*TaskAssigneeReadAllInput](srv, r, OpReadAll, "List all users assigned to a task")
	}
	if r.Ops&OpDelete != 0 && tokenAuthorizes(token, r.Name, OpDelete) {
		addTool[*TaskAssigneeDeleteInput](srv, r, OpDelete, "Unassign a user from a task")
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
