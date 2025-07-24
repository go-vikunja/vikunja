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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/migration"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/ldap"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	"code.vikunja.io/api/pkg/plugins"
	"code.vikunja.io/api/pkg/red"
	"code.vikunja.io/api/pkg/user"
)

// LightInit will only init config, redis, logger but no db connection.
func LightInit() {
	// Set logger
	log.InitLogger()

	// Init the config
	config.InitConfig()

	// Check if the configured time zone is valid
	if _, err := time.LoadLocation(config.ServiceTimeZone.GetString()); err != nil {
		log.Criticalf("Error parsing default time zone: %s", err)
	}

	// Init redis
	red.InitRedis()

	// Init keyvalue store
	keyvalue.InitStorage()
}

// InitEngines intializes all db connections
func InitEngines() {
	err := models.SetEngine()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = files.SetEngine()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.CreateParadeDBIndexes()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// FullInitWithoutAsync does a full init without any async handlers (cron or events)
func FullInitWithoutAsync() {
	LightInit()

	// Initialize the files handler
	files.InitFileHandler()

	// Run the migrations
	migration.Migrate(nil)

	// Set Engine
	InitEngines()

	// Init Typesense
	models.InitTypesense()

	// Start the mail daemon
	mail.StartMailDaemon()

	// Connect to ldap if enabled
	ldap.InitializeLDAPConnection()

	// Load translations
	i18n.Init()

	// Initialize plugins
	plugins.Initialize()
}

// FullInit initializes all kinds of things in the right order
func FullInit() {

	FullInitWithoutAsync()

	// Start the cron
	cron.Init()
	models.RegisterReminderCron()
	models.RegisterOverdueReminderCron()
	models.RegisterUserDeletionCron()
	models.RegisterOldExportCleanupCron()
	models.RegisterAddTaskToFilterViewCron()
	user.RegisterTokenCleanupCron()
	user.RegisterDeletionNotificationCron()
	openid.CleanupSavedOpenIDProviders()
	openid.RegisterEmptyOpenIDTeamCleanupCron()

	// Start processing events
	go func() {
		models.RegisterListeners()
		user.RegisterListeners()
		migrationHandler.RegisterListeners()
		err := events.InitEvents()
		if err != nil {
			log.Fatal(err.Error())
		}

		err = events.Dispatch(&BootedEvent{
			BootedAt: time.Now(),
		})
		if err != nil {
			log.Fatal(err)
		}
	}()
}
