# Exploring OpenClaw/NanoClaw Integration with Vikunja

## What Are OpenClaw and NanoClaw?

**OpenClaw** is an open-source AI agent framework (the fastest-growing GitHub project in history as of early 2026). It connects to messaging apps and enables LLMs to autonomously perform tasks. It exposes an HTTP Agent API with endpoints like `/api/agent/createTask`, `/api/agent/updateTask`, etc., and has 100+ preconfigured "AgentSkills."

**NanoClaw** is a lightweight, security-focused alternative built on Anthropic's Claude Agent SDK. It prioritizes containerized isolation and radical simplicity (~500 lines of core code vs OpenClaw's ~500k). It uses a task-scheduler with cron/interval scheduling and isolated container sandboxes per group.

Both frameworks can connect to external services, execute code, and manage tasks autonomously.

---

## Design Principle

**Assigning an agent to a task should feel exactly like assigning a coworker.**

- Bot users appear in the same assignee picker as human users
- You assign them the same way — search, click, done
- Progress shows up as regular comments in the task thread
- The task gets marked done when the bot finishes
- The only visual difference is a small bot badge next to the name

No separate "agent panels," no special buttons. The existing assignee UX *is* the interface.

---

## Integration Concept

The core idea: **assign an AI agent (via OpenClaw or NanoClaw) to a Vikunja task, and the agent autonomously works to complete it.**

This requires two directions of communication:

1. **Vikunja → Agent**: Assigning a bot user to a task triggers a dispatch to the agent's endpoint
2. **Agent → Vikunja**: The agent calls back via the standard Vikunja API to post comments, update the task, attach files, and mark it done

---

## What Already Exists in Vikunja That Helps

### Task Assignees (the primary UX surface)
- Tasks can have multiple assignees (users)
- Assignment dispatches `TaskAssigneeCreatedEvent`
- Existing assignee picker UI with search
- **This becomes the integration trigger** — assigning a bot user dispatches work to the agent

### REST API + API Tokens (inbound control)
- Full CRUD API for tasks, comments, attachments, assignees
- Scoped API tokens (`tk_` prefix) with per-route permissions
- **This handles Agent → Vikunja updates** — the bot user gets its own API token

### Event System (Watermill-based)
- Async event dispatch with retry logic
- Existing task lifecycle events (created, updated, deleted, assignee changes)
- Listener registration pattern for extending behavior
- **New listener on `TaskAssigneeCreatedEvent`** checks if assignee is a bot and dispatches

### Cron System
- `robfig/cron/v3` for scheduled background work
- Used for reminders, overdue checks, cleanup jobs
- **Can monitor running agent tasks** for timeouts/failures

### Webhooks (outbound notifications)
- Project-level and user-level webhooks already exist
- HMAC-SHA256 signing, Basic Auth support
- Could supplement the integration but not the primary mechanism

---

## What Would Need to Change

### 1. Bot User Identity (`pkg/user/user.go`)

Add `IsBot` flag to the User struct:
```go
IsBot bool `xorm:"bool default false" json:"is_bot"`
```

Bot users are real users with a special flag. They:
- Appear in the assignee picker alongside regular users
- Show a bot badge in the UI (avatar, comments, assignee lists)
- Cannot log in via the web UI — API token only
- Don't receive email notifications (reminders, overdue, etc.)
- Don't need email confirmation, password, or TOTP
- Have their own audit trail — every action is attributed to the bot

**Database migration:** Add `is_bot` column to `users` table.

### 2. Agent Connection Config on Bot Users (`pkg/user/user.go` or new model)

Bot users need agent connection details. Two approaches:

**Option A: Fields on the User model** (simpler)
```go
AgentEndpointURL string `xorm:"text null" json:"agent_endpoint_url,omitempty"`
AgentAPIKey      string `xorm:"text null" json:"-"` // never exposed in API responses
AgentType        string `xorm:"varchar(50) null" json:"agent_type,omitempty"` // "openclaw", "nanoclaw"
```

**Option B: Separate `bot_configs` table** (cleaner separation)
```go
type BotConfig struct {
    ID          int64
    UserID      int64  // FK to users, unique (one config per bot user)
    EndpointURL string
    APIKey      string // encrypted at rest
    Type        string // "openclaw", "nanoclaw"
    Created     time.Time
    Updated     time.Time
}
```

**Recommendation: Option B** — keeps the User model clean, allows the config to be managed independently, and makes it easier to encrypt the API key at rest.

### 3. Agent Dispatch Listener (`pkg/models/listeners.go`)

New listener on `TaskAssigneeCreatedEvent`:

```
When a bot user is assigned to a task:
  1. Load the bot's agent config (endpoint URL, API key, type)
  2. Load full task context (title, description, comments, attachments, related tasks)
  3. Generate a scoped API token for the bot user (if one doesn't exist)
  4. POST to the agent endpoint with task details + callback token
  5. The agent starts working autonomously
```

Corresponding listener on `TaskAssigneeDeletedEvent`:
```
When a bot user is unassigned from a task:
  1. POST a cancellation request to the agent endpoint
  2. Agent stops work on that task
```

### 4. Agent Adapter Interface (`pkg/modules/agents/`)

Generic interface with provider-specific implementations:

```go
type AgentProvider interface {
    // Dispatch sends a task to the agent for processing
    Dispatch(ctx context.Context, task *models.Task, callbackURL string, callbackToken string) error
    // Cancel tells the agent to stop working on a task
    Cancel(ctx context.Context, taskID int64) error
    // Status checks if the agent is healthy
    Status(ctx context.Context) (AgentStatus, error)
}
```

Implementations:
- `pkg/modules/agents/openclaw/` — OpenClaw HTTP API adapter
- `pkg/modules/agents/nanoclaw/` — NanoClaw adapter

### 5. Bot User Creation API

New endpoint or flag on existing user creation:

- `PUT /api/v1/bots` — create a bot user (admin or project admin only)
  - Takes: name, username, agent type, endpoint URL, API key
  - Creates user with `is_bot: true`
  - Creates associated `BotConfig`
  - Auto-generates an API token for the bot
  - Returns the bot user + token (token shown only once)

- `GET /api/v1/bots` — list bot users accessible to current user
- `POST /api/v1/bots/{id}` — update bot config
- `DELETE /api/v1/bots/{id}` — deactivate bot user

### 6. Frontend Changes

**Minimal — that's the point.** The UX piggybacks on existing patterns:

#### Assignee Picker (modify existing)
- Bot users already appear in the user list (they're real users)
- Add a small bot icon/badge next to bot user names
- No other changes needed — assignment works identically

#### User/Avatar Display (`frontend/src/components/misc/User.vue`)
- Show bot badge (small robot icon) on avatar
- Applied everywhere users are displayed: assignee lists, comments, activity log

#### Task Comments
- Comments from bot users get a subtle "bot" indicator
- No separate rendering — same comment thread, same layout
- Users reply to bot comments naturally (the agent sees replies via API polling or webhooks)

#### Bot Management UI (new, minimal)
- Settings page for creating/managing bot users
- Form: name, username, agent type, endpoint URL, API key
- Shows the generated API token once on creation
- Could live under team/workspace settings

#### Model Types (`frontend/src/modelTypes/IUser.ts`)
- Add `isBot: boolean` to `IUser`
- Add `agentType?: string` to `IUser`

### 7. Guard Rails & Edge Cases

**Authentication:**
- Bot users cannot authenticate via username/password or OIDC
- Only API token auth is valid for bots
- Login endpoint rejects users with `is_bot: true`

**Notifications:**
- Skip email notifications for bot users (reminders, overdue, mentions)
- Bot users shouldn't trigger "user mentioned" notifications when they @-mention someone? Or should they? (configurable)

**Permissions:**
- Bot users need project access just like regular users (added to project/team)
- Bot actions are scoped by the same permission system
- Bot can only modify tasks in projects it has write access to

**Rate Limiting:**
- Consider rate limits on bot API calls to prevent runaway agents
- Configurable per bot or globally

**Failure Handling:**
- Cron job checks for stalled agent tasks (assigned to bot, no activity for X minutes)
- Notify the user who assigned the bot if the agent appears stuck
- Auto-unassign after configurable timeout

---

## The User Experience, End to End

### Setup (one-time)
1. Admin goes to Settings → Bots
2. Creates a bot: "CodeReview Bot", type: OpenClaw, endpoint: `https://openclaw.example.com/api/agent`
3. Vikunja creates a bot user, generates an API token, shows it once
4. Admin adds the bot user to the relevant project(s)/team(s)

### Daily Use
1. User creates a task: "Review PR #42 for security issues"
2. User clicks the assignee picker, sees both coworkers and bots
3. User assigns "CodeReview Bot" — looks just like assigning a coworker
4. Behind the scenes: Vikunja dispatches the task to OpenClaw
5. Minutes later, comments start appearing on the task from the bot:
   - "Started reviewing PR #42..."
   - "Found 2 potential issues: SQL injection in `user_query.go:45`, missing auth check in `api/handler.go:112`"
   - "Full report attached."
6. Bot attaches a detailed report file
7. Bot marks the task as done
8. User sees it all in the normal task detail view — no special UI needed

### Collaboration
- User can reply to bot comments with clarifications
- User can unassign the bot to stop it
- User can assign a different bot or a human to take over
- Multiple bots can be assigned (one researches, another implements)

---

## Implementation Phases

### Phase 1: Bot User Foundation
- Add `is_bot` field to User model + migration
- Add `BotConfig` model + migration
- Bot creation API endpoint
- Skip email/password/TOTP requirements for bots
- Block web login for bots
- Skip email notifications for bots
- Frontend: `isBot` on IUser, bot badge on User.vue

**~8-12 files changed**

### Phase 2: Agent Dispatch
- Agent provider interface (`pkg/modules/agents/`)
- OpenClaw adapter
- NanoClaw adapter
- Listener on `TaskAssigneeCreatedEvent` to dispatch to agent
- Listener on `TaskAssigneeDeletedEvent` to cancel agent work
- Scoped API token auto-generation for bots

**~6-10 new files**

### Phase 3: Monitoring & Resilience
- Cron job for stalled agent detection
- Notify assigning user on agent failure
- Agent health check endpoint
- Rate limiting for bot API calls

**~3-5 files**

### Phase 4: Polish
- Bot management UI in frontend settings
- Bot comment styling
- Configuration for agent dispatch behavior (what context to send, timeouts)
- Documentation

**~4-6 files**

---

## Suggested File Structure

```
pkg/
├── user/
│   └── user.go                          # + IsBot field
├── models/
│   ├── bot_config.go                    # BotConfig model (endpoint, API key, type)
│   ├── bot_config_permissions.go        # BotConfig permissions
│   ├── events.go                        # + BotAssignedToTaskEvent
│   └── listeners.go                     # + bot dispatch listener
├── modules/
│   └── agents/
│       ├── agent.go                     # AgentProvider interface
│       ├── dispatch.go                  # Dispatch logic (called by listener)
│       ├── openclaw/
│       │   └── openclaw.go              # OpenClaw adapter
│       └── nanoclaw/
│           └── nanoclaw.go              # NanoClaw adapter
├── routes/
│   └── api/v1/
│       └── bots.go                      # Bot CRUD endpoints

frontend/src/
├── modelTypes/
│   └── IUser.ts                         # + isBot, agentType
├── models/
│   └── user.ts                          # + isBot default
├── components/
│   └── misc/
│       └── User.vue                     # + bot badge
├── views/
│   └── settings/
│       └── Bots.vue                     # Bot management page
```

---

## Key Design Decisions

| Decision | Recommendation | Rationale |
|----------|---------------|-----------|
| Bot identity | `is_bot` flag on User | Bots are first-class citizens, appear in all user contexts naturally |
| Agent config storage | Separate `BotConfig` table | Clean separation, easier to encrypt API keys |
| UX surface | Existing assignee picker | "Assign work to a coworker" feeling — no new concepts to learn |
| Dispatch trigger | `TaskAssigneeCreatedEvent` listener | Piggybacks on existing event system, zero new UI needed |
| Agent communication | Standard Vikunja REST API | Bot uses same API as any other client — comments, updates, attachments |
| Provider support | Adapter pattern with interface | Support OpenClaw and NanoClaw (and future providers) cleanly |
| Progress reporting | Task comments | Shows up naturally in the existing task detail view |
| Completion | Bot marks task done via API | Same as a human marking it done — no special mechanism |

---

## Summary

The integration is built on one core insight: **bot users are users.** By adding an `is_bot` flag and wiring up agent dispatch on assignment, the entire existing Vikunja UX — assignee picker, comments, task status, attachments — becomes the agent interface. No new UI paradigms needed.

The new code is focused on:
1. Bot user identity (`is_bot` flag + `BotConfig`)
2. Agent dispatch (listener + provider adapters)
3. Visual distinction (bot badge in frontend)

Everything else — permissions, API, events, comments, task lifecycle — already works.
