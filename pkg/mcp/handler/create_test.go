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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	config.InitDefaultConfig()
	i18n.Init()
	files.InitTests()
	user.InitTests()
	models.SetupTests()
	events.Fake()
	if err := db.LoadFixtures(); err != nil {
		log.Fatal(err)
	}
	m.Run()
}

func getResultText(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		return ""
	}
	return textContent.Text
}

func TestMcpHandler_CreateTool(t *testing.T) {
	t.Run("task handler", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		tool := h.CreateTool()
		assert.Equal(t, "create_task", tool.Name)
		assert.Contains(t, tool.Description, "task")
		assert.NotEmpty(t, tool.InputSchema.Properties)
	})

	t.Run("project handler", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.CreateTool()
		assert.Equal(t, "create_project", tool.Name)
		assert.Contains(t, tool.Description, "project")
		assert.NotEmpty(t, tool.InputSchema.Properties)
	})
}

func TestMcpHandler_CreateHandlerValidation(t *testing.T) {
	t.Run("create task without title returns error", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": float64(1),
				},
			},
		}
		result, err := h.CreateHandler(context.Background(), request)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create task without project_id returns error", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title": "Test Task",
				},
			},
		}
		result, err := h.CreateHandler(context.Background(), request)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create project without title returns error", func(t *testing.T) {
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
		result, err := h.CreateHandler(context.Background(), request)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMcpHandler_CreateToolProperties(t *testing.T) {
	h := &McpHandler{
		EmptyStruct: func() CObject {
			return &models.Task{}
		},
	}
	tool := h.CreateTool()
	props := tool.InputSchema.Properties
	assert.Contains(t, props, "title")
	assert.Contains(t, props, "project_id")
	assert.Contains(t, props, "description")
}
