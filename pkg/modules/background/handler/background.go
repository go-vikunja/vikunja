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

package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/background"
	"code.vikunja.io/api/pkg/modules/background/unsplash"
	"code.vikunja.io/web"
	"code.vikunja.io/web/handler"
	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
)

// BackgroundProvider represents a thing which holds a background provider
// Lets us get a new fresh provider every time we need one.
type BackgroundProvider struct {
	Provider func() background.Provider
}

// SearchBackgrounds is the web handler to search for backgrounds
func (bp *BackgroundProvider) SearchBackgrounds(c echo.Context) error {
	p := bp.Provider()

	err := c.Bind(p)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided: "+err.Error())
	}

	search := c.QueryParam("s")
	var page int64 = 1
	pg := c.QueryParam("p")
	if pg != "" {
		page, err = strconv.ParseInt(pg, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid page number: "+err.Error())
		}
	}

	s := db.NewSession()
	defer s.Close()

	result, err := p.Search(s, search, page)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "An error occurred: "+err.Error())
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "An error occurred: "+err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// This function does all kinds of preparations for setting and uploading a background
func (bp *BackgroundProvider) setBackgroundPreparations(s *xorm.Session, c echo.Context) (list *models.List, auth web.Auth, err error) {
	auth, err = auth2.GetAuthFromClaims(c)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid auth token: "+err.Error())
	}

	listID, err := strconv.ParseInt(c.Param("list"), 10, 64)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid list ID: "+err.Error())
	}

	// Check if the user has the right to change the list background
	list = &models.List{ID: listID}
	can, err := list.CanUpdate(s, auth)
	if err != nil {
		return
	}
	if !can {
		log.Infof("Tried to update list background of list %d while not having the rights for it (User: %v)", listID, auth)
		return list, auth, models.ErrGenericForbidden{}
	}
	// Load the list
	list, err = models.GetListSimpleByID(s, list.ID)
	return
}

// SetBackground sets an Image as list background
func (bp *BackgroundProvider) SetBackground(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	list, auth, err := bp.setBackgroundPreparations(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	p := bp.Provider()

	image := &background.Image{}
	err = c.Bind(image)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided: "+err.Error())
	}

	err = p.Set(s, image, list, auth)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}
	return c.JSON(http.StatusOK, list)
}

// UploadBackground uploads a background and passes the id of the uploaded file as an Image to the Set function of the BackgroundProvider.
func (bp *BackgroundProvider) UploadBackground(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	list, auth, err := bp.setBackgroundPreparations(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	p := bp.Provider()

	// Get + upload the image
	file, err := c.FormFile("background")
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
		_ = s.Rollback()
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Uploaded file is no image."})
	}
	_, _ = src.Seek(0, io.SeekStart)

	// Save the file
	f, err := files.CreateWithMime(src, file.Filename, uint64(file.Size), auth, mime.String())
	if err != nil {
		_ = s.Rollback()
		if files.IsErrFileIsTooLarge(err) {
			return echo.ErrBadRequest
		}

		return handler.HandleHTTPError(err, c)
	}

	image := &background.Image{ID: strconv.FormatInt(f.ID, 10)}

	err = p.Set(s, image, list, auth)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, list)
}

func checkListBackgroundRights(s *xorm.Session, c echo.Context) (list *models.List, auth web.Auth, err error) {
	auth, err = auth2.GetAuthFromClaims(c)
	if err != nil {
		return nil, auth, echo.NewHTTPError(http.StatusBadRequest, "Invalid auth token: "+err.Error())
	}

	listID, err := strconv.ParseInt(c.Param("list"), 10, 64)
	if err != nil {
		return nil, auth, echo.NewHTTPError(http.StatusBadRequest, "Invalid list ID: "+err.Error())
	}

	// Check if a background for this list exists + Rights
	list = &models.List{ID: listID}
	can, _, err := list.CanRead(s, auth)
	if err != nil {
		_ = s.Rollback()
		return nil, auth, handler.HandleHTTPError(err, c)
	}
	if !can {
		_ = s.Rollback()
		log.Infof("Tried to get list background of list %d while not having the rights for it (User: %v)", listID, auth)
		return nil, auth, echo.NewHTTPError(http.StatusForbidden)
	}

	return
}

// GetListBackground serves a previously set background from a list
// It has no knowledge of the provider that was responsible for setting the background.
// @Summary Get the list background
// @Description Get the list background of a specific list. **Returns json on error.**
// @tags list
// @Produce octet-stream
// @Param id path int true "List ID"
// @Security JWTKeyAuth
// @Success 200 {} string "The list background file."
// @Failure 403 {object} models.Message "No access to this list."
// @Failure 404 {object} models.Message "The list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/background [get]
func GetListBackground(c echo.Context) error {

	s := db.NewSession()
	defer s.Close()

	list, _, err := checkListBackgroundRights(s, c)
	if err != nil {
		return err
	}

	if list.BackgroundFileID == 0 {
		_ = s.Rollback()
		return echo.NotFoundHandler(c)
	}

	// Get the file
	bgFile := &files.File{
		ID: list.BackgroundFileID,
	}
	if err := bgFile.LoadFileByID(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	// Unsplash requires pingbacks as per their api usage guidelines.
	// To do this in a privacy-preserving manner, we do the ping from inside of Vikunja to not expose any user details.
	// FIXME: This should use an event once we have events
	unsplash.Pingback(s, bgFile)

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	// Serve the file
	return c.Stream(http.StatusOK, "image/jpg", bgFile.File)
}

// RemoveListBackground removes a list background, no matter the background provider
// @Summary Remove a list background
// @Description Removes a previously set list background, regardless of the list provider used to set the background. It does not throw an error if the list does not have a background.
// @tags list
// @Produce json
// @Param id path int true "List ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.List "The list"
// @Failure 403 {object} models.Message "No access to this list."
// @Failure 404 {object} models.Message "The list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/background [delete]
func RemoveListBackground(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	list, auth, err := checkListBackgroundRights(s, c)
	if err != nil {
		return err
	}

	list.BackgroundFileID = 0
	list.BackgroundInformation = nil
	err = models.UpdateList(s, list, auth, true)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, list)
}
