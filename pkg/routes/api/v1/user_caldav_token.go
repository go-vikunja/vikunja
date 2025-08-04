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
	"strconv"

	"code.vikunja.io/api/pkg/models"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// GenerateCaldavToken is the handler to create a caldav token
// @Summary Generate a caldav token
// @Description Generates a caldav token which can be used for the caldav api. It is not possible to see the token again after it was generated.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} user.Token
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/token/caldav [put]
func GenerateCaldavToken(c echo.Context) (err error) {

	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	token, err := user.GenerateNewCaldavToken(u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, token)
}

// GetCaldavTokens is the handler to return a project of all caldav tokens for the current user
// @Summary Returns the caldav tokens for the current user
// @Description Return the IDs and created dates of all caldav tokens for the current user.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} user.Token
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/token/caldav [get]
func GetCaldavTokens(c echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	tokens, err := user.GetCaldavTokens(u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, tokens)
}

// DeleteCaldavToken is the handler to delete a caldv token
// @Summary Delete a caldav token by id
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Token ID"
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/token/caldav/{id} [delete]
func DeleteCaldavToken(c echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = user.DeleteCaldavTokenByID(u, id)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "The token was deleted successfully."})
}
