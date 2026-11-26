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
	"strings"

	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/labstack/echo/v5/middleware"
)

// oauthTokenPathV2 takes no cookies, so any origin may read its responses.
const oauthTokenPathV2 = "/api/v2/oauth/token" //nolint:gosec // a route path, not a credential

// Allow-Origin * must not meet the instance middleware's Allow-Credentials.
func skipsInstanceCORS(path string) bool {
	return path == oauthTokenPathV2 || strings.HasPrefix(path, oauth2server.MetadataPath)
}

var anyOriginCORS = pathScoped(skipsInstanceCORS, middleware.CORSWithConfig(middleware.CORSConfig{
	AllowOrigins: []string{"*"},
	MaxAge:       86400,
}))
