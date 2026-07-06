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

// Presence-based argument application: arguments are decoded as a raw key →
// value map and only the keys the caller actually sent are written onto the
// model. This gives partial-update semantics without pointer-typed wrapper
// fields — an explicit `"done": false` clears the flag, an omitted key
// leaves the row untouched.

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"code.vikunja.io/api/pkg/web/handler"
)

// validateAndDecodeArgs checks the raw arguments against the op's schema
// (types, required properties, unknown-key rejection) and returns them as a
// key → raw-value map for presence-based application.
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

// applyArgs unmarshals each supplied argument into its model field. The
// schema has already validated names and types; errors here mean a value
// that passed JSON Schema but not Go unmarshalling (e.g. a malformed
// RFC 3339 timestamp).
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

// popReadAllParams extracts (and removes) the reserved search/page/per_page
// arguments so applyArgs only sees model-bound keys. They map onto
// handler.DoReadAll's positional parameters.
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
	err = pop(argPerPage, &perPage)
	return
}
