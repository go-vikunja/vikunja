# Feature Specification: MCP HTTP/SSE Transport

**Feature Branch**: `006-mcp-http-transport`  
**Created**: 2025-10-22  
**Status**: Draft  
**Input**: User description: "Add HTTP and Server-Sent Events (SSE) transport support to the Vikunja MCP server alongside the existing stdio transport, with seamless integration into the Proxmox LXC deployment scripts."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic HTTP Transport for External Clients (Priority: P1)

External automation tools (n8n workflows, Python scripts) can connect to the MCP server over HTTP to execute Vikunja operations without requiring direct process access.

**Why this priority**: This is the core value proposition - enabling remote client connectivity. Without HTTP transport, the MCP server cannot be used by external tools when deployed as a systemd service in Proxmox LXC containers.

**Independent Test**: Deploy MCP server with HTTP transport enabled, connect from external n8n instance using HTTP POST requests with valid API token, execute a task creation tool, verify task appears in Vikunja. This delivers immediate value by enabling remote automation.

**Acceptance Scenarios**:

1. **Given** MCP server is running with HTTP transport enabled, **When** n8n sends HTTP POST request to `/sse` endpoint with valid Vikunja API token in Authorization header, **Then** MCP server establishes SSE connection and returns 200 OK
2. **Given** active SSE connection from n8n client, **When** client sends MCP protocol message to create a task, **Then** server processes request, creates task in Vikunja, and returns success response via SSE
3. **Given** MCP server with HTTP transport, **When** client sends request with invalid API token, **Then** server returns 401 Unauthorized error immediately without establishing SSE connection
4. **Given** MCP server with HTTP transport, **When** Python script connects using MCP SDK's SSEServerTransport with valid token in query parameter, **Then** connection succeeds and all MCP tools/resources/prompts work identically to stdio transport

---

### User Story 2 - Automated Proxmox Deployment (Priority: P2)

System administrators deploying Vikunja via Proxmox deployment scripts get HTTP transport configured automatically without manual intervention.

**Why this priority**: Eliminates manual configuration burden and prevents deployment errors. Makes HTTP transport the "default happy path" for Proxmox deployments.

**Independent Test**: Run `vikunja-install.sh` on clean Proxmox LXC container, verify MCP systemd service starts with HTTP transport, test connectivity from external host, confirm deployment summary displays correct connection instructions. This is testable independently and delivers value even without implementing stdio backward compatibility.

**Acceptance Scenarios**:

1. **Given** clean Proxmox LXC container, **When** administrator runs `vikunja-install.sh` with default settings, **Then** MCP systemd service is created with `TRANSPORT_TYPE=http` environment variable
2. **Given** MCP service deployed via Proxmox scripts, **When** service starts, **Then** health check verifies HTTP endpoint responds successfully at configured port (3010 for blue, 3011 for green)
3. **Given** successful Proxmox deployment, **When** installation completes, **Then** deployment summary displays MCP HTTP URL with connection examples for n8n and Python clients
4. **Given** deployed MCP service in blue environment, **When** administrator runs `vikunja-update.sh` to switch to green, **Then** HTTP transport configuration persists and new service starts with same transport type

---

### User Story 3 - Backward Compatibility for Stdio Users (Priority: P3)

Existing integrations using stdio transport (subprocess communication) continue to work without changes after HTTP transport is added.

**Why this priority**: Prevents breaking changes for existing users. Lower priority because stdio users are likely manual/development setups, not production Proxmox deployments.

**Independent Test**: Deploy MCP server manually with `TRANSPORT_TYPE=stdio` (or omitted), connect client via subprocess stdin/stdout, execute MCP operations, verify behavior is unchanged from pre-HTTP implementation. This can be tested in isolation using existing test suites.

**Acceptance Scenarios**:

1. **Given** MCP server started without `TRANSPORT_TYPE` environment variable, **When** client connects via stdio (subprocess stdin/stdout), **Then** server operates in stdio mode as it did before HTTP feature was added
2. **Given** MCP server with `TRANSPORT_TYPE=stdio`, **When** server starts, **Then** it does not bind to HTTP port and only accepts stdio communication
3. **Given** MCP server with stdio transport, **When** health check Express server starts, **Then** it remains operational on separate port regardless of MCP transport type
4. **Given** existing stdio-based integration tests, **When** tests run after HTTP feature is deployed, **Then** all tests pass without modification

---

### Edge Cases

- **Invalid transport configuration**: What happens when `TRANSPORT_TYPE` is set to unsupported value (e.g., "websocket")? Server should fail fast with clear error message indicating valid options.
- **Missing MCP_PORT with HTTP**: What happens when `TRANSPORT_TYPE=http` but `MCP_PORT` is not configured? Server should fail to start with error message requiring port configuration.
- **Concurrent stdio and HTTP**: What happens if administrator tries to enable both transports simultaneously? Single transport mode is enforced - server chooses one based on `TRANSPORT_TYPE` priority.
- **Mid-flight requests during blue-green switch**: What happens to active SSE connections when nginx switches from blue to green environment? Existing connections gracefully close, clients retry with exponential backoff to new environment.
- **Token expiration during SSE session**: What happens when Vikunja API token expires mid-session? Next MCP request with expired token returns authentication error, client must re-authenticate with fresh token.
- **Malformed SSE messages**: What happens when client sends invalid MCP protocol message format? Server returns protocol error response via SSE, connection remains open for retry.
- **Rate limiting across transports**: Are rate limits applied consistently to HTTP and stdio? Yes, existing rate limiting (100 req/min per user) applies uniformly regardless of transport type.
- **CORS for browser-based clients**: What happens when browser tries to connect to MCP HTTP endpoint? Server must validate CORS headers and restrict to configured allowed origins (default: deny browser access, document how to enable).

## Requirements *(mandatory)*

### Functional Requirements

#### HTTP Transport Core

- **FR-001**: System MUST support runtime selection between stdio and HTTP transport modes via `TRANSPORT_TYPE` environment variable
- **FR-002**: System MUST default to stdio transport when `TRANSPORT_TYPE` is not specified (backward compatibility)
- **FR-003**: System MUST validate `TRANSPORT_TYPE` value is either "stdio" or "http", rejecting other values with descriptive error
- **FR-004**: When HTTP transport is selected, system MUST bind SSE endpoint to port specified by `MCP_PORT` environment variable
- **FR-005**: System MUST require `MCP_PORT` configuration when `TRANSPORT_TYPE=http`, failing to start if port is missing

#### SSE Protocol Implementation

- **FR-006**: HTTP transport MUST implement Server-Sent Events protocol for MCP message exchange
- **FR-007**: System MUST accept SSE connection requests at `/sse` HTTP endpoint using POST method
- **FR-008**: System MUST maintain persistent SSE connection for bidirectional MCP protocol communication
- **FR-009**: System MUST encode MCP protocol messages according to MCP SDK SSE transport specification
- **FR-010**: All MCP capabilities (tools, resources, prompts) MUST function identically across stdio and HTTP transports

#### Authentication & Security

- **FR-011**: HTTP transport MUST authenticate every incoming request using per-request Vikunja API tokens (no server-level shared tokens)
- **FR-012**: System MUST extract API token from HTTP `Authorization: Bearer <token>` header as primary authentication method
- **FR-013**: System MUST support API token in URL query parameter `?token=<token>` as fallback authentication method
- **FR-014**: System MUST validate token by querying Vikunja backend user info endpoint before establishing SSE connection
- **FR-015**: System MUST cache validated tokens for 5 minutes using existing Authenticator class to reduce backend load
- **FR-016**: System MUST return 401 Unauthorized HTTP status for invalid or missing tokens without establishing SSE connection
- **FR-017**: System MUST apply existing rate limiting (100 requests/minute per user) uniformly to HTTP transport
- **FR-018**: System MUST validate CORS headers for HTTP requests and restrict cross-origin access to configured allowed origins

#### Deployment Integration

- **FR-019**: Proxmox deployment script MUST generate MCP systemd service with `TRANSPORT_TYPE=http` environment variable by default
- **FR-020**: Proxmox deployment script MUST set `MCP_PORT` to 3010 for blue environment and 3011 for green environment
- **FR-021**: Proxmox installation script MUST verify MCP HTTP endpoint health after service start before marking deployment successful
- **FR-022**: Proxmox installation script MUST display MCP HTTP connection URL and authentication instructions in deployment summary
- **FR-023**: Proxmox update script MUST preserve `TRANSPORT_TYPE` configuration across blue-green environment switches
- **FR-024**: Health check endpoint MUST verify HTTP transport is accepting connections when `TRANSPORT_TYPE=http`
- **FR-025**: Blue-green deployment MUST ensure HTTP transport configuration persists in both environments during updates

#### Error Handling & Monitoring

- **FR-026**: System MUST log transport type selection at startup (stdio vs HTTP)
- **FR-027**: System MUST log HTTP port binding success or failure with actionable error messages
- **FR-028**: System MUST return HTTP 503 Service Unavailable when server is shutting down or overloaded
- **FR-029**: System MUST gracefully close active SSE connections during shutdown with proper cleanup
- **FR-030**: Health check endpoint MUST remain operational on separate Express server regardless of MCP transport failures

### Key Entities

- **Transport Configuration**: Specifies runtime transport mode (stdio or HTTP), port binding for HTTP, and CORS settings
- **SSE Connection**: Represents active Server-Sent Events session between client and MCP server, maintains authentication context and message queue
- **Authentication Context**: Stores validated API token, associated Vikunja user, token expiry time, and cached user permissions

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: External clients (n8n, Python) can successfully connect to MCP server over HTTP and execute all MCP tools with 100% feature parity to stdio transport
- **SC-002**: Proxmox deployment completes successfully with HTTP transport enabled in under 5 minutes without manual configuration steps
- **SC-003**: MCP server handles at least 50 concurrent HTTP SSE connections without performance degradation (response time under 500ms for tool execution)
- **SC-004**: Token validation caching reduces Vikunja backend authentication requests by at least 80% compared to validating every request
- **SC-005**: Zero regression in existing stdio transport functionality - all pre-existing integration tests pass without modification
- **SC-006**: Deployment health checks catch HTTP transport configuration errors before marking installation successful (100% error detection for misconfigured ports, missing environment variables)
- **SC-007**: Blue-green deployments complete with zero downtime for MCP HTTP clients - active connections gracefully migrate during environment switch
- **SC-008**: MCP server startup fails fast (within 2 seconds) with clear error messages for invalid transport configuration
- **SC-009**: Documentation enables new users to connect n8n or Python clients to MCP HTTP endpoint in under 10 minutes using provided examples

## Assumptions

- **Assumption 1**: Vikunja backend API provides stable `/user` endpoint for token validation that returns user info consistently
- **Assumption 2**: MCP SDK SSE transport implementation is stable and compatible with current MCP protocol version used in Vikunja
- **Assumption 3**: Proxmox LXC containers have network connectivity allowing inbound HTTP connections on MCP ports (3010/3011)
- **Assumption 4**: Redis instance used for token caching is already deployed and accessible in Proxmox environment (existing infrastructure)
- **Assumption 5**: Blue-green deployment scripts have exclusive control over systemd service files - no external modifications to service configuration
- **Assumption 6**: Client retry logic handles SSE connection drops during blue-green switches - server does not need to implement connection migration
- **Assumption 7**: Default CORS policy (deny all cross-origin) is acceptable for initial release - specific origin whitelisting can be added later if needed
- **Assumption 8**: Existing rate limiting implementation (100 req/min) is sufficient for HTTP transport - no separate HTTP-specific limits needed
- **Assumption 9**: MCP HTTP endpoint does not require TLS termination at application level - handled by nginx reverse proxy if needed
- **Assumption 10**: Node.js runtime version in Proxmox containers supports SSE implementation requirements (Node.js 18+)

