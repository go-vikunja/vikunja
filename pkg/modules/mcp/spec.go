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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/jsonschema-go/jsonschema"
)

type toolSpec struct {
	schema   *jsonschema.Schema
	resolved *jsonschema.Resolved
	params   map[string]*huma.Param
	hasBody  bool
}

func falseSchema() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }

const maxInlineDepth = 8

const formatParamDescription = "Rich-text format for description fields: html (default) or markdown. Updates always exchange HTML."

func buildToolSpec(oapi *huma.OpenAPI, op *huma.Operation) (*toolSpec, error) {
	props := map[string]*jsonschema.Schema{}
	params := map[string]*huma.Param{}
	var required []string
	for _, p := range op.Parameters {
		if p.In != "path" && p.In != "query" {
			continue
		}
		ps, err := toJSONSchema(oapi, p.Schema)
		if err != nil {
			return nil, fmt.Errorf("mcp: %s: param %s: %w", op.OperationID, p.Name, err)
		}
		if ps.Description == "" {
			ps.Description = p.Description
		}
		// The API description this parameter refers to is not reachable over MCP.
		if p.In == "query" && p.Name == "format" {
			ps.Description = formatParamDescription
		}
		props[p.Name] = ps
		params[p.Name] = p
		if p.Required {
			required = append(required, p.Name)
		}
	}
	hasBody := false
	if _, body := bodyMedia(bodySchemaOp(oapi, op)); body != nil {
		fromPath := boundToPathParams(oapi, body, params)
		body = inlineRefs(oapi, body, 0)
		hasBody = true
		for name, prop := range body.Properties {
			if prop.ReadOnly || fromPath[name] {
				continue
			}
			if _, clash := props[name]; clash {
				return nil, fmt.Errorf("mcp: %s: body property %q collides with a parameter", op.OperationID, name)
			}
			ps, err := toJSONSchema(oapi, prop)
			if err != nil {
				return nil, fmt.Errorf("mcp: %s: body property %s: %w", op.OperationID, name, err)
			}
			if op.Method == http.MethodPatch {
				ps = allowNull(ps)
			}
			props[name] = ps
		}
		if op.Method != http.MethodPatch {
			required = append(required, body.Required...)
		}
		if op.Method == http.MethodPost {
			required = append(required, requiredByValidTag(oapi, op)...)
			// Task creation requires a title in model logic, outside valid tags.
			if op.OperationID == "tasks-create" {
				required = append(required, "title")
			}
		}
	}
	required = slices.DeleteFunc(required, func(name string) bool { return props[name] == nil })
	sort.Strings(required)
	schema := &jsonschema.Schema{
		Type:                 "object",
		Properties:           props,
		Required:             slices.Compact(required),
		AdditionalProperties: falseSchema(),
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("mcp: resolve schema for %s: %w", op.OperationID, err)
	}
	return &toolSpec{
		schema:   schema,
		resolved: resolved,
		params:   params,
		hasBody:  hasBody,
	}, nil
}

// Handlers bind these from the path and ignore the body value.
func boundToPathParams(oapi *huma.OpenAPI, body *huma.Schema, params map[string]*huma.Param) map[string]bool {
	out := map[string]bool{}
	if body.Ref == "" {
		return out
	}
	t := oapi.Components.Schemas.TypeFromRef(body.Ref)
	if t == nil {
		return out
	}
	walkFields(t, func(f reflect.StructField) {
		p, ok := params[f.Tag.Get("param")]
		if !ok || p.In != "path" {
			return
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name != "" && name != "-" {
			out[name] = true
		}
	})
	return out
}

// Merge-patch clears a field by sending null; the shape read from the PUT does not allow it.
func allowNull(s *jsonschema.Schema) *jsonschema.Schema {
	switch {
	case s.Type == "null" || slices.Contains(s.Types, "null"):
		return s
	case s.Type != "":
		s.Types = []string{
			s.Type,
			"null",
		}
		s.Type = ""
	case len(s.Types) > 0:
		s.Types = append(s.Types, "null")
	case len(s.AnyOf) > 0:
		s.AnyOf = append(s.AnyOf, &jsonschema.Schema{Type: "null"})
	case len(s.OneOf) > 0:
		s.OneOf = append(s.OneOf, &jsonschema.Schema{Type: "null"})
	}
	if len(s.Enum) > 0 {
		s.Enum = append(s.Enum, nil)
	}
	return s
}

// AutoPatch's PATCH body drops refs and nullability, collapsing nested schemas to {}; read the shape from the PUT instead.
func bodySchemaOp(oapi *huma.OpenAPI, op *huma.Operation) *huma.Operation {
	if op.Method != http.MethodPatch {
		return op
	}
	item := oapi.Paths[op.Path]
	if item == nil || item.Put == nil {
		return op
	}
	return item.Put
}
func inlineRefs(oapi *huma.OpenAPI, s *huma.Schema, depth int) *huma.Schema {
	if s == nil {
		return nil
	}
	if depth > maxInlineDepth {
		return &huma.Schema{ReadOnly: s.ReadOnly}
	}
	copied := *s
	if s.Ref != "" {
		if target := oapi.Components.Schemas.SchemaFromRef(s.Ref); target != nil {
			copied = *target
			copied.ReadOnly = copied.ReadOnly || s.ReadOnly
			if s.Description != "" {
				copied.Description = s.Description
			}
		}
	}
	s = &copied
	if s.Properties != nil {
		props := make(map[string]*huma.Schema, len(s.Properties))
		for k, v := range s.Properties {
			props[k] = inlineRefs(oapi, v, depth+1)
		}
		s.Properties = props
	}
	s.Items = inlineRefs(oapi, s.Items, depth+1)
	if ap, ok := s.AdditionalProperties.(*huma.Schema); ok {
		s.AdditionalProperties = inlineRefs(oapi, ap, depth+1)
	}
	for _, list := range []*[]*huma.Schema{
		&s.AnyOf,
		&s.OneOf,
		&s.AllOf,
	} {
		if *list == nil {
			continue
		}
		out := make([]*huma.Schema, len(*list))
		for i, v := range *list {
			out[i] = inlineRefs(oapi, v, depth+1)
		}
		*list = out
	}
	return s
}
func toJSONSchema(oapi *huma.OpenAPI, s *huma.Schema) (*jsonschema.Schema, error) {
	raw, err := json.Marshal(inlineRefs(oapi, s, 0))
	if err != nil {
		return nil, err
	}
	out := &jsonschema.Schema{}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Huma leaves body fields optional; govalidator requires these on creation.
func requiredByValidTag(oapi *huma.OpenAPI, op *huma.Operation) []string {
	_, body := bodyMedia(op)
	if body == nil || body.Ref == "" {
		return nil
	}
	t := oapi.Components.Schemas.TypeFromRef(body.Ref)
	if t == nil {
		return nil
	}
	var out []string
	walkFields(t, func(f reflect.StructField) {
		if !slices.Contains(strings.Split(f.Tag.Get("valid"), ","), "required") {
			return
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name != "" && name != "-" {
			out = append(out, name)
		}
	})
	return out
}
func walkFields(t reflect.Type, visit func(reflect.StructField)) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			walkFields(f.Type, visit)
			continue
		}
		if f.IsExported() {
			visit(f)
		}
	}
}

func mustResolveSpec(name string, schema *jsonschema.Schema) *toolSpec {
	resolved, err := schema.Resolve(nil)
	if err != nil {
		panic(fmt.Sprintf("mcp: resolve %s schema: %v", name, err))
	}
	return &toolSpec{
		schema:   schema,
		resolved: resolved,
	}
}
