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

// Tool input schemas are reflected from the model's struct tags at
// registration time — the same tag contract the Huma-backed /api/v2 reads:
// `json:` for property names, `doc:` for descriptions, `readOnly:"true"`
// for server-controlled fields, `minLength`/`valid:"required"` for
// create-required detection, and `param:` for fields the REST layer binds
// from the URL. MCP has no URL, so param-bound fields become JSON
// properties (hidden `json:"-"` ones under the snake_cased Go field name).
//
// Field selection per op:
//   - create: every writable field (json-named, not readOnly); required if
//     the tags say so (valid:"required", minLength ≥ 1, or URL-bound).
//   - update: `id` (required) + every writable field, all optional. Only
//     fields present in the arguments are applied — see apply.go.
//   - read_one / delete: `id` + hidden param fields. Models without an
//     exposed id (e.g. task assignees) instead require their param-tagged
//     identifying fields.
//   - read_all: search/page/per_page + `query:`-tagged fields (the
//     TaskCollection filter surface) + hidden param fields.

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
)

// Reserved read_all argument names. They map to handler.DoReadAll's
// positional parameters, not to model fields.
const (
	argSearch  = "search"
	argPage    = "page"
	argPerPage = "per_page"
)

// opSpec is the cached per-(resource, op) tool contract: the input schema
// exposed over MCP, its resolved form for validation, and the json-name →
// struct-field mapping the apply step uses.
type opSpec struct {
	schema   *jsonschema.Schema
	resolved *jsonschema.Resolved
	fields   map[string]int
}

var timeType = reflect.TypeOf(time.Time{})

// falseSchema is JSON Schema `false` — used as additionalProperties so
// unknown argument names are rejected at validation time with a clear error
// instead of being silently dropped.
func falseSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

func buildOpSpec(modelType reflect.Type, op Op, r *Resource) (*opSpec, error) {
	props := map[string]*jsonschema.Schema{}
	fields := map[string]int{}
	var required []string

	excluded := func(name string) bool { return slices.Contains(r.Exclude, name) }
	optional := func(name string) bool { return slices.Contains(r.OptionalFields, name) }

	hasExposedID := false
	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)
		if f.Name != "ID" {
			continue
		}
		if _, ok := jsonName(f); ok {
			hasExposedID = true
		}
	}

	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)
		if !f.IsExported() || f.Anonymous {
			continue
		}
		name, hasJSON := jsonName(f)
		param := f.Tag.Get("param")

		identity := func(name string) bool { return slices.Contains(r.IdentityFields, name) }

		switch {
		case f.Name == "ID":
			if !hasJSON || excluded("id") {
				continue
			}
			if op != OpReadOne && op != OpUpdate && op != OpDelete {
				continue
			}
			// Resources whose rows aren't addressed by their id (e.g. team
			// members, addressed by team + username) declare IdentityFields
			// without "id" and the property disappears entirely.
			if len(r.IdentityFields) > 0 && !identity("id") {
				continue
			}
			if f.Type.Kind() != reflect.Int64 {
				return nil, fmt.Errorf("mcp: %s: ID field must be int64, got %s", modelType, f.Type)
			}
			props["id"] = propWithDoc(&jsonschema.Schema{Type: "integer"}, f)
			fields["id"] = i
			required = append(required, "id")

		case !hasJSON && param != "":
			// URL-bound in REST with no JSON name: expose it under the
			// snake_cased Go field name so MCP callers can supply it.
			hidden := snakeCase(f.Name)
			if excluded(hidden) {
				continue
			}
			if f.Type.Kind() != reflect.Int64 {
				continue
			}
			props[hidden] = propWithDoc(&jsonschema.Schema{Type: "integer"}, f)
			fields[hidden] = i
			if !optional(hidden) {
				required = append(required, hidden)
			}

		// readOnly with a param tag means "REST takes this from the URL,
		// not the body" (e.g. TaskRelation.TaskID) — MCP has no URL, so it
		// stays an argument.
		case !hasJSON, f.Tag.Get("readOnly") == "true" && param == "", excluded(name):
			continue

		default:
			ps, ok := propSchema(f)
			if !ok {
				continue
			}
			include, req := false, false
			switch op {
			case OpCreate:
				include = true
				req = requiredForCreate(f, name, r)
			case OpUpdate:
				include = true
				req = identity(name)
			case OpReadOne, OpDelete:
				// Models without an exposed id (e.g. TaskAssginee) are
				// identified by their param-tagged fields instead;
				// IdentityFields declares the set explicitly when the
				// derivation can't know it (e.g. views need project_id too).
				if (!hasExposedID && param != "") || identity(name) {
					include, req = true, true
				}
			case OpReadAll:
				include = f.Tag.Get("query") != ""
			}
			if !include {
				continue
			}
			props[name] = ps
			fields[name] = i
			if req {
				required = append(required, name)
			}
		}
	}

	if op == OpReadAll {
		props[argSearch] = &jsonschema.Schema{Type: "string", Description: "Filter results by a case-insensitive substring match on the resource's primary text field."}
		props[argPage] = &jsonschema.Schema{Type: "integer", Description: "1-based page number; 0 or omitted uses the server default (first page)."}
		props[argPerPage] = &jsonschema.Schema{Type: "integer", Description: "Page size; 0 or omitted uses the server default."}
	}

	sort.Strings(required)
	schema := &jsonschema.Schema{
		Type:                 "object",
		Properties:           props,
		Required:             required,
		AdditionalProperties: falseSchema(),
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("mcp: resolve schema for %s_%s: %w", r.Name, op.ToolSuffix(), err)
	}
	return &opSpec{schema: schema, resolved: resolved, fields: fields}, nil
}

// propSchema maps a struct field's Go type onto a JSON Schema property.
// Returns false for kinds MCP doesn't expose (nested model structs/slices —
// those relations have their own resources).
func propSchema(f reflect.StructField) (*jsonschema.Schema, bool) {
	s := &jsonschema.Schema{}
	t := f.Type
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch {
	case t == timeType:
		s.Type = "string"
		s.Format = "date-time"
	case t.Kind() == reflect.String:
		s.Type = "string"
		if v, err := strconv.Atoi(f.Tag.Get("minLength")); err == nil {
			s.MinLength = &v
		}
		if v, err := strconv.Atoi(f.Tag.Get("maxLength")); err == nil {
			s.MaxLength = &v
		}
	case t.Kind() == reflect.Bool:
		s.Type = "boolean"
	case t.Kind() >= reflect.Int && t.Kind() <= reflect.Uint64:
		s.Type = "integer"
	case t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64:
		s.Type = "number"
	case t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.String:
		s.Type = "array"
		s.Items = &jsonschema.Schema{Type: "string"}
	default:
		return nil, false
	}
	// Named int types with a custom string MarshalJSON declare their wire
	// type via swaggertype (e.g. ProjectViewKind).
	if f.Tag.Get("swaggertype") == "string" {
		s.Type = "string"
		s.Format = ""
	}
	// Both huma-style `enum` and swaggo-style `enums` list allowed values.
	enum := f.Tag.Get("enum")
	if enum == "" {
		enum = f.Tag.Get("enums")
	}
	if enum != "" && s.Type == "string" {
		for _, v := range strings.Split(enum, ",") {
			s.Enum = append(s.Enum, v)
		}
	}
	return propWithDoc(s, f), true
}

func propWithDoc(s *jsonschema.Schema, f reflect.StructField) *jsonschema.Schema {
	if d := f.Tag.Get("doc"); d != "" {
		s.Description = d
	}
	return s
}

// requiredForCreate reports whether a writable field must be supplied on
// create: an explicit per-resource override, a `valid:"required"` rule, a
// non-zero minLength, or a `param:` tag (REST binds it from the URL, so a
// create without it can never succeed).
func requiredForCreate(f reflect.StructField, name string, r *Resource) bool {
	if slices.Contains(r.RequiredCreate, name) {
		return true
	}
	if strings.Contains(f.Tag.Get("valid"), "required") {
		return true
	}
	if ml, err := strconv.Atoi(f.Tag.Get("minLength")); err == nil && ml > 0 {
		return true
	}
	return f.Tag.Get("param") != ""
}

// jsonName extracts the JSON property name from a struct field's `json`
// tag. Returns ("", false) for fields with no tag or tagged "-".
func jsonName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		return "", false
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

// snakeCase converts a Go field name to snake_case, collapsing acronyms:
// TaskID → task_id, OtherTaskID → other_task_id.
func snakeCase(name string) string {
	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			prevLower := i > 0 && runes[i-1] >= 'a' && runes[i-1] <= 'z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if i > 0 && (prevLower || nextLower) {
				b.WriteByte('_')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}
