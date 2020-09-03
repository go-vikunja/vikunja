---
date: "2020-01-19:16:00+02:00"
title: "Migrations"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Writing a migrator for Vikunja

It is possible to migrate data from other to-do services to Vikunja.
To make this easier, we have put together a few helpers which are documented on this page.

In general, each migrator implements a migrator interface which is then called from a client.
The interface makes it possible to use helper methods which handle http an focus only on the implementation of the migrator itself.

{{< table_of_contents >}}

## Structure

All migrator implementations live in their own package in `pkg/modules/migration/<name-of-the-service>`.
When creating a new migrator, you should place all related code inside that module.

## Migrator interface

The migrator interface is defined as follows:

```go
// Migrator is the basic migrator interface which is shared among all migrators
type Migrator interface {
	// Migrate is the interface used to migrate a user's tasks from another platform to vikunja.
	// The user object is the user who's tasks will be migrated.
	Migrate(user *models.User) error
	// AuthURL returns a url for clients to authenticate against.
	// The use case for this are Oauth flows, where the server token should remain hidden and not
	// known to the frontend.
	AuthURL() string
	// Name holds the name of the migration.
	// This is used to show the name to users and to keep track of users who already migrated.
	Name() string
}
```

## Defining http routes

Once your migrator implements the migration interface, it becomes possible to use the helper http handlers.
Their usage is very similar to the [general web handler](https://kolaente.dev/vikunja/web#user-content-defining-routes-using-the-standard-web-handler):

The `RegisterRoutes(m)` method registers all routes with the scheme `/[MigratorName]/(auth|migrate|status)` for the 
authUrl, Status and Migrate methods.

```go
// This is an example for the Wunderlist migrator
if config.MigrationWunderlistEnable.GetBool() {
    wunderlistMigrationHandler := &migrationHandler.MigrationWeb{
		MigrationStruct: func() migration.Migrator {
			return &wunderlist.Migration{}
		},
	}
    wunderlistMigrationHandler.RegisterRoutes(m)
}
```

You should also document the routes with [swagger annotations]({{< ref "../practical-instructions/swagger-docs.md" >}}).

## Insertion helper method

There is a method available in the `migration` package which takes a fully nested Vikunja structure and creates it with all relations. 
This means you start by adding a namespace, then add lists inside of that namespace, then tasks in the lists and so on.

The root structure must be present as `[]*models.NamespaceWithLists`.

Then call the method like so:

```go
fullVikunjaHierachie, err := convertWunderlistToVikunja(wContent)
if err != nil {
    return
}

err = migration.InsertFromStructure(fullVikunjaHierachie, user)
```

## Configuration

You should add at least an option to enable or disable the migration.
Chances are, you'll need some more options for things like client ID and secret 
(if the other service uses oAuth as an authentication flow).

The easiest way to implement an on/off switch is to check whether your migration service is enabled or not when 
registering the routes, and then simply don't registering the routes in the case it is disabled.

### Making the migrator public in `/info` 

You should make your migrator available in the `/info` endpoint so that frontends can display options to enable them or not.
To do this, add an entry to `pkg/routes/api/v1/info.go`.
