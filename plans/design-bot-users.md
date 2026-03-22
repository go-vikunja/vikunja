# Design: Bot Users

## Overview

Bot users are first-class users with an `is_bot` flag. They appear alongside human users in assignee pickers, comment threads, and everywhere else users show up. The only visual difference is a small "Bot" badge next to their name.

Bot users are always owned by a human user who created them. They cannot log in via username/password — they only interact via the API using tokens. The owning user manages the bot's API tokens.

The feature is gated behind a config flag `service.enablebotusers` (default: `false`).

---

## Data Model

### User table changes

Add one column to the existing `users` table:

```go
// In pkg/user/user.go, add to User struct:
IsBot bool `xorm:"bool default false index" json:"is_bot"`
```

This is the only schema change to `users`. Bot users are regular rows in the `users` table with `is_bot = true`.

Bot users:
- Have a `username` (required, unique as usual)
- Have a `name` (optional display name)
- Have **no password** (empty string, never hashed)
- Have **no email** (empty string — no notifications needed)
- Use `Issuer = "local"` — but we skip the password/email validation for bots (see below)
- Have `Status = StatusActive`
- Have `EmailRemindersEnabled = false`
- Have `OverdueTasksRemindersEnabled = false`

### New table: `bot_owners`

Tracks which human user owns which bot:

```go
// New file: pkg/models/bot_users.go
type BotOwner struct {
    ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
    BotID   int64     `xorm:"bigint not null unique index" json:"bot_id"`    // FK → users.id (the bot)
    OwnerID int64     `xorm:"bigint not null index" json:"owner_id"`         // FK → users.id (the human)
    Created time.Time `xorm:"created not null" json:"created"`
}

func (*BotOwner) TableName() string {
    return "bot_owners"
}
```

**Why a separate table instead of a field on `User`?**
- Keeps the User struct clean — most code doesn't need to know about bot ownership
- Allows querying "all bots owned by user X" efficiently
- Avoids a self-referential FK on the users table

---

## Config

Add to `pkg/config/config.go`:

```go
ServiceEnableBotUsers Key = `service.enablebotusers`
```

Default: `false`. Set in `InitDefaultConfig()`:

```go
ServiceEnableBotUsers.setDefault(false)
```

Add to `config-raw.json` under the `service` section:

```json
{
    "key": "enablebotusers",
    "default_value": "false",
    "comment": "If enabled, users can create bot users that interact with Vikunja via the API only."
}
```

Expose in the `/info` endpoint so the frontend knows whether to show bot-related UI:

```go
// In vikunjaInfos struct:
BotUsersEnabled bool `json:"bot_users_enabled"`

// In Info handler:
info.BotUsersEnabled = config.ServiceEnableBotUsers.GetBool()
```

---

## API Endpoints

All bot endpoints require `service.enablebotusers` to be `true`. Return `403` (or a specific error) if disabled.

### Create Bot

**`PUT /api/v1/user/bots`**

Request body:
```json
{
    "username": "my-review-bot",
    "name": "Code Review Bot"
}
```

Handler logic:
1. Check `config.ServiceEnableBotUsers.GetBool()` — return error if disabled
2. Validate username (same rules: 1-250 chars, no spaces, not reserved)
3. Create user with `IsBot = true`, no password, no email, `Status = Active`
4. Skip email confirmation flow entirely
5. Don't create an initial project for the bot (bots get added to projects manually)
6. Insert `BotOwner{BotID: newUser.ID, OwnerID: currentUser.ID}`
7. Return the created bot user

Response:
```json
{
    "id": 42,
    "username": "my-review-bot",
    "name": "Code Review Bot",
    "is_bot": true,
    "created": "2026-03-22T...",
    "updated": "2026-03-22T..."
}
```

### List My Bots

**`GET /api/v1/user/bots`**

Returns all bots owned by the current user. Supports search via `?s=` query param.

Handler logic:
1. Query `bot_owners WHERE owner_id = currentUser.ID`
2. Join with `users` to get bot details
3. Return list of bot users

### Get Bot

**`GET /api/v1/user/bots/:bot`**

Returns a single bot by ID. Must be owned by current user.

### Update Bot

**`POST /api/v1/user/bots/:bot`**

Update bot's `name` or `username`. Must be owned by current user.

Request body:
```json
{
    "name": "New Bot Name"
}
```

### Delete Bot

**`DELETE /api/v1/user/bots/:bot`**

Soft-deletes (sets `Status = StatusDisabled`) or hard-deletes the bot user. Must be owned by current user.

Considerations:
- What happens to tasks the bot is assigned to? → Unassign the bot from all tasks on deletion.
- What about comments the bot posted? → Keep them (same as when a human user is deleted).

### Create API Token for Bot

**`PUT /api/v1/user/bots/:bot/tokens`**

Creates an API token owned by the bot user, but only the owning human can call this endpoint.

Handler logic:
1. Verify current user owns this bot (check `bot_owners`)
2. Create an `APIToken` with `OwnerID = bot.ID`
3. Return the token (visible only once, same as regular token creation)

Request body:
```json
{
    "title": "Main bot token",
    "permissions": {
        "tasks": ["read_all", "update"],
        "task_comments": ["read_all", "create"]
    },
    "expires_at": "2027-01-01T00:00:00Z"
}
```

### List Bot's API Tokens

**`GET /api/v1/user/bots/:bot/tokens`**

Lists API tokens belonging to the bot. Only the owning human can view these.

### Delete Bot's API Token

**`DELETE /api/v1/user/bots/:bot/tokens/:token`**

Deletes an API token belonging to the bot. Only the owning human can call this.

---

## Auth Changes

### Block password login for bots

In `pkg/routes/api/v1/login.go`, after retrieving the user:

```go
if user.IsBot {
    return &user2.ErrAccountIsBot{UserID: user.ID}
}
```

Also in `CheckUserCredentials` (`pkg/user/user.go`) — bots have no password, so `bcrypt.CompareHashAndPassword` would fail naturally. But adding an explicit check is clearer and avoids the timing cost.

### Block CalDAV auth for bots

In `pkg/routes/caldav/auth.go`, reject bot users.

### API token auth works as-is

The existing API token middleware (`pkg/routes/api_tokens.go`) resolves the token owner and puts them in context. If the token belongs to a bot user, the bot user becomes the authenticated user. No changes needed here — the bot acts as itself.

### User creation validation bypass for bots

`checkIfUserIsValid()` in `pkg/user/user_create.go` currently requires password + username for local users and email for all. For bots:
- Skip password requirement
- Skip email requirement
- Still require username

This could be handled by either:
- **Option A**: Adding a bot-specific code path in `checkIfUserIsValid()` that checks `IsBot`
- **Option B**: Creating a separate `CreateBotUser()` function that skips those checks

**Recommendation: Option B** — a separate `CreateBotUser()` function in `pkg/user/` that handles bot-specific creation logic cleanly without polluting the existing user creation path. It would:
1. Validate username only
2. Check uniqueness (same as regular users)
3. Set `IsBot = true`, `Status = Active`, `EmailRemindersEnabled = false`, `OverdueTasksRemindersEnabled = false`
4. Insert user
5. Skip email confirmation, skip initial project creation
6. Dispatch `CreatedEvent` as usual

---

## Notification Changes

### Skip all email notifications for bots

In `User.ShouldNotify()` (`pkg/user/user.go`):

```go
func (u *User) ShouldNotify(sessions ...*xorm.Session) (bool, error) {
    // ... existing session setup ...
    user, err := getUser(s, &User{ID: u.ID}, true)
    if err != nil {
        return false, err
    }

    if user.IsBot {
        return false, nil
    }

    return user.Status != StatusDisabled && user.Status != StatusAccountLocked, err
}
```

This is comprehensive — it prevents all notification types (email, overdue reminders, etc.) for bot users.

---

## User Search / Assignee Picker Changes

### Bots appear in user search

The `ListUsers()` function in `pkg/user/users_project.go` searches by username/name. Bot users will naturally appear in search results since they're regular users. No backend changes needed.

### Bots appear in project member lists

Bot users can be added to projects via the existing project user/team mechanisms. Once added, they appear in the assignee picker for that project's tasks. No changes needed to the project membership system.

---

## Frontend Changes

### IUser type

Add `isBot` to `frontend/src/modelTypes/IUser.ts`:

```typescript
export interface IUser extends IAbstract {
    // ... existing fields ...
    isBot: boolean
}
```

### User model

Add default in `frontend/src/models/user.ts`:

```typescript
isBot = false
```

### User.vue component — Bot badge

Add a "Bot" badge next to the username when `user.isBot` is true:

```vue
<template>
    <div class="user" :class="{'is-inline': isInline}">
        <img
            v-tooltip="displayName"
            :height="avatarSize"
            :src="avatarSrc"
            :width="avatarSize"
            :alt="'Avatar of ' + displayName"
            class="avatar"
        >
        <span v-if="showUsername" class="username">{{ displayName }}</span>
        <span v-if="user.isBot" class="bot-badge">Bot</span>
    </div>
</template>
```

Style the badge as a small pill/tag:

```scss
.bot-badge {
    display: inline-flex;
    align-items: center;
    font-size: 0.65rem;
    font-weight: 600;
    padding: 0.1rem 0.35rem;
    border-radius: 0.25rem;
    background: var(--grey-200);
    color: var(--grey-600);
    margin-inline-start: 0.25rem;
    vertical-align: middle;
    text-transform: uppercase;
    letter-spacing: 0.02em;
}
```

This badge appears everywhere the `User` component is used: assignee lists, comment headers, task detail sidebar, etc.

### Info store

Add `botUsersEnabled` to the frontend's info/config store so components know whether to show bot-related UI (e.g., the "Manage Bots" settings page).

### Bot management settings page

New page at `frontend/src/views/user/settings/Bots.vue`:
- List bots owned by current user
- Create new bot (form: username, name)
- Edit bot (update name)
- Delete bot (with confirmation)
- Manage bot API tokens:
  - Create token (show once, copy to clipboard)
  - List existing tokens
  - Delete token

This page only appears in the settings sidebar when `botUsersEnabled` is `true`.

---

## Database Migration

Single migration adding:
1. `is_bot` column to `users` table (bool, default false, indexed)
2. `bot_owners` table

```go
func init() {
    migrations = append(migrations, &xormigrate.Migration{
        ID:          "<timestamp>",
        Description: "Add bot user support",
        Migrate: func(tx *xorm.Engine) error {
            // 1. Add is_bot column
            err := tx.Exec("ALTER TABLE users ADD COLUMN is_bot BOOLEAN NOT NULL DEFAULT FALSE")
            if err != nil { return err }

            err = tx.Exec("CREATE INDEX IDX_users_is_bot ON users (is_bot)")
            if err != nil { return err }

            // 2. Create bot_owners table
            type BotOwner struct {
                ID      int64     `xorm:"bigint autoincr not null unique pk"`
                BotID   int64     `xorm:"bigint not null unique index"`
                OwnerID int64     `xorm:"bigint not null index"`
                Created time.Time `xorm:"created not null"`
            }
            return tx.Sync2(&BotOwner{})
        },
        Rollback: func(tx *xorm.Engine) error {
            return nil
        },
    })
}
```

---

## Error Types

New error types in `pkg/user/error.go` (or `pkg/models/error.go`):

| Error | Code | When |
|-------|------|------|
| `ErrAccountIsBot` | TBD | Bot tries to log in with password |
| `ErrBotUsersDisabled` | TBD | Bot endpoint called when feature is off |
| `ErrBotNotOwned` | TBD | User tries to manage a bot they don't own |
| `ErrCannotMakeBotOwner` | TBD | Trying to set a bot as owner of another bot |

---

## What's NOT in scope (for now)

These are deferred to the agent dispatch phase:

- Agent endpoint configuration (OpenClaw/NanoClaw connection details)
- Auto-dispatching work when a bot is assigned to a task
- Agent status tracking
- Health monitoring / timeout cron jobs
- Bot-specific avatar generation (bots use the default avatar provider for now)

---

## Files to Create or Modify

### New Files
| File | Purpose |
|------|---------|
| `pkg/models/bot_users.go` | BotOwner model, bot CRUD logic |
| `pkg/models/bot_users_permissions.go` | Permission checks for bot operations |
| `pkg/routes/api/v1/bots.go` | Bot API endpoint handlers |
| `pkg/migration/<timestamp>.go` | Database migration |
| `frontend/src/views/user/settings/Bots.vue` | Bot management UI |
| `frontend/src/services/botUser.ts` | Bot API service |

### Modified Files
| File | Change |
|------|--------|
| `pkg/user/user.go` | Add `IsBot` field to `User` struct |
| `pkg/user/user.go` | Update `ShouldNotify()` to return false for bots |
| `pkg/user/user_create.go` | Add `CreateBotUser()` function |
| `pkg/config/config.go` | Add `ServiceEnableBotUsers` config key |
| `config-raw.json` | Add config schema entry |
| `pkg/routes/api/v1/info.go` | Expose `BotUsersEnabled` in `/info` |
| `pkg/routes/api/v1/login.go` | Block bot login |
| `pkg/routes/routes.go` | Register bot API routes |
| `frontend/src/modelTypes/IUser.ts` | Add `isBot` field |
| `frontend/src/models/user.ts` | Add `isBot` default |
| `frontend/src/components/misc/User.vue` | Add bot badge |
| `frontend/src/i18n/lang/en.json` | Add translation strings |

---

## Open Questions

1. **Should bots be sharable?** Can user A transfer ownership of a bot to user B? Or share management access? For now: no. One owner per bot, non-transferable.

2. **Hard delete vs soft delete?** When a bot is deleted, should we hard-delete the user row or set `Status = Disabled`? Soft delete (disabled) is safer since it preserves audit trail and comment attribution. Recommend soft delete.

3. **Bot-to-bot ownership?** A bot should not be able to own another bot. The owner must be a human user (`is_bot = false`). Enforce in the creation handler.

4. **Bot API token scopes — should we restrict?** Should bots be able to have tokens with any permission, or only a subset? For now: same as regular users — any valid permission scope. The owning user decides what the bot can do.

5. **Rate limiting for bot API calls?** Not in this phase. Can be added as a config option later (`service.botratelimit`).
