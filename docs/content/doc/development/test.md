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

{{< table_of_contents >}}

## API Tests

The following parts are about the kinds of tests in the API package and how to run them.

### Prerequesites

To run any kind of test, you need to specify Vikunja's [root path](https://vikunja.io/docs/config-options/#rootpath).
This is required to make sure all test fixtures are correctly loaded.

The easies way to do that is to set the environment variable `VIKUNJA_SERVICE_ROOTPATH` to the path where you cloned the working directory.

### Unit tests

To run unit tests with [mage]({{< ref "mage.md">}}), execute

{{< highlight bash >}}
mage test:unit
{{< /highlight >}}

In Vikunja, everything that is not an integration test counts as unit test - even if it accesses the db.
This definition is a bit blurry, but we haven't found a better one yet.

### Integration tests

All integration tests live in `pkg/integrations`.
You can run them by executing `mage test:integration`.

The integration tests use the same config and fixtures as the unit tests and therefor have the same options available,
see at the beginning of this document.

To run integration tests, use `mage test:integration`.

### Running tests with config

You can run tests with all available config variables if you want, enabeling you to run tests for a lot of scenarios.
We use this in CI to run all tests with different databases.

To use the normal config set the enviroment variable `VIKUNJA_TESTS_USE_CONFIG=1`.

### Showing sql queries

When the environment variable `UNIT_TESTS_VERBOSE=1` is set, all sql queries will be shown during the test run.

### Fixtures

All tests are run against a set of db fixtures.
These fixtures are defined in `pkg/models/fixtures` in YAML-Files which represent the database structure.

When you add a new test case which requires new database entries to test against, update these files.

#### Initializing db fixtures when writing tests

All db fixtures for all tests live in the `pkg/db/fixtures/` folder as yaml files.
Each file has the same name as the table the fixtures are for.
You should put new fixtures in this folder.

When initializing db fixtures, you are responsible for defining which tables your package needs in your test init function.
Usually, this is done as follows (this code snippet is taken from the `user` package):

{{< highlight go >}}
err = db.InitTestFixtures("users")
if err != nil {
	log.Fatal(err)
}
{{< /highlight >}}

In your actual tests, you then load the fixtures into the in-memory db like so:

{{< highlight go >}}
db.LoadAndAssertFixtures(t)
{{< /highlight >}}

This will load all fixtures you defined in your test init method.
You should always use this method to load fixtures, the only exception is when your package tests require extra test 
fixtures other than db fixtures (like files).

## Frontend tests

The frontend has end to end tests with Cypress that use a Vikunja instance and drive a browser against it.
Check out the docs [in the frontend repo](https://kolaente.dev/vikunja/frontend/src/branch/main/cypress/README.md) about how they work and how to get them running.

### Unit Tests

To run the frontend unit tests, run

{{< highlight bash >}}
yarn test:unit
{{< /highlight >}}

The frontend also has a watcher available that re-runs all unit tests every time you change something.
To use it, simply run

{{< highlight bash >}}
yarn test:unit-watch
{{< /highlight >}}
