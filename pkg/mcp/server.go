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
	"code.vikunja.io/api/pkg/mcp/handler"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/version"

	"github.com/labstack/echo/v5"

	"github.com/mark3labs/mcp-go/server"
)

type ServerWrapper struct {
	httpServer *server.StreamableHTTPServer
	handler    *Server
}

func NewMCPServerWrapper(Config) *ServerWrapper {
	srv := server.NewMCPServer("Vikunja", version.Version)

	mcpHandler := NewMCPServer()

	taskHandler := &handler.McpHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Task{}
		},
	}
	srv.AddTool(taskHandler.CreateTool(), taskHandler.CreateHandler)
	srv.AddTool(taskHandler.ReadOneTool(), taskHandler.ReadOneHandler)
	srv.AddTool(taskHandler.UpdateTool(), taskHandler.UpdateHandler)
	srv.AddTool(taskHandler.DeleteTool(), taskHandler.DeleteHandler)

	projectHandler := &handler.McpHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Project{}
		},
	}
	srv.AddTool(projectHandler.CreateTool(), projectHandler.CreateHandler)
	srv.AddTool(projectHandler.ReadOneTool(), projectHandler.ReadOneHandler)
	srv.AddTool(projectHandler.ReadAllTool(), projectHandler.ReadAllMCP)
	srv.AddTool(projectHandler.UpdateTool(), projectHandler.UpdateHandler)
	srv.AddTool(projectHandler.DeleteTool(), projectHandler.DeleteHandler)

	httpServer := server.NewStreamableHTTPServer(srv)

	return &ServerWrapper{
		httpServer: httpServer,
		handler:    mcpHandler,
	}
}

func (w *ServerWrapper) HandleRequest(c *echo.Context) error {
	w.httpServer.ServeHTTP(c.Response(), c.Request())
	return nil
}
