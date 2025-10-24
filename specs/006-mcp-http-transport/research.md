# Phase 0: Research & Technical Decisions

**Feature**: HTTP Transport for MCP Server  
**Date**: October 22, 2025  
**Status**: Complete

## Overview

This document captures research findings and technical decisions for implementing HTTP transport (SSE and HTTP Streamable) for the Vikunja MCP Server. All unknowns from Technical Context have been researched and resolved.

---

## Research Tasks Completed

### R1: MCP HTTP Streamable Protocol Specification

**Question**: What is the HTTP Streamable protocol format defined in the MCP specification?

**Research Findings**:

The Model Context Protocol (MCP) defines HTTP Streamable as the recommended bidirectional transport for HTTP-based MCP servers:

- **Protocol**: HTTP POST endpoint accepting JSON-RPC 2.0 messages
- **Streaming**: Server sends newline-delimited JSON (NDJSON) responses
- **Message Format**: Each line is a complete JSON-RPC message
- **Capabilities**: Full bidirectional communication (client → server via POST body, server → client via response stream)
- **Connection Lifecycle**: Long-lived HTTP request/response where server keeps connection open to stream multiple responses
- **Headers**: 
  - `Content-Type: application/json` for requests
  - `Content-Type: application/x-ndjson` for streaming responses
  - Custom headers allowed (enables Bearer authentication)

**Decision**: Implement HTTP Streamable as primary transport protocol using @modelcontextprotocol/sdk's built-in support.

**Rationale**: 
- Official MCP specification compliance
- Recommended by n8n and modern MCP clients
- Cleaner architecture than SSE (no EventSource limitations)
- Better debugging (standard HTTP tools work)

**Alternatives Considered**:
- SSE only: Rejected - deprecated, EventSource API limitations (no custom headers)
- WebSocket: Rejected - not part of MCP spec HTTP transports (yet)
- Custom protocol: Rejected - breaks MCP spec compliance

**References**:
- MCP Specification: https://spec.modelcontextprotocol.io/specification/basic/transports/
- MCP TypeScript SDK: https://github.com/modelcontextprotocol/typescript-sdk
- NDJSON format: https://github.com/ndjson/ndjson-spec

---

### R2: Server-Sent Events (SSE) Transport for Backward Compatibility

**Question**: How should SSE transport be implemented for clients that don't support HTTP Streamable yet?

**Research Findings**:

Server-Sent Events (SSE) is an older HTTP transport supported by MCP but being deprecated:

- **Protocol**: 
  - GET endpoint for server → client events (EventSource API)
  - POST endpoint for client → server messages
- **Format**: `text/event-stream` with `data:` prefixed JSON messages
- **Limitation**: EventSource API doesn't support custom headers (auth must use query params)
- **Lifecycle**: Persistent connection for events, separate POST requests for messages
- **Session Management**: Requires session ID to correlate GET stream with POST messages

**Decision**: Implement SSE as secondary transport for backward compatibility, with deprecation notice.

**Rationale**:
- Some older MCP clients may not support HTTP Streamable yet
- Provides migration path for existing deployments
- Can be removed in future major version

**Implementation Notes**:
- Use query parameter for token authentication (`?token=xxx`)
- Generate session ID on first GET request
- Store session state in memory (Map<sessionId, SessionData>)
- POST endpoint validates session ID and routes to correct MCP server instance
- Add deprecation warning in documentation and server logs

**Alternatives Considered**:
- SSE only: Rejected - deprecated, limiting
- No SSE support: Rejected - breaks existing clients during transition period

**References**:
- SSE Specification: https://html.spec.whatwg.org/multipage/server-sent-events.html
- EventSource API: https://developer.mozilla.org/en-US/docs/Web/API/EventSource
- MCP SDK SSE Transport: @modelcontextprotocol/sdk/server/sse.js

---

### R3: Authentication Token Validation Strategy

**Question**: How should Vikunja API tokens be validated efficiently without excessive API calls?

**Research Findings**:

Token validation approaches evaluated:

1. **Direct API validation (every request)**:
   - Pro: Always current, immediate revocation
   - Con: High latency (200ms+), excessive API load

2. **Redis caching with TTL**:
   - Pro: Fast (<10ms), reduces API load
   - Con: Requires Redis, delayed revocation (up to TTL)

3. **In-memory caching with TTL**:
   - Pro: Fastest (<1ms), no external dependency
   - Con: Delayed revocation, memory usage, lost on restart

4. **JWT token decoding (if Vikunja uses JWT)**:
   - Pro: No external calls, self-contained validation
   - Con: Vikunja uses opaque tokens, not JWTs

**Decision**: Redis caching with 5-minute TTL, in-memory fallback if Redis unavailable.

**Rationale**:
- Balances performance (<100ms target) with security (reasonable TTL)
- Redis already required for rate limiting
- Graceful degradation if Redis fails
- 5-minute TTL acceptable for most use cases (tokens rarely revoked immediately)

**Implementation**:
```typescript
async validateToken(token: string): Promise<UserContext> {
  // 1. Check Redis cache
  const cached = await redis.get(`vikunja:mcp:token:${hash(token)}`);
  if (cached) return JSON.parse(cached);
  
  // 2. Validate against Vikunja API
  const userData = await vikunjaClient.validateToken(token);
  
  // 3. Cache result (5 min TTL)
  await redis.setex(
    `vikunja:mcp:token:${hash(token)}`, 
    300, 
    JSON.stringify(userData)
  );
  
  return userData;
}
```

**Security Considerations**:
- Hash token before using as cache key (prevent token leakage in Redis)
- Store minimal user context in cache (ID, permissions only)
- Invalidate cache on server restart (fresh validation)
- Log all authentication attempts (success & failure)

**Alternatives Considered**:
- No caching: Rejected - unacceptable latency
- Longer TTL (>15 min): Rejected - security risk
- JWT migration: Rejected - requires Vikunja backend changes

---

### R4: Rate Limiting Implementation Best Practices

**Question**: What rate limiting strategy works best for per-token abuse prevention?

**Research Findings**:

Rate limiting approaches for MCP server:

1. **Sliding window (Redis sorted sets)**:
   - Pro: Accurate, smooth limit enforcement
   - Con: Memory intensive, complex cleanup

2. **Fixed window (Redis counters)**:
   - Pro: Simple, low memory
   - Con: Burst at window boundaries

3. **Token bucket (in-memory)**:
   - Pro: Allows bursts, simple
   - Con: Lost on restart, no distributed support

4. **Leaky bucket (rate-limiter-flexible library)**:
   - Pro: Production-ready, Redis support, configurable
   - Con: External dependency

**Decision**: Use `rate-limiter-flexible` library with Redis backend, 100 requests per 15 minutes per token.

**Rationale**:
- Battle-tested library (used by major platforms)
- Redis backend for distributed rate limiting (future horizontal scaling)
- Flexible configuration (requests/window, burst allowance)
- Good DX with TypeScript support

**Configuration**:
```typescript
const rateLimiter = new RateLimiterRedis({
  storeClient: redisClient,
  keyPrefix: 'vikunja:mcp:ratelimit',
  points: 100,          // 100 requests
  duration: 15 * 60,    // per 15 minutes
  blockDuration: 60,    // block for 1 minute after limit
});
```

**Rate Limit Rationale** (100 requests / 15 minutes):
- Supports typical workflow automation (n8n polling every 5 min = ~3 requests)
- Prevents abuse (DDoS, credential stuffing)
- Allows legitimate bursts (user manually triggering multiple tools)
- Adjustable via config for different use cases

**Error Response**:
```json
{
  "error": {
    "code": -32000,
    "message": "Rate limit exceeded",
    "data": {
      "retryAfter": 123,  // seconds until reset
      "limit": 100,
      "window": 900
    }
  }
}
```

**Alternatives Considered**:
- Per-IP limiting: Rejected - doesn't prevent token abuse
- No rate limiting: Rejected - security risk
- Custom implementation: Rejected - reinventing wheel

**References**:
- rate-limiter-flexible: https://github.com/animir/node-rate-limiter-flexible
- Redis rate limiting patterns: https://redis.io/docs/latest/develop/reference/patterns/rate-limiting/

---

### R5: Session Management Architecture

**Question**: How should HTTP sessions be managed for multiple concurrent MCP clients?

**Research Findings**:

Session management challenges:

1. **Single shared MCP server instance**:
   - Pro: Simple, low memory
   - Con: All clients share one user context (CURRENT LIMITATION from 006)

2. **MCP server instance per session**:
   - Pro: Perfect isolation, proper multi-user support
   - Con: Higher memory (~50MB per instance), more complex lifecycle

3. **Session multiplexing with context switching**:
   - Pro: Memory efficient, supports multiple users
   - Con: Complex, potential for context leakage bugs

**Decision**: In-memory session Map for Phase 1 (single user context limitation), plan per-instance architecture for Phase 2.

**Phase 1 Implementation** (Current Feature):
```typescript
interface Session {
  id: string;
  token: string;
  userContext: UserContext;
  transport: 'http-streamable' | 'sse';
  createdAt: Date;
  lastActivity: Date;
}

const sessions = new Map<string, Session>();
```

**Known Limitation**: All HTTP sessions share same user context (acceptable for single-user deployments, needs addressing for multi-tenant).

**Phase 2 Migration Path** (Future):
- One MCP server instance per session
- Connection pool pattern
- Session routing layer
- Documented as DEBT-001 in tasks.md

**Cleanup Strategy**:
- Graceful disconnect: Immediate cleanup
- Idle timeout: 30 minutes of inactivity
- Connection drop: Detect via transport events, cleanup after 60 seconds

**Rationale**:
- Phase 1 delivers value for single-user use cases (most current deployments)
- Documented limitation prevents surprises
- Clear migration path for multi-user support

**Alternatives Considered**:
- Per-instance from start: Rejected - over-engineering for MVP
- No session management: Rejected - resource leaks
- Database-backed sessions: Rejected - unnecessary persistence

---

### R6: Health Check Endpoint Design

**Question**: What health check format works best for monitoring MCP server status?

**Research Findings**:

Health check standards:

1. **Simple 200 OK**: Just HTTP status
2. **Kubernetes liveness/readiness**: Separate endpoints
3. **RFC Health Check Response**: JSON format with detailed status
4. **Custom format**: Tailored to application

**Decision**: RFC-inspired JSON health check with MCP-specific metrics.

**Implementation**:
```typescript
GET /health

Response (200 OK):
{
  "status": "healthy",  // healthy | degraded | unhealthy
  "timestamp": "2025-10-22T10:30:00Z",
  "version": "1.1.0",
  "uptime": 3600,  // seconds
  "checks": {
    "redis": {
      "status": "healthy",
      "latency": 5  // ms
    },
    "vikunja_api": {
      "status": "healthy",
      "latency": 120  // ms
    }
  },
  "sessions": {
    "active": 12,
    "total_created": 145
  }
}

Response (503 Service Unavailable) if critical dependency down:
{
  "status": "unhealthy",
  "timestamp": "2025-10-22T10:30:00Z",
  "checks": {
    "vikunja_api": {
      "status": "unhealthy",
      "error": "Connection refused"
    }
  }
}
```

**Rationale**:
- JSON format parseable by monitoring tools
- Detailed status helps debugging
- Separate checks for dependencies (Redis optional, Vikunja API critical)
- Session metrics useful for capacity planning

**Alternatives Considered**:
- Simple 200/503: Rejected - insufficient debugging info
- Kubernetes-style multi-endpoint: Rejected - overkill for single service
- No health check: Rejected - poor operational visibility

**References**:
- RFC Health Check: https://tools.ietf.org/id/draft-inadarei-api-health-check-06.html

---

### R7: Express.js Middleware Architecture

**Question**: How should Express middleware be structured for authentication, rate limiting, and error handling?

**Research Findings**:

Express middleware patterns:

1. **Monolithic middleware** (all in one function)
2. **Composable middleware chain** (separate concerns)
3. **Route-level middleware** (apply per endpoint)
4. **Router-level middleware** (apply to router)

**Decision**: Composable middleware chain with clear separation of concerns.

**Middleware Stack**:
```typescript
// Global middleware (all routes)
app.use(express.json());
app.use(requestLogger);
app.use(errorHandler);

// HTTP Streamable routes (authentication required)
app.post('/mcp', 
  authenticateToken,      // Validate Bearer header
  rateLimitMiddleware,    // Check rate limits
  httpStreamableHandler   // Handle MCP protocol
);

// SSE routes (query param auth)
app.get('/sse',
  authenticateQuery,      // Validate ?token=xxx
  rateLimitMiddleware,
  sseStreamHandler
);

app.post('/sse',
  authenticateQuery,
  rateLimitMiddleware,
  sseMessageHandler
);

// Health check (no auth)
app.get('/health', healthCheckHandler);
```

**Middleware Order** (critical for correctness):
1. Body parsing (express.json)
2. Request logging (before auth for visibility)
3. Authentication (reject early if invalid token)
4. Rate limiting (after auth to limit by token)
5. Business logic (route handler)
6. Error handling (catch-all at end)

**Error Handling Pattern**:
```typescript
class AuthenticationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthenticationError';
  }
}

app.use((err, req, res, next) => {
  if (err instanceof AuthenticationError) {
    return res.status(401).json({
      error: { code: -32001, message: err.message }
    });
  }
  // ... other error types
});
```

**Rationale**:
- Clear separation of concerns (auth, rate limiting, logging)
- Easy to test middleware in isolation
- Standard Express patterns (familiar to developers)
- Type-safe with TypeScript

**Alternatives Considered**:
- Custom framework: Rejected - Express is standard
- No middleware: Rejected - duplicate code in handlers
- Class-based handlers: Rejected - over-engineering

---

### R8: Deployment Integration with Proxmox Scripts

**Question**: How should MCP HTTP server be integrated into existing Proxmox deployment automation?

**Research Findings**:

Deployment options:

1. **Separate systemd service** (mcp-server runs independently)
2. **Same process as Vikunja** (single binary)
3. **Docker sidecar** (container orchestration)
4. **Process manager (PM2)** (Node.js process management)

**Decision**: Separate systemd service with PM2 process manager for Node.js reliability.

**Deployment Architecture**:
```
Proxmox LXC Container (192.168.50.64)
├── Vikunja API (systemd: vikunja.service)
│   └── Port 3456
├── Vikunja Frontend
│   └── Port 3457  
├── MCP Server (systemd: vikunja-mcp.service)
│   ├── Port 3010 (HTTP Streamable / SSE)
│   └── PM2 managed (auto-restart, logging)
└── Redis (systemd: redis.service)
    └── Port 6379
```

**Systemd Service** (`vikunja-mcp.service`):
```ini
[Unit]
Description=Vikunja MCP Server
After=network.target redis.service vikunja.service
Requires=vikunja.service

[Service]
Type=forking
User=vikunja
WorkingDirectory=/opt/vikunja/mcp-server
Environment="NODE_ENV=production"
Environment="VIKUNJA_API_URL=http://localhost:3456"
Environment="MCP_HTTP_PORT=3010"
Environment="REDIS_URL=redis://localhost:6379"
ExecStart=/usr/bin/pm2 start dist/index.js --name vikunja-mcp
ExecReload=/usr/bin/pm2 reload vikunja-mcp
ExecStop=/usr/bin/pm2 stop vikunja-mcp
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

**PM2 Configuration** (`ecosystem.config.js`):
```javascript
module.exports = {
  apps: [{
    name: 'vikunja-mcp',
    script: './dist/index.js',
    instances: 1,
    exec_mode: 'fork',
    env: {
      NODE_ENV: 'production',
      VIKUNJA_API_URL: 'http://localhost:3456',
      MCP_HTTP_PORT: 3010,
    },
    error_file: '/var/log/vikunja-mcp/error.log',
    out_file: '/var/log/vikunja-mcp/out.log',
    log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
  }]
};
```

**Deployment Script Updates** (`deploy/proxmox/deploy.sh`):
```bash
# Install Node.js 22+ and PM2
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt-get install -y nodejs
npm install -g pm2

# Build and deploy MCP server
cd /opt/vikunja/mcp-server
pnpm install --production
pnpm build

# Setup systemd service
cp vikunja-mcp.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable vikunja-mcp
systemctl start vikunja-mcp

# Verify health
sleep 5
curl -f http://localhost:3010/health || exit 1
```

**Rationale**:
- PM2 handles Node.js-specific concerns (auto-restart, graceful reload, logging)
- Systemd integrates with server lifecycle (boot, monitoring)
- Separate service allows independent restarts (don't affect Vikunja API)
- Consistent with existing deployment patterns

**Alternatives Considered**:
- Raw Node.js: Rejected - no auto-restart, poor logging
- Docker: Rejected - adds complexity, existing deployment uses systemd
- Combined service: Rejected - tight coupling, restart affects both

---

## Technology Stack Summary

Based on research, the final technology stack:

**Core**:
- TypeScript 5.x (type safety)
- Node.js 22+ (LTS, modern features)
- @modelcontextprotocol/sdk 1.0+ (HTTP transports)
- Express 4.x (HTTP server)

**Dependencies**:
- ioredis 5.x (Redis client)
- rate-limiter-flexible 5.x (rate limiting)
- uuid 11.x (session IDs)
- winston 3.x (structured logging)
- zod 3.x (config validation)
- axios 1.x (Vikunja API client)

**Development**:
- Vitest 1.x (unit testing, 80%+ coverage)
- supertest 6.x (HTTP endpoint testing)
- ESLint + Prettier (code quality)
- TypeScript strict mode

**Infrastructure**:
- Redis 6+ (caching & rate limiting)
- PM2 (process management)
- systemd (service management)
- nginx/Caddy (reverse proxy, TLS)

---

## Open Questions Resolved

All items marked "NEEDS CLARIFICATION" in Technical Context have been researched:

✅ HTTP Streamable protocol format → Researched (R1)  
✅ SSE implementation strategy → Researched (R2)  
✅ Token validation approach → Researched (R3)  
✅ Rate limiting strategy → Researched (R4)  
✅ Session management architecture → Researched (R5)  
✅ Health check format → Researched (R6)  
✅ Express middleware structure → Researched (R7)  
✅ Deployment integration → Researched (R8)

**Status**: Ready for Phase 1 (Design & Contracts)

---

## Next Steps

1. ✅ Research complete
2. ⏭️ Create data-model.md (session entities, state transitions)
3. ⏭️ Generate API contracts (OpenAPI specs for HTTP Streamable & SSE)
4. ⏭️ Write quickstart.md (deployment guide, client setup)
5. ⏭️ Update agent context (copilot-instructions.md)
6. ⏭️ Re-validate Constitution Check
