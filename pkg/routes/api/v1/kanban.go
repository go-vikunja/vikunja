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

package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// KanbanRoutes defines all Kanban/bucket API routes with their explicit permission scopes.
// This enables API tokens to be scoped for Kanban board management operations.
var KanbanRoutes = []APIRoute{
	{Method: "GET", Path: "/projects/:project/views/:view/buckets", Handler: handler.WithDBAndUser(getAllBuckets, false), PermissionScope: "read_all"},
	{Method: "PUT", Path: "/projects/:project/views/:view/buckets", Handler: handler.WithDBAndUser(createBucket, true), PermissionScope: "create"},
	{Method: "POST", Path: "/projects/:project/views/:view/buckets/:bucket", Handler: handler.WithDBAndUser(updateBucket, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/projects/:project/views/:view/buckets/:bucket", Handler: handler.WithDBAndUser(deleteBucket, true), PermissionScope: "delete"},
	{Method: "POST", Path: "/projects/:project/views/:view/buckets/:bucket/tasks", Handler: handler.WithDBAndUser(moveTaskToBucket, true), PermissionScope: "move_task"},
}

// RegisterKanbanRoutes registers all Kanban routes
func RegisterKanbanRoutes(a *echo.Group) {
	registerRoutes(a, KanbanRoutes)

	// Move task route creates separate projects_views_buckets_tasks group due to /tasks suffix,
	// but we want it in projects_views_buckets for logical grouping with other bucket operations
	routes := models.GetAPITokenRoutes()
	if tasksGroup, ok := routes["v1"]["projects_views_buckets_tasks"]; ok {
		if moveRoute, ok := tasksGroup["move_task"]; ok {
			if routes["v1"]["projects_views_buckets"] == nil {
				routes["v1"]["projects_views_buckets"] = make(models.APITokenRoute)
			}
			routes["v1"]["projects_views_buckets"]["move_task"] = moveRoute
		}
	}
}

// getAllBuckets gets all buckets for a project view
func getAllBuckets(s *xorm.Session, u *user.User, c echo.Context) error {
	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID").SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	kanbanService := services.NewKanbanService(s.Engine())
	buckets, err := kanbanService.GetAllBuckets(s, viewID, projectID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, buckets)
}

// createBucket creates a new bucket in a project view
func createBucket(s *xorm.Session, u *user.User, c echo.Context) error {
	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID").SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	bucket := new(models.Bucket)
	if err := c.Bind(bucket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bucket object provided.").SetInternal(err)
	}

	bucket.ProjectViewID = viewID
	bucket.ProjectID = projectID

	if err := c.Validate(bucket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	kanbanService := services.NewKanbanService(s.Engine())
	if err := kanbanService.CreateBucket(s, bucket, u); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, bucket)
}

// updateBucket updates a bucket in a project view
func updateBucket(s *xorm.Session, u *user.User, c echo.Context) error {
	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID").SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	bucketID, err := strconv.ParseInt(c.Param("bucket"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bucket ID").SetInternal(err)
	}

	bucket := new(models.Bucket)
	if err := c.Bind(bucket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bucket object provided.").SetInternal(err)
	}

	bucket.ID = bucketID
	bucket.ProjectViewID = viewID
	bucket.ProjectID = projectID

	if err := c.Validate(bucket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	kanbanService := services.NewKanbanService(s.Engine())
	if err := kanbanService.UpdateBucket(s, bucket, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, bucket)
}

// deleteBucket deletes a bucket from a project view
func deleteBucket(s *xorm.Session, u *user.User, c echo.Context) error {
	bucketID, err := strconv.ParseInt(c.Param("bucket"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bucket ID").SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	kanbanService := services.NewKanbanService(s.Engine())
	if err := kanbanService.DeleteBucket(s, bucketID, projectID, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// moveTaskToBucket moves a task to a bucket
func moveTaskToBucket(s *xorm.Session, u *user.User, c echo.Context) error {
	viewID, err := strconv.ParseInt(c.Param("view"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid view ID").SetInternal(err)
	}

	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID").SetInternal(err)
	}

	bucketID, err := strconv.ParseInt(c.Param("bucket"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bucket ID").SetInternal(err)
	}

	taskBucket := new(models.TaskBucket)
	if err := c.Bind(taskBucket); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task bucket object provided.").SetInternal(err)
	}

	taskBucket.ProjectViewID = viewID
	taskBucket.ProjectID = projectID
	taskBucket.BucketID = bucketID

	kanbanService := services.NewKanbanService(s.Engine())
	if err := kanbanService.MoveTaskToBucket(s, taskBucket, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
