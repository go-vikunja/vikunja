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
	_ "embed"

	"code.vikunja.io/api/pkg/models"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed filter_syntax.md
var filterSyntax string

const (
	toolFilterSyntax = "filter_syntax"
	filterSyntaxURI  = "vikunja://docs/filter-syntax"
)

const filterSyntaxDescription = "Reference for the filter query syntax used by task list, project view and saved filter tools: fields, operators, date math and examples."

var filterErrorCodes = map[int]bool{
	models.ErrCodeInvalidFilterExpression:       true,
	models.ErrCodeInvalidTaskFilterValue:        true,
	models.ErrCodeInvalidTaskFilterConcatinator: true,
	models.ErrCodeInvalidTaskFilterComparator:   true,
	models.ErrCodeInvalidTaskField:              true,
}

const filterErrorHint = "call " + toolFilterSyntax + " for the filter syntax"

// Exposed as a tool too: many clients never hand resources to the model.
func installDocs(srv *mcp.Server) {
	srv.AddTool(&mcp.Tool{
		Name:        toolFilterSyntax,
		Description: filterSyntaxDescription,
		InputSchema: &jsonschema.Schema{Type: "object", AdditionalProperties: falseSchema()},
	}, func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: filterSyntax}}}, nil
	})
	srv.AddResource(&mcp.Resource{
		URI:         filterSyntaxURI,
		Name:        toolFilterSyntax,
		Description: filterSyntaxDescription,
		MIMEType:    "text/markdown",
	}, func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
			URI:      filterSyntaxURI,
			MIMEType: "text/markdown",
			Text:     filterSyntax,
		}}}, nil
	})
}
