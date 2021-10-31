---
date: "2019-02-12:00:00+02:00"
title: "Database"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Database

Vikunja uses [xorm](https://xorm.io/) as an abstraction layer to handle the database connection.
Please refer to [their](https://xorm.io/docs/) documentation on how to exactly use it.

{{< table_of_contents >}}

## Using the database

When using the common web handlers, you get an `xorm.Session` to do database manipulations.
In other packages, use the `db.NewSession()` method to get a new database session.

## Adding new database tables

To add a new table to the database, create the struct and [add a migration for it]({{< ref "db-migrations.md" >}}).

To learn more about how to configure your struct to create "good" tables, refer to [the xorm documentaion](https://xorm.io/docs/).

In most cases you will also need to implement the `TableName() string` method on the new struct to make sure the table 
name matches the rest of the tables - plural.

## Adding data to test fixtures

Adding data for test fixtures can be done via `yaml` files in `pkg/models/fixtures`.

The name of the yaml file should match the table name in the database.
Adding values to it is done via array definition inside it.

**Note**: Table and column names need to be in snake_case as that's what is used internally in the database 
and for mapping values from the database to xorm so your structs can use it.
