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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCP_Projects_ToolsListAll(t *testing.T) {
	// Token 12 also carries non-project scopes, so assert presence rather
	// than an exact tool count.
	c := newMCPClient(t, mcpFullProjectsToken)
	resp := c.rpc("tools/list", map[string]any{})
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok)
	tools, ok := result["tools"].([]any)
	require.True(t, ok)

	names := make(map[string]bool, len(tools))
	for _, raw := range tools {
		tool := raw.(map[string]any)
		names[tool["name"].(string)] = true
	}
	for _, want := range []string{
		"projects_create",
		"projects_read_one",
		"projects_read_all",
		"projects_update",
		"projects_delete",
	} {
		assert.Truef(t, names[want], "missing tool %q in %v", want, names)
	}
}

func TestMCP_Projects_Create(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_create", map[string]any{
		"title":       "MCP created project",
		"description": "Created by mcp_projects_test",
	})
	require.NotContains(t, result, "isError", "create unexpectedly errored: %v", result)

	text := toolResultText(t, result)
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &project), "text was: %s", text)
	assert.Equal(t, "MCP created project", project["title"])
	assert.Equal(t, "Created by mcp_projects_test", project["description"])
	id, ok := project["id"].(float64)
	require.Truef(t, ok, "id missing or not a number: %v", project)
	assert.Positive(t, int(id))
}

func TestMCP_Projects_CreateMissingTitle(t *testing.T) {
	// "title" has no omitempty, so the SDK's schema validation rejects the
	// call inside the tool handler — an isError result, not an RPC error.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_create", map[string]any{})
	isErr, _ := result["isError"].(bool)
	require.Truef(t, isErr, "expected isError for missing required title: %v", result)
	assert.Contains(t, toolResultText(t, result), "mcp: invalid arguments for projects_create")
}

func TestMCP_Projects_ReadOneOwned(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_one", map[string]any{"id": 1})
	require.NotContains(t, result, "isError")

	text := toolResultText(t, result)
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &project))
	assert.InDelta(t, float64(1), project["id"], 0.0001)
	assert.Equal(t, "Test1", project["title"])
}

func TestMCP_Projects_ReadOneForbidden(t *testing.T) {
	// Project 20 belongs to user 13, so user 1's read must be denied.
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_one", map[string]any{"id": 20})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for forbidden project, got: %v", result)
}

func TestMCP_Projects_ReadOneNonexistent(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_one", map[string]any{"id": 999999})
	isErr, _ := result["isError"].(bool)
	require.True(t, isErr, "expected isError for nonexistent project, got: %v", result)
}

func TestMCP_Projects_ReadAll(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_all", map[string]any{})
	require.NotContains(t, result, "isError", "read_all errored: %v", result)

	var projects []map[string]any
	readAllItems(t, result, &projects)
	require.NotEmpty(t, projects, "expected at least one project")

	// User 1 owns Test1 (project id 1); confirm it's in the response.
	titles := make(map[string]bool, len(projects))
	for _, p := range projects {
		title, _ := p["title"].(string)
		titles[title] = true
	}
	assert.True(t, titles["Test1"], "expected Test1 in: %v", titles)
}

func TestMCP_Projects_ReadAllSearch(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)
	result := c.callTool("projects_read_all", map[string]any{
		"search":   "Test1",
		"page":     1,
		"per_page": 50,
	})
	require.NotContains(t, result, "isError")

	var projects []map[string]any
	readAllItems(t, result, &projects)
	// At minimum the matching project Test1 should appear.
	require.NotEmpty(t, projects)
	for _, p := range projects {
		title, _ := p["title"].(string)
		assert.NotEmpty(t, title, "project missing title: %v", p)
	}
}

func TestMCP_Projects_Update(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	// Create rather than reuse a fixture: project 1 is referenced from a lot of test data.
	createResult := c.callTool("projects_create", map[string]any{
		"title": "mcp project to update",
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))

	updateResult := c.callTool("projects_update", map[string]any{
		"id":          pid,
		"title":       "mcp project updated",
		"description": "Updated description",
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	require.NotContains(t, readResult, "isError")
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &project))
	assert.Equal(t, "mcp project updated", project["title"])
	assert.Equal(t, "Updated description", project["description"])
}

// An explicitly sent `is_archived: false` must un-archive, not read as "omitted".
func TestMCP_Projects_UpdateClearsArchived(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("projects_create", map[string]any{
		"title":       "mcp project to un-archive",
		"is_archived": true,
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))
	require.True(t, created["is_archived"].(bool), "project should have been created archived")

	updateResult := c.callTool("projects_update", map[string]any{
		"id":          pid,
		"is_archived": false,
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	require.NotContains(t, readResult, "isError")
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &project))
	assert.False(t, project["is_archived"].(bool), "is_archived must be false after explicit clear")
	assert.Equal(t, "mcp project to un-archive", project["title"], "a partial update must not blank the title")
}

// The model updates a fixed column list, so a partial payload has to be merged
// onto the stored row or every omitted column is wiped.
func TestMCP_Projects_UpdateKeepsOmittedFields(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("projects_create", map[string]any{
		"title":      "mcp project with all fields",
		"identifier": "MCPKEEP",
		"hex_color":  "ff8800",
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))

	updateResult := c.callTool("projects_update", map[string]any{
		"id":          pid,
		"description": "only the description changes",
	})
	require.NotContains(t, updateResult, "isError", "update errored: %v", updateResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	require.NotContains(t, readResult, "isError")
	var project map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, readResult)), &project))
	assert.Equal(t, "only the description changes", project["description"])
	assert.Equal(t, "mcp project with all fields", project["title"])
	assert.Equal(t, "MCPKEEP", project["identifier"])
	assert.Equal(t, "ff8800", project["hex_color"])
}

func TestMCP_Projects_Delete(t *testing.T) {
	c := newMCPClient(t, mcpFullProjectsToken)

	createResult := c.callTool("projects_create", map[string]any{
		"title": "mcp project to delete",
	})
	require.NotContains(t, createResult, "isError")
	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(toolResultText(t, createResult)), &created))
	pid := int64(created["id"].(float64))

	deleteResult := c.callTool("projects_delete", map[string]any{"id": pid})
	require.NotContains(t, deleteResult, "isError", "delete errored: %v", deleteResult)

	readResult := c.callTool("projects_read_one", map[string]any{"id": pid})
	isErr, _ := readResult["isError"].(bool)
	require.True(t, isErr, "expected isError for deleted project, got: %v", readResult)
	text := strings.ToLower(toolResultText(t, readResult))
	assert.NotEmpty(t, text)
}
