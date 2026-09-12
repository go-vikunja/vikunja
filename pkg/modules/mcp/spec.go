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
		props[p.Name] = ps
		params[p.Name] = p
		if p.Required {
			required = append(required, p.Name)
		}
	}
	hasBody := false
	if _, body := bodyMedia(op); body != nil {
		body = inlineRefs(oapi, body, 0)
		hasBody = true
		for name, prop := range body.Properties {
			if prop.ReadOnly {
				continue
			}
			if _, clash := props[name]; clash {
				return nil, fmt.Errorf("mcp: %s: body property %q collides with a parameter", op.OperationID, name)
			}
			ps, err := toJSONSchema(oapi, prop)
			if err != nil {
				return nil, fmt.Errorf("mcp: %s: body property %s: %w", op.OperationID, name, err)
			}
			props[name] = ps
		}
		if op.Method != http.MethodPut && op.Method != http.MethodPatch {
			required = append(required, body.Required...)
		}
		if op.Method == http.MethodPost {
			required = append(required, requiredByValidTag(oapi, op)...)
		}
	}
	required = slices.DeleteFunc(required, func(name string) bool { return props[name] == nil })
	sort.Strings(required)
	schema := &jsonschema.Schema{Type: "object", Properties: props, Required: slices.Compact(required), AdditionalProperties: falseSchema()}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("mcp: resolve schema for %s: %w", op.OperationID, err)
	}
	return &toolSpec{schema: schema, resolved: resolved, params: params, hasBody: hasBody}, nil
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
	for _, list := range []*[]*huma.Schema{&s.AnyOf, &s.OneOf, &s.AllOf} {
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
