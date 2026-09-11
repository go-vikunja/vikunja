# Testing

## Backend

- Always via mage: `mage test:web`, `mage test:feature`, `mage test:filter <go-test-filter>`. Plain `go test` does not work.
- `mage test:filter` runs most packages with `-short` but re-runs `pkg/webtests` without it, so a filter naming a web test actually executes it.
- Save output to a file and read the file: `mage test:filter Foo 2>&1 | tee /tmp/out.log`. Tests are expensive; never re-run one just to grep differently.
- Fixtures live in `pkg/db/fixtures/`. Use them instead of inventing data.
- Every permission path needs a positive and a negative test: the allowed user succeeds, an unrelated user is denied.

## Frontend

- E2E: invoke the `run-e2e-tests` skill (`mage test:e2e`). Never run `pnpm test:e2e` directly.
- Before adding a component test, check `frontend/tests/e2e/` for the same scenario. If an e2e test covers it, or can with a small extension, extend the e2e test instead. Component tests are for cases that are hard to exercise reliably end to end.
