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

	"code.vikunja.io/api/pkg/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func rawToolHandler(name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := callTool(ctx, name, req.Params.Arguments)
		if err != nil {
			//nolint:nilerr // Domain errors use MCP tool results.
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil
		}
		body, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("mcp: marshal %s result: %w", name, err)
		}
		res := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(body)}}}
		if _, isObject := result.(map[string]any); isObject {
			res.StructuredContent = result
		}
		return res, nil
	}
}

func installTools(srv *mcp.Server, token *models.APIToken) {
	for _, t := range snapshotTools() {
		if t.tier != TierTyped || !t.authorized(token) {
			continue
		}
		srv.AddTool(&mcp.Tool{
			Name:        t.name,
			Description: t.description,
			InputSchema: t.spec.schema,
		}, rawToolHandler(t.name))
	}
	installCatalogTools(srv)
}
