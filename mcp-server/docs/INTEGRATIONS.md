# MCP Server Integration Guide

This guide shows how to integrate the Vikunja MCP Server with various AI agent platforms.

## Table of Contents
- [Claude Desktop](#claude-desktop)
- [n8n AI Agent](#n8n-ai-agent)
- [Custom Integrations](#custom-integrations)

---

## Claude Desktop

### Prerequisites
- Claude Desktop installed
- Vikunja MCP Server running (see [DEPLOYMENT.md](./DEPLOYMENT.md))
- Valid Vikunja API token

### Configuration

1. **Locate Claude Desktop config file:**
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`

2. **Add MCP server configuration:**

```json
{
  "mcpServers": {
    "vikunja": {
      "command": "node",
      "args": ["/path/to/vikunja-mcp-server/dist/index.js"],
      "env": {
        "VIKUNJA_API_URL": "http://localhost:3456",
        "VIKUNJA_API_TOKEN": "your-vikunja-api-token-here",
        "MCP_PORT": "3457",
        "REDIS_HOST": "localhost",
        "REDIS_PORT": "6379",
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

3. **For Docker deployment:**

```json
{
  "mcpServers": {
    "vikunja": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e", "VIKUNJA_API_URL=http://host.docker.internal:3456",
        "-e", "VIKUNJA_API_TOKEN=your-token-here",
        "-p", "3457:3457",
        "vikunja-mcp-server"
      ]
    }
  }
}
```

4. **Restart Claude Desktop**

### Usage Examples

Once configured, you can interact with Vikunja through Claude:

```
You: "Create a new project called 'Website Redesign' with description 'Redesign the company website'"

Claude: I'll create that project for you using the Vikunja MCP server.
[Uses create_project tool]
✓ Created project "Website Redesign" (ID: 42)

You: "Add a task 'Design mockups' to that project, due next Friday, high priority"

Claude: I'll add that task to the Website Redesign project.
[Uses create_task tool]
✓ Created task "Design mockups" in project 42, due 2025-10-24, priority: 4
```

---

## n8n AI Agent

### Prerequisites
- n8n instance running (version with AI Agent support)
- Vikunja MCP Server running and accessible
- n8n AI Agent and Tool MCP nodes

### Setup with MCP Tool Node (Recommended)

n8n has native MCP support through the [Tool MCP node](https://docs.n8n.io/integrations/builtin/cluster-nodes/sub-nodes/n8n-nodes-langchain.toolmcp/).

#### Step-by-Step Setup:

1. **Create a new workflow in n8n**

2. **Add an AI Agent node** (requires n8n AI features)
   - This is the LangChain AI Agent that orchestrates tool usage

3. **Add a Tool MCP sub-node** to your AI Agent
   - Click "Add Tool" → Select "Tool MCP"

4. **Configure MCP Connection:**

**For Docker deployment:**
```json
{
  "command": "docker",
  "args": [
    "exec", "-i", "vikunja-mcp-server",
    "node", "/app/dist/index.js"
  ],
  "env": {
    "VIKUNJA_API_TOKEN": "your-vikunja-token-here"
  }
}
```

**For local/systemd deployment:**
```json
{
  "command": "node",
  "args": ["/path/to/mcp-server/dist/index.js"],
  "env": {
    "VIKUNJA_API_URL": "http://localhost:3456",
    "VIKUNJA_API_TOKEN": "your-vikunja-token-here"
  }
}
```

5. **Test the connection:**
   - The Tool MCP node will automatically discover all 21 Vikunja tools
   - Your AI agent can now use them based on natural language prompts

#### Example Workflow:

```
Trigger (Webhook/Schedule)
  ↓
AI Agent (with Tool MCP sub-node)
  → Prompt: "Create a task in project 1 called 'Review PR' due tomorrow"
  → AI uses create_task tool automatically
  ↓
Output task details
```

### Alternative: HTTP Requests (For workflows without AI Agent)

If you need direct tool calls without an AI agent:

1. **Create HTTP Request Node:**
   - Method: POST
   - URL: `http://your-mcp-server:3457/tools/execute`
   - Authentication: None (handled in body)

2. **Request Body Template:**

```json
{
  "tool": "create_task",
  "arguments": {
    "project_id": 1,
    "title": "{{ $json.taskTitle }}",
    "description": "{{ $json.taskDescription }}",
    "priority": 3
  },
  "auth": {
    "token": "{{ $env.VIKUNJA_API_TOKEN }}"
  }
}
```

3. **Available Tools:**
   - All 21 MCP tools - see [API.md](./API.md) for complete reference
const request = {
  jsonrpc: '2.0',
  id: 1,
  method: 'tools/call',
  params: {
    name: 'create_task',
    arguments: {
      project_id: 1,
      title: 'Task from n8n'
    }
  }
};

mcp.stdin.write(JSON.stringify(request) + '\n');
```

### Example n8n Workflow

**Workflow: Slack → Vikunja Task**

1. **Trigger**: Webhook from Slack (slash command `/task`)
2. **Extract Data**: Set node to parse Slack message
3. **Create Task**: HTTP Request to MCP server
4. **Respond**: Send success message back to Slack

```json
{
  "nodes": [
    {
      "name": "Slack Webhook",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "slack-task"
      }
    },
    {
      "name": "Parse Task",
      "type": "n8n-nodes-base.set",
      "parameters": {
        "values": {
          "string": [
            {
              "name": "taskTitle",
              "value": "={{ $json.body.text }}"
            }
          ]
        }
      }
    },
    {
      "name": "Create in Vikunja",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "method": "POST",
        "url": "http://localhost:3457/tools/execute",
        "jsonParameters": true,
        "options": {},
        "bodyParametersJson": "={\n  \"tool\": \"create_task\",\n  \"arguments\": {\n    \"project_id\": 1,\n    \"title\": \"{{ $json.taskTitle }}\"\n  },\n  \"auth\": {\n    \"token\": \"{{ $env.VIKUNJA_TOKEN }}\"\n  }\n}"
      }
    },
    {
      "name": "Respond to Slack",
      "type": "n8n-nodes-base.respondToWebhook",
      "parameters": {
        "respondWith": "text",
        "responseBody": "=Task created: {{ $json.result.title }}"
      }
    }
  ]
}
```

---

## Custom Integrations

### Python Client

```python
import asyncio
import json
from typing import Any, Dict

class VikunjaMCPClient:
    def __init__(self, api_url: str, api_token: str):
        self.api_url = api_url
        self.api_token = api_token
        
    async def call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Call an MCP tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        # Send over stdio or HTTP depending on your setup
        # This is a simplified example
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.api_url}/tools/execute",
                json=request,
                headers={"Authorization": f"Bearer {self.api_token}"}
            ) as resp:
                return await resp.json()

# Usage
async def main():
    client = VikunjaMCPClient(
        api_url="http://localhost:3457",
        api_token="your-vikunja-token"
    )
    
    # Create a task
    result = await client.call_tool("create_task", {
        "project_id": 1,
        "title": "Task from Python",
        "description": "Created via MCP",
        "priority": 3
    })
    print(f"Created task: {result}")

asyncio.run(main())
```

### JavaScript/Node.js Client

```javascript
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

class VikunjaMCPClient {
  constructor(serverPath, apiToken) {
    this.serverPath = serverPath;
    this.apiToken = apiToken;
    this.client = null;
  }

  async connect() {
    const transport = new StdioClientTransport({
      command: 'node',
      args: [this.serverPath],
      env: {
        VIKUNJA_API_TOKEN: this.apiToken,
        VIKUNJA_API_URL: 'http://localhost:3456',
      },
    });

    this.client = new Client({
      name: 'vikunja-client',
      version: '1.0.0',
    }, {
      capabilities: {},
    });

    await this.client.connect(transport);
  }

  async callTool(name, args) {
    if (!this.client) {
      throw new Error('Client not connected. Call connect() first.');
    }

    const result = await this.client.callTool({
      name,
      arguments: args,
    });

    return result;
  }

  async close() {
    if (this.client) {
      await this.client.close();
    }
  }
}

// Usage
const client = new VikunjaMCPClient(
  '/path/to/mcp-server/dist/index.js',
  'your-vikunja-token'
);

await client.connect();

// Create a project
const project = await client.callTool('create_project', {
  title: 'New Project',
  description: 'Created via MCP SDK',
});

console.log('Created project:', project);

await client.close();
```

### REST API Bridge (for non-MCP clients)

If you need to expose MCP tools via REST API:

```typescript
// Simple Express bridge
import express from 'express';
import { VikunjaMCPClient } from './mcp-client';

const app = express();
app.use(express.json());

const mcpClient = new VikunjaMCPClient(/* config */);

app.post('/api/tools/:toolName', async (req, res) => {
  try {
    const { toolName } = req.params;
    const result = await mcpClient.callTool(toolName, req.body);
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.listen(3458, () => {
  console.log('MCP Bridge API listening on port 3458');
});
```

---

## Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `VIKUNJA_API_URL` | Vikunja API endpoint | `http://localhost:3456` | Yes |
| `VIKUNJA_API_TOKEN` | Vikunja API token | - | Yes |
| `MCP_PORT` | MCP server port | `3457` | No |
| `REDIS_HOST` | Redis host for rate limiting | `localhost` | No |
| `REDIS_PORT` | Redis port | `6379` | No |
| `REDIS_PASSWORD` | Redis password | - | No |
| `RATE_LIMIT_DEFAULT` | Default rate limit (req/min) | `100` | No |
| `RATE_LIMIT_BURST` | Burst rate limit | `120` | No |
| `LOG_LEVEL` | Logging level | `info` | No |

---

## Troubleshooting

### MCP Server Not Starting

1. **Check logs**: `docker logs vikunja-mcp-server`
2. **Verify Vikunja connection**: `curl http://localhost:3456/api/v1/info`
3. **Check Redis**: `redis-cli ping`

### Authentication Failures

1. **Verify token**: Test token with Vikunja API directly
2. **Check token format**: Should be the raw token, not base64 encoded
3. **Token permissions**: Ensure token has necessary permissions

### Rate Limiting Issues

1. **Check Redis connection**: `redis-cli -h localhost -p 6379 ping`
2. **View rate limit data**: `redis-cli keys "ratelimit:*"`
3. **Adjust limits**: Set `RATE_LIMIT_DEFAULT` and `RATE_LIMIT_BURST` in env

### Tool Execution Errors

1. **Validate input**: Check tool schema in [API.md](./API.md)
2. **Check Vikunja permissions**: Ensure user has access to resources
3. **Review logs**: Set `LOG_LEVEL=debug` for detailed output

---

## Next Steps

- **Production Deployment**: See [DEPLOYMENT.md](./DEPLOYMENT.md)
- **API Reference**: See [API.md](./API.md) for all tools and schemas
- **Examples**: See [EXAMPLES.md](./EXAMPLES.md) for workflow examples
