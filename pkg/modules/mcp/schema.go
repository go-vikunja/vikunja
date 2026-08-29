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

// Tool input schemas are reflected from the same struct tags the Huma-backed
// /api/v2 reads. MCP has no URL, so `param:`-bound fields become plain JSON
// properties — hidden `json:"-"` ones under their snake_cased Go field name.

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

// These map to handler.DoReadAll's positional parameters, not to model fields.
const (
	argSearch  = "search"
	argPage    = "page"
	argPerPage = "per_page"
)

// opSpec caches the per-(resource, op) input schema plus the json-name → struct-field index apply.go needs.
type opSpec struct {
	schema   *jsonschema.Schema
	resolved *jsonschema.Resolved
	fields   map[string]int
}

var timeType = reflect.TypeOf(time.Time{})

// falseSchema is JSON Schema `false`: as additionalProperties it rejects unknown argument names instead of dropping them.
func falseSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

// mustResolveSpec builds an opSpec for the hand-written tool schemas; resource schemas come from buildOpSpec.
func mustResolveSpec(toolName string, schema *jsonschema.Schema) *opSpec {
	resolved, err := schema.Resolve(nil)
	if err != nil {
		panic(fmt.Sprintf("mcp: resolve %s schema: %v", toolName, err))
	}
	return &opSpec{schema: schema, resolved: resolved}
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
			// Rows not addressed by id (team members go by team + username) declare IdentityFields without "id".
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
			// URL-bound in REST with no JSON name, so MCP needs a synthetic property name.
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

		// readOnly with a param tag means REST reads it from the URL, not the body — MCP keeps it as an argument.
		case !hasJSON, f.Tag.Get("readOnly") == "true" && param == "", excluded(name):
			continue

		default:
			ps, ok := propSchema(f)
			if !ok {
				continue
			}
			include, req := writableInclusion(op, f, name, param, hasExposedID, r)
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
		addQueryOnlyArgs(modelType, props, fields, excluded)
		props[argSearch] = &jsonschema.Schema{Type: "string", Description: "Filter results by a case-insensitive substring match on the resource's primary text field."}
		props[argPage] = &jsonschema.Schema{Type: "integer", Description: "1-based page number; 0 or omitted means the first page. Negative values are rejected."}
		props[argPerPage] = &jsonschema.Schema{Type: "integer", Description: "Page size; 0 or omitted uses the server maximum, and larger values are clamped to it. The response reports the page size actually applied."}
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
		return nil, fmt.Errorf("mcp: resolve schema for %s_%s: %w", r.Name, op.Permission(), err)
	}
	return &opSpec{schema: schema, resolved: resolved, fields: fields}, nil
}

// addQueryOnlyArgs exposes json:"-" fields carrying a query tag (e.g. TaskCollection.Expand) as listing arguments.
func addQueryOnlyArgs(modelType reflect.Type, props map[string]*jsonschema.Schema, fields map[string]int, excluded func(string) bool) {
	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)
		name := f.Tag.Get("query")
		if _, hasJSON := jsonName(f); hasJSON || name == "" || !f.IsExported() || excluded(name) {
			continue
		}
		if ps, ok := propSchema(f); ok {
			props[name] = ps
			fields[name] = i
		}
	}
}

// propSchema returns false for nested model structs/slices; those relations have their own resources.
func propSchema(f reflect.StructField) (*jsonschema.Schema, bool) {
	s := &jsonschema.Schema{}
	t := derefType(f.Type)
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
	// Named int types with a custom string MarshalJSON declare their wire type via swaggertype.
	if f.Tag.Get("swaggertype") == "string" {
		s.Type = "string"
		s.Format = ""
	}
	// Both huma-style `enum` and swaggo-style `enums` list allowed values.
	enum := f.Tag.Get("enum")
	if enum == "" {
		enum = f.Tag.Get("enums")
	}
	if enum != "" {
		target := s
		if s.Type == "array" {
			target = s.Items
		}
		if target.Type == "string" {
			for _, v := range strings.Split(enum, ",") {
				target.Enum = append(target.Enum, v)
			}
		}
	}
	return propWithDoc(s, f), true
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func propWithDoc(s *jsonschema.Schema, f reflect.StructField) *jsonschema.Schema {
	if d := f.Tag.Get("doc"); d != "" {
		s.Description = d
	}
	return s
}

func writableInclusion(op Op, f reflect.StructField, name, param string, hasExposedID bool, r *Resource) (include, required bool) {
	identity := slices.Contains(r.IdentityFields, name)
	switch op {
	case OpCreate:
		return true, requiredForCreate(f)
	case OpUpdate:
		// Without a read_one op the dispatcher can't hydrate the stored row, so every
		// omitted writable field would be written as its zero value.
		return true, identity || r.Ops&OpReadOne == 0
	case OpReadOne, OpDelete:
		// Models without an exposed id (e.g. TaskAssginee) are addressed by their param-tagged fields instead.
		if (!hasExposedID && param != "") || identity {
			return true, true
		}
	case OpReadAll:
		if f.Tag.Get("query") != "" {
			return true, false
		}
		// REST scopes the listing by this field from the URL path; without it the listing runs against id 0.
		if param != "" && f.Tag.Get("readOnly") == "true" {
			return true, !slices.Contains(r.OptionalFields, name)
		}
	}
	return false, false
}

// A `param:` tag counts as required: REST binds it from the URL, so a create without it can never succeed.
func requiredForCreate(f reflect.StructField) bool {
	if strings.Contains(f.Tag.Get("valid"), "required") {
		return true
	}
	if ml, err := strconv.Atoi(f.Tag.Get("minLength")); err == nil && ml > 0 {
		return true
	}
	return f.Tag.Get("param") != ""
}

// jsonName returns ("", false) for fields with no json tag or tagged "-".
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

// snakeCase collapses acronyms: TaskID → task_id, OtherTaskID → other_task_id.
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
