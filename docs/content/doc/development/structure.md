---
date: "2019-02-12:00:00+02:00"
title: "Project Structure"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Project structure

This document explains what each package does.

{{< table_of_contents >}}

## Root level

The root directory is where [the config file]({{< ref "../setup/config.md">}}), [Magefile]({{< ref "mage.md">}}), license, drone config, 
application entry point (`main.go`) and so on are located.

## pkg

This is where most of the magic happens. Most packages with actual code are located in this folder.

### caldav

This folder holds a simple caldav implementation which is responsible for the caldav feature.

### cmd

This package contains all cli-related files and functions.

To learn more about how to add a new command, see [the cli docs]({{< ref "cli.md">}}).

To learn more about how to use this cli, see [the cli usage docs]({{< ref "../usage/cli.md">}}).

### config

This package configures handling of Vikunja's runtime configuration.
It sets default values and sets up viper and tells it where to look for config files, how to interpret which env variables 
for config etc.

See also the [docs about adding a new configuration parameter]({{< ref "config.md" >}}).

### cron

See [how to add a cron task]({{< ref "cron.md" >}}).

### db

This package contains the db connection handling and db fixtures for testing.
Each other package gets its db connection object from this package.

### files

This package is responsible for all file-related things.
This means it handles saving and retrieving files from the db and the underlying file system.

### integration

All integration tests live here.
See [integration tests]({{< ref "test.md" >}}#integration-tests) for more details.

### log

Similar to `config`, this will set up the logging, based on differen logging backends.
This init is called in `main.go` after the config init is done.

### mail

This package handles all mail sending. To learn how to send a mail, see [notifications]({{< ref "notifications.md" >}}).

### metrics

This package handles all metrics which are exposed to the prometheus endpoint.
To learn how it works and how to add new metrics, take a look at [how metrics work]({{< ref "metrics.md">}}).

### migration

This package handles all migrations.
All migrations are stored and executed in this package.

To learn more, take a look at the [migrations docs]({{< ref "../development/db-migrations.md">}}).

### models

This is where most of the magic happens.
When adding new features or upgrading existing ones, that most likely happens here.

Because this package is pretty huge, there are several documents and how-to's about it:

* [Adding a feature]({{< ref "feature.md">}})
* [Making calls to the database]({{< ref "database.md">}})

### modules

Everything that can have multiple implementations (like a task migrator from a third-party task provider) lives in a 
respective sub package in this package.

#### auth

Contains openid related authentication.

#### avatar

Contains all possible avatar providers a user can choose to set their avatar.

#### background

All list background providers are in sub-packages of this package.

#### dump

Handles everything related to the `dump` and `restore` commands of Vikunja.

#### keyvalue

A simple key-value store with an implementation for memory and redis. 
Can be used to cache values.

#### migration

See [writing a migrator]({{< ref "migration.md" >}}).

### red (redis)

This package initializes a connection to a redis server.
This inizialization is automatically done at the startup of vikunja.

It also has a function (`GetRedis()`) which returns a redis client object you can then use in your package 
to talk to redis.

It uses the [go-redis](https://github.com/go-redis/redis) library, please see their configuration on how to use it.

**Note**: Only use this package directly if you have to use a direct redis connection.
In most cases, using the `keyvalue` package is a better fit.

### routes

This package defines all routes which are available for vikunja clients to use.
To add a new route, see [adding a new route]({{< ref "feature.md">}}).

#### api/v1

This is where all http-handler functions for the api are stored. 
Every handler function which does not use the standard web handler should live here.

### swagger

This is where the [generated]({{< ref "mage.md#generate-swagger-definitions-from-code-comments">}} [api docs]({{< ref "../usage/api.md">}}) live. 
You usually don't need to touch this package.

### user

All user-related things like registration etc. live in this package.

### utils

A small package, containing some helper functions:

* `MakeRandomString`: Generates a random string of a given length.
* `Sha256`: Calculates a sha256 hash from a given string.

See their function definitions for instructions on how to use them. 

### version

The single purpouse of this package is to hold the current vikunja version which gets overridden through build flags 
each time `mage release` or `mage build` is run.
It is a seperate package to avoid import cycles with other packages.
