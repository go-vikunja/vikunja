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
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// RestoreProject restores a soft-deleted project.
// @Summary Restore a deleted project
// @Description Restores a project that was previously deleted (soft-deleted). Also restores all descendant projects that were soft-deleted at the same time.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Success 200 {object} models.Project "The restored project."
// @Failure 400 {object} web.HTTPError "Invalid project ID."
// @Failure 403 {object} web.HTTPError "The user does not have admin access to the project."
// @Failure 404 {object} web.HTTPError "The project does not exist or is not deleted."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /projects/{id}/restore [post]
func RestoreProject(c *echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid project ID",
		})
	}

	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	project, err := models.RestoreProject(s, projectID, auth)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, project)
}

// ListDeletedProjects returns all soft-deleted projects the user has admin access to.
// @Summary Get all deleted projects
// @Description Returns all soft-deleted projects that the current user has admin access to, along with their deletion date and days remaining before permanent purge.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.Project "All deleted projects."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /projects/deleted [get]
func ListDeletedProjects(c *echo.Context) error {
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	projects, err := models.GetDeletedProjects(s, auth)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, projects)
}
