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
	"errors"
	"fmt"
	"reflect"

	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"
)

// ErrToolNotFound is returned when Dispatch is called for a tool name that
// has not been registered. Callers should map this to an MCP tool result
// with IsError=true (per the SDK convention for missing tools), not to a
// JSON-RPC protocol error.
var ErrToolNotFound = errors.New("mcp: tool not found")

// ErrNoUserInContext is returned when Dispatch is invoked without a user
// in ctx. Task 2's entry handler always sets one, so hitting this means
// either a programming bug or someone calling Dispatch outside the HTTP
// pipeline.
var ErrNoUserInContext = errors.New("mcp: no user in context")

// inputAdapter is the Task 3/Task 4 seam. Each per-op input wrapper struct
// (defined in inputs.go, added by Task 4) implements ApplyTo, which copies
// the wrapper's fields onto a fresh handler.CObject. The dispatcher
// allocates a wrapper from Resource.Inputs[op] via reflection,
// json.Unmarshals tool arguments into it, then calls ApplyTo on the model
// returned by Resource.EmptyStruct().
//
// Defining the interface here (rather than in inputs.go) keeps the
// dispatcher buildable in Task 3 before any wrappers exist; the
// dispatcher tests provide their own ApplyTo implementation to exercise
// the code path.
type inputAdapter interface {
	ApplyTo(dst handler.CObject) error
}

// readAllInput is the optional interface a wrapper for OpReadAll may
// implement to expose pagination fields to the dispatcher. Wrappers that
// don't implement it get search="", page=0, perPage=0 (the same defaults
// the REST layer applies when callers omit the query parameters).
type readAllInput interface {
	ReadAllParams() (search string, page int, perPage int)
}

// crudFuncs are the framework-agnostic Do* entry points the dispatcher
// invokes. The package-level defaults point at handler.Do*; tests swap
// them out so they can run without a database connection (handler.Do*
// opens an xorm session, which is fine in integration tests but not in
// the dispatcher unit tests that exercise routing logic only).
type crudFuncs struct {
	doCreate  func(context.Context, handler.CObject, web.Auth) error
	doReadOne func(context.Context, handler.CObject, web.Auth) (int, error)
	doReadAll func(context.Context, handler.CObject, web.Auth, string, int, int) (any, int, int64, error)
	doUpdate  func(context.Context, handler.CObject, web.Auth) error
	doDelete  func(context.Context, handler.CObject, web.Auth) error
}

var defaultCRUD = crudFuncs{
	doCreate:  handler.DoCreate,
	doReadOne: handler.DoReadOne,
	doReadAll: handler.DoReadAll,
	doUpdate:  handler.DoUpdate,
	doDelete:  handler.DoDelete,
}

// crud is the live set of Do* functions Dispatch uses. Tests swap it out
// via withCRUD and restore it on teardown.
var crud = defaultCRUD

// Dispatch is the single entry point for every tools/call. It returns
// either the result the SDK should serialize (a model on read_one/update,
// the slice from ReadAll on read_all, or the model on create) or an error.
//
// Errors fall into two categories:
//   - ErrToolNotFound / ErrNoUserInContext / JSON-unmarshal errors are
//     dispatcher-level failures the caller should translate into an
//     IsError=true tool result. We return them as errors here (rather than
//     constructing a *mcp.CallToolResult) so the dispatcher stays
//     SDK-agnostic; the thin AddTool handler in Task 5 does the wrapping.
//   - Errors returned by handler.Do* (model-layer permission denials,
//     validation failures, etc.) are propagated as-is. The tool handler
//     in Task 5 wraps them with SetError per the SDK's convention that
//     domain failures be reported as tool results, not protocol errors.
func Dispatch(ctx context.Context, toolName string, rawArgs json.RawMessage) (any, error) {
	ref, ok := lookupTool(toolName)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	u := UserFromContext(ctx)
	if u == nil {
		return nil, ErrNoUserInContext
	}

	// Allocate a fresh wrapper for this call so concurrent dispatches
	// don't share state through the prototype stored in r.Inputs.
	wrapperProto, ok := ref.resource.Inputs[ref.op]
	if !ok {
		return nil, fmt.Errorf("mcp: resource %q has no input wrapper for op %s", ref.resource.Name, ref.op.ToolSuffix())
	}
	wrapper, err := allocateWrapper(wrapperProto)
	if err != nil {
		return nil, err
	}

	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, wrapper); err != nil {
			return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
		}
	}

	model := ref.resource.EmptyStruct()
	if adapter, ok := wrapper.(inputAdapter); ok {
		if err := adapter.ApplyTo(model); err != nil {
			return nil, fmt.Errorf("mcp: copy input for %s: %w", toolName, err)
		}
	}

	switch ref.op {
	case OpCreate:
		if err := crud.doCreate(ctx, model, u); err != nil {
			return nil, err
		}
		return model, nil

	case OpReadOne:
		if _, err := crud.doReadOne(ctx, model, u); err != nil {
			return nil, err
		}
		return model, nil

	case OpReadAll:
		search, page, perPage := "", 0, 0
		if ra, ok := wrapper.(readAllInput); ok {
			search, page, perPage = ra.ReadAllParams()
		}
		result, _, _, err := crud.doReadAll(ctx, model, u, search, page, perPage)
		if err != nil {
			return nil, err
		}
		return result, nil

	case OpUpdate:
		if err := crud.doUpdate(ctx, model, u); err != nil {
			return nil, err
		}
		return model, nil

	case OpDelete:
		if err := crud.doDelete(ctx, model, u); err != nil {
			return nil, err
		}
		return model, nil
	}

	return nil, fmt.Errorf("mcp: unsupported op %d for tool %s", ref.op, toolName)
}

// allocateWrapper returns a fresh pointer of the same concrete type as the
// prototype stored in Resource.Inputs. Resource.Inputs is conventionally a
// pointer-to-zero (e.g. &ProjectCreateInput{}); allocateWrapper takes its
// reflect.Type, allocates a fresh value, and hands back a pointer suitable
// for json.Unmarshal.
func allocateWrapper(proto any) (any, error) {
	if proto == nil {
		return nil, errors.New("mcp: nil input prototype")
	}
	t := reflect.TypeOf(proto)
	if t.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("mcp: input prototype must be a pointer, got %s", t.Kind())
	}
	return reflect.New(t.Elem()).Interface(), nil
}
