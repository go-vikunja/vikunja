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

	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func specFor(t *testing.T, api huma.API, method, path string) *toolSpec {
	t.Helper()
	item := api.OpenAPI().Paths[path]
	require.NotNil(t, item)
	var op *huma.Operation
	switch method {
	case http.MethodGet:
		op = item.Get
	case http.MethodPost:
		op = item.Post
	case http.MethodPatch:
		op = item.Patch
	}
	require.NotNil(t, op)
	spec, err := buildToolSpec(api.OpenAPI(), op)
	require.NoError(t, err)
	return spec
}
func TestBuildToolSpec_Create(t *testing.T) {
	spec := specFor(t, newTestAPI(t), http.MethodPost, "/things")
	props := spec.schema.Properties
	for _, name := range []string{
		"title",
		"description",
		"reminders",
	} {
		assert.Contains(t, props, name)
	}
	assert.NotContains(t, props, "id")
	assert.NotContains(t, props, "owner")
	assert.Equal(t, []string{"title"}, spec.schema.Required)
	assert.ElementsMatch(t, []string{
		"array",
		"null",
	}, props["reminders"].Types)
	require.NotNil(t, props["reminders"].Items)
	assert.Equal(t, "object", props["reminders"].Items.Type)
	assert.Contains(t, props["reminders"].Items.Properties, "at")
	assert.True(t, spec.hasBody)
	assert.Empty(t, spec.params)
	assert.NotNil(t, spec.schema.AdditionalProperties)
}
func TestBuildToolSpec_ListParams(t *testing.T) {
	spec := specFor(t, newTestAPI(t), http.MethodGet, "/things")
	assert.Equal(t, "query", spec.params["q"].In)
	assert.ElementsMatch(t, []string{
		"array",
		"null",
	}, spec.schema.Properties["expand"].Types)
	assert.NotContains(t, spec.schema.Properties, "Authorization")
	assert.False(t, spec.hasBody)
}
func TestBuildToolSpec_Patch(t *testing.T) {
	spec := specFor(t, newTestAPI(t), http.MethodPatch, "/things/{id}")
	assert.Equal(t, "path", spec.params["id"].In)
	assert.Equal(t, []string{"id"}, spec.schema.Required)
	assert.Contains(t, spec.schema.Properties, "title")
	assert.NotContains(t, spec.schema.Properties, "owner")
	reminders := spec.schema.Properties["reminders"]
	require.NotNil(t, reminders)
	require.NotNil(t, reminders.Items)
	assert.Equal(t, "object", reminders.Items.Type)
	assert.Contains(t, reminders.Items.Properties, "at")
}
func TestBuildToolSpec_PutKeepsRequiredBodyFields(t *testing.T) {
	cfg := huma.DefaultConfig("test", "1")
	cfg.FieldsOptionalByDefault = true
	_, api := humatest.New(t, cfg)
	huma.Register(api, huma.Operation{
		OperationID: "thing-position-update",
		Method:      http.MethodPut,
		Path:        "/things/{id}/position",
	}, func(_ context.Context, _ *struct {
		ID   int64 `path:"id"`
		Body struct {
			Position float64 `json:"position" required:"true"`
		}
	}) (*struct{}, error) {
		return nil, nil
	})
	spec, err := buildToolSpec(api.OpenAPI(), api.OpenAPI().Paths["/things/{id}/position"].Put)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"id",
		"position",
	}, spec.schema.Required)
}
func TestBuildToolSpec_ParamBodyCollision(t *testing.T) {
	_, api := humatest.New(t)
	huma.Register(api, huma.Operation{
		OperationID: "clash",
		Method:      http.MethodPost,
		Path:        "/clash/{title}",
	}, func(_ context.Context, _ *struct {
		Title string `path:"title"`
		Body  testThing
	}) (*struct{}, error) {
		return nil, nil
	})
	_, err := buildToolSpec(api.OpenAPI(), api.OpenAPI().Paths["/clash/{title}"].Post)
	require.Error(t, err)
}
func TestInit_BuildsToolIndex(t *testing.T) {
	Init(newTestAPI(t), "")
	var names []string
	for _, tl := range snapshotTools() {
		names = append(names, tl.name)
	}
	assert.ElementsMatch(t, []string{
		"things_list",
		"things_read",
		"things_create",
		"things_update",
		"things_delete",
	}, names)
	update, ok := findTool("things_update")
	require.True(t, ok)
	assert.Equal(t, http.MethodPatch, update.op.Method)
	assert.Equal(t, "application/merge-patch+json", update.contentType)
	assert.Equal(t, "/things/:id", update.echoPath)
	assert.Contains(t, update.description, "Only fields present")
	read, _ := findTool("things_read")
	assert.Equal(t, TierCatalog, read.tier)
	_, denied := findTool("tokens_create")
	assert.False(t, denied)
}
func TestInit_PrefixesEchoPath(t *testing.T) {
	Init(newTestAPI(t), "/api/v2")
	read, _ := findTool("things_read")
	assert.Equal(t, "/api/v2/things/:id", read.echoPath)
}

type recursiveInput struct {
	Title    string            `json:"title"`
	Children []*recursiveInput `json:"children"`
}

func TestBuildToolSpec_RecursiveSchemaIsSelfContained(t *testing.T) {
	_, api := humatest.New(t)
	huma.Register(api, huma.Operation{
		OperationID: "recursive-create",
		Method:      http.MethodPost,
		Path:        "/recursive",
	},
		func(_ context.Context, _ *struct{ Body recursiveInput }) (*struct{}, error) { return nil, nil })
	before, err := json.Marshal(api.OpenAPI())
	require.NoError(t, err)
	spec := specFor(t, api, http.MethodPost, "/recursive")
	raw, err := json.Marshal(spec.schema)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `"$ref"`)
	after, err := json.Marshal(api.OpenAPI())
	require.NoError(t, err)
	assert.JSONEq(t, string(before), string(after))
}

func TestBuildToolSpec_TaskCreationRequiresTitle(t *testing.T) {
	cfg := huma.DefaultConfig("test", "1")
	cfg.FieldsOptionalByDefault = true
	_, api := humatest.New(t, cfg)
	huma.Register(api, huma.Operation{
		OperationID: "tasks-create",
		Method:      http.MethodPost,
		Path:        "/tasks",
	},
		func(_ context.Context, _ *struct{ Body models.Task }) (*struct{}, error) { return nil, nil })
	spec := specFor(t, api, http.MethodPost, "/tasks")
	assert.Equal(t, []string{"title"}, spec.schema.Required)
}
