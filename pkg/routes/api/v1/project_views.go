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

// RegisterProjectViews registers all project view routes
func RegisterProjectViews(a *echo.Group) {
	a.GET("/projects/:project/views", handler.WithDBAndUser(getAllProjectViews, false))
	a.GET("/projects/:project/views/:view", handler.WithDBAndUser(getProjectView, false))
	a.PUT("/projects/:project/views", handler.WithDBAndUser(createProjectView, true))
	a.POST("/projects/:project/views/:view", handler.WithDBAndUser(updateProjectView, true))
	a.DELETE("/projects/:project/views/:view", handler.WithDBAndUser(deleteProjectView, true))
}

// getAllProjectViews gets all project views for a project
// @Summary Get all project views
// @Description Returns all project views for a specific project
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Success 200 {array} models.ProjectView "The project views"
// @Failure 403 {object} web.HTTPError "The user does not have access to this project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views [get]
func getAllProjectViews(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	service := services.NewProjectViewService(s.Engine())
	views, totalCount, err := service.GetAll(s, projectID, u)
	if err != nil {
		return err
	}

	// Calculate pagination
	resultCount := len(views)
	totalPages := int64(1)
	if resultCount > 0 {
		totalPages = totalCount / int64(resultCount)
		if totalCount%int64(resultCount) > 0 {
			totalPages++
		}
	}

	// Set pagination headers
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalPages, 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, views)
}

// getProjectView gets a single project view
// @Summary Get one project view
// @Description Returns a project view by its ID
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view path int true "Project View ID"
// @Success 200 {object} models.ProjectView "The project view"
// @Failure 403 {object} web.HTTPError "The user does not have access to this project view"
// @Failure 404 {object} web.HTTPError "The project view does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{view} [get]
func getProjectView(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID")
	}

	service := services.NewProjectViewService(s.Engine())
	view, err := service.GetByIDAndProject(s, viewID, projectID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, view)
}

// createProjectView creates a new project view
// @Summary Create a project view
// @Description Create a project view in a specific project
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view body models.ProjectView true "The project view you want to create"
// @Success 200 {object} models.ProjectView "The created project view"
// @Failure 400 {object} web.HTTPError "Invalid project view object provided"
// @Failure 403 {object} web.HTTPError "The user does not have access to create a project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views [put]
func createProjectView(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse request body
	view := &models.ProjectView{}
	if err := c.Bind(view); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Set the project ID from the URL parameter
	view.ProjectID = projectID

	// Validate the view
	if err := c.Validate(view); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	service := services.NewProjectViewService(s.Engine())
	if err := service.Create(s, view, u, true, true); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, view)
}

// updateProjectView updates a project view
// @Summary Update a project view
// @Description Updates a project view
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view path int true "Project View ID"
// @Param viewData body models.ProjectView true "The project view with updated values"
// @Success 200 {object} models.ProjectView "The updated project view"
// @Failure 400 {object} web.HTTPError "Invalid project view object provided"
// @Failure 403 {object} web.HTTPError "The user does not have access to update this project view"
// @Failure 404 {object} web.HTTPError "The project view does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{view} [post]
func updateProjectView(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID")
	}

	// Parse request body
	view := &models.ProjectView{}
	if err := c.Bind(view); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Set IDs from URL parameters
	view.ID = viewID
	view.ProjectID = projectID

	// Validate the view
	if err := c.Validate(view); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	service := services.NewProjectViewService(s.Engine())
	if err := service.Update(s, view); err != nil {
		return err
	}

	// Fetch the updated view to return
	updatedView, err := service.GetByIDAndProject(s, viewID, projectID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updatedView)
}

// deleteProjectView deletes a project view
// @Summary Delete a project view
// @Description Deletes a project view
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view path int true "Project View ID"
// @Success 200 {object} models.Message "The project view was successfully deleted"
// @Failure 403 {object} web.HTTPError "The user does not have access to delete this project view"
// @Failure 404 {object} web.HTTPError "The project view does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{view} [delete]
func deleteProjectView(s *xorm.Session, u *user.User, c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID")
	}

	service := services.NewProjectViewService(s.Engine())
	if err := service.Delete(s, viewID, projectID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The project view was successfully deleted."})
}
