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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// WebhookSettingRequest is the request body for creating/updating a webhook setting
type WebhookSettingRequest struct {
	// The target URL where webhook payloads will be sent
	TargetURL string `json:"target_url"`
	// Whether this webhook is enabled (defaults to true when target_url is provided)
	Enabled *bool `json:"enabled,omitempty"`
}

// GetUserWebhookSettings returns all webhook settings for the current user
// @Summary Get all webhook settings
// @Description Returns all webhook settings for the current user.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.UserWebhookSetting "The list of webhook settings"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks [get]
func GetUserWebhookSettings(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	setting := &models.UserWebhookSetting{}
	result, _, _, err := setting.ReadAll(s, u, "", 0, 0)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, result)
}

// GetUserWebhookSettingByType returns a specific webhook setting by notification type
// @Summary Get a webhook setting by type
// @Description Returns a specific webhook setting for the current user by notification type.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param type path string true "The notification type (e.g., 'all', 'task.reminder', 'task.undone.overdue')"
// @Success 200 {object} models.UserWebhookSetting "The webhook setting"
// @Failure 404 {object} web.HTTPError "Webhook setting not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{type} [get]
func GetUserWebhookSettingByType(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	notificationType := c.Param("type")
	setting := &models.UserWebhookSetting{
		NotificationType: notificationType,
	}

	err = setting.ReadOne(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, setting)
}

// CreateOrUpdateUserWebhookSetting creates or updates a webhook setting
// @Summary Create or update a webhook setting
// @Description Creates or updates a webhook setting for a specific notification type.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param type path string true "The notification type (e.g., 'all', 'task.reminder', 'task.undone.overdue')"
// @Param setting body WebhookSettingRequest true "The webhook setting"
// @Success 200 {object} models.UserWebhookSetting "The created/updated webhook setting"
// @Failure 400 {object} web.HTTPError "Invalid notification type or URL"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{type} [put]
func CreateOrUpdateUserWebhookSetting(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	req := &WebhookSettingRequest{}
	if err := c.Bind(req); err != nil {
		return err
	}

	notificationType := c.Param("type")

	// Default enabled to true if not explicitly set
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	setting := &models.UserWebhookSetting{
		NotificationType: notificationType,
		TargetURL:        req.TargetURL,
		Enabled:          enabled,
	}

	err = setting.Create(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, setting)
}

// DeleteUserWebhookSetting deletes a webhook setting
// @Summary Delete a webhook setting
// @Description Deletes a webhook setting for a specific notification type.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param type path string true "The notification type (e.g., 'all', 'task.reminder', 'task.undone.overdue')"
// @Success 200 {object} models.Message "Successfully deleted"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{type} [delete]
func DeleteUserWebhookSetting(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	notificationType := c.Param("type")
	setting := &models.UserWebhookSetting{
		NotificationType: notificationType,
	}

	err = setting.Delete(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "Webhook setting deleted successfully."})
}

// GetAvailableWebhookNotificationTypes returns all available webhook notification types
// @Summary Get available webhook notification types
// @Description Returns all notification types that support webhooks.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.WebhookNotificationType "The list of available webhook notification types"
// @Router /user/settings/webhooks/types [get]
func GetAvailableWebhookNotificationTypes(c *echo.Context) error {
	return c.JSON(http.StatusOK, models.AvailableWebhookNotificationTypes())
}
