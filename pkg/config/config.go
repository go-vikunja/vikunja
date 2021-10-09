// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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

package config

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	_ "time/tzdata" // Imports time zone data instead of relying on the os

	"github.com/spf13/viper"
)

// Key is used as a config key
type Key string

// These constants hold all config value keys
const (
	// #nosec
	ServiceJWTSecret       Key = `service.JWTSecret`
	ServiceJWTTTL          Key = `service.jwtttl`
	ServiceInterface       Key = `service.interface`
	ServiceUnixSocket      Key = `service.unixsocket`
	ServiceUnixSocketMode  Key = `service.unixsocketmode`
	ServiceFrontendurl     Key = `service.frontendurl`
	ServiceEnableCaldav    Key = `service.enablecaldav`
	ServiceRootpath        Key = `service.rootpath`
	ServiceMaxItemsPerPage Key = `service.maxitemsperpage`
	// Deprecated. Use metrics.enabled
	ServiceEnableMetrics         Key = `service.enablemetrics`
	ServiceMotd                  Key = `service.motd`
	ServiceEnableLinkSharing     Key = `service.enablelinksharing`
	ServiceEnableRegistration    Key = `service.enableregistration`
	ServiceEnableTaskAttachments Key = `service.enabletaskattachments`
	ServiceTimeZone              Key = `service.timezone`
	ServiceEnableTaskComments    Key = `service.enabletaskcomments`
	ServiceEnableTotp            Key = `service.enabletotp`
	ServiceSentryDsn             Key = `service.sentrydsn`
	ServiceTestingtoken          Key = `service.testingtoken`
	ServiceEnableEmailReminders  Key = `service.enableemailreminders`
	ServiceEnableUserDeletion    Key = `service.enableuserdeletion`

	AuthLocalEnabled      Key = `auth.local.enabled`
	AuthOpenIDEnabled     Key = `auth.openid.enabled`
	AuthOpenIDRedirectURL Key = `auth.openid.redirecturl`
	AuthOpenIDProviders   Key = `auth.openid.providers`

	LegalImprintURL Key = `legal.imprinturl`
	LegalPrivacyURL Key = `legal.privacyurl`

	DatabaseType                  Key = `database.type`
	DatabaseHost                  Key = `database.host`
	DatabaseUser                  Key = `database.user`
	DatabasePassword              Key = `database.password`
	DatabaseDatabase              Key = `database.database`
	DatabasePath                  Key = `database.path`
	DatabaseMaxOpenConnections    Key = `database.maxopenconnections`
	DatabaseMaxIdleConnections    Key = `database.maxidleconnections`
	DatabaseMaxConnectionLifetime Key = `database.maxconnectionlifetime`
	DatabaseSslMode               Key = `database.sslmode`
	DatabaseTLS                   Key = `database.tls`

	CacheEnabled        Key = `cache.enabled`
	CacheType           Key = `cache.type`
	CacheMaxElementSize Key = `cache.maxelementsize`

	MailerEnabled       Key = `mailer.enabled`
	MailerHost          Key = `mailer.host`
	MailerPort          Key = `mailer.port`
	MailerUsername      Key = `mailer.username`
	MailerPassword      Key = `mailer.password`
	MailerSkipTLSVerify Key = `mailer.skiptlsverify`
	MailerFromEmail     Key = `mailer.fromemail`
	MailerQueuelength   Key = `mailer.queuelength`
	MailerQueueTimeout  Key = `mailer.queuetimeout`
	MailerForceSSL      Key = `mailer.forcessl`

	RedisEnabled  Key = `redis.enabled`
	RedisHost     Key = `redis.host`
	RedisPassword Key = `redis.password`
	RedisDB       Key = `redis.db`

	LogEnabled       Key = `log.enabled`
	LogStandard      Key = `log.standard`
	LogLevel         Key = `log.level`
	LogDatabase      Key = `log.database`
	LogDatabaseLevel Key = `log.databaselevel`
	LogHTTP          Key = `log.http`
	LogEcho          Key = `log.echo`
	LogPath          Key = `log.path`
	LogEvents        Key = `log.events`
	LogEventsLevel   Key = `log.eventslevel`

	RateLimitEnabled Key = `ratelimit.enabled`
	RateLimitKind    Key = `ratelimit.kind`
	RateLimitPeriod  Key = `ratelimit.period`
	RateLimitLimit   Key = `ratelimit.limit`
	RateLimitStore   Key = `ratelimit.store`

	FilesBasePath Key = `files.basepath`
	FilesMaxSize  Key = `files.maxsize`

	MigrationWunderlistEnable          Key = `migration.wunderlist.enable`
	MigrationWunderlistClientID        Key = `migration.wunderlist.clientid`
	MigrationWunderlistClientSecret    Key = `migration.wunderlist.clientsecret`
	MigrationWunderlistRedirectURL     Key = `migration.wunderlist.redirecturl`
	MigrationTodoistEnable             Key = `migration.todoist.enable`
	MigrationTodoistClientID           Key = `migration.todoist.clientid`
	MigrationTodoistClientSecret       Key = `migration.todoist.clientsecret`
	MigrationTodoistRedirectURL        Key = `migration.todoist.redirecturl`
	MigrationTrelloEnable              Key = `migration.trello.enable`
	MigrationTrelloKey                 Key = `migration.trello.key`
	MigrationTrelloRedirectURL         Key = `migration.trello.redirecturl`
	MigrationMicrosoftTodoEnable       Key = `migration.microsofttodo.enable`
	MigrationMicrosoftTodoClientID     Key = `migration.microsofttodo.clientid`
	MigrationMicrosoftTodoClientSecret Key = `migration.microsofttodo.clientsecret`
	MigrationMicrosoftTodoRedirectURL  Key = `migration.microsofttodo.redirecturl`

	CorsEnable  Key = `cors.enable`
	CorsOrigins Key = `cors.origins`
	CorsMaxAge  Key = `cors.maxage`

	AvatarGravaterExpiration Key = `avatar.gravatarexpiration`

	BackgroundsEnabled               Key = `backgrounds.enabled`
	BackgroundsUploadEnabled         Key = `backgrounds.providers.upload.enabled`
	BackgroundsUnsplashEnabled       Key = `backgrounds.providers.unsplash.enabled`
	BackgroundsUnsplashAccessToken   Key = `backgrounds.providers.unsplash.accesstoken`
	BackgroundsUnsplashApplicationID Key = `backgrounds.providers.unsplash.applicationid`

	KeyvalueType Key = `keyvalue.type`

	MetricsEnabled  Key = `metrics.enabled`
	MetricsUsername Key = `metrics.username`
	MetricsPassword Key = `metrics.password`
)

// GetString returns a string config value
func (k Key) GetString() string {
	return viper.GetString(string(k))
}

// GetBool returns a bool config value
func (k Key) GetBool() bool {
	return viper.GetBool(string(k))
}

// GetInt returns an int config value
func (k Key) GetInt() int {
	return viper.GetInt(string(k))
}

// GetInt64 returns an int64 config value
func (k Key) GetInt64() int64 {
	return viper.GetInt64(string(k))
}

// GetDuration returns a duration config value
func (k Key) GetDuration() time.Duration {
	return viper.GetDuration(string(k))
}

// GetStringSlice returns a string slice from a config option
func (k Key) GetStringSlice() []string {
	return viper.GetStringSlice(string(k))
}

// Get returns the raw value from a config option
func (k Key) Get() interface{} {
	return viper.Get(string(k))
}

var timezone *time.Location

// GetTimeZone returns the time zone configured for vikunja
// It is a separate function and not done through viper because that makes handling
// it way easier, especially when testing.
func GetTimeZone() *time.Location {
	if timezone == nil {
		loc, err := time.LoadLocation(ServiceTimeZone.GetString())
		if err != nil {
			fmt.Printf("Error parsing time zone: %s", err)
			os.Exit(1)
		}
		timezone = loc
	}
	return timezone
}

// Set sets a value
func (k Key) Set(i interface{}) {
	viper.Set(string(k), i)
}

// sets the default config value
func (k Key) setDefault(i interface{}) {
	viper.SetDefault(string(k), i)
}

// InitDefaultConfig sets default config values
// This is an extra function so we can call it when initializing tests without initializing the full config
func InitDefaultConfig() {
	// Service config
	random, err := random(32)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Service
	ServiceJWTSecret.setDefault(random)
	ServiceJWTTTL.setDefault(259200)
	ServiceInterface.setDefault(":3456")
	ServiceUnixSocket.setDefault("")
	ServiceFrontendurl.setDefault("")
	ServiceEnableCaldav.setDefault(true)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	ServiceRootpath.setDefault(exPath)
	ServiceMaxItemsPerPage.setDefault(50)
	ServiceEnableMetrics.setDefault(false)
	ServiceMotd.setDefault("")
	ServiceEnableLinkSharing.setDefault(true)
	ServiceEnableRegistration.setDefault(true)
	ServiceEnableTaskAttachments.setDefault(true)
	ServiceTimeZone.setDefault("GMT")
	ServiceEnableTaskComments.setDefault(true)
	ServiceEnableTotp.setDefault(true)
	ServiceEnableEmailReminders.setDefault(true)
	ServiceEnableUserDeletion.setDefault(true)

	// Auth
	AuthLocalEnabled.setDefault(true)
	AuthOpenIDEnabled.setDefault(false)

	// Database
	DatabaseType.setDefault("sqlite")
	DatabaseHost.setDefault("localhost")
	DatabaseUser.setDefault("vikunja")
	DatabasePassword.setDefault("")
	DatabaseDatabase.setDefault("vikunja")
	DatabasePath.setDefault("./vikunja.db")
	DatabaseMaxOpenConnections.setDefault(100)
	DatabaseMaxIdleConnections.setDefault(50)
	DatabaseMaxConnectionLifetime.setDefault(10000)
	DatabaseSslMode.setDefault("disable")
	DatabaseTLS.setDefault("false")

	// Cacher
	CacheEnabled.setDefault(false)
	CacheType.setDefault("memory")
	CacheMaxElementSize.setDefault(1000)
	// Mailer
	MailerEnabled.setDefault(false)
	MailerHost.setDefault("")
	MailerPort.setDefault("587")
	MailerUsername.setDefault("user")
	MailerPassword.setDefault("")
	MailerSkipTLSVerify.setDefault(false)
	MailerFromEmail.setDefault("mail@vikunja")
	MailerQueuelength.setDefault(100)
	MailerQueueTimeout.setDefault(30)
	MailerForceSSL.setDefault(false)
	// Redis
	RedisEnabled.setDefault(false)
	RedisHost.setDefault("localhost:6379")
	RedisPassword.setDefault("")
	RedisDB.setDefault(0)
	// Logger
	LogEnabled.setDefault(true)
	LogStandard.setDefault("stdout")
	LogLevel.setDefault("INFO")
	LogDatabase.setDefault("off")
	LogDatabaseLevel.setDefault("WARNING")
	LogHTTP.setDefault("stdout")
	LogEcho.setDefault("off")
	LogPath.setDefault(ServiceRootpath.GetString() + "/logs")
	LogEvents.setDefault("stdout")
	LogEventsLevel.setDefault("INFO")
	// Rate Limit
	RateLimitEnabled.setDefault(false)
	RateLimitKind.setDefault("user")
	RateLimitLimit.setDefault(100)
	RateLimitPeriod.setDefault(60)
	RateLimitStore.setDefault("memory")
	// Files
	FilesBasePath.setDefault("files")
	FilesMaxSize.setDefault("20MB")
	// Cors
	CorsEnable.setDefault(true)
	CorsOrigins.setDefault([]string{"*"})
	CorsMaxAge.setDefault(0)
	// Migration
	MigrationWunderlistEnable.setDefault(false)
	MigrationTodoistEnable.setDefault(false)
	MigrationTrelloEnable.setDefault(false)
	MigrationMicrosoftTodoEnable.setDefault(false)
	// Avatar
	AvatarGravaterExpiration.setDefault(3600)
	// List Backgrounds
	BackgroundsEnabled.setDefault(true)
	BackgroundsUploadEnabled.setDefault(true)
	BackgroundsUnsplashEnabled.setDefault(false)
	// Key Value
	KeyvalueType.setDefault("memory")
	// Metrics
	MetricsEnabled.setDefault(false)
}

// InitConfig initializes the config, sets defaults etc.
func InitConfig() {

	// Set defaults
	InitDefaultConfig()

	// Init checking for environment variables
	viper.SetEnvPrefix("vikunja")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Load the config file
	viper.AddConfigPath(ServiceRootpath.GetString())
	viper.AddConfigPath("/etc/vikunja/")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("No home directory found, not using config from ~/.config/vikunja/. Error was: %s\n", err.Error())
	} else {
		viper.AddConfigPath(path.Join(homeDir, ".config", "vikunja"))
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err = viper.ReadInConfig()
	if err != nil {
		log.Println(err.Error())
		log.Println("Using default config.")
		return
	}

	if CacheType.GetString() == "keyvalue" {
		CacheType.Set(KeyvalueType.GetString())
	}

	if RateLimitStore.GetString() == "keyvalue" {
		RateLimitStore.Set(KeyvalueType.GetString())
	}

	if ServiceFrontendurl.GetString() != "" && !strings.HasSuffix(ServiceFrontendurl.GetString(), "/") {
		ServiceFrontendurl.Set(ServiceFrontendurl.GetString() + "/")
	}

	if AuthOpenIDRedirectURL.GetString() == "" {
		AuthOpenIDRedirectURL.Set(ServiceFrontendurl.GetString() + "auth/openid/")
	}

	if MigrationTodoistRedirectURL.GetString() == "" {
		MigrationTodoistRedirectURL.Set(ServiceFrontendurl.GetString() + "migrate/todoist")
	}

	if MigrationTrelloRedirectURL.GetString() == "" {
		MigrationTrelloRedirectURL.Set(ServiceFrontendurl.GetString() + "migrate/trello")
	}

	if MigrationMicrosoftTodoRedirectURL.GetString() == "" {
		MigrationMicrosoftTodoRedirectURL.Set(ServiceFrontendurl.GetString() + "migrate/microsoft-todo")
	}

	if ServiceEnableMetrics.GetBool() {
		log.Println("WARNING: service.enablemetrics is deprecated and will be removed in a future release. Please use metrics.enable.")
		MetricsEnabled.Set(true)
	}

	log.Printf("Using config file: %s", viper.ConfigFileUsed())
}

func random(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", b), nil
}
