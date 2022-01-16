// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tkuchiki/go-timezone"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
)

// UserAvatarProvider holds the user avatar provider type
type UserAvatarProvider struct {
	// The avatar provider. Valid types are `gravatar` (uses the user email), `upload`, `initials`, `marble` (generates a random avatar for each user), `default`.
	AvatarProvider string `json:"avatar_provider"`
}

// UserSettings holds all user settings
type UserSettings struct {
	// The new name of the current user.
	Name string `json:"name"`
	// If enabled, sends email reminders of tasks to the user.
	EmailRemindersEnabled bool `json:"email_reminders_enabled"`
	// If true, this user can be found by their name or parts of it when searching for it.
	DiscoverableByName bool `json:"discoverable_by_name"`
	// If true, the user can be found when searching for their exact email.
	DiscoverableByEmail bool `json:"discoverable_by_email"`
	// If enabled, the user will get an email for their overdue tasks each morning.
	OverdueTasksRemindersEnabled bool `json:"overdue_tasks_reminders_enabled"`
	// If a task is created without a specified list this value should be used. Applies
	// to tasks made directly in API and from clients.
	DefaultListID int64 `json:"default_list_id"`
	// The day when the week starts for this user. 0 = sunday, 1 = monday, etc.
	WeekStart int `json:"week_start"`
	// The user's language
	Language string `json:"language"`
	// The user's time zone. Used to send task reminders in the time zone of the user.
	Timezone string `json:"timezone"`
}

// GetUserAvatarProvider returns the currently set user avatar
// @Summary Return user avatar setting
// @Description Returns the current user's avatar setting.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} UserAvatarProvider
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/avatar [get]
func GetUserAvatarProvider(c echo.Context) error {

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserWithEmail(s, &user2.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	uap := &UserAvatarProvider{AvatarProvider: user.AvatarProvider}
	return c.JSON(http.StatusOK, uap)
}

// ChangeUserAvatarProvider changes the user's avatar provider
// @Summary Set the user's avatar
// @Description Changes the user avatar. Valid types are gravatar (uses the user email), upload, initials, default.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param avatar body UserAvatarProvider true "The user's avatar setting"
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/avatar [post]
func ChangeUserAvatarProvider(c echo.Context) error {

	uap := &UserAvatarProvider{}
	err := c.Bind(uap)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad avatar type provided.")
	}

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserWithEmail(s, &user2.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	user.AvatarProvider = uap.AvatarProvider

	_, err = user2.UpdateUser(s, user)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "Avatar was changed successfully."})
}

// UpdateGeneralUserSettings is the handler to change general user settings
// @Summary Change general user settings of the current user.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param avatar body UserSettings true "The updated user settings"
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/general [post]
func UpdateGeneralUserSettings(c echo.Context) error {
	us := &UserSettings{}
	err := c.Bind(us)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad user name provided.")
	}

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserWithEmail(s, &user2.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	user.Name = us.Name
	user.EmailRemindersEnabled = us.EmailRemindersEnabled
	user.DiscoverableByEmail = us.DiscoverableByEmail
	user.DiscoverableByName = us.DiscoverableByName
	user.OverdueTasksRemindersEnabled = us.OverdueTasksRemindersEnabled
	user.DefaultListID = us.DefaultListID
	user.WeekStart = us.WeekStart
	user.Language = us.Language
	user.Timezone = us.Timezone

	_, err = user2.UpdateUser(s, user)
	if err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return handler.HandleHTTPError(err, c)
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "The settings were updated successfully."})
}

// GetAvailableTimezones
// @Summary Get all available time zones on this vikunja instance
// @Description Because available time zones depend on the system Vikunja is running on, this endpoint returns a list of all valid time zones this particular Vikunja instance can handle. The list of time zones is not sorted, you should sort it on the client.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} string "All available time zones."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/timezones [get]
func GetAvailableTimezones(c echo.Context) error {

	allTimezones := timezone.New().Timezones()
	timezoneMap := make(map[string]bool) // to filter all duplicates
	for _, s := range allTimezones {
		for _, t := range s {
			timezoneMap[t] = true
		}
	}

	ts := []string{}
	for s := range timezoneMap {
		ts = append(ts, s)
	}

	return c.JSON(http.StatusOK, ts)
}
