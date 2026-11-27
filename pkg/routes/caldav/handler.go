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

package caldav

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
)

func getBasicAuthUserFromContext(c *echo.Context) (*user.User, error) {
	u, is := c.Get("userBasicAuth").(*user.User)
	if !is {
		return &user.User{}, fmt.Errorf("user is not user element, is %s", reflect.TypeOf(c.Get("userBasicAuth")))
	}
	return u, nil
}

func EntryHandler(c *echo.Context) error { return handle(c, resolveEntry, serveMethod) }

func PrincipalHandler(c *echo.Context) error { return handle(c, resolvePrincipal, serveMethod) }

func ProjectHandler(c *echo.Context) error { return handle(c, resolveProjects, serveMethod) }

func TaskHandler(c *echo.Context) error { return handle(c, resolveTask, serveMethod) }

func EntryMethodNotAllowed(c *echo.Context) error {
	return handle(c, resolveEntry, methodNotAllowed)
}

func PrincipalMethodNotAllowed(c *echo.Context) error {
	return handle(c, resolvePrincipal, methodNotAllowed)
}

func ProjectMethodNotAllowed(c *echo.Context) error {
	return handle(c, resolveProjects, methodNotAllowed)
}

func TaskMethodNotAllowed(c *echo.Context) error {
	return handle(c, resolveTask, methodNotAllowed)
}

func handle(c *echo.Context, resolve targetResolver, serve func(*echo.Context, *user.User, target) error) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return writeError(c, err)
	}

	t, err := resolve(c, u)
	if err != nil {
		return writeError(c, err)
	}
	return serve(c, u, t)
}

func serveMethod(c *echo.Context, u *user.User, t target) error {
	switch c.Request().Method {
	case http.MethodOptions:
		return serveOptions(c, u, t)
	case http.MethodGet, http.MethodHead:
		return serveGet(c, u, t)
	case http.MethodDelete:
		return serveDelete(c, u, t)
	case http.MethodPut:
		return servePut(c, u, t)
	case "PROPFIND":
		return servePropfind(c, u, t)
	case "REPORT":
		return serveReport(c, u, t)
	case "PROPPATCH":
		switch t.(type) {
		case collectionTarget, taskTarget:
			return servePropPatch(c, u, t)
		}
	}
	return methodNotAllowed(c, u, t)
}

func readBody(c *echo.Context, limit int) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, int64(limit)+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errMalformedBody, err)
	}
	if len(body) > limit {
		return nil, errBodyTooLarge
	}
	logRequest(c.Request().Header, body)
	return body, nil
}

const maxLoggedBodySize = 4096

func logRequest(header http.Header, body []byte) {
	if len(body) > maxLoggedBodySize {
		body = body[:maxLoggedBodySize]
	}
	log.Debugf("[CALDAV] Request Body: %s", body)
	log.Debugf("[CALDAV] Request Headers: %v", log.SafeHeaders(header))
}
