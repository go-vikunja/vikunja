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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

// UploadTaskAttachment handles everything needed for the upload of a task attachment
func UploadTaskAttachment(c echo.Context) error {
	var taskAttachment models.TaskAttachment
	if err := c.Bind(&taskAttachment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No task ID provided").SetInternal(err)
	}

	// Permissions check
	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	s := db.NewSession()
	defer s.Close()

	can, err := taskAttachment.CanCreate(s, aut)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) {
			return echo.NewHTTPError(http.StatusBadRequest, "No multipart form provided")
		}
		return handler.HandleHTTPError(err)
	}

	type result struct {
		Errors  []*echo.HTTPError        `json:"errors"`
		Success []*v2.TaskAttachment `json:"success"`
	}
	r := &result{}
	fileHeaders := form.File["files"]
	for _, file := range fileHeaders {
		// We create a new attachment object here to have a clean start
		ta := &models.TaskAttachment{
			TaskID: taskAttachment.TaskID,
		}

		f, err := file.Open()
		if err != nil {
			r.Errors = append(r.Errors, handler.HandleHTTPError(err))
			continue
		}
		defer f.Close()

		err = ta.NewAttachment(s, f, file.Filename, uint64(file.Size), aut)
		if err != nil {
			r.Errors = append(r.Errors, handler.HandleHTTPError(err))
			continue
		}

		v2Attachment := &v2.TaskAttachment{
			ID:      ta.ID,
			TaskID:  ta.TaskID,
			FileID:  ta.FileID,
			Created: ta.Created,
			Links: &v2.TaskAttachmentLinks{
				Self: &v2.Link{Href: fmt.Sprintf("/api/v2/tasks/%d/attachments/%d", ta.TaskID, ta.ID)},
				Task: &v2.Link{Href: fmt.Sprintf("/api/v2/tasks/%d", ta.TaskID)},
				File: &v2.Link{Href: fmt.Sprintf("/api/v2/files/%d", ta.FileID)},
			},
		}
		r.Success = append(r.Success, v2Attachment)
	}

	return c.JSON(http.StatusOK, r)
}

// GetTaskAttachment returns a task attachment to download for the user
func GetTaskAttachment(c echo.Context) error {
	var taskAttachment models.TaskAttachment
	if err := c.Bind(&taskAttachment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No task ID provided").SetInternal(err)
	}

	// Permissions check
	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	s := db.NewSession()
	defer s.Close()

	can, _, err := taskAttachment.CanRead(s, aut)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if !can {
		return echo.ErrForbidden
	}

	// Get the attachment incl file
	err = taskAttachment.ReadOne(s, aut)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// If the preview query parameter is set, get the preview (cached or generate)
	previewSize := models.GetPreviewSizeFromString(c.QueryParam("preview_size"))
	if previewSize != models.PreviewSizeUnknown && strings.HasPrefix(taskAttachment.File.Mime, "image") {
		previewFileBytes := taskAttachment.GetPreview(previewSize)
		if previewFileBytes != nil {
			return c.Blob(http.StatusOK, "image/png", previewFileBytes)
		}
	}

	// Open and send the file to the client
	err = taskAttachment.File.LoadFileByID()
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	http.ServeContent(c.Response(), c.Request(), taskAttachment.File.Name, taskAttachment.File.Created, taskAttachment.File.File)
	return nil
}
