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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/ldap"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/modules/avatar/openid"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
)

// GetAvatar returns a user's avatar
// @Summary User Avatar
// @Description Returns the user avatar as image.
// @tags user
// @Produce octet-stream
// @Param username path string true "The username of the user who's avatar you want to get"
// @Param size query int false "The size of the avatar you want to get. If bigger than the max configured size this will be adjusted to the maximum size."
// @Success 200 {file} blob "The avatar"
// @Failure 404 {object} models.Message "The user does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /{username}/avatar [get]
func GetAvatar(c echo.Context) error {
	// Get the username
	username := c.Param("username")

	s := db.NewSession()
	defer s.Close()

	// Get the user
	u, err := user.GetUserWithEmail(s, &user.User{Username: username})
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		log.Errorf("Error getting user for avatar: %v", err)
		return handler.HandleHTTPError(err)
	}

	found := err == nil || !user.IsErrUserDoesNotExist(err)

	var avatarProvider avatar.Provider
	switch u.AvatarProvider {
	case "gravatar":
		avatarProvider = &gravatar.Provider{}
	case "initials":
		avatarProvider = &initials.Provider{}
	case "upload":
		avatarProvider = &upload.Provider{}
	case "marble":
		avatarProvider = &marble.Provider{}
	case "ldap":
		avatarProvider = &ldap.Provider{}
	case "openid":
		avatarProvider = &openid.Provider{}
	default:
		avatarProvider = &empty.Provider{}
	}

	if !found {
		avatarProvider = &empty.Provider{}
	}

	size := c.QueryParam("size")
	var sizeInt int64 = 250 // Default size of 250
	if size != "" {
		sizeInt, err = strconv.ParseInt(size, 10, 64)
		if err != nil {
			log.Errorf("Error parsing size: %v", err)
			return handler.HandleHTTPError(err)
		}
	}
	if sizeInt > config.ServiceMaxAvatarSize.GetInt64() {
		sizeInt = config.ServiceMaxAvatarSize.GetInt64()
	}

	// Get the avatar
	a, mimeType, err := avatarProvider.GetAvatar(u, sizeInt)
	if err != nil {
		log.Errorf("Error getting avatar for user %d: %v", u.ID, err)
		return handler.HandleHTTPError(err)
	}

	return c.Blob(http.StatusOK, mimeType, a)
}

// UploadAvatar uploads and sets a user avatar
// @Summary Upload a user avatar
// @Description Upload a user avatar. This will also set the user's avatar provider to "upload"
// @tags user
// @Accept mpfd
// @Produce json
// @Param avatar formData string true "The avatar as single file."
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "The avatar was set successfully."
// @Failure 400 {object} models.Message "File is no image."
// @Failure 403 {object} models.Message "File too large."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/settings/avatar/upload [put]
func UploadAvatar(c echo.Context) (err error) {

	s := db.NewSession()
	defer s.Close()

	uc, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	u, err := user.GetUserByID(s, uc.ID)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	// Get + upload the image
	file, err := c.FormFile("avatar")
	if err != nil {
		_ = s.Rollback()
		return err
	}
	src, err := file.Open()
	if err != nil {
		_ = s.Rollback()
		return err
	}
	defer src.Close()

	// Validate we're dealing with an image
	mime, err := mimetype.DetectReader(src)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}
	if !strings.HasPrefix(mime.String(), "image") {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Uploaded file is no image."})
	}
	_, _ = src.Seek(0, io.SeekStart)

	u.AvatarProvider = "upload"
	err = upload.StoreAvatarFile(s, u, src)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	avatar.FlushAllCaches(u)

	return c.JSON(http.StatusOK, models.Message{Message: "Avatar was uploaded successfully."})
}
