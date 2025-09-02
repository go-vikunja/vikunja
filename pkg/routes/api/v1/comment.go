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

// CommentRoutes defines all task comment related routes
var CommentRoutes = []APIRoute{
	{Method: "GET", Path: "/tasks/:task/comments", Handler: handler.WithDBAndUser(getAllCommentsLogic, false), PermissionScope: "read_all"},
	{Method: "PUT", Path: "/tasks/:task/comments", Handler: handler.WithDBAndUser(createCommentLogic, true), PermissionScope: "create"},
	{Method: "GET", Path: "/tasks/:task/comments/:commentid", Handler: handler.WithDBAndUser(getCommentLogic, false), PermissionScope: "read_one"},
	{Method: "POST", Path: "/tasks/:task/comments/:commentid", Handler: handler.WithDBAndUser(updateCommentLogic, true), PermissionScope: "update"},
	{Method: "DELETE", Path: "/tasks/:task/comments/:commentid", Handler: handler.WithDBAndUser(deleteCommentLogic, true), PermissionScope: "delete"},
}

// RegisterComments registers all comment routes
func RegisterComments(a *echo.Group) {
	registerRoutes(a, CommentRoutes)
}

// getAllCommentsLogic handles retrieving all comments for a task
func getAllCommentsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	commentService := services.NewCommentService(s.Engine())

	search := c.QueryParam("s")
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 {
		perPage = 50
	}

	comments, resultCount, totalItems, err := commentService.GetAllForTask(s, taskID, u, search, page, perPage)
	if err != nil {
		return err
	}

	// Set pagination headers
	if totalItems > 0 {
		totalPages := (totalItems + int64(perPage) - 1) / int64(perPage)
		c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalPages, 10))
		c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))
	}

	return c.JSON(http.StatusOK, comments)
}

// createCommentLogic creates a new comment for a task
func createCommentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	comment := new(models.TaskComment)
	if err := c.Bind(comment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment data").SetInternal(err)
	}

	comment.TaskID = taskID

	commentService := services.NewCommentService(s.Engine())
	result, err := commentService.Create(s, comment, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, result)
}

// getCommentLogic handles retrieving a single comment
func getCommentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	commentID, err := strconv.ParseInt(c.Param("commentid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID").SetInternal(err)
	}

	commentService := services.NewCommentService(s.Engine())
	comment, err := commentService.GetByID(s, commentID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, comment)
}

// updateCommentLogic handles updating an existing comment
func updateCommentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").SetInternal(err)
	}

	commentID, err := strconv.ParseInt(c.Param("commentid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID").SetInternal(err)
	}

	comment := new(models.TaskComment)
	if err := c.Bind(comment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment data").SetInternal(err)
	}

	comment.TaskID = taskID
	comment.ID = commentID

	commentService := services.NewCommentService(s.Engine())
	result, err := commentService.Update(s, comment, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, result)
}

// deleteCommentLogic handles deleting a comment
func deleteCommentLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	commentID, err := strconv.ParseInt(c.Param("commentid"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID").SetInternal(err)
	}

	commentService := services.NewCommentService(s.Engine())
	err = commentService.Delete(s, commentID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The comment was deleted successfully."})
}
