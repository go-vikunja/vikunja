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

package services

import (
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// TeamService is a service for managing teams.
type TeamService struct{}

// NewTeamService returns a new TeamService.
func NewTeamService() *TeamService {
	return &TeamService{}
}

func (ts *TeamService) GetByProject(s *xorm.Session, projectID int64, a web.Auth, search string, page, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, u)
	if err != nil {
		return nil, 0, 0, err
	}
	if perm < int(models.PermissionAdmin) {
		return nil, 0, 0, echo.ErrForbidden
	}

	tp := &models.TeamProject{ProjectID: projectID}
	return tp.ReadAll(s, u, search, page, perPage)
}

func (ts *TeamService) Create(s *xorm.Session, tp *models.TeamProject, a web.Auth) (*models.TeamProject, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	p := &models.Project{ID: tp.ProjectID}
	_, perm, err := p.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if perm < int(models.PermissionAdmin) {
		return nil, echo.ErrForbidden
	}

	if err := tp.Create(s, u); err != nil {
		return nil, err
	}
	return tp, nil
}

func (ts *TeamService) Update(s *xorm.Session, tp *models.TeamProject, a web.Auth) (*models.TeamProject, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	p := &models.Project{ID: tp.ProjectID}
	_, perm, err := p.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if perm < int(models.PermissionAdmin) {
		return nil, echo.ErrForbidden
	}

	if err := tp.Update(s); err != nil {
		return nil, err
	}
	return tp, nil
}

func (ts *TeamService) Delete(s *xorm.Session, projectID, teamID int64, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	p := &models.Project{ID: projectID}
	_, perm, err := p.CanRead(s, u)
	if err != nil {
		return err
	}
	if perm < int(models.PermissionAdmin) {
		return echo.ErrForbidden
	}

	tp := &models.TeamProject{
		ProjectID: projectID,
		TeamID:    teamID,
	}

	return tp.Delete(s)
}

func AddTeamLinks(c echo.Context, t *models.TeamWithPermission) {
	t.Links = models.Links{
		"self": {
			HREF:   "/api/v2/teams/" + strconv.FormatInt(t.ID, 10),
			Method: "GET",
		},
		"update": {
			HREF:   "/api/v2/teams/" + strconv.FormatInt(t.ID, 10),
			Method: "PUT",
		},
		"delete": {
			HREF:   "/api/v2/teams/" + strconv.FormatInt(t.ID, 10),
			Method: "DELETE",
		},
	}
}
