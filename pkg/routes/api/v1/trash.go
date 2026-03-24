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
	"math"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v5"
)

// GetTrashedTasks returns all trashed tasks the user has access to
// @Summary Get all trashed tasks
// @Description Returns all tasks that have been soft-deleted (moved to trash).
// @tags trash
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination."
// @Param per_page query int false "The maximum number of items per page."
// @Param project_id query int false "If set, only return trashed tasks from this project."
// @Success 200 {array} models.Task "The trashed tasks"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /trash [get]
func GetTrashedTasks(c *echo.Context) error {
	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	var projectID int64
	if pid := c.QueryParam("project_id"); pid != "" {
		projectID, err = strconv.ParseInt(pid, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid project_id").Wrap(err)
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid page").Wrap(err)
		}
	}

	perPage := 50
	if pp := c.QueryParam("per_page"); pp != "" {
		perPage, err = strconv.Atoi(pp)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid per_page").Wrap(err)
		}
	}

	s := db.NewSession()
	defer s.Close()

	tasks, totalCount, err := models.GetTrashedTasks(s, a, projectID, page, perPage)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	numberOfPages := math.Ceil(float64(totalCount) / float64(perPage))
	if len(tasks) == 0 {
		numberOfPages = 0
	}

	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(numberOfPages, 'f', 0, 64))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(len(tasks)))
	c.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	return c.JSON(http.StatusOK, tasks)
}

// RestoreTrashedTask restores a task from the trash
// @Summary Restore a trashed task
// @Description Restores a previously soft-deleted task by clearing its deleted_at timestamp.
// @tags trash
// @Accept json
// @Produce json
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.Message "The task was restored."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 404 {object} web.HTTPError "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /trash/{taskID}/restore [post]
func RestoreTrashedTask(c *echo.Context) error {
	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	task := &models.Task{ID: taskID}
	canDo, err := models.CanDoTrashOperation(s, task, a)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if !canDo {
		return echo.ErrForbidden
	}

	err = task.Restore(s, a)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "success"})
}

// PermanentlyDeleteTrashedTask permanently deletes a trashed task
// @Summary Permanently delete a trashed task
// @Description Permanently deletes a task that is in the trash, removing all associated data.
// @tags trash
// @Accept json
// @Produce json
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.Message "The task was permanently deleted."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 404 {object} web.HTTPError "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /trash/{taskID} [delete]
func PermanentlyDeleteTrashedTask(c *echo.Context) error {
	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	task := &models.Task{ID: taskID}
	canDo, err := models.CanDoTrashOperation(s, task, a)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if !canDo {
		return echo.ErrForbidden
	}

	err = task.HardDelete(s, a)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "success"})
}

// EmptyTrash permanently deletes all trashed tasks the user has access to
// @Summary Empty the trash
// @Description Permanently deletes all trashed tasks the user has delete permission for.
// @tags trash
// @Accept json
// @Produce json
// @Success 200 {object} models.Message "The trash was emptied."
// @Failure 500 {object} models.Message "Internal error"
// @Router /trash [delete]
func EmptyTrash(c *echo.Context) error {
	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	count, err := models.EmptyTrash(s, a)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: strconv.FormatInt(count, 10) + " tasks permanently deleted"})
}
