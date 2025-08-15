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

// UserSettings represents the settings of a user.
type UserSettings struct {
	Name                         string      `json:"name"`
	EmailRemindersEnabled        bool        `json:"email_reminders_enabled"`
	DiscoverableByName           bool        `json:"discoverable_by_name"`
	DiscoverableByEmail          bool        `json:"discoverable_by_email"`
	OverdueTasksRemindersEnabled bool        `json:"overdue_tasks_reminders_enabled"`
	OverdueTasksRemindersTime    string      `json:"overdue_tasks_reminders_time"`
	DefaultProjectID             int64       `json:"default_project_id"`
	WeekStart                    int         `json:"week_start"`
	Language                     string      `json:"language"`
	Timezone                     string      `json:"timezone"`
	FrontendSettings             interface{} `json:"frontend_settings"`
	AvatarProvider               string      `json:"avatar_provider"`
}

// EmailUpdate represents the request body for updating a user's email.
type EmailUpdate struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

// PasswordUpdate represents the request body for updating a user's password.
type PasswordUpdate struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
