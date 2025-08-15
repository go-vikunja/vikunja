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

import "code.vikunja.io/api/pkg/modules/auth/openid"

// Info represents information about the Vikunja instance.
type Info struct {
	Version                    string    `json:"version"`
	FrontendURL                string    `json:"frontend_url"`
	Motd                       string    `json:"motd"`
	LinkSharingEnabled         bool      `json:"link_sharing_enabled"`
	MaxFileSize                string    `json:"max_file_size"`
	AvailableMigrators         []string  `json:"available_migrators"`
	TaskAttachmentsEnabled     bool      `json:"task_attachments_enabled"`
	EnabledBackgroundProviders []string  `json:"enabled_background_providers"`
	TotpEnabled                bool      `json:"totp_enabled"`
	Legal                      LegalInfo `json:"legal"`
	CaldavEnabled              bool      `json:"caldav_enabled"`
	Auth                       AuthInfo  `json:"auth"`
	EmailRemindersEnabled      bool      `json:"email_reminders_enabled"`
	UserDeletionEnabled        bool      `json:"user_deletion_enabled"`
	TaskCommentsEnabled        bool      `json:"task_comments_enabled"`
	DemoModeEnabled            bool      `json:"demo_mode_enabled"`
	WebhooksEnabled            bool      `json:"webhooks_enabled"`
	PublicTeamsEnabled         bool      `json:"public_teams_enabled"`
	Links                      *InfoLinks `json:"_links"`
}

// InfoLinks represents the links for the info endpoint.
type InfoLinks struct {
	Self *Link `json:"self"`
}

// AuthInfo represents authentication information.
type AuthInfo struct {
	Local         LocalAuthInfo  `json:"local"`
	Ldap          LDAPAuthInfo   `json:"ldap"`
	OpenIDConnect OpenIDAuthInfo `json:"openid_connect"`
}

// LocalAuthInfo represents local authentication information.
type LocalAuthInfo struct {
	Enabled             bool `json:"enabled"`
	RegistrationEnabled bool `json:"registration_enabled"`
}

// LDAPAuthInfo represents LDAP authentication information.
type LDAPAuthInfo struct {
	Enabled bool `json:"enabled"`
}

// OpenIDAuthInfo represents OpenID Connect authentication information.
type OpenIDAuthInfo struct {
	Enabled   bool               `json:"enabled"`
	Providers []*openid.Provider `json:"providers"`
}

// LegalInfo represents legal information.
type LegalInfo struct {
	ImprintURL       string `json:"imprint_url"`
	PrivacyPolicyURL string `json:"privacy_policy_url"`
}
