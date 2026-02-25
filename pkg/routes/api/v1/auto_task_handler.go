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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// TriggerAutoTask manually creates a task from an auto-task template.
// @Summary Trigger auto-task
// @tags autotask
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Auto-task template ID"
// @Success 200 {object} models.Task
// @Router /autotasks/{id}/trigger [post]
func TriggerAutoTask(c *echo.Context) error {
	templateID, err := strconv.ParseInt(c.Param("autotask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	s := db.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not start transaction")
	}

	task, err := models.TriggerAutoTaskFromAuth(s, templateID, auth)
	if err != nil {
		_ = s.Rollback()
		if _, ok := err.(models.ErrAutoTaskTemplateNotFound); ok {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	if err := s.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not commit")
	}

	return c.JSON(http.StatusOK, task)
}

// CheckAutoTasks checks all active templates for the current user and creates due tasks.
// @Summary Check and create auto-tasks
// @tags autotask
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /autotasks/check [post]
func CheckAutoTasks(c *echo.Context) error {
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	s := db.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not start transaction")
	}

	created, err := models.CheckAutoTasksFromAuth(s, auth)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := s.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not commit")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"created": created,
		"count":   len(created),
	})
}

// TruncateAutoTaskLog removes old log entries for a given auto-task template.
// @Summary Truncate auto-task log
// @tags autotask
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Auto-task template ID"
// @Param keep query int false "Number of most recent entries to keep (0 = clear all)"
// @Success 200 {object} map[string]interface{}
// @Router /autotasks/{id}/log/truncate [post]
func TruncateAutoTaskLog(c *echo.Context) error {
	templateID, err := strconv.ParseInt(c.Param("autotask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID")
	}

	keep := 0
	if keepStr := c.QueryParam("keep"); keepStr != "" {
		keep, err = strconv.Atoi(keepStr)
		if err != nil || keep < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid keep parameter")
		}
	}

	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	s := db.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not start transaction")
	}

	deleted, err := models.TruncateAutoTaskLog(s, templateID, auth.GetID(), keep)
	if err != nil {
		_ = s.Rollback()
		if _, ok := err.(models.ErrAutoTaskTemplateNotFound); ok {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := s.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not commit")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}
