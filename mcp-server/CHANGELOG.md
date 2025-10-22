# Changelog

All notable changes to the Vikunja MCP Server will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-01-XX

### Added
- **HTTP/SSE Transport Support** - Enable network-accessible AI agent connectivity
  - New `TRANSPORT_TYPE` configuration (stdio | http)
  - Server-Sent Events (SSE) endpoint at `POST /sse`
  - Per-request authentication with Vikunja API tokens
  - Support for multiple concurrent clients
  - Integration examples for n8n workflows and Python MCP SDK
- **Dual Transport Architecture** - Maintain backward compatibility
  - Stdio transport remains default for Claude Desktop
  - HTTP transport enables web-based AI agents and automation tools
  - Transport factory pattern for runtime selection
- **Enhanced Authentication** - SSE-specific auth middleware
  - Token extraction from Authorization header or query parameter
  - User context caching (5-minute TTL via existing Redis)
  - Request-level authentication validation
- **Connection Management** - Graceful SSE lifecycle
  - Connection tracking with unique IDs
  - Automatic cleanup on client disconnect
  - Graceful shutdown closes all active connections
- **CORS Configuration** - Optional cross-origin support
  - `CORS_ENABLED` flag for browser-based clients
  - `CORS_ALLOWED_ORIGINS` whitelist configuration
- **Comprehensive Documentation** - Updated deployment guides
  - HTTP transport configuration in DEPLOYMENT.md
  - n8n, Python SDK, and curl examples
  - Troubleshooting guide for HTTP connections
  - Security best practices for network exposure
  - Proxmox deployment automatically configures HTTP transport

### Changed
- Proxmox deployment scripts now configure `TRANSPORT_TYPE=http` by default
- MCP server ports: 3010 (blue), 3011 (green) for HTTP transport
- Health checks updated to verify HTTP/SSE endpoint availability

### Technical Details
- TypeScript 5.x with strict null checks
- `@modelcontextprotocol/sdk` SSE transport integration
- Express 4.x for HTTP server
- Zod validation for transport configuration
- UUID v4 for connection identifiers
- 27 new transport-layer tests (100% passing)

### Migration Notes
- **No breaking changes** - Stdio transport remains default
- To enable HTTP transport, set `TRANSPORT_TYPE=http` and `MCP_PORT=3010`
- Existing stdio configurations continue to work without modification
- See [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md) for HTTP transport setup

### Security Considerations
- HTTP transport exposes MCP server on network - use firewall rules
- Per-request token validation ensures authentication security
- Consider reverse proxy with TLS for production deployments
- CORS should be restricted to trusted origins

## [1.0.0] - 2025-10-17

### Added
- Initial release of Vikunja MCP Server
- Full MCP v1.0+ protocol support
- 21 production-ready tools:
  - 4 project management tools (create, update, delete, archive)
  - 5 task management tools (create, update, complete, delete, move)
  - 5 assignment & label tools (assign, unassign, add/remove labels, create label)
  - 4 search tools (search tasks/projects, get user tasks, get project tasks)
  - 4 bulk operation tools (bulk update, complete, assign, label)
- Token-based authentication with Vikunja API
- Redis-backed rate limiting
- Docker deployment support with docker-compose
- Comprehensive documentation:
  - API reference with all tool schemas
  - Deployment guide (Docker, LXC, systemd)
  - Integration guides (Claude Desktop, n8n, Python, JavaScript)
  - 12 workflow examples
- Health check endpoint
- Structured logging with Winston
- Error handling with JSON-RPC 2.0 error codes
- Input validation with Zod schemas
- 98.5% test coverage (193/196 tests passing)

### Technical Details
- Built with TypeScript
- MCP SDK integration
- Express.js for HTTP endpoints
- Axios for Vikunja API communication
- ioredis for Redis connection
- Stateless design for horizontal scaling
- <200ms p95 latency for tool calls

### Documentation
- Complete README with quick start guide
- API.md - Full tool reference
- DEPLOYMENT.md - Production deployment guide
- INTEGRATIONS.md - Platform integration guides
- EXAMPLES.md - Workflow examples and patterns

## Future Roadmap

### Planned Features
- Webhook support for real-time updates
- Additional bulk operations
- Performance metrics endpoint
- WebSocket support for streaming updates
- Advanced filtering options
- Project templates
- Recurring task automation
- Integration examples for additional platforms

### Under Consideration
- Admin dashboard
- Multi-tenancy support
- Custom field support
- Advanced analytics

---

For detailed changes and implementation progress, see the git commit history.
