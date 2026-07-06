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

	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"
)

// ErrToolNotFound is returned when Dispatch is called for a tool name that
// has not been registered. Callers should map this to an MCP tool result
// with IsError=true (per the SDK convention for missing tools), not to a
// JSON-RPC protocol error.
var ErrToolNotFound = errors.New("mcp: tool not found")

// ErrNoUserInContext is returned when Dispatch is invoked without a user
// in ctx. The entry handler always sets one, so hitting this means either a
// programming bug or someone calling Dispatch outside the HTTP pipeline.
var ErrNoUserInContext = errors.New("mcp: no user in context")

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
// and restore it on teardown.
var crud = defaultCRUD

// Dispatch is the single entry point for every tools/call — the typed
// per-resource tools and the do_action catalog both funnel through here
// with raw JSON arguments. It validates the arguments against the op's
// tag-derived schema, applies the supplied keys onto a fresh model
// (presence-based, see apply.go), and invokes the matching handler.Do*.
//
// Errors fall into three categories:
//   - ErrToolNotFound / ErrNoUserInContext / ErrScopeDenied and argument
//     validation failures are dispatcher-level; callers translate them into
//     IsError=true tool results. They're returned as errors (rather than
//     *mcp.CallToolResult) so the dispatcher stays SDK-agnostic.
//   - Errors returned by handler.Do* (model-layer permission denials,
//     validation failures, etc.) are propagated as-is; the tool handler
//     wraps them per the SDK's convention that domain failures be reported
//     as tool results, not protocol errors.
func Dispatch(ctx context.Context, toolName string, rawArgs json.RawMessage) (any, error) {
	ref, ok := lookupTool(toolName)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	// Scope check first — never touch model state for a tool the caller
	// isn't authorized to invoke. This guards against the (rare) case where
	// the per-session tool registration in newServer registered a tool the
	// current request's token doesn't have a scope for: the SDK caches the
	// *Server across requests, but the API token is per-HTTP-request.
	if !tokenAuthorizes(TokenFromContext(ctx), ref.resource.Name, ref.op) {
		return nil, fmt.Errorf("%w: %s", ErrScopeDenied, toolName)
	}

	u := UserFromContext(ctx)
	if u == nil {
		return nil, ErrNoUserInContext
	}

	spec := ref.resource.spec(ref.op)
	args, err := validateAndDecodeArgs(spec, rawArgs)
	if err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
	}

	var search string
	var page, perPage int
	if ref.op == OpReadAll {
		if search, page, perPage, err = popReadAllParams(args); err != nil {
			return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
		}
	}

	model := ref.resource.modelFor(ref.op)()
	if err := applyArgs(model, spec, args); err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
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
