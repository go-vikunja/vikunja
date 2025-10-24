# vikunja Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-19

## Active Technologies
- Bash 4.0+ (deployment scripts), Go 1.21+ (Vikunja backend), Node.js 22+ (frontend build, MCP server) (004-proxmox-deployment)
- TypeScript 5.x, Node.js 22+, Bash 4.0+ (deployment scripts) + `@modelcontextprotocol/sdk` (SSE transport), Express 4.x (health checks), Zod (config validation), Redis (token caching) (006-mcp-http-transport)
- Redis (authentication token cache, 5-minute TTL) (006-mcp-http-transport)
- TypeScript 5.x, Node.js 22+ + @modelcontextprotocol/sdk (SSE & HTTP Streamable transports), Express 4.x (HTTP server), Zod (config validation), ioredis (token caching), rate-limiter-flexible (abuse prevention), uuid (session IDs), winston (logging) (006-mcp-http-transport)
- Redis (token cache & rate limiting state, 5-minute TTL), in-memory fallback if Redis unavailable (006-mcp-http-transport)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for Bash 4.0+ (deployment scripts), Go 1.21+ (Vikunja backend), Node.js 22+ (frontend build, MCP server)

## Code Style
Bash 4.0+ (deployment scripts), Go 1.21+ (Vikunja backend), Node.js 22+ (frontend build, MCP server): Follow standard conventions

## Recent Changes
- 006-mcp-http-transport: Added TypeScript 5.x, Node.js 22+ + @modelcontextprotocol/sdk (SSE & HTTP Streamable transports), Express 4.x (HTTP server), Zod (config validation), ioredis (token caching), rate-limiter-flexible (abuse prevention), uuid (session IDs), winston (logging)
- 006-mcp-http-transport: Added TypeScript 5.x, Node.js 22+, Bash 4.0+ (deployment scripts) + `@modelcontextprotocol/sdk` (SSE transport), Express 4.x (health checks), Zod (config validation), Redis (token caching)
- 004-proxmox-deployment: Added Bash 4.0+ (deployment scripts), Go 1.21+ (Vikunja backend), Node.js 22+ (frontend build, MCP server)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
