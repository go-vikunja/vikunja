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

// RegisterSubscriptions registers all subscription routes
func RegisterSubscriptions(a *echo.Group) {
	a.PUT("/subscriptions/:entity/:entityID", handler.WithDBAndUser(createSubscriptionLogic, true))
	a.DELETE("/subscriptions/:entity/:entityID", handler.WithDBAndUser(deleteSubscriptionLogic, true))
}

// createSubscriptionLogic subscribes a user to an entity (project or task).
//
// @Summary Subscribe to an entity
// @Description Subscribe the current user to a project or task to get notifications about changes.
// @tags subscriptions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entity path string true "The entity type. Can be either 'project' or 'task'."
// @Param entityID path int true "The entity ID"
// @Success 201 {object} models.Subscription "The subscription"
// @Failure 400 {object} web.HTTPError "Invalid entity type or entity ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to subscribe"
// @Failure 412 {object} web.HTTPError "The user is already subscribed to this entity."
// @Failure 500 {object} models.Message "Internal error"
// @Router /subscriptions/{entity}/{entityID} [put]
func createSubscriptionLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse entity type from URL parameter
	entityTypeStr := c.Param("entity")

	// Parse entity ID
	entityID, err := strconv.ParseInt(c.Param("entityID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid entity ID")
	}

	// Convert string entity type to enum
	var entityType models.SubscriptionEntityType
	switch entityTypeStr {
	case "project":
		entityType = models.SubscriptionEntityProject
	case "task":
		entityType = models.SubscriptionEntityTask
	default:
		return &models.ErrUnknownSubscriptionEntityType{EntityType: entityType}
	}

	// Create subscription object
	subscription := &models.Subscription{
		EntityType: entityType,
		EntityID:   entityID,
	}

	// Use model's Create method (which delegates to service)
	err = subscription.Create(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, subscription)
}

// deleteSubscriptionLogic unsubscribes a user from an entity.
//
// @Summary Unsubscribe from an entity
// @Description Unsubscribes the current user from a project or task.
// @tags subscriptions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entity path string true "The entity type. Can be either 'project' or 'task'."
// @Param entityID path int true "The entity ID"
// @Success 200 {object} models.Message "The subscription was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid entity type or entity ID"
// @Failure 403 {object} web.HTTPError "The user does not have access to unsubscribe"
// @Failure 404 {object} web.HTTPError "The subscription does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /subscriptions/{entity}/{entityID} [delete]
func deleteSubscriptionLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse entity type from URL parameter
	entityTypeStr := c.Param("entity")

	// Parse entity ID
	entityID, err := strconv.ParseInt(c.Param("entityID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid entity ID")
	}

	// Convert string entity type to enum
	var entityType models.SubscriptionEntityType
	switch entityTypeStr {
	case "project":
		entityType = models.SubscriptionEntityProject
	case "task":
		entityType = models.SubscriptionEntityTask
	default:
		return &models.ErrUnknownSubscriptionEntityType{EntityType: entityType}
	}

	// Create subscription object for deletion
	subscription := &models.Subscription{
		EntityType: entityType,
		EntityID:   entityID,
	}

	// Use model's Delete method (which delegates to service)
	err = subscription.Delete(s, u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The subscription was successfully deleted."})
}
