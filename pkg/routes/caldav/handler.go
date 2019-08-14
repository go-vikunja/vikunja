//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"bytes"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/samedi/caldav-go"
	"github.com/samedi/caldav-go/lib"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

func getBasicAuthUserFromContext(c echo.Context) (user models.User, err error) {
	u, is := c.Get("userBasicAuth").(models.User)
	if !is {
		return models.User{}, fmt.Errorf("user is not user element, is %s", reflect.TypeOf(c.Get("userBasicAuth")))
	}
	return u, nil
}

// ListHandler returns all tasks from a list
func ListHandler(c echo.Context) error {
	listID, err := getIntParam(c, "list")
	if err != nil {
		return err
	}

	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		log.Error(err)
		return echo.ErrInternalServerError
	}

	storage := &VikunjaCaldavListStorage{
		list: &models.List{ID: listID},
		user: &u,
	}

	// Try to parse a task from the request payload
	body, _ := ioutil.ReadAll(c.Request().Body)
	// Restore the io.ReadCloser to its original state
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	// Parse it
	vtodo := string(body)
	if vtodo != "" && strings.HasPrefix(vtodo, `BEGIN:VCALENDAR`) {
		storage.task, err = parseTaskFromVTODO(vtodo)
		if err != nil {
			log.Error(err)
			return echo.ErrInternalServerError
		}
	}

	log.Debugf("[CALDAV] Request Body: %v\n", string(body))
	log.Debugf("[CALDAV] Request Headers: %v\n", c.Request().Header)

	caldav.SetupStorage(storage)
	caldav.SetupUser("dav/lists")
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})
	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

// TaskHandler is the handler which manages updating/deleting a single task
func TaskHandler(c echo.Context) error {
	listID, err := getIntParam(c, "list")
	if err != nil {
		return err
	}

	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		log.Error(err)
		return echo.ErrInternalServerError
	}

	// Get the task uid
	taskUID := strings.TrimSuffix(c.Param("task"), ".ics")

	storage := &VikunjaCaldavListStorage{
		list: &models.List{ID: listID},
		task: &models.Task{UID: taskUID},
		user: &u,
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
		log.Error(err)
		return echo.ErrInternalServerError
	}

	storage := &VikunjaCaldavListStorage{
		user:        &u,
		isPrincipal: true,
	}

	// Try to parse a task from the request payload
	body, _ := ioutil.ReadAll(c.Request().Body)
	// Restore the io.ReadCloser to its original state
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))

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
		log.Error(err)
		return echo.ErrInternalServerError
	}

	storage := &VikunjaCaldavListStorage{
		user:    &u,
		isEntry: true,
	}

	// Try to parse a task from the request payload
	body, _ := ioutil.ReadAll(c.Request().Body)
	// Restore the io.ReadCloser to its original state
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))

	log.Debugf("[CALDAV] Request Body: %v\n", string(body))
	log.Debugf("[CALDAV] Request Headers: %v\n", c.Request().Header)

	caldav.SetupStorage(storage)
	caldav.SetupUser("dav/principals/" + u.Username)
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})

	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

func getIntParam(c echo.Context, paramName string) (intParam int64, err error) {
	param := c.Param(paramName)
	if param == "" {
		return 0, nil
	}

	intParam, err = strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, handler.HandleHTTPError(err, c)
	}
	return
}
