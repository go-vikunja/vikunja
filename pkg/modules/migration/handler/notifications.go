// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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

package handler

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
)

// MigrationDoneNotification represents a MigrationDoneNotification notification
type MigrationDoneNotification struct {
	MigratorName string
}

// ToMail returns the mail notification for MigrationDoneNotification
func (n *MigrationDoneNotification) ToMail() *notifications.Mail {
	kind := cases.Title(language.English).String(n.MigratorName)

	return notifications.NewMail().
		Subject("The migration from "+kind+" to Vikunja was completed").
		Line("Vikunja has imported all lists/projects, tasks, notes, reminders and files from "+kind+" you have access to.").
		Action("View your imported projects in Vikunja", config.ServiceFrontendurl.GetString()).
		Line("Have fun with your new (old) projects!")
}

// ToDB returns the MigrationDoneNotification notification in a format which can be saved in the db
func (n *MigrationDoneNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *MigrationDoneNotification) Name() string {
	return "migration.done"
}
