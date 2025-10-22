# Data Model: MCP HTTP/SSE Transport

**Feature**: MCP HTTP/SSE Transport  
**Phase**: 1 - Design  
**Date**: 2025-10-22

## Overview

This document defines the data structures and state management for HTTP/SSE transport implementation. Since this is a transport layer feature, the "data model" focuses on configuration schemas, runtime state, and connection lifecycle rather than persistent database entities.

## Configuration Schema

### Transport Configuration

**Source File**: `mcp-server/src/config/index.ts`

```typescript
import { z } from 'zod';

/**
 * Transport type enumeration
 */
export const TransportType = z.enum(['stdio', 'http']);
export type TransportType = z.infer<typeof TransportType>;

/**
 * Extended configuration schema with transport support
 */
const ConfigSchema = z.object({
  // ... existing fields (vikunjaApiUrl, redis, rateLimits, llm, logging) ...
  
  /**
   * MCP transport type
   * - stdio: Standard input/output (default, for subprocess communication)
   * - http: HTTP with Server-Sent Events (for remote clients)
   */
  transportType: TransportType.default('stdio'),
  
  /**
   * MCP server port (required for HTTP transport)
   * Blue environment: 3010
   * Green environment: 3011
   */
  mcpPort: z.number().int().positive().default(3457),
  
  /**
   * CORS configuration for HTTP transport
   */
  cors: z.object({
    enabled: z.boolean().default(false),
    allowedOrigins: z.array(z.string().url()).default([]),
  }).optional(),
});

export type Config = z.infer<typeof ConfigSchema>;

/**
 * Validate configuration with cross-field constraints
 */
export function validateConfig(config: Config): Config {
  // Require mcpPort when using HTTP transport
  if (config.transportType === 'http' && !config.mcpPort) {
    throw new Error(
      'Configuration error: MCP_PORT is required when TRANSPORT_TYPE=http'
    );
  }
  
  // Warn if CORS enabled without allowed origins
  if (config.cors?.enabled && config.cors.allowedOrigins.length === 0) {
    console.warn(
      'Warning: CORS enabled but no allowed origins configured. All origins will be denied.'
    );
  }
  
  return config;
}
```

**Environment Variable Mapping**:
- `TRANSPORT_TYPE` → `transportType` (values: "stdio" | "http")
- `MCP_PORT` → `mcpPort` (integer, required for HTTP)
- `CORS_ENABLED` → `cors.enabled` (boolean, optional)
- `CORS_ALLOWED_ORIGINS` → `cors.allowedOrigins` (comma-separated URLs, optional)

**Validation Rules**:
- `transportType` must be exactly "stdio" or "http"
- `mcpPort` must be positive integer (1-65535 range)
- HTTP transport MUST have `mcpPort` configured
- CORS origins must be valid URLs (if provided)

**Default Values**:
- `transportType`: "stdio" (backward compatibility)
- `mcpPort`: 3457 (overridden to 3010/3011 by deployment scripts)
- `cors.enabled`: false
- `cors.allowedOrigins`: [] (deny all cross-origin)

---

## Runtime State Models

### SSE Connection State

**Source File**: `mcp-server/src/transport/types.ts`

```typescript
import type { UserContext } from '../auth/types.js';
import type { Response } from 'express';

/**
 * Represents an active SSE connection
 */
export interface SSEConnection {
  /**
   * Unique connection identifier
   */
  id: string;
  
  /**
   * Authenticated user context
   */
  userContext: UserContext;
  
  /**
   * Express response object (for SSE streaming)
   */
  response: Response;
  
  /**
   * Connection establishment timestamp
   */
  connectedAt: Date;
  
  /**
   * Last activity timestamp (for idle detection)
   */
  lastActivityAt: Date;
  
  /**
   * Connection state
   */
  state: 'connected' | 'closing' | 'closed';
}

/**
 * SSE connection manager
 * Tracks all active connections for graceful shutdown
 */
export class SSEConnectionManager {
  private connections: Map<string, SSEConnection>;
  
  constructor() {
    this.connections = new Map();
  }
  
  /**
   * Add new connection
   */
  add(connection: SSEConnection): void {
    this.connections.set(connection.id, connection);
  }
  
  /**
   * Remove connection
   */
  remove(connectionId: string): void {
    this.connections.delete(connectionId);
  }
  
  /**
   * Get connection by ID
   */
  get(connectionId: string): SSEConnection | undefined {
    return this.connections.get(connectionId);
  }
  
  /**
   * Get all active connections
   */
  getAll(): SSEConnection[] {
    return Array.from(this.connections.values());
  }
  
  /**
   * Get connection count
   */
  count(): number {
    return this.connections.size;
  }
  
  /**
   * Gracefully close all connections
   */
  async closeAll(): Promise<void> {
    const closePromises = Array.from(this.connections.values()).map(
      async (conn) => {
        conn.state = 'closing';
        // Send close event via SSE
        conn.response.write('event: close\ndata: {"reason": "server shutdown"}\n\n');
        conn.response.end();
        conn.state = 'closed';
      }
    );
    
    await Promise.all(closePromises);
    this.connections.clear();
  }
}
```

**State Transitions**:
1. **Initial**: Connection request received at POST `/sse`
2. **Authenticating**: Middleware validates token
3. **Connected**: SSE transport created, bidirectional communication active
4. **Closing**: Server shutdown or client disconnect initiated
5. **Closed**: Connection terminated, resources released

**Lifecycle Management**:
- Connections stored in-memory (not persisted)
- Each connection has unique ID (UUID v4)
- Manager tracks connections for graceful shutdown on SIGTERM
- Idle connections (no activity >5 minutes) automatically cleaned up

---

## Authentication Context

**Source File**: `mcp-server/src/auth/types.ts` (EXISTING)

```typescript
/**
 * Authenticated user context (already exists)
 */
export interface UserContext {
  userId: number;
  username: string;
  token: string;
  permissions: UserPermissions;
  createdAt: Date;
}

/**
 * User permissions (already exists)
 */
export interface UserPermissions {
  isAdmin: boolean;
  canCreateProjects: boolean;
  canManageTeams: boolean;
}
```

**Usage in HTTP Transport**:
- Middleware extracts token from request (header or query param)
- `Authenticator.validateToken()` returns `UserContext`
- Context stored in `req.userContext` for SSE handler
- Context passed to MCP tool execution for permission checks

---

## Request/Response Models

### SSE Connection Request

**HTTP Method**: POST  
**Endpoint**: `/sse`  
**Headers**:
- `Authorization: Bearer <vikunja-api-token>` (preferred)
- `Content-Type: application/json` (optional)

**Query Parameters** (alternative auth):
- `token`: Vikunja API token (fallback if header not supported)

**Request Body**: None (SSE is initiated, not data-driven)

**Example Request**:
```http
POST /sse HTTP/1.1
Host: localhost:3010
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Accept: text/event-stream
```

### SSE Connection Response

**Success Response** (200 OK):
```http
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: connected
data: {"connectionId": "123e4567-e89b-12d3-a456-426614174000"}

```

**Error Responses**:

**401 Unauthorized** (missing/invalid token):
```json
{
  "error": "Unauthorized",
  "message": "Missing or invalid authentication token"
}
```

**503 Service Unavailable** (server shutting down):
```json
{
  "error": "Service Unavailable",
  "message": "Server is shutting down, please retry"
}
```

---

## State Diagrams

### Transport Initialization Flow

```
[Server Start]
      |
      v
[Load Config] --> TRANSPORT_TYPE?
      |                |
      |                v
      |         [stdio] --> [Create StdioServerTransport]
      |                              |
      |                              v
      |                     [Connect MCP Server]
      |                              |
      |         [http]  --> [Create Express App]
      |                              |
      |                              v
      |                     [Setup POST /sse Route]
      |                              |
      |                              v
      |                     [Start Express Server on MCP_PORT]
      |                              |
      +------------------------------+
      |
      v
[Server Ready]
```

### SSE Connection Lifecycle

```
[Client POST /sse]
      |
      v
[Auth Middleware] --> Token Valid?
      |                    |
      | YES                | NO
      v                    v
[Extract UserContext] [Return 401]
      |
      v
[Create SSE Transport]
      |
      v
[Connect MCP Server to Transport]
      |
      v
[Send 'connected' Event]
      |
      v
[Bidirectional Communication Active]
      |
      +---> [Client Request] --> [Execute MCP Tool] --> [Stream Response]
      |
      +---> [Client Disconnect] --> [Close Connection] --> [Remove from Manager]
      |
      +---> [Server Shutdown] --> [Send 'close' Event] --> [Gracefully Close]
```

---

## Relationships

### Component Dependencies

```
Config
  |
  +---> TransportFactory
          |
          +---> StdioServerTransport (if stdio)
          |
          +---> Express Server (if http)
                  |
                  +---> SSEAuthMiddleware
                  |       |
                  |       +---> Authenticator
                  |       |       |
                  |       |       +---> UserContext
                  |       |
                  |       +---> SSEConnectionHandler
                  |               |
                  |               +---> SSEServerTransport
                  |               |
                  |               +---> SSEConnectionManager
                  |                       |
                  |                       +---> SSEConnection[]
                  |
                  +---> MCP Server
```

### Data Flow

```
[Client] --POST /sse with token--> [Express Middleware]
                                           |
                                           v
                                    [Authenticator.validateToken()]
                                           |
                                           v
                                    [UserContext (cached 5min)]
                                           |
                                           v
                                    [SSE Connection Handler]
                                           |
                                           v
                                    [SSEServerTransport]
                                           |
                                           v
                                    [MCP Server.connect()]
                                           |
                                           v
                                    [Bidirectional SSE Stream]
                                           |
                                           v
                                    [MCP Tool Execution]
                                           |
                                           v
                                    [Response Streamed Back]
```

---

## Performance Considerations

### Memory Usage

| Component | Memory per Instance | Max Instances | Total Memory |
|-----------|---------------------|---------------|--------------|
| SSEConnection | ~10 KB | 50 connections | ~500 KB |
| UserContext (cached) | ~2 KB | 100 users | ~200 KB |
| Express Server | ~20 MB | 1 instance | ~20 MB |
| Total Overhead | | | **~21 MB** |

**Baseline**: Node.js MCP server uses ~50-80 MB, HTTP transport adds ~21 MB overhead.

### Caching Strategy

**Token Validation Cache**:
- **Implementation**: Existing `Authenticator` class with Redis backend
- **TTL**: 5 minutes
- **Key**: `mcp:auth:token:<sha256-hash>`
- **Value**: Serialized `UserContext`
- **Hit Rate**: Expected 80%+ (clients maintain persistent connections)

**Cache Invalidation**:
- Automatic expiry after 5 minutes
- Manual invalidation on password change/logout (future enhancement)

---

## Validation & Error Handling

### Configuration Validation

**Startup Checks**:
1. `TRANSPORT_TYPE` is valid enum value
2. `MCP_PORT` is provided and valid for HTTP transport
3. `VIKUNJA_API_URL` is reachable (existing check)
4. CORS origins are valid URLs (if provided)

**Error Messages**:
- ❌ `"Invalid TRANSPORT_TYPE: expected 'stdio' or 'http', got 'websocket'"`
- ❌ `"MCP_PORT required when TRANSPORT_TYPE=http"`
- ❌ `"Invalid CORS origin: must be valid URL"`

### Runtime Error Handling

**Authentication Errors**:
- 401 Unauthorized: Missing token, invalid token, expired token
- Logged at WARN level with user identifier (not token value)

**Connection Errors**:
- 503 Service Unavailable: Server shutting down, too many connections
- Logged at ERROR level with connection details

**MCP Protocol Errors**:
- Invalid tool name: Returned via SSE error event
- Tool execution failure: Logged at ERROR, returned to client with sanitized message

---

## Testing Validation

### Unit Test Coverage

**Config Validation**:
- ✅ Default values applied correctly
- ✅ HTTP transport requires MCP_PORT
- ✅ Invalid transport type rejected
- ✅ CORS validation

**SSE Connection Manager**:
- ✅ Add/remove connections
- ✅ Connection count tracking
- ✅ Graceful closeAll()

**Auth Middleware**:
- ✅ Token extraction from header
- ✅ Token extraction from query param
- ✅ Missing token returns 401
- ✅ Invalid token returns 401
- ✅ Valid token populates req.userContext

### Integration Test Scenarios

1. **Successful Connection**: POST with valid token → 200 OK with SSE stream
2. **Missing Token**: POST without auth → 401 Unauthorized
3. **Invalid Token**: POST with bad token → 401 Unauthorized
4. **Concurrent Connections**: 50 simultaneous connections → all succeed
5. **Graceful Shutdown**: SIGTERM during active connections → close events sent

---

## Summary

**Key Data Structures**:
- `Config` (extended with `transportType`, `mcpPort`, `cors`)
- `SSEConnection` (connection state tracking)
- `SSEConnectionManager` (centralized connection lifecycle)
- `UserContext` (existing, reused for authentication)

**State Management**:
- Configuration validated at startup (fail-fast)
- Connections tracked in-memory (ephemeral)
- Authentication cached in Redis (5-minute TTL)

**Validation Strategy**:
- Zod schema validation for config
- Express middleware for auth
- TypeScript types for compile-time safety
- Comprehensive unit + integration tests

This data model provides the foundation for implementing HTTP/SSE transport with proper type safety, state management, and error handling.
