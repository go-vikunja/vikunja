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
	"encoding/json"
	"fmt"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"github.com/mark3labs/mcp-go/mcp"
)

func (c *McpHandler) ReadAllTool() mcp.Tool {
	name := c.getTypeName()

	return mcp.Tool{
		Name:        fmt.Sprintf("get_%ss", name),
		Description: fmt.Sprintf("Get all %ss the authenticated user has access to", name),
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"limit":  map[string]any{"type": "integer", "description": "Maximum number of results"},
				"offset": map[string]any{"type": "integer", "description": "Number of results to skip"},
			},
		},
	}

}

func (c *McpHandler) ReadAllMCP(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currentStruct := c.EmptyStruct()

	u, err := c.getUser(request)
	if err != nil {
		return nil, err
	}

	var filter struct {
		Limit  *int `json:"limit"`
		Offset *int `json:"offset"`
	}
	paramsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("could not marshal request: %v", err)), nil
	}

	if err = json.Unmarshal(paramsBytes, &filter); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	perPage := 50
	if filter.Limit != nil && *filter.Limit > 0 && *filter.Limit < 50 {
		perPage = *filter.Limit
	}
	page := 1
	if filter.Offset != nil && *filter.Offset > 0 {
		page = *filter.Offset/perPage + 1
	}

	// Create the db session
	s := db.NewSession()
	defer func() {
		err = s.Close()
		if err != nil {
			log.Errorf("Could not close session: %s", err)
		}
	}()

	result, count, total, err := currentStruct.ReadAll(s, u, "", page, perPage)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	list := List[interface{}]{
		Result: result,
		Cont:   count,
		Total:  total,
	}

	return mcp.NewToolResultStructured(list, toJSON(list)), nil
}
