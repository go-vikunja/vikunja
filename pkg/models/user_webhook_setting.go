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

package models

import (
	"errors"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

func init() {
	// Register the webhook URL lookup function with the notifications package
	notifications.SetWebhookURLLookup(lookupWebhookURL)
}

// lookupWebhookURL looks up the webhook URL for a user and notification type with fallback to "all"
func lookupWebhookURL(userID int64, notificationType string) (string, error) {
	setting, err := GetUserWebhookSettingByTypeWithFallback(userID, notificationType)
	if err != nil {
		return "", err
	}
	if setting == nil || !setting.Enabled || setting.TargetURL == "" {
		return "", nil
	}
	return setting.TargetURL, nil
}

// UserWebhookSetting represents a per-notification-type webhook configuration for a user
type UserWebhookSetting struct {
	// The unique, numeric id of this webhook setting
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"type"`
	// The user this setting belongs to
	UserID int64 `xorm:"bigint not null unique(user_notification_type)" json:"user_id"`
	// The notification type this setting applies to (e.g., "task.reminder", "task.undone.overdue", or "all" for general fallback)
	NotificationType string `xorm:"varchar(100) not null unique(user_notification_type)" json:"notification_type"`
	// Whether this webhook is enabled
	Enabled bool `xorm:"bool default true" json:"enabled"`
	// The target URL where webhook payloads will be sent
	TargetURL string `xorm:"text not null" json:"target_url"`
	// A timestamp when this setting was created
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this setting was last updated
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
}

// TableName returns the table name for user webhook settings
func (*UserWebhookSetting) TableName() string {
	return "user_webhook_settings"
}

// WebhookNotificationType represents an available webhook notification type
type WebhookNotificationType struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// AvailableWebhookNotificationTypes returns all notification types that support webhooks
func AvailableWebhookNotificationTypes() []WebhookNotificationType {
	return []WebhookNotificationType{
		{Type: "all", Description: "General webhook for all notification types (fallback)"},
		{Type: "task.reminder", Description: "Task reminder notifications"},
		{Type: "task.undone.overdue", Description: "Overdue task notifications"},
		// Task notifications
		// {Type: "task.comment", Description: "Task comment notifications"},
		// {Type: "task.assigned", Description: "Task assigned notifications"},
		// {Type: "task.deleted", Description: "Task deleted notifications"},
		// {Type: "task.mentioned", Description: "User mentioned in task notifications"},
		// Project notifications
		// {Type: "project.created", Description: "Project created notifications"},
		// Team notifications
		// {Type: "team.member.added", Description: "Team member added notifications"},
		// Data export notifications
		// {Type: "data.export.ready", Description: "Data export ready notifications"},
		// Migration notifications
		// {Type: "migration.done", Description: "Migration completed notifications"},
		// {Type: "migration.failed", Description: "Migration failed notifications"},
		// {Type: "migration.failed.reported", Description: "Migration failure reported notifications"},
		// User account notifications
		// {Type: "totp.invalid", Description: "Invalid TOTP code notifications"},
		// {Type: "password.account.locked.after.invalid.totp", Description: "Account locked after invalid TOTP notifications"},
		// {Type: "failed.login.attempt", Description: "Failed login attempt notifications"},
		// {Type: "user.deletion.confirm", Description: "Account deletion confirmation notifications"},
		// {Type: "user.deletion", Description: "Account deletion scheduled notifications"},
		// {Type: "user.deleted", Description: "Account deleted notifications"},
	}
}

// isValidNotificationType checks if a notification type is valid
func isValidNotificationType(notificationType string) bool {
	for _, t := range AvailableWebhookNotificationTypes() {
		if t.Type == notificationType {
			return true
		}
	}
	return false
}

// GetUserWebhookSettingByType retrieves a webhook setting for a specific user and notification type
func GetUserWebhookSettingByType(s *xorm.Session, userID int64, notificationType string) (*UserWebhookSetting, error) {
	setting := &UserWebhookSetting{}
	has, err := s.Where("user_id = ? AND notification_type = ?", userID, notificationType).Get(setting)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return setting, nil
}

// GetUserWebhookSettingByTypeWithFallback retrieves a webhook setting with fallback to "all" type
func GetUserWebhookSettingByTypeWithFallback(userID int64, notificationType string) (*UserWebhookSetting, error) {
	s := db.NewSession()
	defer s.Close()

	// Try specific type first
	setting, err := GetUserWebhookSettingByType(s, userID, notificationType)
	if err != nil {
		return nil, err
	}

	// If no specific setting or disabled, try "all" fallback
	if setting == nil || !setting.Enabled || setting.TargetURL == "" {
		setting, err = GetUserWebhookSettingByType(s, userID, "all")
		if err != nil {
			return nil, err
		}
	}

	return setting, nil
}

// ReadAll returns all webhook settings for the current user
// @Summary Get all webhook settings
// @Description Returns all webhook settings for the current user.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.UserWebhookSetting "The list of webhook settings"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks [get]
func (s *UserWebhookSetting) ReadAll(sess *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	settings := []*UserWebhookSetting{}

	err = sess.Where("user_id = ?", a.GetID()).Find(&settings)
	if err != nil {
		return nil, 0, 0, err
	}

	return settings, len(settings), int64(len(settings)), nil
}

// ReadOne returns a specific webhook setting by notification type
// @Summary Get a webhook setting by type
// @Description Returns a specific webhook setting for the current user by notification type.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param type path string true "The notification type"
// @Success 200 {object} models.UserWebhookSetting "The webhook setting"
// @Failure 404 {object} web.HTTPError "Webhook setting not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{type} [get]
func (s *UserWebhookSetting) ReadOne(sess *xorm.Session, a web.Auth) error {
	setting, err := GetUserWebhookSettingByType(sess, a.GetID(), s.NotificationType)
	if err != nil {
		return err
	}
	if setting == nil {
		return ErrUserWebhookSettingDoesNotExist{NotificationType: s.NotificationType}
	}

	*s = *setting
	return nil
}

// Create creates or updates a webhook setting
// @Summary Create or update a webhook setting
// @Description Creates or updates a webhook setting for a specific notification type.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param type path string true "The notification type"
// @Param setting body models.UserWebhookSetting true "The webhook setting"
// @Success 200 {object} models.UserWebhookSetting "The created/updated webhook setting"
// @Failure 400 {object} web.HTTPError "Invalid notification type or URL"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{type} [put]
func (s *UserWebhookSetting) Create(sess *xorm.Session, a web.Auth) error {
	// Validate notification type
	if !isValidNotificationType(s.NotificationType) {
		return ErrInvalidWebhookNotificationType{NotificationType: s.NotificationType}
	}

	// Validate URL
	if s.TargetURL == "" {
		return ErrWebhookURLRequired{}
	}

	s.UserID = a.GetID()

	// Auto-enable when URL is provided (unless explicitly set to false)
	// The enabled field defaults to true, so this is the expected behavior

	// Check if setting already exists
	existing, err := GetUserWebhookSettingByType(sess, s.UserID, s.NotificationType)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing setting
		s.ID = existing.ID
		_, err = sess.ID(existing.ID).Cols("enabled", "target_url", "updated").Update(s)
		return err
	}

	// Create new setting
	_, err = sess.Insert(s)
	return err
}

// Update updates a webhook setting (alias for Create since we use upsert logic)
func (s *UserWebhookSetting) Update(sess *xorm.Session, a web.Auth) error {
	return s.Create(sess, a)
}

// Delete deletes a webhook setting
// @Summary Delete a webhook setting
// @Description Deletes a webhook setting for a specific notification type.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param type path string true "The notification type"
// @Success 200 {object} models.Message "Successfully deleted"
// @Failure 404 {object} web.HTTPError "Webhook setting not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{type} [delete]
func (s *UserWebhookSetting) Delete(sess *xorm.Session, a web.Auth) error {
	_, err := sess.Where("user_id = ? AND notification_type = ?", a.GetID(), s.NotificationType).Delete(&UserWebhookSetting{})
	return err
}

// CanRead checks if the user can read this webhook setting
func (s *UserWebhookSetting) CanRead(_ *xorm.Session, a web.Auth) (bool, int, error) {
	return s.UserID == a.GetID(), int(PermissionRead), nil
}

// CanUpdate checks if the user can update this webhook setting
func (s *UserWebhookSetting) CanUpdate(_ *xorm.Session, a web.Auth) (bool, error) {
	return s.UserID == a.GetID() || s.UserID == 0, nil
}

// CanDelete checks if the user can delete this webhook setting
func (s *UserWebhookSetting) CanDelete(_ *xorm.Session, a web.Auth) (bool, error) {
	return s.UserID == a.GetID() || s.UserID == 0, nil
}

// CanCreate checks if the user can create this webhook setting
func (s *UserWebhookSetting) CanCreate(_ *xorm.Session, _ web.Auth) (bool, error) {
	return true, nil
}

// Error types for webhook settings

// ErrUserWebhookSettingDoesNotExist represents an error when a webhook setting is not found
type ErrUserWebhookSettingDoesNotExist struct {
	NotificationType string
}

func (e ErrUserWebhookSettingDoesNotExist) Error() string {
	return "Webhook setting not found for notification type: " + e.NotificationType
}

// IsErrUserWebhookSettingDoesNotExist checks if an error is ErrUserWebhookSettingDoesNotExist
func IsErrUserWebhookSettingDoesNotExist(err error) bool {
	var errUserWebhookSettingDoesNotExist ErrUserWebhookSettingDoesNotExist
	ok := errors.As(err, &errUserWebhookSettingDoesNotExist)
	return ok
}

// HTTPError returns the HTTP error code for this error
func (e ErrUserWebhookSettingDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 404,
		Code:     10001,
		Message:  "Webhook setting not found for notification type: " + e.NotificationType,
	}
}

// ErrInvalidWebhookNotificationType represents an error when an invalid notification type is provided
type ErrInvalidWebhookNotificationType struct {
	NotificationType string
}

func (e ErrInvalidWebhookNotificationType) Error() string {
	return "Invalid webhook notification type: " + e.NotificationType
}

// IsErrInvalidWebhookNotificationType checks if an error is ErrInvalidWebhookNotificationType
func IsErrInvalidWebhookNotificationType(err error) bool {
	var errInvalidWebhookNotificationType ErrInvalidWebhookNotificationType
	ok := errors.As(err, &errInvalidWebhookNotificationType)
	return ok
}

// HTTPError returns the HTTP error code for this error
func (e ErrInvalidWebhookNotificationType) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 400,
		Code:     10002,
		Message:  "Invalid webhook notification type: " + e.NotificationType,
	}
}

// ErrWebhookURLRequired represents an error when webhook URL is required but not provided
type ErrWebhookURLRequired struct{}

func (e ErrWebhookURLRequired) Error() string {
	return "Webhook URL is required"
}

// IsErrWebhookURLRequired checks if an error is ErrWebhookURLRequired
func IsErrWebhookURLRequired(err error) bool {
	var errWebhookURLRequired ErrWebhookURLRequired
	ok := errors.As(err, &errWebhookURLRequired)
	return ok
}

// HTTPError returns the HTTP error code for this error
func (e ErrWebhookURLRequired) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 400,
		Code:     10003,
		Message:  "Webhook URL is required",
	}
}
