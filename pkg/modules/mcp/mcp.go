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

// Package mcp derives MCP tools from the v2 OpenAPI document and dispatches them through REST handlers.
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
	"code.vikunja.io/api/pkg/version"
	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const RoutePrefix = "/api/v2/mcp"
const (
	maxRequestBytes       = 4 << 20
	maxMessagesPerRequest = 20
)

func newServer(req *http.Request) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "vikunja",
		Version: version.Version,
	}, nil)
	installTools(srv, TokenFromContext(req.Context()))
	return srv
}

// Stateless prevents session IDs from carrying identity across requests.
// Localhost protection would reject deployments behind a loopback reverse proxy.
var streamableHandler = mcp.NewStreamableHTTPHandler(newServer, &mcp.StreamableHTTPOptions{
	Stateless:                  true,
	DisableLocalhostProtection: true,
})

// Handler rejects JWTs, which bypass API-token route scopes.
func Handler(c *echo.Context) error {
	tokenAny := c.Get("api_token")
	if tokenAny == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "MCP requires an API token")
	}
	token, ok := tokenAny.(*models.APIToken)
	if !ok || token == nil {
		log.Errorf("[mcp] api_token has unexpected type %T", tokenAny)
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid token in context")
	}
	if !token.HasMCPAccess() {
		return echo.NewHTTPError(http.StatusForbidden, "token does not have mcp:access scope")
	}
	req := c.Request()
	if err := limitRequestBody(c, req); err != nil {
		return err
	}
	ctx := WithCaller(WithToken(req.Context(), token), req)
	http.StripPrefix(RoutePrefix, streamableHandler).ServeHTTP(c.Response(), req.WithContext(ctx))
	return nil
}
func limitRequestBody(c *echo.Context, req *http.Request) error {
	req.Body = http.MaxBytesReader(c.Response(), req.Body, maxRequestBytes)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, fmt.Sprintf("MCP request body must not exceed %d bytes", maxRequestBytes))
		}
		return echo.NewHTTPError(http.StatusBadRequest, "could not read request body")
	}
	if count, isBatch := countBatchMessages(body); isBatch && count > maxMessagesPerRequest {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("MCP requests are limited to %d batched messages", maxMessagesPerRequest))
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	return nil
}
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
