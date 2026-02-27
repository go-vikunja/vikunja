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
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// GetUserWebhooks returns all webhook targets for the current user
// @Summary Get all user-level webhook targets
// @Description Get all webhook targets configured for the current user (not project-specific).
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.Webhook "The list of webhook targets"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks [get]
func GetUserWebhooks(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	ws := []*models.Webhook{}
	err = s.Where("user_id = ?", u.ID).Find(&ws)
	if err != nil {
		return err
	}

	userIDs := []int64{}
	for _, w := range ws {
		userIDs = append(userIDs, w.CreatedByID)
	}

	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return err
	}

	for _, w := range ws {
		w.Secret = ""
		if createdBy, has := users[w.CreatedByID]; has {
			w.CreatedBy = createdBy
		}
	}

	return c.JSON(http.StatusOK, ws)
}

// CreateUserWebhook creates a new user-level webhook target
// @Summary Create a user-level webhook target
// @Description Create a webhook target for the current user that receives events across all projects.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param webhook body models.Webhook true "The webhook target"
// @Success 200 {object} models.Webhook "The created webhook target"
// @Failure 400 {object} web.HTTPError "Invalid webhook"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks [put]
func CreateUserWebhook(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	w := &models.Webhook{}
	if err := c.Bind(w); err != nil {
		return err
	}

	// Force user-level webhook
	w.UserID = u.ID
	w.ProjectID = 0

	s := db.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return err
	}

	err = w.Create(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, w)
}

// UpdateUserWebhook updates a user-level webhook target
// @Summary Update a user-level webhook target
// @Description Update the events for a user-level webhook target.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Webhook ID"
// @Success 200 {object} models.Webhook "The updated webhook target"
// @Failure 404 {object} web.HTTPError "Webhook not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{id} [post]
func UpdateUserWebhook(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	webhookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrNotFound
	}

	w := &models.Webhook{}
	if err := c.Bind(w); err != nil {
		return err
	}

	// Use path param as canonical ID
	w.ID = webhookID

	s := db.NewSession()
	defer s.Close()

	// Verify webhook belongs to user
	existing := &models.Webhook{}
	has, err := s.Where("id = ? AND user_id = ?", w.ID, u.ID).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		return echo.ErrNotFound
	}

	if err := s.Begin(); err != nil {
		return err
	}

	err = w.Update(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, w)
}

// DeleteUserWebhook deletes a user-level webhook target
// @Summary Delete a user-level webhook target
// @Description Delete a user-level webhook target.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Webhook ID"
// @Success 200 {object} models.Message "Successfully deleted"
// @Failure 404 {object} web.HTTPError "Webhook not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{id} [delete]
func DeleteUserWebhook(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	webhookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.ErrNotFound
	}

	// Verify webhook belongs to user
	existing := &models.Webhook{}
	has, err := s.Where("id = ? AND user_id = ?", webhookID, u.ID).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		return echo.ErrNotFound
	}

	if err := s.Begin(); err != nil {
		return err
	}

	err = existing.Delete(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "Successfully deleted."})
}

// GetUserDirectedWebhookEvents returns events available for user-level webhooks
// @Summary Get available user-directed webhook events
// @Description Get all webhook events that can be used with user-level webhook targets.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} string "The list of user-directed webhook events"
// @Router /user/settings/webhooks/events [get]
func GetUserDirectedWebhookEvents(c *echo.Context) error {
	return c.JSON(http.StatusOK, models.GetUserDirectedWebhookEvents())
}
