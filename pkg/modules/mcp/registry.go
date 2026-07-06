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
	"errors"
	"fmt"
	"reflect"
	"sync"

	"code.vikunja.io/api/pkg/web/handler"
)

// Op is a bitmask of the CRUD operations a resource exposes. Bitmask was
// chosen because resources rarely need anything beyond a simple
// allow/disallow per op and OR-ing flags reads cleanly at the registration
// site (e.g. OpCreate | OpReadOne | OpReadAll). No other corner of the
// codebase uses bitmasks; this is local to the MCP registry.
type Op uint8

const (
	OpCreate Op = 1 << iota
	OpReadOne
	OpReadAll
	OpUpdate
	OpDelete
)

// AllOps returns the ops in registration-and-iteration order. Keeping the
// list in one place ensures the registry, dispatcher, and any future
// tools/list filter walk the same five.
func AllOps() []Op {
	return []Op{OpCreate, OpReadOne, OpReadAll, OpUpdate, OpDelete}
}

// Permission returns the API-token permission string for this op. The
// strings must match the permission names that pkg/models/api_routes.go
// stores under apiTokenRoutes[group][...]; CanDoAPIRoute in the REST layer
// and the MCP per-tool scope filter both look up by these exact strings.
func (o Op) Permission() string {
	switch o {
	case OpCreate:
		return "create"
	case OpReadOne:
		return "read_one"
	case OpReadAll:
		return "read_all"
	case OpUpdate:
		return "update"
	case OpDelete:
		return "delete"
	}
	return ""
}

// ToolSuffix returns the snake_case suffix used to form a tool name. Tool
// names are <resource.Name>_<op-suffix>; the suffix is identical to the
// permission string today but kept separate so the two can evolve
// independently if MCP and the REST scope system diverge.
func (o Op) ToolSuffix() string {
	return o.Permission()
}

// Tier controls how a resource is exposed to MCP clients.
type Tier uint8

const (
	// TierTyped resources register one first-class tool per op — visible in
	// tools/list with a full input schema. For the assistant-core surface.
	TierTyped Tier = iota
	// TierCatalog resources are only reachable through the find_action /
	// do_action meta-tools, keeping tools/list small for the long tail.
	TierCatalog
)

// Resource describes a CRUD-able model exposed over MCP. Everything beyond
// this declaration — input schemas, argument application, dispatch — is
// derived at registration time from the model's struct tags (json / doc /
// readOnly / valid / minLength / param / query), the same contract the
// Huma-backed /api/v2 reflects. See schema.go for the derivation rules.
type Resource struct {
	// Name matches the API-token scope group exactly (e.g. "projects",
	// "tasks_comments"). It is also the prefix of every tool name this
	// resource produces.
	Name string

	// Description is used in each generated tool's description text.
	Description string

	// Model returns a fresh, zero-valued model instance for each dispatched
	// call. Mirrors handler.WebHandler.EmptyStruct.
	Model func() handler.CObject

	// Models overrides Model per op — e.g. tasks list through
	// models.TaskCollection while the other ops use models.Task.
	Models map[Op]func() handler.CObject

	// Ops is the bitmask of CRUD operations this resource supports.
	Ops Op

	// Tier selects typed tools vs. catalog-only exposure.
	Tier Tier

	// Gate, when set, is consulted at session-init time; a false return
	// hides the resource entirely (live config checks, e.g. task comments).
	Gate func() bool

	// Exclude hides fields from every op's schema, by JSON property name.
	Exclude []string

	// OptionalFields downgrades hidden param-derived fields the derivation
	// would mark required, by JSON property name (e.g. TaskCollection's
	// project_id — omitting it lists tasks across all projects). Writable
	// fields are unaffected.
	OptionalFields []string

	// RequiredCreate marks additional fields required on create when the
	// tags alone don't say so.
	RequiredCreate []string

	// IdentityFields overrides how read_one/update/delete address a record,
	// by JSON property name, for models whose row isn't addressed by its id
	// (team members go by team + username) or that need parent context the
	// derivation can't infer (views need project_id alongside id).
	IdentityFields []string

	specs map[Op]*opSpec
}

// modelFor returns the constructor for the given op, honouring the per-op
// override map.
func (r *Resource) modelFor(op Op) func() handler.CObject {
	if f, ok := r.Models[op]; ok {
		return f
	}
	return r.Model
}

// spec returns the cached tool contract for the given op.
func (r *Resource) spec(op Op) *opSpec {
	return r.specs[op]
}

// enabled reports whether the resource's config gate (if any) allows it.
func (r *Resource) enabled() bool {
	return r.Gate == nil || r.Gate()
}

// toolDescription renders the generated tool's description. Update spells
// out the partial-update contract because it is the one op whose semantics
// an agent cannot infer from the schema.
func (r *Resource) toolDescription(op Op) string {
	switch op {
	case OpCreate:
		return "Create a new record. Resource: " + r.Description
	case OpReadOne:
		return "Fetch a single record. Resource: " + r.Description
	case OpReadAll:
		return "List records the caller has access to. Resource: " + r.Description
	case OpUpdate:
		return "Update an existing record; only fields present in the arguments are changed. Resource: " + r.Description
	case OpDelete:
		return "Delete a record. Resource: " + r.Description
	}
	return r.Description
}

// toolRef points a tool name back at its resource + op. Built once at
// registration time so the dispatcher never has to parse tool names.
type toolRef struct {
	resource *Resource
	op       Op
}

var (
	registryMu sync.RWMutex
	resources  []*Resource
	toolIndex  = map[string]toolRef{}
)

// ErrDuplicateResource is returned when Register is called twice with the
// same Name.
var ErrDuplicateResource = errors.New("mcp: resource already registered")

// Register adds a resource to the package-level registry, builds the per-op
// input schemas from the model's struct tags, and populates the tool-name
// lookup table so the dispatcher never has to string-parse tool names like
// "tasks_comments_read_all".
func Register(r Resource) error {
	if r.Name == "" {
		return errors.New("mcp: resource Name must not be empty")
	}
	if r.Model == nil {
		return fmt.Errorf("mcp: resource %q has no Model", r.Name)
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := findResourceLocked(r.Name); exists {
		return fmt.Errorf("%w: %s", ErrDuplicateResource, r.Name)
	}

	stored := r
	stored.specs = make(map[Op]*opSpec)
	for _, op := range AllOps() {
		if stored.Ops&op == 0 {
			continue
		}
		model := stored.modelFor(op)()
		mt := reflect.TypeOf(model)
		if mt == nil || mt.Kind() != reflect.Pointer || mt.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("mcp: resource %q model for op %s must be a pointer to struct, got %T", r.Name, op.ToolSuffix(), model)
		}
		spec, err := buildOpSpec(mt.Elem(), op, &stored)
		if err != nil {
			return err
		}
		stored.specs[op] = spec
	}

	resources = append(resources, &stored)

	for _, op := range AllOps() {
		if stored.Ops&op == 0 {
			continue
		}
		toolName := stored.Name + "_" + op.ToolSuffix()
		toolIndex[toolName] = toolRef{resource: &stored, op: op}
	}

	return nil
}

// lookupResource returns the registered resource with the given name.
// Intended for tests and internal callers; external code should resolve
// via tool name.
func lookupResource(name string) (*Resource, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return findResourceLocked(name)
}

func findResourceLocked(name string) (*Resource, bool) {
	for _, r := range resources {
		if r.Name == name {
			return r, true
		}
	}
	return nil, false
}

// snapshotResources returns the registered resources in registration order.
func snapshotResources() []*Resource {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]*Resource, len(resources))
	copy(out, resources)
	return out
}

// lookupTool returns the (resource, op) pair the given tool name was
// registered for.
func lookupTool(toolName string) (toolRef, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	ref, ok := toolIndex[toolName]
	return ref, ok
}
