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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

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
