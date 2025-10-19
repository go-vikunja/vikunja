# Changelog

All notable changes to the Vikunja MCP Server will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
