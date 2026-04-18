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

package config

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	_ "time/tzdata" // Imports time zone data instead of relying on the os

	"code.vikunja.io/api/pkg/log"

	"github.com/c2h5oh/datasize"
	"github.com/spf13/viper"
)

// Key is used as a config key
type Key string

// These constants hold all config value keys
const (
	// #nosec
	ServiceSecret                         Key = `service.secret`
	ServiceJWTSecret                      Key = `service.JWTSecret` // #nosec G101 -- Deprecated config key alias, not a credential
	ServiceJWTTTL                         Key = `service.jwtttl`
	ServiceJWTTTLLong                     Key = `service.jwtttllong`
	ServiceJWTTTLShort                    Key = `service.jwtttlshort`
	ServiceInterface                      Key = `service.interface`
	ServiceUnixSocket                     Key = `service.unixsocket`
	ServiceUnixSocketMode                 Key = `service.unixsocketmode`
	ServicePublicURL                      Key = `service.publicurl`
	ServiceEnableCaldav                   Key = `service.enablecaldav`
	ServiceRootpath                       Key = `service.rootpath`
	ServiceMaxItemsPerPage                Key = `service.maxitemsperpage`
	ServiceDemoMode                       Key = `service.demomode`
	ServiceMotd                           Key = `service.motd`
	ServiceEnableLinkSharing              Key = `service.enablelinksharing`
	ServiceEnableRegistration             Key = `service.enableregistration`
	ServiceEnableTaskAttachments          Key = `service.enabletaskattachments`
	ServiceTimeZone                       Key = `service.timezone`
	ServiceEnableTaskComments             Key = `service.enabletaskcomments`
	ServiceEnableTotp                     Key = `service.enabletotp`
	ServiceTestingtoken                   Key = `service.testingtoken`
	ServiceEnableEmailReminders           Key = `service.enableemailreminders`
	ServiceEnableUserDeletion             Key = `service.enableuserdeletion`
	ServiceEnableBotUsers                 Key = `service.enablebotusers`
	ServiceMaxAvatarSize                  Key = `service.maxavatarsize`
	ServiceAllowIconChanges               Key = `service.allowiconchanges`
	ServiceCustomLogoURL                  Key = `service.customlogourl`
	ServiceCustomLogoURLDark              Key = `service.customlogourldark`
	ServiceEnablePublicTeams              Key = `service.enablepublicteams`
	ServiceBcryptRounds                   Key = `service.bcryptrounds`
	ServiceEnableOpenIDTeamUserOnlySearch Key = `service.enableopenidteamusersearch`
	ServiceIPExtractionMethod             Key = `service.ipextractionmethod`
	ServiceTrustedProxies                 Key = `service.trustedproxies`

	SentryEnabled         Key = `sentry.enabled`
	SentryDsn             Key = `sentry.dsn`
	SentryFrontendEnabled Key = `sentry.frontendenabled`
	SentryFrontendDsn     Key = `sentry.frontenddsn`

	AuthLocalEnabled    Key = `auth.local.enabled`
	AuthOpenIDEnabled   Key = `auth.openid.enabled`
	AuthOpenIDProviders Key = `auth.openid.providers`

	AuthLdapEnabled    Key = `auth.ldap.enabled`
	AuthLdapHost       Key = `auth.ldap.host`
	AuthLdapPort       Key = `auth.ldap.port`
	AuthLdapBaseDN     Key = `auth.ldap.basedn`
	AuthLdapUserFilter Key = `auth.ldap.userfilter`
	AuthLdapUseTLS     Key = `auth.ldap.usetls`
	AuthLdapVerifyTLS  Key = `auth.ldap.verifytls`
	AuthLdapBindDN     Key = `auth.ldap.binddn`
	// #nosec G101
	AuthLdapBindPassword         Key = `auth.ldap.bindpassword`
	AuthLdapGroupSyncEnabled     Key = `auth.ldap.groupsyncenabled`
	AuthLdapGroupSyncFilter      Key = `auth.ldap.groupsyncfilter`
	AuthLdapAvatarSyncAttribute  Key = `auth.ldap.avatarsyncattribute`
	AuthLdapAttributeUsername    Key = `auth.ldap.attribute.username`
	AuthLdapAttributeEmail       Key = `auth.ldap.attribute.email`
	AuthLdapAttributeDisplayname Key = `auth.ldap.attribute.displayname`
	AuthLdapAttributeMemberID    Key = `auth.ldap.attribute.memberid`

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
	DatabaseSslCert               Key = `database.sslcert`
	DatabaseSslKey                Key = `database.sslkey`
	DatabaseSslRootCert           Key = `database.sslrootcert`
	DatabaseTLS                   Key = `database.tls`
	DatabaseSchema                Key = `database.schema`

	MailerEnabled       Key = `mailer.enabled`
	MailerHost          Key = `mailer.host`
	MailerPort          Key = `mailer.port`
	MailerUsername      Key = `mailer.username`
	MailerPassword      Key = `mailer.password`
	MailerAuthType      Key = `mailer.authtype`
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
	LogFormat        Key = `log.format`
	LogDatabase      Key = `log.database`
	LogDatabaseLevel Key = `log.databaselevel`
	LogHTTP          Key = `log.http`
	LogPath          Key = `log.path`
	LogEvents        Key = `log.events`
	LogEventsLevel   Key = `log.eventslevel`
	LogMail          Key = `log.mail`
	LogMailLevel     Key = `log.maillevel`

	RateLimitEnabled           Key = `ratelimit.enabled`
	RateLimitKind              Key = `ratelimit.kind`
	RateLimitPeriod            Key = `ratelimit.period`
	RateLimitLimit             Key = `ratelimit.limit`
	RateLimitStore             Key = `ratelimit.store`
	RateLimitNoAuthRoutesLimit Key = `ratelimit.noauthlimit`

	FilesBasePath Key = `files.basepath`
	FilesMaxSize  Key = `files.maxsize`
	FilesType     Key = `files.type`

	// S3 Configuration
	FilesS3Endpoint       Key = `files.s3.endpoint`
	FilesS3Bucket         Key = `files.s3.bucket`
	FilesS3Region         Key = `files.s3.region`
	FilesS3AccessKey      Key = `files.s3.accesskey`
	FilesS3SecretKey      Key = `files.s3.secretkey`
	FilesS3UsePathStyle   Key = `files.s3.usepathstyle`
	FilesS3DisableSigning Key = `files.s3.disablesigning`
	FilesS3TempDir        Key = `files.s3.tempdir`

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
	AvatarGravatarBaseURL    Key = `avatar.gravatarbaseurl`

	BackgroundsEnabled               Key = `backgrounds.enabled`
	BackgroundsUploadEnabled         Key = `backgrounds.providers.upload.enabled`
	BackgroundsUnsplashEnabled       Key = `backgrounds.providers.unsplash.enabled`
	BackgroundsUnsplashAccessToken   Key = `backgrounds.providers.unsplash.accesstoken`
	BackgroundsUnsplashApplicationID Key = `backgrounds.providers.unsplash.applicationid`

	KeyvalueType Key = `keyvalue.type`

	MetricsEnabled  Key = `metrics.enabled`
	MetricsUsername Key = `metrics.username`
	MetricsPassword Key = `metrics.password`

	DefaultSettingsAvatarProvider              Key = `defaultsettings.avatar_provider`
	DefaultSettingsAvatarFileID                Key = `defaultsettings.avatar_file_id`
	DefaultSettingsEmailRemindersEnabled       Key = `defaultsettings.email_reminders_enabled`
	DefaultSettingsDiscoverableByName          Key = `defaultsettings.discoverable_by_name`
	DefaultSettingsDiscoverableByEmail         Key = `defaultsettings.discoverable_by_email`
	DefaultSettingsOverdueTaskRemindersEnabled Key = `defaultsettings.overdue_tasks_reminders_enabled`
	DefaultSettingsDefaultProjectID            Key = `defaultsettings.default_project_id`
	DefaultSettingsWeekStart                   Key = `defaultsettings.week_start`
	DefaultSettingsLanguage                    Key = `defaultsettings.language`
	DefaultSettingsTimezone                    Key = `defaultsettings.timezone`
	DefaultSettingsOverdueTaskRemindersTime    Key = `defaultsettings.overdue_tasks_reminders_time`

	WebhooksEnabled             Key = `webhooks.enabled`
	WebhooksTimeoutSeconds      Key = `webhooks.timeoutseconds`
	WebhooksProxyURL            Key = `webhooks.proxyurl`
	WebhooksProxyPassword       Key = `webhooks.proxypassword`
	WebhooksAllowNonRoutableIPs Key = `webhooks.allownonroutableips`

	OutgoingRequestsAllowNonRoutableIPs Key = `outgoingrequests.allownonroutableips`
	OutgoingRequestsProxyURL            Key = `outgoingrequests.proxyurl`
	OutgoingRequestsProxyPassword       Key = `outgoingrequests.proxypassword`
	OutgoingRequestsTimeoutSeconds      Key = `outgoingrequests.timeoutseconds`

	AutoTLSEnabled     Key = `autotls.enabled`
	AutoTLSEmail       Key = `autotls.email`
	AutoTLSRenewBefore Key = `autotls.renewbefore`

	PluginsEnabled Key = `plugins.enabled`
	PluginsDir     Key = `plugins.dir`
	PluginsLoader  Key = `plugins.loader`

	// LicenseKey gates optional paid features and funds Vikunja's development.
	// See the package comment in pkg/license/license.go before removing.
	LicenseKey Key = `license.key`
)

var maxFileSizeInBytes uint64

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
			log.Fatalf("Error parsing time zone: %s", err)
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

// getRootpathLocation determines the default root path for Vikunja data.
// It prefers the current working directory, which respects systemd's
// WorkingDirectory= setting and is the most intuitive default.
// Falls back to the binary's directory if Getwd fails.
func getRootpathLocation() string {
	// Prefer working directory — this respects systemd WorkingDirectory=
	// and is the intuitive default for most deployment scenarios.
	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	// Fall back to the binary's directory.
	if ex, err := os.Executable(); err == nil {
		return filepath.Dir(ex)
	}

	// Last resort: search $PATH.
	exeSuffix := ""
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	if exeLocation, err := exec.LookPath("vikunja" + exeSuffix); err == nil {
		return filepath.Dir(exeLocation)
	}

	log.Fatal("Could not determine root path. Set service.rootpath in your config.")
	return ""
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
	ServiceSecret.setDefault(random)
	ServiceJWTTTL.setDefault(259200)      // 72 hours
	ServiceJWTTTLLong.setDefault(2592000) // 30 days
	ServiceJWTTTLShort.setDefault(600)    // 10 minutes
	ServiceInterface.setDefault(":3456")
	ServiceUnixSocket.setDefault("")
	ServicePublicURL.setDefault("")
	ServiceEnableCaldav.setDefault(true)

	ServiceRootpath.setDefault(getRootpathLocation())
	ServiceMaxItemsPerPage.setDefault(50)
	ServiceMotd.setDefault("")
	ServiceEnableLinkSharing.setDefault(true)
	ServiceEnableRegistration.setDefault(true)
	ServiceEnableTaskAttachments.setDefault(true)
	ServiceTimeZone.setDefault("GMT")
	ServiceEnableTaskComments.setDefault(true)
	ServiceEnableTotp.setDefault(true)
	ServiceEnableEmailReminders.setDefault(true)
	ServiceEnableUserDeletion.setDefault(true)
	ServiceEnableBotUsers.setDefault(true)
	ServiceMaxAvatarSize.setDefault(1024)
	ServiceDemoMode.setDefault(false)
	ServiceAllowIconChanges.setDefault(true)
	ServiceEnablePublicTeams.setDefault(false)
	ServiceBcryptRounds.setDefault(11)
	ServiceEnableOpenIDTeamUserOnlySearch.setDefault(false)
	ServiceIPExtractionMethod.setDefault("direct")
	ServiceTrustedProxies.setDefault("")

	// Sentry
	SentryDsn.setDefault("https://440eedc957d545a795c17bbaf477497c@o1047380.ingest.sentry.io/4504254983634944")
	SentryFrontendDsn.setDefault("https://85694a2d757547cbbc90cd4b55c5a18d@o1047380.ingest.sentry.io/6024480")

	// Auth
	AuthLocalEnabled.setDefault(true)
	AuthOpenIDEnabled.setDefault(false)

	AuthLdapEnabled.setDefault(false)
	AuthLdapHost.setDefault("localhost")
	AuthLdapPort.setDefault(389)
	AuthLdapUseTLS.setDefault(true)
	AuthLdapVerifyTLS.setDefault(true)
	AuthLdapGroupSyncEnabled.setDefault(false)
	AuthLdapGroupSyncFilter.setDefault("(&(objectclass=*)(|(objectclass=group)(objectclass=groupOfNames)))")
	AuthLdapAttributeUsername.setDefault("uid")
	AuthLdapAttributeEmail.setDefault("mail")
	AuthLdapAttributeDisplayname.setDefault("displayName")
	AuthLdapAttributeMemberID.setDefault("member")

	// Database
	DatabaseType.setDefault("sqlite")
	DatabaseHost.setDefault("localhost")
	DatabaseUser.setDefault("vikunja")
	DatabasePassword.setDefault("")
	DatabaseDatabase.setDefault("vikunja")
	DatabasePath.setDefault(ResolvePath("vikunja.db"))
	DatabaseMaxOpenConnections.setDefault(100)
	DatabaseMaxIdleConnections.setDefault(50)
	DatabaseMaxConnectionLifetime.setDefault(10000)
	DatabaseSslMode.setDefault("disable")
	DatabaseSslCert.setDefault("")
	DatabaseSslKey.setDefault("")
	DatabaseSslRootCert.setDefault("")
	DatabaseTLS.setDefault("false")
	DatabaseSchema.setDefault("public")

	// Mailer
	MailerEnabled.setDefault(false)
	MailerHost.setDefault("")
	MailerPort.setDefault("587")
	MailerUsername.setDefault("")
	MailerPassword.setDefault("")
	MailerSkipTLSVerify.setDefault(false)
	MailerFromEmail.setDefault("mail@vikunja")
	MailerQueuelength.setDefault(100)
	MailerQueueTimeout.setDefault(30)
	MailerForceSSL.setDefault(false)
	MailerAuthType.setDefault("plain")
	// Redis
	RedisEnabled.setDefault(false)
	RedisHost.setDefault("localhost:6379")
	RedisPassword.setDefault("")
	RedisDB.setDefault(0)
	// Logger
	LogEnabled.setDefault(true)
	LogStandard.setDefault("stdout")
	LogLevel.setDefault("INFO")
	LogFormat.setDefault("text")
	LogDatabase.setDefault("off")
	LogDatabaseLevel.setDefault("WARNING")
	LogHTTP.setDefault("stdout")
	LogPath.setDefault(ResolvePath("logs"))
	LogEvents.setDefault("off")
	LogEventsLevel.setDefault("INFO")
	LogMail.setDefault("off")
	LogMailLevel.setDefault("INFO")
	// Rate Limit
	RateLimitEnabled.setDefault(false)
	RateLimitKind.setDefault("user")
	RateLimitLimit.setDefault(100)
	RateLimitPeriod.setDefault(60)
	RateLimitStore.setDefault("memory")
	RateLimitNoAuthRoutesLimit.setDefault(10)
	// Files
	FilesBasePath.setDefault("files")
	FilesMaxSize.setDefault("20MB")
	FilesType.setDefault("local")
	// S3 Configuration
	FilesS3Endpoint.setDefault("")
	FilesS3Bucket.setDefault("")
	FilesS3Region.setDefault("")
	FilesS3AccessKey.setDefault("")
	FilesS3SecretKey.setDefault("")
	FilesS3UsePathStyle.setDefault(false)
	FilesS3DisableSigning.setDefault(false)
	FilesS3TempDir.setDefault("")
	// Cors
	CorsEnable.setDefault(true)
	CorsOrigins.setDefault([]string{"http://127.0.0.1:*", "http://localhost:*"})
	CorsMaxAge.setDefault(0)
	// Migration
	MigrationTodoistEnable.setDefault(false)
	MigrationTrelloEnable.setDefault(false)
	MigrationMicrosoftTodoEnable.setDefault(false)
	// Avatar
	AvatarGravaterExpiration.setDefault(3600)
	AvatarGravatarBaseURL.setDefault("https://www.gravatar.com")
	// Project Backgrounds
	BackgroundsEnabled.setDefault(true)
	BackgroundsUploadEnabled.setDefault(true)
	BackgroundsUnsplashEnabled.setDefault(false)
	// Key Value
	KeyvalueType.setDefault("memory")
	// Metrics
	MetricsEnabled.setDefault(false)
	// Settings
	DefaultSettingsAvatarProvider.setDefault("initials")
	DefaultSettingsOverdueTaskRemindersEnabled.setDefault(true)
	DefaultSettingsOverdueTaskRemindersTime.setDefault("9:00")
	// Webhook
	WebhooksEnabled.setDefault(true)
	WebhooksTimeoutSeconds.setDefault(30)
	WebhooksAllowNonRoutableIPs.setDefault(false)
	// Outgoing Requests
	OutgoingRequestsAllowNonRoutableIPs.setDefault(false)
	OutgoingRequestsTimeoutSeconds.setDefault(30)
	// AutoTLS
	AutoTLSRenewBefore.setDefault("720h") // 30days in hours
	// Plugins
	PluginsEnabled.setDefault(false)
	PluginsDir.setDefault(ResolvePath("plugins"))
	PluginsLoader.setDefault("native")

	// Migrate deprecated webhook config keys to outgoingrequests.*
	// This allows removing the old keys in a single place later.
	if WebhooksAllowNonRoutableIPs.GetBool() && !OutgoingRequestsAllowNonRoutableIPs.GetBool() {
		log.Warningf("Config key %q is deprecated and will be removed in a future release. Please use %q instead.", WebhooksAllowNonRoutableIPs, OutgoingRequestsAllowNonRoutableIPs)
		OutgoingRequestsAllowNonRoutableIPs.Set("true")
	}
	if proxyURL := WebhooksProxyURL.GetString(); proxyURL != "" && OutgoingRequestsProxyURL.GetString() == "" {
		log.Warningf("Config key %q is deprecated and will be removed in a future release. Please use %q instead.", WebhooksProxyURL, OutgoingRequestsProxyURL)
		OutgoingRequestsProxyURL.Set(proxyURL)
	}
	if proxyPassword := WebhooksProxyPassword.GetString(); proxyPassword != "" && OutgoingRequestsProxyPassword.GetString() == "" {
		log.Warningf("Config key %q is deprecated and will be removed in a future release. Please use %q instead.", WebhooksProxyPassword, OutgoingRequestsProxyPassword)
		OutgoingRequestsProxyPassword.Set(proxyPassword)
	}
	// License
	LicenseKey.setDefault("")
}

// ResolvePath resolves a path relative to service.rootpath.
// If the path is already absolute, it is returned as-is (cleaned).
// If the path is relative (or empty), it is joined with service.rootpath.
func ResolvePath(p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(ServiceRootpath.GetString(), p)
}

func GetConfigValueFromFile(configKey string) string {
	if !strings.HasSuffix(configKey, ".file") {
		configKey += ".file"
	}
	var valuePath = os.ExpandEnv(viper.GetString(configKey))
	if valuePath == "" {
		return ""
	}

	valuePath = ResolvePath(valuePath)

	contents, err := os.ReadFile(valuePath)
	if err == nil {
		return strings.Trim(string(contents), "\n")
	}

	log.Fatalf("Failed to read the config file at %s for key %s: %v", valuePath, configKey, err)
	return ""
}

func readConfigValuesFromFiles() {
	keys := viper.AllKeys()
	for _, key := range keys {
		if strings.Contains(key, "auth.openid.providers") {
			// Setting openid provider values will remove everything but the value from file
			continue
		}
		// Env is evaluated manually at runtime, so we need to check this for each key
		value := GetConfigValueFromFile(key)
		if value != "" {
			viper.Set(strings.TrimSuffix(key, ".file"), value)
		}
	}
}

func setConfigFromEnv() error {
	envKeys := os.Environ()
	configMap := make(map[string]any)
	for _, envKeyValue := range envKeys {
		keyValue := strings.SplitN(envKeyValue, "=", 2)
		if len(keyValue) != 2 {
			continue
		}
		key, value := keyValue[0], keyValue[1]

		if strings.HasPrefix(key, "VIKUNJA_") {
			formattedKey := strings.ToLower(strings.TrimPrefix(key, "VIKUNJA_"))
			keys := strings.Split(formattedKey, "_")
			currentMap := configMap

			for i, part := range keys {
				_, isString := currentMap[part].(string)
				if i == len(keys)-1 || isString {
					// Set the value at the final level
					currentMap[part] = value
				} else {
					// Check if the key exists at the current level, create a new map if not
					if _, exists := currentMap[part]; !exists {
						currentMap[part] = make(map[string]any)
					}

					// Move into the nested map
					typed, is := currentMap[part].(map[string]any)
					if !is {
						log.Errorf("Failed to set config value from environment variable %s: %s, failed on part %s, type is not map, is %T", key, value, part, currentMap[part])
						continue
					}

					currentMap = typed
				}
			}
		}
	}
	return viper.MergeConfigMap(configMap)
}

// InitConfig initializes the config, sets defaults etc.
func InitConfig() {

	// Set defaults
	InitDefaultConfig()

	// Init checking for environment variables
	viper.SetEnvPrefix("vikunja")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	log.ConfigureStandardLogger(LogEnabled.GetBool(), LogStandard.GetString(), LogPath.GetString(), LogLevel.GetString(), LogFormat.GetString())

	// Load the config file
	viper.AddConfigPath(ServiceRootpath.GetString())
	viper.AddConfigPath("/etc/vikunja/")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Debugf("No home directory found, not using config from ~/.config/vikunja/. Error was: %s\n", err.Error())
	} else {
		viper.AddConfigPath(path.Join(homeDir, ".config", "vikunja"))
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	err = viper.ReadInConfig()

	if viper.ConfigFileUsed() != "" {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())

		if err != nil {
			log.Warning(err.Error())
			log.Warning("Using default config.")
		} else {
			log.ConfigureStandardLogger(LogEnabled.GetBool(), LogStandard.GetString(), LogPath.GetString(), LogLevel.GetString(), LogFormat.GetString())
		}
	} else {
		log.Info("No config file found, using default or config from environment variables.")
	}

	err = setConfigFromEnv()
	if err != nil {
		log.Warningf("Failed to set config from environment variables: %s", err.Error())
	}

	readConfigValuesFromFiles()

	// Deprecation: migrate service.JWTSecret → service.secret only when the
	// user has not explicitly set service.secret (so the new key takes precedence).
	if ServiceJWTSecret.GetString() != "" {
		if viper.IsSet(string(ServiceSecret)) {
			log.Warning("config: both service.secret and service.jwtsecret are set. Using service.secret. Please remove service.jwtsecret, it is deprecated and will be removed in a future release.")
		} else {
			log.Warning("config: service.jwtsecret is deprecated and will be removed in a future release. Please use service.secret instead.")
			ServiceSecret.Set(ServiceJWTSecret.GetString())
		}
	}

	if _, err := url.ParseRequestURI(AvatarGravatarBaseURL.GetString()); err != nil {
		log.Fatalf("Could not parse gravatarbaseurl: %s", err)
	}

	AvatarGravatarBaseURL.Set(strings.TrimRight(AvatarGravatarBaseURL.GetString(), "/"))

	if RateLimitStore.GetString() == "keyvalue" {
		RateLimitStore.Set(KeyvalueType.GetString())
	}

	if loader := PluginsLoader.GetString(); loader != "yaegi" && loader != "native" {
		log.Fatalf("Invalid value for plugins.loader: %q (must be \"yaegi\" or \"native\")", loader)
	}

	if CorsEnable.GetBool() && ServicePublicURL.GetString() == "" {
		log.Fatalf("service.publicurl is required when cors.enable is true")
	}

	if ServicePublicURL.GetString() != "" {
		if !strings.HasSuffix(ServicePublicURL.GetString(), "/") {
			ServicePublicURL.Set(ServicePublicURL.GetString() + "/")
		}

		parsed, err := url.Parse(ServicePublicURL.GetString())
		if err != nil {
			log.Fatalf("Could not parse publicurl: %s", err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			log.Fatalf("service.publicurl must include http:// or https:// scheme, got: %s", ServicePublicURL.GetString())
		}
	}

	if MigrationTodoistRedirectURL.GetString() == "" {
		MigrationTodoistRedirectURL.Set(ServicePublicURL.GetString() + "migrate/todoist")
	}

	if MigrationTrelloRedirectURL.GetString() == "" {
		MigrationTrelloRedirectURL.Set(ServicePublicURL.GetString() + "migrate/trello")
	}

	if MigrationMicrosoftTodoRedirectURL.GetString() == "" {
		MigrationMicrosoftTodoRedirectURL.Set(ServicePublicURL.GetString() + "migrate/microsoft-todo")
	}

	if tz := DefaultSettingsTimezone.GetString(); tz == "" {
		DefaultSettingsTimezone.Set(ServiceTimeZone.GetString())
	} else if _, err := time.LoadLocation(tz); err != nil {
		DefaultSettingsTimezone.Set(ServiceTimeZone.GetString())
	}

	publicURL := strings.TrimSuffix(ServicePublicURL.GetString(), "/")
	CorsOrigins.Set(append(CorsOrigins.GetStringSlice(), publicURL))

	err = SetMaxFileSizeMBytesFromString(FilesMaxSize.GetString())
	if err != nil {
		log.Fatalf("Could not parse files.maxsize: %s", err)
	}
}

func random(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", b), nil
}

func SetMaxFileSizeMBytesFromString(size string) error {
	var maxSize datasize.ByteSize
	err := maxSize.UnmarshalText([]byte(size))
	if err != nil {
		return err
	}

	maxFileSizeInBytes = uint64(maxSize.MBytes())
	return nil
}

func GetMaxFileSizeInMBytes() uint64 {
	if maxFileSizeInBytes == 0 {
		return 20
	}
	return maxFileSizeInBytes
}
