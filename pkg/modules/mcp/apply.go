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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

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

// decodeToolArgs validates before unmarshalling, so an unknown or mistyped key is
// named back to the caller instead of being silently dropped.
func decodeToolArgs(spec *opSpec, raw json.RawMessage, dst any) error {
	if _, err := validateAndDecodeArgs(spec, raw); err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
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
		if isIntegerKind(derefType(field.Type()).Kind()) {
			var err error
			if rawVal, err = integralNumber(rawVal); err != nil {
				return fmt.Errorf("invalid value for %q: %w", name, err)
			}
		}
		if err := json.Unmarshal(rawVal, field.Addr().Interface()); err != nil {
			return fmt.Errorf("invalid value for %q: %w", name, err)
		}
	}
	return nil
}

// The read_one schema's required set is exactly the fields that address a single row.
func identityArgs(spec *opSpec, args map[string]json.RawMessage) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(spec.schema.Required))
	for _, name := range spec.schema.Required {
		if raw, ok := args[name]; ok {
			out[name] = raw
		}
	}
	return out
}

func isIntegerKind(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Uint64
}

// JSON Schema's "integer" accepts 1.0 and 1e3, which json.Unmarshal into an
// int64 then rejects with an internals-flavoured message.
func integralNumber(raw json.RawMessage) (json.RawMessage, error) {
	num := json.Number(bytes.TrimSpace(raw))
	if _, err := num.Int64(); err == nil {
		return raw, nil
	}
	// Int64 rejected a plain integer literal, so it is out of range: report that
	// through the real unmarshal instead of rounding it through a float64.
	if !strings.ContainsAny(string(num), ".eE") {
		return raw, nil
	}
	f, err := num.Float64()
	if err != nil {
		//nolint:nilerr // not a number at all, so the real unmarshal names the field
		return raw, nil
	}
	if f != math.Trunc(f) {
		return nil, errors.New("must be an integer")
	}
	// Out of int64 range: leave it to the real unmarshal to say so. The upper bound is
	// 2^63 rather than MaxInt64 because float64(math.MaxInt64) rounds up to exactly 2^63.
	if f < math.MinInt64 || f >= float64(1<<63) {
		return raw, nil
	}
	return json.RawMessage(strconv.FormatInt(int64(f), 10)), nil
}

// popReadAllParams consumes the reserved listing arguments, leaving the
// model-bound ones for applyArgs.
func popReadAllParams(args map[string]json.RawMessage) (search string, page, perPage int, err error) {
	pop := func(name string, dst any) error {
		raw, ok := args[name]
		if !ok {
			return nil
		}
		delete(args, name)
		if _, isInt := dst.(*int); isInt {
			normalized, err := integralNumber(raw)
			if err != nil {
				return fmt.Errorf("invalid value for %q: %w", name, err)
			}
			raw = normalized
		}
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

	page, perPage, err = handler.NormalizePagination(page, perPage)
	if err != nil {
		return "", 0, 0, err
	}
	return search, page, perPage, nil
}
