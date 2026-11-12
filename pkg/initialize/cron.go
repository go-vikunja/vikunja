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

package initialize

import (
	"context"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/jobs"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	migrationmodule "code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// RegisterCron starts shared schedules and process maintenance.
func RegisterCron() {
	cron.Init()
	if err := cron.Register("job-maintenance", "* * * * *", func(ctx context.Context, _ *xorm.Session, _ cron.Window) error { return jobs.Maintain(ctx) }); err != nil {
		log.Fatalf("Could not register durable job maintenance: %v", err)
	}
	models.RegisterReminderCron()
	models.RegisterOverdueReminderCron()
	models.RegisterUserDeletionCron()
	models.RegisterTaskCleanupCron()
	models.RegisterProjectCleanupCron()
	models.RegisterOldExportCleanupCron()
	migrationmodule.RegisterImportUploadCleanupCron()
	models.RegisterAddTaskToFilterViewCron()
	user.RegisterTokenCleanupCron()
	models.RegisterSessionCleanupCron()
	user.RegisterDeletionNotificationCron()
	if config.ServiceEnableScheduler.GetBool() {
		openid.CleanupSavedOpenIDProviders()
	}
	openid.RegisterEmptyOpenIDTeamCleanupCron()
	openid.RegisterProviderAvailabilityCron()
	models.RegisterAPITokenExpiryCheckCron()
}
