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
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	apiv2 "code.vikunja.io/api/pkg/routes/api/v2"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExposure_TaskOccurrences: the occurrence query and the preview, generated from the real v2
// operations, are typed tools on tools/list, with their parameters and their operation
// descriptions, so an agent can ask for a task's next dates or check a rule before saving it.
func TestExposure_TaskOccurrences(t *testing.T) {
	config.InitDefaultConfig()
	e := echo.New()
	api := apiv2.NewAPI(e, e.Group(apiv2.GroupPrefix))
	apiv2.RegisterTaskOccurrencesRoutes(api)
	apiv2.EnableAutoPatch(api)

	m, err := New(api, nil)
	require.NoError(t, err)
	m.authorize = func(*models.APIToken, string, string) bool { return true }
	cs := connect(t, func(srv *sdk.Server) { m.addToolsAuthorizedBy(srv, nil) })
	listed, err := cs.ListTools(context.Background(), nil)
	require.NoError(t, err)

	tools := map[string]map[string]any{}
	descriptions := map[string]string{}
	for _, tl := range listed.Tools {
		raw, err := json.Marshal(tl.InputSchema)
		require.NoError(t, err)
		schema := map[string]any{}
		require.NoError(t, json.Unmarshal(raw, &schema))
		tools[tl.Name] = schema
		descriptions[tl.Name] = tl.Description
	}

	for name, want := range map[string]struct {
		operationID string
		properties  []string
	}{
		"tasks_occurrences":         {"tasks-occurrences", []string{"task", "count"}},
		"tasks_occurrences_preview": {"tasks-occurrences-preview", []string{"repeat", "repeats_rrule", "due_date", "timezone", "count", "repeats_from_current_date"}},
	} {
		t.Run(name, func(t *testing.T) {
			schema, ok := tools[name]
			require.True(t, ok, "%s is a typed tool on tools/list: %v", name, descriptions)
			properties, ok := schema["properties"].(map[string]any)
			require.True(t, ok, "%s has input properties: %v", name, schema)
			for _, property := range want.properties {
				assert.Contains(t, properties, property)
			}

			op := operationByID(t, api.OpenAPI(), want.operationID)
			assert.Contains(t, descriptions[name], op.Summary)
			assert.Contains(t, descriptions[name], op.Description)
		})
	}

	assert.Contains(t, descriptions["tasks_occurrences"], "task_done")
	assert.Contains(t, descriptions["tasks_occurrences_preview"], "exactly one of repeat or repeats_rrule")
}

func operationByID(t *testing.T, doc *huma.OpenAPI, id string) *huma.Operation {
	t.Helper()
	for _, item := range doc.Paths {
		for _, op := range []*huma.Operation{item.Get, item.Put, item.Post, item.Patch, item.Delete} {
			if op != nil && op.OperationID == id {
				return op
			}
		}
	}
	require.Failf(t, "operation not found", "%s", id)
	return nil
}
