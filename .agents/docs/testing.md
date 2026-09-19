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
- E2E tests assert stored state, not just a toast: reload the page, and check the API through the `apiContext`/`userToken` fixtures where the UI could lie.
- Component tests are only for what e2e can't exercise reliably: request races, stale responses after navigation, identity changes mid-request, and similar timing cases. Mocked component tests pass while the real page is broken, so they don't replace an e2e test for the same behaviour.
- Same for cache modules: a unit test covers what e2e can't see (`dataUpdateCount` untouched, no request made, `total`/`count` bookkeeping). "Edit in the detail shows on the list and the board" is an e2e test, not a `QueryClient` test.
- E2E tests that assert requests after navigating away and back must `page.reload()` first: the in-memory query cache serves the return navigation within `staleTime`, so "exactly one request" passes without proving anything.
- Drag/position tests need asymmetric fixture positions (100/250/300). With 100/200/300 the computed midpoint equals the dragged item's own position and the assertion cannot fail.
- A `@/client/generated` mock must not adapt the SDK into a legacy shape (a `getAll(scope, {...query, s: query.q}, page)` shim). Assertions then read fields the mock invented; assert the `path`/`query` passed to the SDK function.
- Unit tests: `pnpm vitest run <file>` in `frontend/`. Mock the generated client with `vi.mock('@/client/generated', () => sdk)` and `@/message` when the code toasts.
- When a component test is justified and reads server data, mount it against a real `QueryClient` seeded through the key factories, and mock only the generated client. Don't mock the composable that owns the behaviour under test; a mocked read hid a table that never updated after mutations. If a mutation invalidates a query, the client mock must answer the refetch with data matching the seeded state.
- A regression test must fail against the unfixed code. Check that before landing the fix.
- A test that can't fail is a defect. Examples: asserting that an unseeded cache is `toBeUndefined()`, checking for hidden controls on the viewer's own row (it never shows them), asserting two key literals differ, or calling the fix itself (`observer.reset()`) instead of going through the component.
- Typecheck with `pnpm typecheck` in `frontend/` and read the log. It has well over a thousand pre-existing errors, so compare the normalized error set against the base branch rather than the count; no new entries allowed:

  ```bash
  pnpm typecheck 2>&1 | grep 'error TS' | sed -E 's/\([0-9]+,[0-9]+\)//' | sort -u > /tmp/tsc-after.txt
  comm -13 /tmp/tsc-before.txt /tmp/tsc-after.txt
  ```

  Plain `vue-tsc --build` without `--force` can skip the build and print nothing, which looks like zero errors. Do not use `vue-tsc -p tsconfig.app.json`: it reports a spurious TS2589 on `i18n.global.t` that the project build does not, and it skips test files, so errors there (`.at()` is not in the lib target) go unnoticed until CI.
