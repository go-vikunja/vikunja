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
// Vikunja's CRUD API to MCP-aware clients (Claude Desktop, Cursor, etc.).
//
// The entry point is Handler, which is mounted by the routes package
// inside the authenticated /api/v2 group. The actual MCP protocol
// (JSON-RPC framing, session management, SSE streaming) is delegated to
// github.com/modelcontextprotocol/go-sdk.
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

// RoutePrefix is the absolute path the MCP endpoint is mounted under. Shared
// with the routes package so the mount point and the token middleware's
// route-check exemption can't drift apart.
const RoutePrefix = "/api/v2/mcp"

// newServer builds a server whose tool set is filtered to what the calling
// request's token authorises. Stateless mode calls this per request, so the
// filtering always reflects the token on the wire rather than whichever token
// happened to open the session.
func newServer(req *http.Request) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "vikunja",
		Version: version.Version,
	}, nil)
	installToolsForToken(srv, TokenFromContext(req.Context()))
	return srv
}

// Stateless keeps identity per-request: a cached session would otherwise pin
// every later message to the context (and token) of whoever initialized it.
// DisableLocalhostProtection turns off the SDK's DNS-rebinding guard, which
// would 403 the standard reverse-proxy-to-127.0.0.1 deployment.
var streamableHandler = mcp.NewStreamableHTTPHandler(newServer, &mcp.StreamableHTTPOptions{
	Stateless:                  true,
	DisableLocalhostProtection: true,
})

// Handler is the Echo entry point for the MCP endpoint. It:
//
//  1. Rejects JWT-authed requests with 401 — MCP is token-only because
//     JWT bypasses CanDoAPIRoute (and therefore the mcp:access scope).
//  2. Pulls the API token from the Echo context and rejects with 403 if
//     it does not have the mcp:access scope.
//  3. Attaches the authenticated user and token to r.Context() via the
//     typed keys in context.go so tool handlers can pull them out
//     without depending on Echo.
//  4. Forwards to the SDK's streamable-HTTP handler with the route
//     prefix stripped.
func Handler(c *echo.Context) error {
	// JWT-authed requests have a *jwt.Token under "user" and do not have
	// "api_token" set. The token middleware only populates "api_token"
	// when it successfully resolves a Bearer tk_… header.
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
