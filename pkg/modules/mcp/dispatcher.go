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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"
)

// ErrToolNotFound maps to an IsError tool result, not a JSON-RPC protocol error.
var ErrToolNotFound = errors.New("mcp: tool not found")

// ErrNoUserInContext means Dispatch was called outside the HTTP pipeline, which always sets a user.
var ErrNoUserInContext = errors.New("mcp: no user in context")

// crudFuncs exists so the dispatcher's routing tests can run without a database.
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

var crud = defaultCRUD

// Dispatch is the single entry point for every tools/call, typed or via
// do_action. Every error is returned as a plain error, not an SDK type, so
// this package stays SDK-agnostic; the caller renders them as tool results.
func Dispatch(ctx context.Context, toolName string, rawArgs json.RawMessage) (any, error) {
	ref, ok := lookupTool(toolName)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	// tools/list and find_action hide gated resources; do_action would reach them by name.
	if !ref.resource.enabled() {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	// Fail closed: do_action must not reach a tool the token was never registered for.
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

	// Updates are read-modify-write: the models write a fixed column list, so a
	// partial payload applied to a zero model would blank every omitted column.
	// Resources without a read_one op leave web.CRUDable nil, and their Update
	// writes a single column anyway.
	// The read and the write run in separate transactions, so a concurrent write
	// in between is lost; v2's AutoPatch If-Match handling is the proper fix.
	if ref.op == OpUpdate && ref.resource.Ops&OpReadOne != 0 {
		readSpec := ref.resource.spec(OpReadOne)
		if err := applyArgs(model, readSpec, identityArgs(readSpec, args)); err != nil {
			return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
		}
		if _, err := crud.doReadOne(ctx, model, u); err != nil {
			return nil, err
		}
	}

	if err := applyArgs(model, spec, args); err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
	}

	// The REST layer runs this via echo's CustomValidator before the handler;
	// without it MCP writes bypass every `valid:` tag rule on the model.
	if ref.op == OpCreate || ref.op == OpUpdate {
		if err := models.ValidateStruct(model); err != nil {
			return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolName, err)
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
		result, _, totalItems, err := crud.doReadAll(ctx, model, u, search, page, perPage)
		if err != nil {
			return nil, err
		}
		return newReadAllResult(result, totalItems, page, perPage), nil

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

// readAllResult mirrors apiv2.Paginated field for field; mcp must not import apiv2.
type readAllResult struct {
	Items      any   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int64 `json:"total_pages"`
}

func newReadAllResult(items any, total int64, page, perPage int) *readAllResult {
	if v := reflect.ValueOf(items); !v.IsValid() || (v.Kind() == reflect.Slice && v.IsNil()) {
		items = []any{}
	}
	var totalPages int64
	if perPage > 0 {
		totalPages = (total + int64(perPage) - 1) / int64(perPage)
	}
	return &readAllResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}
