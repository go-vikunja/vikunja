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

// ProjectService is a service for managing projects.
type ProjectService struct{}

// NewProjectService returns a new ProjectService.
func NewProjectService() *ProjectService {
	return &ProjectService{}
}

func (ps *ProjectService) GetAll(s *xorm.Session, a web.Auth, p *models.Project, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}
	prs, resultCount, totalItems, err := p.ReadAll(s, u, search, page, perPage)
	if err != nil {
		return nil, 0, 0, err
	}

	projects, ok := prs.([]*models.Project)
	if !ok {
		return prs, resultCount, totalItems, nil
	}

	err = models.AddProjectDetails(s, projects, u)
	if err != nil {
		return
	}

	if p.Expand == "permissions" {
		err = models.AddMaxPermissionToProjects(s, projects, u)
		if err != nil {
			return
		}
	} else {
		for _, pr := range projects {
			pr.MaxPermission = models.PermissionUnknown
		}
	}
	return projects, resultCount, totalItems, err
}

func (ps *ProjectService) Create(s *xorm.Session, p *models.Project, a web.Auth) (*models.Project, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}
	if err := p.Create(s, u); err != nil {
		return nil, err
	}
	return p, nil
}

func (ps *ProjectService) Get(s *xorm.Session, projectID int64, a web.Auth) (*models.Project, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	p := &models.Project{ID: projectID}
	can, _, err := p.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err = p.ReadOne(s, u); err != nil {
		return nil, err
	}
	return p, nil
}

func (ps *ProjectService) Update(s *xorm.Session, p *models.Project, a web.Auth) (*models.Project, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	can, err := p.CanUpdate(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err := models.UpdateProject(s, p, u, false); err != nil {
		return nil, err
	}
	return p, nil
}

func (ps *ProjectService) Delete(s *xorm.Session, projectID int64, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	p := &models.Project{ID: projectID}
	can, err := p.CanDelete(s, u)
	if err != nil {
		return err
	}
	if !can {
		return echo.ErrForbidden
	}
	return p.Delete(s, u)
}

func AddProjectLinks(c echo.Context, p *models.Project) {
	p.Links = models.Links{
		"self": {
			HREF:   "/api/v2/projects/" + strconv.FormatInt(p.ID, 10),
			Method: "GET",
		},
		"update": {
			HREF:   "/api/v2/projects/" + strconv.FormatInt(p.ID, 10),
			Method: "PUT",
		},
		"delete": {
			HREF:   "/api/v2/projects/" + strconv.FormatInt(p.ID, 10),
			Method: "DELETE",
		},
	}
}
