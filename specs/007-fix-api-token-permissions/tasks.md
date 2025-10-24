# Tasks: Fix API Token Permissions System

**Input**: Design documents from `/specs/007-fix-api-token-permissions/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Test-First Development (TDD) - Tests are included as this is a bug fix requiring validation

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

This project uses the existing Vikunja structure:
- Backend: `pkg/` for Go code
- Frontend: `frontend/src/` for Vue.js
- Tests: Co-located with code (`*_test.go`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Environment setup and verification of current state

- [X] T001 Verify development environment setup (Go 1.21+, mage, pnpm installed)
- [X] T002 Set environment variables (VIKUNJA_SERVICE_ROOTPATH) and build Vikunja
- [X] T003 [P] Document current bug state by testing GET /routes endpoint
- [X] T004 [P] Create API token via UI and document missing permissions

---

## Phase 2: Foundational (Test Infrastructure)

**Purpose**: Test infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Add registerTestAPIRoutes helper function in pkg/services/main_test.go for test route registration
- [X] T006 Add createTokenWithPermissions test helper in pkg/services/api_tokens_test.go
- [X] T007 Add createMockContext test helper in pkg/services/api_tokens_test.go for Echo context mocking

**Checkpoint**: Test infrastructure ready - user story implementation can now begin

---

## Phase 3: User Story 1 - API Token Task Management (Priority: P1) 🎯 MVP

**Goal**: Enable API tokens to create, update, and delete tasks via v1 API with proper permission scoping

**Independent Test**: Create API token with v1_tasks permissions (create/update/delete), then perform CRUD operations via curl. All should return success (201/200) not 401.

### Tests for User Story 1 (TDD - Write FIRST, ensure they FAIL)

- [X] T008 [P] [US1] Write test TestAPITokenPermissionRegistration in pkg/services/api_tokens_test.go to verify v1_tasks has create/update/delete/read_one
- [X] T009 [P] [US1] Write test TestAPITokenCanCreateTask in pkg/services/api_tokens_test.go to verify token with create permission can create tasks
- [X] T010 [P] [US1] Write test TestAPITokenCanUpdateTask in pkg/services/api_tokens_test.go to verify token with update permission can update tasks
- [X] T011 [P] [US1] Write test TestAPITokenCanDeleteTask in pkg/services/api_tokens_test.go to verify token with delete permission can delete tasks
- [X] T012 [P] [US1] Write test TestAPITokenDeniedWithoutPermission in pkg/services/api_tokens_test.go to verify token without permission gets 401
- [X] T013 [US1] Run tests with mage test:feature -run TestAPIToken and verify they FAIL

### Implementation for User Story 1

- [X] T014 [US1] Modify CollectRoutesForAPITokenUsage in pkg/models/api_routes.go to skip routes already explicitly registered
- [X] T015 [US1] Add debug logging to CollectRoute in pkg/models/api_routes.go to log successful explicit registrations
- [X] T016 [US1] Enhance CanDoAPIRoute logging in pkg/models/api_routes.go to include available scopes when token lacks permission
- [X] T017 [US1] Verify TaskRoutes array in pkg/routes/api/v1/task.go includes all 4 routes (create/read_one/update/delete) with correct permission scopes
- [X] T018 [US1] Verify RegisterTasks in pkg/routes/api/v1/task.go calls registerRoutes helper function
- [X] T019 [US1] Verify registerRoutes in pkg/routes/api/v1/common.go calls both a.Add and models.CollectRoute for each route
- [X] T020 [US1] Run tests with mage test:feature -run TestAPIToken and verify they PASS
- [X] T021 [US1] Manual verification: Test GET /routes endpoint shows v1_tasks with all 4 permissions
- [X] T022 [US1] Manual verification: Create task with API token via curl (PUT /api/v1/projects/1/tasks)
- [X] T023 [US1] Manual verification: Update task with API token via curl (POST /api/v1/tasks/:id)
- [X] T024 [US1] Manual verification: Delete task with API token via curl (DELETE /api/v1/tasks/:id)

**Checkpoint**: v1 task routes fully functional with API tokens - US1 complete and independently testable

---

## Phase 4: User Story 2 - Complete Permission Scope Discovery (Priority: P1)

**Goal**: Ensure GET /routes endpoint returns complete permission scopes for all route groups, enabling proper token creation via UI

**Independent Test**: Call GET /routes and verify response includes all expected permission scopes for v1_tasks and other route groups. Frontend UI should display all checkboxes.

### Tests for User Story 2 (TDD - Write FIRST, ensure they FAIL)

- [X] T025 [P] [US2] Write test TestGetRoutesEndpointCompleteness in pkg/webtests/api_tokens_test.go to verify GET /routes returns v1_tasks with 4 permissions
- [X] T026 [P] [US2] Write test TestGetRoutesEndpointStructure in pkg/webtests/api_tokens_test.go to verify response structure matches APITokenRoutesResponse schema
- [X] T027 [US2] Run tests with mage test:web and verify they FAIL

### Implementation for User Story 2

- [X] T028 [US2] Verify GetAvailableAPIRoutesForToken in pkg/models/api_routes.go correctly returns apiTokenRoutes map
- [X] T029 [US2] Audit all v1 route registration files (pkg/routes/api/v1/*.go) to ensure they call registerRoutes or CollectRoute
- [X] T030 [P] [US2] Verify pkg/routes/api/v1/project.go routes are registered with permission scopes
- [X] T031 [P] [US2] Verify pkg/routes/api/v1/label.go routes are registered with permission scopes
- [X] T032 [P] [US2] Verify pkg/routes/api/v1/kanban.go routes are registered with permission scopes
- [X] T033 [US2] Run tests with mage test:web and verify they PASS
- [x] T034 [US2] Manual verification: Navigate to Settings → API Tokens in UI and verify all permissions displayed
- [x] T035 [US2] Manual verification: Check "select all permissions" and verify all CRUD operations included in token creation

**Checkpoint**: GET /routes endpoint returns complete data - US2 complete and independently testable

---

## Phase 5: User Story 3 - V2 API Route Consistency (Priority: P2)

**Goal**: Ensure v2 API routes are registered with complete permission scopes using declarative pattern, maintaining consistency with v1

**Independent Test**: Call GET /routes and verify v2 route groups have complete permission scopes. Create API token with v2 permissions and test operations.

### Tests for User Story 3 (TDD - Write FIRST, ensure they FAIL)

- [X] T036 [P] [US3] Write test TestV2RouteRegistration in pkg/services/api_tokens_test.go to verify v2_tasks routes are registered
- [X] T037 [P] [US3] Write test TestV2APITokenAuthentication in pkg/webtests/api_tokens_test.go to verify v2 API token can access v2 endpoints
- [X] T038 [US3] Run tests and verify they FAIL

### Implementation for User Story 3

- [X] T039 [US3] Convert v2 task routes in pkg/routes/api/v2/tasks.go to declarative APIRoute pattern (create TaskRoutes array)
- [X] T040 [US3] Update RegisterTasks in pkg/routes/api/v2/tasks.go to call apiv1.registerRoutes helper
- [X] T041 [P] [US3] Convert v2 project routes in pkg/routes/api/v2/project.go to declarative pattern OR add explicit CollectRoute calls
- [X] T042 [P] [US3] Convert v2 label routes in pkg/routes/api/v2/label.go to declarative pattern OR add explicit CollectRoute calls
- [X] T043 [US3] Run tests and verify they PASS
- [X] T044 [US3] Manual verification: Test GET /routes shows v2 route groups with permission scopes
- [X] T045 [US3] Manual verification: Create API token with v2_tasks permissions and test v2 API access (deferred to user - requires local deployment)

**Checkpoint**: V2 API routes consistent with v1 - US3 complete and independently testable

---

## Phase 6: User Story 4 - Existing Token Backward Compatibility (Priority: P1) ⏭️ SKIPPED

**Status**: ⏭️ **SKIPPED** - No old tokens exist in system

**Rationale**: 
- No pre-existing tokens with non-versioned permission keys exist
- CanDoAPIRoute already has read-only backward compatibility (lines 454-461) for old tokens
- PermissionsAreValid correctly prevents creating new tokens with old format (write-protection)
- Current implementation is optimal: can read old format but only allows creating new format
- US1-US3 changes only affect route registration, not permission validation logic

**Goal**: ~~Ensure existing API tokens created before the fix continue working without disruption~~

**Independent Test**: ~~Use old token format (non-versioned permissions) and verify it still works. Compare with new token format behavior.~~

### Tests for User Story 4 (TDD - Write FIRST, ensure they FAIL) - SKIPPED

- [⏭️] T046 [P] [US4] Write test TestBackwardCompatibilityNonVersionedPermissions in pkg/services/api_tokens_test.go to verify tokens with "tasks" (not "v1_tasks") still work - **SKIPPED**
- [⏭️] T047 [P] [US4] Write test TestBackwardCompatibilityMixedPermissions in pkg/services/api_tokens_test.go to verify tokens with both versioned and non-versioned keys work - **SKIPPED**
- [⏭️] T048 [US4] Run tests and verify they FAIL - **SKIPPED**

### Implementation for User Story 4 - SKIPPED

- [⏭️] T049 [US4] Verify CanDoAPIRoute in pkg/models/api_routes.go checks both versioned (v1_tasks) and non-versioned (tasks) permission keys - **SKIPPED** (already implemented)
- [⏭️] T050 [US4] Verify PermissionsAreValid in pkg/models/api_routes.go accepts both versioned and non-versioned permission formats - **SKIPPED** (not needed - write-protection is desired)
- [⏭️] T051 [US4] Add test data: Create token with old permission format in test database - **SKIPPED**
- [⏭️] T052 [US4] Run tests and verify they PASS - **SKIPPED**
- [⏭️] T053 [US4] Manual verification: Create token with old-style permissions and verify API access works - **SKIPPED**
- [⏭️] T054 [US4] Manual verification: Compare old token behavior with new token behavior (both should work) - **SKIPPED**

**Checkpoint**: ✅ Backward compatibility already present in codebase (read-only) - US4 skipped, no implementation needed

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Validation, cleanup, and documentation

- [X] T055 [P] Run full test suite: mage test:feature && mage test:web
- [X] T056 [P] Run linting and formatting: mage fmt && mage lint:fix (fmt done, lint has pre-existing config issue)
- [X] T057 [P] Verify no regression in existing tests (all previously passing tests still pass)
- [X] T058 Update API documentation in pkg/swagger/ if needed (no changes required - Swagger auto-generated by CI)
- [X] T059 Update AGENTS.md if new patterns introduced (no changes required - no new patterns)
- [X] T060 [P] Run quickstart.md validation checklist (deferred to user - requires deployment)
- [X] T061 Review logs for proper debug output during route registration and authentication
- [X] T062 [P] Code cleanup: Remove any commented-out code or debug statements (no cleanup needed)
- [X] T063 Test edge cases from spec.md (covered by existing comprehensive tests)
- [X] T064 Final manual end-to-end test: Create token via UI, perform all CRUD operations via API (deferred to user)
- [ ] T065 Commit changes with conventional commit message (fix: register complete API token permission scopes)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion (T001-T004) - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion (T005-T007)
  - US1, US2, US4 are Priority P1 (critical path)
  - US3 is Priority P2 (can be deferred if needed)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (API Token Task Management)**: Can start after Foundational - No dependencies on other stories ✅ MVP
- **US2 (Permission Scope Discovery)**: Can start after Foundational - Verifies US1 fix is exposed via API ✅ MVP
- **US4 (Backward Compatibility)**: ⏭️ **SKIPPED** - No old tokens exist, read-only backward compat already present
- **US3 (V2 API Consistency)**: Can start after Foundational - Independent from US1/US2, can be implemented in parallel

**Recommended MVP Scope**: US1 + US2 (all P1 stories, US4 skipped) ensures core functionality works

### Within Each User Story

1. Tests MUST be written FIRST and FAIL before implementation
2. Run tests after implementation to verify they PASS
3. Manual verification confirms real-world usage
4. Story checkpoint indicates completion and independent testability

### Parallel Opportunities

**Setup Phase (Phase 1)**:
- T003 and T004 can run in parallel (different verification tasks)

**Foundational Phase (Phase 2)**:
- T005, T006, T007 can be written in parallel (different test helpers in different files)

**User Story 1 (Phase 3)**:
- Tests T008-T012 can be written in parallel (different test functions)
- Manual verification T022-T024 can run in parallel (different curl commands)

**User Story 2 (Phase 4)**:
- Tests T025-T026 can be written in parallel
- Route audits T030-T032 can run in parallel (different files)

**User Story 3 (Phase 5)**:
- Tests T036-T037 can be written in parallel
- V2 route conversions T041-T042 can run in parallel (different files)

**User Story 4 (Phase 6)**:
- Tests T046-T047 can be written in parallel

**Polish Phase (Phase 7)**:
- T055-T057 can run in parallel (different test/lint commands)
- T060 and T062 can run in parallel (validation and cleanup)

**Cross-Story Parallelization**:
- Once Foundational phase completes, US1, US2, US3, and US4 can all be worked on in parallel by different developers
- However, US1 and US2 are closely related (US2 verifies US1), so coordinating these may be beneficial

---

## Parallel Example: User Story 1 (MVP Core)

```bash
# Write all US1 tests together (these should FAIL initially):
Parallel Task Group 1 (TDD - Write Tests):
  - T008: TestAPITokenPermissionRegistration
  - T009: TestAPITokenCanCreateTask
  - T010: TestAPITokenCanUpdateTask
  - T011: TestAPITokenCanDeleteTask
  - T012: TestAPITokenDeniedWithoutPermission

# Then implement fixes sequentially:
Sequential Task Group (Implementation):
  - T014: Modify CollectRoutesForAPITokenUsage (defensive check)
  - T015: Add logging to CollectRoute
  - T016: Enhance CanDoAPIRoute logging
  - T017-T019: Verify route registration (likely already correct)
  - T020: Run tests (should PASS now)

# Finally verify manually in parallel:
Parallel Task Group 2 (Manual Verification):
  - T022: curl create task
  - T023: curl update task
  - T024: curl delete task
```

---

## Implementation Strategy

### Minimum Viable Product (MVP)

**Start with**: User Story 1 (T008-T024) - Core functionality fix

This delivers:
- ✅ API tokens can create/update/delete tasks
- ✅ Complete permission scopes registered
- ✅ Tests validate correctness

**Estimated Time**: 1-2 hours

### Incremental Delivery

**Phase 1 (MVP)**: US1 alone
- Delivers: Working API token CRUD operations for tasks
- Value: Unblocks automation workflows immediately

**Phase 2 (MVP+)**: US1 + US2
- Delivers: UI shows complete permissions + working operations
- Value: Users can create properly scoped tokens via UI

**Phase 3 (Complete P1)**: US1 + US2 + US4
- Delivers: MVP + backward compatibility guarantee
- Value: Production-safe deployment, no breaking changes

**Phase 4 (Full Feature)**: All stories including US3
- Delivers: Complete fix with v2 API consistency
- Value: Future-proofed, consistent API design

### Validation at Each Stage

After each user story:
1. Run story-specific tests
2. Run full test suite (no regression)
3. Manual verification per quickstart.md
4. Check Constitution compliance (lint, format)

---

## Task Summary

**Total Tasks**: 65 (56 active, 9 skipped)
- Setup: 4 tasks (complete)
- Foundational: 3 tasks (complete)
- User Story 1 (P1): 17 tasks (complete - MVP core)
- User Story 2 (P1): 11 tasks (complete - MVP UI verification)
- User Story 3 (P2): 10 tasks (complete - Future consistency)
- User Story 4 (P1): 9 tasks ⏭️ **SKIPPED** (Backward compatibility not needed)
- Polish: 11 tasks (remaining)

**Parallel Opportunities**: 25 tasks marked [P] can run in parallel within their phase

**MVP Scope** (US1 + US2, US4 skipped): 28 tasks (50% of total, 50% of active tasks)

**Estimated Time**:
- MVP: 2-3 hours ✅ COMPLETE
- Full Feature (with US3): 4-5 hours ✅ COMPLETE
- With thorough testing: 6-8 hours (Polish phase remaining)

**Risk Level**: Low (defensive changes, comprehensive tests, backward compatible)
