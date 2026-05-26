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
// inside the existing authenticated /api/v1 group. The actual MCP protocol
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

// routePrefix is the URL prefix the MCP endpoint is mounted under. The
// SDK handler does not care about path — it dispatches on HTTP method
// alone — so this is only used to strip the prefix before forwarding so
// the underlying http.Request looks like it was routed to "/".
const routePrefix = "/api/v1/mcp"

// newServer constructs a fresh *mcp.Server with Vikunja's implementation
// metadata and the per-session set of registered tools. The SDK calls the
// factory passed to NewStreamableHTTPHandler exactly once per session
// (when no Mcp-Session-Id matches an existing session, i.e. at the
// initialize request); the returned *mcp.Server is cached for the
// lifetime of that session.
//
// Per-token tool filtering happens here: we pull the API token from the
// request context (placed there by the Echo entry handler in Handler) and
// register only the tools the token's scopes authorise. tools/list then
// returns the filtered subset naturally; tools/call is additionally
// re-checked in the dispatcher.
//
// RegisterResources is idempotent and called here so production startup
// doesn't need to know about a separate init step — the first incoming
// MCP request triggers registration on demand.
func newServer(req *http.Request) *mcp.Server {
	RegisterResources()
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "vikunja",
		Version: version.Version,
	}, nil)
	// The token may legitimately be nil if a future code path forgets to
	// attach one — installToolsForToken treats that as "no tools allowed".
	// In the production flow Handler rejects unauthenticated requests
	// before reaching the SDK, so this is purely defensive.
	token := TokenFromContext(req.Context())
	installToolsForToken(srv, token)
	return srv
}

// streamableHandler is package-level so the SDK can manage its internal
// session map across requests. The factory returns a fresh *mcp.Server
// per session, scoped to the requesting token's permissions.
var streamableHandler = mcp.NewStreamableHTTPHandler(newServer, nil)

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

	u, ok := c.Get("api_user").(*user.User)
	if !ok || u == nil {
		log.Errorf("[mcp] api_user missing from context for token %d", token.ID)
		return echo.NewHTTPError(http.StatusInternalServerError, "missing user in context")
	}

	req := c.Request()
	ctx := WithUser(req.Context(), u)
	ctx = WithToken(ctx, token)
	req = req.WithContext(ctx)

	// Strip the mount prefix before forwarding. The SDK's ServeHTTP
	// dispatches on req.Method, not req.URL.Path, so this is mostly
	// cosmetic — but it keeps the request looking the way the SDK's own
	// tests/examples expect (requests served at "/").
	http.StripPrefix(routePrefix, streamableHandler).ServeHTTP(c.Response(), req)
	return nil
}
