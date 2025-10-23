# Implementation Plan: HTTP Transport for MCP Server

**Branch**: `007-mcp-http-transport` | **Date**: October 22, 2025 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-mcp-http-transport/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement HTTP transport layer for the Vikunja MCP Server to enable remote client connections via modern HTTP Streamable protocol and backward-compatible SSE transport. This allows AI workflow tools (n8n, Claude Desktop) to access Vikunja task management capabilities over HTTP with token-based authentication, rate limiting, and session management. The implementation extends the existing stdio-based MCP server with HTTP endpoints while maintaining full compatibility with the established tool suite.

## Technical Context

**Language/Version**: TypeScript 5.x, Node.js 22+  
**Primary Dependencies**: @modelcontextprotocol/sdk (SSE & HTTP Streamable transports), Express 4.x (HTTP server), Zod (config validation), ioredis (token caching), rate-limiter-flexible (abuse prevention), uuid (session IDs), winston (logging)  
**Storage**: Redis (token cache & rate limiting state, 5-minute TTL), in-memory fallback if Redis unavailable  
**Testing**: Vitest (unit tests, 80%+ coverage target), supertest (HTTP endpoint testing), manual integration testing with n8n & Claude Desktop  
**Target Platform**: Linux server (Proxmox LXC containers), Docker-compatible  
**Project Type**: Single Node.js TypeScript project (mcp-server/)  
**Performance Goals**: <2s connection establishment, 50 concurrent clients, <100ms token validation (cached), <500KB memory per session  
**Constraints**: <200ms HTTP transport overhead vs direct API calls, <100MB memory for 10 sessions, must work with EventSource API limitations (no custom headers)  
**Scale/Scope**: Single MCP server instance, 50 concurrent sessions, existing 15+ MCP tools, 2 transport protocols (SSE + HTTP Streamable), 3 authentication methods (Bearer header, query param, future OAuth)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality Standards ✅

**Architecture**: 
- ✅ TypeScript project with clear module separation (transports/, auth/, ratelimit/)
- ✅ No Chef/Waiter/Pantry pattern needed (Node.js project, not Go backend)
- ✅ Clean separation: Transport handlers → Auth middleware → Vikunja API client
- ✅ Service-like architecture: VikunjaMcpServer (orchestration), separate transport handlers

**Quality Gates**:
- ✅ ESLint + Prettier configured (existing mcp-server setup)
- ✅ TypeScript strict mode for type safety
- ✅ No technical debt from this implementation (greenfield HTTP transport)

**Compliance**: PASS - TypeScript best practices, modular design, linting enforced

---

### II. Test-First Development ✅

**TDD Approach**:
- ✅ Write Vitest tests for each transport module before implementation
- ✅ Test authentication flow: valid token → success, invalid → 401
- ✅ Test rate limiting: under limit → success, over limit → 429
- ✅ Test session lifecycle: connect → active → disconnect → cleanup
- ✅ Test both positive and negative scenarios

**Coverage Target**: 80%+ for new HTTP transport code (src/transports/http/)

**Test Structure**:
```
tests/
├── transports/
│   ├── http-streamable.test.ts  # HTTP Streamable protocol tests
│   ├── sse-transport.test.ts    # SSE transport tests
│   └── session-manager.test.ts  # Session lifecycle tests
├── auth/
│   └── token-validator.test.ts  # Authentication tests
└── integration/
    └── end-to-end.test.ts        # Full connection flow
```

**DO NOT Test**: Vikunja API backend (external dependency, use mocks)

**Compliance**: PASS - TDD workflow planned, comprehensive test coverage

---

### III. User Experience Consistency ⚠️

**Not Applicable**: This is a backend HTTP transport server, no frontend UI changes.

**Client Experience** (MCP client configuration):
- ✅ Clear error messages for auth failures, rate limiting
- ✅ Standard HTTP status codes (401, 429, 500)
- ✅ Consistent JSON error format
- ✅ Documentation for connection setup

**Compliance**: PASS - Backend-only feature, appropriate error handling planned

---

### IV. Performance Requirements ✅

**Connection Performance**:
- ✅ Target: <2s connection establishment (meets <3s general guideline)
- ✅ Token validation: <100ms (cached) vs <200ms (fresh)
- ✅ Concurrent connections: 50 clients (reasonable for single instance)
- ✅ Memory: <100MB per 10 sessions (<512MB total typical workload)

**Optimization Strategy**:
- ✅ Redis token caching (5-min TTL) reduces API calls
- ✅ In-memory session storage for low overhead
- ✅ Connection pooling for Vikunja API client
- ✅ Async/await for non-blocking I/O

**Monitoring**:
- ✅ Health check endpoint (/health)
- ✅ Winston structured logging
- ✅ Session count tracking

**Compliance**: PASS - Performance targets defined and achievable

---

### V. Security & Reliability Standards ✅

**Authentication** (NON-NEGOTIABLE):
- ✅ ALL HTTP endpoints require Vikunja API token authentication
- ✅ Bearer header support (standard)
- ✅ Query parameter fallback (EventSource API limitation)
- ✅ Token validation against Vikunja API backend
- ✅ Permissions enforced by Vikunja backend (not duplicated)

**Input Validation**:
- ✅ Zod schema validation for configuration
- ✅ MCP SDK handles protocol message validation
- ✅ Rate limiting per token (100 requests / 15 minutes)

**Data Protection**:
- ✅ NO tokens in plaintext logs (hash or redact)
- ✅ Secure token handling in memory
- ✅ HTTPS enforced via reverse proxy (nginx/Caddy)
- ✅ Redis password-protected

**Error Handling**:
- ✅ Graceful degradation: Redis failure → in-memory fallback or reject
- ✅ Clear error messages for clients
- ✅ Security event logging (auth failures, rate limits)

**Reliability**:
- ✅ Session timeout and cleanup (no resource leaks)
- ✅ Graceful shutdown on SIGTERM
- ✅ Health check for monitoring

**Compliance**: PASS - Security requirements fully addressed

---

## Constitution Summary

**Overall Status**: ✅ **PASS** - Ready for Phase 0 Research

All five principles met:
1. ✅ Code Quality: TypeScript, modular architecture, linting
2. ✅ Test-First: TDD workflow, 80%+ coverage target
3. ✅ UX Consistency: N/A (backend), appropriate error handling
4. ✅ Performance: Defined targets, optimization strategy
5. ✅ Security: Authentication, validation, rate limiting, logging

**No violations requiring justification.**

## Project Structure

### Documentation (this feature)

```
specs/007-mcp-http-transport/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (MCP protocol research)
├── data-model.md        # Phase 1 output (session & state models)
├── quickstart.md        # Phase 1 output (deployment & client setup)
├── contracts/           # Phase 1 output (API contracts)
│   ├── http-streamable.openapi.yaml
│   └── sse-transport.openapi.yaml
├── checklists/
│   └── requirements.md  # Quality validation (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT yet created)
```

### Source Code (repository root)

```
mcp-server/                          # Existing TypeScript MCP server
├── src/
│   ├── index.ts                     # Entry point (stdio OR http mode)
│   ├── server.ts                    # VikunjaMcpServer class (existing)
│   ├── config/
│   │   ├── index.ts                 # Config loading & validation
│   │   └── schema.ts                # [NEW] Zod schemas for HTTP transport
│   ├── auth/
│   │   ├── token-validator.ts       # [NEW] Vikunja token authentication
│   │   └── middleware.ts            # [NEW] Express auth middleware
│   ├── ratelimit/
│   │   ├── limiter.ts               # [NEW] Rate limiter implementation
│   │   └── redis-store.ts           # [NEW] Redis backend for limits
│   ├── transports/
│   │   ├── stdio/                   # [EXISTING] stdio transport
│   │   │   └── stdio-transport.ts
│   │   └── http/                    # [NEW] HTTP transports
│   │       ├── index.ts             # Export all HTTP transports
│   │       ├── http-streamable.ts   # HTTP Streamable transport handler
│   │       ├── sse-transport.ts     # SSE transport handler (backward compat)
│   │       ├── session-manager.ts   # Session lifecycle management
│   │       └── health-check.ts      # Health check endpoint
│   ├── vikunja/
│   │   ├── client.ts                # [MODIFY] Add connection pooling
│   │   └── types.ts                 # Vikunja API types
│   ├── tools/                       # [EXISTING] 15+ MCP tools
│   ├── resources/                   # [EXISTING] MCP resources
│   └── utils/
│       ├── logger.ts                # Winston logger
│       └── errors.ts                # Custom error classes
├── tests/
│   ├── transports/
│   │   ├── http-streamable.test.ts  # [NEW] HTTP Streamable tests
│   │   ├── sse-transport.test.ts    # [NEW] SSE tests
│   │   └── session-manager.test.ts  # [NEW] Session tests
│   ├── auth/
│   │   └── token-validator.test.ts  # [NEW] Auth tests
│   ├── ratelimit/
│   │   └── limiter.test.ts          # [NEW] Rate limit tests
│   └── integration/
│       └── http-transport.test.ts   # [NEW] End-to-end tests
├── package.json                     # [MODIFY] Add new dependencies
├── tsconfig.json                    # TypeScript config
├── vitest.config.ts                 # Test configuration
├── .env.example                     # [MODIFY] Add HTTP transport vars
├── README.md                        # [MODIFY] Document HTTP usage
└── CHANGELOG.md                     # [MODIFY] Add v1.1.0 entry

deploy/proxmox/                      # Deployment automation
├── deploy.sh                        # [MODIFY] Add MCP port config
├── mcp-server.service               # [NEW] systemd service file
└── README.md                        # [MODIFY] MCP deployment docs
```

**Structure Decision**: Single TypeScript project (mcp-server/) extending existing stdio-based MCP server with HTTP transport capabilities. Clean modular separation: transports/ (protocol handlers), auth/ (authentication), ratelimit/ (abuse prevention), vikunja/ (API client). No frontend changes needed (backend-only feature).

## Complexity Tracking

*No Constitution violations - this section intentionally left empty.*

All complexity is justified:
- TypeScript project matches existing mcp-server codebase
- Modular architecture follows single responsibility principle
- No additional abstractions beyond necessary (transport handlers, auth, rate limiting)
- Complexity controlled via test-first development and 80%+ coverage target

---

## Phase 0: Research ✅

**Status**: Complete

All technical unknowns researched and documented in [research.md](research.md):

- ✅ R1: MCP HTTP Streamable protocol specification
- ✅ R2: SSE transport for backward compatibility
- ✅ R3: Token validation strategy (Redis caching, 5-min TTL)
- ✅ R4: Rate limiting implementation (rate-limiter-flexible library)
- ✅ R5: Session management architecture (in-memory Map, future per-instance)
- ✅ R6: Health check endpoint design (RFC-inspired JSON format)
- ✅ R7: Express middleware architecture (composable chain)
- ✅ R8: Proxmox deployment integration (PM2 + systemd)

**Key Decisions**:
- HTTP Streamable as primary transport (modern MCP standard)
- SSE for backward compatibility (deprecated, will remove in v2.0)
- Redis caching for performance, in-memory fallback
- 100 requests / 15 minutes rate limit per token
- Session-based architecture with cleanup (30 min idle timeout)

---

## Phase 1: Design & Contracts ✅

**Status**: Complete

### Data Model

Defined in [data-model.md](data-model.md):

**Core Entities**:
1. **Session**: Active MCP connection (id, token, userContext, transport, timestamps)
2. **UserContext**: Authenticated user identity (userId, username, permissions)
3. **RateLimitState**: Request tracking (tokenHash, points, window)
4. **ClientInfo**: Optional client metadata (userAgent, mcpVersion)
5. **TransportMessage**: JSON-RPC 2.0 message structure

**State Machine**: Session lifecycle (Created → Active → Orphaned → Terminated)

**Storage Strategy**:
- Sessions: In-memory Map (fast, ephemeral)
- UserContext: Redis cache (5-min TTL, shared)
- RateLimitState: Redis (distributed, 16-min TTL)

**Validation**: Zod schemas for all entities

### API Contracts

Generated OpenAPI 3.1 specifications in [contracts/](contracts/):

1. **http-streamable.openapi.yaml**: 
   - POST /mcp (bidirectional MCP protocol)
   - GET /health (monitoring)
   - Bearer authentication
   - NDJSON streaming responses

2. **sse-transport.openapi.yaml**:
   - GET /sse (event stream, deprecated)
   - POST /sse (client messages, deprecated)
   - Query parameter authentication
   - Deprecation notices

**Error Codes**:
- -32001: Authentication failed
- -32002: Rate limit exceeded
- -32003: Session not found
- Standard JSON-RPC codes (-32700 to -32603)

### Quickstart Guide

Deployment and usage documentation in [quickstart.md](quickstart.md):

**Covers**:
- 5-minute quick start (install, configure, deploy)
- Client setup (n8n, Claude Desktop, custom)
- Production deployment (Proxmox LXC, nginx TLS)
- Monitoring (health checks, logs, metrics)
- Troubleshooting (common issues, solutions)
- Security best practices
- Migration from stdio to HTTP
- Performance tuning

### Agent Context

Updated `.github/copilot-instructions.md` with:
- TypeScript 5.x, Node.js 22+ stack
- MCP SDK, Express, Zod, ioredis, rate-limiter-flexible
- Redis for caching and rate limiting
- Feature 007 technologies and patterns

---

## Constitution Check (Post-Design) ✅

**Re-validation after Phase 1 design**:

### I. Code Quality Standards ✅

**Verification**:
- ✅ Modular TypeScript architecture (transports/, auth/, ratelimit/)
- ✅ Clear separation of concerns (no business logic in routes)
- ✅ Zod validation schemas prevent runtime errors
- ✅ ESLint + Prettier configured and enforced

**Status**: PASS - Architecture aligns with quality standards

---

### II. Test-First Development ✅

**Verification**:
- ✅ Test structure planned (tests/transports/, tests/auth/, tests/integration/)
- ✅ 80%+ coverage target for HTTP transport code
- ✅ Both positive and negative test cases planned
- ✅ Vitest + supertest for HTTP endpoint testing

**Status**: PASS - TDD workflow ready for implementation

---

### III. User Experience Consistency ✅

**Verification**:
- ✅ Clear, actionable error messages defined (401, 429, 500)
- ✅ Standard HTTP status codes and JSON-RPC errors
- ✅ Comprehensive client documentation (quickstart.md)
- ✅ Migration guide for stdio → HTTP

**Status**: PASS - Backend-only feature with appropriate UX considerations

---

### IV. Performance Requirements ✅

**Verification**:
- ✅ <2s connection establishment target (meets guideline)
- ✅ <100ms token validation (cached, meets <200ms)
- ✅ 50 concurrent clients supported
- ✅ <100MB memory per 10 sessions (meets <512MB)
- ✅ Redis caching reduces API load
- ✅ Health check endpoint for monitoring

**Status**: PASS - Performance targets defined and achievable

---

### V. Security & Reliability Standards ✅

**Verification**:
- ✅ 100% authentication enforcement (all endpoints except /health)
- ✅ Bearer header + query param (EventSource limitation) auth
- ✅ Tokens hashed before storage (SHA256)
- ✅ No tokens in plaintext logs
- ✅ Rate limiting per token (100 req / 15 min)
- ✅ Zod validation for all inputs
- ✅ TLS via reverse proxy (nginx)
- ✅ Graceful error handling and logging
- ✅ Session cleanup (no resource leaks)

**Status**: PASS - Security requirements fully addressed

---

## Post-Design Constitution Summary

**Overall Status**: ✅ **PASS** - Ready for Phase 2 (Tasks)

All five principles validated after design:
1. ✅ Code Quality: Modular TypeScript, clean architecture, linting
2. ✅ Test-First: TDD planned, 80%+ coverage, comprehensive tests
3. ✅ UX Consistency: Clear errors, documentation, migration guides
4. ✅ Performance: Targets met, optimization strategy, monitoring
5. ✅ Security: Authentication, validation, rate limiting, TLS

**No technical debt introduced during design phase.**

---

## Summary

**Planning Phase Complete** ✅

**Artifacts Generated**:
1. ✅ plan.md (this file) - Implementation plan
2. ✅ research.md - Technical research and decisions
3. ✅ data-model.md - Entity definitions and state machines
4. ✅ contracts/http-streamable.openapi.yaml - HTTP Streamable API spec
5. ✅ contracts/sse-transport.openapi.yaml - SSE API spec (deprecated)
6. ✅ quickstart.md - Deployment and usage guide
7. ✅ .github/copilot-instructions.md - Updated agent context

**Constitution Checks**:
- ✅ Pre-research: PASS (all principles met)
- ✅ Post-design: PASS (validated after Phase 1)

**Next Phase**: `/speckit.tasks` to generate implementation task breakdown

**Ready for Implementation**: Yes ✅
