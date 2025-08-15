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

// DeleteTeam handles deleting a team.
func DeleteTeam(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID.")
	}

	t := &models.Team{ID: teamID}
	if err := t.Delete(s, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateTeam handles updating a team.
func UpdateTeam(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID.")
	}

	var t models.Team
	if err := c.Bind(&t); err != nil {
		return err
	}
	t.ID = teamID

	if err := t.Update(s, u); err != nil {
		return err
	}

	v2Team := &v2.Team{
		Team: t,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/teams/%d", t.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Team)
}

// GetTeam handles getting a team by its ID.
func GetTeam(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID.")
	}

	t := &models.Team{ID: teamID}
	if err := t.ReadOne(s, u); err != nil {
		return err
	}

	v2Team := &v2.Team{
		Team: *t,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/teams/%d", t.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Team)
}

// CreateTeam handles creating a new team.
func CreateTeam(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	var t models.Team
	if err := c.Bind(&t); err != nil {
		return err
	}

	if err := t.Create(s, u); err != nil {
		return err
	}

	v2Team := &v2.Team{
		Team: t,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/teams/%d", t.ID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2Team)
}

// GetTeams handles getting all teams for the current user.
func GetTeams(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	page, perPage := v2.GetPageAndPerPage(c)
	search := c.QueryParam("s")
	includePublic := c.QueryParam("include_public") == "true"

	t := models.Team{IncludePublic: includePublic}
	teamsInterface, _, _, err := t.ReadAll(s, aut, search, page, perPage)
	if err != nil {
		return err
	}
	teams, ok := teamsInterface.([]*models.Team)
	if !ok {
		return fmt.Errorf("could not convert teams to []*models.Team")
	}

	v2Teams := make([]*v2.Team, len(teams))
	for i, t := range teams {
		v2Teams[i] = &v2.Team{
			Team: *t,
			Links: &v2.Links{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/teams/%d", t.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Teams)
}
