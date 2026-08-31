---
name: run-e2e-tests
description: Run Vikunja's Playwright end-to-end tests through `mage test:e2e`. Use when asked to run e2e tests, reproduce a bug in the browser, or verify a frontend change end to end.
---

# Running E2E Tests

**ALWAYS use `mage test:e2e`.** Do NOT run `pnpm test:e2e` directly. The mage command builds the API, starts it with an isolated SQLite database, builds and serves the frontend, runs the Playwright tests, and tears everything down automatically.

```bash
mage test:e2e ""                                      # run all tests
mage test:e2e "tests/e2e/misc/menu.spec.ts"           # specific file
mage test:e2e "--grep menu"                            # filter by name
mage test:e2e "--headed tests/e2e/misc/menu.spec.ts"  # headed mode
```

**Always save test output to a file.** E2E tests are expensive (they rebuild the API, start servers, run browsers). NEVER re-run tests just to look at the output differently (e.g., with different `grep`/`tail` filters). Save the output on the first run and then read the file:

```bash
# First run: save output to a file
mage test:e2e "tests/e2e/misc/menu.spec.ts" 2>&1 | tee /tmp/e2e-output.log

# Subsequent analysis: read the file, don't re-run
cat /tmp/e2e-output.log | grep -E '(passed|failed)'
cat /tmp/e2e-output.log | tail -20
```

Set `VIKUNJA_E2E_SKIP_BUILD=true` to skip rebuilding the API binary when iterating on frontend-only changes.
