// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
)

// LinkShareToken represents a link share auth token with extra infos about the actual link share
type LinkShareToken struct {
	auth.Token
	*models.LinkSharing
	ListID int64 `json:"list_id"`
}

// AuthenticateLinkShare gives a jwt auth token for valid share hashes
// @Summary Get an auth token for a share
// @Description Get a jwt auth token for a shared list from a share hash.
// @tags sharing
// @Accept json
// @Produce json
// @Param share path string true "The share hash"
// @Success 200 {object} auth.Token "The valid jwt auth token."
// @Failure 400 {object} web.HTTPError "Invalid link share object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /shares/{share}/auth [post]
func AuthenticateLinkShare(c echo.Context) error {
	hash := c.Param("share")

	s := db.NewSession()
	defer s.Close()

	share, err := models.GetLinkShareByHash(s, hash)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	t, err := auth.NewLinkShareJWTAuthtoken(share)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, LinkShareToken{
		Token:       auth.Token{Token: t},
		LinkSharing: share,
		ListID:      share.ListID,
	})
}
