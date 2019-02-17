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