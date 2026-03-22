# Exploring OpenClaw/NanoClaw Integration with Vikunja

## What Are OpenClaw and NanoClaw?

**OpenClaw** is an open-source AI agent framework (the fastest-growing GitHub project in history as of early 2026). It connects to messaging apps and enables LLMs to autonomously perform tasks. It exposes an HTTP Agent API with endpoints like `/api/agent/createTask`, `/api/agent/updateTask`, etc., and has 100+ preconfigured "AgentSkills."

**NanoClaw** is a lightweight, security-focused alternative built on Anthropic's Claude Agent SDK. It prioritizes containerized isolation and radical simplicity (~500 lines of core code vs OpenClaw's ~500k). It uses a task-scheduler with cron/interval scheduling and isolated container sandboxes per group.

Both frameworks can connect to external services, execute code, and manage tasks autonomously.

---

## Integration Concept

The core idea: **assign an AI agent (via OpenClaw or NanoClaw) to a Vikunja task, and the agent autonomously works to complete it.**

This requires two directions of communication:

1. **Vikunja → Agent**: "Here's a task, go do it" (task assignment triggers agent work)
2. **Agent → Vikunja**: "Here's my progress/result" (agent updates task status, adds comments, attaches files)

---

## What Already Exists in Vikunja That Helps

### Webhooks (outbound notifications)
- Project-level and user-level webhooks already exist
- Events like `task.created`, `task.updated`, `task.assignee.created` can trigger HTTP calls
- HMAC-SHA256 signing, Basic Auth support
- **This handles Vikunja → Agent notifications**

### REST API + API Tokens (inbound control)
- Full CRUD API for tasks, comments, attachments, assignees
- Scoped API tokens (`tk_` prefix) with per-route permissions
- **This handles Agent → Vikunja updates**

### Event System (Watermill-based)
- Async event dispatch with retry logic
- Existing task lifecycle events (created, updated, deleted, assignee changes)
- Listener registration pattern for extending behavior

### Cron System
- `robfig/cron/v3` for scheduled background work
- Used for reminders, overdue checks, cleanup jobs

### Task Assignees
- Tasks can have multiple assignees (users)
- Assignment dispatches `TaskAssigneeCreatedEvent`

---

## What Would Need to Change

### Option A: Lightweight / Webhook-Based Integration (Recommended Starting Point)

This approach uses existing Vikunja primitives and puts the orchestration logic in an external bridge service.

#### New Components

1. **Agent Configuration Model** (`pkg/models/agents.go`)
   - New database table `agents` storing agent connections:
     ```go
     type Agent struct {
         ID          int64
         Name        string     // "My OpenClaw Agent"
         Type        string     // "openclaw" or "nanoclaw"
         EndpointURL string     // Agent API base URL
         APIKey      string     // Auth credential for the agent
         ProjectID   int64      // Scoped to a project
         CreatedByID int64
         Created     time.Time
         Updated     time.Time
     }
     ```
   - CRUD + Permissions (project-level admin only)
   - Migration file for the new table

2. **Agent Assignment on Tasks** (`pkg/models/task_agent.go`)
   - New table `task_agents` linking tasks to agents:
     ```go
     type TaskAgent struct {
         ID       int64
         TaskID   int64
         AgentID  int64
         Status   string // "pending", "running", "completed", "failed"
         Created  time.Time
         Updated  time.Time
     }
     ```
   - When an agent is assigned to a task, Vikunja sends the task details to the agent's endpoint
   - Agent status tracked and displayed on the task

3. **Agent Dispatch Listener** (`pkg/models/listeners.go`)
   - New event listener for `TaskAgentAssignedEvent`
   - Sends HTTP POST to the agent's endpoint with task details:
     ```json
     {
       "task_id": 123,
       "title": "Fix the login bug",
       "description": "Users report...",
       "callback_url": "https://vikunja.example.com/api/v1",
       "callback_token": "tk_..."
     }
     ```
   - The callback URL and token let the agent call back into Vikunja's API

4. **Agent Callback API Token Generation**
   - When dispatching to an agent, auto-generate a scoped API token
   - Permissions limited to: update this task, add comments, add attachments
   - Token expires when task is marked done or after a configurable TTL

5. **API Routes** (`pkg/routes/api/v1/`)
   - `GET /projects/{id}/agents` — list configured agents for a project
   - `PUT /projects/{id}/agents` — add an agent configuration
   - `POST /projects/{id}/agents/{agentID}` — update agent config
   - `DELETE /projects/{id}/agents/{agentID}` — remove agent
   - `PUT /tasks/{id}/agents` — assign agent to task
   - `DELETE /tasks/{id}/agents/{agentID}` — unassign agent
   - `GET /tasks/{id}/agents` — list agents on a task with status
   - `POST /tasks/{id}/agents/{agentID}/status` — agent reports status back (webhook receiver)

6. **Frontend Changes**
   - **Agent config UI** in project settings (`frontend/src/views/project/settings/`)
     - Form to add/edit agent connections (name, type, URL, API key)
   - **Task detail agent panel** (`frontend/src/components/tasks/`)
     - Show assigned agents and their status
     - Button to assign an available agent to the task
     - Live status indicator (pending/running/completed/failed)
   - **New Pinia store** (`frontend/src/stores/agents.ts`)
   - **New service** (`frontend/src/services/agent.ts`)
   - **New model types** (`frontend/src/modelTypes/IAgent.ts`)

#### Flow

```
User assigns agent to task in Vikunja UI
    → Vikunja creates TaskAgent record
    → Vikunja generates scoped API token
    → Vikunja POSTs task details + callback token to agent endpoint
    → Agent starts working autonomously
    → Agent calls back to Vikunja API:
        - POST /tasks/{id}/comments → progress updates
        - PUT /tasks/{id} → update status, mark done
        - PUT /tasks/{id}/attachments → attach deliverables
    → Vikunja shows agent progress in task detail view
```

---

### Option B: Deep Integration (More Ambitious)

Builds on Option A with additional capabilities:

7. **Agent Status Polling Cron**
   - Background cron job checking agent health/status
   - Polls agents with `status: "running"` to detect stalls
   - Updates task status if agent becomes unreachable
   - Auto-retry logic for transient failures

8. **Agent Chat Interface**
   - Extend task comments to support "agent messages"
   - Add a `source` field to `TaskComment` ("user" vs "agent")
   - Frontend renders agent comments differently (distinct styling)
   - Users can reply to agent comments to provide guidance

9. **Agent Templates / Skills**
   - Pre-configured agent profiles for common task types
   - Map Vikunja labels/tags to agent skills
   - "Code Review" agent, "Research" agent, "Writing" agent, etc.

10. **Multi-Agent Orchestration**
    - Multiple agents on a single task (one researches, another implements)
    - Agent-to-agent handoff via task relations
    - Parent task decomposition: agent creates subtasks and assigns sub-agents

---

## Implementation Effort Estimate

### Option A (Recommended MVP)

| Component | Files to Create/Modify | Scope |
|-----------|----------------------|-------|
| Agent model + migration | 3-4 new Go files | New DB table, CRUD, permissions |
| TaskAgent model + migration | 3-4 new Go files | New DB table, CRUD, permissions |
| Agent dispatch listener | Modify `listeners.go`, new dispatch logic | HTTP client to agent |
| API token generation | Modify `api_tokens.go` | Auto-scoped token creation |
| API routes | Modify `routes.go`, new handler files | 8-10 new endpoints |
| Frontend: model + service | 2-4 new TS files | Types, API service layer |
| Frontend: store | 1 new file | Pinia store for agents |
| Frontend: project settings UI | 1-2 new Vue components | Agent config form |
| Frontend: task detail panel | 1-2 new Vue components | Agent status display |
| Events | Modify `events.go` | New agent-related events |
| i18n | Modify `en.json` | Translation strings |

**Roughly 15-25 files to create or modify.**

### Key Decisions to Make

1. **OpenClaw vs NanoClaw vs Both?**
   - OpenClaw has a documented HTTP API (`/api/agent/*` endpoints)
   - NanoClaw is simpler, built on Claude Agent SDK
   - Could support both via an adapter pattern (common interface, provider-specific implementations)
   - Recommendation: **Start with a generic interface, implement OpenClaw adapter first** since it has the richer API

2. **Security model for agent callbacks**
   - Scoped API tokens (recommended) vs. separate agent auth mechanism
   - Should agents get their own "user" identity or act as the assigning user?
   - Recommendation: **Agents act as a special "agent" user type** — auditable, distinct from human users

3. **Task context sent to agent**
   - Just the task? Task + comments? Task + related tasks? Task + project context?
   - Recommendation: **Task + description + comments**, expandable later

4. **Agent failure handling**
   - What happens when an agent fails or goes silent?
   - Recommendation: **Configurable timeout, cron-based health check, notify assigning user on failure**

5. **Where does the bridge service run?**
   - Option A keeps it inside Vikunja (direct HTTP calls to agent endpoints)
   - Could also be an external microservice that consumes Vikunja webhooks
   - Recommendation: **Inside Vikunja** for simplicity (new `pkg/modules/agents/` package)

---

## Suggested File Structure

```
pkg/
├── models/
│   ├── agents.go                    # Agent configuration model
│   ├── agents_permissions.go        # Agent permissions
│   ├── task_agents.go               # Task-Agent assignment model
│   ├── task_agents_permissions.go   # Task-Agent permissions
│   └── events.go                    # + new agent events
├── modules/
│   └── agents/
│       ├── agent.go                 # Common agent interface
│       ├── openclaw/
│       │   └── openclaw.go          # OpenClaw adapter
│       ├── nanoclaw/
│       │   └── nanoclaw.go          # NanoClaw adapter
│       └── dispatch.go              # Task dispatch logic
├── routes/
│   └── api/v1/
│       └── agents.go                # Agent API handlers

frontend/src/
├── modelTypes/
│   └── IAgent.ts                    # Agent TypeScript interfaces
├── services/
│   └── agent.ts                     # Agent API service
├── stores/
│   └── agents.ts                    # Pinia store
├── components/
│   └── tasks/
│       └── partials/
│           └── agentPanel.vue       # Agent status on task detail
├── views/
│   └── project/
│       └── settings/
│           └── agents.vue           # Agent configuration page
```

---

## Summary

The integration is very feasible because Vikunja already has the key building blocks:
- **Webhooks** for outbound notifications
- **API tokens** for secure inbound callbacks
- **Event system** for reactive dispatch
- **Cron** for health monitoring

The core new work is:
1. A new `Agent` model for storing agent configurations
2. A new `TaskAgent` model for task-agent assignments
3. HTTP dispatch logic to send tasks to agents
4. Auto-generated scoped API tokens for callbacks
5. Frontend UI for configuration and status display

Start with Option A (webhook-based, ~15-25 files), validate the concept, then expand to Option B features as needed.
