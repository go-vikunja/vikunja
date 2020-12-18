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

You can either use a `config.yml` file in the root directory of vikunja or set all config option with 
environment variables. If you have both, the value set in the config file is used.

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
We reccomend yaml or toml, but you're free to use whatever you want.

Vikunja provides a default [`config.yml`](https://kolaente.dev/vikunja/api/src/branch/master/config.yml.sample) file which you can use as a starting point.

# Config file locations

Vikunja will search on various places for a config file:

* Next to the location of the binary
* In the `service.rootpath` location set in a config (remember you can set config arguments via environment variables)
* In `/etc/vikunja`
* In `~/.config/vikunja`

# Default configuration with explanations

The following explains all possible config variables and their defaults.
You can find a full example configuration file in [here](https://code.vikunja.io/api/src/branch/master/config.yml.sample).

If you don't provide a value in your config file, their default will be used.

## Nesting

Most config variables are nested under some "higher-level" key.
For example, the `interface` config variable is a child of the `service` key.

The docs below aim to reflect that leveling, but please also have a lookt at [the default config](https://code.vikunja.io/api/src/branch/master/config.yml.sample) file
to better grasp how the nesting looks like.

<!-- Generated config will be injected here -->

---

## service



### JWTSecret

This token is used to verify issued JWT tokens.
Default is a random token which will be generated at each startup of vikunja.
(This means all already issued tokens will be invalid once you restart vikunja)

Default: `<jwt-secret>`

### interface

The interface on which to run the webserver

Default: `:3456`

### frontendurl

The URL of the frontend, used to send password reset emails.

Default: `<empty>`

### rootpath

The base path on the file system where the binary and assets are.
Vikunja will also look in this path for a config file, so you could provide only this variable to point to a folder
with a config file which will then be used.

Default: `<rootpath>`

### maxitemsperpage

The max number of items which can be returned per page

Default: `50`

### enablemetrics

If set to true, enables a /metrics endpoint for prometheus to collect metrics about the system
You'll need to use redis for this in order to enable common metrics over multiple nodes

Default: `false`

### enablecaldav

Enable the caldav endpoint, see the docs for more details

Default: `true`

### motd

Set the motd message, available from the /info endpoint

Default: `<empty>`

### enablelinksharing

Enable sharing of lists via a link

Default: `true`

### enableregistration

Whether to let new users registering themselves or not

Default: `true`

### enabletaskattachments

Whether to enable task attachments or not

Default: `true`

### timezone

The time zone all timestamps are in

Default: `GMT`

### enabletaskcomments

Whether task comments should be enabled or not

Default: `true`

### enabletotp

Whether totp is enabled. In most cases you want to leave that enabled.

Default: `true`

### sentrydsn

If not empty, enables logging of crashes and unhandled errors in sentry.

Default: `<empty>`

### testingtoken

If not empty, this will enable `/test/{table}` endpoints which allow to put any content in the database.
Used to reset the db before frontend tests. Because this is quite a dangerous feature allowing for lots of harm,
each request made to this endpoint neefs to provide an `Authorization: <token>` header with the token from below. <br/>
**You should never use this unless you know exactly what you're doing**

Default: `<empty>`

---

## database



### type

Database type to use. Supported types are mysql, postgres and sqlite.

Default: `sqlite`

### user

Database user which is used to connect to the database.

Default: `vikunja`

### password

Databse password

Default: `<empty>`

### host

Databse host

Default: `localhost`

### database

Databse to use

Default: `vikunja`

### path

When using sqlite, this is the path where to store the data

Default: `./vikunja.db`

### maxopenconnections

Sets the max open connections to the database. Only used when using mysql and postgres.

Default: `100`

### maxidleconnections

Sets the maximum number of idle connections to the db.

Default: `50`

### maxconnectionlifetime

The maximum lifetime of a single db connection in miliseconds.

Default: `10000`

### sslmode

Secure connection mode. Only used with postgres.
(see https://pkg.go.dev/github.com/lib/pq?tab=doc#hdr-Connection_String_Parameters)

Default: `disable`

---

## cache



### enabled

If cache is enabled or not

Default: `false`

### type

Cache type. Possible values are "keyvalue", "memory" or "redis".
When choosing "keyvalue" this setting follows the one configured in the "keyvalue" section.
When choosing "redis" you will need to configure the redis connection seperately.

Default: `keyvalue`

### maxelementsize

When using memory this defines the maximum size an element can take

Default: `1000`

---

## redis



### enabled

Whether to enable redis or not

Default: `false`

### host

The host of the redis server including its port.

Default: `localhost:6379`

### password

The password used to authenicate against the redis server

Default: `<empty>`

### db

0 means default database

Default: `0`

---

## cors



### enable

Whether to enable or disable cors headers.
Note: If you want to put the frontend and the api on seperate domains or ports, you will need to enable this.
      Otherwise the frontend won't be able to make requests to the api through the browser.

Default: `true`

### origins

A list of origins which may access the api. These need to include the protocol (`http://` or `https://`) and port, if any.

Default: `<empty>`

### maxage

How long (in seconds) the results of a preflight request can be cached.

Default: `0`

---

## mailer



### enabled

Whether to enable the mailer or not. If it is disabled, all users are enabled right away and password reset is not possible.

Default: `false`

### host

SMTP Host

Default: `<empty>`

### port

SMTP Host port

Default: `587`

### username

SMTP username

Default: `user`

### password

SMTP password

Default: `<empty>`

### skiptlsverify

Wether to skip verification of the tls certificate on the server

Default: `false`

### fromemail

The default from address when sending emails

Default: `mail@vikunja`

### queuelength

The length of the mail queue.

Default: `100`

### queuetimeout

The timeout in seconds after which the current open connection to the mailserver will be closed.

Default: `30`

### forcessl

By default, vikunja will try to connect with starttls, use this option to force it to use ssl.

Default: `false`

---

## log



### path

A folder where all the logfiles should go.

Default: `<rootpath>logs`

### enabled

Whether to show any logging at all or none

Default: `true`

### standard

Where the normal log should go. Possible values are stdout, stderr, file or off to disable standard logging.

Default: `stdout`

### level

Change the log level. Possible values (case-insensitive) are CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG.

Default: `INFO`

### database

Whether or not to log database queries. Useful for debugging. Possible values are stdout, stderr, file or off to disable database logging.

Default: `off`

### databaselevel

The log level for database log messages. Possible values (case-insensitive) are CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG.

Default: `WARNING`

### http

Whether to log http requests or not. Possible values are stdout, stderr, file or off to disable http logging.

Default: `stdout`

### echo

Echo has its own logging which usually is unnessecary, which is why it is disabled by default. Possible values are stdout, stderr, file or off to disable standard logging.

Default: `off`

---

## ratelimit



### enabled

whether or not to enable the rate limit

Default: `false`

### kind

The kind on which rates are based. Can be either "user" for a rate limit per user or "ip" for an ip-based rate limit.

Default: `user`

### period

The time period in seconds for the limit

Default: `60`

### limit

The max number of requests a user is allowed to do in the configured time period

Default: `100`

### store

The store where the limit counter for each user is stored.
Possible values are "keyvalue", "memory" or "redis".
When choosing "keyvalue" this setting follows the one configured in the "keyvalue" section.

Default: `keyvalue`

---

## files



### basepath

The path where files are stored

Default: `./files`

### maxsize

The maximum size of a file, as a human-readable string.
Warning: The max size is limited 2^64-1 bytes due to the underlying datatype

Default: `20MB`

---

## migration



### wunderlist

These are the settings for the wunderlist migrator

Default: `<empty>`

### todoist

Default: `<empty>`

### trello

Default: `<empty>`

### microsofttodo

Default: `<empty>`

---

## avatar



### gravatarexpiration

When using gravatar, this is the duration in seconds until a cached gravatar user avatar expires

Default: `3600`

---

## backgrounds



### enabled

Whether to enable backgrounds for lists at all.

Default: `true`

### providers

Default: `<empty>`

---

## legal

Legal urls
Will be shown in the frontend if configured here



### imprinturl

Default: `<empty>`

### privacyurl

Default: `<empty>`

---

## keyvalue

Key Value Storage settings
The Key Value Storage is used for different kinds of things like metrics and a few cache systems.



### type

The type of the storage backend. Can be either "memory" or "redis". If "redis" is chosen it needs to be configured seperately.

Default: `memory`

---

## auth



### local

Local authentication will let users log in and register (if enabled) through the db.
This is the default auth mechanism and does not require any additional configuration.

Default: `<empty>`

### openid

OpenID configuration will allow users to authenticate through a third-party OpenID Connect compatible provider.<br/>
The provider needs to support the `openid`, `profile` and `email` scopes.<br/>
**Note:** The frontend expects to be redirected after authentication by the third party
to <frontend-url>/auth/openid/<auth key>. Please make sure to configure the redirect url with your third party
auth service accordingy if you're using the default vikunja frontend.
Take a look at the [default config file](https://kolaente.dev/vikunja/api/src/branch/master/config.yml.sample) for more information about how to configure openid authentication.

Default: `<empty>`

