# Research: AI Agent Control via MCP

**Feature**: 002-ai-agent-control  
**Date**: 2025-10-17  
**Status**: Phase 0 - Research Complete

## Executive Summary

This research validates the feasibility of implementing an MCP (Model Context Protocol) server for Vikunja with self-hosted Proxmox deployment, independent versioning, per-token rate limiting, and external LLM integration.

**Key Finding**: All requirements are technically feasible with established tools and patterns.

---

## 0.1 MCP Protocol Deep Dive

### Protocol Overview
**MCP Version**: v1.0+ (Anthropic specification)  
**Transport**: JSON-RPC 2.0 over stdio, SSE, or WebSocket  
**Best Choice**: SSE (Server-Sent Events) for agent connections

### Required Capabilities
1. **Resources**: MUST expose queryable data (projects, tasks, etc.)
2. **Tools**: MUST provide callable operations (create, update, delete)
3. **Prompts**: OPTIONAL workflow templates for agents

### Optional Capabilities
4. **Sampling**: Agent can request LLM completions (not needed - agents have their own LLM)
5. **Roots**: File system roots (not applicable)
6. **Logging**: Server-side logging (useful for debugging)

### Authentication in MCP
- MCP itself has NO built-in auth
- Custom implementation: Pass Vikunja API token in initial connection handshake
- Store token in connection context, use for all Vikunja API calls

### Resource URI Scheme
**Standard**: `protocol://domain/path`  
**Vikunja**: `vikunja://projects/123`, `vikunja://tasks/456/comments`

**Example Resource**:
```json
{
  "uri": "vikunja://tasks/456",
  "name": "Task: Fix bug in authentication",
  "mimeType": "application/json",
  "text": "{...task JSON...}"
}
```

### MCP Subscriptions
- **Not in v1.0**: Real-time updates not part of initial spec
- **Future**: Could implement via WebSocket transport
- **Workaround**: Agents poll resources, acceptable for v1

### Error Format
```json
{
  "jsonrpc": "2.0",
  "id": 123,
  "error": {
    "code": -32000,
    "message": "Permission denied",
    "data": {
      "vikunjaError": "User does not have write access to project",
      "projectId": 5
    }
  }
}
```

**Error Codes** (JSON-RPC standard):
- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error
- `-32000 to -32099`: Custom application errors

**Decision**: Use -32000 range for Vikunja-specific errors (permission denied, not found, validation failed)

---

## 0.2 Proxmox Deployment Architecture

### LXC vs VM Decision
**Winner**: **LXC Containers**

**Rationale**:
- Much lower overhead (MB RAM vs GB)
- Faster startup (<1s vs 10-30s)
- Better density (50+ containers per host vs 10-20 VMs)
- Native Node.js performance (no virtualization overhead)
- Easier to template and clone

**VM Use Case**: Only if strict kernel-level isolation needed (not required here)

### Multi-Version Architecture

**Strategy**: One LXC container per MCP server version

```
Proxmox Host
├── vikunja-mcp-v1.0 (LXC 100) → Port 9001
├── vikunja-mcp-v1.1 (LXC 101) → Port 9002
├── vikunja-mcp-v2.0 (LXC 102) → Port 9003
└── vikunja-redis (LXC 200) → Port 6379 (shared)
```

**Port Allocation**:
- Base port: 9000
- Version port: 9000 + (major * 100) + minor
- Examples: v1.0 → 9100, v1.1 → 9101, v2.0 → 9200

### Network Configuration

**Option A: Bridge Mode** (Recommended)
- Each LXC gets IP on main network (192.168.1.x)
- Direct access from agent machines
- Simple firewall rules
- Pros: Easy to understand, standard setup
- Cons: Uses IPs from main pool

**Option B: NAT with Port Forwarding**
- LXCs on private network (10.0.0.x)
- Port forward to Proxmox host IP
- Pros: IP conservation, centralized access control
- Cons: More complex routing, single point of entry

**Decision**: Bridge mode for simplicity, can switch to NAT later if needed

### Resource Allocation

**Per MCP Container**:
- **CPU**: 2 vCPUs (Node.js is single-threaded, but allows OS tasks)
- **RAM**: 2GB (1.5GB for Node, 512MB for OS)
- **Storage**: 8GB (2GB for Node modules, 6GB for logs/cache)
- **Network**: 1Gbps

**Scaling**: Add more containers (horizontal) rather than increasing resources (vertical)

**Redis Container**:
- CPU: 1 vCPU
- RAM: 1GB (512MB for Redis, 512MB for OS)
- Storage: 4GB

### Backup Strategy

**LXC Backup** (Proxmox built-in):
- Daily backup of container
- Retention: 7 days
- Backup storage: Proxmox Backup Server or NAS
- Backup includes: `/etc/vikunja-mcp/`, `/var/log/vikunja-mcp/`

**Configuration Backup** (Git):
- `/etc/vikunja-mcp/` → Git repository
- Commit on every config change
- Allows rollback and audit trail

**Redis Backup**:
- RDB snapshots every 15 minutes
- AOF (Append-Only File) for durability
- Backup to NAS nightly

### Update Process (Zero Downtime)

1. **Deploy new version** in new LXC container
2. **Test** with single agent
3. **Update DNS/load balancer** to point to new container
4. **Monitor** for issues
5. **Decommission old container** after 7 days (or keep for rollback)

**Rollback**: Point DNS back to old container

---

## 0.3 Rate Limiting Strategy

### Redis Architecture

**Deployment**: Single Redis instance in LXC container

**Why Not Cluster?**:
- Single Redis handles 100K+ ops/sec (more than enough)
- Cluster adds complexity without benefit at this scale
- Easier to backup and restore
- Lower operational overhead

**When to Cluster**: If MCP server scales to 10+ instances or 1M+ requests/minute

### Rate Limit Algorithm

**Choice**: **Sliding Window Counter** (Redis-based)

**Why**:
- Smooth rate limiting (no burst at window boundaries)
- Accurate counting within time window
- Efficient (2 Redis operations per request)
- Better than Fixed Window (avoids boundary gaming)

**Algorithm**:
```typescript
// Key: ratelimit:{token}
// Value: Sorted set of timestamps

const now = Date.now();
const windowStart = now - 60000; // 1 minute ago

// Remove old entries
await redis.zremrangebyscore(key, 0, windowStart);

// Count entries in window
const count = await redis.zcard(key);

if (count >= 100) {
  throw new RateLimitError('Limit exceeded');
}

// Add new entry
await redis.zadd(key, now, `${now}-${randomId()}`);
await redis.expire(key, 120); // Expire after 2 minutes
```

**Memory**: ~100 bytes per request * 100 requests = 10KB per token per minute

### Rate Limit Configuration

**Per-Token Limits**:
- **Default**: 100 requests/minute
- **Burst**: 120 requests (20% burst allowance)
- **Admin tokens**: Unlimited (bypass rate limiting)

**Error Response**:
```json
{
  "error": {
    "code": -32000,
    "message": "Rate limit exceeded",
    "data": {
      "limit": 100,
      "remaining": 0,
      "resetAt": "2025-10-17T10:05:00Z"
    }
  }
}
```

### Token Cleanup

**Strategy**: Expire keys automatically with Redis TTL

- Active tokens: TTL refreshed on each request (2 minutes)
- Inactive tokens: Auto-expire after 2 minutes of no activity
- No manual cleanup needed

**Storage Estimate**:
- 1000 active tokens * 10KB = 10MB RAM
- 10,000 tokens * 10KB = 100MB RAM (still tiny)

---

## 0.4 External LLM Integration

### LLM Provider Selection

**Primary**: **OpenAI GPT-4 Turbo**
- Most reliable for task parsing
- Good balance of speed and accuracy
- Well-documented API
- $10/1M tokens input, $30/1M tokens output

**Secondary**: **Anthropic Claude 3 Sonnet**
- Excellent at structured extraction
- Alternative if OpenAI unavailable
- Similar pricing

**Self-Hosted**: **Ollama + Llama 3.1**
- For air-gapped deployments
- Free but requires GPU (RTX 3090 or better)
- Lower accuracy than commercial models

### Provider Abstraction

**Interface**:
```typescript
interface LLMProvider {
  parseTask(input: string): Promise<ParsedTask>;
  isAvailable(): Promise<boolean>;
}

interface ParsedTask {
  title: string;
  description?: string;
  dueDate?: Date;
  priority?: number;
  labels?: string[];
  assignees?: string[];
}
```

**Providers**:
```typescript
class OpenAIProvider implements LLMProvider { }
class AnthropicProvider implements LLMProvider { }
class OllamaProvider implements LLMProvider { }
```

### Prompt Structure

**System Prompt**:
```
You are a task parser for Vikunja. Extract structured task information from natural language.

Output JSON with these fields (omit if not present):
- title (required): Concise task title
- description (optional): Detailed description
- dueDate (optional): ISO 8601 date
- priority (optional): 1=low, 2=medium, 3=high, 4=urgent, 5=critical
- labels (optional): Array of label names
- assignees (optional): Array of usernames with @ prefix

Examples:
Input: "Fix authentication bug by Friday @alice high priority"
Output: {"title":"Fix authentication bug","dueDate":"2025-10-20","priority":3,"assignees":["alice"]}

Input: "Write documentation #docs"
Output: {"title":"Write documentation","labels":["docs"]}
```

**User Message**: The natural language task input

### Fallback Behavior

**If LLM unavailable**:
1. Return error to agent: "Natural language parsing unavailable"
2. Agent can retry with structured input (JSON)
3. Log failure for monitoring

**If LLM returns invalid JSON**:
1. Use regex fallback for simple parsing
2. Extract title (required) at minimum
3. Log parse failure

### Caching Strategy

**Cache Key**: Hash of input text  
**Cache TTL**: 5 minutes  
**Storage**: Redis (same instance as rate limiting)

**Why Cache**:
- Agents may retry same input
- Reduces LLM API costs
- Faster response time

**Cache Size**: Limit to 1000 entries (LRU eviction)

### Cost Estimation

**Assumptions**:
- Average task input: 50 tokens
- Average output: 100 tokens
- 1000 tasks parsed per day

**Cost** (GPT-4 Turbo):
- Input: 1000 * 50 * $0.00001 = $0.50/day
- Output: 1000 * 100 * $0.00003 = $3.00/day
- **Total**: ~$3.50/day or $105/month

**With 50% cache hit rate**: $52.50/month

---

## 0.5 Vikunja API Integration

### Required Endpoints

**Authentication**:
- `POST /api/v1/login` - Not used (agents provide token directly)
- Token validation: Implicit in all requests

**Projects**:
- `GET /api/v1/projects` - List user's projects
- `GET /api/v1/projects/:id` - Get project details
- `POST /api/v1/projects` - Create project (uses PUT in v1!)
- `POST /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project

**Tasks**:
- `GET /api/v1/tasks/all` - List all user tasks
- `GET /api/v1/projects/:id/tasks` - List project tasks
- `GET /api/v1/tasks/:id` - Get task details
- `PUT /api/v1/projects/:id/tasks` - Create task
- `POST /api/v1/tasks/:id` - Update task
- `DELETE /api/v1/tasks/:id` - Delete task
- `POST /api/v1/tasks/bulk` - Bulk update tasks

**Labels**:
- `GET /api/v1/labels` - List labels
- `PUT /api/v1/labels` - Create label
- `PUT /api/v1/tasks/:id/labels` - Add label to task
- `DELETE /api/v1/tasks/:id/labels/:labelId` - Remove label

**Teams**:
- `GET /api/v1/teams` - List teams
- `GET /api/v1/teams/:id` - Get team details

**Users**:
- `GET /api/v1/user` - Get current user
- `GET /api/v1/users` - Search users

### Authentication Flow

1. **Agent connection**: Provides Vikunja API token in handshake
2. **MCP server**: Validates token with `GET /api/v1/user`
3. **If valid**: Store token in connection context
4. **All requests**: Use token for Vikunja API calls

**Token Storage**: In-memory Map keyed by connection ID

### Error Mapping

**Vikunja Error → MCP Error**:
- 400 Bad Request → -32602 Invalid params
- 401 Unauthorized → -32000 Authentication failed
- 403 Forbidden → -32000 Permission denied
- 404 Not Found → -32000 Resource not found
- 500 Internal Error → -32603 Internal error

**Example**:
```typescript
function mapVikunjaError(apiError: VikunjaError): MCPError {
  if (apiError.status === 403) {
    return {
      code: -32000,
      message: 'Permission denied',
      data: { vikunjaError: apiError.message }
    };
  }
  // ... more mappings
}
```

### Pagination Handling

**Vikunja Pagination**:
- Query params: `page=1&per_page=50`
- Response headers: `x-pagination-total-pages`, `x-pagination-result-count`

**MCP Resource Listing**:
- Default page size: 50
- Max page size: 100
- Return pagination metadata in resource response

### Permission Checking Strategy

**Decision**: **Rely on Vikunja API for all permission checks**

**Why**:
- Vikunja already has complex permission model (users, teams, link shares)
- Reduces duplication and potential security bugs
- Service layer (already refactored) handles permissions correctly
- MCP server just proxies - no business logic

**Approach**:
1. MCP server calls Vikunja API with user token
2. If Vikunja returns 403, map to MCP permission denied error
3. No permission caching in MCP server (keep it simple)

**Performance**: Acceptable since service layer is fast (<50ms)

---

## 0.6 Version Management Strategy

### Semantic Versioning

**Format**: `vMAJOR.MINOR.PATCH`

**Increment Rules**:
- **MAJOR**: Breaking changes to resources, tools, or authentication
  - Example: Change tool input schema, remove resource type
- **MINOR**: Backwards-compatible additions
  - Example: Add new tool, add optional field to resource
- **PATCH**: Bug fixes, no API changes
  - Example: Fix error handling, improve performance

### Version Discovery

**MCP Server Info**:
```json
{
  "name": "vikunja-mcp-server",
  "version": "1.0.0",
  "protocolVersion": "2024-11-05",
  "capabilities": {
    "resources": {},
    "tools": {},
    "prompts": {}
  }
}
```

Agents get this on connection via `initialize` request.

### Simultaneous Version Support

**Architecture**: Independent LXC containers per major version

```
v1.x servers (LXC 100-199)
v2.x servers (LXC 200-299)
```

**Configuration Isolation**:
- Each version has own config directory: `/etc/vikunja-mcp/v{major}/`
- Separate log directories: `/var/log/vikunja-mcp/v{major}/`
- Separate systemd services: `vikunja-mcp-v1@.service`, `vikunja-mcp-v2@.service`

**Shared Resources**:
- Redis (rate limiting) - shared across versions
- Vikunja API - same backend for all versions

### Configuration Differences

**v1.x config** (`/etc/vikunja-mcp/v1/config.json`):
```json
{
  "version": "1.0.0",
  "vikunjaApiUrl": "http://vikunja:3456/api/v1",
  "port": 9100,
  "redis": { "host": "redis.lxc", "port": 6379 },
  "rateLimits": { "default": 100 }
}
```

**v2.x config** (future - example of breaking changes):
```json
{
  "version": "2.0.0",
  "vikunjaApiUrl": "http://vikunja:3456/api/v1",
  "port": 9200,
  "redis": { "host": "redis.lxc", "port": 6379 },
  "rateLimits": { 
    "default": 200,  // Different default
    "perTool": true   // New feature
  },
  "auth": {
    "type": "oauth2"  // Breaking change from token auth
  }
}
```

### Agent Migration Path

**v1 → v2 migration**:

1. **Agent discovers v2**: Checks Vikunja documentation or attempts connection
2. **Agent tests v2**: Connects to v2 port, validates compatibility
3. **Agent switches**: Updates connection string from v1 to v2
4. **Gradual rollout**: Agents migrate over weeks/months
5. **v1 deprecation**: After 6 months, v1 marked deprecated
6. **v1 shutdown**: After 12 months, v1 containers removed

**Compatibility Layer** (optional):
- v2 server could include v1 compatibility mode
- Helps agents migrate gradually
- Not required but nice to have

### Version Support Policy

**Active Support**:
- Current major version: Full support (features + fixes)
- Previous major version: Security fixes only for 6 months

**End of Life**:
- 6 months after new major version release
- 3-month warning before shutdown
- Migration guide provided

**Example Timeline**:
- 2025-11-01: v1.0 released
- 2026-05-01: v2.0 released
- 2026-05-01 → 2026-11-01: v1.x receives security fixes only
- 2026-08-01: Warning email to v1 users
- 2026-11-01: v1.x EOL, containers shut down

---

## Architectural Decisions

### AD-001: SSE Transport for MCP
**Decision**: Use Server-Sent Events (SSE) as primary transport  
**Rationale**: Better than stdio (easier to deploy), simpler than WebSocket  
**Alternative**: WebSocket if bidirectional push needed later

### AD-002: LXC over VM
**Decision**: Deploy MCP server in LXC containers  
**Rationale**: Lower overhead, faster startup, better density  
**Trade-off**: Less isolation than VMs (acceptable for this use case)

### AD-003: Single Redis Instance
**Decision**: Use single Redis for rate limiting and caching  
**Rationale**: Sufficient performance, simpler operations  
**Threshold**: Cluster if >10 MCP instances or >1M req/min

### AD-004: OpenAI Primary LLM
**Decision**: Use OpenAI GPT-4 Turbo for task parsing  
**Rationale**: Best accuracy, good speed, well-documented  
**Alternatives**: Anthropic (fallback), Ollama (self-hosted)

### AD-005: Permission Delegation
**Decision**: Delegate all permission checks to Vikunja API  
**Rationale**: Avoid duplicating complex permission logic  
**Trade-off**: Slight latency increase (50ms per check) vs security correctness

### AD-006: Version Isolation
**Decision**: Separate LXC per major version  
**Rationale**: Clean isolation, independent upgrades  
**Alternative**: Multi-version in single container (rejected - too complex)

---

## Open Questions & Resolutions

### Q1: How to handle WebSocket vs SSE for MCP?
**Resolution**: Start with SSE (simpler), add WebSocket if bidirectional push needed

### Q2: Should MCP server cache Vikunja API responses?
**Resolution**: No for v1.0 - keep it simple, add caching in v1.1 if needed

### Q3: How to monitor MCP server health?
**Resolution**: Expose `/health` endpoint returning JSON with status, metrics

### Q4: Should we support API v2 endpoints when they exist?
**Resolution**: No - MCP uses v1 endpoints, v2 is separate track (may be cancelled)

### Q5: How to handle Vikunja backend downtime?
**Resolution**: Circuit breaker pattern - stop calling Vikunja after 10 consecutive failures, retry after 30 seconds

### Q6: Rate limiting granularity - per tool or global?
**Resolution**: Global per token for v1.0, can add per-tool limits in v2.0 if needed

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| MCP spec changes | Low | High | Abstract protocol layer, monitor spec updates |
| OpenAI API outage | Medium | Medium | Anthropic fallback, cache recent parses |
| Redis failure | Low | High | Redis persistence (AOF), automatic restart |
| Vikunja API changes | Low | Medium | Version lock Vikunja API, test before upgrade |
| Container overhead | Low | Low | LXC is lightweight, monitored resource usage |
| Rate limit evasion | Medium | Medium | Token-based (not IP), admin review of high-usage |

---

## Technology Stack Summary

### MCP Server
- **Language**: TypeScript 5.x
- **Runtime**: Node.js 20 LTS
- **Framework**: `@modelcontextprotocol/sdk`
- **HTTP**: Express (health checks only)
- **Validation**: Zod
- **Logging**: Winston
- **Testing**: Vitest + Supertest

### Infrastructure
- **Virtualization**: Proxmox VE 8.x
- **Containers**: LXC (Debian 12 base)
- **Cache/Rate Limit**: Redis 7.x
- **Reverse Proxy**: Nginx (optional)
- **Monitoring**: Node exporter + Prometheus (optional)

### External Services
- **Vikunja API**: Existing v1 endpoints
- **LLM**: OpenAI GPT-4 Turbo / Anthropic Claude 3 / Ollama

---

## Next Steps

1. ✅ Research complete - All questions answered
2. **Phase 1**: Create data model and contracts
3. **Phase 1**: Generate quickstart guide
4. **Phase 2**: `/tasks` command to generate implementation tasks
5. **Phase 3**: Begin TDD implementation

---

**Research Status**: ✅ **COMPLETE**  
**Blockers**: None  
**Ready for Phase 1**: Yes
