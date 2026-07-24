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

package shared

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/modules/migration/clickup"
	csvmigrator "code.vikunja.io/api/pkg/modules/migration/csv"
	microsofttodo "code.vikunja.io/api/pkg/modules/migration/microsoft-todo"
	"code.vikunja.io/api/pkg/modules/migration/ticktick"
	"code.vikunja.io/api/pkg/modules/migration/todoist"
	"code.vikunja.io/api/pkg/modules/migration/trello"
	vikunja_file "code.vikunja.io/api/pkg/modules/migration/vikunja-file"
	"code.vikunja.io/api/pkg/modules/migration/wekan"
	"code.vikunja.io/api/pkg/version"
)

// VikunjaInfos holds public information about this Vikunja instance.
type VikunjaInfos struct {
	Version                    string            `json:"version" doc:"The Vikunja version this instance runs."`
	FrontendURL                string            `json:"frontend_url" doc:"The publicly configured frontend URL of this instance."`
	Motd                       string            `json:"motd" doc:"The message of the day, shown to all users."`
	LinkSharingEnabled         bool              `json:"link_sharing_enabled" doc:"Whether sharing projects via public links is enabled."`
	MaxFileSize                string            `json:"max_file_size" doc:"The maximum allowed upload size, as a human-readable string (e.g. 20MB)."`
	MaxItemsPerPage            int               `json:"max_items_per_page" doc:"The maximum number of items a paginated endpoint returns per page."`
	AvailableMigrators         []string          `json:"available_migrators" doc:"The migrators enabled on this instance."`
	TaskAttachmentsEnabled     bool              `json:"task_attachments_enabled" doc:"Whether task attachments are enabled."`
	EnabledBackgroundProviders []string          `json:"enabled_background_providers" doc:"The project-background providers enabled on this instance (e.g. upload, unsplash)."`
	TotpEnabled                bool              `json:"totp_enabled" doc:"Whether TOTP two-factor authentication is enabled."`
	Legal                      LegalInfo         `json:"legal" doc:"Links to the instance's legal documents."`
	CaldavEnabled              bool              `json:"caldav_enabled" doc:"Whether the CalDAV interface is enabled."`
	AuthInfo                   AuthInfo          `json:"auth" doc:"The authentication methods enabled on this instance."`
	EmailRemindersEnabled      bool              `json:"email_reminders_enabled" doc:"Whether email reminders are enabled."`
	UserDeletionEnabled        bool              `json:"user_deletion_enabled" doc:"Whether users may delete their own account."`
	TaskCommentsEnabled        bool              `json:"task_comments_enabled" doc:"Whether task comments are enabled."`
	DemoModeEnabled            bool              `json:"demo_mode_enabled" doc:"Whether this instance runs in demo mode (data is periodically reset)."`
	WebhooksEnabled            bool              `json:"webhooks_enabled" doc:"Whether webhooks are enabled."`
	PublicTeamsEnabled         bool              `json:"public_teams_enabled" doc:"Whether public teams are enabled."`
	AllowIconChanges           bool              `json:"allow_icon_changes" doc:"Whether users may change project icons."`
	EnabledProFeatures         []license.Feature `json:"enabled_pro_features" doc:"The licensed pro features enabled on this instance."`
	// ConcurrentWrites reports whether the configured database can handle concurrent writes. It is false on SQLite, where overlapping write transactions deadlock, so clients should serialize batched writes instead of firing them in parallel.
	ConcurrentWrites bool `json:"concurrent_writes" doc:"Whether the configured database supports concurrent writes. False on SQLite; clients should serialize batched writes when this is false."`
}

// AuthInfo describes the authentication methods enabled on this instance.
type AuthInfo struct {
	Local         LocalAuthInfo  `json:"local"`
	Ldap          LdapAuthInfo   `json:"ldap"`
	OpenIDConnect OpenIDAuthInfo `json:"openid_connect"`
}

// LocalAuthInfo describes the local (username/password) authentication method.
type LocalAuthInfo struct {
	Enabled             bool `json:"enabled"`
	RegistrationEnabled bool `json:"registration_enabled"`
}

// LdapAuthInfo describes the LDAP authentication method.
type LdapAuthInfo struct {
	Enabled bool `json:"enabled"`
}

// OpenIDAuthInfo describes the OpenID Connect authentication method.
type OpenIDAuthInfo struct {
	Enabled   bool               `json:"enabled"`
	Providers []*openid.Provider `json:"providers"`
}

// LegalInfo holds links to the instance's legal documents.
type LegalInfo struct {
	ImprintURL       string `json:"imprint_url"`
	PrivacyPolicyURL string `json:"privacy_policy_url"`
}

// BuildInfo assembles the public instance information returned by GET /info on
// both API versions.
func BuildInfo() VikunjaInfos {
	info := VikunjaInfos{
		Version:                version.Version,
		FrontendURL:            config.ServicePublicURL.GetString(),
		Motd:                   config.ServiceMotd.GetString(),
		LinkSharingEnabled:     config.ServiceEnableLinkSharing.GetBool(),
		MaxFileSize:            config.FilesMaxSize.GetString(),
		MaxItemsPerPage:        config.ServiceMaxItemsPerPage.GetInt(),
		TaskAttachmentsEnabled: config.ServiceEnableTaskAttachments.GetBool(),
		TotpEnabled:            config.ServiceEnableTotp.GetBool(),
		CaldavEnabled:          config.ServiceEnableCaldav.GetBool(),
		EmailRemindersEnabled:  config.ServiceEnableEmailReminders.GetBool(),
		UserDeletionEnabled:    config.ServiceEnableUserDeletion.GetBool(),
		TaskCommentsEnabled:    config.ServiceEnableTaskComments.GetBool(),
		DemoModeEnabled:        config.ServiceDemoMode.GetBool(),
		WebhooksEnabled:        config.WebhooksEnabled.GetBool(),
		PublicTeamsEnabled:     config.ServiceEnablePublicTeams.GetBool(),
		AllowIconChanges:       config.ServiceAllowIconChanges.GetBool(),
		ConcurrentWrites:       config.DatabaseType.GetString() != "sqlite",
		EnabledProFeatures:     license.EnabledProFeatures(),
		AvailableMigrators: []string{
			(&vikunja_file.FileMigrator{}).Name(),
			(&ticktick.Migrator{}).Name(),
			(&wekan.Migrator{}).Name(),
			(&csvmigrator.Migrator{}).Name(),
		},
		Legal: LegalInfo{
			ImprintURL:       config.LegalImprintURL.GetString(),
			PrivacyPolicyURL: config.LegalPrivacyURL.GetString(),
		},
		AuthInfo: AuthInfo{
			Local: LocalAuthInfo{
				Enabled:             config.AuthLocalEnabled.GetBool(),
				RegistrationEnabled: config.AuthLocalEnabled.GetBool() && config.ServiceEnableRegistration.GetBool(),
			},
			Ldap: LdapAuthInfo{
				Enabled: config.AuthLdapEnabled.GetBool(),
			},
			OpenIDConnect: OpenIDAuthInfo{
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

	// ClickUp is always available - it needs no server-side configuration.
	info.AvailableMigrators = append(info.AvailableMigrators, (&clickup.Migration{}).Name())

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

	return info
}
