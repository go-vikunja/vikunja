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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// RegisterWebhooks registers all webhook routes
func RegisterWebhooks(a *echo.Group) {
	a.GET("/projects/:project/webhooks", handler.WithDBAndUser(getAllWebhooksLogic, false))
	a.PUT("/projects/:project/webhooks", handler.WithDBAndUser(createWebhookLogic, true))
	a.DELETE("/projects/:project/webhooks/:webhook", handler.WithDBAndUser(deleteWebhookLogic, true))
	a.POST("/projects/:project/webhooks/:webhook", handler.WithDBAndUser(updateWebhookLogic, true))
	a.GET("/webhooks/events", GetAvailableWebhookEvents)
}

// getAllWebhooksLogic retrieves all webhooks for a project.
//
// @Summary Get all webhooks
// @Description Returns all webhooks for the specified project
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param page query int false "The page number for pagination. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Success 200 {array} models.Webhook "The webhooks"
// @Failure 400 {object} web.HTTPError "Invalid project ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/webhooks [get]
func getAllWebhooksLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
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

	// Create webhook object for ReadAll
	webhook := &models.Webhook{ProjectID: projectID}

	// Use model's ReadAll method (which delegates to service)
	result, resultCount, totalItems, err := webhook.ReadAll(s, u, "", page, perPage)
	if err != nil {
		return err
	}

	// Set pagination headers
	totalPages := totalItems / int64(perPage)
	if totalItems%int64(perPage) > 0 {
		totalPages++
	}
	c.Response().Header().Set("x-pagination-total-pages", strconv.FormatInt(totalPages, 10))
	c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(resultCount))

	return c.JSON(http.StatusOK, result)
}

// createWebhookLogic creates a new webhook.
//
// @Summary Create a webhook
// @Description Creates a new webhook for the specified project
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param webhook body models.Webhook true "The webhook object"
// @Success 201 {object} models.Webhook "The created webhook"
// @Failure 400 {object} web.HTTPError "Invalid webhook object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/webhooks [put]
func createWebhookLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse webhook from request body
	var webhook models.Webhook
	if err := c.Bind(&webhook); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook object")
	}

	webhook.ProjectID = projectID

	// Use model's Create method (which delegates to service)
	err = webhook.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, webhook)
}

// deleteWebhookLogic deletes a webhook.
//
// @Summary Delete a webhook
// @Description Deletes a webhook from a project
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param webhook path int true "Webhook ID"
// @Success 200 {object} models.Message "The webhook was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid project ID or webhook ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The webhook does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/webhooks/{webhook} [delete]
func deleteWebhookLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse webhook ID
	webhookID, err := strconv.ParseInt(c.Param("webhook"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook ID")
	}

	// Create webhook object for deletion
	webhook := &models.Webhook{
		ID:        webhookID,
		ProjectID: projectID,
	}

	// Use model's Delete method (which delegates to service)
	err = webhook.Delete(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The webhook was successfully deleted."})
}

// updateWebhookLogic updates a webhook.
//
// @Summary Update a webhook
// @Description Updates a webhook
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param webhook path int true "Webhook ID"
// @Param webhookData body models.Webhook true "The webhook with updated values"
// @Success 200 {object} models.Webhook "The updated webhook"
// @Failure 400 {object} web.HTTPError "Invalid webhook object"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "The webhook does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/webhooks/{webhook} [post]
func updateWebhookLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse project ID
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID")
	}

	// Parse webhook ID
	webhookID, err := strconv.ParseInt(c.Param("webhook"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook ID")
	}

	// Parse webhook from request body
	var webhook models.Webhook
	if err := c.Bind(&webhook); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook object")
	}

	webhook.ID = webhookID
	webhook.ProjectID = projectID

	// Use model's Update method (which delegates to service)
	err = webhook.Update(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, webhook)
}

// GetAvailableWebhookEvents returns a list of all possible webhook target events
// @Summary Get all possible webhook events
// @Description Get all possible webhook events to use when creating or updating a webhook target.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} string "The list of all possible webhook events"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /webhooks/events [get]
func GetAvailableWebhookEvents(c echo.Context) error {
	return c.JSON(http.StatusOK, models.GetAvailableWebhookEvents())
}
