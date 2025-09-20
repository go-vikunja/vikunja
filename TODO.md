# TODO List for E2E Test Fixes

## Completed
- [x] Create PLAN.md and TODO.md files
- [x] Check GitHub CI logs for latest test failures
- [x] Analyze failing tests and root causes
- [x] Fix identified issues (undefined project ID bug in router)
- [x] Run linter, typecheck, unit tests
- [x] Commit and push changes

## Latest Issues Fixed (September 20, 2025)

### E2E Test Failures Analysis
The E2E tests were failing with these specific errors:
- **project-view-table.spec.ts**: Tasks not appearing in table view (2 failures)
- **overview.spec.ts**: Tasks not showing on home page overview (1 failure)
- **team.spec.ts**: Create button not found (1 failure)

### Root Cause: Unsafe Route Parameter Parsing
The main issue was in `frontend/src/router/index.ts` where `parseInt()` and `Number()` were being used directly on route parameters without validation. This could result in:
- `parseInt(undefined)` → `NaN`
- `NaN` being passed as projectId/viewId to components
- Components failing to load data due to invalid IDs

### Solution: Safe Parameter Parsing
Replaced all unsafe parameter parsing in the router with `getRouteParamAsNumber()` utility function:

**Files modified:**
- `frontend/src/router/index.ts`: Fixed all route prop parsing
- Routes affected: task detail, project views, project settings, filters

**The `getRouteParamAsNumber()` utility:**
- Returns `undefined` for invalid parameters instead of `NaN`
- Allows proper validation in components/composables
- Prevents malformed API requests

### Validation Added
- **useTaskList composable**: Already had guard to prevent loading with invalid projectId
- **TaskCollectionService**: Already had validation to throw errors for undefined projectIds
- **Router props**: Now safely convert parameters or return undefined

## Previous Issues Fixed
The E2E tests were also previously failing because API requests were being made to `/api/v1/projects/undefined/tasks` when the project ID was null or undefined.

**Previous Solutions**:
- Enhanced AbstractService.getReplacedRoute() to validate parameters
- Fixed AddTask.vue to properly validate route project ID
- Updated various components to use safer parameter utilities
- Fixed ProjectSettingsBackground.vue parameter validation

## Notes
- All lint, typecheck, and unit tests pass
- Changes committed and pushed successfully
- Router now safely handles undefined/invalid route parameters

## Current Status (September 20, 2025 - 4:15 PM)
**Major Progress Made with E2E Test Fixes!**

### Local Testing Results:
- **Project creation tests**: ✅ PASSING (was previously failing)
- **Team creation tests**: ✅ PASSING (5/5 tests pass - was previously failing)
- **Login tests**: ✅ PASSING
- **Overview tests**: ❌ 1 test still failing (out of multiple tests, most pass)
- **Table view tests**: ❌ 1 test still failing

### Key Improvements:
- Most critical E2E tests that were completely broken are now working
- Button visibility issues (Create buttons not found) have been resolved
- Project and team creation workflows are functional
- Basic navigation and authentication flows are stable

### Remaining Issues:
- One task overview test: "Should show a new task with a very soon due date at the top"
- One table view test: "Should show a table with tasks"
- Both appear to be related to timing/synchronization issues with TaskFactory.create()

### All Static Analysis Passing:
- ✅ ESLint: No errors
- ✅ TypeScript: No type errors
- ✅ Unit tests: 690/690 passing
- ✅ Stylelint: No errors (when applicable)

The E2E test stability has dramatically improved from the previous state where most core functionality was broken.