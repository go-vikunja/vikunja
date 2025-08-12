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

package handler

import (
	"bytes"
	_ "image/gif"  // To make sure the decoder used for generating blurHashes recognizes gifs
	_ "image/jpeg" // To make sure the decoder used for generating blurHashes recognizes jpgs
	_ "image/png"  // To make sure the decoder used for generating blurHashes recognizes pngs

	"github.com/disintegration/imaging"

	_ "golang.org/x/image/bmp"  // To make sure the decoder used for generating blurHashes recognizes bmps
	_ "golang.org/x/image/tiff" // To make sure the decoder used for generating blurHashes recognizes tiffs
	_ "golang.org/x/image/webp" // To make sure the decoder used for generating blurHashes recognizes tiffs

	"image"
	"io"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/background"
	"code.vikunja.io/api/pkg/modules/background/unsplash"
	"code.vikunja.io/api/pkg/modules/background/upload"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/bbrks/go-blurhash"
	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
	"golang.org/x/image/draw"
	"xorm.io/xorm"
)

var allowedImageMimes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/bmp",
	"image/tiff",
	"image/webp",
}

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
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided: "+err.Error()).SetInternal(err)
	}

	search := c.QueryParam("s")
	var page int64 = 1
	pg := c.QueryParam("p")
	if pg != "" {
		page, err = strconv.ParseInt(pg, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid page number: "+err.Error()).SetInternal(err)
		}
	}

	s := db.NewSession()
	defer s.Close()

	result, err := p.Search(s, search, page)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "An error occurred: "+err.Error()).SetInternal(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "An error occurred: "+err.Error()).SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}

// This function does all kinds of preparations for setting and uploading a background
func (bp *BackgroundProvider) setBackgroundPreparations(s *xorm.Session, c echo.Context) (project *models.Project, auth web.Auth, err error) {
	auth, err = auth2.GetAuthFromClaims(c)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid auth token: "+err.Error()).SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID: "+err.Error()).SetInternal(err)
	}

	// Check if the user has the permission to change the project background
	project = &models.Project{ID: projectID}
	can, err := project.CanUpdate(s, auth)
	if err != nil {
		return
	}
	if !can {
		log.Infof("Tried to update project background of project %d while not having the permissions for it (User: %v)", projectID, auth)
		return project, auth, models.ErrGenericForbidden{}
	}
	// Load the project
	project, err = models.GetProjectSimpleByID(s, project.ID)
	return
}

// SetBackground sets an Image as project background
func (bp *BackgroundProvider) SetBackground(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	project, auth, err := bp.setBackgroundPreparations(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	p := bp.Provider()

	image := &background.Image{}
	err = c.Bind(image)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided: "+err.Error()).SetInternal(err)
	}

	err = p.Set(s, image, project, auth)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = project.ReadOne(s, auth)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, project)
}

func CreateBlurHash(srcf io.Reader) (hash string, err error) {
	src, _, err := image.Decode(srcf)
	if err != nil {
		return "", err
	}

	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	return blurhash.Encode(4, 3, dst)
}

// UploadBackground uploads a background and passes the id of the uploaded file as an Image to the Set function of the BackgroundProvider.
func (bp *BackgroundProvider) UploadBackground(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	project, auth, err := bp.setBackgroundPreparations(s, c)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	// Get + upload the image
	file, err := c.FormFile("background")
	if err != nil {
		_ = s.Rollback()
		return err
	}
	srcf, err := file.Open()
	if err != nil {
		_ = s.Rollback()
		return err
	}
	defer srcf.Close()

	// Validate we're dealing with an image
	mime, err := mimetype.DetectReader(srcf)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}
	if !strings.HasPrefix(mime.String(), "image") {
		_ = s.Rollback()
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Uploaded file is no image."})
	}
	supported := false
	for _, m := range allowedImageMimes {
		if mime.Is(m) {
			supported = true
			break
		}
	}
	if !supported {
		_ = s.Rollback()
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Unsupported image format. Allowed: " + strings.Join(allowedImageMimes, ",")})
	}

	err = SaveBackgroundFile(s, auth, project, srcf, file.Filename, uint64(file.Size))
	if err != nil {
		_ = s.Rollback()
		if files.IsErrFileIsTooLarge(err) {
			return echo.ErrBadRequest
		}
		if IsErrFileUnsupportedImageFormat(err) {
			return c.JSON(http.StatusBadRequest, models.Message{Message: "Unsupported image format. Allowed: " + strings.Join(allowedImageMimes, ",")})
		}

		return handler.HandleHTTPError(err)
	}

	err = project.ReadOne(s, auth)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, project)
}

func SaveBackgroundFile(s *xorm.Session, auth web.Auth, project *models.Project, srcf io.ReadSeeker, filename string, filesize uint64) (err error) {
	mime, _ := mimetype.DetectReader(srcf)
	_, _ = srcf.Seek(0, io.SeekStart)
	src, err := imaging.Decode(srcf)
	if err != nil {
		if strings.Contains(err.Error(), "unknown format") {
			return ErrFileUnsupportedImageFormat{Mime: mime.String()}
		}
		return err
	}

	_, _ = srcf.Seek(0, io.SeekStart)
	imgConfig, _, err := image.DecodeConfig(srcf)
	if err != nil {
		return err
	}

	height := imgConfig.Height
	if imgConfig.Height > background.MaxBackgroundImageHeight {
		height = background.MaxBackgroundImageHeight
	}

	buf := bytes.Buffer{}
	dst := imaging.Resize(src, 0, height, imaging.Lanczos)
	err = imaging.Encode(&buf, dst, imaging.JPEG, imaging.JPEGQuality(80))
	if err != nil {
		return err
	}

	f, err := files.Create(&buf, filename, filesize, auth)
	if err != nil {
		return err
	}

	// Generate a blurHash
	_, _ = srcf.Seek(0, io.SeekStart)
	project.BackgroundBlurHash, err = CreateBlurHash(srcf)
	if err != nil {
		return err
	}

	// Save it
	p := upload.Provider{}
	img := &background.Image{ID: strconv.FormatInt(f.ID, 10)}
	err = p.Set(s, img, project, auth)
	return err
}

func checkProjectBackgroundRights(s *xorm.Session, c echo.Context) (project *models.Project, auth web.Auth, err error) {
	auth, err = auth2.GetAuthFromClaims(c)
	if err != nil {
		return nil, auth, echo.NewHTTPError(http.StatusBadRequest, "Invalid auth token: "+err.Error()).SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return nil, auth, echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID: "+err.Error()).SetInternal(err)
	}

	// Check if a background for this project exists + Permissions
	project = &models.Project{ID: projectID}
	can, _, err := project.CanRead(s, auth)
	if err != nil {
		_ = s.Rollback()
		return nil, auth, handler.HandleHTTPError(err)
	}
	if !can {
		_ = s.Rollback()
		log.Infof("Tried to get project background of project %d while not having the permissions for it (User: %v)", projectID, auth)
		return nil, auth, echo.NewHTTPError(http.StatusForbidden)
	}

	return
}

// GetProjectBackground serves a previously set background from a project
// It has no knowledge of the provider that was responsible for setting the background.
// @Summary Get the project background
// @Description Get the project background of a specific project. **Returns json on error.**
// @tags project
// @Produce octet-stream
// @Param id path int true "Project ID"
// @Security JWTKeyAuth
// @Success 200 {file} blob "The project background file."
// @Failure 403 {object} models.Message "No access to this project."
// @Failure 404 {object} models.Message "The project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/background [get]
func GetProjectBackground(c echo.Context) error {

	s := db.NewSession()
	defer s.Close()

	project, _, err := checkProjectBackgroundRights(s, c)
	if err != nil {
		return err
	}

	if project.BackgroundFileID == 0 {
		_ = s.Rollback()
		return echo.NotFoundHandler(c)
	}

	// Get the file
	bgFile := &files.File{
		ID: project.BackgroundFileID,
	}
	if err := bgFile.LoadFileByID(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}
	stat, err := bgFile.File.Stat()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	// Unsplash requires pingbacks as per their api usage guidelines.
	// To do this in a privacy-preserving manner, we do the ping from inside of Vikunja to not expose any user details.
	// FIXME: This should use an event once we have events
	unsplash.Pingback(s, bgFile)

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	// Set Last-Modified header if we have the file stat, so clients can decide whether to use cached files
	if stat != nil {
		c.Response().Header().Set(echo.HeaderLastModified, stat.ModTime().UTC().Format(http.TimeFormat))
	}

	// Serve the file
	return c.Stream(http.StatusOK, "image/jpg", bgFile.File)
}

// RemoveProjectBackground removes a project background, no matter the background provider
// @Summary Remove a project background
// @Description Removes a previously set project background, regardless of the project provider used to set the background. It does not throw an error if the project does not have a background.
// @tags project
// @Produce json
// @Param id path int true "Project ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Project "The project"
// @Failure 403 {object} models.Message "No access to this project."
// @Failure 404 {object} models.Message "The project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/background [delete]
func RemoveProjectBackground(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	project, auth, err := checkProjectBackgroundRights(s, c)
	if err != nil {
		return err
	}

	err = project.DeleteBackgroundFileIfExists()
	if err != nil {
		return err
	}

	project.BackgroundFileID = 0
	project.BackgroundInformation = nil
	project.BackgroundBlurHash = ""
	err = models.UpdateProject(s, project, auth, true)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, project)
}
