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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"

	"bytes"
	"image"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
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

	s := db.NewSession()
	defer s.Close()

	// Get the user
	u, err := user.GetUserWithEmail(s, &user.User{Username: username})
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		log.Errorf("Error getting user for avatar: %v", err)
		return handler.HandleHTTPError(err, c)
	}

	found := !(err != nil && user.IsErrUserDoesNotExist(err))

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
			return handler.HandleHTTPError(err, c)
		}
	}

	// Get the avatar
	a, mimeType, err := avatarProvider.GetAvatar(u, sizeInt)
	if err != nil {
		log.Errorf("Error getting avatar for user %d: %v", u.ID, err)
		return handler.HandleHTTPError(err, c)
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
		return handler.HandleHTTPError(err, c)
	}
	u, err := user.GetUserByID(s, uc.ID)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
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
		return handler.HandleHTTPError(err, c)
	}
	if !strings.HasPrefix(mime.String(), "image") {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Uploaded file is no image."})
	}
	_, _ = src.Seek(0, io.SeekStart)

	// Remove the old file if one exists
	if u.AvatarFileID != 0 {
		f := &files.File{ID: u.AvatarFileID}
		if err := f.Delete(); err != nil {
			if !files.IsErrFileDoesNotExist(err) {
				_ = s.Rollback()
				return handler.HandleHTTPError(err, c)
			}
		}
		u.AvatarFileID = 0
	}

	// Resize the new file to a max height of 1024
	img, _, err := image.Decode(src)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}
	resizedImg := imaging.Resize(img, 0, 1024, imaging.Lanczos)
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, resizedImg); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	upload.InvalidateCache(u)

	// Save the file
	f, err := files.CreateWithMime(buf, file.Filename, uint64(file.Size), u, "image/png")
	if err != nil {
		_ = s.Rollback()
		if files.IsErrFileIsTooLarge(err) {
			return echo.ErrBadRequest
		}

		return handler.HandleHTTPError(err, c)
	}

	u.AvatarFileID = f.ID
	u.AvatarProvider = "upload"

	if _, err := user.UpdateUser(s, u); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Avatar was uploaded successfully."})
}
