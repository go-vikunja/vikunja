# Tasks: HTTP Transport for MCP Server

**Feature**: HTTP Transport for Vikunja MCP Server  
**Branch**: `007-mcp-http-transport`  
**Input**: Design documents from `/specs/007-mcp-http-transport/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD approach - tests written FIRST before implementation per Constitution requirement

**Organization**: Tasks grouped by user story to enable independent implementation and testing

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions
- Single TypeScript project: `mcp-server/src/`, `mcp-server/tests/`
- Deployment: `deploy/proxmox/`
- Documentation: Root level files (README, CHANGELOG)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency setup

- [X] T001 Install new dependencies in mcp-server/package.json (ioredis, rate-limiter-flexible, uuid, express, @types/express)
- [X] T002 [P] Update mcp-server/.env.example with HTTP transport configuration (MCP_HTTP_ENABLED, MCP_HTTP_PORT, REDIS_URL, rate limiting config)
- [X] T003 [P] Update mcp-server/tsconfig.json to include new source directories (auth/, ratelimit/, transports/http/)
- [X] T004 [P] Create mcp-server/src/utils/errors.ts with custom error classes (AuthenticationError, RateLimitError, SessionError)
- [X] T005 [P] Update mcp-server/src/utils/logger.ts to add HTTP transport logging contexts

**Checkpoint**: Dependencies installed, configuration files ready, basic utilities available

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 Create mcp-server/src/config/schema.ts with Zod schemas for HTTP transport config validation
- [X] T007 Update mcp-server/src/config/index.ts to load and validate HTTP transport configuration
- [X] T008 [P] Create mcp-server/src/auth/token-validator.ts with TokenValidator class (validateToken method, Redis caching)
- [X] T009 [P] Create mcp-server/src/auth/middleware.ts with Express authentication middleware (authenticateBearer, authenticateQuery)
- [X] T010 [P] Create mcp-server/src/ratelimit/limiter.ts with RateLimiter class using rate-limiter-flexible
- [X] T011 [P] Create mcp-server/src/ratelimit/redis-store.ts with Redis backend configuration for rate limiting
- [X] T012 [P] Create mcp-server/src/transports/http/session-manager.ts with SessionManager class (createSession, getSession, updateActivity, cleanupStaleSessions)
- [X] T013 Update mcp-server/src/vikunja/client.ts to add connection pooling for HTTP transport
- [X] T014 Update mcp-server/src/index.ts to detect HTTP mode vs stdio mode (check MCP_HTTP_ENABLED env var)

**Checkpoint**: Foundation ready - authentication, rate limiting, session management, and configuration all operational

---

## Phase 3: User Story 1 - Remote Client Connection (Priority: P1) 🎯 MVP

**Goal**: Enable MCP clients to connect remotely via HTTP Streamable, authenticate with Vikunja tokens, and execute tools

**Independent Test**: Configure MCP client with server URL and token, establish connection, list available tools, call a tool (e.g., get_tasks)

### Tests for User Story 1 (TDD - Write FIRST, ensure FAIL)

- [X] T015 [P] [US1] Create mcp-server/tests/unit/transports/http-streamable.test.ts with HTTP Streamable connection tests (connection success, protocol compliance, NDJSON streaming)
- [X] T016 [P] [US1] Create mcp-server/tests/unit/auth/token-validator.test.ts with authentication tests (valid token, invalid token, expired token, cached token)
- [X] T017 [P] [US1] Create mcp-server/tests/integration/http-transport.test.ts with end-to-end tests (full connection flow, tool listing, tool execution)

### Implementation for User Story 1

- [X] T018 [US1] Create mcp-server/src/transports/http/http-streamable.ts with HTTPStreamableTransport class (@modelcontextprotocol/sdk integration, NDJSON streaming, message handling) ✅ **COMPLETE** - Shared MCP server architecture with session-specific contexts
- [X] T019 [US1] Create mcp-server/src/transports/http/health-check.ts with health check endpoint handler (check Redis, check Vikunja API, return session stats) ✅ **COMPLETE**
- [X] T020 [US1] Create mcp-server/src/transports/http/index.ts to export HTTP transport handlers and start Express server ✅ **COMPLETE** - Exports complete, server wiring in index.ts
- [X] T020b [US1] **CRITICAL**: Wire VikunjaMCPServer to HTTPStreamableTransport (create MCP Server instances, connect to transport, route messages to tool handlers) ✅ **COMPLETE** - Shared server connected via transport.connect()
- [X] T020c [US1] Implement MCP Server connection architecture (decide per-session vs shared server, implement connection lifecycle) ✅ **COMPLETE** - Using shared VikunjaMCPServer with session contexts
- [X] T021 [US1] Update mcp-server/src/index.ts to initialize HTTP server when MCP_HTTP_ENABLED=true (create Express app, setup routes, start listening) ✅ **COMPLETE** - POST /mcp endpoint wired with auth, rate limiting, and session management
- [X] T022 [US1] ~~Wire authentication middleware to HTTP Streamable endpoint POST /mcp (authenticateBearer)~~ ✅ **DONE** - Embedded in HTTPStreamableTransport.handleRequest
- [X] T023 [US1] ~~Wire rate limiting middleware to HTTP Streamable endpoint POST /mcp~~ ✅ **DONE** - Embedded in HTTPStreamableTransport.handleRequest
- [X] T024 [US1] ~~Add session creation and management to HTTP Streamable connection flow~~ ✅ **DONE** - Embedded in HTTPStreamableTransport.handleRequest
- [X] T025 [US1] Add error handling and logging for HTTP Streamable transport ✅ **COMPLETE** - Enhanced with detailed error logging, timing metrics, IP tracking, and transport lifecycle logging
- [X] T026 [US1] Verify tests pass: Run mcp-server/tests/transports/http-streamable.test.ts ✅ **PASS** - 16/16 tests passed (HTTP integration tests with supertest)
- [X] T026b [US1] Implement full HTTP integration tests with supertest ✅ **COMPLETE** - Comprehensive HTTP integration tests: protocol compliance (4 tests), authentication flow (4 tests), session management (3 tests), rate limiting (2 tests), error handling (3 tests)
- [X] T027 [US1] Verify tests pass: Run mcp-server/tests/auth/token-validator.test.ts ✅ **PASS** - 19/19 tests passed (fixed hanging test with ioredis mock + added required field validation)
- [X] T028 [US1] Verify tests pass: Run mcp-server/tests/integration/http-transport.test.ts ✅ **PASS** - 24/24 tests passed
- [ ] T028b [US1] Implement full end-to-end integration tests (rewrite http-transport.test.ts to test real server, tool execution, connection flow)

**Checkpoint**: HTTP Streamable transport fully functional - clients can connect, authenticate, list tools, execute tools

---

## Phase 4: User Story 2 - Modern Transport Protocol Support (Priority: P1)

**Goal**: Add SSE transport for backward compatibility while keeping HTTP Streamable as primary (with deprecation notices)

**Independent Test**: Configure MCP client to use SSE mode (if supported), verify connection, tool listing, and execution work identically to HTTP Streamable

### Tests for User Story 2 (TDD - Write FIRST, ensure FAIL)

- [ ] T029 [P] [US2] Create mcp-server/tests/transports/sse-transport.test.ts with SSE connection tests (GET /sse stream, POST /sse messages, session correlation, EventSource compliance)

### Implementation for User Story 2

- [ ] T030 [P] [US2] Create mcp-server/src/transports/http/sse-transport.ts with SSETransport class (GET /sse event stream handler, POST /sse message handler, session management)
- [ ] T031 [US2] Wire authentication middleware to SSE endpoints (authenticateQuery for GET /sse, authenticateQuery for POST /sse)
- [ ] T032 [US2] Wire rate limiting middleware to SSE endpoints
- [ ] T033 [US2] Add session ID generation and correlation between GET /sse and POST /sse
- [ ] T034 [US2] Add deprecation warnings to SSE transport (logs, response headers with deprecation notice)
- [ ] T035 [US2] Update mcp-server/src/transports/http/index.ts to add SSE route handlers (GET /sse, POST /sse)
- [ ] T036 [US2] Add error handling and logging for SSE transport
- [ ] T037 [US2] Verify tests pass: Run mcp-server/tests/transports/sse-transport.test.ts

**Checkpoint**: Both HTTP Streamable and SSE transports functional - clients can choose either protocol

---

## Phase 5: User Story 3 - Secure Authenticated Access (Priority: P2)

**Goal**: Ensure robust authentication enforcement, token caching, and permission validation

**Independent Test**: Attempt connections with valid/invalid/expired tokens, verify only valid tokens grant access and permissions are enforced

### Tests for User Story 3 (TDD - Write FIRST, ensure FAIL)

- [ ] T038 [P] [US3] Add authentication edge case tests to mcp-server/tests/auth/token-validator.test.ts (token revocation during session, Redis cache expiry, fallback to API validation)
- [ ] T039 [P] [US3] Create mcp-server/tests/transports/session-manager.test.ts with session lifecycle tests (creation, activity tracking, graceful cleanup, timeout cleanup)

### Implementation for User Story 3

- [ ] T040 [US3] Enhance mcp-server/src/auth/token-validator.ts with Redis caching implementation (5-min TTL, SHA256 token hashing, in-memory fallback)
- [ ] T041 [US3] Enhance mcp-server/src/auth/middleware.ts with detailed error responses (401 with reason, clear error messages)
- [ ] T042 [US3] Add permission enforcement to tool execution (verify token permissions match Vikunja API permissions)
- [ ] T043 [US3] Add security audit logging to mcp-server/src/utils/logger.ts (log all auth attempts, failures, token validation results)
- [ ] T044 [US3] Implement session timeout and cleanup in mcp-server/src/transports/http/session-manager.ts (30-min idle timeout, 60-sec orphaned cleanup)
- [ ] T045 [US3] Add graceful disconnect handling to both HTTP Streamable and SSE transports
- [ ] T046 [US3] Verify tests pass: Run mcp-server/tests/auth/token-validator.test.ts (all edge cases)
- [ ] T047 [US3] Verify tests pass: Run mcp-server/tests/transports/session-manager.test.ts

**Checkpoint**: Authentication robust with caching, permissions enforced, sessions properly managed

---

## Phase 6: User Story 4 - Rate Limiting and Resource Protection (Priority: P3)

**Goal**: Prevent abuse with per-token rate limiting (100 req/15min) and clear error responses

**Independent Test**: Make rapid requests to exceed rate limit, verify 429 errors with retry information

### Tests for User Story 4 (TDD - Write FIRST, ensure FAIL)

- [ ] T048 [P] [US4] Create mcp-server/tests/ratelimit/limiter.test.ts with rate limiting tests (under limit success, over limit 429, per-token isolation, window reset, Redis persistence)

### Implementation for User Story 4

- [ ] T049 [US4] Enhance mcp-server/src/ratelimit/limiter.ts with configurable limits from config (points, duration from env vars)
- [ ] T050 [US4] Enhance mcp-server/src/ratelimit/limiter.ts with error responses (429 with retryAfter, limit info in response data)
- [ ] T051 [US4] Add rate limit enforcement to all HTTP endpoints (POST /mcp, GET /sse, POST /sse)
- [ ] T052 [US4] Add rate limit metrics to health check endpoint (current usage, limits, blocked tokens)
- [ ] T053 [US4] Add rate limit logging to mcp-server/src/utils/logger.ts (log limit exceeded events, reset events)
- [ ] T054 [US4] Verify tests pass: Run mcp-server/tests/ratelimit/limiter.test.ts

**Checkpoint**: Rate limiting operational, abuse prevention active, clear error responses

---

## Phase 7: User Story 5 - Session Management and Cleanup (Priority: P3)

**Goal**: Efficient session lifecycle management with automatic cleanup

**Independent Test**: Establish multiple sessions, disconnect gracefully/abruptly, verify cleanup and resource release

### Tests for User Story 5 (TDD - Write FIRST, ensure FAIL)

- [ ] T055 [P] [US5] Add session cleanup tests to mcp-server/tests/transports/session-manager.test.ts (graceful disconnect cleanup, timeout cleanup, concurrent sessions, resource tracking)

### Implementation for User Story 5

- [ ] T056 [US5] Enhance mcp-server/src/transports/http/session-manager.ts with cleanup interval (run every 5 minutes)
- [ ] T057 [US5] Add session metrics to SessionManager (active count, total created, cleanup stats)
- [ ] T058 [US5] Wire session metrics to health check endpoint in mcp-server/src/transports/http/health-check.ts
- [ ] T059 [US5] Add session event logging (creation, activity, cleanup) to mcp-server/src/utils/logger.ts
- [ ] T060 [US5] Implement graceful shutdown handling in mcp-server/src/index.ts (SIGTERM handler, cleanup all sessions)
- [ ] T061 [US5] Add connection drop detection to HTTP Streamable transport (detect client disconnect)
- [ ] T062 [US5] Add connection drop detection to SSE transport (detect EventSource close)
- [ ] T063 [US5] Verify tests pass: Run mcp-server/tests/transports/session-manager.test.ts (all cleanup scenarios)

**Checkpoint**: Session management robust, resources properly cleaned up, graceful shutdown implemented

---

## Phase 8: Deployment Integration

**Purpose**: Enable production deployment with Proxmox automation

**⚠️ NOTE**: These tasks should be done AFTER Phase 3 (User Story 1) is complete, as they require the HTTP transport to be functional.

- [ ] T064 [P] Update deploy/proxmox/templates/vikunja-mcp.service to add HTTP transport environment variables (MCP_HTTP_ENABLED, MCP_HTTP_PORT, MCP_HTTP_HOST, REDIS_URL, AUTH_*, RATE_LIMIT_*, SESSION_*)
- [ ] T065 Update deploy/proxmox/lib/service-setup.sh generate_systemd_unit() function to populate HTTP transport variables in MCP service template
- [ ] T066 [P] Update deploy/proxmox/README.md with MCP HTTP transport deployment documentation (ports, configuration, testing, health check URLs)
- [ ] T067 [P] Add health check verification to deployment scripts (curl http://localhost:MCP_HTTP_PORT/health after MCP service start)
- [ ] T068 [P] Update deployment summary output to include MCP HTTP transport status (enabled/disabled, port, health URL)

**Checkpoint**: Automated deployment ready for Proxmox LXC containers with HTTP transport support

---

## Phase 9: Documentation & Polish

**Purpose**: Comprehensive documentation and final improvements

- [ ] T071 [P] Update mcp-server/README.md with HTTP transport usage (configuration, client setup, examples for n8n and Claude Desktop)
- [ ] T072 [P] Update mcp-server/CHANGELOG.md with v1.1.0 entry (HTTP Streamable, SSE, authentication, rate limiting, session management)
- [ ] T073 [P] Create mcp-server/docs/migration-sse-to-http-streamable.md migration guide
- [ ] T074 [P] Create example client configurations in mcp-server/examples/ (n8n-config.json, claude-desktop-config.json, custom-client.ts)
- [ ] T075 [P] Add OpenAPI spec files to mcp-server/docs/api/ (copy from specs/007-mcp-http-transport/contracts/)
- [ ] T076 Run full test suite with coverage: pnpm test:coverage (verify 80%+ coverage for HTTP transport code)
- [ ] T077 Run linting: pnpm lint:fix (ensure code quality)
- [ ] T078 Validate deployment: Follow specs/007-mcp-http-transport/quickstart.md on clean Proxmox LXC
- [ ] T079 [P] Add Prometheus metrics endpoint /metrics (optional enhancement from plan)
- [ ] T080 Manual integration test with n8n (connect, list tools, execute get_tasks)
- [ ] T081 Manual integration test with Claude Desktop (configure, test in conversation)

**Checkpoint**: Documentation complete, tests passing, deployment validated, ready for production

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - User Story 1 (Phase 3): Can start after Foundational - HIGHEST PRIORITY
  - User Story 2 (Phase 4): Can start after US1 complete (builds on HTTP infrastructure)
  - User Story 3 (Phase 5): Can start after Foundational - Independent but enhances US1/US2
  - User Story 4 (Phase 6): Can start after Foundational - Independent but enhances US1/US2
  - User Story 5 (Phase 7): Can start after Foundational - Independent but enhances US1/US2
- **Deployment (Phase 8)**: Depends on US1 minimum (MVP), ideally US1+US2 complete
- **Documentation (Phase 9)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Core HTTP Streamable transport - NO dependencies on other stories
- **User Story 2 (P1)**: SSE transport - Depends on US1 infrastructure but can be developed in parallel after T021
- **User Story 3 (P2)**: Enhanced authentication - Independent, can start after Foundational
- **User Story 4 (P3)**: Rate limiting - Independent, can start after Foundational
- **User Story 5 (P3)**: Session management - Independent, can start after Foundational

### Within Each User Story

**Standard TDD Flow** (per Constitution):
1. Write tests FIRST (tasks with test files)
2. Verify tests FAIL (run test suite)
3. Implement feature (tasks with src/ files)
4. Verify tests PASS (checkpoint tasks)
5. Refactor if needed (keep tests green)

**Task Order within Story**:
- Tests before implementation
- Core infrastructure before integrations
- Happy path before edge cases
- Implementation before verification

### Parallel Opportunities

**Setup (Phase 1)**: Tasks T002, T003, T004, T005 can run in parallel

**Foundational (Phase 2)**: Tasks T008, T009, T010, T011, T012 can run in parallel after T006-T007

**User Story 1 Tests**: Tasks T015, T016, T017 can run in parallel (different test files)

**User Story 2**: Can start in parallel with US3/US4/US5 after US1 T021 complete

**User Story 3 Tests**: Tasks T038, T039 can run in parallel (different test files)

**User Story 3-5**: Can all proceed in parallel after Foundational complete (if team capacity allows)

**Deployment**: Tasks T064, T070 can run in parallel

**Documentation**: Tasks T071, T072, T073, T074, T075 can run in parallel

---

## Parallel Execution Examples

### Parallel Example: User Story 1 Tests (TDD Phase)

```bash
# Launch all test creation for User Story 1 together:
Task T015: "Create mcp-server/tests/transports/http-streamable.test.ts"
Task T016: "Create mcp-server/tests/auth/token-validator.test.ts"
Task T017: "Create mcp-server/tests/integration/http-transport.test.ts"

# All can be written in parallel (different files)
```

### Parallel Example: Foundational Infrastructure

```bash
# After config setup (T006-T007), launch in parallel:
Task T008: "Create mcp-server/src/auth/token-validator.ts"
Task T009: "Create mcp-server/src/auth/middleware.ts"
Task T010: "Create mcp-server/src/ratelimit/limiter.ts"
Task T011: "Create mcp-server/src/ratelimit/redis-store.ts"
Task T012: "Create mcp-server/src/transports/http/session-manager.ts"

# All different files, no dependencies between them
```

### Parallel Example: Multiple User Stories

```bash
# After Foundational complete, with 3 developers:
Developer A: User Story 1 (Phase 3) - HTTP Streamable core
Developer B: User Story 3 (Phase 5) - Enhanced authentication  
Developer C: User Story 4 (Phase 6) - Rate limiting

# US2 should follow US1 due to infrastructure dependencies
# US5 can also proceed in parallel (session management)
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2 Only)

1. **Complete Phase 1**: Setup (T001-T005) - ~1 hour
2. **Complete Phase 2**: Foundational (T006-T014) - ~4 hours
3. **Complete Phase 3**: User Story 1 (T015-T028) - ~8 hours
   - **STOP and VALIDATE**: Test HTTP Streamable independently
4. **Complete Phase 4**: User Story 2 (T029-T037) - ~4 hours
   - **STOP and VALIDATE**: Test both transports work
5. **Deploy & Demo**: Basic HTTP transport functional

**MVP Deliverables**: Remote HTTP connection, HTTP Streamable + SSE transports, basic authentication, tool execution

---

### Full Feature (All User Stories)

1. Complete Setup + Foundational (as above)
2. Complete User Story 1 → Test independently
3. Complete User Story 2 → Test independently  
4. Complete User Story 3 → Enhanced authentication with caching
5. Complete User Story 4 → Rate limiting active
6. Complete User Story 5 → Session management robust
7. Complete Deployment Integration → Production ready
8. Complete Documentation → Fully documented

**Each story adds value without breaking previous stories**

---

### Parallel Team Strategy

**With 3 developers**:

1. **Week 1**: All developers complete Setup + Foundational together (T001-T014)
2. **Week 2-3**: Once Foundational done:
   - **Developer A**: User Story 1 (T015-T028) - HTTP Streamable core
   - **Developer B**: User Story 2 (T029-T037) - SSE transport (starts after A completes T021)
   - **Developer C**: User Story 3 (T038-T047) - Enhanced auth (parallel with A/B)
3. **Week 3**: 
   - **Developer A**: User Story 4 (T048-T054) - Rate limiting
   - **Developer B**: User Story 5 (T055-T063) - Session management
   - **Developer C**: Deployment (T064-T070)
4. **Week 4**: All developers on Documentation & Testing (T071-T081)

---

## Test Coverage Target

Per Constitution requirement: **80%+ coverage for HTTP transport code**

**Coverage areas**:
- `src/auth/` - token validation, middleware (80%+)
- `src/ratelimit/` - rate limiting logic (80%+)
- `src/transports/http/` - HTTP Streamable, SSE, session management (80%+)
- `src/config/schema.ts` - config validation (80%+)

**Exclusions** (don't need coverage):
- External dependencies (Vikunja API client calls - use mocks)
- Main entry point boilerplate (index.ts)
- Logger initialization

**Verification**: Task T076 runs `pnpm test:coverage` and confirms 80%+ achieved

---

## Notes

- **[P] tasks**: Different files, can run in parallel
- **[Story] labels**: Map tasks to user stories (US1-US5) for traceability
- **TDD mandatory**: Per Constitution, write tests FIRST, watch them FAIL, then implement
- **Independent stories**: Each user story should be independently completable and testable
- **Commit frequently**: After each task or logical group
- **Checkpoint validation**: Stop at checkpoints to verify story independently
- **Coverage gates**: Must achieve 80%+ coverage before considering story complete

---

## Total Task Count: 81 Tasks

**By Phase**:
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundational): 9 tasks
- Phase 3 (US1): 14 tasks
- Phase 4 (US2): 9 tasks  
- Phase 5 (US3): 10 tasks
- Phase 6 (US4): 7 tasks
- Phase 7 (US5): 9 tasks
- Phase 8 (Deployment): 7 tasks
- Phase 9 (Documentation): 11 tasks

**By User Story**:
- US1 (Remote Connection): 14 tasks
- US2 (Modern Transport): 9 tasks
- US3 (Authentication): 10 tasks
- US4 (Rate Limiting): 7 tasks
- US5 (Session Management): 9 tasks
- Infrastructure (Setup + Foundational): 14 tasks
- Deployment & Docs: 18 tasks

**Parallelization**:
- 28 tasks marked [P] can run in parallel within their phase
- User Stories 3, 4, 5 can all run in parallel after Foundational
- Estimated serial completion: ~6-8 weeks (1 developer)
- Estimated parallel completion: ~3-4 weeks (3 developers)

**MVP Scope** (Recommended):
- Phase 1, 2, 3, 4: Tasks T001-T037 (37 tasks)
- Delivers: HTTP Streamable + SSE transports, basic auth, tool execution
- Estimated time: 2-3 weeks (1 developer), 1-2 weeks (2 developers)
