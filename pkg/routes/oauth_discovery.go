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

	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/labstack/echo/v5"
)

// RFC 8414 §3.1 inserts the well-known path before the issuer's path, so a subpath
// deployment's proxy can forward that URL here unchanged.
func registerOAuthDiscoveryRoutes(e *echo.Echo) {
	paths := []string{oauth2server.MetadataPath}
	if issuerPath := oauth2server.IssuerPath(); issuerPath != "" {
		paths = append(paths, oauth2server.MetadataPath+issuerPath)
	}
	for _, p := range paths {
		e.GET(p, oauthAuthorizationServerMetadata)
	}
}

// oauthAuthorizationServerMetadata serves the RFC 8414 authorization server metadata.
func oauthAuthorizationServerMetadata(c *echo.Context) error {
	metadata, ok := oauth2server.Metadata()
	if !ok {
		// Written rather than returned: the static file handler answers a returned 404 with index.html.
		return c.NoContent(http.StatusNotFound)
	}
	c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(http.StatusOK, metadata)
}
