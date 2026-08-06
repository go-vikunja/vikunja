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
	"github.com/labstack/echo/v5"
	"github.com/samedi/caldav-go"
	"github.com/samedi/caldav-go/lib"
)

func getBasicAuthUserFromContext(c *echo.Context) (*user.User, error) {
	u, is := c.Get("userBasicAuth").(*user.User)
	if !is {
		return &user.User{}, fmt.Errorf("user is not user element, is %s", reflect.TypeOf(c.Get("userBasicAuth")))
	}
	return u, nil
}

// caldav-go renders current-user-principal as "/<name>/", so pass the path
// without surrounding slashes or the href ends up with a double slash.
func setupUser(path string) {
	caldav.SetupUser(strings.Trim(path, "/"))
}

// ProjectHandler returns all tasks from a project
func ProjectHandler(c *echo.Context) error {
	project, err := getProjectFromParam(c)
	if err != nil && models.IsErrProjectDoesNotExist(err) {
		return c.String(http.StatusNotFound, "Project not found")
	}
	if err != nil {
		return err
	}

	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
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
		storage.task, _, err = caldav2.ParseTaskFromVTODO(vtodo)
		if err != nil {
			log.Warningf("[CALDAV] Failed to parse task: %v", err)
			return models.ErrInvalidData{Message: "Invalid task"}
		}
	}

	log.Debugf("[CALDAV] Request Body: %v\n", string(body))
	log.Debugf("[CALDAV] Request Headers: %v\n", c.Request().Header)

	// RFC 6578: intercept sync-collection REPORT before caldav-go sees it.
	// caldav-go returns 412 (empty body) for unknown REPORT types, which causes
	// iOS to silently stop syncing after the initial account add.
	if c.Request().Method == "REPORT" && strings.Contains(string(body), "sync-collection") {
		log.Debugf("[CALDAV] sync-collection REPORT for project %d by user %s", storage.project.ID, u.Username)
		return handleSyncCollectionReport(c, string(body), storage)
	}

	// caldav-go has no PROPPATCH dispatch case and falls back to a blanket 501;
	// intercept it here, same as the sync-collection REPORT above. The home-set
	// root (project.ID == 0) still falls through to that 501, unhandled for now.
	if c.Request().Method == "PROPPATCH" && storage.project.ID != 0 {
		log.Debugf("[CALDAV] PROPPATCH for project %d by user %s", storage.project.ID, u.Username)
		return handlePropPatch(c, string(body), storage.project, u)
	}

	caldav.SetupStorage(storage)
	setupUser(ProjectHomeSetPath)
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})
	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

// TaskHandler is the handler which manages updating/deleting a single task
func TaskHandler(c *echo.Context) error {
	project, err := getProjectFromParam(c)

	// PROPPATCH on a task resource gets the same 207/403 refusal as on a
	// collection, rather than caldav-go's blanket 501 (see ProjectHandler).
	if c.Request().Method == "PROPPATCH" {
		if err != nil && models.IsErrProjectDoesNotExist(err) {
			return c.String(http.StatusNotFound, "Project not found")
		}
		if err != nil {
			return err
		}

		u, err := getBasicAuthUserFromContext(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
		}

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not read request body").Wrap(err)
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		return handlePropPatch(c, string(body), project, u)
	}

	if err != nil {
		return err
	}

	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
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

// Sub-paths and foreign usernames must 404 (RFC 4918) instead of falling
// through to the merged principal pseudo-resource, which would silently serve
// the authenticated user's data and act as a username enumeration oracle.
// /.well-known/caldav also routes here (RFC 6764 bootstrap).
func isOwnPrincipalPath(path, username string) bool {
	if strings.TrimSuffix(path, "/") == "/.well-known/caldav" {
		return true
	}
	rest := strings.TrimPrefix(strings.TrimPrefix(path, PrincipalBasePath), "/")
	return strings.TrimSuffix(rest, "/") == username
}

// PrincipalHandler handles all request to principal resources
func PrincipalHandler(c *echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
	}

	if !isOwnPrincipalPath(c.Request().URL.Path, u.Username) {
		return c.String(http.StatusNotFound, "Not found")
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
	setupUser(principalPathForUser(u.Username))
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})

	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

// EntryHandler handles all request to principal resources
func EntryHandler(c *echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
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
	setupUser(principalPathForUser(u.Username))
	caldav.SetupSupportedComponents([]string{lib.VCALENDAR, lib.VTODO})

	response := caldav.HandleRequest(c.Request())
	response.Write(c.Response())
	return nil
}

func getProjectFromParam(c *echo.Context) (project *models.ProjectWithTasksAndBuckets, err error) {
	param := c.Param("project")
	if param == "" {
		return &models.ProjectWithTasksAndBuckets{}, nil
	}

	s := db.NewSession()
	defer s.Close()

	intParam, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		// The project ID given is not an integer, it cannot exist
		return nil, models.ErrProjectDoesNotExist{}
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
