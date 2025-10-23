# Phase 1: Data Model

**Feature**: HTTP Transport for MCP Server  
**Date**: October 22, 2025  
**Status**: Complete

## Overview

This document defines the data entities and their relationships for HTTP transport session management, authentication, and rate limiting in the Vikunja MCP Server.

---

## Core Entities

### 1. Session

Represents an active MCP client connection over HTTP transport.

**Purpose**: Track connection state, user context, and lifecycle for each client.

**Attributes**:

| Attribute | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `id` | string (UUID v4) | Yes | Unique session identifier | Non-empty, valid UUID format |
| `token` | string | Yes | Vikunja API token (hashed for storage) | Non-empty, min 20 chars |
| `userContext` | UserContext | Yes | Authenticated user information | Valid UserContext object |
| `transport` | enum | Yes | Transport protocol used | 'http-streamable' \| 'sse' |
| `createdAt` | Date | Yes | Session creation timestamp | Valid ISO 8601 date |
| `lastActivity` | Date | Yes | Last request/message timestamp | Valid ISO 8601 date, >= createdAt |
| `clientInfo` | ClientInfo | No | Client metadata (user agent, version) | Valid ClientInfo object |

**Relationships**:
- One Session → One UserContext (composition)
- One Session → One ClientInfo (composition, optional)
- One Session → Many RateLimitRecords (via token, external)

**State Transitions**:

```
[Created] ──authenticate──> [Active]
    │
    └──auth_failure──> [Terminated]

[Active] ──activity──> [Active] (update lastActivity)
    │
    ├──disconnect──> [Terminated]
    │
    ├──idle_timeout──> [Terminated] (30 min)
    │
    └──connection_lost──> [Orphaned] ──cleanup_timeout──> [Terminated] (60 sec)
```

**Example**:
```typescript
{
  id: "550e8400-e29b-41d4-a716-446655440000",
  token: "sha256:abc123...", // hashed
  userContext: {
    userId: 42,
    username: "alice",
    permissions: ["read", "write"]
  },
  transport: "http-streamable",
  createdAt: "2025-10-22T10:00:00Z",
  lastActivity: "2025-10-22T10:15:00Z",
  clientInfo: {
    userAgent: "n8n/1.0.0",
    mcpVersion: "1.0"
  }
}
```

**Storage**: In-memory Map<sessionId, Session> (Phase 1)

**Indexes**: By session ID (Map key), by token (for lookup)

---

### 2. UserContext

Represents an authenticated Vikunja user's identity and permissions.

**Purpose**: Associate MCP operations with a Vikunja user, enforce permissions.

**Attributes**:

| Attribute | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `userId` | number | Yes | Vikunja user ID | Positive integer |
| `username` | string | Yes | Vikunja username | Non-empty, alphanumeric |
| `email` | string | No | User email | Valid email format |
| `permissions` | string[] | Yes | User capabilities | Non-empty array, valid permission strings |
| `tokenScopes` | string[] | No | API token-specific scopes | Array of scope strings |
| `validatedAt` | Date | Yes | When token was last validated | Valid ISO 8601 date |

**Relationships**:
- One UserContext → One Vikunja User (external, via userId)
- Many Sessions → One UserContext (if same token reused)

**Validation Rules**:
- `userId` must exist in Vikunja backend
- `permissions` must be subset of user's actual permissions
- `validatedAt` must be within cache TTL (5 minutes)

**Example**:
```typescript
{
  userId: 42,
  username: "alice",
  email: "alice@example.com",
  permissions: ["task:read", "task:write", "project:read"],
  tokenScopes: ["api:tasks", "api:projects"],
  validatedAt: "2025-10-22T10:00:00Z"
}
```

**Storage**: 
- In-memory (part of Session object)
- Redis cache (for token validation, 5-min TTL)

**Cache Key**: `vikunja:mcp:token:${sha256(token)}`

---

### 3. RateLimitState

Represents current rate limit consumption for a token.

**Purpose**: Track request counts and enforce rate limits per token.

**Attributes**:

| Attribute | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `tokenHash` | string | Yes | SHA256 hash of API token | 64 hex characters |
| `points` | number | Yes | Remaining points (requests) | 0 to maxPoints |
| `maxPoints` | number | Yes | Total points per window | Positive integer (default: 100) |
| `windowStart` | Date | Yes | Rate limit window start time | Valid ISO 8601 date |
| `windowDuration` | number | Yes | Window duration in seconds | Positive integer (default: 900 = 15 min) |
| `blockedUntil` | Date | No | When block expires (if over limit) | Valid ISO 8601 date |

**Relationships**:
- One RateLimitState per unique token
- Referenced by Session.token (hashed)

**State Transitions**:

```
[New Window] ──first_request──> [Active] (points = maxPoints - 1)
    │
    └──window_expired──> [New Window] (reset points)

[Active] ──request──> [Active] (points -= 1)
    │
    ├──points_remain──> [Active]
    │
    └──points_exhausted──> [Blocked] ──block_duration_expires──> [New Window]
```

**Calculation Logic**:
```typescript
// Check if rate limit allows request
function checkRateLimit(tokenHash: string): {allowed: boolean, retryAfter?: number} {
  const state = getRateLimitState(tokenHash);
  
  // Reset if window expired
  if (Date.now() - state.windowStart.getTime() > state.windowDuration * 1000) {
    state.points = state.maxPoints;
    state.windowStart = new Date();
    state.blockedUntil = null;
  }
  
  // Check if blocked
  if (state.blockedUntil && Date.now() < state.blockedUntil.getTime()) {
    const retryAfter = Math.ceil((state.blockedUntil.getTime() - Date.now()) / 1000);
    return { allowed: false, retryAfter };
  }
  
  // Check if points available
  if (state.points > 0) {
    state.points -= 1;
    return { allowed: true };
  }
  
  // No points, block for 1 minute
  state.blockedUntil = new Date(Date.now() + 60 * 1000);
  return { allowed: false, retryAfter: 60 };
}
```

**Example**:
```typescript
{
  tokenHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  points: 75,
  maxPoints: 100,
  windowStart: "2025-10-22T10:00:00Z",
  windowDuration: 900,  // 15 minutes
  blockedUntil: null
}
```

**Storage**: Redis (managed by rate-limiter-flexible library)

**Redis Key**: `vikunja:mcp:ratelimit:${tokenHash}`

**TTL**: windowDuration + blockDuration (16 minutes total)

---

### 4. ClientInfo

Optional metadata about the MCP client software.

**Purpose**: Debugging, analytics, version compatibility checks.

**Attributes**:

| Attribute | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `userAgent` | string | No | HTTP User-Agent header | Max 500 chars |
| `mcpVersion` | string | No | MCP protocol version | Semver format (e.g., "1.0.0") |
| `clientName` | string | No | Client application name | Max 100 chars |
| `clientVersion` | string | No | Client application version | Semver format |

**Relationships**:
- One ClientInfo → One Session (optional, embedded)

**Example**:
```typescript
{
  userAgent: "Mozilla/5.0 (compatible; n8n/1.0.0)",
  mcpVersion: "1.0.0",
  clientName: "n8n",
  clientVersion: "1.0.0"
}
```

**Storage**: In-memory (part of Session object)

**Validation**: 
- `mcpVersion` must be compatible (major version match)
- Optional fields, no enforcement if missing

---

### 5. TransportMessage

Represents a message exchanged via MCP protocol.

**Purpose**: Define structure of client ↔ server communication.

**Attributes**:

| Attribute | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `jsonrpc` | string | Yes | JSON-RPC version | Must be "2.0" |
| `id` | string \| number | No | Request/response ID | Unique per request |
| `method` | string | No | RPC method name | Required for requests |
| `params` | object | No | Method parameters | Valid JSON object |
| `result` | any | No | Response result | Required for success responses |
| `error` | ErrorObject | No | Error details | Required for error responses |

**Subtypes**:

**Request Message**:
```typescript
{
  jsonrpc: "2.0",
  id: 1,
  method: "tools/list",
  params: {}
}
```

**Success Response**:
```typescript
{
  jsonrpc: "2.0",
  id: 1,
  result: {
    tools: [...]
  }
}
```

**Error Response**:
```typescript
{
  jsonrpc: "2.0",
  id: 1,
  error: {
    code: -32001,
    message: "Authentication failed",
    data: { ... }
  }
}
```

**Notification** (no response expected):
```typescript
{
  jsonrpc: "2.0",
  method: "notifications/initialized"
}
```

**Validation**: Handled by @modelcontextprotocol/sdk

**Storage**: Not persisted (ephemeral, in transit only)

---

## Entity Relationships Diagram

```
┌─────────────────┐
│   Session       │
│                 │
│ - id            │
│ - token         │───────────┐
│ - transport     │           │
│ - createdAt     │           │
│ - lastActivity  │           ▼
└────────┬────────┘    ┌──────────────┐
         │             │ RateLimitState│
         │ 1           │                │
         │             │ - tokenHash    │
         │ owns        │ - points       │
         │             │ - windowStart  │
         ▼ 1           └────────────────┘
┌─────────────────┐           ▲
│  UserContext    │           │
│                 │           │ referenced by
│ - userId        │           │ (via token hash)
│ - username      │           │
│ - permissions   │           │
│ - validatedAt   │───────────┘
└─────────────────┘    cached in Redis
         │
         │ 1
         │ optionally has
         ▼ 0..1
┌─────────────────┐
│   ClientInfo    │
│                 │
│ - userAgent     │
│ - mcpVersion    │
│ - clientName    │
└─────────────────┘
```

**Key Relationships**:
1. Session owns UserContext (composition, 1:1)
2. Session optionally has ClientInfo (composition, 1:0..1)
3. RateLimitState referenced by Session.token (hashed, 1:1)
4. Multiple Sessions can share same UserContext (if same token, cached)

---

## State Machine: Session Lifecycle

```
┌──────────┐
│  START   │
└────┬─────┘
     │ create()
     ▼
┌──────────────┐  authenticate(token)  ┌──────────────┐
│   CREATED    │──────────────────────>│    ACTIVE    │
└──────┬───────┘                       └──────┬───────┘
       │                                      │
       │ auth_failure()                       │ activity()
       │                                      ├──────────┐
       ▼                                      │          │
┌──────────────┐                             │◄─────────┘
│ TERMINATED   │                             │
└──────────────┘                             │ disconnect()
       ▲                                     │ idle_timeout(30min)
       │                                     │
       │                                     ▼
       │                              ┌──────────────┐
       │                              │   ORPHANED   │
       │                              │ (conn lost)  │
       │                              └──────┬───────┘
       │                                     │
       │ cleanup_timeout(60sec)              │
       └─────────────────────────────────────┘
```

**State Descriptions**:

- **CREATED**: Session object initialized, awaiting authentication
- **ACTIVE**: Token validated, client can send MCP requests
- **ORPHANED**: Connection lost unexpectedly, awaiting cleanup timeout
- **TERMINATED**: Session ended, resources freed

**Triggers**:

- `create()`: New HTTP connection established
- `authenticate(token)`: Token validated successfully
- `auth_failure()`: Invalid token, reject session
- `activity()`: Client sends message, update lastActivity
- `disconnect()`: Graceful client disconnect
- `idle_timeout(30min)`: No activity for 30 minutes
- `cleanup_timeout(60sec)`: Connection lost, cleanup after timeout

---

## Validation Rules Summary

### Session Validation
```typescript
const sessionSchema = z.object({
  id: z.string().uuid(),
  token: z.string().min(20),  // Hashed format
  userContext: userContextSchema,
  transport: z.enum(['http-streamable', 'sse']),
  createdAt: z.date(),
  lastActivity: z.date(),
  clientInfo: clientInfoSchema.optional(),
}).refine(data => data.lastActivity >= data.createdAt, {
  message: "lastActivity must be >= createdAt"
});
```

### UserContext Validation
```typescript
const userContextSchema = z.object({
  userId: z.number().int().positive(),
  username: z.string().min(1).regex(/^[a-zA-Z0-9_-]+$/),
  email: z.string().email().optional(),
  permissions: z.array(z.string()).min(1),
  tokenScopes: z.array(z.string()).optional(),
  validatedAt: z.date(),
});
```

### RateLimitState Validation
```typescript
const rateLimitStateSchema = z.object({
  tokenHash: z.string().length(64).regex(/^[a-f0-9]+$/),
  points: z.number().int().min(0),
  maxPoints: z.number().int().positive(),
  windowStart: z.date(),
  windowDuration: z.number().int().positive(),
  blockedUntil: z.date().optional(),
}).refine(data => data.points <= data.maxPoints, {
  message: "points cannot exceed maxPoints"
});
```

### ClientInfo Validation
```typescript
const clientInfoSchema = z.object({
  userAgent: z.string().max(500).optional(),
  mcpVersion: z.string().regex(/^\d+\.\d+\.\d+$/).optional(),
  clientName: z.string().max(100).optional(),
  clientVersion: z.string().regex(/^\d+\.\d+\.\d+$/).optional(),
});
```

---

## Data Access Patterns

### Pattern 1: Create Session
```typescript
async function createSession(
  token: string, 
  transport: 'http-streamable' | 'sse',
  clientInfo?: ClientInfo
): Promise<Session> {
  // 1. Validate token and get user context
  const userContext = await validateToken(token);
  
  // 2. Create session object
  const session: Session = {
    id: uuidv4(),
    token: sha256(token),  // Hash for storage
    userContext,
    transport,
    createdAt: new Date(),
    lastActivity: new Date(),
    clientInfo,
  };
  
  // 3. Store in session map
  sessions.set(session.id, session);
  
  return session;
}
```

### Pattern 2: Validate Token (with caching)
```typescript
async function validateToken(token: string): Promise<UserContext> {
  const cacheKey = `vikunja:mcp:token:${sha256(token)}`;
  
  // 1. Check Redis cache
  const cached = await redis.get(cacheKey);
  if (cached) {
    const userContext = JSON.parse(cached);
    // Verify not expired (5 min TTL)
    if (Date.now() - new Date(userContext.validatedAt).getTime() < 5 * 60 * 1000) {
      return userContext;
    }
  }
  
  // 2. Validate against Vikunja API
  const response = await vikunjaClient.get('/api/v1/user', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const userContext: UserContext = {
    userId: response.data.id,
    username: response.data.username,
    email: response.data.email,
    permissions: extractPermissions(response.data),
    validatedAt: new Date(),
  };
  
  // 3. Cache in Redis (5 min TTL)
  await redis.setex(cacheKey, 300, JSON.stringify(userContext));
  
  return userContext;
}
```

### Pattern 3: Check Rate Limit
```typescript
async function checkRateLimit(token: string): Promise<void> {
  const tokenHash = sha256(token);
  
  try {
    await rateLimiter.consume(tokenHash, 1);  // Consume 1 point
  } catch (rateLimiterRes) {
    // Rate limit exceeded
    const retryAfter = Math.ceil(rateLimiterRes.msBeforeNext / 1000);
    throw new RateLimitError(
      `Rate limit exceeded. Retry after ${retryAfter} seconds.`,
      retryAfter
    );
  }
}
```

### Pattern 4: Update Session Activity
```typescript
function updateSessionActivity(sessionId: string): void {
  const session = sessions.get(sessionId);
  if (session) {
    session.lastActivity = new Date();
  }
}
```

### Pattern 5: Cleanup Stale Sessions
```typescript
function cleanupStaleSessions(): void {
  const now = Date.now();
  const idleTimeout = 30 * 60 * 1000;  // 30 minutes
  
  for (const [sessionId, session] of sessions.entries()) {
    const idle = now - session.lastActivity.getTime();
    if (idle > idleTimeout) {
      sessions.delete(sessionId);
      logger.info('Session cleaned up due to idle timeout', { sessionId });
    }
  }
}

// Run cleanup every 5 minutes
setInterval(cleanupStaleSessions, 5 * 60 * 1000);
```

---

## Storage Strategy

| Entity | Storage | Persistence | TTL | Indexing |
|--------|---------|-------------|-----|----------|
| Session | In-memory Map | No (lost on restart) | Idle timeout (30 min) | By sessionId |
| UserContext | Redis cache | No (cache only) | 5 minutes | By token hash |
| RateLimitState | Redis | No (ephemeral) | 16 minutes (window + block) | By token hash |
| ClientInfo | In-memory (embedded) | No | Same as Session | N/A |

**Rationale**:
- **In-memory Sessions**: Fast access, acceptable to lose on restart (clients reconnect)
- **Redis UserContext**: Shared across restarts/instances (future), reduces API calls
- **Redis RateLimitState**: Distributed rate limiting (future horizontal scaling)
- **No Database**: Sessions are ephemeral, no need for persistence

**Future Enhancements** (out of scope for Phase 1):
- Database-backed sessions for persistence across restarts
- Multi-instance session sharing via Redis Pub/Sub
- Session migration for zero-downtime deployments

---

## Performance Considerations

**Memory Usage**:
- Session object: ~2KB (including UserContext)
- 50 concurrent sessions: ~100KB total
- Redis cache: ~5KB per token (300 tokens = ~1.5MB)
- Well within <100MB target

**Lookup Performance**:
- Session by ID: O(1) via Map
- Token validation: O(1) Redis GET (cached)
- Rate limit check: O(1) Redis INCR

**Scalability**:
- Single instance: 50 concurrent sessions (target)
- Future: Horizontal scaling with Redis-backed sessions

---

## Next Steps

1. ✅ Data model defined
2. ⏭️ Generate API contracts (OpenAPI specs)
3. ⏭️ Write quickstart.md (deployment guide)
4. ⏭️ Update agent context
5. ⏭️ Re-validate Constitution Check
