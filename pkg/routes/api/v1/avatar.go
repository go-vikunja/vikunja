// Vikunja is a to-do-list application to facilitate your life.
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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// GetAvatar returns a user's avatar
// @Summary User Avatar
// @Description Returns the user avatar as image.
// @tags user
// @Produce octet-stream
// @Param username path string true "The username of the user who's avatar you want to get"
// @Param size query int false "The size of the avatar you want to get"
// @Success 200 {} blob "The avatar"
// @Failure 404 {object} models.Message "The user does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /{username}/avatar [get]
func GetAvatar(c echo.Context) error {
	// Get the username
	username := c.Param("username")

	// Get the user
	user, err := user2.GetUserWithEmail(&user2.User{Username: username})
	if err != nil {
		log.Errorf("Error getting user for avatar: %v", err)
		return handler.HandleHTTPError(err, c)
	}

	// Initialize the avatar provider
	// For now, we only have one avatar provider, in the future there could be multiple which
	// could be changed based on user settings etc.
	var avatarProvider avatar.Provider
	switch config.AvatarProvider.GetString() {
	case "gravatar":
		avatarProvider = &gravatar.Provider{}
	default:
		avatarProvider = &empty.Provider{}
	}

	size := c.QueryParam("size")
	var sizeInt int64 = 250 // Default size of 250
	if size != "" {
		sizeInt, err = strconv.ParseInt(size, 10, 64)
		if err != nil {
			log.Errorf("Error parsing size: %v", err)
			return handler.HandleHTTPError(err, c)
		}
	}

	// Get the avatar
	a, mimeType, err := avatarProvider.GetAvatar(user, sizeInt)
	if err != nil {
		log.Errorf("Error getting avatar for user %d: %v", user.ID, err)
		return handler.HandleHTTPError(err, c)
	}

	return c.Blob(http.StatusOK, mimeType, a)
}
