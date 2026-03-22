# Bot Users - Gap Analysis & Implementation Resolutions

## Context

The [design plan](https://github.com/go-vikunja/vikunja/blob/claude/explore-openclaw-integration-KQEzg/plans/design-bot-users.md) for bot users is solid overall but has gaps that need resolution before implementation. This document identifies each gap and provides a concrete resolution based on codebase analysis and design decisions.

---

## Gap 1: Empty Email Uniqueness Conflict (Critical)

**Problem:** `checkIfUserExists()` (`pkg/user/user_create.go:155-194`) checks email uniqueness for local users. Multiple bots with `Email=""` would all conflict via `getUser(s, &User{Email: "", Issuer: "local"}, false)`.

**Resolution:** `CreateBotUser()` must:
- Only check username uniqueness (reuse `GetUserByUsername`)
- Skip `checkIfUserIsValid()` entirely (requires email + password)
- Skip the email existence check

---

## Gap 2: Bot Token Creation - Minimal Branching (Critical)

**Problem:** `APIToken.Create()` (`pkg/models/api_tokens.go:86-110`) hardcodes `t.OwnerID = a.GetID()`.

**Resolution:** Add a `CreateForUser()` method or modify `Create()` to accept an optional owner override. Something like: if `t.OwnerID` is already set (non-zero) when `Create()` is called, preserve it instead of overwriting. The bot token handler would pre-set `t.OwnerID = bot.ID` before calling `Create()`. This is minimal branching - just one `if` check:

```go
if t.OwnerID == 0 {
    t.OwnerID = a.GetID()
}
```

The bot endpoint handler verifies ownership of the bot, then calls `Create()` with `OwnerID` pre-set.

For `ReadAll()` and `Delete()` on bot tokens, the handler would similarly need to filter by bot's OwnerID rather than `a.GetID()`. Custom handlers for the bot token endpoints can query with the bot's ID directly.

---

## Gap 3: Error Codes (Required)

**Problem:** All error codes are TBD.

**Resolution:** Existing user error codes go up to ~1021 (`ErrAccountIsNotLocal`). Assign:
- `ErrAccountIsBot` → 1022
- `ErrBotUsersDisabled` → 1023
- `ErrBotNotOwned` → 1024
- `ErrCannotMakeBotOwner` → 1025

Verify the next available code by checking `pkg/user/error.go` and `pkg/models/error.go` at implementation time.

---

## Gap 4: CalDAV Bot Blocking Location (Required)

**Problem:** Plan doesn't specify exact location in `BasicAuth()` (`pkg/routes/caldav/auth.go:31-68`).

**Resolution:** Add check before `c.Set("userBasicAuth", u)` at line 63-64. Both the token path and password path converge there:

```go
if u != nil && u.IsBot {
    log.Warningf("CalDAV basic auth rejected for bot user %d", u.ID)
    return false, nil
}
```

---

## Gap 5: BotOwner Table Registration (Required)

**Problem:** `BotOwner` not mentioned for registration in `pkg/models/models.go`.

**Resolution:** Add `&BotOwner{}` to the `GetTables()` return slice.

---

## Gap 6: Registration Config Independence (Clarification)

**Problem:** Unclear if `service.enableregistration = false` blocks bot creation.

**Resolution:** Non-issue. Bot creation uses `PUT /user/bots` which is gated only by `service.enablebotusers`. The registration endpoint (`POST /register`) is completely separate. No changes needed.

---

## Gap 7: LDAP Auth Blocking (Required)

**Problem:** Login handler tries LDAP first. Need bot check before LDAP attempt.

**Resolution:** In `pkg/routes/api/v1/login.go`, after resolving the user by username but before LDAP/local auth:
```go
if user.IsBot {
    return ErrAccountIsBot{UserID: user.ID}
}
```

Note: Need to check exact flow - the current login may not resolve the user before attempting auth. If LDAP auth creates/matches users by username, add a post-LDAP check too.

---

## Gap 8: Admin Capabilities → No Admin Panel Exists

**Problem:** Plan discusses admin management of bots.

**Resolution:** There is no admin panel in Vikunja. Remove all admin-related discussion from the plan. Bot management is owner-only.

---

## Gap 9: Frontend Route & Navigation (Required)

**Problem:** Missing router config and Settings.vue navigation details.

**Resolution:**

**Router** (`frontend/src/router/index.ts`): Add as child of `user.settings`:
```ts
{
    path: '/user/settings/bots',
    name: 'user.settings.bots',
    component: () => import('@/views/user/settings/Bots.vue'),
}
```

**Settings.vue** (`frontend/src/views/user/Settings.vue`): Add to `navigationItems`:
```ts
{
    title: t('user.settings.bots.title'),
    routeName: 'user.settings.bots',
    condition: botUsersEnabled.value,
}
```

Add `const botUsersEnabled = computed(() => configStore.botUsersEnabled)` alongside other feature flags.

---

## Gap 10: Frontend Config Store (Required)

**Problem:** Missing config store changes.

**Resolution:** In `frontend/src/stores/config.ts`:
- Add `botUsersEnabled: false` to the state
- Map from API response's `bot_users_enabled` field

---

## Gap 11: Bot Deletion - Follow User Deletion Pattern (Critical)

**Problem:** Plan says "unassign bot from all tasks" but deletion semantics were unclear.

**Resolution (per user decision):** Two operations:

### Delete (hard delete, like users)
Reuse the existing `DeleteUser()` function in `pkg/models/user_delete.go:132-183`. It already handles:
- Deleting task assignees, subscriptions, team members, saved filters, reactions, favorites, API tokens
- Reassigning/deleting owned projects
- Deleting the user row

For bots, skip the `AccountDeletedNotification` (bots don't get notifications). Consider adding a bot-specific check in `DeleteUser()` or a wrapper function.

### Disable (soft disable)
Set `Status = StatusDisabled`. Bot tokens still exist but the API token middleware resolves the user and checks... actually need to verify: does the API token auth check user status? If not, disabling a bot might not revoke API access.

**Action item:** Check if `pkg/routes/api_tokens.go` checks user status after resolving the token owner. If not, add a check.

### Frontend
- Delete button with confirmation dialog (make it clear this is permanent)
- Disable/Enable toggle button (reversible)

---

## Gap 12: Initial Project → Non-Issue

Initial project creation happens in `pkg/routes/api/v1/user_register.go:84-89`, not in `CreateUser()`. Bot creation uses a separate endpoint, so this is automatically skipped.

---

## Gap 13: is_bot in User Search Responses → Automatic

Adding `IsBot` to the `User` struct with `json:"is_bot"` includes it in all API responses. No additional changes needed.

---

## Gap 14: Bot Username Convention (Design Decision)

**Resolution (per user):** Require `bot-` prefix for usernames. Display name (the `name` field) is free-form.

Add validation in bot creation:
```go
if !strings.HasPrefix(bot.Username, "bot-") {
    return ErrBotUsernameMustHavePrefix{}
}
```

Also reserve the `bot-` prefix for non-bot users (add check in regular `checkIfUserIsValid()`).

---

## Gap 15: API Token Auth - User Status Check (Confirmed Critical)

**Problem:** When a bot is disabled (`StatusDisabled`), its API tokens still work. Verified in `pkg/routes/api_tokens.go:76-103` - `checkAPITokenAndPutItInContext()` fetches the user at line 94 but never checks `Status`. A disabled bot's tokens remain functional.

**Resolution:** Add status check after line 97 in `checkAPITokenAndPutItInContext()`:
```go
if u.Status == user.StatusDisabled {
    log.Debugf("[auth] Tried authenticating with token %d but user %d is disabled", token.ID, u.ID)
    return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
}
```

This benefits both bots and regular disabled users - it's a general security improvement.

---

## Gap 16: OpenID Connect → Non-Issue by Design

OIDC matches users by Subject+Issuer, which bots don't have. No blocking needed.

---

## Summary of Resolutions

| Gap | Status | Key Change |
|-----|--------|-----------|
| 1. Email uniqueness | Resolved | Skip email check in `CreateBotUser()` |
| 2. Token creation | Resolved | Branch in `Create()`: preserve pre-set OwnerID |
| 3. Error codes | Resolved | 1022-1025 range |
| 4. CalDAV blocking | Resolved | Check before `c.Set("userBasicAuth")` |
| 5. Table registration | Resolved | Add to `GetTables()` |
| 6. Registration independence | Non-issue | Separate endpoints |
| 7. LDAP blocking | Resolved | Check before auth attempt |
| 8. Admin capabilities | Removed | No admin panel exists |
| 9. Frontend routes | Resolved | Router + Settings.vue details specified |
| 10. Config store | Resolved | Add `botUsersEnabled` to config store |
| 11. Bot deletion | Resolved | Hard delete via `DeleteUser()`, plus disable toggle |
| 12. Initial project | Non-issue | Separate code path |
| 13. User search | Automatic | `json:"is_bot"` covers it |
| 14. Username prefix | Resolved | Require `bot-` prefix, free display name |
| 15. Token auth status | Confirmed gap | Add status check in `checkAPITokenAndPutItInContext()` |
| 16. OIDC | Non-issue | By design |

---

## Files to Verify During Implementation

- `pkg/routes/api_tokens.go` - **Confirmed: no user status check** - must add one (line 94-97)
- `pkg/user/error.go` - Confirm next available error code before assigning 1022+
- `pkg/routes/api/v1/login.go` - Exact location for bot check vs LDAP flow
- `frontend/src/stores/config.ts` - Exact state structure for adding `botUsersEnabled`
- `frontend/src/router/index.ts` - Exact child route pattern under `user.settings`

---

## Verification Plan

1. **Backend**: `mage test:filter` for bot-related tests
2. **Login blocking**: Attempt password login as bot → expect `ErrAccountIsBot`
3. **CalDAV blocking**: Attempt CalDAV auth as bot → expect rejection
4. **Token flow**: Create bot → create token for bot → use token to call API as bot
5. **Disable flow**: Disable bot → verify its tokens stop working
6. **Delete flow**: Delete bot → verify all related data is cleaned up
7. **Frontend**: Bot badge shows in User.vue, settings page CRUD works, config gate works
