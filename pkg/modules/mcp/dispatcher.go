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
	"strings"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
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

	// tools/list and find_action already hide gated resources; do_action
	// would otherwise reach them by name.
	if !ref.resource.enabled() {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	// Fail closed: do_action must not reach a tool the token was never
	// registered for.
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

	// The REST layer runs this via echo's CustomValidator before the handler;
	// without it MCP writes bypass every `valid:` tag rule on the model.
	if ref.op == OpCreate || ref.op == OpUpdate {
		if err := models.ValidateStructFields(model, suppliedFieldNames(model, spec, args)); err != nil {
			return nil, validationFailure(toolName, err)
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
		result, resultCount, totalItems, err := crud.doReadAll(ctx, model, u, search, page, perPage)
		if err != nil {
			return nil, err
		}
		return newReadAllResult(result, resultCount, totalItems, page, perPage), nil

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

// validationFailure renders a `valid:` tag failure for a tool result, which is
// plain text — ValidationHTTPError keeps the offending field names out of its
// message.
func validationFailure(toolName string, err error) error {
	var invalid models.ValidationHTTPError
	if errors.As(err, &invalid) && len(invalid.InvalidFields) > 0 {
		return fmt.Errorf("mcp: invalid arguments for %s: %s", toolName, strings.Join(invalid.InvalidFields, "; "))
	}
	return fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
}

// suppliedFieldNames returns the names govalidator may report for the
// arguments the caller actually sent: the JSON property name plus the Go
// field name, which govalidator falls back to for `json:"-"` fields.
func suppliedFieldNames(model handler.CObject, spec *opSpec, args map[string]json.RawMessage) map[string]bool {
	modelType := reflect.TypeOf(model).Elem()
	names := make(map[string]bool, len(args)*2)
	for name := range args {
		names[name] = true
		if idx, ok := spec.fields[name]; ok {
			names[modelType.Field(idx).Name] = true
		}
	}
	return names
}

// readAllResult is the read_all envelope. A bare array left clients no way to
// tell a truncated page from the last one, and no way to page on from it.
type readAllResult struct {
	Items       any   `json:"items"`
	ResultCount int   `json:"result_count"`
	TotalItems  int64 `json:"total_items"`
	Page        int   `json:"page"`
	PerPage     int   `json:"per_page"`
}

func newReadAllResult(items any, resultCount int, totalItems int64, page, perPage int) *readAllResult {
	// read_all hands out user rows directly, skipping the per-parent
	// serialisation the REST layer relies on to hide addresses.
	if users, ok := items.([]*user.User); ok {
		for _, u := range users {
			if u != nil {
				u.Email = ""
			}
		}
	}
	if v := reflect.ValueOf(items); !v.IsValid() || (v.Kind() == reflect.Slice && v.IsNil()) {
		items = []any{}
	}
	return &readAllResult{
		Items:       items,
		ResultCount: resultCount,
		TotalItems:  totalItems,
		Page:        page,
		PerPage:     perPage,
	}
}
