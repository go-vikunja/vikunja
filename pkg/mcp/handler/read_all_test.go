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

package handler

import (
	"context"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMcpHandler_ReadAllTool(t *testing.T) {
	t.Run("project handler returns get_projects tool", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.ReadAllTool()
		assert.Equal(t, "get_projects", tool.Name)
		assert.Contains(t, tool.Description, "projects")
		assert.NotEmpty(t, tool.InputSchema.Properties)
	})

	t.Run("tool has limit and offset properties", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.ReadAllTool()
		props := tool.InputSchema.Properties
		assert.Contains(t, props, "limit")
		assert.Contains(t, props, "offset")
	})

	t.Run("limit has correct type and description", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.ReadAllTool()
		limitProp := tool.InputSchema.Properties["limit"].(map[string]any)
		assert.Equal(t, "integer", limitProp["type"])
		assert.Contains(t, limitProp["description"], "Maximum")
	})

	t.Run("offset has correct type and description", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.ReadAllTool()
		offsetProp := tool.InputSchema.Properties["offset"].(map[string]any)
		assert.Equal(t, "integer", offsetProp["type"])
		assert.Contains(t, offsetProp["description"], "skip")
	})
}

func TestMcpHandler_ReadAllMCPValidation(t *testing.T) {
	t.Run("empty arguments should work", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}
		result, err := h.ReadAllMCP(context.Background(), request)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid arguments type should return error", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: "invalid",
			},
		}
		result, err := h.ReadAllMCP(context.Background(), request)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
