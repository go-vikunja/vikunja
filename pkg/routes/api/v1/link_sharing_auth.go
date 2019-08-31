//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"net/http"
)

// AuthenticateLinkShare gives a jwt auth token for valid share hashes
// @Summary Get an auth token for a share
// @Description Get a jwt auth token for a shared list from a share hash.
// @tags sharing
// @Accept json
// @Produce json
// @Param share path string true "The share hash"
// @Success 200 {object} v1.Token "The valid jwt auth token."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid link share object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /shares/{share}/auth [post]
func AuthenticateLinkShare(c echo.Context) error {
	hash := c.Param("share")
	share, err := models.GetLinkShareByHash(hash)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	t, err := NewLinkShareJWTAuthtoken(share)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, Token{Token: t})
}
