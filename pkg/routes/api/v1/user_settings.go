// Vikunja is a to-do list application to facilitate your life.
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
	"github.com/labstack/echo/v4"
	"net/http"
)

// UserAvatarProvider holds the user avatar provider type
type UserAvatarProvider struct {
	AvatarProvider string `json:"avatar_provider"`
}

// GetUserAvatarProvider returns the currently set user avatar
// @Summary Return user avatar setting
// @Description Returns the current user's avatar setting.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} UserAvatarProvider
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/avatar [get]
func GetUserAvatarProvider(c echo.Context) error {

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	user, err := user2.GetUserWithEmail(u)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	uap := &UserAvatarProvider{AvatarProvider: user.AvatarProvider}
	return c.JSON(http.StatusOK, uap)
}

// ChangeUserAvatarProvider changes the user's avatar provider
// @Summary Set the user's avatar
// @Description Changes the user avatar. Valid types are gravatar (uses the user email), upload, initials, default.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param avatar body UserAvatarProvider true "The user's avatar setting"
// @Success 200 {object} UserAvatarProvider
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/avatar [post]
func ChangeUserAvatarProvider(c echo.Context) error {

	uap := &UserAvatarProvider{}
	err := c.Bind(uap)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad avatar type provided.")
	}

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	user, err := user2.GetUserWithEmail(u)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	user.AvatarProvider = uap.AvatarProvider

	_, err = user2.UpdateUser(user)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "Avatar was changed successfully."})
}
