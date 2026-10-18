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
	"net/http"
	"testing"

	apiv2 "code.vikunja.io/api/pkg/routes/api/v2"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	exposedOperations["things-page"] = catalogTool
}

func TestFieldTree_Apply(t *testing.T) {
	v := map[string]any{
		"id":    1,
		"title": "t",
		"owner": map[string]any{"id": 2, "name": "n"},
		"labels": []any{
			map[string]any{"id": 3, "title": "a"},
			map[string]any{"id": 4, "title": "b"},
		},
	}
	got := parseFields([]string{"id", "owner.name", "labels.title", "missing", "owner.missing"}).apply(v)
	assert.Equal(t, map[string]any{
		"id":    1,
		"owner": map[string]any{"name": "n"},
		"labels": []any{
			map[string]any{"title": "a"},
			map[string]any{"title": "b"},
		},
	}, got)
}
func TestParseFields_WholeValueWins(t *testing.T) {
	assert.Equal(t, fieldTree{"owner": nil}, parseFields([]string{"owner.name", "owner"}))
	assert.Equal(t, fieldTree{"owner": nil}, parseFields([]string{"owner", "owner.name"}))
}
func TestBuildToolSpec_FieldsOnlyOnReads(t *testing.T) {
	api := newTestAPI(t)
	assert.Contains(t, specFor(t, api, http.MethodGet, "/things/{id}").schema.Properties, fieldsParam)
	assert.Contains(t, specFor(t, api, http.MethodGet, "/things").schema.Properties, fieldsParam)
	assert.NotContains(t, specFor(t, api, http.MethodPost, "/things").schema.Properties, fieldsParam)
	assert.NotContains(t, specFor(t, api, http.MethodPatch, "/things/{id}").schema.Properties, fieldsParam)
}
func TestBuildToolSpec_FieldsParamCollision(t *testing.T) {
	_, api := humatest.New(t)
	huma.Register(api, huma.Operation{
		OperationID: "clash-list",
		Method:      http.MethodGet,
		Path:        "/clash",
	}, func(_ context.Context, _ *struct {
		Fields string `query:"fields"`
	}) (*testThingBody, error) {
		return nil, nil
	})
	_, err := buildToolSpec(api.OpenAPI(), api.OpenAPI().Paths["/clash"].Get)
	require.Error(t, err)
}
func TestCallTool_FieldsProjectsTheResponse(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_read", json.RawMessage(`{"id":7,"fields":["title","owner.id"]}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"title": "stored",
		"owner": map[string]any{"id": json.Number("1")},
	}, res)
}
func TestCallTool_FieldsRejectsEmptyList(t *testing.T) {
	m, ctx := newTestCaller(t)
	_, err := m.callTool(ctx, "things_read", json.RawMessage(`{"id":7,"fields":[]}`))
	require.Error(t, err)
}
func TestCallTool_FieldsKeepsThePageEnvelope(t *testing.T) {
	cfg := huma.DefaultConfig("test", "1")
	cfg.FieldsOptionalByDefault = true
	_, api := humatest.New(t, cfg)
	huma.Register(api, huma.Operation{
		OperationID: "things-page",
		Method:      http.MethodGet,
		Path:        "/things",
	}, func(_ context.Context, _ *struct{}) (*struct{ Body apiv2.Paginated[testThing] }, error) {
		out := &struct{ Body apiv2.Paginated[testThing] }{}
		out.Body = apiv2.NewPaginated([]testThing{{ID: 1, Title: "a"}, {ID: 2, Title: "b"}}, 2, 1, 50)
		return out, nil
	})
	mod, err := New(api, nil)
	require.NoError(t, err)
	m, ctx := newTestCallerFor(t, mod)
	res, err := m.callTool(ctx, "things_page", json.RawMessage(`{"fields":["title"]}`))
	require.NoError(t, err)
	out := echoResult(t, res)
	assert.Equal(t, []any{
		map[string]any{"title": "a"},
		map[string]any{"title": "b"},
	}, out["items"])
	assert.Equal(t, json.Number("2"), out["total"])
	assert.Equal(t, json.Number("1"), out["total_pages"])
}
