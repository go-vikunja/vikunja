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
	"fmt"

	"code.vikunja.io/api/pkg/log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ServerWrapper struct {
	httpServer *server.StreamableHTTPServer
	handler    *Server
}

func NewMCPServerWrapper(authToken string) *ServerWrapper {
	srv := server.NewMCPServer("Vikunja", "1.0.0")

	mcpHandler := NewMCPServer(authToken)

	srv.AddTool(mcp.Tool{
		Name:        "get_tasks",
		Description: "Get tasks from Vikunja with optional filters",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Filter by project ID"},
				"list_id":    map[string]any{"type": "integer", "description": "Filter by list ID"},
				"is_done":    map[string]any{"type": "boolean", "description": "Filter by completion status"},
				"limit":      map[string]any{"type": "integer", "description": "Maximum number of results"},
				"offset":     map[string]any{"type": "integer", "description": "Number of results to skip"},
				"search":     map[string]any{"type": "string", "description": "Search tasks by title"},
			},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleGetTasks(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "get_task",
		Description: "Get a single task by ID",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"id": map[string]any{"type": "integer", "description": "The task ID"},
			},
			Required: []string{"id"},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleGetTask(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task in Vikunja",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"title":       map[string]any{"type": "string", "description": "Task title"},
				"description": map[string]any{"type": "string", "description": "Task description"},
				"project_id":  map[string]any{"type": "integer", "description": "Project ID"},
				"due_date":    map[string]any{"type": "string", "description": "Due date in ISO 8601 format"},
				"priority":    map[string]any{"type": "integer", "description": "Task priority"},
			},
			Required: []string{"title", "project_id"},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleCreateTask(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "update_task",
		Description: "Update an existing task in Vikunja",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"id":          map[string]any{"type": "integer", "description": "The task ID"},
				"title":       map[string]any{"type": "string", "description": "Task title"},
				"description": map[string]any{"type": "string", "description": "Task description"},
				"done":        map[string]any{"type": "boolean", "description": "Whether task is done"},
				"due_date":    map[string]any{"type": "string", "description": "Due date in ISO 8601 format"},
				"priority":    map[string]any{"type": "integer", "description": "Task priority"},
			},
			Required: []string{"id"},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleUpdateTask(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "delete_task",
		Description: "Delete a task from Vikunja",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"id": map[string]any{"type": "integer", "description": "The task ID to delete"},
			},
			Required: []string{"id"},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleDeleteTask(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "get_projects",
		Description: "Get all projects the authenticated user has access to",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"limit":  map[string]any{"type": "integer", "description": "Maximum number of results"},
				"offset": map[string]any{"type": "integer", "description": "Number of results to skip"},
			},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleGetProjects(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "get_lists",
		Description: "Get all lists in a project",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "The project ID"},
				"limit":      map[string]any{"type": "integer", "description": "Maximum number of results"},
				"offset":     map[string]any{"type": "integer", "description": "Number of results to skip"},
			},
			Required: []string{"project_id"},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleGetLists(paramsBytes, token)
	})

	srv.AddTool(mcp.Tool{
		Name:        "get_kanban_board",
		Description: "Get kanban board data for a project including buckets and tasks",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "The project ID"},
			},
			Required: []string{"project_id"},
		},
	}, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := extractToken(request)
		paramsBytes, _ := json.Marshal(request.Params.Arguments)
		return mcpHandler.HandleGetKanbanBoard(paramsBytes, token)
	})

	httpServer := server.NewStreamableHTTPServer(srv)

	return &ServerWrapper{
		httpServer: httpServer,
		handler:    mcpHandler,
	}
}

func extractToken(request mcp.CallToolRequest) string {
	return request.Header.Get("Authorization")
}

func (w *ServerWrapper) RunHTTP(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Infof("MCP server listening on %s", addr)
	return w.httpServer.Start(addr)
}
