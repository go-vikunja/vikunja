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
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/tkuchiki/go-timezone"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
)

// UserAvatarProvider holds the user avatar provider type
type UserAvatarProvider struct {
	// The avatar provider. Valid types are `gravatar` (uses the user email), `upload`, `initials`, `marble` (generates a random avatar for each user), `ldap` (synced from LDAP server), `openid` (synced from OpenID provider), `default`.
	AvatarProvider string `json:"avatar_provider"`
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
func GetUserAvatarProvider(c *echo.Context) error {

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserWithEmail(s, &user2.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	uap := &UserAvatarProvider{AvatarProvider: user.AvatarProvider}
	return c.JSON(http.StatusOK, uap)
}

// ChangeUserAvatarProvider changes the user's avatar provider
// @Summary Set the user's avatar
// @Description Changes the user avatar. Valid types are gravatar (uses the user email), upload, initials, marble, ldap (synced from LDAP server), openid (synced from OpenID provider), default.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param avatar body UserAvatarProvider true "The user's avatar setting"
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/avatar [post]
func ChangeUserAvatarProvider(c *echo.Context) error {

	uap := &UserAvatarProvider{}
	err := c.Bind(uap)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad avatar type provided.").Wrap(err)
	}

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserWithEmail(s, &user2.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := models.UpdateUserAvatarProvider(s, user, uap.AvatarProvider); err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "Avatar was changed successfully."})
}

// UpdateGeneralUserSettings is the handler to change general user settings
// @Summary Change general user settings of the current user.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param avatar body models.UserGeneralSettings true "The updated user settings"
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/general [post]
func UpdateGeneralUserSettings(c *echo.Context) error {
	us := &models.UserGeneralSettings{}
	err := c.Bind(us)
	if err != nil {
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return models.ErrInvalidModel{Message: fmt.Sprintf("%v", he.Message), Err: err}
		}
		return models.ErrInvalidModel{Err: err}
	}

	err = c.Validate(us)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).Wrap(err)
	}

	u, err := user2.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	user, err := user2.GetUserWithEmail(s, &user2.User{ID: u.ID})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := models.UpdateUserGeneralSettings(s, user, us); err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "The settings were updated successfully."})
}

// GetAvailableTimezones
// @Summary Get all available time zones on this vikunja instance
// @Description Because available time zones depend on the system Vikunja is running on, this endpoint returns a project of all valid time zones this particular Vikunja instance can handle. The project of time zones is not sorted, you should sort it on the client.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} string "All available time zones."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/timezones [get]
func GetAvailableTimezones(c *echo.Context) error {

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
