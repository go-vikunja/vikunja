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

package apiv2

import (
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskRepeatSchema_OpenAPI: every field of the structured repeat, and the task's repeat
// fields, are described in the v2 OpenAPI document, which agents (and the MCP tools generated from
// it) read to create, change and clear recurring tasks.
func TestTaskRepeatSchema_OpenAPI(t *testing.T) {
	config.InitDefaultConfig()
	e := echo.New()
	api := NewAPI(e, e.Group(GroupPrefix))
	RegisterTaskRoutes(api)
	schemas := api.OpenAPI().Components.Schemas.Map()

	repeat := schemas["TaskRepeat"]
	require.NotNil(t, repeat, "the TaskRepeat component schema")
	require.NotEmpty(t, repeat.Properties)
	for name, prop := range repeat.Properties {
		assert.NotEmpty(t, prop.Description, "TaskRepeat.%s has no description", name)
	}
	facts := map[string][]string{
		"freq":       {"secondly", "lowercase"},
		"interval":   {"100 years"},
		"by_day":     {"2tu", "-1fr", "lowercase"},
		"by_set_pos": {"10"},
		"count":      {"10000", "completion"},
		"until":      {"RFC 3339", "UTC"},
		"wkst":       {"mo"},
		"read_only":  {"removed"},
	}
	for name, want := range facts {
		prop := repeat.Properties[name]
		require.NotNil(t, prop, "TaskRepeat.%s", name)
		for _, w := range want {
			assert.Contains(t, prop.Description, w, "TaskRepeat.%s", name)
		}
	}

	task := schemas["Task"]
	require.NotNil(t, task, "the Task component schema")
	require.NotNil(t, task.Properties["repeat"])
	assert.Contains(t, task.Properties["repeat"].Description, "null")
	assert.Contains(t, task.Properties["repeat"].Description, "500 characters")
	require.NotNil(t, task.Properties["repeats_from_current_date"])
	assert.Contains(t, task.Properties["repeats_from_current_date"].Description, "completed")
}
