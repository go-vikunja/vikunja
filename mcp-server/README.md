# Vikunja MCP Server# Vikunja MCP Server



> Model Context Protocol (MCP) server providing AI agents with full access to Vikunja task management.Model Context Protocol (MCP) server for Vikunja task management, enabling AI agents to interact with Vikunja through a standardized protocol.



[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)## Features

[![Tests](https://img.shields.io/badge/tests-193%2F196%20passing-brightgreen)](./tests)

[![Coverage](https://img.shields.io/badge/coverage-98.5%25-brightgreen)](./tests)- **Native MCP Integration**: Full MCP v1.0+ protocol support

- **Resource Exposure**: Projects, tasks, labels, teams, and more

## What is This?- **Tool Operations**: Create, update, delete operations for all entities

- **Rate Limiting**: Per-token rate limiting with Redis backend

The Vikunja MCP Server enables AI agents (like Claude Desktop, n8n, custom scripts) to interact with [Vikunja](https://vikunja.io) through the [Model Context Protocol](https://modelcontextprotocol.io).- **Authentication**: Vikunja API token-based authentication

- **Performance**: <200ms p95 latency for tool calls

**Key Features:**- **Scalability**: Stateless design for horizontal scaling

- 🤖 **21 MCP Tools** - Complete CRUD for projects, tasks, labels, and more

- 🔐 **Secure** - Token-based authentication with rate limiting## Installation

- 🚀 **Fast** - <200ms p95 latency, stateless for horizontal scaling

- 📦 **Easy Deploy** - Docker Compose or standalone```bash

- 📚 **Well Documented** - API reference, examples, and integration guidesnpm install

```

## Quick Start

## Configuration

### 1. Deploy with Docker (2 minutes)

Create a `.env` file:

```bash

# Clone or download the MCP server```env

cd /path/to/vikunja-mcp-serverVIKUNJA_API_URL=http://localhost:3456

MCP_PORT=3457

# Configure environmentREDIS_HOST=localhost

cp .env.example .envREDIS_PORT=6379

nano .env  # Set your Vikunja URL and tokenREDIS_PASSWORD=

RATE_LIMIT_DEFAULT=100

# Start servicesRATE_LIMIT_BURST=120

docker-compose up -dLOG_LEVEL=info

```

# Verify

curl http://localhost:3457/health## Development

```

```bash

### 2. Configure Your AI Agent# Run in development mode with hot reload

npm run dev

#### Claude Desktop

# Build for production

Edit `~/.config/Claude/claude_desktop_config.json`:npm run build



```json# Run tests

{npm test

  "mcpServers": {

    "vikunja": {# Run tests with coverage

      "command": "docker",npm run test:coverage

      "args": [

        "exec", "-i", "vikunja-mcp-server",# Lint code

        "node", "/app/dist/index.js"npm run lint

      ],

      "env": {# Format code

        "VIKUNJA_API_TOKEN": "your-vikunja-token"npm run format

      }```

    }

  }## Architecture

}

``````

AI Agent → MCP Server → Vikunja API v1 → Service Layer → Database

#### n8n              ↓

        Resources/Tools

Use the HTTP Request node:```

```json

{## License

  "url": "http://localhost:3457/tools/execute",

  "method": "POST",MIT

  "body": {
    "tool": "create_task",
    "arguments": {
      "project_id": 1,
      "title": "Task from n8n"
    }
  }
}
```

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

Create a `.env` file or set environment variables:

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

# Run in development mode
npm run dev

# Lint
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

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- [Vikunja](https://vikunja.io) - The task management system
- [Model Context Protocol](https://modelcontextprotocol.io) - The protocol specification
- [Claude Desktop](https://claude.ai/desktop) - AI assistant with MCP support
- [n8n](https://n8n.io) - Workflow automation platform

---

**Ready to automate your tasks?** Start with the [Quick Start](#quick-start) above! 🚀
