# Testing

## Backend

- Always via mage: `mage test:web`, `mage test:feature`, `mage test:filter <go-test-filter>`. Plain `go test` does not work.
- `mage test:filter` runs most packages with `-short` but re-runs `pkg/webtests` without it, so a filter naming a web test actually executes it.
- Save output to a file and read the file: `mage test:filter Foo 2>&1 | tee /tmp/out.log`. Tests are expensive; never re-run one just to grep differently.
- Fixtures live in `pkg/db/fixtures/`. Use them instead of inventing data.
- Every permission path needs a positive and a negative test: the allowed user succeeds, an unrelated user is denied.

## Frontend

- E2E: invoke the `run-e2e-tests` skill (`mage test:e2e`). Never run `pnpm test:e2e` directly.
- Prefer e2e tests over component tests. User-visible behaviour (a created item appears, an edit sticks after reload, a delete removes the row) and the "How to verify" steps of a PR belong in `frontend/tests/e2e/`. Extend an existing spec when one covers the page; add one when none does.
- Component tests are only for what e2e can't exercise reliably: request races, stale responses after navigation, identity changes mid-request, and similar timing cases. Mocked component tests pass while the real page is broken, so they don't replace an e2e test for the same behaviour.
- Unit tests: `pnpm vitest run <file>` in `frontend/`. Mock the generated client with `vi.mock('@/client/generated', () => sdk)` and `@/message` when the code toasts.
- When a component test is justified and reads server data, mount it against a real `QueryClient` seeded through the key factories, and mock only the generated client. Don't mock the composable that owns the behaviour under test; a mocked read hid a table that never updated after mutations. If a mutation invalidates a query, the client mock must answer the refetch with data matching the seeded state.
- A regression test must fail against the unfixed code. Check that before landing the fix.
- Typecheck with `pnpm typecheck` in `frontend/` and read the log. It has well over a thousand pre-existing errors, so compare the normalized error set against the base branch rather than the count; no new entries allowed:

  ```bash
  pnpm typecheck 2>&1 | grep 'error TS' | sed -E 's/\([0-9]+,[0-9]+\)//' | sort -u > /tmp/tsc-after.txt
  comm -13 /tmp/tsc-before.txt /tmp/tsc-after.txt
  ```

  Plain `vue-tsc --build` without `--force` can skip the build and print nothing, which looks like zero errors. Do not use `vue-tsc -p tsconfig.app.json`: it reports a spurious TS2589 on `i18n.global.t` that the project build does not.
