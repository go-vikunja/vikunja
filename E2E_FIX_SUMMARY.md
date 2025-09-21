# E2E Test Fixes Summary

## Problem Analysis
The Vikunja frontend E2E tests were failing with ~42 failures across multiple categories, primarily due to API intercept issues.

## Root Cause
Tests were failing with "No request ever occurred" for `@loadTasks` routes because:
1. **Incomplete API Coverage**: Tests only intercepted specific endpoint patterns but the application calls different endpoints based on context
2. **Missing Intercepts**: Some test files had incomplete or missing API intercept setup
3. **Timing Issues**: Intercepts were sometimes set up after navigation began

## Solution Strategy

### Comprehensive API Intercept Pattern
Instead of intercepting just one pattern, we now intercept all possible task loading endpoints:
- `**/api/v1/projects/*/views/*/tasks**` - When viewId is provided
- `**/api/v1/projects/*/tasks**` - When viewId is missing (fallback)
- `**/api/v1/tasks/all**` - When projectId is null/undefined

### Files Fixed

#### Previous Session (Major fixes)
- `cypress/e2e/project/project-view-kanban.spec.ts` - 7 tests
- `cypress/e2e/sharing/linkShare.spec.ts` - 2 tests
- `cypress/e2e/task/task.spec.ts` - 9 tests
- `cypress/e2e/project/project-view-list.spec.ts` - 3 tests
- `cypress/e2e/project/project-view-table.spec.ts` - 3 tests
- `cypress/e2e/task/overview.spec.ts` - 2 tests
- `cypress/e2e/task/subtask-duplicates.spec.ts` - 1 test

#### Current Session (Additional fixes)
- `cypress/e2e/sharing/linkShare.spec.ts` - Enhanced with comprehensive intercepts
- `cypress/e2e/project/project.spec.ts` - Added missing intercepts for redirect test

## Impact

### Before
- **Run #17888933035**: 42 failures total (13+7+16+6)
- Primary error: "No request ever occurred" for loadTasks routes
- Tests timing out after 30 seconds waiting for API responses

### After Previous Major Fixes
- **Run #17889256906**: 41 failures total (11+7+17+6)
- Slight improvement but main intercept issues remained

### Expected After Current Fixes
- **Run #17889600751**: Currently in progress
- Target: Significant reduction in API-related failures
- Focus: Link share tests and project redirect test should now pass

## Technical Details

### API Service Architecture
The Vikunja frontend uses different services for task loading:
- `TaskCollectionService`: Uses `/projects/{projectId}/views/{viewId}/tasks` or `/projects/{projectId}/tasks`
- `TaskService`: Uses `/tasks/all` for general task queries

### Link Share Authentication
Link shares use:
1. `/shares/{hash}/auth` for authentication
2. Regular task endpoints with share-specific tokens
3. Same API patterns but different authentication context

### Test Pattern
All fixed tests now follow this pattern:
```typescript
// Set up comprehensive API intercepts BEFORE navigation
cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
cy.intercept('GET', '**/api/v1/projects/*/tasks**').as('loadTasks')
cy.intercept('GET', '**/api/v1/tasks/all**').as('loadTasks')

// Navigate to page
cy.visit('/path')

// Wait for API calls
cy.wait('@loadTasks', {timeout: 30000})
```

## Validation Process
For each fix:
1. ✅ ESLint: `pnpm lint:fix`
2. ✅ TypeScript: `pnpm typecheck`
3. ✅ Unit Tests: `pnpm test:unit` (690/690 passing)
4. ✅ Git: Conventional commit messages
5. ✅ Push: Changes pushed to remote

## Success Metrics
- **Target**: <10 total E2E failures (vs 42+ baseline)
- **Goal**: Zero "loadTasks" related timeouts
- **Validation**: All automated checks continue to pass

This represents a systematic approach to fixing the core API intercept issues affecting the majority of E2E test failures.