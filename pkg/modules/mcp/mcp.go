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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"
	"code.vikunja.io/api/pkg/version"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	routeSuffix = "/mcp"
	RoutePrefix = "/api/v2" + routeSuffix
)

const (
	maxRequestBytes       = 4 << 20
	maxMessagesPerRequest = 20
)

// Register must follow apiv2.RegisterAll so tools pick up the AutoPatch operations.
func Register(api huma.API, group *echo.Group, groupPrefix string) {
	initTools(api, groupPrefix)
	group.POST(routeSuffix, handler)
}

func newServerForRequest(req *http.Request) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "vikunja",
		Version: version.Version,
	}, nil)
	addToolsAuthorizedBy(srv, tokenFrom(echoContextFrom(req.Context())))
	return srv
}

func echoContextFrom(ctx context.Context) *echo.Context {
	ec, _ := ctx.Value(humabridge.EchoContextKey).(*echo.Context)
	return ec
}
func tokenFrom(ec *echo.Context) *models.APIToken {
	if ec == nil {
		return nil
	}
	token, _ := ec.Get("api_token").(*models.APIToken)
	return token
}

func addToolsAuthorizedBy(srv *mcp.Server, token *models.APIToken) {
	var catalog []*tool
	for _, t := range snapshotTools() {
		switch {
		case !t.authorized(token):
			continue
		case !t.typed:
			catalog = append(catalog, t)
			continue
		}
		srv.AddTool(&mcp.Tool{
			Name:        t.name,
			Description: t.description,
			InputSchema: t.spec.schema,
		}, rawToolHandler(t.name))
	}
	installCatalogTools(srv, catalog)
}

func rawToolHandler(name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := callTool(ctx, name, req.Params.Arguments)
		if err != nil {
			//nolint:nilerr // Domain errors use MCP tool results.
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil
		}
		body, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("mcp: marshal %s result: %w", name, err)
		}
		res := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(body)}}}
		if _, isObject := result.(map[string]any); isObject {
			res.StructuredContent = result
		}
		return res, nil
	}
}

var streamableHandler = sync.OnceValue(newStreamableHandler)

// Built on first use because the CORS config is not loaded at package init.
var originProtection = sync.OnceValue(newCrossOriginProtection)

// Stateless builds a server per request, so tools/list is filtered by the caller's token; localhost protection would reject deployments behind a loopback reverse proxy.
func newStreamableHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(newServerForRequest, &mcp.StreamableHTTPOptions{
		Stateless:                  true,
		DisableLocalhostProtection: true,
	})
}

func newCrossOriginProtection() *http.CrossOriginProtection {
	protection := http.NewCrossOriginProtection()
	if !config.CorsEnable.GetBool() {
		return protection
	}
	for _, origin := range config.CorsOrigins.GetStringSlice() {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			log.Debugf("[mcp] not trusting cors origin %q: %s", origin, err)
		}
	}
	return protection
}

// handler rejects JWTs, which bypass API-token route scopes.
func handler(c *echo.Context) error {
	req := c.Request()
	// MCP is not a browser transport: an Origin a browser would not send to itself is rejected before the token is looked at.
	if err := originProtection().Check(req); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}
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
	if proceed, err := limitRequestBody(c, req); !proceed {
		return err
	}
	http.StripPrefix(RoutePrefix, streamableHandler()).ServeHTTP(c.Response(), req)
	return nil
}

// Written instead of returned: error_handler.go rewrites a returned 413 into the generic "file is too large" error.
func limitRequestBody(c *echo.Context, req *http.Request) (proceed bool, err error) {
	req.Body = http.MaxBytesReader(c.Response(), req.Body, maxRequestBytes)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return false, c.JSON(http.StatusRequestEntityTooLarge, map[string]string{
				"message": fmt.Sprintf("MCP request body must not exceed %d bytes", maxRequestBytes),
			})
		}
		return false, echo.NewHTTPError(http.StatusBadRequest, "could not read request body")
	}
	if count, isBatch := countBatchMessages(body); isBatch && count > maxMessagesPerRequest {
		return false, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("MCP requests are limited to %d batched messages", maxMessagesPerRequest))
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	return true, nil
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
