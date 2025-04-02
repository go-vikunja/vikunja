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
	"time"

	"code.vikunja.io/api/pkg/modules/auth/openid"

	"code.vikunja.io/api/pkg/user"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"code.vikunja.io/api/pkg/db"

	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
)

type UserWithSettings struct {
	user.User
	Settings            *UserSettings `json:"settings"`
	DeletionScheduledAt time.Time     `json:"deletion_scheduled_at"`
	IsLocalUser         bool          `json:"is_local_user"`
	AuthProvider        string        `json:"auth_provider"`
}

// UserShow gets all information about the current user
// @Summary Get user information
// @Description Returns the current user object with their settings.
// @tags user
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} v1.UserWithSettings
// @Failure 404 {object} web.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user [get]
func UserShow(c echo.Context) error {
	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.").SetInternal(err)
	}

	s := db.NewSession()
	defer s.Close()

	u, err := models.GetUserOrLinkShareUser(s, a)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	us := &UserWithSettings{
		User: *u,
		Settings: &UserSettings{
			Name:                         u.Name,
			EmailRemindersEnabled:        u.EmailRemindersEnabled,
			DiscoverableByName:           u.DiscoverableByName,
			DiscoverableByEmail:          u.DiscoverableByEmail,
			OverdueTasksRemindersEnabled: u.OverdueTasksRemindersEnabled,
			DefaultProjectID:             u.DefaultProjectID,
			WeekStart:                    u.WeekStart,
			Language:                     u.Language,
			Timezone:                     u.Timezone,
			OverdueTasksRemindersTime:    u.OverdueTasksRemindersTime,
			FrontendSettings:             u.FrontendSettings,
			ExtraSettingsLinks:           u.ExtraSettingsLinks,
		},
		DeletionScheduledAt: u.DeletionScheduledAt,
		IsLocalUser:         u.Issuer == user.IssuerLocal,
	}

	us.AuthProvider, err = getAuthProviderName(u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, us)
}

func getAuthProviderName(u *user.User) (name string, err error) {
	if u.Issuer == user.IssuerLocal {
		return "local", nil
	}

	if u.Issuer == user.IssuerLDAP {
		return "ldap", nil
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		return "", err
	}

	for _, provider := range providers {
		issuerURL, err := provider.Issuer()
		if err != nil {
			return "", err
		}
		if issuerURL == u.Issuer {
			return provider.Name, nil
		}
	}

	return
}
