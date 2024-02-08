---
date: "2019-02-12:00:00+02:00"
title: "Config options"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Configuration options

You can either use a `config.yml` file in the root directory of vikunja or set almost all config option with environment variables. If you have both, the value set in the config file is used.
Right now it is not possible to configure openid authentication via environment variables.

Variables are nested in the `config.yml`, these nested variables become `VIKUNJA_FIRST_CHILD` when configuring via
environment variables. So setting

{{< highlight bash >}}
export VIKUNJA_FIRST_CHILD=true
{{< /highlight >}}

is the same as defining it in a `config.yml` like so:

{{< highlight yaml >}}
first:
    child: true
{{< /highlight >}}

# Formats

Vikunja supports using `toml`, `yaml`, `hcl`, `ini`, `json`, envfile, env variables and Java Properties files.
We recommend yaml or toml, but you're free to use whatever you want.

Vikunja provides a default [`config.yml`](https://kolaente.dev/vikunja/vikunja/src/branch/main/config.yml.sample) file which you can use as a starting point.

# Config file locations

Vikunja will search on various places for a config file:

* Next to the location of the binary
* In the `service.rootpath` location set in a config (remember you can set config arguments via environment variables)
* In `/etc/vikunja`
* In `~/.config/vikunja`

# Default configuration with explanations

The following explains all possible config variables and their defaults.
You can find a full example configuration file in [here](https://code.vikunja.io/api/src/branch/main/config.yml.sample).

If you don't provide a value in your config file, their default will be used.

## Nesting

Most config variables are nested under some "higher-level" key.
For example, the `interface` config variable is a child of the `service` key.

The docs below aim to reflect that leveling, but please also have a look at [the default config](https://code.vikunja.io/api/src/branch/main/config.yml.sample) file
to better grasp how the nesting looks like.

<!-- Generated config will be injected here -->

---

## service



### JWTSecret

This token is used to verify issued JWT tokens.
Default is a random token which will be generated at each startup of vikunja.
(This means all already issued tokens will be invalid once you restart vikunja)

Default: `<jwt-secret>`

Full path: `service.JWTSecret`

Environment path: `VIKUNJA_SERVICE_JWTSECRET`


### jwtttl

The duration of the issued JWT tokens in seconds.
The default is 259200 seconds (3 Days).

Default: `259200`

Full path: `service.jwtttl`

Environment path: `VIKUNJA_SERVICE_JWTTTL`


### jwtttllong

The duration of the "remember me" time in seconds. When the login request is made with 
the long param set, the token returned will be valid for this period.
The default is 2592000 seconds (30 Days).

Default: `2592000`

Full path: `service.jwtttllong`

Environment path: `VIKUNJA_SERVICE_JWTTTLLONG`


### interface

The interface on which to run the webserver

Default: `:3456`

Full path: `service.interface`

Environment path: `VIKUNJA_SERVICE_INTERFACE`


### unixsocket

Path to Unix socket. If set, it will be created and used instead of tcp

Default: `<empty>`

Full path: `service.unixsocket`

Environment path: `VIKUNJA_SERVICE_UNIXSOCKET`


### unixsocketmode

Permission bits for the Unix socket. Note that octal values must be prefixed by "0o", e.g. 0o660

Default: `<empty>`

Full path: `service.unixsocketmode`

Environment path: `VIKUNJA_SERVICE_UNIXSOCKETMODE`


### frontendurl

The URL of the frontend, used to send password reset emails.

Default: `<empty>`

Full path: `service.frontendurl`

Environment path: `VIKUNJA_SERVICE_FRONTENDURL`


### rootpath

The base path on the file system where the binary and assets are.
Vikunja will also look in this path for a config file, so you could provide only this variable to point to a folder
with a config file which will then be used.

Default: `<rootpath>`

Full path: `service.rootpath`

Environment path: `VIKUNJA_SERVICE_ROOTPATH`


### staticpath

Path on the file system to serve static files from. Set to the path of the frontend files to host frontend alongside the api.

Default: `<empty>`

Full path: `service.staticpath`

Environment path: `VIKUNJA_SERVICE_STATICPATH`


### maxitemsperpage

The max number of items which can be returned per page

Default: `50`

Full path: `service.maxitemsperpage`

Environment path: `VIKUNJA_SERVICE_MAXITEMSPERPAGE`


### enablecaldav

Enable the caldav endpoint, see the docs for more details

Default: `true`

Full path: `service.enablecaldav`

Environment path: `VIKUNJA_SERVICE_ENABLECALDAV`


### motd

Set the motd message, available from the /info endpoint

Default: `<empty>`

Full path: `service.motd`

Environment path: `VIKUNJA_SERVICE_MOTD`


### enablelinksharing

Enable sharing of project via a link

Default: `true`

Full path: `service.enablelinksharing`

Environment path: `VIKUNJA_SERVICE_ENABLELINKSHARING`


### enableregistration

Whether to let new users registering themselves or not

Default: `true`

Full path: `service.enableregistration`

Environment path: `VIKUNJA_SERVICE_ENABLEREGISTRATION`


### enabletaskattachments

Whether to enable task attachments or not

Default: `true`

Full path: `service.enabletaskattachments`

Environment path: `VIKUNJA_SERVICE_ENABLETASKATTACHMENTS`


### timezone

The time zone all timestamps are in. Please note that time zones have to use [the official tz database names](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). UTC or GMT offsets won't work.

Default: `GMT`

Full path: `service.timezone`

Environment path: `VIKUNJA_SERVICE_TIMEZONE`


### enabletaskcomments

Whether task comments should be enabled or not

Default: `true`

Full path: `service.enabletaskcomments`

Environment path: `VIKUNJA_SERVICE_ENABLETASKCOMMENTS`


### enabletotp

Whether totp is enabled. In most cases you want to leave that enabled.

Default: `true`

Full path: `service.enabletotp`

Environment path: `VIKUNJA_SERVICE_ENABLETOTP`


### sentrydsn

If not empty, enables logging of crashes and unhandled errors in sentry.

Default: `<empty>`

Full path: `service.sentrydsn`

Environment path: `VIKUNJA_SERVICE_SENTRYDSN`


### testingtoken

If not empty, this will enable `/test/{table}` endpoints which allow to put any content in the database.
Used to reset the db before frontend tests. Because this is quite a dangerous feature allowing for lots of harm,
each request made to this endpoint needs to provide an `Authorization: <token>` header with the token from below. <br/>
**You should never use this unless you know exactly what you're doing**

Default: `<empty>`

Full path: `service.testingtoken`

Environment path: `VIKUNJA_SERVICE_TESTINGTOKEN`


### enableemailreminders

If enabled, vikunja will send an email to everyone who is either assigned to a task or created it when a task reminder
is due.

Default: `true`

Full path: `service.enableemailreminders`

Environment path: `VIKUNJA_SERVICE_ENABLEEMAILREMINDERS`


### enableuserdeletion

If true, will allow users to request the complete deletion of their account. When using external authentication methods 
it may be required to coordinate with them in order to delete the account. This setting will not affect the cli commands
for user deletion.

Default: `true`

Full path: `service.enableuserdeletion`

Environment path: `VIKUNJA_SERVICE_ENABLEUSERDELETION`


### maxavatarsize

The maximum size clients will be able to request for user avatars.
If clients request a size bigger than this, it will be changed on the fly.

Default: `1024`

Full path: `service.maxavatarsize`

Environment path: `VIKUNJA_SERVICE_MAXAVATARSIZE`


### demomode

If set to true, the frontend will show a big red warning not to use this instance for real data as it will be cleared out.
You probably don't need to set this value, it was created specifically for usage on [try](https://try.vikunja.io).

Default: `false`

Full path: `service.demomode`

Environment path: `VIKUNJA_SERVICE_DEMOMODE`


---

## database



### type

Database type to use. Supported types are mysql, postgres and sqlite.

Default: `sqlite`

Full path: `database.type`

Environment path: `VIKUNJA_DATABASE_TYPE`


### user

Database user which is used to connect to the database.

Default: `vikunja`

Full path: `database.user`

Environment path: `VIKUNJA_DATABASE_USER`


### password

Database password

Default: `<empty>`

Full path: `database.password`

Environment path: `VIKUNJA_DATABASE_PASSWORD`


### host

Database host

Default: `localhost`

Full path: `database.host`

Environment path: `VIKUNJA_DATABASE_HOST`


### database

Database to use

Default: `vikunja`

Full path: `database.database`

Environment path: `VIKUNJA_DATABASE_DATABASE`


### path

When using sqlite, this is the path where to store the data

Default: `./vikunja.db`

Full path: `database.path`

Environment path: `VIKUNJA_DATABASE_PATH`


### maxopenconnections

Sets the max open connections to the database. Only used when using mysql and postgres.

Default: `100`

Full path: `database.maxopenconnections`

Environment path: `VIKUNJA_DATABASE_MAXOPENCONNECTIONS`


### maxidleconnections

Sets the maximum number of idle connections to the db.

Default: `50`

Full path: `database.maxidleconnections`

Environment path: `VIKUNJA_DATABASE_MAXIDLECONNECTIONS`


### maxconnectionlifetime

The maximum lifetime of a single db connection in milliseconds.

Default: `10000`

Full path: `database.maxconnectionlifetime`

Environment path: `VIKUNJA_DATABASE_MAXCONNECTIONLIFETIME`


### sslmode

Secure connection mode. Only used with postgres.
(see https://pkg.go.dev/github.com/lib/pq?tab=doc#hdr-Connection_String_Parameters)

Default: `disable`

Full path: `database.sslmode`

Environment path: `VIKUNJA_DATABASE_SSLMODE`


### sslcert

The path to the client cert. Only used with postgres.

Default: `<empty>`

Full path: `database.sslcert`

Environment path: `VIKUNJA_DATABASE_SSLCERT`


### sslkey

The path to the client key. Only used with postgres.

Default: `<empty>`

Full path: `database.sslkey`

Environment path: `VIKUNJA_DATABASE_SSLKEY`


### sslrootcert

The path to the ca cert. Only used with postgres.

Default: `<empty>`

Full path: `database.sslrootcert`

Environment path: `VIKUNJA_DATABASE_SSLROOTCERT`


### tls

Enable SSL/TLS for mysql connections. Options: false, true, skip-verify, preferred

Default: `false`

Full path: `database.tls`

Environment path: `VIKUNJA_DATABASE_TLS`


---

## typesense



### enabled

Whether to enable the Typesense integration. If true, all tasks will be synced to the configured Typesense
instance and all search and filtering will run through Typesense instead of only through the database.
Typesense allows fast fulltext search including fuzzy matching support. It may return different results than 
what you'd get with a database-only search.

Default: `false`

Full path: `typesense.enabled`

Environment path: `VIKUNJA_TYPESENSE_ENABLED`


### url

The url to the Typesense instance you want to use. Can be hosted locally or in Typesense Cloud as long
as Vikunja is able to reach it.

Default: `<empty>`

Full path: `typesense.url`

Environment path: `VIKUNJA_TYPESENSE_URL`


### apikey

The Typesense API key you want to use.

Default: `<empty>`

Full path: `typesense.apikey`

Environment path: `VIKUNJA_TYPESENSE_APIKEY`


---

## redis



### enabled

Whether to enable redis or not

Default: `false`

Full path: `redis.enabled`

Environment path: `VIKUNJA_REDIS_ENABLED`


### host

The host of the redis server including its port.

Default: `localhost:6379`

Full path: `redis.host`

Environment path: `VIKUNJA_REDIS_HOST`


### password

The password used to authenticate against the redis server

Default: `<empty>`

Full path: `redis.password`

Environment path: `VIKUNJA_REDIS_PASSWORD`


### db

0 means default database

Default: `0`

Full path: `redis.db`

Environment path: `VIKUNJA_REDIS_DB`


---

## cors



### enable

Whether to enable or disable cors headers.
Note: If you want to put the frontend and the api on separate domains or ports, you will need to enable this.
      Otherwise the frontend won't be able to make requests to the api through the browser.

Default: `true`

Full path: `cors.enable`

Environment path: `VIKUNJA_CORS_ENABLE`


### origins

A list of origins which may access the api. These need to include the protocol (`http://` or `https://`) and port, if any.

Default: `<empty>`

Full path: `cors.origins`

Environment path: `VIKUNJA_CORS_ORIGINS`


### maxage

How long (in seconds) the results of a preflight request can be cached.

Default: `0`

Full path: `cors.maxage`

Environment path: `VIKUNJA_CORS_MAXAGE`


---

## mailer



### enabled

Whether to enable the mailer or not. If it is disabled, all users are enabled right away and password reset is not possible.

Default: `false`

Full path: `mailer.enabled`

Environment path: `VIKUNJA_MAILER_ENABLED`


### host

SMTP Host

Default: `<empty>`

Full path: `mailer.host`

Environment path: `VIKUNJA_MAILER_HOST`


### port

SMTP Host port.
**NOTE:** If you're unable to send mail and the only error you see in the logs is an `EOF`, try setting the port to `25`.

Default: `587`

Full path: `mailer.port`

Environment path: `VIKUNJA_MAILER_PORT`


### authtype

SMTP Auth Type. Can be either `plain`, `login` or `cram-md5`.

Default: `plain`

Full path: `mailer.authtype`

Environment path: `VIKUNJA_MAILER_AUTHTYPE`


### username

SMTP username

Default: `user`

Full path: `mailer.username`

Environment path: `VIKUNJA_MAILER_USERNAME`


### password

SMTP password

Default: `<empty>`

Full path: `mailer.password`

Environment path: `VIKUNJA_MAILER_PASSWORD`


### skiptlsverify

Wether to skip verification of the tls certificate on the server

Default: `false`

Full path: `mailer.skiptlsverify`

Environment path: `VIKUNJA_MAILER_SKIPTLSVERIFY`


### fromemail

The default from address when sending emails

Default: `mail@vikunja`

Full path: `mailer.fromemail`

Environment path: `VIKUNJA_MAILER_FROMEMAIL`


### queuelength

The length of the mail queue.

Default: `100`

Full path: `mailer.queuelength`

Environment path: `VIKUNJA_MAILER_QUEUELENGTH`


### queuetimeout

The timeout in seconds after which the current open connection to the mailserver will be closed.

Default: `30`

Full path: `mailer.queuetimeout`

Environment path: `VIKUNJA_MAILER_QUEUETIMEOUT`


### forcessl

By default, vikunja will try to connect with starttls, use this option to force it to use ssl.

Default: `false`

Full path: `mailer.forcessl`

Environment path: `VIKUNJA_MAILER_FORCESSL`


---

## log



### path

A folder where all the logfiles should go.

Default: `<rootpath>logs`

Full path: `log.path`

Environment path: `VIKUNJA_LOG_PATH`


### enabled

Whether to show any logging at all or none

Default: `true`

Full path: `log.enabled`

Environment path: `VIKUNJA_LOG_ENABLED`


### standard

Where the normal log should go. Possible values are stdout, stderr, file or off to disable standard logging.

Default: `stdout`

Full path: `log.standard`

Environment path: `VIKUNJA_LOG_STANDARD`


### level

Change the log level. Possible values (case-insensitive) are CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG.

Default: `INFO`

Full path: `log.level`

Environment path: `VIKUNJA_LOG_LEVEL`


### database

Whether or not to log database queries. Useful for debugging. Possible values are stdout, stderr, file or off to disable database logging.

Default: `off`

Full path: `log.database`

Environment path: `VIKUNJA_LOG_DATABASE`


### databaselevel

The log level for database log messages. Possible values (case-insensitive) are CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG.

Default: `WARNING`

Full path: `log.databaselevel`

Environment path: `VIKUNJA_LOG_DATABASELEVEL`


### http

Whether to log http requests or not. Possible values are stdout, stderr, file or off to disable http logging.

Default: `stdout`

Full path: `log.http`

Environment path: `VIKUNJA_LOG_HTTP`


### echo

Echo has its own logging which usually is unnecessary, which is why it is disabled by default. Possible values are stdout, stderr, file or off to disable standard logging.

Default: `off`

Full path: `log.echo`

Environment path: `VIKUNJA_LOG_ECHO`


### events

Whether or not to log events. Useful for debugging. Possible values are stdout, stderr, file or off to disable events logging.

Default: `off`

Full path: `log.events`

Environment path: `VIKUNJA_LOG_EVENTS`


### eventslevel

The log level for event log messages. Possible values (case-insensitive) are ERROR, INFO, DEBUG.

Default: `info`

Full path: `log.eventslevel`

Environment path: `VIKUNJA_LOG_EVENTSLEVEL`


### mail

Whether or not to log mail log messages. This will not log mail contents. Possible values are stdout, stderr, file or off to disable mail-related logging.

Default: `off`

Full path: `log.mail`

Environment path: `VIKUNJA_LOG_MAIL`


### maillevel

The log level for mail log messages. Possible values (case-insensitive) are ERROR, WARNING, INFO, DEBUG.

Default: `info`

Full path: `log.maillevel`

Environment path: `VIKUNJA_LOG_MAILLEVEL`


---

## ratelimit



### enabled

whether or not to enable the rate limit

Default: `false`

Full path: `ratelimit.enabled`

Environment path: `VIKUNJA_RATELIMIT_ENABLED`


### kind

The kind on which rates are based. Can be either "user" for a rate limit per user or "ip" for an ip-based rate limit.

Default: `user`

Full path: `ratelimit.kind`

Environment path: `VIKUNJA_RATELIMIT_KIND`


### period

The time period in seconds for the limit

Default: `60`

Full path: `ratelimit.period`

Environment path: `VIKUNJA_RATELIMIT_PERIOD`


### limit

The max number of requests a user is allowed to do in the configured time period

Default: `100`

Full path: `ratelimit.limit`

Environment path: `VIKUNJA_RATELIMIT_LIMIT`


### store

The store where the limit counter for each user is stored.
Possible values are "keyvalue", "memory" or "redis".
When choosing "keyvalue" this setting follows the one configured in the "keyvalue" section.

Default: `keyvalue`

Full path: `ratelimit.store`

Environment path: `VIKUNJA_RATELIMIT_STORE`


### noauthlimit

The number of requests a user can make from the same IP to all unauthenticated routes (login, register, 
password confirmation, email verification, password reset request) per minute. This limit cannot be disabled.
You should only change this if you know what you're doing.

Default: `10`

Full path: `ratelimit.noauthlimit`

Environment path: `VIKUNJA_RATELIMIT_NOAUTHLIMIT`


---

## files



### basepath

The path where files are stored

Default: `./files`

Full path: `files.basepath`

Environment path: `VIKUNJA_FILES_BASEPATH`


### maxsize

The maximum size of a file, as a human-readable string.
Warning: The max size is limited 2^64-1 bytes due to the underlying datatype

Default: `20MB`

Full path: `files.maxsize`

Environment path: `VIKUNJA_FILES_MAXSIZE`


---

## migration



### todoist

Default: `<empty>`

Full path: `migration.todoist`

Environment path: `VIKUNJA_MIGRATION_TODOIST`


### trello

Default: `<empty>`

Full path: `migration.trello`

Environment path: `VIKUNJA_MIGRATION_TRELLO`


### microsofttodo

Default: `<empty>`

Full path: `migration.microsofttodo`

Environment path: `VIKUNJA_MIGRATION_MICROSOFTTODO`


---

## avatar



### gravatarexpiration

When using gravatar, this is the duration in seconds until a cached gravatar user avatar expires

Default: `3600`

Full path: `avatar.gravatarexpiration`

Environment path: `VIKUNJA_AVATAR_GRAVATAREXPIRATION`


---

## backgrounds



### enabled

Whether to enable backgrounds for projects at all.

Default: `true`

Full path: `backgrounds.enabled`

Environment path: `VIKUNJA_BACKGROUNDS_ENABLED`


### providers

Default: `<empty>`

Full path: `backgrounds.providers`

Environment path: `VIKUNJA_BACKGROUNDS_PROVIDERS`


---

## legal

Legal urls
Will be shown in the frontend if configured here



### imprinturl

Default: `<empty>`

Full path: `legal.imprinturl`

Environment path: `VIKUNJA_LEGAL_IMPRINTURL`


### privacyurl

Default: `<empty>`

Full path: `legal.privacyurl`

Environment path: `VIKUNJA_LEGAL_PRIVACYURL`


---

## keyvalue

Key Value Storage settings
The Key Value Storage is used for different kinds of things like metrics and a few cache systems.



### type

The type of the storage backend. Can be either "memory" or "redis". If "redis" is chosen it needs to be configured separately.

Default: `memory`

Full path: `keyvalue.type`

Environment path: `VIKUNJA_KEYVALUE_TYPE`


---

## auth



### local

Local authentication will let users log in and register (if enabled) through the db.
This is the default auth mechanism and does not require any additional configuration.

Default: `<empty>`

Full path: `auth.local`

Environment path: `VIKUNJA_AUTH_LOCAL`


### openid

OpenID configuration will allow users to authenticate through a third-party OpenID Connect compatible provider.<br/>
The provider needs to support the `openid`, `profile` and `email` scopes.<br/>
**Note:** Some openid providers (like gitlab) only make the email of the user available through openid claims if they have set it to be publicly visible.
If the email is not public in those cases, authenticating will fail.
**Note 2:** The frontend expects to be redirected after authentication by the third party
to <frontend-url>/auth/openid/<auth key>. Please make sure to configure the redirect url with your third party
auth service accordingly if you're using the default vikunja frontend.
Take a look at the [default config file](https://kolaente.dev/vikunja/vikunja/src/branch/main/config.yml.sample) for more information about how to configure openid authentication.

Default: `<empty>`

Full path: `auth.openid`

Environment path: `VIKUNJA_AUTH_OPENID`


---

## metrics

Prometheus metrics endpoint



### enabled

If set to true, enables a /metrics endpoint for prometheus to collect metrics about Vikunja. You can query it from `/api/v1/metrics`.

Default: `false`

Full path: `metrics.enabled`

Environment path: `VIKUNJA_METRICS_ENABLED`


### username

If set to a non-empty value the /metrics endpoint will require this as a username via basic auth in combination with the password below.

Default: `<empty>`

Full path: `metrics.username`

Environment path: `VIKUNJA_METRICS_USERNAME`


### password

If set to a non-empty value the /metrics endpoint will require this as a password via basic auth in combination with the username below.

Default: `<empty>`

Full path: `metrics.password`

Environment path: `VIKUNJA_METRICS_PASSWORD`


---

## defaultsettings

Provide default settings for new users. When a new user is created, these settings will automatically be set for the user. If you change them in the config file afterwards they will not be changed back for existing users.



### avatar_provider

The avatar source for the user. Can be `gravatar`, `initials`, `upload` or `marble`. If you set this to `upload` you'll also need to specify `defaultsettings.avatar_file_id`.

Default: `initials`

Full path: `defaultsettings.avatar_provider`

Environment path: `VIKUNJA_DEFAULTSETTINGS_AVATAR_PROVIDER`


### avatar_file_id

The id of the file used as avatar.

Default: `0`

Full path: `defaultsettings.avatar_file_id`

Environment path: `VIKUNJA_DEFAULTSETTINGS_AVATAR_FILE_ID`


### email_reminders_enabled

If set to true users will get task reminders via email.

Default: `false`

Full path: `defaultsettings.email_reminders_enabled`

Environment path: `VIKUNJA_DEFAULTSETTINGS_EMAIL_REMINDERS_ENABLED`


### discoverable_by_name

If set to true will allow other users to find this user when searching for parts of their name.

Default: `false`

Full path: `defaultsettings.discoverable_by_name`

Environment path: `VIKUNJA_DEFAULTSETTINGS_DISCOVERABLE_BY_NAME`


### discoverable_by_email

If set to true will allow other users to find this user when searching for their exact email.

Default: `false`

Full path: `defaultsettings.discoverable_by_email`

Environment path: `VIKUNJA_DEFAULTSETTINGS_DISCOVERABLE_BY_EMAIL`


### overdue_tasks_reminders_enabled

If set to true will send an email every day with all overdue tasks at a configured time.

Default: `true`

Full path: `defaultsettings.overdue_tasks_reminders_enabled`

Environment path: `VIKUNJA_DEFAULTSETTINGS_OVERDUE_TASKS_REMINDERS_ENABLED`


### overdue_tasks_reminders_time

When to send the overdue task reminder email.

Default: `9:00`

Full path: `defaultsettings.overdue_tasks_reminders_time`

Environment path: `VIKUNJA_DEFAULTSETTINGS_OVERDUE_TASKS_REMINDERS_TIME`


### default_project_id

The id of the default project. Make sure users actually have access to this project when setting this value.

Default: `0`

Full path: `defaultsettings.default_project_id`

Environment path: `VIKUNJA_DEFAULTSETTINGS_DEFAULT_PROJECT_ID`


### week_start

Start of the week for the user. `0` is sunday, `1` is monday and so on.

Default: `0`

Full path: `defaultsettings.week_start`

Environment path: `VIKUNJA_DEFAULTSETTINGS_WEEK_START`


### language

The language of the user interface. Must be an ISO 639-1 language code followed by an ISO 3166-1 alpha-2 country code. Check https://kolaente.dev/vikunja/vikunja/src/branch/main/frontend/src/i18n/lang for a list of possible languages. Will default to the browser language the user uses when signing up.

Default: `<unset>`

Full path: `defaultsettings.language`

Environment path: `VIKUNJA_DEFAULTSETTINGS_LANGUAGE`


### timezone

The time zone of each individual user. This will affect when users get reminders and overdue task emails.

Default: `<time zone set at service.timezone>`

Full path: `defaultsettings.timezone`

Environment path: `VIKUNJA_DEFAULTSETTINGS_TIMEZONE`


---

## webhooks



### enabled

Whether to enable support for webhooks

Default: `true`

Full path: `webhooks.enabled`

Environment path: `VIKUNJA_WEBHOOKS_ENABLED`


### timoutseconds

The timout in seconds until a webhook request fails when no response has been received.

Default: `30`

Full path: `webhooks.timoutseconds`

Environment path: `VIKUNJA_WEBHOOKS_TIMOUTSECONDS`


### proxyurl

The URL of [a mole instance](https://github.com/frain-dev/mole) to use to proxy outgoing webhook requests. You should use this and configure appropriately if you're not the only one using your Vikunja instance. More info about why: https://webhooks.fyi/best-practices/webhook-providers#implement-security-on-egress-communication. Must be used in combination with `webhooks.password` (see below).

Default: `<empty>`

Full path: `webhooks.proxyurl`

Environment path: `VIKUNJA_WEBHOOKS_PROXYURL`


### proxypassword

The proxy password to use when authenticating against the proxy.

Default: `<empty>`

Full path: `webhooks.proxypassword`

Environment path: `VIKUNJA_WEBHOOKS_PROXYPASSWORD`


