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
// and the (future) MCP per-tool scope filter both look up by these exact
// strings.
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

// Resource describes a CRUD-able model exposed over MCP. Mirrors the
// handler.WebHandler{EmptyStruct: ...} shape used in pkg/routes/routes.go.
//
// Inputs maps each enabled op to a pointer-to-zero of the wrapper struct
// the dispatcher should unmarshal tool arguments into. The wrapper carries
// json:/jsonschema: tags consumed by the SDK's AddTool for input-schema
// generation, and implements the inputAdapter seam below so the dispatcher
// can copy wrapper -> fresh model before invoking handler.Do*.
//
// The wrapper structs themselves live in inputs.go (introduced in Task 4).
// Task 3 only carries them through the registry.
type Resource struct {
	// Name matches the API-token scope group exactly (e.g. "projects",
	// "task_comments"). It is also the prefix of every tool name this
	// resource produces.
	Name string

	// Description is used as the prefix of each generated tool's
	// description text.
	Description string

	// EmptyStruct returns a fresh, zero-valued model instance for each
	// dispatched call. Mirrors handler.WebHandler.EmptyStruct.
	EmptyStruct func() handler.CObject

	// Ops is the bitmask of CRUD operations this resource supports.
	Ops Op

	// Inputs holds the per-op wrapper type. The dispatcher allocates a
	// fresh value with reflection (via reflect.TypeOf(v).Elem()), JSON-
	// unmarshals the call arguments into it, and then asks the wrapper to
	// copy itself onto a fresh model via the inputAdapter interface.
	Inputs map[Op]any
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

// Register adds a resource to the package-level registry. It validates the
// shape (non-empty name, EmptyStruct present, an Inputs entry for each op
// in the Ops bitmask) and populates the tool-name lookup table so the
// dispatcher never has to string-parse tool names like
// "task_comments_read_all".
func Register(r Resource) error {
	if r.Name == "" {
		return errors.New("mcp: resource Name must not be empty")
	}
	if r.EmptyStruct == nil {
		return fmt.Errorf("mcp: resource %q has no EmptyStruct", r.Name)
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := findResourceLocked(r.Name); exists {
		return fmt.Errorf("%w: %s", ErrDuplicateResource, r.Name)
	}

	// Make sure every enabled op has an input wrapper, otherwise the
	// dispatcher would crash later with a less useful error.
	for _, op := range AllOps() {
		if r.Ops&op == 0 {
			continue
		}
		if _, has := r.Inputs[op]; !has {
			return fmt.Errorf("mcp: resource %q is missing input for op %s", r.Name, op.ToolSuffix())
		}
	}

	stored := r
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

// lookupTool returns the (resource, op) pair the given tool name was
// registered for.
func lookupTool(toolName string) (toolRef, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	ref, ok := toolIndex[toolName]
	return ref, ok
}
