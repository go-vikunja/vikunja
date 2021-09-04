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
The interface makes it possible to use helper methods which handle http and focus only on the implementation of the migrator itself.

There are two ways of migrating data from another service:
1. Through the auth-based flow where the user gives you access to their data at the third-party service through an 
   oauth flow. You can then call the service's api on behalf of your user to get all the data.
   The Todoist, Trello and Microsoft To-Do Migrators use this pattern.
2. A file migration where the user uploads a file obtained from some third-party service. In your migrator, you need 
   to parse the file and create the lists, tasks etc.
   The Vikunja File Import uses this pattern.

To differentiate the two, there are two different interfaces you must implement.

{{< table_of_contents >}}

## Structure

All migrator implementations live in their own package in `pkg/modules/migration/<name-of-the-service>`.
When creating a new migrator, you should place all related code inside that module.

## Migrator Interface

The migrator interface is defined as follows:

```go
// Migrator is the basic migrator interface which is shared among all migrators
type Migrator interface {
	// Name holds the name of the migration.
	// This is used to show the name to users and to keep track of users who already migrated.
	Name() string
	// Migrate is the interface used to migrate a user's tasks from another platform to vikunja.
	// The user object is the user who's tasks will be migrated.
	Migrate(user *models.User) error
	// AuthURL returns a url for clients to authenticate against.
	// The use case for this are Oauth flows, where the server token should remain hidden and not
	// known to the frontend.
	AuthURL() string
}
```

## File Migrator Interface

```go
// FileMigrator handles importing Vikunja data from a file. The implementation of it determines the format.
type FileMigrator interface {
	// Name holds the name of the migration.
	// This is used to show the name to users and to keep track of users who already migrated.
	Name() string
	// Migrate is the interface used to migrate a user's tasks, list and other things from a file to vikunja.
	// The user object is the user who's tasks will be migrated.
	Migrate(user *user.User, file io.ReaderAt, size int64) error
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

And for the file migrator:

```go
vikunjaFileMigrationHandler := &migrationHandler.FileMigratorWeb{
	MigrationStruct: func() migration.FileMigrator {
		return &vikunja_file.FileMigrator{}
	},
}
vikunjaFileMigrationHandler.RegisterRoutes(m)
```

You should also document the routes with [swagger annotations]({{< ref "swagger-docs.md" >}}).

## Insertion helper method

There is a method available in the `migration` package which takes a fully nested Vikunja structure and creates it with all relations. 
This means you start by adding a namespace, then add lists inside of that namespace, then tasks in the lists and so on.

The root structure must be present as `[]*models.NamespaceWithListsAndTasks`. It allows to represent all of Vikunja's 
hierachie as a single data structure.

Then call the method like so:

```go
fullVikunjaHierachie, err := convertWunderlistToVikunja(wContent)
if err != nil {
    return
}

err = migration.InsertFromStructure(fullVikunjaHierachie, user)
```

## Configuration

If your migrator is an oauth-based one, you should add at least an option to enable or disable it.
Chances are, you'll need some more options for things like client ID and secret 
(if the other service uses oAuth as an authentication flow).

The easiest way to implement an on/off switch is to check whether your migration service is enabled or not when 
registering the routes, and then simply don't registering the routes in case it is disabled.

File based migrators can always be enabled.

### Making the migrator public in `/info` 

You should make your migrator available in the `/info` endpoint so that frontends can display options to enable them or not.
To do this, add an entry to the `AvailableMigrators` field in `pkg/routes/api/v1/info.go`.
