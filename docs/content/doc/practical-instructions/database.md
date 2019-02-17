---
date: "2019-02-12:00:00+02:00"
title: "Database"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "practical instructions"
---

# Database

Vikunja uses [xorm](http://xorm.io/) as an abstraction layer to handle the database connection.
Please refer to [their](http://xorm.io/docs/) documentation on how to exactly use it.

Inside the `models` package, a variable `x` is available which contains a pointer to an instance of `xorm.Engine`.
This is used whenever you make a call to the database to get or update data.

This xorm instance is set up and initialized every time vikunja is started.

### Adding new database tables

To add a new table to the database, add a an instance of your struct to the `tables` variable in the 
init function in `pkg/models/models.go`. Xorm will sync them automatically.

You also need to add a pointer to the `tablesWithPointer` slice to enable caching for all instances of this struct.

To learn more about how to configure your struct to create "good" tables, refer to [the xorm documentaion](http://xorm.io/docs/).

### Adding data to test fixtures

Adding data for test fixtures is done in via `yaml` files insinde of `pkg/models/fixtures`.

The name of the yaml file should equal the table name in the database.
Adding values to it is done via array definition inside of the yaml file.

**Note**: Table and column names need to be in snake_case as that's what is used internally in the database 
and for mapping values from the database to xorm so your structs can use it.