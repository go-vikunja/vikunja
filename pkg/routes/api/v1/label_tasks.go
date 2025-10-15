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

// RegisterLabelTasks registers all label-task routes
func RegisterLabelTasks(a *echo.Group) {
	a.PUT("/tasks/:projecttask/labels", handler.WithDBAndUser(addLabelToTaskLogic, true))
	a.DELETE("/tasks/:projecttask/labels/:label", handler.WithDBAndUser(removeLabelFromTaskLogic, true))
	a.GET("/tasks/:projecttask/labels", handler.WithDBAndUser(getTaskLabelsLogic, false))
	a.POST("/tasks/:projecttask/labels/bulk", handler.WithDBAndUser(updateTaskLabelsLogic, true))
}

// addLabelToTaskLogic adds a single label to a task.
//
// @Summary Add a label to a task
// @Description Adds a label to a task. The user needs to have write access to the task and read access to the label.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param label body models.LabelTask true "The label task relation object with the label ID to add"
// @Success 201 {object} models.Label "The label was successfully added to the task."
// @Failure 400 {object} web.HTTPError "Invalid task ID or label object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task or label"
// @Failure 404 {object} web.HTTPError "The task or label does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/labels [put]
func addLabelToTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse label task relation from request body
	var labelTask models.LabelTask
	if err := c.Bind(&labelTask); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object")
	}

	if labelTask.LabelID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Label ID is required")
	}

	// Add label to task via service
	service := services.NewLabelService(s.Engine())
	err = service.AddLabelToTask(s, labelTask.LabelID, taskID, u)
	if err != nil {
		return err
	}

	// Return the added label
	addedLabel, err := service.Get(s, labelTask.LabelID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, addedLabel)
}

// removeLabelFromTaskLogic removes a label from a task.
//
// @Summary Remove a label from a task
// @Description Removes a label from a task. The user needs to have write access to the task.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param label path int true "Label ID"
// @Success 200 {object} models.Message "The label was successfully removed from the task."
// @Failure 400 {object} web.HTTPError "Invalid task ID or label ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/labels/{label} [delete]
func removeLabelFromTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse label ID
	labelID, err := strconv.ParseInt(c.Param("label"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID")
	}

	// Remove label from task via service
	service := services.NewLabelService(s.Engine())
	err = service.RemoveLabelFromTask(s, labelID, taskID, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The label was successfully removed from the task."})
}

// getTaskLabelsLogic retrieves all labels for a task.
//
// @Summary Get all labels for a task
// @Description Returns all labels that are associated with a task.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param page query int false "The page number for pagination. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search labels by title."
// @Success 200 {array} models.Label "All labels for the task."
// @Failure 400 {object} web.HTTPError "Invalid task ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/labels [get]
func getTaskLabelsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 50
	}

	search := c.QueryParam("s")

	// Get labels from service
	service := services.NewLabelService(s.Engine())
	opts := &services.GetLabelsByTaskIDsOptions{
		User:    u,
		TaskIDs: []int64{taskID},
		Search:  []string{search},
		Page:    page,
		PerPage: perPage,
	}

	labels, resultCount, totalItems, err := service.GetLabelsByTaskIDs(s, opts)
	if err != nil {
		return err
	}

	// Extract just the labels (remove task ID grouping)
	uniqueLabels := make([]*models.Label, 0, len(labels))
	seen := make(map[int64]bool)
	for _, labelWithTaskID := range labels {
		if !seen[labelWithTaskID.ID] {
			seen[labelWithTaskID.ID] = true
			uniqueLabels = append(uniqueLabels, &labelWithTaskID.Label)
		}
	}

	// Set pagination headers
	totalPages := totalItems / int64(perPage)
	if totalItems%int64(perPage) > 0 {
		totalPages++
	}
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalPages, 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, uniqueLabels)
}

// updateTaskLabelsLogic performs bulk label updates on a task.
//
// @Summary Bulk update labels on a task
// @Description Updates all labels on a task. This will remove all existing labels and add the new ones.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projecttask path int true "Task ID"
// @Param labels body models.LabelTaskBulk true "The label IDs to set on the task"
// @Success 200 {array} models.Label "The updated labels."
// @Failure 400 {object} web.HTTPError "Invalid task ID or label objects"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task or labels"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{projecttask}/labels/bulk [post]
func updateTaskLabelsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse task ID
	taskID, err := strconv.ParseInt(c.Param("projecttask"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	// Parse labels from request body
	var bulkUpdate models.LabelTaskBulk
	if err := c.Bind(&bulkUpdate); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label objects")
	}

	// Update task labels via service
	service := services.NewLabelService(s.Engine())
	err = service.UpdateTaskLabels(s, taskID, bulkUpdate.Labels, u)
	if err != nil {
		return err
	}

	// Return the updated labels
	opts := &services.GetLabelsByTaskIDsOptions{
		User:    u,
		TaskIDs: []int64{taskID},
	}
	labels, _, _, err := service.GetLabelsByTaskIDs(s, opts)
	if err != nil {
		return err
	}

	// Extract just the labels
	uniqueLabels := make([]*models.Label, 0, len(labels))
	seen := make(map[int64]bool)
	for _, labelWithTaskID := range labels {
		if !seen[labelWithTaskID.ID] {
			seen[labelWithTaskID.ID] = true
			uniqueLabels = append(uniqueLabels, &labelWithTaskID.Label)
		}
	}

	return c.JSON(http.StatusOK, uniqueLabels)
}
