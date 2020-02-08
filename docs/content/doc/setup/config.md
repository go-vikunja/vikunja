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
  # The max number of items which can be returned per page
  maxitemsperpage: 50
  # If set to true, enables a /metrics endpoint for prometheus to collect metrics about the system
  # You'll need to use redis for this in order to enable common metrics over multiple nodes
  enablemetrics: false
  # Enable the caldav endpoint, see the docs for more details
  enablecaldav: true
  # Set the motd message, available from the /info endpoint
  motd: ""
  # Enable sharing of lists via a link
  enablelinksharing: true
  # Whether to let new users registering themselves or not
  enableregistration: true
  # Whether to enable task attachments or not
  enabletaskattachments: true
  # The time zone all timestamps are in
  timezone: GMT

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
  maxopenconnections: 100
  # Sets the maximum number of idle connections to the db.
  maxidleconnections: 50
  # The maximum lifetime of a single db connection in miliseconds.
  maxconnectionlifetime: 10000

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

cors:
  # Whether to enable or disable cors headers.
  enable: true
  # A list of origins which may access the api.
  origins:
    - *
  # How long (in seconds) the results of a preflight request can be cached.
  maxage: 0

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
  
ratelimit:
  # whether or not to enable the rate limit
  enabled: false
  # The kind on which rates are based. Can be either "user" for a rate limit per user or "ip" for an ip-based rate limit.
  kind: user
  # The time period in seconds for the limit
  period: 60
  # The max number of requests a user is allowed to do in the configured time period
  limit: 100
  # The store where the limit counter for each user is stored. Possible values are "memory" or "redis"
  store: memory

files:
  # The path where files are stored
  basepath: ./files # relative to the binary
  # The maximum size of a file, as a human-readable string.
  # Warning: The max size is limited 2^64-1 bytes due to the underlying datatype
  maxsize: 20MB

migration:
  # These are the settings for the wunderlist migrator
  wunderlist:
    # Wheter to enable the wunderlist migrator or not
    enable: true
    # The client id, required for making requests to the wunderlist api
    # You need to register your vikunja instance at https://developer.wunderlist.com/apps/new to get this
    clientid:
    # The client secret, also required for making requests to the wunderlist api
    clientsecret:
    # The url where clients are redirected after they authorized Vikunja to access their wunderlist stuff.
    # This needs to match the url you entered when registering your Vikunja instance at wunderlist.
    # This is usually the frontend url where the frontend then makes a request to /migration/wunderlist/migrate
    # with the code obtained from the wunderlist api.
    redirecturl:
{{< /highlight >}}
