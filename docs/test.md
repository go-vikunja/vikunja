# Testing

You can run unit tests with our `Makefile` with

```bash
make test
```

### Running tests with config

You can run tests with all available config variables if you want, enabeling you to run tests for a lot of scenarios.

To use the normal config set the enviroment variable `VIKUNJA_TESTS_USE_CONFIG=1`.

### Show sql queries

When `UNIT_TESTS_VERBOSE=1` is set, all sql queries will be shown when tests are run.