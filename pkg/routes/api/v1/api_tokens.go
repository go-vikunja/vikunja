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
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// RegisterAPITokens registers all API token routes
func RegisterAPITokens(a *echo.Group) {
	a.GET("/tokens", handler.WithDBAndUser(getAllAPITokensLogic, false))
	a.PUT("/tokens", handler.WithDBAndUser(createAPITokenLogic, true))
	a.DELETE("/tokens/:token", handler.WithDBAndUser(deleteAPITokenLogic, true))
}

// getAllAPITokensLogic handles retrieving all API tokens for the current user
//
// @Summary Get all api tokens of the current user
// @Description Returns all api tokens the current user has created.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "The page number, used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of tokens per page. This parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tokens by their title."
// @Success 200 {array} models.APIToken "The list of all tokens"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /tokens [get]
func getAllAPITokensLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Extract query parameters for search and pagination
	search := c.QueryParam("s")
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		page = 0
	}
	perPage, err := strconv.Atoi(c.QueryParam("per_page"))
	if err != nil {
		perPage = 50 // Default items per page
	}

	apiTokenService := services.NewAPITokenService(s.Engine())
	tokens, _, totalCount, err := apiTokenService.GetAll(s, u, search, page, perPage)
	if err != nil {
		return err
	}

	// Set pagination header
	if totalCount > 0 {
		c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalCount, 10))
	}

	return c.JSON(http.StatusOK, tokens)
}

// createAPITokenLogic creates a new API token
//
// @Summary Create a new api token
// @Description Create a new api token to use on behalf of the user creating it.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param token body models.APIToken true "The token object with required fields"
// @Success 200 {object} models.APIToken "The created token."
// @Failure 400 {object} web.HTTPError "Invalid token object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tokens [put]
func createAPITokenLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	token := new(models.APIToken)
	if err := c.Bind(token); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token object provided.").SetInternal(err)
	}

	if err := c.Validate(token); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	apiTokenService := services.NewAPITokenService(s.Engine())
	if err := apiTokenService.Create(s, token, u); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, token)
}

// deleteAPITokenLogic handles deleting an API token
//
// @Summary Deletes an existing api token
// @Description Delete any of the user's api tokens.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param token path int true "Token ID"
// @Success 200 {object} models.Message "Successfully deleted."
// @Failure 404 {object} web.HTTPError "The token does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tokens/{token} [delete]
func deleteAPITokenLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	tokenID, err := strconv.ParseInt(c.Param("token"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token ID").SetInternal(err)
	}

	apiTokenService := services.NewAPITokenService(s.Engine())
	if err := apiTokenService.Delete(s, tokenID, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Successfully deleted."})
}
