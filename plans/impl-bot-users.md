# Implementation Plan: Bot Users

This plan implements bot users as first-class users with an `is_bot` flag, gated behind a `service.enablebotusers` config flag. It incorporates all gap resolutions from the design review.

Reference: [Design Plan](https://github.com/go-vikunja/vikunja/blob/claude/explore-openclaw-integration-KQEzg/plans/design-bot-users.md) | [Gap Analysis](plans/review-bot-design-plan.md)

---

## Phase 1: Backend Foundation

### Step 1.1: Config Key

**File:** `pkg/config/config.go`

- Add `ServiceEnableBotUsers Key = "service.enablebotusers"` alongside existing `ServiceEnable*` keys (around line 52-65).
- In `InitDefaultConfig()`, add `ServiceEnableBotUsers.setDefault(false)`.

**File:** `config-raw.json`

- Add entry in the `service` section (after existing `enableuserdeletion` block around line 107):
```json
{
    "key": "enablebotusers",
    "default_value": "false",
    "comment": "If enabled, users can create bot users that interact with Vikunja via the API only."
}
```

---

### Step 1.2: User Struct Changes

**File:** `pkg/user/user.go`

- Add `IsBot` field to the `User` struct:
```go
IsBot bool `xorm:"bool default false index" json:"is_bot"`
```

- Update `ShouldNotify()` (line 146-160) to return `false` for bots:
```go
if user.IsBot {
    return false, nil
}
```

---

### Step 1.3: Error Types

**File:** `pkg/user/error.go`

Add four new error types after `ErrorCodeTokenUserMismatch = 1029` (line 707+):

| Error Type | Code | HTTP Status | When |
|---|---|---|---|
| `ErrAccountIsBot` | 1030 | 412 | Bot attempts password login |
| `ErrBotUsersDisabled` | 1031 | 403 | Bot endpoint called when feature is off |
| `ErrBotNotOwned` | 1032 | 403 | User tries to manage a bot they don't own |
| `ErrCannotMakeBotOwner` | 1033 | 400 | Trying to set a bot as owner of another bot |
| `ErrBotUsernameMustHavePrefix` | 1034 | 400 | Bot username doesn't start with `bot-` |

Follow the existing error pattern: struct, `IsErr*()` check, `Error()` string, `ErrCode*` constant, `HTTPError()` method.

---

### Step 1.4: Bot Owner Model & Database Migration

**New file:** `pkg/models/bot_users.go`

Define `BotOwner` struct:
```go
type BotOwner struct {
    ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
    BotID   int64     `xorm:"bigint not null unique index" json:"bot_id"`
    OwnerID int64     `xorm:"bigint not null index" json:"owner_id"`
    Created time.Time `xorm:"created not null" json:"created"`
}

func (*BotOwner) TableName() string {
    return "bot_owners"
}
```

**File:** `pkg/models/models.go`

- Add `&BotOwner{}` to `GetTables()` return slice (line 45-73).

**New file:** `pkg/migration/<timestamp>.go`

- Run `mage dev:make-migration BotOwner` to generate migration file.
- Migration adds:
  1. `is_bot` column to `users` table (bool, default false, indexed)
  2. `bot_owners` table via `tx.Sync2(&BotOwner{})`

---

### Step 1.5: CreateBotUser Function

**File:** `pkg/user/user_create.go`

Add `CreateBotUser()` function. This does NOT call `checkIfUserIsValid()` or `checkIfUserExists()` — it has its own validation:

1. Validate username: not empty, no spaces, starts with `bot-`, not a duplicate (via `GetUserByUsername`)
2. Validate name (display name): optional, free-form
3. Create user with:
   - `IsBot = true`
   - `Status = StatusActive`
   - `Issuer = IssuerLocal`
   - `Password = ""` (no password)
   - `Email = ""` (no email)
   - `EmailRemindersEnabled = false`
   - `OverdueTasksRemindersEnabled = false`
   - All other defaults from config
4. Insert user
5. Dispatch `CreatedEvent`
6. Return the new user

Also add to `checkIfUserIsValid()` (line 130-153): reject `bot-` prefix for non-bot users:
```go
if strings.HasPrefix(user.Username, "bot-") {
    return ErrUsernameReserved{Username: user.Username}
}
```

---

### Step 1.6: Block Login for Bots

**File:** `pkg/routes/api/v1/login.go`

After the status check at line 74-77, add:
```go
if user.IsBot {
    _ = s.Rollback()
    return &user2.ErrAccountIsBot{UserID: user.ID}
}
```

This blocks bots after both LDAP and local auth paths converge. Even if LDAP somehow resolves a bot user, it gets blocked here.

**File:** `pkg/routes/caldav/auth.go`

Before `c.Set("userBasicAuth", u)` at line 63-64:
```go
if u != nil && u.IsBot {
    log.Warningf("CalDAV basic auth rejected for bot user %d", u.ID)
    return false, nil
}
```

---

### Step 1.7: API Token Auth - User Status Check (Security Fix)

**File:** `pkg/routes/api_tokens.go`

In `checkAPITokenAndPutItInContext()`, after `user.GetUserByID()` at line 94-97, add:
```go
if u.Status == user.StatusDisabled {
    log.Debugf("[auth] Tried authenticating with token %d but user %d is disabled", token.ID, u.ID)
    return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
}
```

This ensures disabling a bot (or any user) immediately revokes API token access.

---

### Step 1.8: APIToken.Create() - Minimal Branching

**File:** `pkg/models/api_tokens.go`

In `Create()` (line 86-110), change line 102 from:
```go
t.OwnerID = a.GetID()
```
to:
```go
if t.OwnerID == 0 {
    t.OwnerID = a.GetID()
}
```

This allows bot token handlers to pre-set `OwnerID` to the bot's ID.

---

## Phase 2: Bot API Endpoints

### Step 2.1: Bot CRUD Model Methods

**File:** `pkg/models/bot_users.go` (extend from Step 1.4)

Add a `BotUser` struct for the API layer (wraps user + ownership info). This struct implements `web.CRUDable` and `web.Permissions` for use with generic WebHandlers where possible.

However, because bot CRUD involves creating users (not just inserting a model row), custom handlers are needed for Create and Delete. ReadAll and ReadOne can use generic patterns.

Define the following methods on `BotUser`:
- `Create(s, a)`: Call `user.CreateBotUser()`, then insert `BotOwner{BotID: newUser.ID, OwnerID: a.GetID()}`
- `ReadAll(s, a, search, page, perPage)`: Query `bot_owners WHERE owner_id = a.GetID()` joined with `users`
- `ReadOne(s, a)`: Get bot by ID, verify ownership
- `Update(s, a)`: Update bot's `name` or `username` (verify ownership, validate `bot-` prefix preserved)
- `Delete(s, a)`: Call `DeleteUser()` from `pkg/models/user_delete.go` to hard-delete the bot and all related data. Also delete the `BotOwner` row. Skip `AccountDeletedNotification` since `ShouldNotify()` returns false for bots.

Add a `Disable`/`Enable` method (or use `Update` with a status field):
- Set `user.Status = StatusDisabled` or `StatusActive`

**New file:** `pkg/models/bot_users_permissions.go`

Permission methods:
- `CanCreate(s, a)`: Return true if `config.ServiceEnableBotUsers.GetBool()` and user is not a bot.
- `CanRead(s, a)`: Return true if user owns this bot (check `bot_owners`).
- `CanUpdate(s, a)`: Same as CanRead.
- `CanDelete(s, a)`: Same as CanRead.

---

### Step 2.2: Bot Token Endpoints

**File:** `pkg/routes/api/v1/bots.go` (new file)

Custom handlers for bot token management since we need to override the token owner:

**`PUT /user/bots/:bot/tokens`** - Create token for bot:
1. Parse bot ID from `:bot` param
2. Verify current user owns this bot (query `bot_owners`)
3. Bind `APIToken` from request body
4. Set `token.OwnerID = bot.ID` (pre-set before calling Create)
5. Call `token.Create(s, a)` — uses the branched logic from Step 1.8
6. Return the token (visible only once)

**`GET /user/bots/:bot/tokens`** - List bot's tokens:
1. Verify ownership
2. Query `api_tokens WHERE owner_id = bot.ID`

**`DELETE /user/bots/:bot/tokens/:token`** - Delete bot's token:
1. Verify ownership
2. Delete `api_tokens WHERE id = :token AND owner_id = bot.ID`

---

### Step 2.3: Route Registration

**File:** `pkg/routes/routes.go`

In the user routes group (after line 462), add:
```go
if config.ServiceEnableBotUsers.GetBool() {
    botHandler := &handler.WebHandler{
        EmptyStruct: func() handler.CObject {
            return &models.BotUser{}
        },
    }
    u.PUT("/bots", botHandler.CreateWeb)
    u.GET("/bots", botHandler.ReadAllWeb)
    u.GET("/bots/:bot", botHandler.ReadOneWeb)
    u.POST("/bots/:bot", botHandler.UpdateWeb)
    u.DELETE("/bots/:bot", botHandler.DeleteWeb)

    // Bot token management (custom handlers)
    u.PUT("/bots/:bot/tokens", apiv1.CreateBotToken)
    u.GET("/bots/:bot/tokens", apiv1.ListBotTokens)
    u.DELETE("/bots/:bot/tokens/:token", apiv1.DeleteBotToken)
}
```

---

### Step 2.4: Info Endpoint

**File:** `pkg/routes/api/v1/info.go`

- Add `BotUsersEnabled bool \`json:"bot_users_enabled"\`` to `vikunjaInfos` struct (line 35-55).
- In the `Info` handler, set: `info.BotUsersEnabled = config.ServiceEnableBotUsers.GetBool()`.

---

## Phase 3: Frontend

### Step 3.1: Config Store

**File:** `frontend/src/stores/config.ts`

- Add to `ConfigState` interface (after `publicTeamsEnabled` at line 46):
```typescript
botUsersEnabled: boolean,
```

- Add default in reactive state (after `publicTeamsEnabled: false` at line 85):
```typescript
botUsersEnabled: false,
```

No other changes needed — `objectToCamelCase(config)` in `update()` automatically maps `bot_users_enabled` → `botUsersEnabled`.

---

### Step 3.2: User Type & Model

**File:** `frontend/src/modelTypes/IUser.ts`

Add to interface:
```typescript
isBot: boolean
```

**File:** `frontend/src/models/user.ts`

Add default in constructor:
```typescript
isBot = false
```

---

### Step 3.3: Bot Badge in User Component

**File:** `frontend/src/components/misc/User.vue`

After the username `<span>`, add:
```vue
<span v-if="user?.isBot" class="bot-badge">{{ $t('user.bot.badge') }}</span>
```

Style:
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

---

### Step 3.4: Bot Service & Model

**New file:** `frontend/src/services/botUser.ts`

```typescript
export default class BotUserService extends AbstractService<IBotUser> {
    constructor() {
        super({
            create: '/user/bots',
            getAll: '/user/bots',
            get: '/user/bots/{id}',
            update: '/user/bots/{id}',
            delete: '/user/bots/{id}',
        })
    }
    modelFactory(data) { return new BotUserModel(data) }
}
```

**New file:** `frontend/src/services/botToken.ts`

```typescript
export default class BotTokenService extends AbstractService<IApiToken> {
    constructor(botId: number) {
        super({
            create: `/user/bots/${botId}/tokens`,
            getAll: `/user/bots/${botId}/tokens`,
            delete: `/user/bots/${botId}/tokens/{id}`,
        })
    }
    modelFactory(data) { return new ApiTokenModel(data) }
}
```

**New file:** `frontend/src/modelTypes/IBotUser.ts`

```typescript
export interface IBotUser extends IAbstract {
    id: number
    username: string
    name: string
    isBot: boolean
    created: Date
    updated: Date
}
```

**New file:** `frontend/src/models/botUser.ts`

Model class extending `AbstractModel<IBotUser>`.

---

### Step 3.5: Bot Settings Page

**New file:** `frontend/src/views/user/settings/Bots.vue`

Settings page with:
- List of bots owned by current user (fetched from `GET /user/bots`)
- Create bot form: username (with `bot-` prefix pre-filled), display name
- Per-bot actions:
  - Edit (update display name)
  - Disable/Enable toggle
  - Delete (with confirmation dialog — make it clear this is permanent and irreversible)
- Per-bot token management:
  - Create token (show token value once, copy-to-clipboard)
  - List existing tokens
  - Delete token

Follow the pattern of `frontend/src/views/user/settings/ApiTokens.vue` for the token management UI.

---

### Step 3.6: Route & Navigation

**File:** `frontend/src/router/index.ts`

Add to `user.settings` children (after the webhooks route, line 147-151):
```typescript
{
    path: '/user/settings/bots',
    name: 'user.settings.bots',
    component: () => import('@/views/user/settings/Bots.vue'),
},
```

**File:** `frontend/src/views/user/Settings.vue`

Add computed:
```typescript
const botUsersEnabled = computed(() => configStore.botUsersEnabled)
```

Add to `navigationItems` array (before the deletion entry):
```typescript
{
    title: t('user.settings.bots.title'),
    routeName: 'user.settings.bots',
    condition: botUsersEnabled.value,
},
```

---

### Step 3.7: Translation Strings

**File:** `frontend/src/i18n/lang/en.json`

Add under `user.settings`:
```json
"bots": {
    "title": "Bots",
    "create": "Create Bot",
    "createDescription": "Bot users can interact with Vikunja via the API only. Bot usernames must start with 'bot-'.",
    "username": "Bot Username",
    "usernamePlaceholder": "bot-my-assistant",
    "displayName": "Display Name",
    "displayNamePlaceholder": "My Assistant Bot",
    "noBotsYet": "You haven't created any bots yet.",
    "delete": "Delete Bot",
    "deleteConfirmation": "This will permanently delete this bot and all its data including task assignments, project memberships, and API tokens. This action cannot be undone.",
    "disable": "Disable Bot",
    "enable": "Enable Bot",
    "disabled": "Disabled",
    "tokens": "API Tokens",
    "createToken": "Create Token",
    "tokenCreated": "Token created. Copy it now — you won't be able to see it again.",
    "noTokens": "No API tokens yet."
}
```

Add under `user`:
```json
"bot": {
    "badge": "Bot"
}
```

---

## Phase 4: Testing

### Step 4.1: Backend Test Fixtures

**File:** `pkg/db/fixtures/users.yml` (or appropriate fixture file)

Add bot user fixture entries for testing.

### Step 4.2: Backend Tests

**New file:** `pkg/models/bot_users_test.go`

Test cases:
- Create bot successfully
- Create bot when feature disabled → error
- Create bot with invalid username (no `bot-` prefix) → error
- Create bot as a bot → error
- List bots returns only owned bots
- Update bot name
- Delete bot (verify cascading: task assignees, tokens, project memberships removed)
- Disable bot → verify API token auth fails
- Enable bot → verify API token auth works again
- Create token for bot
- Create token for bot not owned → error
- Login as bot → error
- CalDAV auth as bot → rejected

### Step 4.3: Frontend Tests

Unit tests for bot service and model. The settings page can be covered by E2E tests if needed.

---

## Implementation Order

1. **Phase 1** (Steps 1.1–1.8): Backend foundation — can be tested independently
2. **Phase 2** (Steps 2.1–2.4): API endpoints — testable with curl/API tests
3. **Phase 3** (Steps 3.1–3.7): Frontend — requires backend running
4. **Phase 4** (Steps 4.1–4.3): Tests — run alongside implementation

Within Phase 1, the steps are sequential (each depends on prior). Phase 2 depends on Phase 1. Phase 3 depends on Phase 2. Phase 4 can be written alongside Phases 1-3.

---

## Files Summary

### New Files (10)
| File | Purpose |
|------|---------|
| `pkg/models/bot_users.go` | BotOwner + BotUser model, CRUD logic |
| `pkg/models/bot_users_permissions.go` | Permission checks |
| `pkg/models/bot_users_test.go` | Backend tests |
| `pkg/routes/api/v1/bots.go` | Bot token endpoint handlers |
| `pkg/migration/<timestamp>.go` | Database migration |
| `frontend/src/views/user/settings/Bots.vue` | Bot management UI |
| `frontend/src/services/botUser.ts` | Bot CRUD service |
| `frontend/src/services/botToken.ts` | Bot token service |
| `frontend/src/modelTypes/IBotUser.ts` | Bot TypeScript interface |
| `frontend/src/models/botUser.ts` | Bot model class |

### Modified Files (15)
| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `ServiceEnableBotUsers` key + default |
| `config-raw.json` | Add `enablebotusers` schema entry |
| `pkg/user/user.go` | Add `IsBot` field; update `ShouldNotify()` |
| `pkg/user/user_create.go` | Add `CreateBotUser()`; reserve `bot-` prefix in `checkIfUserIsValid()` |
| `pkg/user/error.go` | Add 5 new error types (codes 1030-1034) |
| `pkg/models/models.go` | Add `&BotOwner{}` to `GetTables()` |
| `pkg/models/api_tokens.go` | Branch `OwnerID` assignment in `Create()` |
| `pkg/routes/routes.go` | Register bot API routes |
| `pkg/routes/api/v1/info.go` | Expose `BotUsersEnabled` in `/info` |
| `pkg/routes/api/v1/login.go` | Block bot login |
| `pkg/routes/api_tokens.go` | Add user status check (security fix) |
| `pkg/routes/caldav/auth.go` | Block bot CalDAV auth |
| `frontend/src/stores/config.ts` | Add `botUsersEnabled` to state |
| `frontend/src/modelTypes/IUser.ts` | Add `isBot` field |
| `frontend/src/models/user.ts` | Add `isBot` default |
| `frontend/src/components/misc/User.vue` | Add bot badge |
| `frontend/src/views/user/Settings.vue` | Add bots nav item |
| `frontend/src/router/index.ts` | Add bots route |
| `frontend/src/i18n/lang/en.json` | Add translation strings |
