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

## Config file locations

Vikunja will search on various places for a config file:

* Next to the location of the binary
* In the `service.rootpath` location set in a config (remember you can set config arguments via environment variables)
* In `/etc/vikunja`
* In `~/.config/vikunja`

# Default configuration with explanations

This is the same as the `config.yml.sample` file you'll find in the root of vikunja.

{{< highlight yaml >}}
service:
  # This token is used to verify issued JWT tokens.
  # Default is a random token which will be generated at each startup of vikunja.
  # (This means all already issued tokens will be invalid once you restart vikunja)
  JWTSecret: "cei6gaezoosah2bao3ieZohkae5aicah"
  # The interface on which to run the webserver
  interface: ":3456"
  # The URL of the frontend, used to send password reset emails.
  frontendurl: ""
  # The base path on the file system where the binary and assets are.
  # Vikunja will also look in this path for a config file, so you could provide only this variable to point to a folder
  # with a config file which will then be used.
  rootpath: <the path of the executable>
  # The number of items which gets returned per page
  pagecount: 50
  # If set to true, enables a /metrics endpoint for prometheus to collect metrics about the system
  # You'll need to use redis for this in order to enable common metrics over multiple nodes
  enablemetrics: false

database:
  # Database type to use. Supported types are mysql and sqlite.
  type: "sqlite"
  # Database user which is used to connect to the database.
  user: "vikunja"
  # Databse password
  password: ""
  # Databse host
  host: "localhost"
  # Databse to use
  database: "vikunja"
  # When using sqlite, this is the path where to store the data
  Path: "./vikunja.db"
  # Sets the max open connections to the database. Only used when using mysql.
  openconnections: 100

cache:
  # If cache is enabled or not
  enabled: false
  # Cache type. Possible values are memory or redis, you'll need to enable redis below when using redis
  type: memory
  # When using memory this defines the maximum size an element can take
  maxelementsize: 1000

redis:
  # Whether to enable redis or not
  enabled: false
  # The host of the redis server including its port.
  host: 'localhost:6379'
  # The password used to authenicate against the redis server
  password: ''
  # 0 means default database
  db: 0

mailer:
  # Whether to enable the mailer or not. If it is disabled, all users are enabled right away and password reset is not possible.
  enabled: false
  # SMTP Host
  host: ""
  # SMTP Host port
  port: 587
  # SMTP username
  username: "user"
  # SMTP password
  password: ""
  # Wether to skip verification of the tls certificate on the server
  skiptlsverify: false
  # The default from address when sending emails
  fromemail: "mail@vikunja"
  # The length of the mail queue.
  queuelength: 100
  # The timeout in seconds after which the current open connection to the mailserver will be closed.
  queuetimeout: 30

log:
  # A folder where all the logfiles should go.
  path: <rootpath>logs
  # Whether to show any logging at all or none
  enabled: true
  # Where the error log should go. Possible values are stdout, stderr, file or off to disable error logging.
  errors: "stdout"
  # Where the normal log should go. Possible values are stdout, stderr, file or off to disable standard logging.
  standard: "stdout"
  # Whether or not to log database queries. Useful for debugging. Possible values are stdout, stderr, file or off to disable database logging.
  database: "off"
  # Whether to log http requests or not. Possible values are stdout, stderr, file or off to disable http logging.
  http: "stdout"
  # Echo has its own logging which usually is unnessecary, which is why it is disabled by default. Possible values are stdout, stderr, file or off to disable standard logging.
  echo: "off"
{{< /highlight >}}
