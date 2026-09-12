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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

const (
	// Descriptions and comments can be long; dbtext caps a single field at 1 MiB.
	maxRequestBytes = 4 << 20
	// A JSON-RPC batch is one POST, hence one rate-limiter hit, no matter how many tool calls it carries.
	maxMessagesPerRequest = 20
)

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
	if err := limitRequestBody(c, req); err != nil {
		return err
	}

	ctx := WithUser(req.Context(), u)
	ctx = WithToken(ctx, token)
	req = req.WithContext(ctx)

	http.StripPrefix(RoutePrefix, streamableHandler).ServeHTTP(c.Response(), req)
	return nil
}

// limitRequestBody buffers the POST body so a batch can be counted before the
// SDK executes it, then hands the same bytes back to the SDK.
func limitRequestBody(c *echo.Context, req *http.Request) error {
	if req.Method != http.MethodPost {
		return nil
	}

	req.Body = http.MaxBytesReader(c.Response(), req.Body, maxRequestBytes)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, fmt.Sprintf("MCP request body must not exceed %d bytes", maxRequestBytes))
		}
		log.Debugf("[mcp] could not read request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "could not read request body")
	}

	if count, isBatch := countBatchMessages(body); isBatch && count > maxMessagesPerRequest {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("MCP requests are limited to %d batched messages", maxMessagesPerRequest))
	}

	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	return nil
}

// countBatchMessages walks the top level only; a batch of large tool calls must not be unmarshalled just to be counted.
// Anything that isn't a well-formed array is left for the SDK to reject.
func countBatchMessages(body []byte) (count int, isBatch bool) {
	dec := json.NewDecoder(bytes.NewReader(body))
	tok, err := dec.Token()
	if err != nil {
		return 0, false
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return 0, false
	}
	for dec.More() {
		count++
		if count > maxMessagesPerRequest {
			return count, true
		}
		var msg json.RawMessage
		if err := dec.Decode(&msg); err != nil {
			return 0, false
		}
	}
	return count, true
}
