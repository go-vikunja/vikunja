# TODO List for E2E Test Fixes

## Completed
- [x] Create PLAN.md and TODO.md files
- [x] Check GitHub CI logs for latest test failures
- [x] Analyze failing tests and root causes
- [x] Fix identified issues (undefined project ID bug)
- [x] Run linter, typecheck, unit tests
- [x] Commit and push changes

## Issue Fixed
The E2E tests were failing because API requests were being made to `/api/v1/projects/undefined/tasks` when the project ID was null or undefined. This caused 400 errors from the backend.

**Root Cause**: Multiple locations were using `Number(route.params.projectId)` which converts `undefined` to `NaN`, and `NaN.toString()` becomes "undefined" in API URLs.

**Solution**:
- Enhanced AbstractService.getReplacedRoute() to validate all route parameters and throw errors for invalid values (undefined, null, NaN, "undefined", "null")
- Fixed AddTask.vue to properly validate route project ID with NaN check before using it
- Updated ContentLinkShare.vue and Filters.vue to use the safer `getRouteParamAsNumber()` utility
- Fixed ProjectSettingsBackground.vue to validate project ID before making API calls
- These changes prevent malformed API URLs and provide better error messages when parameters are invalid

## Notes
- All lint, typecheck, and unit tests pass
- Changes committed and pushed successfully
- Ready for next E2E test run to verify fix