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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// MaxFrontendSettingsSize is the largest marshalled frontend_settings value in
// bytes. It stays below the 65,535 bytes of the MySQL TEXT column.
const MaxFrontendSettingsSize = 32 * 1024

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
	FrontendSettings             any    `json:"frontend_settings" doc:"Arbitrary settings used only by the frontend. A JSON object of at most 32 KiB; stored and returned verbatim."`
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
func ChangeUserPassword(ctx context.Context, s *xorm.Session, u *user.User, oldPassword, newPassword string) error {
	if oldPassword == "" {
		return user.ErrEmptyOldPassword{}
	}

	target, err := user.GetUserByID(s, u.ID)
	if err != nil {
		return err
	}
	if !target.IsLocalUser() {
		return &user.ErrAccountIsNotLocal{UserID: target.ID}
	}

	if err := user.CheckReauthPassword(ctx, s, target, oldPassword); err != nil {
		return err
	}

	return SetUserPasswordAndInvalidateSessions(s, target, newPassword)
}

// SetUserPasswordAndInvalidateSessions is shared by every password-change
// flow: a changed password must always kill existing sessions. Callers must
// dispatch pending events after committing.
func SetUserPasswordAndInvalidateSessions(s *xorm.Session, u *user.User, newPassword string) error {
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
	if err := checkFrontendSettings(settings.FrontendSettings); err != nil {
		return err
	}

	// Unchanged, so a default the user lost access to doesn't block saving.
	if settings.DefaultProjectID != 0 && settings.DefaultProjectID != u.DefaultProjectID {
		if err := checkDefaultProject(s, u, settings.DefaultProjectID); err != nil {
			return err
		}
	}

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

// changedGeneralSettings returns the sorted json names of the fields that
// differ. frontend_settings counts as one field.
func changedGeneralSettings(before, after *UserGeneralSettings) ([]string, error) {
	b, err := settingsFieldsJSON(before)
	if err != nil {
		return nil, err
	}
	a, err := settingsFieldsJSON(after)
	if err != nil {
		return nil, err
	}

	changed := []string{}
	for name, value := range a {
		if !bytes.Equal(value, b[name]) {
			changed = append(changed, name)
		}
	}
	slices.Sort(changed)
	return changed, nil
}

func settingsFieldsJSON(settings *UserGeneralSettings) (map[string]json.RawMessage, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	delete(fields, "extra_settings_links")
	return fields, nil
}

// checkFrontendSettings accepts no value or a JSON object up to
// MaxFrontendSettingsSize. Clients crash on other JSON types.
func checkFrontendSettings(settings any) error {
	raw, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	if bytes.Equal(raw, []byte("null")) {
		return nil
	}
	if raw[0] != '{' {
		return InvalidFieldErrorWithMessage([]string{"frontend_settings"}, "The frontend settings must be a JSON object.")
	}
	if len(raw) > MaxFrontendSettingsSize {
		return InvalidFieldErrorWithMessage([]string{"frontend_settings"}, "The frontend settings are too large.")
	}
	return nil
}

// validStoredFrontendSettings applies the rule of checkFrontendSettings to a
// stored value. It checks the stored bytes and does not decode them, so the
// size check comes first and is cheap. A value that Vikunja stored has the
// same length as its marshalled form.
func validStoredFrontendSettings(stored string) bool {
	raw := bytes.TrimSpace([]byte(stored))
	return len(raw) > 0 && len(raw) <= MaxFrontendSettingsSize &&
		raw[0] == '{' && json.Valid(raw)
}

// A var, so a test can make the batches small.
var frontendSettingsResetBatch = 500

// A text column, so xorm returns the raw stored value and does not decode it.
type storedFrontendSettings struct {
	ID               int64  `xorm:"bigint autoincr not null unique pk"`
	FrontendSettings string `xorm:"text null"`
}

func (storedFrontendSettings) TableName() string {
	return "users"
}

// ResetInvalidFrontendSettings sets every stored frontend_settings value that
// checkFrontendSettings rejects to NULL. Otherwise the user cannot save any
// general setting, because a settings PATCH sends the stored value back.
// A restore needs it: restored rows skip the migration that does the same.
func ResetInvalidFrontendSettings(s *xorm.Session) error {
	lastID := int64(0)
	for {
		rows := []*storedFrontendSettings{}
		err := s.
			Where(builder.Gt{"id": lastID}).
			And("frontend_settings IS NOT NULL").
			OrderBy("id").
			Limit(frontendSettingsResetBatch).
			Find(&rows)
		if err != nil {
			return fmt.Errorf("could not read frontend settings: %w", err)
		}

		for _, row := range rows {
			if validStoredFrontendSettings(row.FrontendSettings) {
				continue
			}
			if _, err := s.Exec("UPDATE users SET frontend_settings = NULL WHERE id = ?", row.ID); err != nil {
				return err
			}
			log.Infof("Reset the frontend settings of user %d: not a JSON object or larger than %d bytes", row.ID, MaxFrontendSettingsSize)
		}

		if len(rows) < frontendSettingsResetBatch {
			return nil
		}
		lastID = rows[len(rows)-1].ID
	}
}

func checkDefaultProject(s *xorm.Session, u *user.User, projectID int64) error {
	p, err := GetProjectSimpleByID(s, projectID)
	if IsErrProjectDoesNotExist(err) {
		return &ErrInvalidDefaultProject{ProjectID: projectID}
	}
	if err != nil {
		return err
	}
	can, err := p.CanWrite(s, u)
	if IsErrProjectIsArchived(err) {
		return &ErrInvalidDefaultProject{ProjectID: projectID}
	}
	if err != nil {
		return err
	}
	if !can {
		return &ErrInvalidDefaultProject{ProjectID: projectID}
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
