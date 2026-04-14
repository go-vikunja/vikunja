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

func (c *McpHandler) ReadOneTool() mcp.Tool {
	name := c.getTypeName()
	idx := c.getIndex()

	var idxName = "id"
	var idxType = "integer"
	if idx != nil {
		idxName = idx.Tag.Get("json")
		idxType = c.goToMCPType(idx.Type)
	}

	return mcp.Tool{
		Name:        fmt.Sprintf("get_%s", name),
		Description: fmt.Sprintf("Get a %s the authenticated user has access to", name),
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				idxName: map[string]any{"type": idxType, "description": fmt.Sprintf("The %s of the %s to get", idxName, name)},
			},
		},
	}
}

func (c *McpHandler) ReadOneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currentStruct := c.EmptyStruct()

	u, err := c.getUser(request)
	if err != nil {
		return nil, err
	}

	paramsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("could not marshal request: %v", err)), nil
	}

	if err = json.Unmarshal(paramsBytes, currentStruct); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	s := db.NewSession()
	defer func() {
		err = s.Close()
		if err != nil {
			log.Errorf("Could not close session: %s", err)
		}
	}()

	canRead, _, err := currentStruct.CanRead(s, u)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	if !canRead {
		_ = s.Rollback()
		events.CleanupPending(s)
		log.Warningf("Tried to read while not having the permissions for it (User: %v)", u)
		return mcp.NewToolResultError("You don't have the permission to see this"), nil
	}

	err = currentStruct.ReadOne(s, u)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	err = s.Commit()
	if err != nil {
		events.CleanupPending(s)
		return nil, err
	}

	events.DispatchPending(s)

	return mcp.NewToolResultStructured(currentStruct, toJSON(currentStruct)), nil
}
