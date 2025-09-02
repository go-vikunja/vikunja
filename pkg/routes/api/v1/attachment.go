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
	"errors"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// AttachmentRoutes defines all task attachment related routes
var AttachmentRoutes = []APIRoute{
	{Method: "GET", Path: "/tasks/:task/attachments", Handler: handler.WithDBAndUser(getAllAttachmentsLogic, false), PermissionScope: "read_all"},
	{Method: "PUT", Path: "/tasks/:task/attachments", Handler: handler.WithDBAndUser(uploadAttachmentLogic, true), PermissionScope: "create"},
	{Method: "GET", Path: "/tasks/:task/attachments/:attachment", Handler: handler.WithDBAndUser(getAttachmentLogic, false), PermissionScope: "read_one"},
	{Method: "DELETE", Path: "/tasks/:task/attachments/:attachment", Handler: handler.WithDBAndUser(deleteAttachmentLogic, true), PermissionScope: "delete"},
}

// RegisterAttachments registers all attachment routes
func RegisterAttachments(a *echo.Group) {
	registerRoutes(a, AttachmentRoutes)
}

// @Summary Get all attachments for a task
// @Description Returns all attachments for a specific task with pagination support.
// @tags task-attachments
// @Accept json
// @Produce json
// @Param task path int true "Task ID"
// @Param page query int false "Page number for pagination" default(1)
// @Param per_page query int false "Number of items per page" default(50)
// @Security JWTKeyAuth
// @Success 200 {array} models.TaskAttachment "The list of attachments"
// @Failure 400 {object} web.HTTPError "Invalid task ID"
// @Failure 403 {object} web.HTTPError "No access to task"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/attachments [get]
func getAllAttachmentsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	attachmentService := services.NewAttachmentService(s.Engine())

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 {
		perPage = 50
	}

	attachments, resultCount, totalItems, err := attachmentService.GetAllForTask(s, taskID, u, page, perPage)
	if err != nil {
		return err
	}

	// Set pagination headers
	if totalItems > 0 {
		totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)
		c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalPages, 10))
		c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	}

	return c.JSON(http.StatusOK, attachments)
}

// @Summary Upload new attachments to a task
// @Description Uploads one or more files as attachments to a specific task. Supports multiple file upload.
// @tags task-attachments
// @Accept multipart/form-data
// @Produce json
// @Param task path int true "Task ID"
// @Param files formData file true "Files to upload"
// @Security JWTKeyAuth
// @Success 200 {object} object "Upload results with success and error arrays"
// @Failure 400 {object} web.HTTPError "Invalid task ID or no multipart form"
// @Failure 403 {object} web.HTTPError "No access to task"
// @Failure 413 {object} models.Message "File too large"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/attachments [put]
func uploadAttachmentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
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
		Success []*models.TaskAttachment `json:"success"`
	}
	r := &result{}

	attachmentService := services.NewAttachmentService(s.Engine())
	fileHeaders := form.File["files"]
	for _, file := range fileHeaders {
		// We create a new attachment object here to have a clean start
		attachment := &models.TaskAttachment{
			TaskID: taskID,
		}

		f, err := file.Open()
		if err != nil {
			r.Errors = append(r.Errors, handler.HandleHTTPError(err))
			continue
		}
		defer f.Close()

		createdAttachment, err := attachmentService.Create(s, attachment, f, file.Filename, uint64(file.Size), u)
		if err != nil {
			r.Errors = append(r.Errors, handler.HandleHTTPError(err))
			continue
		}
		r.Success = append(r.Success, createdAttachment)
	}

	return c.JSON(http.StatusOK, r)
}

// @Summary Download or preview an attachment
// @Description Downloads an attachment file or returns a preview image if requested and the file is an image.
// @tags task-attachments
// @Produce application/octet-stream
// @Produce image/png
// @Param task path int true "Task ID"
// @Param attachment path int true "Attachment ID"
// @Param preview_size query string false "Preview size for images" Enums(sm, md, lg, xl)
// @Security JWTKeyAuth
// @Success 200 {file} file "The attachment file or preview image"
// @Failure 400 {object} web.HTTPError "Invalid task or attachment ID"
// @Failure 403 {object} web.HTTPError "No access to task"
// @Failure 404 {object} web.HTTPError "Attachment not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/attachments/{attachment} [get]
func getAttachmentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	attachmentID, err := strconv.ParseInt(c.Param("attachment"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid attachment ID").SetInternal(err)
	}

	attachmentService := services.NewAttachmentService(s.Engine())

	// Check if preview is requested
	previewSize := models.GetPreviewSizeFromString(c.QueryParam("preview_size"))
	if previewSize != models.PreviewSizeUnknown {
		// Get attachment first to check if it's an image
		attachment, err := attachmentService.GetByID(s, attachmentID, taskID, u)
		if err != nil {
			return err
		}

		if attachment.File != nil && strings.HasPrefix(attachment.File.Mime, "image") {
			previewBytes, err := attachmentService.GetPreview(s, attachmentID, taskID, previewSize, u)
			if err != nil {
				return err
			}
			if previewBytes != nil {
				return c.Blob(http.StatusOK, "image/png", previewBytes)
			}
		}
	}

	// Get the attachment for download
	attachment, err := attachmentService.GetByID(s, attachmentID, taskID, u)
	if err != nil {
		return err
	}

	// Load the file content for serving
	err = attachment.File.LoadFileByID()
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	http.ServeContent(c.Response(), c.Request(), attachment.File.Name, attachment.File.Created, attachment.File.File)
	return nil
}

// @Summary Delete an attachment
// @Description Removes an attachment from a task and deletes the underlying file.
// @tags task-attachments
// @Produce json
// @Param task path int true "Task ID"
// @Param attachment path int true "Attachment ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "Attachment deleted successfully"
// @Failure 400 {object} web.HTTPError "Invalid task or attachment ID"
// @Failure 403 {object} web.HTTPError "No access to task"
// @Failure 404 {object} web.HTTPError "Attachment not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/attachments/{attachment} [delete]
func deleteAttachmentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	attachmentID, err := strconv.ParseInt(c.Param("attachment"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid attachment ID").SetInternal(err)
	}

	attachmentService := services.NewAttachmentService(s.Engine())
	err = attachmentService.Delete(s, attachmentID, taskID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The attachment was deleted successfully."})
}
