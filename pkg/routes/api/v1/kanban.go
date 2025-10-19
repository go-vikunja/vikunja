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

// RegisterKanbanRoutes registers all Kanban routes
func RegisterKanbanRoutes(a *echo.Group) {
	a.GET("/projects/:project/views/:view/buckets", handler.WithDBAndUser(getAllBuckets, false))
	a.PUT("/projects/:project/views/:view/buckets", handler.WithDBAndUser(createBucket, true))
	a.POST("/projects/:project/views/:view/buckets/:bucket", handler.WithDBAndUser(updateBucket, true))
	a.DELETE("/projects/:project/views/:view/buckets/:bucket", handler.WithDBAndUser(deleteBucket, true))
	a.POST("/projects/:project/views/:view/buckets/:bucket/tasks", handler.WithDBAndUser(moveTaskToBucket, true))
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
