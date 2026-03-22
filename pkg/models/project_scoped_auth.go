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

package models

import (
	"code.vikunja.io/api/pkg/web"
)

// ProjectScopedAuth wraps a web.Auth and adds project scope restrictions
// from a project-scoped API token. Models can type-assert to this type
// to check and enforce project scope.
type ProjectScopedAuth struct {
	Auth       web.Auth
	ProjectIDs []int64
}

// GetID implements the web.Auth interface by delegating to the wrapped auth.
func (p *ProjectScopedAuth) GetID() int64 {
	return p.Auth.GetID()
}

// UnwrapAuth implements web.AuthUnwrapper to allow type assertions on the inner auth.
func (p *ProjectScopedAuth) UnwrapAuth() web.Auth {
	return p.Auth
}

// GetProjectScope extracts project scope IDs from a web.Auth, if present.
// Returns nil if the auth is not project-scoped (meaning all projects are accessible).
func GetProjectScope(a web.Auth) []int64 {
	if scoped, ok := a.(*ProjectScopedAuth); ok {
		return scoped.ProjectIDs
	}
	return nil
}
