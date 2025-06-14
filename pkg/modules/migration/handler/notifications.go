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

package handler

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
)

// MigrationDoneNotification represents a MigrationDoneNotification notification
type MigrationDoneNotification struct {
	MigratorName string
}

// ToMail returns the mail notification for MigrationDoneNotification
func (n *MigrationDoneNotification) ToMail(lang string) *notifications.Mail {
	kind := cases.Title(language.English).String(n.MigratorName)

	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.migration.done.subject", kind)).
		Line(i18n.T(lang, "notifications.migration.done.imported", kind)).
		Action(i18n.T(lang, "notifications.common.actions.open_vikunja"), config.ServicePublicURL.GetString()).
		Line(i18n.T(lang, "notifications.migration.done.have_fun"))
}

// ToDB returns the MigrationDoneNotification notification in a format which can be saved in the db
func (n *MigrationDoneNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *MigrationDoneNotification) Name() string {
	return "migration.done"
}

// MigrationFailedReportedNotification represents a MigrationFailedReportedNotification notification
type MigrationFailedReportedNotification struct {
	MigratorName string
}

// ToMail returns the mail notification for MigrationFailedReportedNotification
func (n *MigrationFailedReportedNotification) ToMail(lang string) *notifications.Mail {
	kind := cases.Title(language.English).String(n.MigratorName)

	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.migration.failed.subject", kind)).
		Line(i18n.T(lang, "notifications.migration.failed.message", kind)).
		Line(i18n.T(lang, "notifications.migration.failed.retry", kind)).
		Line(i18n.T(lang, "notifications.migration.failed.working_on_it"))
}

// ToDB returns the MigrationFailedReportedNotification notification in a format which can be saved in the db
func (n *MigrationFailedReportedNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *MigrationFailedReportedNotification) Name() string {
	return "migration.failed.reported"
}

// MigrationFailedNotification represents a MigrationFailedNotification notification
type MigrationFailedNotification struct {
	MigratorName string
	Error        error
}

// ToMail returns the mail notification for MigrationFailedNotification
func (n *MigrationFailedNotification) ToMail(lang string) *notifications.Mail {
	kind := cases.Title(language.English).String(n.MigratorName)

	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.migration.failed.subject", kind)).
		Line(i18n.T(lang, "notifications.migration.failed.message", kind)).
		Line(i18n.T(lang, "notifications.migration.failed.retry", kind)).
		Line(i18n.T(lang, "notifications.migration.failed.error", n.Error.Error())).
		Line(i18n.T(lang, "notifications.migration.failed.report"))
}

// ToDB returns the MigrationFailedNotification notification in a format which can be saved in the db
func (n *MigrationFailedNotification) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *MigrationFailedNotification) Name() string {
	return "migration.failed"
}
