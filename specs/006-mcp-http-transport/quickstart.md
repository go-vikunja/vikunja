# Quickstart: Connecting to MCP HTTP Transport

**Feature**: MCP HTTP/SSE Transport  
**Audience**: Developers integrating n8n, Python scripts, or other external clients with Vikunja MCP server

## Overview

This guide shows how to connect external clients to the Vikunja MCP server using HTTP/SSE transport. After completing this guide, you'll be able to execute Vikunja operations (create tasks, manage projects, etc.) from remote automation tools.

**Prerequisites**:
- Vikunja MCP server deployed with HTTP transport (`TRANSPORT_TYPE=http`)
- Vikunja API token (obtain from Vikunja UI: Settings → API Tokens)
- Server address and port (e.g., `http://192.168.1.100:3010`)

**Time to Complete**: 10 minutes

---

## Quick Reference

| Client | Connection Method | Documentation Link |
|--------|------------------|-------------------|
| **n8n** | HTTP POST with SSE | [n8n Section](#connecting-from-n8n) |
| **Python (MCP SDK)** | `SSEServerTransport` | [Python Section](#connecting-from-python-mcp-sdk) |
| **curl (Testing)** | Manual SSE stream | [curl Section](#testing-with-curl) |
| **JavaScript/TypeScript** | Fetch API + EventSource | [JavaScript Section](#connecting-from-javascripttypescript) |

---

## Connecting from n8n

n8n is a workflow automation tool that can call HTTP APIs. Use the **HTTP Request** node to connect to MCP server.

### Step 1: Create HTTP Request Node

1. In n8n workflow, add **HTTP Request** node
2. Set **Method**: `POST`
3. Set **URL**: `http://YOUR_SERVER_IP:3010/sse`
4. Enable **Stream Response** (under Options)

### Step 2: Add Authentication

**Option A: Authorization Header (Recommended)**

1. Under **Authentication** section, select **Generic Credential Type**
2. Choose **Header Auth**
3. Add header:
   - **Name**: `Authorization`
   - **Value**: `Bearer YOUR_VIKUNJA_API_TOKEN`

**Option B: Query Parameter**

1. Under **Query Parameters**, add:
   - **Name**: `token`
   - **Value**: `YOUR_VIKUNJA_API_TOKEN`

### Step 3: Send MCP Protocol Message

1. Under **Body**, select **JSON**
2. Add MCP tool call message:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_task",
    "arguments": {
      "title": "New task from n8n",
      "project_id": 1
    }
  }
}
```

### Step 4: Process Response

1. Response arrives as SSE stream with MCP protocol messages
2. Parse JSON from `data:` fields in SSE events
3. Extract `result.content[0].text` for tool execution result

### Example n8n Workflow

```json
{
  "nodes": [
    {
      "name": "Create Vikunja Task",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "method": "POST",
        "url": "http://192.168.1.100:3010/sse",
        "authentication": "genericCredentialType",
        "genericAuthType": "headerAuth",
        "headerAuth": {
          "name": "Authorization",
          "value": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        },
        "options": {
          "response": {
            "stream": true
          }
        },
        "bodyParametersJson": {
          "jsonrpc": "2.0",
          "id": 1,
          "method": "tools/call",
          "params": {
            "name": "create_task",
            "arguments": {
              "title": "{{ $json.taskTitle }}",
              "project_id": 1
            }
          }
        }
      }
    }
  ]
}
```

---

## Connecting from Python (MCP SDK)

Use the official MCP Python SDK with `SSEServerTransport` for type-safe MCP integration.

### Step 1: Install MCP SDK

```bash
pip install modelcontextprotocol
```

### Step 2: Create Client Script

```python
import asyncio
from modelcontextprotocol.client import Client
from modelcontextprotocol.client.sse import SSEServerTransport

# Your Vikunja configuration
VIKUNJA_MCP_URL = "http://192.168.1.100:3010/sse"
VIKUNJA_API_TOKEN = "YOUR_VIKUNJA_API_TOKEN"

async def connect_to_vikunja_mcp():
    """Connect to Vikunja MCP server via HTTP/SSE"""
    
    # Create SSE transport with authentication
    transport = SSEServerTransport(
        url=VIKUNJA_MCP_URL,
        headers={
            "Authorization": f"Bearer {VIKUNJA_API_TOKEN}"
        }
    )
    
    # Initialize MCP client
    async with Client(transport) as client:
        # Initialize protocol handshake
        await client.initialize()
        
        # List available tools
        tools = await client.list_tools()
        print(f"Available tools: {[tool.name for tool in tools]}")
        
        # Call a tool (create task example)
        result = await client.call_tool(
            name="create_task",
            arguments={
                "title": "New task from Python",
                "project_id": 1,
                "description": "Created via MCP SDK"
            }
        )
        
        print(f"Task created: {result}")

# Run the async function
asyncio.run(connect_to_vikunja_mcp())
```

### Step 3: Run the Script

```bash
python vikunja_mcp_client.py
```

### Advanced: Error Handling

```python
from modelcontextprotocol.client import ClientError

async def safe_connect():
    try:
        transport = SSEServerTransport(
            url=VIKUNJA_MCP_URL,
            headers={"Authorization": f"Bearer {VIKUNJA_API_TOKEN}"}
        )
        
        async with Client(transport) as client:
            await client.initialize()
            
            # Execute with error handling
            try:
                result = await client.call_tool(
                    name="create_task",
                    arguments={"title": "Test", "project_id": 1}
                )
                return result
            except ClientError as e:
                print(f"Tool execution failed: {e}")
                return None
                
    except ConnectionError:
        print("Failed to connect to MCP server. Check URL and token.")
    except Exception as e:
        print(f"Unexpected error: {e}")

asyncio.run(safe_connect())
```

---

## Testing with curl

Quick test to verify HTTP transport is working:

### Step 1: Establish SSE Connection

```bash
curl -N -H "Authorization: Bearer YOUR_VIKUNJA_API_TOKEN" \
     -H "Accept: text/event-stream" \
     -X POST http://192.168.1.100:3010/sse
```

**Expected Response** (SSE stream):
```
event: connected
data: {"connectionId": "123e4567-e89b-12d3-a456-426614174000"}

```

### Step 2: Send MCP Message (Advanced)

```bash
# Note: curl doesn't support bidirectional SSE easily
# Use Python/JavaScript for actual tool calls
# This example shows initial connection only
```

---

## Connecting from JavaScript/TypeScript

Use native `fetch` API with `EventSource` polyfill for SSE streaming.

### Step 1: Install Dependencies (Node.js)

```bash
npm install eventsource
```

### Step 2: Create Client

```typescript
import EventSource from 'eventsource';

const VIKUNJA_MCP_URL = 'http://192.168.1.100:3010/sse';
const VIKUNJA_API_TOKEN = 'YOUR_VIKUNJA_API_TOKEN';

// Establish SSE connection with authentication
const eventSource = new EventSource(VIKUNJA_MCP_URL, {
  headers: {
    'Authorization': `Bearer ${VIKUNJA_API_TOKEN}`
  }
});

// Handle connection established
eventSource.addEventListener('connected', (event) => {
  const data = JSON.parse(event.data);
  console.log('Connected:', data.connectionId);
});

// Handle MCP messages
eventSource.addEventListener('message', (event) => {
  const message = JSON.parse(event.data);
  console.log('MCP Message:', message);
});

// Handle errors
eventSource.onerror = (error) => {
  console.error('Connection error:', error);
  eventSource.close();
};

// Send MCP tool call (requires separate HTTP POST)
async function callTool(name: string, args: object) {
  const response = await fetch(`${VIKUNJA_MCP_URL}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${VIKUNJA_API_TOKEN}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      id: Date.now(),
      method: 'tools/call',
      params: { name, arguments: args }
    })
  });
  
  return response.json();
}

// Example: Create task
callTool('create_task', {
  title: 'Task from JavaScript',
  project_id: 1
}).then(result => console.log(result));
```

---

## Troubleshooting

### Connection Refused

**Symptom**: `curl: (7) Failed to connect to 192.168.1.100 port 3010`

**Solutions**:
1. Verify MCP server is running: `systemctl status vikunja-mcp-blue`
2. Check port is correct (blue: 3010, green: 3011)
3. Verify firewall allows connections: `sudo ufw allow 3010`
4. Test from server itself: `curl http://localhost:3010/health`

### 401 Unauthorized

**Symptom**: `{"error":"Unauthorized","message":"Invalid token"}`

**Solutions**:
1. Verify token is correct (copy from Vikunja UI)
2. Check token hasn't expired
3. Test token with Vikunja API: `curl -H "Authorization: Bearer TOKEN" http://SERVER:3456/api/v1/user`
4. Ensure token is passed correctly (header vs query param)

### 503 Service Unavailable

**Symptom**: `{"error":"Service Unavailable","message":"Server is shutting down"}`

**Solutions**:
1. Check server logs: `journalctl -u vikunja-mcp-blue -n 50`
2. Verify dependencies are running: Redis (`systemctl status redis`), Vikunja backend
3. Check connection limit (max 50 concurrent)
4. Retry with exponential backoff

### No Response from SSE Stream

**Symptom**: Connection established but no events received

**Solutions**:
1. Verify MCP protocol messages are correct (use `jsonrpc: "2.0"`)
2. Check tool name is valid: List tools first with `tools/list` method
3. Review server logs for tool execution errors
4. Ensure request body is properly formatted JSON

---

## Available MCP Tools

Once connected, you can call these tools:

| Tool Name | Description | Arguments |
|-----------|-------------|-----------|
| `create_task` | Create new task | `title`, `project_id`, `description` (optional) |
| `update_task` | Update existing task | `task_id`, `title`, `done`, etc. |
| `list_tasks` | Get tasks for project | `project_id`, `filter` (optional) |
| `create_project` | Create new project | `title`, `description` (optional) |
| `list_projects` | Get all projects | None |
| `assign_task` | Assign task to user | `task_id`, `user_id` |

**Full tool documentation**: Call `tools/list` method for complete list with schemas.

---

## Next Steps

1. **Review API Contract**: See [`contracts/sse-transport.openapi.yaml`](./contracts/sse-transport.openapi.yaml) for complete API specification
2. **Explore Tools**: Use `tools/list` MCP method to discover all available operations
3. **Build Workflows**: Integrate MCP calls into your automation pipelines
4. **Monitor Usage**: Check server logs for tool execution history and errors

## Support

- **MCP Server Logs**: `journalctl -u vikunja-mcp-blue -f`
- **Vikunja API Docs**: https://vikunja.io/docs/api/
- **MCP Protocol Spec**: https://spec.modelcontextprotocol.io/

---

**Time Spent**: ~10 minutes  
**Success Criteria**: Connected client can execute `create_task` tool and receive response
