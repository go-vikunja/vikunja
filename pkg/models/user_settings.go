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
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// UserGeneralSettings is the single user-settings wire struct shared by v1 and
// v2 — both the update request body and the nested settings on GET /user. A
// dedicated struct (not user.User) is required: user.User's settings fields are
// json:"-" so they don't leak when it is embedded in other responses
// (assignees, created_by, members …).
type UserGeneralSettings struct {
	Name                         string `json:"name" doc:"The full name of the user."`
	EmailRemindersEnabled        bool   `json:"email_reminders_enabled" doc:"If enabled, sends email reminders of tasks to the user."`
	DiscoverableByName           bool   `json:"discoverable_by_name" doc:"If true, this user can be found by their name or parts of it when searching."`
	DiscoverableByEmail          bool   `json:"discoverable_by_email" doc:"If true, the user can be found when searching for their exact email."`
	OverdueTasksRemindersEnabled bool   `json:"overdue_tasks_reminders_enabled" doc:"If enabled, the user gets an email for their overdue tasks each morning."`
	OverdueTasksRemindersTime    string `json:"overdue_tasks_reminders_time" valid:"time,required" doc:"The time the daily overdue-tasks summary is sent, as HH:MM."`
	DefaultProjectID             int64  `json:"default_project_id" doc:"Project a task is filed under when created without an explicit project."`
	WeekStart                    int    `json:"week_start" valid:"range(0|6)" minimum:"0" maximum:"6" doc:"The day the week starts on: 0=sunday, 1=monday, … 6=saturday."`
	Language                     string `json:"language" doc:"The user's language."`
	Timezone                     string `json:"timezone" doc:"The user's time zone, used to send task reminders in their local time."`
	FrontendSettings             any    `json:"frontend_settings" doc:"Arbitrary settings used only by the frontend. Any JSON value; stored and returned verbatim."`
	// Server/OpenID-provided; populated on read, ignored on write.
	ExtraSettingsLinks map[string]any `json:"extra_settings_links" readOnly:"true" doc:"Additional settings links provided by the OpenID provider. Server-controlled."`
}

// NewUserGeneralSettings projects a user's stored settings into the shared wire
// struct for GET /user. Used by both the v1 and v2 user-show handlers.
func NewUserGeneralSettings(u *user.User) *UserGeneralSettings {
	return &UserGeneralSettings{
		Name:                         u.Name,
		EmailRemindersEnabled:        u.EmailRemindersEnabled,
		DiscoverableByName:           u.DiscoverableByName,
		DiscoverableByEmail:          u.DiscoverableByEmail,
		OverdueTasksRemindersEnabled: u.OverdueTasksRemindersEnabled,
		OverdueTasksRemindersTime:    u.OverdueTasksRemindersTime,
		DefaultProjectID:             u.DefaultProjectID,
		WeekStart:                    u.WeekStart,
		Language:                     u.Language,
		Timezone:                     u.Timezone,
		FrontendSettings:             u.FrontendSettings,
		ExtraSettingsLinks:           u.ExtraSettingsLinks,
	}
}

// ChangeUserPassword verifies the old password, sets the new one, and
// invalidates all of the user's sessions. Lives here (not in pkg/user) because
// it needs DeleteAllUserSessions, which pkg/user cannot import.
func ChangeUserPassword(s *xorm.Session, u *user.User, oldPassword, newPassword string) error {
	if oldPassword == "" {
		return user.ErrEmptyOldPassword{}
	}

	if _, err := user.CheckUserCredentials(s, &user.Login{Username: u.Username, Password: oldPassword}); err != nil {
		return err
	}

	if err := user.UpdateUserPassword(s, u, newPassword); err != nil {
		return err
	}

	return DeleteAllUserSessions(s, u.ID)
}

// UpdateUserGeneralSettings copies the general settings onto the user, persists
// them, and flushes the avatar cache when an initials avatar's name changed.
// Lives here (not in pkg/user) because the avatar flush needs pkg/modules/avatar,
// which pkg/user cannot import.
func UpdateUserGeneralSettings(s *xorm.Session, u *user.User, settings *UserGeneralSettings) error {
	invalidateAvatar := u.AvatarProvider == "initials" && u.Name != settings.Name

	u.Name = settings.Name
	u.EmailRemindersEnabled = settings.EmailRemindersEnabled
	u.DiscoverableByEmail = settings.DiscoverableByEmail
	u.DiscoverableByName = settings.DiscoverableByName
	u.OverdueTasksRemindersEnabled = settings.OverdueTasksRemindersEnabled
	u.DefaultProjectID = settings.DefaultProjectID
	u.WeekStart = settings.WeekStart
	u.Language = settings.Language
	u.Timezone = settings.Timezone
	u.OverdueTasksRemindersTime = settings.OverdueTasksRemindersTime
	u.FrontendSettings = settings.FrontendSettings

	if _, err := user.UpdateUser(s, u, true); err != nil {
		return err
	}

	if invalidateAvatar {
		avatar.FlushAllCaches(u)
	}
	return nil
}

// UpdateUserAvatarProvider sets the user's avatar provider, persists it, and
// flushes the avatar cache when the provider changes (or is set to initials).
func UpdateUserAvatarProvider(s *xorm.Session, u *user.User, provider string) error {
	oldProvider := u.AvatarProvider
	u.AvatarProvider = provider

	if _, err := user.UpdateUser(s, u, false); err != nil {
		return err
	}

	if u.AvatarProvider == "initials" || oldProvider != u.AvatarProvider {
		avatar.FlushAllCaches(u)
	}
	return nil
}
