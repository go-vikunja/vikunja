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

package v2

import (
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// AuthenticateLinkShare gives a jwt auth token for valid share hashes
func AuthenticateLinkShare(c echo.Context) error {
	var sh v2.LinkShareAuth
	if err := c.Bind(&sh); err != nil {
		return handler.HandleHTTPError(err)
	}

	s := db.NewSession()
	defer s.Close()

	share, err := models.GetLinkShareByHash(s, c.Param("token"))
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if share.SharingType == models.SharingTypeWithPassword {
		err := models.VerifyLinkSharePassword(share, sh.Password)
		if err != nil {
			return handler.HandleHTTPError(err)
		}
	}

	t, err := auth.NewLinkShareJWTAuthtoken(share)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	share.Password = ""

	return c.JSON(http.StatusOK, &v2.LinkShareToken{
		Token:       t,
		LinkSharing: share,
		ProjectID:   share.ProjectID,
		Links: &v2.LinkShareTokenLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/shares/%s/auth", share.Hash),
			},
		},
	})
}
