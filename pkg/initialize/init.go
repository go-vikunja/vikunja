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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/audit"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/health"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/migration"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/ldap"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	"code.vikunja.io/api/pkg/plugins"
	_ "code.vikunja.io/api/pkg/plugins/yaegi" // register yaegi plugin loader
	"code.vikunja.io/api/pkg/red"
	ws "code.vikunja.io/api/pkg/websocket"
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

	if config.ServiceEnableMigrations.GetBool() {
		if err := db.CreateParadeDBIndexes(); err != nil {
			log.Fatal(err.Error())
		}
	}
}

// FullInitWithoutAsync does a full init without any async handlers (cron or events)
func FullInitWithoutAsync() {
	LightInit()

	// Initialize the files handler
	err := files.InitFileHandler(context.Background())
	if err != nil {
		log.Fatalf("Could not init file handler: %s", err)
	}

	// Run the migrations
	if config.ServiceEnableMigrations.GetBool() {
		migration.Migrate(nil)
	}

	// Set Engine
	InitEngines()

	// Initialize license validation — funds ongoing development of Vikunja.
	// See the package comment in pkg/license/license.go before removing.
	license.Init()

	if config.AuditEnabled.GetBool() {
		if err := audit.Init(); err != nil {
			log.Fatalf("Could not initialize audit logging: %s", err)
		}
	}

	// Start the mail daemon
	mail.StartMailDaemon()

	// Connect to ldap if enabled
	ldap.InitializeLDAPConnection()

	// Check all OpenID Connect providers at startup
	_, err = openid.GetAllProviders()
	if err != nil {
		if openid.IsErrDuplicateOIDCIssuer(err) {
			log.Fatalf("OpenID Connect configuration error: %s", err)
		}
		log.Errorf("Error initializing OpenID Connect providers: %s", err)
	}

	// Load translations
	i18n.Init()

	// Initialize plugins
	plugins.Initialize()
}

// FullInit initializes the API and background services.
func FullInit() {
	FullInitWithoutAsync()
	StartAsync()
}

func StartAsync() {
	ws.InitHub()
	if config.WebsocketRedisEnabled.GetBool() {
		channel := fmt.Sprintf("%s:%d:websocket", config.RedisPrefix.GetString(), config.RedisDB.GetInt())
		if err := ws.GetHub().StartRedis(context.Background(), red.GetRedis(), channel); err != nil {
			log.Fatalf("Could not initialize websocket fanout: %s", err)
		}
	}

	models.RegisterListeners()
	migrationHandler.RegisterListeners()
	ws.RegisterListeners()
	if err := events.Start(context.Background()); err != nil {
		log.Fatalf("Could not initialize events: %s", err)
	}

	health.RequireEvents()

	// Jobs can publish events immediately, so start them after the consumers.
	RegisterCron()
	if err := events.Dispatch(&BootedEvent{BootedAt: time.Now()}); err != nil {
		log.Fatalf("Could not dispatch boot event: %s", err)
	}
}
