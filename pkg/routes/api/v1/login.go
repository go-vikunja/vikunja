// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Token represents an authentification token
type Token struct {
	Token string `json:"token"`
}

// Login is the login handler
// @Summary Login
// @Description Logs a user in. Returns a JWT-Token to authenticate further requests.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body models.UserLogin true "The login credentials"
// @Success 200 {object} v1.Token
// @Failure 400 {object} models.Message "Invalid user password model."
// @Failure 403 {object} models.Message "Invalid username or password."
// @Router /login [post]
func Login(c echo.Context) error {
	u := user2.Login{}
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Please provide a username and password."})
	}

	// Check user
	user, err := user2.CheckUserCredentials(&u)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	// Create token
	t, err := NewUserJWTAuthtoken(user)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Token{Token: t})
}

// RenewToken gives a new token to every user with a valid token
// If the token is valid is checked in the middleware.
// @Summary Renew user token
// @Description Returns a new valid jwt user token with an extended length.
// @tags user
// @Accept json
// @Produce json
// @Success 200 {object} v1.Token
// @Failure 400 {object} models.Message "Only user token are available for renew."
// @Router /user/token [post]
func RenewToken(c echo.Context) error {

	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
	typ := int(claims["type"].(float64))
	if typ != AuthTypeUser {
		return echo.ErrBadRequest
	}

	user, err := user2.GetUserFromClaims(claims)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	// Create token
	t, err := NewUserJWTAuthtoken(user)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Token{Token: t})
}
