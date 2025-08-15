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
	"fmt"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v4"
)

// DeleteProject handles deleting a project.
func DeleteProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	p := &models.Project{ID: projectID}
	if err := p.Delete(s, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateProject handles updating a project.
func UpdateProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	var p models.Project
	if err := c.Bind(&p); err != nil {
		return err
	}
	p.ID = projectID

	if err := p.Update(s, u); err != nil {
		return err
	}

	v2Project := &v2.Project{
		Project: p,
		Links: &v2.ProjectLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d", p.ID),
			},
			Tasks: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Project)
}

// GetProject handles getting a project by its ID.
func GetProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	p := &models.Project{ID: projectID}
	if err := p.ReadOne(s, u); err != nil {
		return err
	}

	v2Project := &v2.Project{
		Project: *p,
		Links: &v2.ProjectLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d", p.ID),
			},
			Tasks: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Project)
}

// CreateProject handles creating a new project.
func CreateProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	var p models.Project
	if err := c.Bind(&p); err != nil {
		return err
	}

	if err := p.Create(s, u); err != nil {
		return err
	}

	v2Project := &v2.Project{
		Project: p,
		Links: &v2.ProjectLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d", p.ID),
			},
			Tasks: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2Project)
}

// GetProjects handles getting all projects for the current user.
func GetProjects(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	page, perPage := v2.GetPageAndPerPage(c)
	search := c.QueryParam("s")
	isArchived := c.QueryParam("is_archived") == "true"

	p := models.Project{IsArchived: isArchived}
	projectsInterface, _, _, err := p.ReadAll(s, aut, search, page, perPage)
	if err != nil {
		return err
	}
	projects, ok := projectsInterface.([]*models.Project)
	if !ok {
		return fmt.Errorf("could not convert projects to []*models.Project")
	}

	v2Projects := make([]*v2.Project, len(projects))
	for i, p := range projects {
		v2Projects[i] = &v2.Project{
			Project: *p,
			Links: &v2.ProjectLinks{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/projects/%d", p.ID),
				},
				Tasks: &v2.Link{
					Href: fmt.Sprintf("/api/v2/projects/%d/tasks", p.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Projects)
}

// ListUsersForProject returns a list with all users who have access to a project, regardless of the method the project was shared with them.
func ListUsersForProject(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	project := models.Project{ID: projectID}
	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	canRead, _, err := project.CanRead(s, aut)
	if err != nil {
		return err
	}
	if !canRead {
		return echo.ErrForbidden
	}

	currentUser, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	search := c.QueryParam("s")
	users, err := models.ListUsersFromProject(s, &project, currentUser, search)
	if err != nil {
		return err
	}

	v2Users := make([]*v2.User, len(users))
	for i, u := range users {
		v2Users[i] = &v2.User{
			User: *u,
			Links: &v2.UserLinks{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/users/%d", u.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Users)
}
