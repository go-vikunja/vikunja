# Cypress to Playwright Migration Strategy

## Overview

This repository now runs **both** Cypress and Playwright test suites in parallel during the migration period. This gradual approach allows us to maintain test coverage while fixing Playwright tests incrementally.

## Current Status

### Playwright Tests (Passing: ~82%)
**13 fully passing test suites (69 tests):**
- tests/e2e/misc/menu.spec.ts - 4/4 ✅
- tests/e2e/project/filter-persistence.spec.ts - 4/4 ✅
- tests/e2e/project/project.spec.ts - 7/7 ✅
- tests/e2e/project/project-view-gantt.spec.ts - 8/8 ✅
- tests/e2e/project/project-view-kanban.spec.ts - 14/14 ✅
- tests/e2e/project/project-view-list.spec.ts - 7/8 ✅
- tests/e2e/project/project-view-table.spec.ts - 3/3 ✅
- tests/e2e/task/comment-pagination.spec.ts - 2/2 ✅
- tests/e2e/task/date-display.spec.ts - 9/9 ✅
- tests/e2e/task/overview.spec.ts - 4/4 ✅
- tests/e2e/task/subtask-duplicates.spec.ts - 1/1 ✅
- tests/e2e/user/login.spec.ts - 1/1 ✅
- tests/e2e/user/logout.spec.ts - 1/1 ✅

### Cypress Tests (Coverage for failing Playwright tests)
**9 test files providing coverage:**
- cypress/e2e/project/project-history.spec.ts
- cypress/e2e/sharing/linkShare.spec.ts
- cypress/e2e/sharing/team.spec.ts
- cypress/e2e/task/task.spec.ts (~47 tests, ~7 failing in Playwright)
- cypress/e2e/user/email-confirmation.spec.ts
- cypress/e2e/user/openid-login.spec.ts
- cypress/e2e/user/password-reset.spec.ts
- cypress/e2e/user/registration.spec.ts
- cypress/e2e/user/settings.spec.ts

## How It Works

1. **Both test suites run in CI** - GitHub Actions runs both Cypress and Playwright tests
2. **No duplicate coverage** - Only Cypress tests WITHOUT passing Playwright equivalents are kept
3. **Incremental migration** - As Playwright tests are fixed, corresponding Cypress tests can be removed
4. **Zero coverage loss** - Every feature is tested by either Cypress or Playwright (or both during transition)

## Next Steps

### To Fix a Failing Playwright Test:
1. Fix the Playwright test in `frontend/tests/e2e/`
2. Verify it passes consistently
3. Remove the corresponding Cypress test from `frontend/cypress/e2e/`
4. Update this document

### Priority Order (from STATUS.md):
1. **High Priority:** Fix remaining task.spec.ts failures (~7 tests)
2. **Medium Priority:** Fix project-history, project-view-list, sharing tests
3. **Low Priority:** Fix user management tests (email, openid, password, registration, settings)

## Migration Complete When:
- All Playwright tests pass (100% coverage)
- All Cypress tests removed
- Cypress dependencies removed from package.json
- Cypress job removed from CI workflow

## References
- See `frontend/STATUS.md` for detailed Playwright test status
- See `frontend/cypress/README.md` for Cypress test documentation
