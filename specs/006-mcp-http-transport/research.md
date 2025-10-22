# Research: MCP HTTP/SSE Transport Implementation

**Feature**: MCP HTTP/SSE Transport  
**Phase**: 0 - Research & Technical Discovery  
**Date**: 2025-10-22

## Overview

This document consolidates research findings for implementing HTTP/SSE transport in the Vikunja MCP server. The goal is to resolve all technical unknowns before proceeding to design and implementation.

## Research Areas

### 1. MCP SDK SSE Transport Patterns

**Research Task**: Understand how `@modelcontextprotocol/sdk/server/sse.js` works and integration patterns

**Decision**: Use MCP SDK's built-in `SSEServerTransport` class with Express server integration

**Rationale**:
- MCP SDK provides first-party SSE transport implementation following the official MCP specification
- Designed for HTTP POST-based SSE connections with standardized message format
- Handles protocol-level concerns (message framing, error handling, connection lifecycle)
- Well-documented in MCP SDK repository with reference implementations

**Integration Pattern**:
```typescript
import { SSEServerTransport } from '@modelcontextprotocol/sdk/server/sse.js';
import express from 'express';

// Create Express app for SSE endpoint
const app = express();

// SSE endpoint accepts POST requests
app.post('/sse', async (req, res) => {
  // 1. Authenticate request (extract token from header/query)
  // 2. Create SSE transport instance
  const transport = new SSEServerTransport('/sse', res);
  // 3. Connect MCP server to transport
  await server.connect(transport);
});

// Listen on MCP_PORT
app.listen(port);
```

**Alternatives Considered**:
- **Custom SSE implementation**: Rejected - reinventing protocol handling is error-prone and MCP SDK already provides this
- **WebSocket transport**: Rejected - out of scope per specification, SSE is sufficient for server-to-client streaming with HTTP POST for client-to-server
- **Socket.io**: Rejected - adds unnecessary dependency when MCP SDK has native SSE support

**Key Findings**:
- SSE transport requires separate Express instance from health check server (different ports: `MCP_PORT` for SSE, existing port for health checks)
- Each POST request to `/sse` establishes a new SSE connection
- Authentication must happen BEFORE creating SSE transport (return 401 without establishing connection for invalid tokens)
- Transport lifecycle: POST request → auth → create transport → connect server → keep-alive until client closes or server shuts down

**References**:
- MCP SDK documentation: https://github.com/modelcontextprotocol/sdk
- SSE specification: https://html.spec.whatwg.org/multipage/server-sent-events.html
- Existing MCP server stdio transport: `mcp-server/src/server.ts` lines 195-200

---

### 2. Express Server Integration for Health Checks + SSE

**Research Task**: Determine if SSE transport can coexist with existing Express health check server

**Decision**: Use **separate Express server instances** for health checks and SSE transport

**Rationale**:
- Health check server (port 3457) is independent lifecycle - must remain available even if MCP transport fails
- SSE transport binds to `MCP_PORT` (3010 blue, 3011 green) which is environment-specific
- Separation of concerns: health checks are operational monitoring, SSE is MCP protocol transport
- Avoids port conflicts and simplifies blue-green deployment (each environment has dedicated MCP port)

**Implementation Pattern**:
```typescript
// Existing health check server (unchanged)
const healthApp = express();
healthApp.get('/health', healthCheckHandler);
healthApp.listen(3457);

// NEW: SSE transport server (only when TRANSPORT_TYPE=http)
if (config.transportType === 'http') {
  const sseApp = express();
  sseApp.post('/sse', sseAuthMiddleware, sseConnectionHandler);
  sseApp.listen(config.mcpPort);
}
```

**Alternatives Considered**:
- **Single Express server with multiple endpoints**: Rejected - couples health checks to MCP transport lifecycle, makes blue-green port management complex
- **Reuse health check server for SSE**: Rejected - different port requirements (health on 3457, MCP on 3010/3011)

**Key Findings**:
- Express servers are lightweight - multiple instances acceptable for separation of concerns
- Health check server starts regardless of transport type
- SSE server only starts when `TRANSPORT_TYPE=http`
- Port conflicts avoided by design (health: 3457, MCP blue: 3010, MCP green: 3011)

---

### 3. Per-Request Authentication Middleware

**Research Task**: Design authentication middleware for SSE endpoint that validates tokens before establishing connection

**Decision**: Create Express middleware that validates Vikunja API token using existing `Authenticator` class

**Rationale**:
- Existing `Authenticator.validateToken()` already implements token validation with 5-minute cache
- Express middleware pattern is idiomatic for pre-request validation
- Returning 401 before SSE connection prevents unauthorized clients from consuming server resources
- Token can come from `Authorization: Bearer <token>` header OR `?token=<token>` query parameter for client flexibility

**Middleware Implementation**:
```typescript
async function sseAuthMiddleware(
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
) {
  try {
    // Extract token from header or query
    const token = req.headers.authorization?.replace('Bearer ', '') 
                  || req.query.token as string;
    
    if (!token) {
      res.status(401).json({ error: 'Missing authentication token' });
      return;
    }

    // Validate with existing Authenticator (cached)
    const userContext = await authenticator.validateToken(token);
    
    // Store in request for SSE handler
    req.userContext = userContext;
    next();
  } catch (error) {
    res.status(401).json({ 
      error: 'Invalid token',
      message: error.message 
    });
  }
}
```

**Alternatives Considered**:
- **Server-level authentication**: Rejected - requires single shared token, eliminates per-user rate limiting and permissions
- **Authentication after SSE connection**: Rejected - wastes resources establishing connection for unauthorized clients
- **OAuth2 flow**: Rejected - Vikunja already uses API tokens, no need to introduce OAuth complexity

**Key Findings**:
- `Authenticator` class already handles caching (5-minute TTL) - no additional caching needed
- Middleware pattern integrates cleanly with Express routing
- Token extraction supports both header (preferred) and query param (fallback for clients that can't set headers)
- User context stored in request object passes to SSE connection handler
- Failed auth returns 401 immediately, preventing SSE handshake

---

### 4. Transport Factory Pattern

**Research Task**: Design clean abstraction for selecting stdio vs HTTP transport at runtime

**Decision**: Factory function that returns appropriate transport based on config

**Rationale**:
- Factory pattern encapsulates transport creation logic
- Server code doesn't need to know transport implementation details
- Easy to add future transports (if needed) without changing server code
- Configuration-driven selection aligns with 12-factor app principles

**Factory Implementation**:
```typescript
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { SSEServerTransport } from '@modelcontextprotocol/sdk/server/sse.js';
import type { Transport } from '@modelcontextprotocol/sdk/shared/transport.js';

export async function createTransport(config: Config): Promise<Transport> {
  if (config.transportType === 'stdio') {
    return new StdioServerTransport();
  } else if (config.transportType === 'http') {
    // HTTP transport requires Express server setup (separate flow)
    throw new Error('HTTP transport requires server.startHttpTransport()');
  } else {
    throw new Error(`Unsupported transport type: ${config.transportType}`);
  }
}
```

**Alternatives Considered**:
- **Strategy pattern with classes**: Rejected - overkill for two transport types, factory is simpler
- **Direct conditional in server.ts**: Rejected - couples server to transport details, harder to test
- **Plugin system**: Rejected - unnecessary complexity for built-in transports

**Key Findings**:
- Stdio transport is synchronous (just instantiate)
- HTTP transport requires async setup (Express server initialization)
- Factory validates transport type and provides clear error messages
- Server code uses factory result without knowing transport details

---

### 5. Configuration Schema Extension

**Research Task**: How to add `TRANSPORT_TYPE` to existing Zod config schema

**Decision**: Add enum field with validation, default to "stdio" for backward compatibility

**Config Schema Update**:
```typescript
const ConfigSchema = z.object({
  // ... existing fields ...
  transportType: z.enum(['stdio', 'http']).default('stdio'),
  mcpPort: z.number().int().positive().default(3457).optional(),
  // mcpPort required when transportType=http (validated separately)
});

// Custom validation
function validateConfig(config: z.infer<typeof ConfigSchema>) {
  if (config.transportType === 'http' && !config.mcpPort) {
    throw new Error('MCP_PORT is required when TRANSPORT_TYPE=http');
  }
  return config;
}
```

**Rationale**:
- Zod enum provides type safety and validation
- Default "stdio" maintains backward compatibility
- Cross-field validation ensures HTTP transport has required port
- Environment variable mapping: `TRANSPORT_TYPE` → `transportType`

**Alternatives Considered**:
- **Boolean flag `useHttp`**: Rejected - enum is more extensible if future transports added
- **Separate config files**: Rejected - environment variables are preferred for 12-factor apps
- **Auto-detect based on environment**: Rejected - explicit configuration is clearer

**Key Findings**:
- Zod validation happens at startup - fails fast if misconfigured
- TypeScript types derived from Zod schema ensure compile-time safety
- `.env.example` needs updates with `TRANSPORT_TYPE=stdio` and `TRANSPORT_TYPE=http` examples

---

### 6. Blue-Green Deployment Considerations

**Research Task**: How to preserve transport configuration across blue-green switches

**Decision**: systemd environment variables automatically preserved by systemd service manager

**Rationale**:
- systemd service files in `/etc/systemd/system/` persist across service restarts
- Blue and green services have separate unit files with their own environment variables
- `systemctl daemon-reload` and `systemctl restart` preserve `Environment=` directives
- No additional script logic needed - systemd handles this

**Deployment Flow**:
1. `vikunja-install.sh`: Generates `vikunja-mcp-blue.service` with `Environment="TRANSPORT_TYPE=http"`
2. Blue environment runs with HTTP transport
3. `vikunja-update.sh`: Generates `vikunja-mcp-green.service` with same `Environment="TRANSPORT_TYPE=http"`
4. Green environment starts with HTTP transport
5. Nginx switches to green - blue keeps HTTP config for next cycle

**Validation Check** (add to `vikunja-update-main.sh`):
```bash
# Before switching nginx, verify MCP HTTP endpoint is responding
if ! curl -f -s -o /dev/null "http://localhost:${MCP_GREEN_PORT}/sse"; then
    log_error "MCP HTTP endpoint not responding on green"
    exit 1
fi
```

**Alternatives Considered**:
- **Config file persistence**: Rejected - environment variables are simpler for container deployments
- **Database-stored config**: Rejected - unnecessary complexity, config is deployment-time concern
- **Script-based config migration**: Rejected - systemd already handles this

**Key Findings**:
- systemd `Environment=` directives persist in unit files
- Each color (blue/green) has independent service file
- Health check addition required to verify HTTP transport before nginx switch
- No config migration code needed

---

### 7. Proxmox Deployment Script Updates

**Research Task**: Identify exact changes needed in deployment scripts

**Decision**: Minimal changes to three scripts - add environment variable, health check, and summary

**Files to Modify**:

**1. `deploy/proxmox/lib/service-setup.sh` (lines 93-127)**
```bash
# Add after line 117 (existing Environment="MCP_PORT=...")
Environment="TRANSPORT_TYPE=http"
# Comment: "HTTP transport for remote client connectivity (n8n, Python)"
```

**2. `deploy/proxmox/lib/vikunja-install-main.sh`**
```bash
# Add after MCP service start
log_info "Waiting for MCP HTTP endpoint..."
for i in {1..30}; do
    if curl -f -s -o /dev/null "http://localhost:${MCP_BLUE_PORT}/health"; then
        log_success "MCP HTTP endpoint ready"
        break
    fi
    sleep 1
done

# Add to deployment summary
echo "MCP Server (HTTP): http://${CONTAINER_IP}:${MCP_BLUE_PORT}/sse"
echo "  Authentication: Bearer token in Authorization header"
echo "  Example (n8n): POST http://IP:${MCP_BLUE_PORT}/sse with header 'Authorization: Bearer <token>'"
```

**3. `deploy/proxmox/lib/vikunja-update-main.sh`**
```bash
# Add before nginx switch
log_info "Validating MCP HTTP transport on green..."
if ! curl -f -s -o /dev/null "http://localhost:${MCP_GREEN_PORT}/health"; then
    log_error "MCP HTTP endpoint not responding"
    exit 1
fi
```

**Rationale**:
- Single environment variable addition is non-invasive
- Health checks align with existing backend/frontend checks
- Deployment summary informs operators how to connect clients
- No changes to blue-green switching logic (transport is stateless)

**Alternatives Considered**:
- **Separate LXC container for MCP**: Deferred - evaluate in future spec if resource isolation needed
- **Config file generation**: Rejected - environment variables are simpler and align with systemd patterns
- **Interactive prompts for transport type**: Rejected - HTTP should be default for Proxmox, stdio for manual installs

**Key Findings**:
- Health check endpoint: `/health` (not `/sse`) - SSE endpoint doesn't respond to GET
- Summary should show full connection example with token placeholder
- Scripts already have retry logic for health checks - reuse pattern
- MCP ports already defined: `MCP_BLUE_PORT=3010`, `MCP_GREEN_PORT=3011`

---

## Summary of Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **SSE Transport** | Use MCP SDK `SSEServerTransport` | First-party implementation, spec-compliant |
| **Server Architecture** | Separate Express servers for health + SSE | Independent lifecycles, different ports |
| **Authentication** | Express middleware with existing `Authenticator` | Pre-connection validation, 5-min cache |
| **Transport Selection** | Factory function based on `TRANSPORT_TYPE` env var | Clean abstraction, config-driven |
| **Configuration** | Zod enum with cross-field validation | Type-safe, fail-fast on misconfiguration |
| **Blue-Green** | systemd preserves `Environment=` directives | No custom migration logic needed |
| **Deployment Scripts** | Minimal changes (env var, health check, summary) | Non-invasive, aligns with existing patterns |

## Next Steps (Phase 1: Design)

1. **Data Model**: Define TypeScript interfaces for transport config, SSE connection state, authentication context
2. **API Contracts**: Document POST `/sse` endpoint (OpenAPI spec)
3. **Quickstart Guide**: Write connection examples for n8n and Python MCP SDK clients
4. **Update Agent Context**: Run `.specify/scripts/bash/update-agent-context.sh copilot` to add TypeScript/Node.js/Express to project context

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| SSE compatibility with old MCP SDK versions | Medium | Test with multiple SDK versions, document minimum version |
| Performance impact of per-request auth | Low | Existing 5-min cache reduces overhead to <10ms |
| Blue-green switchover drops active connections | Low | Document client retry logic, SSE reconnect is standard pattern |
| Misconfigured transport type breaks deployment | High | Fail-fast validation, health checks catch before nginx switch |

All research complete - no remaining NEEDS CLARIFICATION markers.
