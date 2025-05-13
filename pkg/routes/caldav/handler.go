// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
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
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
)

func getBasicAuthUserFromContext(c echo.Context) (*user.User, error) {
	u, is := c.Get("userBasicAuth").(*user.User)
	if !is {
		return nil, fmt.Errorf("user is not user.User element, is %s", reflect.TypeOf(c.Get("userBasicAuth")))
	}
	if u == nil {
		return nil, fmt.Errorf("userBasicAuth from context is nil")
	}
	return u, nil
}

// ProjectHandler handles requests related to project collections and individual project resources (calendars).
func ProjectHandler(c echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		log.Errorf("Error getting user from basic auth context: %v", err)
		return echo.ErrUnauthorized.SetInternal(fmt.Errorf("invalid user context: %w", err))
	}

	projectIDParam := c.Param("project")

	// Log CalDAV request details
	bodyBytes, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for further processing if any
	log.Debugf("[CALDAV] ProjectHandler: Method=%s, Path=%s, ProjectIDParam=%s, User=%s, Headers=%v, Body=%s",
		c.Request().Method, c.Path(), projectIDParam, u.Username, c.Request().Header, string(bodyBytes))


	if projectIDParam == "" {
		// Request to /dav/projects/ or /dav/projects
		// This should list all accessible calendars (projects) for the user.
		switch c.Request().Method {
		case "PROPFIND":
			return ListCalendars(c, u)
		case "REPORT":
			// CalDAV sync clients might send REPORT to the calendar collection.
			// For now, we can treat it like a PROPFIND or return not implemented.
			// Depending on the REPORT body, it could be a sync-collection report.
			// Let's assume for now it's asking for a list of resources.
			log.Debugf("[CALDAV] ProjectHandler (collection) received REPORT, treating as PROPFIND for now.")
			return ListCalendars(c, u)
		default:
			log.Warningf("[CALDAV] ProjectHandler (collection) received unhandled method %s", c.Request().Method)
			c.Response().Header().Set("Allow", "PROPFIND, REPORT, OPTIONS")
			return c.NoContent(http.StatusMethodNotAllowed)
		}
	}

	// Request to /dav/projects/:projectid/
	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		log.Errorf("Invalid project ID parameter '%s': %v", projectIDParam, err)
		return c.String(http.StatusBadRequest, "Invalid project ID")
	}

	// Verify project existence and user access (basic check, ListTasksInProject will do more)
	s := db.NewSession()
	defer s.Close()
	proj := models.Project{ID: projectID}
	canRead, _, errDb := proj.CanRead(s, u)
	if errDb != nil {
		log.Errorf("Error checking read permission for project %d by user %d: %v", projectID, u.ID, errDb)
		return c.NoContent(http.StatusInternalServerError)
	}
	if !canRead {
		log.Warningf("User %s (ID: %d) forbidden to access project %d", u.Username, u.ID, projectID)
		return c.NoContent(http.StatusForbidden)
	}
	if err := s.Commit(); err != nil { // commit the read-only transaction
		log.Errorf("Error committing after project read check: %v", err)
		// Not returning error here as it's a read check, but logging is important.
	}


	switch c.Request().Method {
	case "PROPFIND":
		return ListTasksInProject(c, u, projectID)
	case "REPORT":
		// REPORT on a specific calendar. This is often a calendar-multiget or calendar-query.
		// The current ListTasksInProject doesn't handle specific REPORT bodies.
		// For simplicity, we can treat it as a PROPFIND for all tasks in the project.
		// A more advanced implementation would parse the REPORT XML.
		log.Debugf("[CALDAV] ProjectHandler (specific project) received REPORT, treating as PROPFIND for tasks.")
		return ListTasksInProject(c, u, projectID)
	case "OPTIONS":
		c.Response().Header().Set("Allow", "OPTIONS, PROPFIND, REPORT, PUT, DELETE") // Methods for a calendar resource
		c.Response().Header().Set("DAV", "1, 3, calendar-access")                  // Basic DAV + CalDAV calendar access
		// Add other capabilities like calendar-schedule if supported
		return c.NoContent(http.StatusOK)
	default:
		log.Warningf("[CALDAV] ProjectHandler (specific project) received unhandled method %s for project %d", c.Request().Method, projectID)
		c.Response().Header().Set("Allow", "OPTIONS, PROPFIND, REPORT, PUT, DELETE") // Set Allow for unhandled methods too
		return c.NoContent(http.StatusMethodNotAllowed)
	}
}

// TaskHandler manages operations on individual task resources (.ics files).
func TaskHandler(c echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		log.Errorf("Error getting user from basic auth context: %v", err)
		return echo.ErrUnauthorized.SetInternal(fmt.Errorf("invalid user context: %w", err))
	}

	projectIDParam := c.Param("project")
	taskUIDParam := c.Param("task") // Includes .ics suffix

	// Log CalDAV request details
	bodyBytes, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
	log.Debugf("[CALDAV] TaskHandler: Method=%s, Path=%s, ProjectIDParam=%s, TaskUIDParam=%s, User=%s, Headers=%v, Body=%s",
		c.Request().Method, c.Path(), projectIDParam, taskUIDParam, u.Username, c.Request().Header, string(bodyBytes))

	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		log.Errorf("Invalid project ID parameter '%s' in TaskHandler: %v", projectIDParam, err)
		return c.String(http.StatusBadRequest, "Invalid project ID")
	}

	taskUID := strings.TrimSuffix(taskUIDParam, ".ics")
	if taskUID == "" {
		log.Errorf("Empty task UID in TaskHandler (param was '%s')", taskUIDParam)
		return c.String(http.StatusBadRequest, "Invalid task UID")
	}

	switch c.Request().Method {
	case http.MethodGet:
		return FetchTaskAsICS(c, u, projectID, taskUID)
	case http.MethodPut:
		return UpsertTaskFromICS(c, u, projectID, taskUID)
	case http.MethodDelete:
		return RemoveTaskICS(c, u, projectID, taskUID)
	case "PROPFIND":
		// PROPFIND on a specific task. We can fetch it and return its properties.
		// This is a simplified PROPFIND for a single resource.
		// FetchTaskAsICS already sets ETag and ContentType headers.
		// We need to wrap this in a multistatus response.
		// For now, let's delegate to FetchTaskAsICS which should return 200 OK with headers.
		// A full PROPFIND would require XML response.
		// TODO: Implement proper PROPFIND for single task resource.
		// As a temporary measure, try fetching it. If it exists, client might get headers.
		// This is not fully CalDAV compliant for PROPFIND on task.
		log.Debugf("[CALDAV] TaskHandler received PROPFIND for specific task, attempting simplified handling.")
		return GetTaskPropertiesAsXML(c, u, projectID, taskUID)
	default:
		log.Warningf("[CALDAV] TaskHandler received unhandled method %s", c.Request().Method)
		c.Response().Header().Set("Allow", "GET, PUT, DELETE, PROPFIND, OPTIONS")
		return c.NoContent(http.StatusMethodNotAllowed)
	}
}

// PrincipalHandler handles requests to principal URLs.
// Typically /dav/principals/users/<username>/ or /.well-known/caldav
func PrincipalHandler(c echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		log.Errorf("Error getting user from basic auth context: %v", err)
		return echo.ErrUnauthorized.SetInternal(fmt.Errorf("invalid user context: %w", err))
	}

	// Log CalDAV request details
	bodyBytes, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Debugf("[CALDAV] PrincipalHandler: Method=%s, Path=%s, User=%s, Headers=%v, Body=%s",
		c.Request().Method, c.Path(), u.Username, c.Request().Header, string(bodyBytes))

	switch c.Request().Method {
	case "PROPFIND":
		// PROPFIND on a principal URL should list properties of the principal,
		// including the calendar-home-set, which points to where calendars are.
		// For Vikunja, ProjectBasePath (/dav/projects/) can be considered the calendar-home-set.
		// It should also list the calendars themselves (projects).
		return ListPrincipalPropertiesAndCalendars(c, u)
	case "REPORT":
		log.Debugf("[CALDAV] PrincipalHandler received REPORT, which is not typically handled at this level. Path: %s", c.Path())
		// Reports are usually on calendar or task resources.
		// If it's a sync-collection on the principal, it's more complex.
		return c.NoContent(http.StatusNotImplemented)

	default:
		log.Warningf("[CALDAV] PrincipalHandler received unhandled method %s", c.Request().Method)
		// Key principal properties for PROPFIND:
		// <d:current-user-principal>, <d:principal-URL>, <c:calendar-home-set>
		c.Response().Header().Set("Allow", "PROPFIND, OPTIONS, REPORT")
		return c.NoContent(http.StatusMethodNotAllowed)
	}
}

// EntryHandler handles requests to the CalDAV service root (e.g., /dav/ or /dav).
func EntryHandler(c echo.Context) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		log.Errorf("Error getting user from basic auth context: %v", err)
		return echo.ErrUnauthorized.SetInternal(fmt.Errorf("invalid user context: %w", err))
	}

	// Log CalDAV request details
	bodyBytes, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Debugf("[CALDAV] EntryHandler: Method=%s, Path=%s, User=%s, Headers=%v, Body=%s",
		c.Request().Method, c.Path(), u.Username, c.Request().Header, string(bodyBytes))

	switch c.Request().Method {
	case "PROPFIND":
		// PROPFIND on the service root should discover the user's principal
		// and their calendar home set.
		// For Vikunja, this essentially means listing their calendars.
		return ListPrincipalPropertiesAndCalendars(c, u)
	case "REPORT":
		log.Debugf("[CALDAV] EntryHandler received REPORT, which is not typically handled at this level. Path: %s", c.Path())
		return c.NoContent(http.StatusNotImplemented)
	default:
		log.Warningf("[CALDAV] EntryHandler received unhandled method %s", c.Request().Method)
		c.Response().Header().Set("Allow", "PROPFIND, OPTIONS, REPORT")
		return c.NoContent(http.StatusMethodNotAllowed)
	}
}

func getProjectFromParam(c echo.Context) (*models.ProjectWithTasksAndBuckets, error) {
	param := c.Param("project")
	if param == "" {
		// This case should ideally be handled by a different route/handler (e.g., for /dav/projects/)
		// If ProjectHandler gets called with no :project param, it means it's the collection.
		// Returning nil, nil here and letting the caller (ProjectHandler) decide.
		return nil, nil
	}

	s := db.NewSession()
	defer s.Close()

	intParam, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("getProjectFromParam: invalid project ID '%s': %w", param, err)
	}

	// No need to handle FavoritesPseudoProjectID or SavedFilter here as ProjectHandler
	// will use ListCalendars or ListTasksInProject which work with actual project IDs.
	// CalDAV paths should resolve to actual calendar resources.

	p, err := models.GetProjectSimpleByID(s, intParam)
	if err != nil {
		if models.IsErrProjectDoesNotExist(err) {
			return nil, err // Return specific error for not found
		}
		return nil, fmt.Errorf("getProjectFromParam: db error getting project %d: %w", intParam, err)
	}
	if err := s.Commit(); err != nil {
		return nil, fmt.Errorf("getProjectFromParam: error committing after GetProjectSimpleByID: %w", err)
	}


	// We need ProjectWithTasksAndBuckets for consistency, though only Project might be used by caller
	// before calling ListTasksInProject which re-fetches with tasks.
	// For basic validation, Project info is enough.
	return &models.ProjectWithTasksAndBuckets{Project: *p}, nil
}
