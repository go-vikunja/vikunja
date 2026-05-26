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

package client

import (
	"context"
	"fmt"
)

// FullPermissions is the broadest set of API token scopes a veans bot needs:
// read+write on every resource it touches. Vikunja's permission map is
// `{resource: [actions]}` shaped; the keys here cover everything the CLI
// calls for normal operation.
//
// We over-grant intentionally — the bot needs to claim, comment, label,
// relate, and update tasks; revoking unused scopes after the fact is cheap.
func FullPermissions() map[string][]string {
	return map[string][]string{
		"tasks":            {"read_one", "read_all", "create", "update", "delete"},
		"projects":         {"read_one", "read_all", "create", "update", "delete"},
		"labels":           {"read_one", "read_all", "create", "update", "delete"},
		"task_comments":    {"read_one", "read_all", "create", "update", "delete"},
		"task_assignees":   {"create", "delete", "read_all"},
		"task_relations":   {"create", "delete"},
		"task_attachments": {"create", "read_one", "delete"},
		"buckets":          {"read_all", "create", "update", "delete"},
		"project_views":    {"read_one", "read_all"},
		"users":            {"read_all"},
	}
}

// CreateToken mints an API token. If t.OwnerID is non-zero, the token is
// minted FOR that user — the caller must be the bot's owner (i.e. created
// the bot in step 8 of init).
func (c *Client) CreateToken(ctx context.Context, t *APIToken) (*APIToken, error) {
	var out APIToken
	if err := c.Do(ctx, "PUT", "/tokens", nil, t, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTokens returns every API token the authenticated user can see.
func (c *Client) ListTokens(ctx context.Context) ([]*APIToken, error) {
	var out []*APIToken
	if err := c.Do(ctx, "GET", "/tokens", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteToken revokes a token by ID. Used by `veans login` rotation.
func (c *Client) DeleteToken(ctx context.Context, id int64) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/tokens/%d", id), nil, nil, nil)
}
