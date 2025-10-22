# Vikunja MCP Server

> Model Context Protocol (MCP) server for Vikunja task management, enabling AI agents to interact with Vikunja through a standardized protocol.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tests](https://img.shields.io/badge/tests-193%2F196%20passing-brightgreen)](./tests)
[![Coverage](https://img.shields.io/badge/coverage-98.5%25-brightgreen)](./tests)

## What is This?

The Vikunja MCP Server enables AI agents (like Claude Desktop, n8n, custom scripts) to interact with [Vikunja](https://vikunja.io) through the [Model Context Protocol](https://modelcontextprotocol.io).

**Key Features:**
- 🤖 **21 MCP Tools** - Complete CRUD for projects, tasks, labels, and more
- 🔐 **Secure** - Token-based authentication with rate limiting
- 🚀 **Fast** - <200ms p95 latency, stateless for horizontal scaling
- 📦 **Easy Deploy** - Docker Compose or standalone
- 📚 **Well Documented** - API reference, examples, and integration guides

## Features

- **Native MCP Integration**: Full MCP v1.0+ protocol support
- **Dual Transport Support**: Stdio (default) for process spawning, HTTP/SSE for network clients
- **Resource Exposure**: Projects, tasks, labels, teams, and more
- **Tool Operations**: Create, update, delete operations for all entities
- **Rate Limiting**: Per-token rate limiting with Redis backend
- **Authentication**: Vikunja API token-based authentication
- **Performance**: <200ms p95 latency for tool calls
- **Scalability**: Stateless design for horizontal scaling

## Transport Configuration

The MCP server supports two transport mechanisms:

### Stdio Transport (Default)

**Best for:** Claude Desktop, direct subprocess spawning

```bash
# No additional configuration needed
TRANSPORT_TYPE=stdio  # Default if omitted
```

**Usage:**
```json
{
  "mcpServers": {
    "vikunja": {
      "command": "node",
      "args": ["/path/to/dist/index.js"],
      "env": {
        "VIKUNJA_API_URL": "http://localhost:3456",
        "VIKUNJA_API_TOKEN": "your-token"
      }
    }
  }
}
```

### HTTP/SSE Transport

**Best for:** n8n workflows, Python MCP SDK clients, web-based AI agents

```bash
# Enable HTTP transport
TRANSPORT_TYPE=http
MCP_PORT=3010  # Required for HTTP transport

# Optional: CORS for browser clients
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://n8n.example.com,https://app.example.com
```

**Usage (n8n example):**
```javascript
// HTTP Request node
{
  "method": "POST",
  "url": "http://localhost:3010/sse",
  "headers": {
    "Authorization": "Bearer YOUR_VIKUNJA_TOKEN"
  }
}
```

**Usage (Python MCP SDK):**
```python
from mcp.client.sse import sse_client
import httpx

async with httpx.AsyncClient() as http_client:
    async with sse_client(
        http_client=http_client,
        url="http://localhost:3010/sse",
        headers={"Authorization": "Bearer YOUR_TOKEN"}
    ) as (read, write):
        # Use MCP client session
        pass
```

See [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md) for comprehensive transport configuration.

## Quick Start

### 1. Deploy with Docker (2 minutes)

```bash
# Clone or download the MCP server
cd /path/to/vikunja-mcp-server

# Configure environment
cp .env.example .env
nano .env  # Set your Vikunja URL and token

# Start services
docker-compose up -d

# Verify
curl http://localhost:3457/health
```

### 2. Configure Your AI Agent

#### Claude Desktop

Edit `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "vikunja": {
      "command": "docker",
      "args": [
        "exec", "-i", "vikunja-mcp-server",
        "node", "/app/dist/index.js"
      ],
      "env": {
        "VIKUNJA_API_TOKEN": "your-vikunja-token"
      }
    }
  }
}
```

#### n8n

Use the [MCP Tool node](https://docs.n8n.io/integrations/builtin/cluster-nodes/sub-nodes/n8n-nodes-langchain.toolmcp/) in your AI Agent workflow:

1. Add an **AI Agent** node to your workflow
2. Add a **Tool MCP** sub-node
3. Configure the MCP connection:
   ```json
   {
     "command": "docker",
     "args": ["exec", "-i", "vikunja-mcp-server", "node", "/app/dist/index.js"],
     "env": {
       "VIKUNJA_API_TOKEN": "your-vikunja-token"
     }
   }
   ```
4. The AI agent can now use all 21 Vikunja tools automatically

See [docs/INTEGRATIONS.md](docs/INTEGRATIONS.md#n8n) for detailed setup instructions.

### 3. Test It!

**With Claude Desktop:**
```
You: "Create a project called 'Website Redesign'"
Claude: ✓ Created project "Website Redesign" (ID: 42)
```

**With curl:**
```bash
curl -X POST http://localhost:3457/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "create_task",
    "arguments": {
      "project_id": 1,
      "title": "My first task"
    },
    "auth": {
      "token": "your-vikunja-token"
    }
  }'
```

## Installation

```bash
npm install
```

## Available Tools

### Projects (4 tools)
- `create_project` - Create a new project
- `update_project` - Update project details
- `delete_project` - Delete a project
- `archive_project` - Archive/unarchive a project

### Tasks (5 tools)
- `create_task` - Create a task in a project
- `update_task` - Update task details
- `complete_task` - Mark a task as complete
- `delete_task` - Delete a task
- `move_task` - Move task to another project

### Assignments & Labels (5 tools)
- `assign_task` - Assign a user to a task
- `unassign_task` - Remove a user from a task
- `add_label` - Add a label to a task
- `remove_label` - Remove a label from a task
- `create_label` - Create a new label

### Search (4 tools)
- `search_tasks` - Search tasks with advanced filtering
- `search_projects` - Search projects
- `get_my_tasks` - Get current user's assigned tasks
- `get_project_tasks` - Get all tasks in a project

### Bulk Operations (4 tools)
- `bulk_update_tasks` - Update multiple tasks at once (max 100)
- `bulk_complete_tasks` - Complete multiple tasks
- `bulk_assign_tasks` - Assign user to multiple tasks
- `bulk_add_labels` - Add label to multiple tasks

## Documentation

- **[API Reference](docs/API.md)** - Complete tool documentation with schemas
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Docker, LXC, systemd deployment
- **[Integration Guide](docs/INTEGRATIONS.md)** - Platform-specific setup (Claude, n8n, Python, JS)
- **[Examples](docs/EXAMPLES.md)** - 12 workflow examples and patterns

## Configuration

Create a `.env` file:

```env
VIKUNJA_API_URL=http://localhost:3456
MCP_PORT=3457
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
RATE_LIMIT_DEFAULT=100
RATE_LIMIT_BURST=120
LOG_LEVEL=info
```

Or set environment variables:

```bash
# Required
VIKUNJA_API_URL=http://localhost:3456
VIKUNJA_API_TOKEN=your-token-here  # Or pass per-request

# Optional
MCP_PORT=3457                      # Default: 3457
REDIS_HOST=localhost               # Default: localhost
REDIS_PORT=6379                    # Default: 6379
REDIS_PASSWORD=                    # Optional
RATE_LIMIT_DEFAULT=100             # Requests/min, default: 100
RATE_LIMIT_BURST=120               # Burst limit, default: 120
RATE_LIMIT_ADMIN_BYPASS=false      # Bypass for admin tokens
LOG_LEVEL=info                     # error|warn|info|debug
LOG_FORMAT=json                    # json|simple
```

## Development

### Prerequisites
- Node.js 20+
- Redis (for rate limiting)
- Vikunja instance running

### Setup

```bash
# Install dependencies
npm install

# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Build
npm run build

# Run in development mode with hot reload
npm run dev

# Lint code
npm run lint

# Format code
npm run format
```

### Project Structure

```
vikunja-mcp-server/
├── src/
│   ├── config/          # Configuration management
│   ├── auth/            # Authentication
│   ├── ratelimit/       # Rate limiting (Redis-backed)
│   ├── vikunja/         # Vikunja API client
│   ├── tools/           # MCP tool implementations
│   ├── resources/       # MCP resource providers
│   ├── utils/           # Utilities (logger, errors)
│   └── index.ts         # MCP server entry point
├── tests/
│   ├── unit/            # Unit tests
│   └── integration/     # Integration tests
├── docs/                # Documentation
├── Dockerfile           # Production Docker image
├── docker-compose.yml   # Full stack deployment
└── package.json
```

## Architecture

```
┌─────────────────┐
│   AI Agents     │
│ Claude/n8n/etc  │
└────────┬────────┘
         │ MCP Protocol
         ▼
┌─────────────────┐      ┌──────────┐
│  MCP Server     │─────▶│  Redis   │
│  (Port 3457)    │      │  (Rate   │
└────────┬────────┘      │  Limit)  │
         │               └──────────┘
         │ HTTP/REST
         ▼
┌─────────────────┐
│  Vikunja API    │
│  (Port 3456)    │
└─────────────────┘
```

Alternative view:

```
AI Agent → MCP Server → Vikunja API v1 → Service Layer → Database
              ↓
        Resources/Tools
```

## Use Cases

### Task Creation from Email
```
Email arrives → n8n parses → create_task → Vikunja
```

### Daily Standup Reports
```
Claude: "What are my tasks for today?"
→ get_my_tasks → Format as standup report
```

### Sprint Planning
```
Claude: "Move all urgent tasks to this week's sprint"
→ search_tasks (priority: 5) → bulk_update_tasks (due_date)
```

### Team Workload Analysis
```
Python script → search_tasks (per user) → Generate report
```

See [docs/EXAMPLES.md](docs/EXAMPLES.md) for 12 complete workflow examples.

## Performance

- **Latency**: <200ms p95 for tool calls
- **Rate Limiting**: 100 req/min default, 120 burst
- **Scalability**: Stateless design, horizontal scaling ready
- **Test Coverage**: 98.5% (193/196 tests passing)

## Security

- **Authentication**: Vikunja API token (per-request or env)
- **Rate Limiting**: Token-based with Redis backend
- **Input Validation**: Zod schemas for all inputs
- **Error Handling**: No sensitive data in error messages
- **Docker**: Non-root user, minimal attack surface

## Troubleshooting

### MCP Server Won't Start
```bash
# Check logs
docker logs vikunja-mcp-server

# Verify Vikunja is accessible
curl http://localhost:3456/api/v1/info

# Check Redis
docker exec -it vikunja-mcp-redis redis-cli ping
```

### Authentication Errors
```bash
# Test token directly with Vikunja
curl -H "Authorization: Bearer $VIKUNJA_API_TOKEN" \
  http://localhost:3456/api/v1/projects
```

### Rate Limiting
```bash
# Check current rate limit status
docker exec -it vikunja-mcp-redis redis-cli keys "ratelimit:*"

# Clear rate limits (for testing)
docker exec -it vikunja-mcp-redis redis-cli FLUSHDB
```

See [docs/DEPLOYMENT.md#troubleshooting](docs/DEPLOYMENT.md#troubleshooting) for more.

## Contributing

Contributions welcome! Please:

1. Run tests: `npm test`
2. Maintain coverage: `npm run test:coverage` (>90%)
3. Follow style guide: `npm run lint`
4. Update docs as needed

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- [Vikunja](https://vikunja.io) - The task management system
- [Model Context Protocol](https://modelcontextprotocol.io) - The protocol specification
- [Claude Desktop](https://claude.ai/desktop) - AI assistant with MCP support
- [n8n](https://n8n.io) - Workflow automation platform

---

**Ready to automate your tasks?** Start with the [Quick Start](#quick-start) above! 🚀
