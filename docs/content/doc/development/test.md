---
date: "2019-02-12:00:00+02:00"
title: "Testing"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Testing

You can run unit tests with [mage]({{< ref "mage.md">}}) with

{{< highlight bash >}}
mage test:unit
{{< /highlight >}}

{{< table_of_contents >}}

## Running tests with config

You can run tests with all available config variables if you want, enabeling you to run tests for a lot of scenarios.

To use the normal config set the enviroment variable `VIKUNJA_TESTS_USE_CONFIG=1`.

## Show sql queries

When `UNIT_TESTS_VERBOSE=1` is set, all sql queries will be shown when tests are run.

## Fixtures

All tests are run against a set of db fixtures.
These fixtures are defined in `pkg/models/fixtures` in YAML-Files which represent the database structure.

When you add a new test case which requires new database entries to test against, update these files.

## Integration tests

All integration tests live in `pkg/integrations`.
You can run them by executing `mage test:integration`.

The integration tests use the same config and fixtures as the unit tests and therefor have the same options available,
see at the beginning of this document.

To run integration tests, use `mage test:integration`.

## Initializing db fixtures when writing tests

All db fixtures for all tests live in the `pkg/db/fixtures/` folder as yaml files.
Each file has the same name as the table the fixtures are for.
You should put new fixtures in this folder.

When initializing db fixtures, you are responsible for defining which tables your package needs in your test init function.
Usually, this is done as follows (this code snippet is taken from the `user` package):

```go
err = db.InitTestFixtures("users")
if err != nil {
	log.Fatal(err)
}
```

In your actual tests, you then load the fixtures into the in-memory db like so:

```go
db.LoadAndAssertFixtures(t)
```

This will load all fixtures you defined in your test init method.
You should always use this method to load fixtures, the only exception is when your package tests require extra test 
fixtures other than db fixtures (like files).
