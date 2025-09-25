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

// LabelRoutes defines all label API routes with their explicit permission scopes.
// This replaces the old implicit permission detection with explicit declarations.
var LabelRoutes = []APIRoute{
	{Method: "GET", Path: "/labels", Handler: handler.WithDBAndUser(getAllLabelsLogic, false), PermissionScope: "read_all"},
	{Method: "POST", Path: "/labels", Handler: handler.WithDBAndUser(createLabelLogic, true), PermissionScope: "create"},
	{Method: "PUT", Path: "/labels", Handler: handler.WithDBAndUser(createLabelLogic, true), PermissionScope: "create"}, // Frontend compatibility: PUT for creation
	{Method: "GET", Path: "/labels/:id", Handler: handler.WithDBAndUser(getLabelLogic, false), PermissionScope: "read_one"},
	{Method: "PUT", Path: "/labels/:id", Handler: handler.WithDBAndUser(updateLabelLogic, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/labels/:id", Handler: handler.WithDBAndUser(deleteLabelLogic, true), PermissionScope: "delete"},
}

// RegisterLabels registers all label routes
func RegisterLabels(a *echo.Group) {
	registerRoutes(a, LabelRoutes)
}

// getAllLabelsLogic handles retrieving all labels for a user
func getAllLabelsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Extract query parameters for search and pagination
	search := c.QueryParam("s")
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		page = 0
	}
	perPage, err := strconv.Atoi(c.QueryParam("per_page"))
	if err != nil {
		perPage = 50 // Default items per page
	}

	labelService := services.NewLabelService(s.Engine())
	labelsWithTaskID, _, _, err := labelService.GetAll(s, u, search, page, perPage)
	if err != nil {
		return err
	}

	// Convert from []*LabelWithTaskID to []*Label for response
	labelsSlice, ok := labelsWithTaskID.([]*models.LabelWithTaskID)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "unexpected type returned from GetAll")
	}

	labels := make([]*models.Label, len(labelsSlice))
	for i, labelWithTaskID := range labelsSlice {
		labels[i] = &labelWithTaskID.Label
	}

	return c.JSON(http.StatusOK, labels)
}

// createLabelLogic creates a new label
func createLabelLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	l := new(models.Label)
	if err := c.Bind(l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object provided.").SetInternal(err)
	}

	if err := c.Validate(l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	labelService := services.NewLabelService(s.Engine())
	if err := labelService.Create(s, l, u); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, l)
}

// getLabelLogic retrieves a single label by its ID
func getLabelLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID").SetInternal(err)
	}

	labelService := services.NewLabelService(s.Engine())
	label, err := labelService.Get(s, labelID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, label)
}

// updateLabelLogic handles updating a label
func updateLabelLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID").SetInternal(err)
	}

	updatePayload := new(models.Label)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object provided.").SetInternal(err)
	}
	updatePayload.ID = labelID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	labelService := services.NewLabelService(s.Engine())
	if err := labelService.Update(s, updatePayload, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updatePayload)
}

// deleteLabelLogic handles deleting a label
func deleteLabelLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID").SetInternal(err)
	}

	labelService := services.NewLabelService(s.Engine())
	label := &models.Label{ID: labelID, CreatedByID: u.ID}
	if err := labelService.Delete(s, label, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
