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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// LinkShareRoutes defines all link share API routes with their explicit permission scopes.
// This replaces the old WebHandler-based implicit permission detection with explicit declarations.
var LinkShareRoutes = []APIRoute{
	{Method: "GET", Path: "/projects/:project/shares", Handler: handler.WithDBAndUser(getAllLinkSharesLogic, false), PermissionScope: "read_all"},
	{Method: "GET", Path: "/projects/:project/shares/:share", Handler: handler.WithDBAndUser(getLinkShareLogic, false), PermissionScope: "read_one"},
	{Method: "PUT", Path: "/projects/:project/shares", Handler: handler.WithDBAndUser(createLinkShareLogic, true), PermissionScope: "create"},
	{Method: "POST", Path: "/projects/:project/shares/:share", Handler: handler.WithDBAndUser(updateLinkShareLogic, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/projects/:project/shares/:share", Handler: handler.WithDBAndUser(deleteLinkShareLogic, true), PermissionScope: "delete"},
}

// RegisterLinkShares registers all link share routes
func RegisterLinkShares(a *echo.Group) {
	registerRoutes(a, LinkShareRoutes)
}

// getAllLinkSharesLogic handles retrieving all link shares for a project
func getAllLinkSharesLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	linkShareService := services.NewLinkShareService(s.Engine())
	shares, err := linkShareService.GetByProjectID(s, projectID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, shares)
}

// getLinkShareLogic retrieves a single link share by its ID
func getLinkShareLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	shareID, err := strconv.ParseInt(c.Param("share"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid share ID").SetInternal(err)
	}

	linkShareService := services.NewLinkShareService(s.Engine())
	share, err := linkShareService.GetByID(s, shareID)
	if err != nil {
		return err
	}

	// Check if user can read this link share
	canRead, _, err := linkShareService.CanRead(s, share, u)
	if err != nil {
		return err
	}
	if !canRead {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	return c.JSON(http.StatusOK, share)
}

// createLinkShareLogic creates a new link share for a project
func createLinkShareLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	share := new(models.LinkSharing)
	if err := c.Bind(share); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid link share object provided.").SetInternal(err)
	}

	// Ensure the project ID matches the URL parameter
	share.ProjectID = projectID

	if err := c.Validate(share); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	linkShareService := services.NewLinkShareService(s.Engine())
	err = linkShareService.Create(s, share, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, share)
}

// updateLinkShareLogic handles updating a link share
func updateLinkShareLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	shareID, err := strconv.ParseInt(c.Param("share"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid share ID").SetInternal(err)
	}

	updatePayload := new(models.LinkSharing)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid link share object provided.").SetInternal(err)
	}
	updatePayload.ID = shareID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	linkShareService := services.NewLinkShareService(s.Engine())
	err = linkShareService.Update(s, updatePayload, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updatePayload)
}

// deleteLinkShareLogic handles deleting a link share
func deleteLinkShareLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	shareID, err := strconv.ParseInt(c.Param("share"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid share ID").SetInternal(err)
	}

	linkShareService := services.NewLinkShareService(s.Engine())
	err = linkShareService.Delete(s, shareID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The link share was deleted successfully."})
}
