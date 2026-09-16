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

package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/routes/api/shared"

	"github.com/labstack/echo/v5"
)

// LinkShareAuth represents everything required to authenticate a link share
type LinkShareAuth struct {
	Hash     string `param:"share" json:"-"`
	Password string `json:"password"`
}

// AuthenticateLinkShare gives a jwt auth token for valid share hashes
// @Summary Get an auth token for a share
// @Description Get a jwt auth token for a shared project from a share hash.
// @tags sharing
// @Accept json
// @Produce json
// @Param password body v1.LinkShareAuth true "The password for link shares which require one."
// @Param share path string true "The share hash"
// @Success 200 {object} auth.Token "The valid jwt auth token."
// @Failure 400 {object} web.HTTPError "Invalid link share object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /shares/{share}/auth [post]
func AuthenticateLinkShare(c *echo.Context) error {
	sh := &LinkShareAuth{}
	if err := c.Bind(sh); err != nil {
		return err
	}

	token, err := shared.AuthenticateLinkShare(sh.Hash, sh.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}
