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

func TestMcpHandler_ReadOneTool(t *testing.T) {
	t.Run("task handler returns get_task tool", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		tool := h.ReadOneTool()
		assert.Equal(t, "get_task", tool.Name)
		assert.Contains(t, tool.Description, "task")
		assert.NotEmpty(t, tool.InputSchema.Properties)
	})

	t.Run("project handler returns get_project tool", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.ReadOneTool()
		assert.Equal(t, "get_project", tool.Name)
		assert.Contains(t, tool.Description, "project")
		assert.NotEmpty(t, tool.InputSchema.Properties)
	})

	t.Run("tool has id property", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		tool := h.ReadOneTool()
		props := tool.InputSchema.Properties
		assert.Contains(t, props, "id")
	})

	t.Run("id has correct type and description", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		tool := h.ReadOneTool()
		idProp := tool.InputSchema.Properties["id"].(map[string]any)
		assert.Equal(t, "integer", idProp["type"])
		assert.Contains(t, idProp["description"], "id")
		assert.Contains(t, idProp["description"], "task")
	})
}

func TestMcpHandler_ReadOneHandlerValidation(t *testing.T) {
	t.Run("empty arguments should return error", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}
		result, err := h.ReadOneHandler(context.Background(), request)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid id type should return error", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id": "not-an-integer",
				},
			},
		}
		result, err := h.ReadOneHandler(context.Background(), request)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid arguments type should return error", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: "invalid",
			},
		}
		result, err := h.ReadOneHandler(context.Background(), request)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
