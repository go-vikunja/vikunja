// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Key is used as a config key
type Key string

// These constants hold all config value keys
const (
	ServiceJWTSecret         Key = `service.JWTSecret`
	ServiceInterface         Key = `service.interface`
	ServiceFrontendurl       Key = `service.frontendurl`
	ServiceEnableCaldav      Key = `service.enablecaldav`
	ServiceRootpath          Key = `service.rootpath`
	ServiceMaxItemsPerPage   Key = `service.maxitemsperpage`
	ServiceEnableMetrics     Key = `service.enablemetrics`
	ServiceMotd              Key = `service.motd`
	ServiceEnableLinkSharing Key = `service.enablelinksharing`

	DatabaseType                  Key = `database.type`
	DatabaseHost                  Key = `database.host`
	DatabaseUser                  Key = `database.user`
	DatabasePassword              Key = `database.password`
	DatabaseDatabase              Key = `database.database`
	DatabasePath                  Key = `database.path`
	DatabaseMaxOpenConnections    Key = `database.maxopenconnections`
	DatabaseMaxIdleConnections    Key = `database.maxidleconnections`
	DatabaseMaxConnectionLifetime Key = `database.maxconnectionlifetime`

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

	RedisEnabled  Key = `redis.enabled`
	RedisHost     Key = `redis.host`
	RedisPassword Key = `redis.password`
	RedisDB       Key = `redis.db`

	LogEnabled  Key = `log.enabled`
	LogErrors   Key = `log.errors`
	LogStandard Key = `log.standard`
	LogDatabase Key = `log.database`
	LogHTTP     Key = `log.http`
	LogEcho     Key = `log.echo`
	LogPath     Key = `log.path`

	RateLimitEnabled Key = `ratelimit.enabled`
	RateLimitKind    Key = `ratelimit.kind`
	RateLimitPeriod  Key = `ratelimit.period`
	RateLimitLimit   Key = `ratelimit.limit`
	RateLimitStore   Key = `ratelimit.store`

	FilesBasePath Key = `files.basepath`
	FilesMaxSize  Key = `files.maxsize`
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
	ServiceInterface.setDefault(":3456")
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
	// Redis
	RedisEnabled.setDefault(false)
	RedisHost.setDefault("localhost:6379")
	RedisPassword.setDefault("")
	RedisDB.setDefault(0)
	// Logger
	LogEnabled.setDefault(true)
	LogErrors.setDefault("stdout")
	LogStandard.setDefault("stdout")
	LogDatabase.setDefault("off")
	LogHTTP.setDefault("stdout")
	LogEcho.setDefault("off")
	LogPath.setDefault(ServiceRootpath.GetString() + "/logs")
	// Rate Limit
	RateLimitEnabled.setDefault(false)
	RateLimitKind.setDefault("user")
	RateLimitLimit.setDefault(100)
	RateLimitPeriod.setDefault(60)
	RateLimitStore.setDefault("memory")
	// Files
	FilesBasePath.setDefault("files")
	FilesMaxSize.setDefault("20MB")
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
	viper.AddConfigPath("~/.config/vikunja")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err.Error())
		log.Println("Using defaults.")
	}
}

func random(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", b), nil
}
