# Configuration options

You can either use a `config.yml` file in the root directory of vikunja or set all config option with 
environment variables. If you have both, the value set in the config file is used.

Variables are nested in the `config.yml`, these nested variables become `VIKUNJA_FIRST_CHILD` when configuring via
environment variables. So setting

```bash
export VIKUNJA_FIRST_CHILD=true
```

is the same as defining it in a `config.yml` like so:

```yaml
first:
    child: true
```

# Default configuration with explanations

This is the same as the `config.yaml` file you'll find in the root of vikunja.

```yaml
service:
  # This token is used to verify issued JWT tokens.
  # Default is a random token which will be generated at each startup of vikunja.
  # (This means all already issued tokens will be invalid once you restart vikunja)
  JWTSecret: "cei6gaezoosah2bao3ieZohkae5aicah"
  # The interface on which to run the webserver
  interface: ":3456"

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
  # Whether to show mysql queries or not. Useful for debugging.
  showqueries: "false"
  # Sets the max open connections to the database. Only used when using mysql.
  openconnections: 100


cache:
  # If cache is enabled or not
  enabled: false
  # Cache type. Possible values are memory or redis
  type: memory
  # When using memory this defines the maximum size an element can take
  maxelementsize: 1000
  # When using redis, this is the host of the redis server including its port.
  redishost: 'localhost:6379'
  # When using redis, this is the password used to authenicate against the redis server
  redispassword: ''
```
