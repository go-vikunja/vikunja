---
date: "2019-02-12:00:00+02:00"
title: "Project structure"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Project structure

In general, this api repo has the following structure:

* `docker`
* `docs`
* `pkg`
  * `caldav`
  * `config`
  * `log`
  * `mail`
  * `metrics`
  * `migration`
  * `models`
  * `modules`
    * `migration`
      * `handler`
      * `wunderlist`
  * `red`
  * `routes`
    * `api/v1`
  * `swagger`
  * `utils`
* `REST-Tests`
* `templates`
* `vendor`

This document will explain what these mean and what you can find where.

## Root level

The root directory is where [the config file]({{< ref "../setup/config.md">}}), [Makefile]({{< ref "make.md">}}), license, drone config, 
application entry point (`main.go`) and so on are located.

## docker

This directory holds additonal files needed to build and run the docker container, mainly service configuration to properly run Vikunja inside a docker 
container.

## pkg

This is where most of the magic happens. Most packages with actual code are located in this folder.

### caldav

This folder holds a simple caldav implementation which is responsible for returning the caldav feature.

### cmd

This package contains all cli-related files and functions.

To learn more about how to add a new command, see [the cli docs]({{< ref "cli.md">}}).

To learn more about how to use this cli, see [the cli usage docs]({{< ref "../usage/cli.md">}}).

### config

This package configures the config. It sets default values and sets up viper and tells it where to look for config files, 
how to interpret which env variables for config etc.

If you want to add a new config parameter, you should add default value in this package.

### log

Similar to `config`, this will set up the logging, based on differen logging backends.
This init is called in `main.go` after the config init is done.

### mail

This package handles all mail sending. To learn how to send a mail, see [sending emails]({{< ref "../practical-instructions/mail.md">}}).

### metrics

This package handles all metrics which are exposed to the prometheus endpoint.
To learn how it works and how to add new metrics, take a look at [how metrics work]({{< ref "../practical-instructions/metrics.md">}}).

### migration

This package handles all migrations.
All migrations are stored and executed here.

To learn more, take a look at the [migrations docs]({{< ref "../development/db-migrations.md">}}).

### models

This is where most of the magic happens.
When adding new features or upgrading existing ones, that most likely happens here.

Because this package is pretty huge, there are several documents and how-to's about it:

* [Adding a feature]({{< ref "../practical-instructions/feature.md">}})
* [Making calls to the database]({{< ref "../practical-instructions/database.md">}})

### modules

#### migration

See [writing a migrator]({{< ref "migration.md" >}}).

### red (redis)

This package initializes a connection to a redis server.
This inizialization is automatically done at the startup of vikunja.

It also has a function (`GetRedis()`) which returns a redis client object you can then use in your package 
to talk to redis.

It uses the [go-redis](https://github.com/go-redis/redis) library, please see their configuration on how to use it.

### routes

This package defines all routes which are available for vikunja clients to use.
To add a new route, see [adding a new route]({{< ref "../practical-instructions/feature.md">}}).

#### api/v1

This is where all http-handler functions for the api are stored. 
Every handler function which does not use the standard web handler should live here.

### swagger

This is where the [generated]({{< ref "make.md#generate-swagger-definitions-from-code-comments">}} [api docs]({{< ref "../usage/api.md">}}) live. 
You usually don't need to touch this package.

### utils

A small package, containing some helper functions:

* `MakeRandomString`: Generates a random string of a given length.
* `Sha256`: Calculates a sha256 hash from a given string.

See their function definitions for instructions on how to use them. 

## REST-Tests

Holds all kinds of test files to directly test the api from inside of [jetbrains ide's](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html).

These files are currently more an experiment, maybe we will drop them in the future to use something we could integrate in the testing process with drone.
Therefore, this has no claim to be complete yet even working, you're free to change whatever is needed to get it working for you.

## templates

Holds the email templates used to send plain text and html emails for new user registration and password changes.

## vendor

All libraries needed to build Vikunja. 

We keep all libraries used for Vikunja around in the `vendor/` folder to still be able to build the project even if
some maintainers take their libraries down like [it happened in the past](https://github.com/jteeuwen/go-bindata/issues/5).

When adding a new  dependency, make sure to run `go mod vendor` to put it inside this directory.
