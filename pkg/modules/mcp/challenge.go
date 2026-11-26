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

package mcp

import (
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/labstack/echo/v5"
)

// bearerError is an RFC 6750 §3.1 error.
type bearerError struct {
	code        string
	description string
	scope       oauth2server.Scope
}

var (
	errInvalidToken = &bearerError{
		code:        "invalid_token",
		description: "The access token is expired, revoked or not valid for this resource",
	}
	errInsufficientScope = &bearerError{
		code:        "insufficient_scope",
		description: "The access token does not grant access to the MCP endpoint",
		scope:       oauth2server.ScopeMCPRead,
	}
)

// Challenge adds the RFC 9728 challenge to every 401 of the MCP endpoint written
// after it, whichever middleware or handler writes it.
func Challenge(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		res, err := echo.UnwrapResponse(c.Response())
		if c.Path() != RoutePrefix || err != nil {
			return next(c)
		}
		var bearerErr *bearerError
		// RFC 6750 §3.1: no error code without credentials or with another auth scheme.
		if strings.HasPrefix(strings.ToLower(c.Request().Header.Get(echo.HeaderAuthorization)), "bearer ") {
			bearerErr = errInvalidToken
		}
		res.Before(func() {
			if res.Status == http.StatusUnauthorized {
				setChallenge(res.Header(), bearerErr)
			}
		})
		return next(c)
	}
}

// setChallenge does nothing without an issuer: there is no metadata to point at.
func setChallenge(h http.Header, bearerErr *bearerError) {
	issuer, ok := oauth2server.Issuer()
	if !ok {
		return
	}
	// Under the issuer rather than path-inserted at the host root, so a subpath
	// deployment serves it without proxy rules; RFC 9728 §3.3 allows any URL here.
	params := []string{"resource_metadata=" + quote(issuer+ResourceMetadataPath+RoutePrefix)}
	if bearerErr != nil {
		params = append(params, "error="+quote(bearerErr.code))
		if bearerErr.scope != "" {
			params = append(params, "scope="+quote(string(bearerErr.scope)))
		}
		params = append(params, "error_description="+quote(bearerErr.description))
	}
	h.Set(echo.HeaderWWWAuthenticate, "Bearer "+strings.Join(params, ", "))
	// Browser-based clients on an origin the instance CORS policy allows must be able to read it.
	if h.Get(echo.HeaderAccessControlAllowOrigin) != "" {
		h.Add(echo.HeaderAccessControlExposeHeaders, echo.HeaderWWWAuthenticate)
	}
}

func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
