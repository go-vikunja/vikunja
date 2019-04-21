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

You can run unit tests with [our `Makefile`]({{< ref "make.md">}}) with

{{< highlight bash >}}
make test
{{< /highlight >}}

### Running tests with config

You can run tests with all available config variables if you want, enabeling you to run tests for a lot of scenarios.

To use the normal config set the enviroment variable `VIKUNJA_TESTS_USE_CONFIG=1`.

### Show sql queries

When `UNIT_TESTS_VERBOSE=1` is set, all sql queries will be shown when tests are run.

### Fixtures

All tests are run against a set of db fixtures.
These fixtures are defined in `pkg/models/fixtures` in YAML-Files which represent the database structure.

When you add a new test case which requires new database entries to test against, update these files.

# Integration tests

All integration tests live in `pkg/integrations`.
You can run them by executing `make integration-test`.

The integration tests use the same config and fixtures as the unit tests and therefor have the same options available,
see at the beginning of this document.

To run integration tests, use `make integration-test`.
