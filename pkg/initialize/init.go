// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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

package initialize

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/migration"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	migrator "code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/red"
	"code.vikunja.io/api/pkg/user"
)

// LightInit will only fullInit config, redis, logger but no db connection.
func LightInit() {
	// Init the config
	config.InitConfig()

	// Init redis
	red.InitRedis()

	// Init keyvalue store
	keyvalue.InitStorage()

	// Set logger
	log.InitLogger()
}

// InitEngines intializes all db connections
func InitEngines() {
	err := models.SetEngine()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = user.InitDB()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = files.SetEngine()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = migrator.InitDB()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// FullInit initializes all kinds of things in the right order
func FullInit() {

	LightInit()

	// Run the migrations
	migration.Migrate(nil)

	// Set Engine
	InitEngines()

	// Initialize the files handler
	files.InitFileHandler()

	// Start the mail daemon
	mail.StartMailDaemon()

	// Start the cron
	cron.Init()
	models.RegisterReminderCron()
}
