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

// TaskService is a service for managing tasks.
type TaskService struct{}

// NewTaskService returns a new TaskService.
func NewTaskService() *TaskService {
	return &TaskService{}
}

func (ts *TaskService) Get(s *xorm.Session, taskID int64, a web.Auth) (*models.Task, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	t := &models.Task{ID: taskID}
	can, err := t.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err := t.ReadOne(s, u); err != nil {
		return nil, err
	}
	return t, nil
}

func (ts *TaskService) GetAll(s *xorm.Session, a web.Auth, search string, page, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}
	tc := &models.TaskCollection{}
	return tc.ReadAll(s, u, search, page, perPage)
}

func (ts *TaskService) GetByProject(s *xorm.Session, projectID int64, a web.Auth, search string, page, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	p := &models.Project{ID: projectID}
	can, _, err := p.CanRead(s, u)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, echo.ErrForbidden
	}

	tc := &models.TaskCollection{
		ProjectID: projectID,
	}

	return tc.ReadAll(s, u, search, page, perPage)
}

func (ts *TaskService) Create(s *xorm.Session, t *models.Task, a web.Auth) (*models.Task, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	p := &models.Project{ID: t.ProjectID}
	can, err := p.CanWrite(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err := t.Create(s, u); err != nil {
		return nil, err
	}
	return t, nil
}

func (ts *TaskService) Update(s *xorm.Session, t *models.Task, a web.Auth) (*models.Task, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	can, err := t.CanUpdate(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err := t.Update(s, u); err != nil {
		return nil, err
	}
	return t, nil
}

func (ts *TaskService) Delete(s *xorm.Session, taskID int64, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	t := &models.Task{ID: taskID}
	can, err := t.CanDelete(s, u)
	if err != nil {
		return err
	}
	if !can {
		return echo.ErrForbidden
	}
	return t.Delete(s, u)
}

func AddTaskLinks(a web.Auth, t *models.Task) {
	t.Links = models.Links{
		"self": {
			HREF:   "/api/v2/tasks/" + strconv.FormatInt(t.ID, 10),
			Method: "GET",
		},
		"update": {
			HREF:   "/api/v2/tasks/" + strconv.FormatInt(t.ID, 10),
			Method: "PUT",
		},
		"delete": {
			HREF:   "/api/v2/tasks/" + strconv.FormatInt(t.ID, 10),
			Method: "DELETE",
		},
	}
}
