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

package caldav

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"

	caldav2 "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/samedi/caldav-go"
	"github.com/samedi/caldav-go/lib"
)

func getBasicAuthUserFromContext(c echo.Context) (*user.User, error) {
	u, is := c.Get("userBasicAuth").(*user.User)
	if !is {
		return &user.User{}, fmt.Errorf("user is not user element, is %s", reflect.TypeOf(c.Get("userBasicAuth")))
	}
	return u, nil
}

// ProjectHandler returns all tasks from a project
func ProjectHandler(c echo.Context) error {
	project, err := getProjectFromParam(c)
	if err != nil && models.IsErrProjectDoesNotExist(err) {
		return c.String(http.StatusNotFound, "Project not found")
	}
	if err != nil {
		return err
	}

	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	storage := &VikunjaCaldavProjectStorage{
		project: project,
		user:    u,
	}

	// Try to parse a task from the request payload
	body, _ := io.ReadAll(c.Request().Body)
	// Restore the io.ReadCloser to its original state
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	// Parse it
	vtodo := string(body)
	if vtodo != "" && strings.HasPrefix(vtodo, `BEGIN:VCALENDAR`) {
		storage.task, err = caldav2.ParseTaskFromVTODO(vtodo)
		if err != nil {
			log.Warningf("[CALDAV] Failed to parse task: %v", err)
			return models.ErrInvalidData{Message: "Invalid task"}
		}
	}

	log.Debugf("[CALDAV] Request Body: %v\n", string(body))
	log.Debugf("[CALDAV] Request Headers: %v\n", c.Request().Header)

	caldav.SetupStorage(storage)
	caldav.SetupUser("dav/projects")
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})
	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

// TaskHandler is the handler which manages updating/deleting a single task
func TaskHandler(c echo.Context) error {
	project, err := getProjectFromParam(c)
	if err != nil {
		return err
	}

	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	// Get the task uid
	taskUID := strings.TrimSuffix(c.Param("task"), ".ics")

	storage := &VikunjaCaldavProjectStorage{
		project: project,
		task:    &models.Task{UID: taskUID},
		user:    u,
	}

	caldav.SetupStorage(storage)
	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

// PrincipalHandler handles all request to principal resources
func PrincipalHandler(c echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	storage := &VikunjaCaldavProjectStorage{
		user:        u,
		isPrincipal: true,
	}

	// Try to parse a task from the request payload
	body, _ := io.ReadAll(c.Request().Body)
	// Restore the io.ReadCloser to its original state
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	log.Debugf("[CALDAV] Request Body: %v\n", string(body))
	log.Debugf("[CALDAV] Request Headers: %v\n", c.Request().Header)

	caldav.SetupStorage(storage)
	caldav.SetupUser("dav/principals/" + u.Username)
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})

	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

// EntryHandler handles all request to principal resources
func EntryHandler(c echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.ErrInternalServerError.SetInternal(err)
	}

	storage := &VikunjaCaldavProjectStorage{
		user:    u,
		isEntry: true,
	}

	// Try to parse a task from the request payload
	body, _ := io.ReadAll(c.Request().Body)
	// Restore the io.ReadCloser to its original state
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	log.Debugf("[CALDAV] Request Body: %v\n", string(body))
	log.Debugf("[CALDAV] Request Headers: %v\n", c.Request().Header)

	caldav.SetupStorage(storage)
	caldav.SetupUser("dav/principals/" + u.Username)
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})

	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

func getProjectFromParam(c echo.Context) (project *models.ProjectWithTasksAndBuckets, err error) {
	param := c.Param("project")
	if param == "" {
		return &models.ProjectWithTasksAndBuckets{}, nil
	}

	s := db.NewSession()
	defer s.Close()

	intParam, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, err
	}

	if intParam == models.FavoritesPseudoProjectID {
		return &models.ProjectWithTasksAndBuckets{Project: models.FavoritesPseudoProject}, nil
	}

	if intParam < models.FavoritesPseudoProjectID {
		var sf *models.SavedFilter
		sf, err = models.GetSavedFilterSimpleByID(s, models.GetSavedFilterIDFromProjectID(intParam))
		if err != nil {
			return nil, err
		}

		project = &models.ProjectWithTasksAndBuckets{Project: *sf.ToProject()}
		return
	}

	p, err := models.GetProjectSimpleByID(s, intParam)
	if err != nil {
		return nil, err
	}

	project = &models.ProjectWithTasksAndBuckets{Project: *p}
	return
}
