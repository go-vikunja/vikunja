# Feature Specification: AI Agent Control Layer - MCP vs API v2 Analysis

**Feature ID**: 002  
**Created**: 2025-10-17  
**Status**: Analysis & Planning  
**Priority**: Strategic  

## Executive Summary

This specification analyzes two competing approaches for enabling full AI agent control of Vikunja:
1. **API v2** - RESTful modernization with HATEOAS
2. **MCP (Model Context Protocol)** - Direct agent integration protocol

**Recommendation**: **Implement MCP directly**, skip API v2 as intermediate step.

## Context & Problem Statement

### Current State
- **API v1**: Inconsistent RESTful patterns, mixed HTTP verbs, working but not agent-optimized
- **V2 Progress**: ~20% complete (basic project endpoints implemented, frontend not migrated)
- **Agent Needs**: Programmatic task management, workflow automation, context-aware operations

### The Question
Should we:
- **Path A**: Complete API v2 → Then build MCP wrapper
- **Path B**: Build MCP directly on service layer (skip v2)

## Analysis: API v2 vs MCP

### API v2 Characteristics

**Goals** (from API_V2_PRD.md):
- RESTful standardization (plural nouns, semantic HTTP verbs)
- HATEOAS for discoverability
- Feature parity with v1
- Better documentation

**Progress**:
- ✅ 5 project endpoints (GET/POST/PUT/DELETE/duplicate)
- ❌ 0 frontend migrations completed
- ❌ 95% of endpoints not implemented
- ❌ No HATEOAS links implemented yet

**Estimated Effort**: 6-8 weeks
- Implement ~60 remaining endpoints
- Frontend client migration for all endpoints
- HATEOAS link implementation
- Comprehensive testing
- Documentation updates

**Value for Agents**: Moderate
- Better API consistency helps
- HATEOAS helps with discoverability
- Still HTTP/REST overhead
- No agent-specific optimizations

### MCP (Model Context Protocol) Characteristics

**What is MCP**:
- Protocol designed by Anthropic for AI agent integration
- Direct tool/resource exposure to LLMs
- Native context management
- Bidirectional communication
- Standard across agent platforms (Claude, Gemini, etc.)

**Architecture**:
```
AI Agent → MCP Server → Service Layer → Database
                ↓
          Resources/Tools/Prompts
```

**MCP Server Capabilities**:
1. **Resources**: Expose Vikunja data (projects, tasks, users)
2. **Tools**: Expose actions (create task, assign user, update status)
3. **Prompts**: Pre-defined agent workflows
4. **Sampling**: Agent can request LLM completions

**Value for Agents**: **High**
- Native agent protocol (no HTTP overhead)
- Context-aware operations
- Batch operations support
- Real-time updates via subscriptions
- Agent-optimized error handling
- Works across all MCP-compatible agents

**Estimated Effort**: 3-4 weeks
- MCP server implementation (TypeScript/Python)
- Service layer integration (already complete!)
- Resource/tool definitions (~30 operations)
- Authentication bridge
- Testing & documentation

## Decision Matrix

| Criteria | API v2 First | MCP Direct | Weight |
|----------|--------------|------------|--------|
| **Time to Agent Control** | 8-10 weeks | 3-4 weeks | ⭐⭐⭐ |
| **Agent UX** | Good (REST) | Excellent (native) | ⭐⭐⭐ |
| **Human Dev UX** | Excellent (REST) | N/A (agent-only) | ⭐ |
| **Maintenance Burden** | High (3 APIs) | Medium (2 APIs + MCP) | ⭐⭐ |
| **Future-Proofing** | Questionable | Protocol standard | ⭐⭐⭐ |
| **Constitutional Compliance** | Full | Full | ⭐⭐⭐ |
| **Reuse of Service Layer** | Yes | Yes | ⭐⭐ |

**Score**: MCP Direct wins 21-15

## Recommendation: MCP Direct

### Rationale

1. **Speed**: 4-6 weeks faster to agent control
2. **Purpose-Built**: MCP designed for this exact use case
3. **Service Layer Ready**: Current refactor provides perfect foundation
4. **Standard Protocol**: Works with Claude, Gemini, ChatGPT, etc.
5. **Lower Maintenance**: Don't maintain 3 parallel APIs (v1, v2, MCP)

### What About Human Developers?

**Keep API v1** for human developers:
- Already works
- Frontend uses it
- Third-party integrations exist
- Documentation exists

**MCP is for agents only**:
- Different use case
- Different consumers
- Different requirements

### Architecture Vision

```
Human Users → Frontend → API v1 → Service Layer → Database
                                      ↑
AI Agents → MCP Server ────────────────┘
```

## Requirements

### Functional Requirements

#### FR-001: MCP Server Implementation
- **MUST** implement Model Context Protocol specification v1.0+
- **MUST** expose Vikunja resources (projects, tasks, labels, teams, users)
- **MUST** expose Vikunja tools (CRUD operations, assignments, status changes)
- **MUST** support authentication via Vikunja API tokens
- **MUST** handle concurrent agent requests

#### FR-002: Resource Exposure
- **MUST** expose all major entities as MCP resources:
  - Projects (with hierarchy and views)
  - Tasks (with full metadata: assignees, labels, dates, attachments)
  - Labels (with color and usage info)
  - Teams (with members and permissions)
  - Users (filtered by permissions)
  - Comments (on tasks)
  - Attachments (with metadata)
  
#### FR-003: Tool Exposure
- **MUST** provide MCP tools for common operations:
  - Project management (create, update, archive, duplicate)
  - Task management (create, update, complete, delete, move)
  - Assignment operations (assign users, assign to buckets)
  - Label operations (create, assign, remove)
  - Comment operations (add, read, update)
  - Search operations (tasks, projects, users)
  - Bulk operations (update multiple tasks)

#### FR-004: Agent-Optimized Features
- **MUST** support batch operations to reduce round-trips
- **MUST** provide context-aware queries (my tasks, team tasks, etc.)
- **MUST** include rich metadata in responses (permissions, capabilities)
- **SHOULD** support natural language task parsing
- **SHOULD** provide agent-friendly error messages

#### FR-005: Authentication & Authorization
- **MUST** support Vikunja API token authentication
- **MUST** enforce same permission model as API v1
- **MUST** validate permissions at service layer
- **MUST** support token expiration and refresh

#### FR-006: Observability
- **MUST** log all MCP operations with agent ID
- **MUST** track operation performance metrics
- **SHOULD** provide agent usage analytics
- **SHOULD** support rate limiting per agent/user

### Non-Functional Requirements

#### NFR-001: Performance
- **MUST** respond to tool calls within 200ms (p95)
- **MUST** support 100+ concurrent agent connections
- **MUST** handle resource listing with pagination (default 50 items)
- **SHOULD** cache frequently accessed resources

#### NFR-002: Reliability
- **MUST** gracefully handle service layer errors
- **MUST** provide detailed error context to agents
- **MUST** implement circuit breaker for database
- **SHOULD** support request retries with exponential backoff

#### NFR-003: Security
- **MUST** validate all inputs from agents
- **MUST** sanitize agent-provided content before storage
- **MUST** prevent SQL injection via parameterized queries
- **MUST** rate limit per authenticated token (100 req/min)
- **MUST NOT** expose internal system details in errors

#### NFR-004: Compatibility
- **MUST** support MCP specification v1.0+
- **MUST** work with Claude Desktop, MCP clients
- **SHOULD** support MCP specification v2.0 when released
- **MUST** maintain backward compatibility within major versions

#### NFR-005: Documentation
- **MUST** provide MCP server configuration guide
- **MUST** document all exposed resources and tools
- **MUST** provide example agent workflows
- **SHOULD** include troubleshooting guide

## Technical Architecture

### MCP Server Stack

**Language**: TypeScript (Node.js)
- MCP SDK has best TypeScript support
- Integrates easily with existing ecosystem
- Strong typing for safety

**Alternative**: Python (if Go binding unavailable)

### Integration Points

1. **Authentication**: 
   - Agent provides Vikunja API token
   - MCP server validates against Vikunja auth service
   - Session maintained for connection lifetime

2. **Service Layer**:
   - MCP server calls existing Go service layer via HTTP or direct DB
   - Prefer: HTTP calls to API v1 (reuse existing)
   - Alternative: Direct service layer if packaging Go as library

3. **Resources**:
   - Map Vikunja entities to MCP resources
   - Include URIs: `vikunja://projects/123`, `vikunja://tasks/456`
   - Support filtering via query parameters

4. **Tools**:
   - Map service methods to MCP tools
   - Strong input validation schemas
   - Rich output with operation results

### Deployment

**Option A**: Standalone MCP Server
```
vikunja-mcp-server (Node.js process)
  ↓ HTTP
vikunja-api (Go process)
  ↓
Database
```

**Option B**: Embedded in Main Process
```
vikunja (Go process with embedded Node MCP server)
  ↓
Database
```

**Recommendation**: Option A (standalone) for:
- Independent scaling
- Language isolation
- Easier development
- Security boundary

## Implementation Phases

### Phase 1: MCP Server Foundation (Week 1)
- Set up MCP server project (TypeScript)
- Implement authentication bridge
- Create resource/tool registration system
- Basic error handling

### Phase 2: Core Resources & Tools (Week 2)
- Expose projects, tasks, labels resources
- Implement task CRUD tools
- Implement project management tools
- Implement search tools

### Phase 3: Advanced Features (Week 3)
- Expose teams, users, comments resources
- Implement bulk operations
- Implement assignment tools
- Add natural language task parsing

### Phase 4: Polish & Documentation (Week 4)
- Performance optimization
- Comprehensive testing
- Documentation
- Example agent workflows

## Success Metrics

1. **Agent Adoption**: 10+ unique agents using MCP server within 3 months
2. **Operation Latency**: p95 < 200ms for tool calls
3. **Reliability**: 99.5% success rate for valid operations
4. **Test Coverage**: 90%+ for MCP server code
5. **Zero Security Incidents**: No authorization bypasses or data leaks

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|---------|------------|-----------|
| MCP spec changes | High | Low | Abstract protocol layer, quick updates |
| Performance issues | Medium | Medium | Load testing, caching, optimization |
| Security vulnerabilities | Critical | Low | Security review, input validation, rate limiting |
| Service layer gaps | Medium | Low | Already 90%+ complete from 001 refactor |
| Agent adoption low | Medium | Low | Excellent documentation, example workflows |

## Dependencies

- ✅ **Complete Service Layer** (from Feature 001): Already done!
- ❌ MCP TypeScript SDK: Available
- ❌ Vikunja API token system: Exists
- ❌ MCP client for testing: Available (Claude Desktop)

## Open Questions

1. **Hosting**: Cloud-hosted MCP server or self-hosted only?
2. **Versioning**: How to version MCP server independently from main app?
3. **Rate Limiting**: Per-user or per-agent or both?
4. **Natural Language**: Use external LLM for task parsing or basic regex?

## Related Documents

- [API v2 PRD](../../vikunja/API_V2_PRD.md) - What we're NOT doing
- [API v2 Tasks](../../vikunja/API_V2_TASKS.md) - Incomplete v2 work
- [Service Layer Spec](../001-complete-service-layer/spec.md) - Foundation
- [Constitution](.specify/memory/constitution.md) - Compliance requirements

## Appendix: API v2 Disposition

**Recommendation**: Pause API v2 work indefinitely

**Rationale**:
- 95% incomplete
- Significant ongoing cost (maintain 3 APIs)
- Limited value over v1 for human developers
- MCP serves agent use case better

**If Later Needed**:
- Can always resume from current 20% completion
- Service layer makes it easier
- Focus on human dev pain points (if they emerge)

**Existing v2 Code**:
- Keep current 5 endpoints
- Mark as experimental
- Don't expand frontend usage
- No breaking changes to v1

---

**Prepared by**: GitHub Copilot  
**Review Status**: Pending stakeholder review  
**Next Step**: Get approval for MCP-first approach
