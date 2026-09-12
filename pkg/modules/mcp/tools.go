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
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type tool struct {
	name        string
	op          *huma.Operation
	echoPath    string
	typed       bool
	spec        *toolSpec
	contentType string
	description string
}

var (
	toolsAPI  huma.API
	toolsMu   sync.RWMutex
	toolIndex map[string]*tool
	toolOrder []*tool
)

func initTools(api huma.API, groupPrefix string) {
	index, order, err := buildTools(api.OpenAPI(), groupPrefix)
	if err != nil {
		panic(err)
	}
	toolsMu.Lock()
	defer toolsMu.Unlock()
	toolsAPI = api
	toolIndex = index
	toolOrder = order
}

type candidate struct {
	id string
	op *huma.Operation
}

func candidates(item *huma.PathItem) []candidate {
	var out []candidate
	for _, op := range []*huma.Operation{
		item.Get,
		item.Post,
		item.Delete,
	} {
		if op != nil {
			out = append(out, candidate{
				id: op.OperationID,
				op: op,
			})
		}
	}
	switch {
	case item.Put != nil && item.Patch != nil:
		out = append(out, candidate{
			id: item.Put.OperationID,
			op: item.Patch,
		})
	case item.Put != nil:
		out = append(out, candidate{
			id: item.Put.OperationID,
			op: item.Put,
		})
	case item.Patch != nil:
		out = append(out, candidate{
			id: item.Patch.OperationID,
			op: item.Patch,
		})
	}
	return out
}
func buildTools(oapi *huma.OpenAPI, groupPrefix string) (map[string]*tool, []*tool, error) {
	index := map[string]*tool{}
	for _, item := range oapi.Paths {
		if item == nil {
			continue
		}
		for _, c := range candidates(item) {
			typed, ok := exposure(c.id, c.op)
			if !ok {
				continue
			}
			name := toolNameFor(c.id)
			if _, dup := index[name]; dup {
				return nil, nil, fmt.Errorf("mcp: duplicate tool name %s", name)
			}
			spec, err := buildToolSpec(oapi, c.op)
			if err != nil {
				return nil, nil, err
			}
			ct, _ := bodyMedia(c.op)
			index[name] = &tool{
				name:        name,
				op:          c.op,
				echoPath:    groupPrefix + echoPath(c.op.Path),
				typed:       typed,
				spec:        spec,
				contentType: ct,
				description: describe(oapi, c.op),
			}
		}
	}
	order := make([]*tool, 0, len(index))
	for _, t := range index {
		order = append(order, t)
	}
	sort.Slice(order, func(i, j int) bool { return order[i].name < order[j].name })
	return index, order, nil
}
func echoPath(path string) string { return strings.NewReplacer("{", ":", "}", "").Replace(path) }

// AutoPatch's PATCH prose is boilerplate about JSON Patch, which MCP callers cannot use; read it from the PUT instead.
func describe(oapi *huma.OpenAPI, op *huma.Operation) string {
	src := bodySchemaOp(oapi, op)
	s := src.Summary
	if src.Description != "" {
		s += ". " + src.Description
	}
	if op.Method == http.MethodPatch {
		s += " Only fields present in the arguments are changed. Rich-text fields are exchanged as HTML here."
	}
	return s
}

func findTool(name string) (*tool, bool) {
	toolsMu.RLock()
	defer toolsMu.RUnlock()
	t, ok := toolIndex[name]
	return t, ok
}
func snapshotTools() []*tool {
	toolsMu.RLock()
	defer toolsMu.RUnlock()
	out := make([]*tool, len(toolOrder))
	copy(out, toolOrder)
	return out
}

var routeAuthorizer = (*models.APIToken).CanUseRoute

func (t *tool) authorized(token *models.APIToken) bool {
	return routeAuthorizer(token, t.echoPath, t.op.Method)
}
func currentAPI() huma.API { toolsMu.RLock(); defer toolsMu.RUnlock(); return toolsAPI }
