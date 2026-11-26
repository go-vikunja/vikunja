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

package routes

import (
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/modules/auth/oauth2server"
	"code.vikunja.io/api/pkg/modules/mcp"

	"github.com/labstack/echo/v5"
)

// oauthTokenPathV2 takes no cookies: its credentials are in the body, so reading
// its responses from any origin exposes nothing a server-side client can't get.
const oauthTokenPathV2 = "/api/v2/oauth/token" //nolint:gosec // a route path, not a credential

// skipsInstanceCORS: these paths answer CORS themselves. Their Allow-Origin * must
// not meet the instance middleware's Allow-Credentials.
func skipsInstanceCORS(path string) bool {
	return path == oauthTokenPathV2 ||
		strings.HasPrefix(path, oauth2server.MetadataPath) ||
		strings.HasPrefix(path, mcp.ResourceMetadataPath)
}

func allowAnyOrigin(c *echo.Context) {
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
}

func allowAnyOriginMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		allowAnyOrigin(c)
		return next(c)
	}
}

func anyOriginPreflight(method string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		allowAnyOrigin(c)
		h := c.Response().Header()
		h.Set(echo.HeaderAccessControlAllowMethods, method+", OPTIONS")
		if requested := c.Request().Header.Get(echo.HeaderAccessControlRequestHeaders); requested != "" {
			h.Set(echo.HeaderAccessControlAllowHeaders, requested)
			h.Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
		}
		h.Set(echo.HeaderAccessControlMaxAge, "86400")
		return c.NoContent(http.StatusNoContent)
	}
}
