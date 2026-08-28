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

// Arguments are decoded as a raw key → value map so only the keys the caller
// actually sent get written onto the model. That gives partial updates without
// pointer-typed fields: `"done": false` clears, an omitted key leaves the row.

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web/handler"
)

func validateAndDecodeArgs(spec *opSpec, raw json.RawMessage) (map[string]json.RawMessage, error) {
	instance := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &instance); err != nil {
			return nil, errors.New("arguments must be a JSON object")
		}
	}
	if err := spec.resolved.Validate(instance); err != nil {
		return nil, err
	}
	args := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, errors.New("arguments must be a JSON object")
		}
	}
	return args, nil
}

// The schema already validated names and types, so errors here mean a value that
// passed JSON Schema but not Go unmarshalling (e.g. a malformed RFC 3339 timestamp).
func applyArgs(model handler.CObject, spec *opSpec, args map[string]json.RawMessage) error {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("mcp: model must be a non-nil pointer, got %s", rv.Kind())
	}
	rv = rv.Elem()
	for name, rawVal := range args {
		idx, ok := spec.fields[name]
		if !ok {
			return fmt.Errorf("unknown argument %q", name)
		}
		field := rv.Field(idx)
		if !field.CanAddr() {
			return fmt.Errorf("mcp: field for argument %q is not addressable", name)
		}
		if err := json.Unmarshal(rawVal, field.Addr().Interface()); err != nil {
			return fmt.Errorf("invalid value for %q: %w", name, err)
		}
	}
	return nil
}

// Normalising is not optional: page < 1 makes the models skip the LIMIT clause
// entirely, so an omitted per_page would dump every row the caller can see.
func popReadAllParams(args map[string]json.RawMessage) (search string, page, perPage int, err error) {
	pop := func(name string, dst any) error {
		raw, ok := args[name]
		if !ok {
			return nil
		}
		delete(args, name)
		if err := json.Unmarshal(raw, dst); err != nil {
			return fmt.Errorf("invalid value for %q: %w", name, err)
		}
		return nil
	}
	if err = pop(argSearch, &search); err != nil {
		return
	}
	if err = pop(argPage, &page); err != nil {
		return
	}
	if err = pop(argPerPage, &perPage); err != nil {
		return
	}

	if page < 0 {
		return "", 0, 0, fmt.Errorf("invalid value for %q: must not be negative", argPage)
	}
	if page == 0 {
		page = 1
	}
	if perPage < 0 {
		return "", 0, 0, fmt.Errorf("invalid value for %q: must not be negative", argPerPage)
	}
	maxPerPage := config.ServiceMaxItemsPerPage.GetInt()
	if perPage == 0 || perPage > maxPerPage {
		perPage = maxPerPage
	}
	return search, page, perPage, nil
}
