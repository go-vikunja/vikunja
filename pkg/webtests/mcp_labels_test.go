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

package webtests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCP_Labels_ToolsListAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	names := toolNamesFromList(t, resp)

	for _, want := range []string{
		"labels_create",
		"labels_read_one",
		"labels_read_all",
		"labels_update",
		"labels_delete",
	} {
		assert.Truef(t, names[want], "missing %s in tools/list: %v", want, names)
	}
}

func TestMCP_Labels_Create(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("labels_create", map[string]any{
		"title":     "mcp label",
		"hex_color": "ff8800",
	})
	require.NotContains(t, result, "isError", "create errored: %v", result)

	text := toolResultText(t, result)
	var label map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &label))
	assert.Equal(t, "mcp label", label["title"])
	id, ok := label["id"].(float64)
	require.Truef(t, ok, "id missing: %v", label)
	assert.Positive(t, int(id))
}

func TestMCP_Labels_ReadAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("labels_read_all", map[string]any{})
	require.NotContains(t, result, "isError")

	text := toolResultText(t, result)
	var labels []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &labels))
	require.NotEmpty(t, labels, "expected at least one label")
}

func TestMCP_Labels_ReadOneForbidden(t *testing.T) {
	// Label 6 is attached only to a private task on project 20 (user 13).
	// User 1 cannot reach it.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("labels_read_one", map[string]any{"id": 6})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for inaccessible label, got: %v", result)
}
