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
	"fmt"
	"strconv"
	"strings"

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

// TaskOptions holds all the options for getting tasks
type TaskOptions struct {
	models.TaskSortBy
	models.TaskFilterBy
	models.TaskPagination
}

func (ts *TaskService) applyTaskOptions(q *xorm.Session, options TaskOptions) (*xorm.Session, error) {
	for _, sortBy := range options.SortBy {
		var direction = " ASC"
		var field = sortBy
		if strings.HasPrefix(sortBy, "-") {
			direction = " DESC"
			field = field[1:]
		}

		err := models.ValidateTaskFieldForSorting(field)
		if err != nil {
			return nil, err
		}
		q = q.OrderBy(fmt.Sprintf("%s %s", field, direction))
	}

	// TODO: Add filtering

	return q, nil
}
func (ts *TaskService) Get(s *xorm.Session, taskID int64, a web.Auth) (*models.Task, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	t := &models.Task{ID: taskID}
	can, _, err := t.CanRead(s, u)
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

func (ts *TaskService) GetAll(s *xorm.Session, a web.Auth, search string, page, perPage int, options TaskOptions) (result interface{}, resultCount int, totalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get all projects the user has access to
	projects, _, err := models.GetAllProjectsForUser(s, u.ID, &models.ProjectOptions{})
	if err != nil {
		return nil, 0, 0, err
	}

	if len(projects) == 0 {
		return []*models.Task{}, 0, 0, nil
	}

	projectIDs := make([]int64, len(projects))
	for i, p := range projects {
		projectIDs[i] = p.ID
	}

	s = s.In("project_id", projectIDs)

	s, err = ts.applyTaskOptions(s, options)
	if err != nil {
		return nil, 0, 0, err
	}

	var tasks []*models.Task
	totalItems, err = s.FindAndCount(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	taskMap := make(map[int64]*models.Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	err = models.AddMoreInfoToTasks(s, taskMap, u, nil, nil)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

func (ts *TaskService) GetByProject(s *xorm.Session, projectID int64, a web.Auth, search string, page, perPage int, options TaskOptions) (result interface{}, resultCount int, totalItems int64, err error) {
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

	s = s.Where("project_id = ?", projectID)
	s, err = ts.applyTaskOptions(s, options)
	if err != nil {
		return nil, 0, 0, err
	}

	var tasks []*models.Task
	totalItems, err = s.FindAndCount(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
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
