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
	"os"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

func checkExportRequest(c echo.Context) (s *xorm.Session, u *user.User, err error) {
	s = db.NewSession()
	defer s.Close()

	err = s.Begin()
	if err != nil {
		return nil, nil, handler.HandleHTTPError(err)
	}

	u, err = user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return nil, nil, handler.HandleHTTPError(err)
	}

	// Users authenticated with a third-party are unable to provide their password.
	if u.Issuer != user.IssuerLocal {
		return
	}

	var pass UserPasswordConfirmation
	if err := c.Bind(&pass); err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, "No password provided.").SetInternal(err)
	}

	err = c.Validate(pass)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, err).SetInternal(err)
	}

	err = user.CheckUserPassword(u, pass.Password)
	if err != nil {
		_ = s.Rollback()
		return nil, nil, handler.HandleHTTPError(err)
	}

	return
}

// RequestUserDataExport is the handler to request a user data export
// @Summary Request a user data export.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param password body v1.UserPasswordConfirmation true "User password to confirm the data export request."
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/export/request [post]
func RequestUserDataExport(c echo.Context) error {
	s, u, err := checkExportRequest(c)
	if err != nil {
		return err
	}

	err = events.Dispatch(&models.UserDataExportRequestedEvent{
		User: u,
	})
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{Message: "Successfully requested data export. We will send you an email when it's ready."})
}

// DownloadUserDataExport is the handler to download a created user data export
// @Summary Download a user data export.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param password body v1.UserPasswordConfirmation true "User password to confirm the download."
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 404 {object} web.HTTPError "No user data export found."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/export/download [post]
func DownloadUserDataExport(c echo.Context) error {
	s, u, err := checkExportRequest(c)
	if err != nil {
		return err
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err)
	}

	// Check if user has an export file
	exportNotFoundError := echo.NewHTTPError(http.StatusNotFound, "No user data export found.")
	if u.ExportFileID == 0 {
		return exportNotFoundError
	}

	// Download
	exportFile := &files.File{ID: u.ExportFileID}
	err = exportFile.LoadFileMetaByID()
	if err != nil {
		if files.IsErrFileDoesNotExist(err) {
			return exportNotFoundError
		}
		return handler.HandleHTTPError(err)
	}
	err = exportFile.LoadFileByID()
	if err != nil {
		if os.IsNotExist(err) {
			return exportNotFoundError
		}
		return handler.HandleHTTPError(err)
	}

	http.ServeContent(c.Response(), c.Request(), exportFile.Name, exportFile.Created, exportFile.File)
	return nil
}

type UserExportStatus struct {
	ID      int64     `json:"id"`
	Size    uint64    `json:"size"`
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"`
}

// GetUserExportStatus returns metadata about the current user export if it exists
// @Summary Get current user data export
// @tags user
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} v1.UserExportStatus
// @Router /user/export [get]
func GetUserExportStatus(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	if u.ExportFileID == 0 {
		return c.JSON(http.StatusOK, struct{}{})
	}

	exportFile := &files.File{ID: u.ExportFileID}
	if err := exportFile.LoadFileMetaByID(); err != nil {
		return handler.HandleHTTPError(err)
	}

	status := UserExportStatus{
		ID:      exportFile.ID,
		Size:    exportFile.Size,
		Created: exportFile.Created,
		Expires: exportFile.Created.Add(7 * 24 * time.Hour),
	}

	return c.JSON(http.StatusOK, status)
}
