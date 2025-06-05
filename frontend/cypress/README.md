# Frontend Testing With Cypress

## Setup

* Enable the [seeder api endpoint](https://vikunja.io/docs/config-options/#testingtoken). You'll then need to add the testingtoken in `cypress.json` or set the `CYPRESS_TEST_SECRET` environment variable.
* Basic configuration happens in the `cypress.json` file
* Overridable with [env](https://docs.cypress.io/guides/guides/environment-variables.html#Option-3-CYPRESS)
* Override base url with `CYPRESS_BASE_URL`

## Fixtures

We're using the [test endpoint](https://vikunja.io/docs/config-options/#testingtoken) of the vikunja api to
seed the database with test data before running the tests.
This ensures better reproducibility of tests.

## Running The Tests Locally

### Using Docker

The easiest way to run all frontend tests locally is by using the `docker-compose` file in this repository.
It uses the same configuration as the CI.

To use it, run

```shell
docker-compose up -d
```

Then, once all containers are started, run

```shell
docker-compose run cypress bash
```

to get a shell inside the cypress container.
In that shell you can then execute the tests with

```shell
pnpm run test:e2e
```

### Using The Cypress Dashboard

To open the Cypress Dashboard and run tests from there, run

```shell
pnpm run test:e2e:dev
```
