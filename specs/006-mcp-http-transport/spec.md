# Feature Specification: HTTP Transport for MCP Server

**Feature Branch**: `006-mcp-http-transport`  
**Created**: October 22, 2025  
**Status**: Draft  
**Input**: User description: "HTTP transport layer for Vikunja MCP Server to enable remote client connections via SSE and HTTP Streamable protocols"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remote Client Connection (Priority: P1)

A user running an AI workflow tool (like n8n or Claude Desktop) on their local machine needs to connect to the Vikunja MCP server running on a remote server to access task management tools.

**Why this priority**: This is the core capability that enables any remote MCP client to use Vikunja. Without this, MCP only works locally via stdio, severely limiting its usefulness.

**Independent Test**: Can be fully tested by configuring an MCP client with the server URL and authentication token, establishing a connection, and listing available tools. This delivers immediate value by enabling remote access.

**Acceptance Scenarios**:

1. **Given** a Vikunja MCP server is running on a remote host, **When** a user configures their MCP client with the server URL and valid authentication token, **Then** the client successfully establishes a connection and can list available tools.

2. **Given** a connected MCP client, **When** the user invokes a tool (e.g., "list tasks"), **Then** the tool executes successfully and returns the expected data.

3. **Given** a user attempts to connect without a valid token, **When** the connection request is made, **Then** the connection is rejected with an authentication error.

---

### User Story 2 - Modern Transport Protocol Support (Priority: P1)

A user wants to connect their n8n workflow to the Vikunja MCP server using the recommended HTTP Streamable transport protocol instead of the deprecated SSE protocol.

**Why this priority**: HTTP Streamable is the current MCP standard. SSE is deprecated and being removed from modern MCP clients. This ensures compatibility with current and future tooling.

**Independent Test**: Can be tested by configuring an MCP client (like n8n) to use HTTP Streamable mode and verifying successful connection, tool listing, and tool execution.

**Acceptance Scenarios**:

1. **Given** a Vikunja MCP server supports HTTP Streamable, **When** a user configures their n8n integration to use HTTP Streamable mode, **Then** the connection succeeds and all MCP operations work correctly.

2. **Given** a client supports multiple transport protocols, **When** the client negotiates capabilities with the server, **Then** the server advertises support for both SSE (for backward compatibility) and HTTP Streamable.

3. **Given** a user attempts to use an unsupported transport protocol, **When** the connection is initiated, **Then** the server returns a clear error message indicating which protocols are supported.

---

### User Story 3 - Secure Authenticated Access (Priority: P2)

A user needs to ensure that only authorized clients can access their Vikunja tasks through the MCP server, using their existing Vikunja API token.

**Why this priority**: Security is critical for production use, but the feature can demonstrate value without complex auth flows initially. Basic token authentication provides sufficient security for MVP.

**Independent Test**: Can be tested by attempting connections with valid tokens, expired tokens, invalid tokens, and no tokens, verifying that only valid tokens grant access.

**Acceptance Scenarios**:

1. **Given** a user has a valid Vikunja API token, **When** they provide this token in the authentication header or query parameter, **Then** they can access their tasks through the MCP server.

2. **Given** a user provides an invalid or expired token, **When** they attempt to connect, **Then** the connection is rejected with a clear authentication error.

3. **Given** a user's token permissions, **When** they execute MCP tools, **Then** the operations respect the same permissions as direct API calls would (e.g., read-only tokens cannot modify data).

---

### User Story 4 - Rate Limiting and Resource Protection (Priority: P3)

A system administrator needs to prevent abuse and ensure fair resource usage across multiple MCP clients connecting to the server.

**Why this priority**: Important for production stability but not essential for initial value delivery. Can be added after core connectivity works.

**Independent Test**: Can be tested by making rapid successive connection attempts or tool calls and verifying that rate limits are enforced with appropriate error responses.

**Acceptance Scenarios**:

1. **Given** rate limiting is configured for 100 requests per 15 minutes, **When** a client exceeds this limit, **Then** subsequent requests are rejected with a rate limit error until the window resets.

2. **Given** multiple clients connect with different tokens, **When** each client makes requests, **Then** rate limits are enforced per-token, not globally.

3. **Given** a client receives a rate limit error, **When** they check the error response, **Then** it includes information about when they can retry.

---

### User Story 5 - Session Management and Cleanup (Priority: P3)

When MCP clients connect and disconnect, the server needs to manage sessions efficiently and clean up resources when connections are terminated.

**Why this priority**: Important for production stability and resource efficiency, but connections can work without sophisticated session management initially.

**Independent Test**: Can be tested by establishing multiple connections, gracefully disconnecting some, abruptly terminating others, and verifying that server resources are properly cleaned up in all cases.

**Acceptance Scenarios**:

1. **Given** a client establishes an MCP connection, **When** the client gracefully disconnects, **Then** the server immediately cleans up the session and associated resources.

2. **Given** a client connection is abruptly terminated (network failure), **When** the timeout period expires, **Then** the server automatically cleans up the orphaned session.

3. **Given** multiple concurrent sessions exist, **When** the server is queried for status, **Then** it accurately reports the number of active sessions and their resource usage.

---

### Edge Cases

- What happens when a client sends malformed MCP protocol messages?
- How does the system handle a Redis connection failure during an active MCP session?
- What happens when a user's Vikunja API token is revoked while an MCP session is active?
- How does the system behave under extremely high connection request load (DDoS-like scenario)?
- What happens when a client keeps a connection open but sends no requests for an extended period?
- How does the system handle concurrent tool executions from the same client session?
- What happens when the Vikunja API backend is unavailable or slow to respond?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support HTTP Streamable transport protocol as defined in the MCP specification for bidirectional communication with MCP clients.

- **FR-002**: System MUST support Server-Sent Events (SSE) transport protocol for backward compatibility with older MCP clients.

- **FR-003**: System MUST implement capability negotiation to advertise supported transport protocols to connecting clients.

- **FR-004**: System MUST authenticate all connection attempts using Vikunja API tokens provided via Bearer authentication header or query parameter.

- **FR-005**: System MUST validate authentication tokens against the Vikunja API backend and reject connections with invalid tokens.

- **FR-006**: System MUST create isolated user contexts for each MCP session based on the authenticated user's permissions.

- **FR-007**: System MUST enforce the same permission model for MCP tool operations as direct Vikunja API calls (read-only tokens cannot modify data).

- **FR-008**: System MUST implement rate limiting to prevent abuse, with configurable limits per authentication token.

- **FR-009**: System MUST cache validated authentication tokens to reduce load on the Vikunja API backend.

- **FR-010**: System MUST maintain active session state for each connected client, including connection metadata and user context.

- **FR-011**: System MUST clean up session resources when clients disconnect gracefully or when connection timeouts expire.

- **FR-012**: System MUST handle malformed MCP protocol messages gracefully with appropriate error responses.

- **FR-013**: System MUST log all connection attempts, authentication events, and errors for security auditing.

- **FR-014**: System MUST provide a health check endpoint for monitoring server availability.

- **FR-015**: System MUST support deployment alongside the main Vikunja API service without port conflicts.

- **FR-016**: System MUST handle Redis connection failures gracefully, falling back to in-memory caching or rejecting requests with clear error messages.

- **FR-017**: System MUST return clear, actionable error messages for common failure scenarios (authentication failed, rate limited, invalid protocol, etc.).

### Key Entities

- **MCP Session**: Represents an active connection between an MCP client and the server. Attributes include session ID, user context, connection timestamp, transport protocol type, authentication token hash, and last activity timestamp.

- **User Context**: Associates an MCP session with a Vikunja user's permissions and identity. Attributes include user ID, authentication token, permission level, and cached user data.

- **Rate Limit State**: Tracks request counts for rate limiting enforcement. Attributes include token identifier, request count, window start time, and limit threshold.

- **Transport Handler**: Manages protocol-specific communication for each transport type (HTTP Streamable, SSE). Handles message serialization, connection lifecycle, and protocol-specific error handling.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: MCP clients can establish connections to the remote server in under 2 seconds under normal network conditions.

- **SC-002**: The server successfully handles 50 concurrent MCP client connections without performance degradation.

- **SC-003**: 100% of MCP tool operations respect the same permission model as direct API calls (verified through test suite).

- **SC-004**: Authentication token validation adds less than 100ms latency to connection establishment for cached tokens.

- **SC-005**: The system successfully rejects 100% of connection attempts with invalid authentication tokens.

- **SC-006**: Rate limiting prevents abuse while allowing 95% of legitimate usage patterns to proceed without throttling.

- **SC-007**: Gracefully disconnected sessions are cleaned up within 5 seconds, and timed-out sessions within 60 seconds.

- **SC-008**: The server maintains 99.5% uptime during normal operation (excluding planned maintenance).

- **SC-009**: Modern MCP clients (n8n, Claude Desktop) successfully connect using HTTP Streamable transport without configuration issues.

- **SC-010**: Server can be deployed to production environment with automated deployment scripts in under 15 minutes.

## Scope & Boundaries *(mandatory)*

### In Scope

- HTTP Streamable and SSE transport protocol implementations
- Token-based authentication using existing Vikunja API tokens
- Basic rate limiting per authentication token
- Session management and cleanup
- Health check endpoint for monitoring
- Deployment automation for Proxmox LXC environments
- Support for query parameter and header-based authentication (SSE limitation)
- Integration with existing Vikunja API backend
- Redis integration for token caching and rate limiting
- Error handling and security logging

### Out of Scope

- OAuth/OIDC authentication flows (future enhancement)
- WebSocket or gRPC transport protocols
- Multi-user support with separate server instances per session (future enhancement)
- TLS/SSL encryption at the application layer (handled by reverse proxy)
- Horizontal scaling with session persistence across instances
- Advanced authentication features like token rotation or refresh tokens
- Custom MCP tools beyond those already implemented in the stdio transport
- Performance benchmarking suite
- Prometheus metrics collection
- Built-in load balancing

## Assumptions *(mandatory)*

1. **Existing Infrastructure**: A Vikunja API server is already deployed and accessible to the MCP server.

2. **Redis Availability**: A Redis instance is available for token caching and rate limiting, or the server can fall back to in-memory storage.

3. **Network Configuration**: The deployment environment allows HTTP traffic on designated MCP ports (default 3010 for HTTP Streamable).

4. **Authentication Model**: Users already have Vikunja API tokens with appropriate permissions.

5. **Single User Per Session**: Each MCP session serves a single authenticated user (multi-user shared sessions are out of scope).

6. **Reverse Proxy for TLS**: Production deployments use a reverse proxy (nginx, Caddy) for TLS termination.

7. **MCP Protocol Stability**: The MCP specification for HTTP Streamable and SSE transports is stable and won't require breaking changes.

8. **Client Compatibility**: MCP clients properly implement the MCP specification for transport negotiation and protocol handling.

9. **Resource Limits**: Server environment has sufficient resources (memory, CPU, file descriptors) for target concurrent connection count.

10. **Monitoring Infrastructure**: External monitoring tools are available to track the health check endpoint.

## Dependencies *(mandatory)*

### External Dependencies

- **Vikunja API Backend**: Required for authentication, user data, and all task management operations. The MCP server is a proxy that cannot function without the backend API.

- **Redis (Optional)**: Used for token caching and rate limiting state. Server can operate with degraded performance using in-memory fallback if Redis is unavailable.

- **MCP Client Software**: n8n, Claude Desktop, or other MCP-compatible clients that support HTTP Streamable or SSE transports.

### Internal Dependencies

- **Existing MCP Tools**: The stdio-based MCP server implementation with all Vikunja tools (task management, project operations, etc.) must be complete and tested.

- **VikunjaClient Module**: The API client library used to communicate with the Vikunja backend must support the required operations.

- **Configuration System**: Existing configuration management for server ports, URLs, and feature flags.

### Technical Prerequisites

- Node.js 22+ runtime environment
- TypeScript 5.x compiler
- @modelcontextprotocol/sdk package with HTTP transport support
- Express.js for HTTP server functionality
- uuid package for session ID generation
- Zod for configuration validation

## Non-Functional Requirements *(mandatory)*

### Performance

- Connection establishment: Under 2 seconds for new connections, under 500ms for cached token authentication
- Tool execution latency: No more than 200ms overhead compared to direct API calls
- Concurrent connections: Support minimum 50 simultaneous clients
- Memory usage: Under 100MB per 10 active sessions

### Reliability

- Session cleanup: 100% of sessions cleaned up within timeout period (no resource leaks)
- Error recovery: Graceful handling of Redis failures, API backend unavailability, and network interruptions
- Uptime target: 99.5% availability during normal operation

### Security

- Authentication: 100% of requests must be authenticated, no unauthenticated access permitted
- Permission enforcement: All operations must respect user permissions from Vikunja API
- Token security: No tokens logged in plaintext, secure handling in memory
- Rate limiting: Protection against abuse and DDoS attempts
- Audit logging: All authentication attempts and errors must be logged

### Scalability

- Initial target: 50 concurrent connections on single server instance
- Resource usage must scale linearly with connection count
- Support for future horizontal scaling architecture (though not implemented initially)

### Maintainability

- Code coverage: Minimum 80% test coverage for HTTP transport modules
- Documentation: Complete API documentation for all endpoints
- Deployment: Automated deployment scripts with validation
- Monitoring: Health check endpoint for integration with monitoring systems

### Compatibility

- MCP Protocol: Full compliance with MCP specification for HTTP Streamable and SSE
- Client Support: Verified compatibility with n8n and Claude Desktop
- Browser Compatibility: SSE transport works with EventSource API in modern browsers
- Backward Compatibility: Continued support for SSE alongside HTTP Streamable

## Open Questions

*None - all critical decisions have reasonable defaults documented in assumptions.*
