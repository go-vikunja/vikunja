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

// Package mcp implements the streamable-HTTP MCP endpoint that exposes
// Vikunja's CRUD API to MCP-aware clients. Protocol framing, sessions and SSE
// streaming are delegated to github.com/modelcontextprotocol/go-sdk.
package mcp

import (
	"net/http"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/version"

	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RoutePrefix is shared with the routes package so the mount point and the token middleware's exemption can't drift apart.
const RoutePrefix = "/api/v2/mcp"

func newServer(req *http.Request) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "vikunja",
		Version: version.Version,
	}, nil)
	installToolsForToken(srv, TokenFromContext(req.Context()))
	return srv
}

// Stateless keeps identity per-request: a cached session would pin every later
// message to the token that initialized it. DisableLocalhostProtection drops the
// DNS-rebinding guard, which would 403 reverse-proxy-to-127.0.0.1 deployments.
var streamableHandler = mcp.NewStreamableHTTPHandler(newServer, &mcp.StreamableHTTPOptions{
	Stateless:                  true,
	DisableLocalhostProtection: true,
})

// Handler is token-only: JWT auth bypasses CanDoAPIRoute and therefore the mcp:access scope.
func Handler(c *echo.Context) error {
	// The token middleware only sets "api_token" for a resolved Bearer tk_… header.
	tokenAny := c.Get("api_token")
	if tokenAny == nil {
		log.Debugf("[mcp] rejecting non-API-token request to %s", c.Request().URL.Path)
		return echo.NewHTTPError(http.StatusUnauthorized, "MCP requires an API token")
	}

	token, ok := tokenAny.(*models.APIToken)
	if !ok || token == nil {
		log.Errorf("[mcp] api_token in context has unexpected type %T", tokenAny)
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid token in context")
	}

	if !token.HasMCPAccess() {
		log.Debugf("[mcp] API token %d does not have mcp:access scope", token.ID)
		return echo.NewHTTPError(http.StatusForbidden, "token does not have mcp:access scope")
	}

	u, err := user.GetCurrentUser(c)
	if err != nil {
		log.Errorf("[mcp] no user in context for token %d: %v", token.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "missing user in context")
	}

	req := c.Request()
	ctx := WithUser(req.Context(), u)
	ctx = WithToken(ctx, token)
	req = req.WithContext(ctx)

	http.StripPrefix(RoutePrefix, streamableHandler).ServeHTTP(c.Response(), req)
	return nil
}
