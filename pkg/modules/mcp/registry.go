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

// Op is a bitmask so registration sites can OR ops together.
type Op uint8

const (
	OpCreate Op = 1 << iota
	OpReadOne
	OpReadAll
	OpUpdate
	OpDelete
)

// AllOps returns the ops in registration-and-iteration order.
func AllOps() []Op {
	return []Op{OpCreate, OpReadOne, OpReadAll, OpUpdate, OpDelete}
}

// Permission must return the exact permission names apiTokenRoutes uses in pkg/models/api_routes.go.
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

// Tier controls how a resource is exposed to MCP clients.
type Tier uint8

const (
	// TierTyped registers one first-class tool per op, visible in tools/list.
	TierTyped Tier = iota
	// TierCatalog is reachable only via find_action / do_action, keeping tools/list small.
	TierCatalog
)

// Resource describes a CRUD-able model exposed over MCP. Schemas, argument
// application and dispatch are all derived from the model's struct tags at
// registration time; see schema.go for the derivation rules.
type Resource struct {
	// Name matches the API-token scope group exactly and prefixes every tool name.
	Name string

	Description string

	// Model mirrors handler.WebHandler.EmptyStruct: a fresh instance per call.
	Model func() handler.CObject

	// Models overrides Model per op, e.g. tasks list through models.TaskCollection.
	Models map[Op]func() handler.CObject

	Ops  Op
	Tier Tier

	// Gate is evaluated at session-init time so live config changes take effect.
	Gate func() bool

	// Exclude hides fields from every op's schema, by JSON property name.
	Exclude []string

	// OptionalFields downgrades param-derived fields the derivation would mark required (e.g. TaskCollection's project_id).
	OptionalFields []string

	// IdentityFields overrides which properties read_one/update/delete address a record by (team members go by team + username, views need project_id alongside id).
	IdentityFields []string

	specs map[Op]*opSpec
}

func (r *Resource) modelFor(op Op) func() handler.CObject {
	if f, ok := r.Models[op]; ok {
		return f
	}
	return r.Model
}

func (r *Resource) spec(op Op) *opSpec {
	return r.specs[op]
}

func (r *Resource) enabled() bool {
	return r.Gate == nil || r.Gate()
}

// Update spells out the partial-update contract because an agent cannot infer it from the schema.
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

// toolRef is built at registration time so the dispatcher never parses tool names.
type toolRef struct {
	resource *Resource
	op       Op
}

var (
	registryMu sync.RWMutex
	resources  []*Resource
	toolIndex  = map[string]toolRef{}
)

// ErrDuplicateResource is returned when Register is called twice with the same Name.
var ErrDuplicateResource = errors.New("mcp: resource already registered")

// Register derives the per-op input schemas from the model's struct tags and indexes them by tool name.
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
			return fmt.Errorf("mcp: resource %q model for op %s must be a pointer to struct, got %T", r.Name, op.Permission(), model)
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
		toolName := stored.Name + "_" + op.Permission()
		toolIndex[toolName] = toolRef{resource: &stored, op: op}
	}

	return nil
}

// lookupResource is for internal callers only; everything else resolves via tool name.
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

func snapshotResources() []*Resource {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]*Resource, len(resources))
	copy(out, resources)
	return out
}

func lookupTool(toolName string) (toolRef, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	ref, ok := toolIndex[toolName]
	return ref, ok
}
