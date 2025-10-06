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

// RegisterProjectUsers registers all project-user routes
func RegisterProjectUsers(a *echo.Group) {
	a.GET("/projects/:project/users", handler.WithDBAndUser(getAllProjectUsersLogic, false))
	a.PUT("/projects/:project/users", handler.WithDBAndUser(createProjectUserLogic, true))
	a.DELETE("/projects/:project/users/:user", handler.WithDBAndUser(deleteProjectUserLogic, true))
	a.POST("/projects/:project/users/:user", handler.WithDBAndUser(updateProjectUserLogic, true))
}

// getAllProjectUsersLogic handles retrieving all users with access to a project.
//
// @Summary Get all users with access to a project
// @Description Returns all users that have access to a project with their permission levels
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param page query int false "The page number for pagination. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search users by username."
// @Success 200 {array} models.UserWithPermission "All users with their permissions."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/users [get]
func getAllProjectUsersLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 50
	}

	search := c.QueryParam("s")

	// Get users from service
	service := services.NewProjectUserService(s.Engine())
	users, resultCount, totalItems, err := service.GetAll(s, projectID, u, search, page, perPage)
	if err != nil {
		return err
	}

	// Set pagination headers
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt((totalItems+int64(perPage)-1)/int64(perPage), 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, users)
}

// createProjectUserLogic handles adding a user to a project.
//
// @Summary Add a user to a project
// @Description Gives a user access to a project with specified permissions.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param projectUser body models.ProjectUser true "The user you want to add to the project."
// @Success 200 {object} models.ProjectUser "The created user-project relation."
// @Failure 400 {object} web.HTTPError "Invalid project user object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The project does not exist."
// @Failure 409 {object} web.HTTPError "The user already has access to the project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/users [put]
func createProjectUserLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse request body
	var projectUser models.ProjectUser
	if err := c.Bind(&projectUser); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project user object")
	}

	projectUser.ProjectID = projectID

	// Create user-project relation via service
	service := services.NewProjectUserService(s.Engine())
	if err := service.Create(s, &projectUser, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, projectUser)
}

// deleteProjectUserLogic handles removing a user from a project.
//
// @Summary Delete a user from a project
// @Description Removes a user's access to a project.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param user path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the project."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The project or user does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/users/{user} [delete]
func deleteProjectUserLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse user ID
	userID, err := strconv.ParseInt(c.Param("user"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	projectUser := &models.ProjectUser{
		ProjectID: projectID,
		UserID:    userID,
	}

	// Delete user-project relation via service
	service := services.NewProjectUserService(s.Engine())
	if err := service.Delete(s, projectUser); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The user was successfully removed from the project."})
}

// updateProjectUserLogic handles updating a user's permission level on a project.
//
// @Summary Update a user's permission level
// @Description Update a user's permission level on a project. The user needs to have admin access to the project.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param user path int true "User ID"
// @Param projectUser body models.ProjectUser true "The user you want to update."
// @Success 200 {object} models.ProjectUser "The updated user-project relation."
// @Failure 400 {object} web.HTTPError "Invalid project user object provided."
// @Failure 403 {object} web.HTTPError "The user does not have admin access to the project"
// @Failure 404 {object} web.HTTPError "The project or user does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/users/{user} [post]
func updateProjectUserLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse user ID
	userID, err := strconv.ParseInt(c.Param("user"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Parse request body
	var projectUser models.ProjectUser
	if err := c.Bind(&projectUser); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project user object")
	}

	projectUser.ProjectID = projectID
	projectUser.UserID = userID

	// Update user-project relation via service
	service := services.NewProjectUserService(s.Engine())
	if err := service.Update(s, &projectUser); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, projectUser)
}
