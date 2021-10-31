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
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

func checkExportRequest(c echo.Context) (s *xorm.Session, u *user.User, err error) {
	s = db.NewSession()
	defer s.Close()

	err = s.Begin()
	if err != nil {
		return nil, nil, handler.HandleHTTPError(err, c)
	}

	u, err = user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return nil, nil, handler.HandleHTTPError(err, c)
	}

	// Users authenticated with a third-party are unable to provide their password.
	if u.Issuer != user.IssuerLocal {
		return
	}

	var pass UserPasswordConfirmation
	if err := c.Bind(&pass); err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	err = c.Validate(pass)
	if err != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = user.CheckUserPassword(u, pass.Password)
	if err != nil {
		_ = s.Rollback()
		return nil, nil, handler.HandleHTTPError(err, c)
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
		return handler.HandleHTTPError(err, c)
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
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
		return handler.HandleHTTPError(err, c)
	}

	// Download
	exportFile := &files.File{ID: u.ExportFileID}
	err = exportFile.LoadFileMetaByID()
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}
	err = exportFile.LoadFileByID()
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	http.ServeContent(c.Response(), c.Request(), exportFile.Name, exportFile.Created, exportFile.File)
	return nil
}
