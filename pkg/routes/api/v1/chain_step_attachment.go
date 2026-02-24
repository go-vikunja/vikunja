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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// UploadChainStepAttachment handles file upload for a chain step
// @Summary Upload an attachment to a chain step
// @Tags task_chains
// @Accept multipart/form-data
// @Produce json
// @Param step path int true "Step ID"
// @Param files formance file true "The file to upload"
// @Security JWTKeyAuth
// @Success 200 {object} models.TaskChainStepAttachment
// @Router /chainsteps/{step}/attachments [put]
func UploadChainStepAttachment(c *echo.Context) error {
	stepID, err := strconv.ParseInt(c.Param("step"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid step ID")
	}

	// Get current user
	authUser, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unable to authenticate")
	}

	s := db.NewSession()
	defer s.Close()

	// Verify the step exists and the user owns the chain
	step := &models.TaskChainStep{}
	exists, err := s.Where("id = ?", stepID).Get(step)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if !exists {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusNotFound, "Step not found")
	}

	chain := &models.TaskChain{ID: step.ChainID}
	can, _, err := chain.CanRead(s, authUser)
	if err != nil || !can {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// Get uploaded file
	file, err := c.FormFile("files")
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}

	src, err := file.Open()
	if err != nil {
		_ = s.Rollback()
		return err
	}
	defer src.Close()

	// Store the file using Vikunja's file storage
	storedFile, err := files.Create(src, file.Filename, uint64(file.Size), authUser)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Create the attachment record
	att := &models.TaskChainStepAttachment{
		StepID:      stepID,
		FileID:      storedFile.ID,
		FileName:    file.Filename,
		CreatedByID: authUser.GetID(),
	}

	if err := s.Begin(); err != nil {
		return err
	}

	_, err = s.Insert(att)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	att.File = storedFile
	return c.JSON(http.StatusOK, att)
}

// DeleteChainStepAttachment removes an attachment from a chain step
// @Summary Delete a chain step attachment
// @Tags task_chains
// @Produce json
// @Param step path int true "Step ID"
// @Param attachment path int true "Attachment ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Message
// @Router /chainsteps/{step}/attachments/{attachment} [delete]
func DeleteChainStepAttachment(c *echo.Context) error {
	stepID, err := strconv.ParseInt(c.Param("step"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid step ID")
	}
	attID, err := strconv.ParseInt(c.Param("attachment"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid attachment ID")
	}

	authUser, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unable to authenticate")
	}

	s := db.NewSession()
	defer s.Close()

	// Verify ownership
	step := &models.TaskChainStep{}
	exists, err := s.Where("id = ?", stepID).Get(step)
	if err != nil || !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Step not found")
	}

	chain := &models.TaskChain{ID: step.ChainID}
	can, _, err := chain.CanRead(s, authUser)
	if err != nil || !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	att := &models.TaskChainStepAttachment{}
	exists, err = s.Where("id = ? AND step_id = ?", attID, stepID).Get(att)
	if err != nil || !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Attachment not found")
	}

	if err := s.Begin(); err != nil {
		return err
	}

	// Delete the file
	f := &files.File{ID: att.FileID}
	_ = f.Delete(s)

	// Delete the record
	_, err = s.Where("id = ?", attID).Delete(&models.TaskChainStepAttachment{})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Attachment deleted"})
}
