# API Reference

Complete API documentation for the Vikunja MCP Server.

## Table of Contents
- [Overview](#overview)
- [Authentication](#authentication)
- [Available Tools](#available-tools)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)

---

## Overview

The Vikunja MCP Server exposes 21 tools for managing tasks, projects, labels, and more through the Model Context Protocol.

### Protocol Version
- **MCP Version**: 1.0+
- **Transport**: stdio, HTTP
- **Message Format**: JSON-RPC 2.0

---

## Authentication

All requests require a valid Vikunja API token.

### Token Types
- **User Token**: Regular user access with standard permissions
- **Admin Token**: Elevated permissions (if `RATE_LIMIT_ADMIN_BYPASS=true`)

### Providing Authentication

**Via Environment Variable:**
```bash
export VIKUNJA_API_TOKEN="your-token-here"
```

**Via MCP Request:**
```json
{
  "auth": {
    "token": "your-token-here"
  }
}
```

---

## Available Tools

### Project Management

#### `create_project`
Create a new project in Vikunja.

**Input Schema:**
```typescript
{
  title: string;           // Required: Project title
  description?: string;    // Optional: Project description
  color?: string;         // Optional: Hex color (e.g., "1973ff")
  is_archived?: boolean;  // Optional: Archive status (default: false)
}
```

**Example:**
```json
{
  "tool": "create_project",
  "arguments": {
    "title": "Website Redesign",
    "description": "Complete redesign of company website",
    "color": "1973ff"
  }
}
```

**Response:**
```json
{
  "id": 42,
  "title": "Website Redesign",
  "description": "Complete redesign of company website",
  "color": "1973ff",
  "is_archived": false,
  "created": "2025-10-17T12:00:00Z",
  "updated": "2025-10-17T12:00:00Z"
}
```

---

#### `update_project`
Update an existing project.

**Input Schema:**
```typescript
{
  project_id: number;      // Required: Project ID
  title?: string;          // Optional: New title
  description?: string;    // Optional: New description
  color?: string;         // Optional: New color
}
```

**Example:**
```json
{
  "tool": "update_project",
  "arguments": {
    "project_id": 42,
    "title": "Website Redesign v2",
    "color": "ff6b35"
  }
}
```

---

#### `delete_project`
Delete a project permanently.

**Input Schema:**
```typescript
{
  project_id: number;  // Required: Project ID to delete
}
```

**Example:**
```json
{
  "tool": "delete_project",
  "arguments": {
    "project_id": 42
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Project deleted successfully"
}
```

---

#### `archive_project`
Archive or unarchive a project.

**Input Schema:**
```typescript
{
  project_id: number;     // Required: Project ID
  is_archived: boolean;   // Required: Archive status
}
```

**Example:**
```json
{
  "tool": "archive_project",
  "arguments": {
    "project_id": 42,
    "is_archived": true
  }
}
```

---

### Task Management

#### `create_task`
Create a new task in a project.

**Input Schema:**
```typescript
{
  project_id: number;           // Required: Parent project ID
  title: string;                // Required: Task title
  description?: string;         // Optional: Task description
  due_date?: string;           // Optional: ISO 8601 date (e.g., "2025-10-24T10:00:00Z")
  priority?: number;           // Optional: Priority 1-5 (1=low, 5=urgent)
  labels?: number[];           // Optional: Array of label IDs
  assignees?: number[];        // Optional: Array of user IDs
  start_date?: string;         // Optional: ISO 8601 date
  end_date?: string;           // Optional: ISO 8601 date
  repeat_after?: number;       // Optional: Repeat interval in seconds
  percent_done?: number;       // Optional: Completion percentage (0-100)
}
```

**Example:**
```json
{
  "tool": "create_task",
  "arguments": {
    "project_id": 1,
    "title": "Design homepage mockup",
    "description": "Create initial design for new homepage",
    "due_date": "2025-10-24T17:00:00Z",
    "priority": 4,
    "labels": [1, 3],
    "assignees": [5]
  }
}
```

**Response:**
```json
{
  "id": 123,
  "title": "Design homepage mockup",
  "description": "Create initial design for new homepage",
  "done": false,
  "priority": 4,
  "labels": [
    {"id": 1, "title": "Design", "hex_color": "1973ff"},
    {"id": 3, "title": "High Priority", "hex_color": "ff6b35"}
  ],
  "assignees": [
    {"id": 5, "username": "designer1", "name": "Jane Designer"}
  ],
  "due_date": "2025-10-24T17:00:00Z",
  "created": "2025-10-17T12:00:00Z",
  "updated": "2025-10-17T12:00:00Z"
}
```

---

#### `update_task`
Update an existing task.

**Input Schema:**
```typescript
{
  task_id: number;             // Required: Task ID
  title?: string;              // Optional: New title
  description?: string;        // Optional: New description
  due_date?: string;          // Optional: New due date (ISO 8601)
  priority?: number;          // Optional: New priority (1-5)
  done?: boolean;             // Optional: Completion status
  percent_done?: number;      // Optional: Completion percentage
}
```

**Example:**
```json
{
  "tool": "update_task",
  "arguments": {
    "task_id": 123,
    "priority": 5,
    "due_date": "2025-10-23T17:00:00Z"
  }
}
```

---

#### `complete_task`
Mark a task as complete.

**Input Schema:**
```typescript
{
  task_id: number;  // Required: Task ID to complete
}
```

**Example:**
```json
{
  "tool": "complete_task",
  "arguments": {
    "task_id": 123
  }
}
```

---

#### `delete_task`
Delete a task permanently.

**Input Schema:**
```typescript
{
  task_id: number;  // Required: Task ID to delete
}
```

---

#### `move_task`
Move a task to a different project.

**Input Schema:**
```typescript
{
  task_id: number;        // Required: Task ID to move
  target_project_id: number;  // Required: Destination project ID
}
```

**Example:**
```json
{
  "tool": "move_task",
  "arguments": {
    "task_id": 123,
    "target_project_id": 2
  }
}
```

---

### Assignment & Labels

#### `assign_task`
Assign a user to a task.

**Input Schema:**
```typescript
{
  task_id: number;   // Required: Task ID
  user_id: number;   // Required: User ID to assign
}
```

**Example:**
```json
{
  "tool": "assign_task",
  "arguments": {
    "task_id": 123,
    "user_id": 5
  }
}
```

---

#### `unassign_task`
Remove a user from a task.

**Input Schema:**
```typescript
{
  task_id: number;   // Required: Task ID
  user_id: number;   // Required: User ID to unassign
}
```

---

#### `add_label`
Add a label to a task.

**Input Schema:**
```typescript
{
  task_id: number;   // Required: Task ID
  label_id: number;  // Required: Label ID
}
```

---

#### `remove_label`
Remove a label from a task.

**Input Schema:**
```typescript
{
  task_id: number;   // Required: Task ID
  label_id: number;  // Required: Label ID to remove
}
```

---

#### `create_label`
Create a new label.

**Input Schema:**
```typescript
{
  title: string;         // Required: Label title
  description?: string;  // Optional: Label description
  hex_color?: string;   // Optional: Hex color without # (e.g., "1973ff")
}
```

**Example:**
```json
{
  "tool": "create_label",
  "arguments": {
    "title": "Bug",
    "description": "Bug reports and fixes",
    "hex_color": "ff0000"
  }
}
```

---

### Search & Filtering

#### `search_tasks`
Search for tasks with advanced filtering.

**Input Schema:**
```typescript
{
  query: string;               // Required: Search query
  done?: boolean;             // Optional: Filter by completion status
  priority?: number;          // Optional: Filter by priority (1-5)
  labels?: number[];          // Optional: Filter by label IDs (AND logic)
  assignees?: number[];       // Optional: Filter by assignee IDs (OR logic)
  page?: number;              // Optional: Page number (default: 1)
  per_page?: number;          // Optional: Results per page (default: 50, max: 100)
}
```

**Example:**
```json
{
  "tool": "search_tasks",
  "arguments": {
    "query": "design",
    "done": false,
    "priority": 4,
    "labels": [1],
    "page": 1,
    "per_page": 20
  }
}
```

**Response:**
```json
{
  "tasks": [
    {
      "id": 123,
      "title": "Design homepage mockup",
      "priority": 4,
      "done": false,
      "labels": [{"id": 1, "title": "Design"}]
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}
```

---

#### `search_projects`
Search for projects.

**Input Schema:**
```typescript
{
  query: string;          // Required: Search query
  is_archived?: boolean;  // Optional: Filter by archive status
  page?: number;          // Optional: Page number
  per_page?: number;      // Optional: Results per page
}
```

---

#### `get_my_tasks`
Get all tasks assigned to the current user.

**Input Schema:**
```typescript
{
  done?: boolean;        // Optional: Filter by completion status
  page?: number;         // Optional: Page number
  per_page?: number;     // Optional: Results per page
}
```

**Example:**
```json
{
  "tool": "get_my_tasks",
  "arguments": {
    "done": false,
    "per_page": 50
  }
}
```

---

#### `get_project_tasks`
Get all tasks in a specific project.

**Input Schema:**
```typescript
{
  project_id: number;    // Required: Project ID
  done?: boolean;        // Optional: Filter by completion status
  priority?: number;     // Optional: Filter by priority
  page?: number;         // Optional: Page number
  per_page?: number;     // Optional: Results per page
}
```

---

### Bulk Operations

#### `bulk_update_tasks`
Update multiple tasks at once (max 100 tasks).

**Input Schema:**
```typescript
{
  task_ids: number[];        // Required: Array of task IDs (max 100)
  updates: {                 // Required: Updates to apply
    priority?: number;
    due_date?: string;
    done?: boolean;
    // ... any task field
  };
}
```

**Example:**
```json
{
  "tool": "bulk_update_tasks",
  "arguments": {
    "task_ids": [123, 124, 125],
    "updates": {
      "priority": 5,
      "due_date": "2025-10-20T17:00:00Z"
    }
  }
}
```

**Response:**
```json
{
  "success_count": 3,
  "failed_count": 0,
  "results": [
    {"task_id": 123, "success": true},
    {"task_id": 124, "success": true},
    {"task_id": 125, "success": true}
  ]
}
```

---

#### `bulk_complete_tasks`
Mark multiple tasks as complete (max 100 tasks).

**Input Schema:**
```typescript
{
  task_ids: number[];  // Required: Array of task IDs (max 100)
}
```

---

#### `bulk_assign_tasks`
Assign a user to multiple tasks (max 100 tasks).

**Input Schema:**
```typescript
{
  task_ids: number[];  // Required: Array of task IDs (max 100)
  user_id: number;     // Required: User ID to assign
}
```

---

#### `bulk_add_labels`
Add a label to multiple tasks (max 100 tasks).

**Input Schema:**
```typescript
{
  task_ids: number[];  // Required: Array of task IDs (max 100)
  label_id: number;    // Required: Label ID to add
}
```

---

## Error Handling

### Error Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "Permission denied",
    "data": {
      "type": "PermissionError",
      "details": "User does not have access to project"
    }
  }
}
```

### Error Codes

| Code | Type | Description |
|------|------|-------------|
| `-32700` | Parse Error | Invalid JSON |
| `-32600` | Invalid Request | Invalid JSON-RPC request |
| `-32601` | Method Not Found | Tool does not exist |
| `-32602` | Invalid Params | Invalid tool arguments |
| `-32603` | Internal Error | Server internal error |
| `-32000` | Application Error | Business logic error (see subtypes) |

### Error Subtypes

- **AuthenticationError**: Invalid or missing API token
- **PermissionError**: User lacks required permissions
- **NotFoundError**: Resource (task/project) not found
- **ValidationError**: Invalid input data
- **RateLimitError**: Rate limit exceeded

**Example:**
```json
{
  "error": {
    "code": -32000,
    "message": "Rate limit exceeded",
    "data": {
      "type": "RateLimitError",
      "limit": 100,
      "remaining": 0,
      "reset_at": "2025-10-17T12:01:00Z"
    }
  }
}
```

---

## Rate Limiting

### Configuration

Rate limits are configured via environment variables:

```env
RATE_LIMIT_DEFAULT=100      # Requests per minute
RATE_LIMIT_BURST=120        # Burst allowance
RATE_LIMIT_ADMIN_BYPASS=false  # Admin token bypass
```

### Rate Limit Headers

Response headers include rate limit information:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 85
X-RateLimit-Reset: 1697551260
```

### Handling Rate Limits

When rate limited, you'll receive a 429 status or RateLimitError:

```json
{
  "error": {
    "code": -32000,
    "message": "Rate limit exceeded. Retry after 45 seconds.",
    "data": {
      "type": "RateLimitError",
      "limit": 100,
      "remaining": 0,
      "reset_at": "2025-10-17T12:01:00Z"
    }
  }
}
```

**Best Practices:**
1. Implement exponential backoff
2. Respect `reset_at` timestamp
3. Use bulk operations when possible
4. Cache results when appropriate

---

## Usage Tips

### Pagination

For search and list operations:

```json
{
  "tool": "search_tasks",
  "arguments": {
    "query": "design",
    "page": 1,
    "per_page": 50
  }
}
```

**Response includes pagination metadata:**
```json
{
  "tasks": [...],
  "total": 150,
  "page": 1,
  "per_page": 50,
  "total_pages": 3
}
```

### Batch Processing

Use bulk operations for efficiency:

```javascript
// ❌ Inefficient (100 requests)
for (const taskId of taskIds) {
  await callTool('complete_task', { task_id: taskId });
}

// ✅ Efficient (1 request)
await callTool('bulk_complete_tasks', { task_ids: taskIds });
```

### Error Recovery

Implement retry logic for transient errors:

```javascript
async function callToolWithRetry(tool, args, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await callTool(tool, args);
    } catch (error) {
      if (error.type === 'RateLimitError') {
        await sleep(error.reset_at - Date.now());
        continue;
      }
      if (error.code === -32603 && i < maxRetries - 1) {
        await sleep(1000 * Math.pow(2, i)); // Exponential backoff
        continue;
      }
      throw error;
    }
  }
}
```

---

## Next Steps

- **Integration Guide**: See [INTEGRATIONS.md](./INTEGRATIONS.md)
- **Deployment**: See [DEPLOYMENT.md](./DEPLOYMENT.md)
- **Examples**: See [EXAMPLES.md](./EXAMPLES.md)
