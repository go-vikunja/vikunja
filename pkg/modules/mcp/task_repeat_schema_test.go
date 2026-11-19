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
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/config"
	apiv2 "code.vikunja.io/api/pkg/routes/api/v2"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskRepeatSchema_MCPTool: the MCP task tools, generated from the real v2 task operations,
// carry the descriptions of the repeat object, its fields and repeats_from_current_date, so an
// agent can create, change and clear recurring tasks without guessing.
func TestTaskRepeatSchema_MCPTool(t *testing.T) {
	config.InitDefaultConfig()
	e := echo.New()
	api := apiv2.NewAPI(e, e.Group(apiv2.GroupPrefix))
	apiv2.RegisterTaskRoutes(api)
	apiv2.EnableAutoPatch(api)
	oapi := api.OpenAPI()

	ops := map[string]*huma.Operation{
		"create (POST)":          oapi.Paths["/projects/{project}/tasks"].Post,
		"full update (PUT)":      oapi.Paths["/tasks/{task}"].Put,
		"partial update (PATCH)": oapi.Paths["/tasks/{task}"].Patch,
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			require.NotNil(t, op, "the operation exists")
			spec, err := buildToolSpec(oapi, op)
			require.NoError(t, err)

			repeat := spec.schema.Properties["repeat"]
			require.NotNil(t, repeat, "the tool takes repeat")
			assert.Contains(t, repeat.Description, "null")
			require.NotEmpty(t, repeat.Properties, "the repeat object's fields are inlined")
			for field, prop := range repeat.Properties {
				assert.NotEmpty(t, prop.Description, "repeat.%s has no description", field)
			}
			assert.Contains(t, repeat.Properties["by_day"].Description, "2tu")
			assert.Contains(t, repeat.Properties["count"].Description, "10000")

			fromCompletion := spec.schema.Properties["repeats_from_current_date"]
			require.NotNil(t, fromCompletion)
			assert.Contains(t, fromCompletion.Description, "completed")
		})
	}
	require.NotNil(t, ops["partial update (PATCH)"], "autopatch adds the PATCH the tasks_update tool calls")
	assert.Equal(t, http.MethodPatch, ops["partial update (PATCH)"].Method)
}
