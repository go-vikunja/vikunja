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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	microsofttodo "code.vikunja.io/api/pkg/modules/migration/microsoft-todo"
	"code.vikunja.io/api/pkg/modules/migration/ticktick"
	"code.vikunja.io/api/pkg/modules/migration/todoist"
	"code.vikunja.io/api/pkg/modules/migration/trello"
	vikunja_file "code.vikunja.io/api/pkg/modules/migration/vikunja-file"
	"code.vikunja.io/api/pkg/version"

	"github.com/labstack/echo/v4"
)

type vikunjaInfos struct {
	Version                    string    `json:"version"`
	FrontendURL                string    `json:"frontend_url"`
	Motd                       string    `json:"motd"`
	LinkSharingEnabled         bool      `json:"link_sharing_enabled"`
	MaxFileSize                string    `json:"max_file_size"`
	AvailableMigrators         []string  `json:"available_migrators"`
	TaskAttachmentsEnabled     bool      `json:"task_attachments_enabled"`
	EnabledBackgroundProviders []string  `json:"enabled_background_providers"`
	TotpEnabled                bool      `json:"totp_enabled"`
	Legal                      legalInfo `json:"legal"`
	CaldavEnabled              bool      `json:"caldav_enabled"`
	AuthInfo                   authInfo  `json:"auth"`
	EmailRemindersEnabled      bool      `json:"email_reminders_enabled"`
	UserDeletionEnabled        bool      `json:"user_deletion_enabled"`
	TaskCommentsEnabled        bool      `json:"task_comments_enabled"`
	DemoModeEnabled            bool      `json:"demo_mode_enabled"`
	WebhooksEnabled            bool      `json:"webhooks_enabled"`
	PublicTeamsEnabled         bool      `json:"public_teams_enabled"`
}

type authInfo struct {
	Local         localAuthInfo  `json:"local"`
	Ldap          ldapAuthInfo   `json:"ldap"`
	OpenIDConnect openIDAuthInfo `json:"openid_connect"`
}

type localAuthInfo struct {
	Enabled             bool `json:"enabled"`
	RegistrationEnabled bool `json:"registration_enabled"`
}

type ldapAuthInfo struct {
	Enabled bool `json:"enabled"`
}

type openIDAuthInfo struct {
	Enabled   bool               `json:"enabled"`
	Providers []*openid.Provider `json:"providers"`
}

type legalInfo struct {
	ImprintURL       string `json:"imprint_url"`
	PrivacyPolicyURL string `json:"privacy_policy_url"`
}

// Info is the handler to get infos about this vikunja instance
// @Summary Info
// @Description Returns the version, frontendurl, motd and various settings of Vikunja
// @tags service
// @Produce json
// @Success 200 {object} v1.vikunjaInfos
// @Router /info [get]
func Info(c echo.Context) error {
	info := vikunjaInfos{
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
		Legal: legalInfo{
			ImprintURL:       config.LegalImprintURL.GetString(),
			PrivacyPolicyURL: config.LegalPrivacyURL.GetString(),
		},
		AuthInfo: authInfo{
			Local: localAuthInfo{
				Enabled:             config.AuthLocalEnabled.GetBool(),
				RegistrationEnabled: config.AuthLocalEnabled.GetBool() && config.ServiceEnableRegistration.GetBool(),
			},
			Ldap: ldapAuthInfo{
				Enabled: config.AuthLdapEnabled.GetBool(),
			},
			OpenIDConnect: openIDAuthInfo{
				Enabled: config.AuthOpenIDEnabled.GetBool(),
			},
		},
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		log.Errorf("Error while getting openid providers for /info: %s", err)
		// No return here to not break /info
	}

	info.AuthInfo.OpenIDConnect.Providers = providers

	// Migrators
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
