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
	"maps"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/jsonschema-go/jsonschema"
)

const fieldsParam = "fields"

type projection int

const (
	noProjection projection = iota
	projectBody
	// Keeps the Paginated envelope's paging metadata.
	projectItems
)

func projectionFor(oapi *huma.OpenAPI, op *huma.Operation) projection {
	if op.Method != http.MethodGet {
		return noProjection
	}
	res := op.Responses["200"]
	if res == nil || res.Content["application/json"] == nil || res.Content["application/json"].Schema == nil {
		return noProjection
	}
	s := res.Content["application/json"].Schema
	if s.Ref != "" {
		s = oapi.Components.Schemas.SchemaFromRef(s.Ref)
	}
	if s != nil && s.Properties["items"] != nil && s.Properties["total_pages"] != nil {
		return projectItems
	}
	return projectBody
}

func fieldsSchema() *jsonschema.Schema {
	one := 1
	return &jsonschema.Schema{
		Type:        "array",
		Items:       &jsonschema.Schema{Type: "string", MinLength: &one},
		MinItems:    &one,
		Description: `Return only these response fields. Dot paths select nested fields, also inside arrays (e.g. ["id", "title", "labels.title"]). Omit for the full response.`,
	}
}

// A nil subtree selects the whole value.
type fieldTree map[string]fieldTree

func parseFields(paths []string) fieldTree {
	root := fieldTree{}
	for _, path := range paths {
		node := root
		parts := strings.Split(path, ".")
		for i, part := range parts {
			sub, seen := node[part]
			if seen && sub == nil {
				break
			}
			if i == len(parts)-1 {
				node[part] = nil
				break
			}
			if sub == nil {
				sub = fieldTree{}
				node[part] = sub
			}
			node = sub
		}
	}
	return root
}

func (t fieldTree) apply(v any) any {
	switch x := v.(type) {
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = t.apply(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, sub := range t {
			val, ok := x[k]
			switch {
			case !ok:
				continue
			case sub == nil:
				out[k] = val
			default:
				out[k] = sub.apply(val)
			}
		}
		return out
	}
	return v
}

// takeFields removes the fields argument so it is not forwarded to the v2 route.
func takeFields(args map[string]json.RawMessage) (fieldTree, error) {
	raw, ok := args[fieldsParam]
	if !ok {
		return nil, nil
	}
	delete(args, fieldsParam)
	var paths []string
	if err := json.Unmarshal(raw, &paths); err != nil {
		return nil, err
	}
	return parseFields(paths), nil
}

func (p projection) apply(fields fieldTree, result any) any {
	if fields == nil {
		return result
	}
	switch p {
	case projectItems:
		envelope, ok := result.(map[string]any)
		if !ok {
			return result
		}
		out := maps.Clone(envelope)
		out["items"] = fields.apply(envelope["items"])
		return out
	case projectBody:
		return fields.apply(result)
	case noProjection:
	}
	return result
}
