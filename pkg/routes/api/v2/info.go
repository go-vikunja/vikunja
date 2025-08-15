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

package v2

import (
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/modules/migration/microsoft-todo"
	"code.vikunja.io/api/pkg/modules/migration/ticktick"
	"code.vikunja.io/api/pkg/modules/migration/todoist"
	"code.vikunja.io/api/pkg/modules/migration/trello"
	vikunja_file "code.vikunja.io/api/pkg/modules/migration/vikunja-file"
	"code.vikunja.io/api/pkg/version"
	"github.com/labstack/echo/v4"
	"github.com/tkuchiki/go-timezone"
)

// GetInfo is the handler to get infos about this vikunja instance
func GetInfo(c echo.Context) error {
	info := &v2.Info{
		Version:                version.Version,
		FrontendURL:            config.ServicePublicURL.GetString(),
		Motd:                   config.ServiceMotd.GetString(),
		LinkSharingEnabled:     config.ServiceEnableLinkSharing.GetBool(),
		MaxFileSize:            config.FilesMaxSize.GetString(),
		TaskAttachmentsEnabled: config.ServiceEnableTaskAttachments.GetBool(),
		TotpEnabled:            config.ServiceEnableTotp.GetBool(),
		CaldavEnabled:          config.ServiceEnableCaldav.GetBool(),
		EmailRemindersEnabled:  config.ServiceEnableEmailReminders.GetBool(),
		UserDeletionEnabled:    config.ServiceEnableUserDeletion.GetBool(),
		TaskCommentsEnabled:    config.ServiceEnableTaskComments.GetBool(),
		DemoModeEnabled:        config.ServiceDemoMode.GetBool(),
		WebhooksEnabled:        config.WebhooksEnabled.GetBool(),
		PublicTeamsEnabled:     config.ServiceEnablePublicTeams.GetBool(),
		AvailableMigrators: []string{
			(&vikunja_file.FileMigrator{}).Name(),
			(&ticktick.Migrator{}).Name(),
		},
		Legal: v2.LegalInfo{
			ImprintURL:       config.LegalImprintURL.GetString(),
			PrivacyPolicyURL: config.LegalPrivacyURL.GetString(),
		},
		Auth: v2.AuthInfo{
			Local: v2.LocalAuthInfo{
				Enabled:             config.AuthLocalEnabled.GetBool(),
				RegistrationEnabled: config.AuthLocalEnabled.GetBool() && config.ServiceEnableRegistration.GetBool(),
			},
			Ldap: v2.LDAPAuthInfo{
				Enabled: config.AuthLdapEnabled.GetBool(),
			},
			OpenIDConnect: v2.OpenIDAuthInfo{
				Enabled: config.AuthOpenIDEnabled.GetBool(),
			},
		},
		Links: &v2.InfoLinks{
			Self: &v2.Link{Href: "/api/v2/info"},
		},
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		log.Errorf("Error while getting openid providers for /info: %s", err)
	}
	info.Auth.OpenIDConnect.Providers = providers

	if config.MigrationTodoistEnable.GetBool() {
		m := &todoist.Migration{}
		info.AvailableMigrators = append(info.AvailableMigrators, m.Name())
	}
	if config.MigrationTrelloEnable.GetBool() {
		m := &trello.Migration{}
		info.AvailableMigrators = append(info.AvailableMigrators, m.Name())
	}
	if config.MigrationMicrosoftTodoEnable.GetBool() {
		m := &microsofttodo.Migration{}
		info.AvailableMigrators = append(info.AvailableMigrators, m.Name())
	}

	if config.BackgroundsEnabled.GetBool() {
		if config.BackgroundsUploadEnabled.GetBool() {
			info.EnabledBackgroundProviders = append(info.EnabledBackgroundProviders, "upload")
		}
		if config.BackgroundsUnsplashEnabled.GetBool() {
			info.EnabledBackgroundProviders = append(info.EnabledBackgroundProviders, "unsplash")
		}
	}

	return c.JSON(http.StatusOK, info)
}

// GetTimezones returns all available timezones
func GetTimezones(c echo.Context) error {
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

// GetWebhookEvents returns all available webhook events
func GetWebhookEvents(c echo.Context) error {
	return c.JSON(http.StatusOK, models.GetAvailableWebhookEvents())
}
