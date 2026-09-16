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

	"github.com/labstack/echo/v5"

	"code.vikunja.io/api/pkg/config"
)

// ChangePasswordRedirect implements https://www.w3.org/TR/change-password-url/
func ChangePasswordRedirect(c *echo.Context) error {
	publicURL := config.ServicePublicURL.GetString()
	if publicURL == "" {
		publicURL = "/"
	}
	return c.Redirect(http.StatusFound, publicURL+"user/settings/password-update")
}
