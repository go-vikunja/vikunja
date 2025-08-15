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
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// GetCaldavTokens is the handler to return a project of all caldav tokens for the current user
func GetCaldavTokens(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	tokens, err := user.GetCaldavTokens(u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	v2Tokens := make([]*v2.CaldavToken, len(tokens))
	for i, t := range tokens {
		v2Tokens[i] = &v2.CaldavToken{
			ID:      t.ID,
			Created: t.Created,
			Links: &v2.CaldavTokenLinks{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/user/caldav-tokens/%d", t.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Tokens)
}

// GenerateCaldavToken is the handler to create a caldav token
func GenerateCaldavToken(c echo.Context) (err error) {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	token, err := user.GenerateNewCaldavToken(u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	v2Token := &v2.CaldavToken{
		ID:      token.ID,
		Token:   token.Token,
		Created: token.Created,
		Links: &v2.CaldavTokenLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/user/caldav-tokens/%d", token.ID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2Token)
}

// DeleteCaldavToken is the handler to delete a caldv token
func DeleteCaldavToken(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}
	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = user.DeleteCaldavTokenByID(u, id)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
