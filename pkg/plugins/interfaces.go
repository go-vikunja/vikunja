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

package plugins

import (
	"github.com/labstack/echo/v4"
	"src.techknowlogick.com/xormigrate"
)

// Plugin is the base interface all plugins need to implement.
type Plugin interface {
	Name() string
	Version() string
	Init() error
	Shutdown() error
}

// MigrationPlugin lets a plugin provide database migrations.
type MigrationPlugin interface {
	Plugin
	Migrations() []*xormigrate.Migration
}

// AuthenticatedRouterPlugin lets a plugin register authenticated web handlers and routes.
type AuthenticatedRouterPlugin interface {
	Plugin
	RegisterAuthenticatedRoutes(g *echo.Group)
}

// UnauthenticatedRouterPlugin lets a plugin register unauthenticated web handlers and routes.
type UnauthenticatedRouterPlugin interface {
	Plugin
	RegisterUnauthenticatedRoutes(g *echo.Group)
}
