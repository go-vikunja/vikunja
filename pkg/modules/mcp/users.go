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
	"errors"
	"fmt"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Scope names follow CollectRoutesForAPITokenUsage: single-segment GET /users
// files under "other", the project variant under "projects".
const (
	toolUsersSearch = "users_search"

	scopeGroupUsers      = "other"
	scopePermUsers       = "users"
	scopeGroupProjects   = "projects"
	scopePermUsersInProj = "users_search"
)

type usersSearchArgs struct {
	Query     string `json:"query"`
	ProjectID int64  `json:"project_id"`
}

var usersSearchSpec = func() *opSpec {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"query":      {Type: "string", Description: "Username, display name or full email address to look up. Name and email only match users who made themselves discoverable."},
			"project_id": {Type: "integer", Description: "Restrict the search to users who already have access to this project (useful for picking assignees). Requires read access to the project."},
		},
		Required:             []string{"query"},
		AdditionalProperties: falseSchema(),
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		panic(fmt.Sprintf("mcp: resolve %s schema: %v", toolUsersSearch, err))
	}
	return &opSpec{schema: schema, resolved: resolved}
}()

func installUsersSearchTool(srv *mcp.Server, token *models.APIToken) {
	if !token.HasPermission(scopeGroupUsers, scopePermUsers) {
		return
	}
	srv.AddTool(&mcp.Tool{
		Name:        toolUsersSearch,
		Description: "Find users by username, name or email — e.g. to resolve the user id needed by tasks_assignees_create, or the username needed by projects_users_create. Email addresses are never returned.",
		InputSchema: usersSearchSpec.schema,
	}, usersSearchHandler)
}

func usersSearchHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	users, err := searchUsers(ctx, req.Params.Arguments)
	if err != nil {
		//nolint:nilerr // IsError tool result, not a JSON-RPC protocol error
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, nil
	}
	body, err := json.Marshal(users)
	if err != nil {
		return nil, fmt.Errorf("mcp: marshal %s result: %w", toolUsersSearch, err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
	}, nil
}

func searchUsers(ctx context.Context, rawArgs json.RawMessage) ([]*user.User, error) {
	token := TokenFromContext(ctx)
	if !token.HasPermission(scopeGroupUsers, scopePermUsers) {
		return nil, fmt.Errorf("%w: %s", ErrScopeDenied, toolUsersSearch)
	}
	u := UserFromContext(ctx)
	if u == nil {
		return nil, ErrNoUserInContext
	}

	if _, err := validateAndDecodeArgs(usersSearchSpec, rawArgs); err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolUsersSearch, err)
	}
	var in usersSearchArgs
	if err := json.Unmarshal(rawArgs, &in); err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", toolUsersSearch, err)
	}
	if in.ProjectID != 0 && !token.HasPermission(scopeGroupProjects, scopePermUsersInProj) {
		return nil, fmt.Errorf("%w: %s with project_id", ErrScopeDenied, toolUsersSearch)
	}

	s := db.NewSession()
	defer s.Close()

	if in.ProjectID == 0 {
		return user.SearchUsers(s, in.Query, u)
	}

	found, canRead, err := models.SearchUsersForProject(s, &models.Project{ID: in.ProjectID}, u, u, in.Query)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, errors.New("forbidden: no read access to the project")
	}
	// ListUsers keeps the email when the query matched it exactly; the global
	// branch strips it via user.SearchUsers, so mirror that here.
	for _, f := range found {
		f.Email = ""
	}
	return found, nil
}
