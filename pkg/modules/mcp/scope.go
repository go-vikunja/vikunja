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
	"errors"
	"slices"

	"code.vikunja.io/api/pkg/models"
)

// ErrScopeDenied is returned by the dispatcher when the token attached to
// the call context does not have the (resource, op) scope required to invoke
// the tool. The AddTool wrapper renders this as an IsError tool result so
// the client sees a structured failure rather than a JSON-RPC protocol error.
var ErrScopeDenied = errors.New("mcp: tool not authorized for this token")

// tokenAuthorizes returns true iff the token's APIPermissions map contains
// op.Permission() under the given resource's scope group. This is the
// (group, permission) lookup that gates both tools/list visibility and
// tools/call invocation; it intentionally duplicates rather than shares
// CanDoAPIRoute's logic because MCP doesn't have a path/method to match —
// the registry already owns the (resource, op) → (group, permission) mapping.
//
// A nil token or nil APIPermissions returns false (slices.Contains on a nil
// slice is also false, so the second case is naturally handled). Defensive
// checks here keep the dispatcher's "fail closed" contract even if the entry
// handler somehow forgets to attach a token.
func tokenAuthorizes(token *models.APIToken, resourceName string, op Op) bool {
	if token == nil {
		return false
	}
	return slices.Contains(token.APIPermissions[resourceName], op.Permission())
}
