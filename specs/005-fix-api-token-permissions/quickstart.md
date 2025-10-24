# Quick Start: Fix API Token Permissions System

**Feature**: Fix API Token Permissions System  
**Branch**: `005-fix-api-token-permissions`  
**Date**: October 23, 2025

## Problem Statement

API tokens cannot perform create, update, or delete operations on tasks (via v1 API) even when "all permissions" are selected in the UI. The GET /routes endpoint returns incomplete permission scopes for v1_tasks, missing create/update/delete permissions.

**Root Cause**: Incomplete route registration system where explicit CollectRoute calls may be overridden by legacy CollectRoutesForAPITokenUsage detection.

## Quick Fix Overview

This is a **minimal infrastructure fix** that ensures the declarative APIRoute registration pattern works correctly without reverting to deprecated "magic" detection.

**Scope**: ~200 LOC changes across 5-10 files  
**Risk Level**: Low (defensive changes, backward compatible)  
**Breaking Changes**: None

## Prerequisites

**Required Tools**:
- Go 1.21+
- mage (build tool)
- pnpm (frontend package manager)
- Running Vikunja instance (for testing)

**Environment Setup**:
```bash
export VIKUNJA_SERVICE_ROOTPATH=$(pwd)
export PATH=$PATH:$(go env GOPATH)/bin
```

## Development Workflow

### Step 1: Verify Current State (5 min)

**Test the bug**:
```bash
# Start Vikunja server
mage build
./vikunja

# In another terminal, get available routes
curl -X GET http://localhost:3456/api/v1/routes \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" | jq '.v1.tasks'

# Expected (BROKEN): Missing create/update/delete
# Expected (FIXED): Shows create, update, delete, read_one
```

**Create an API token via UI**:
1. Login to Vikunja
2. Navigate to Settings → API Tokens
3. Create new token with "all permissions"
4. Note which permissions show up for "tasks"

**Expected Bug Symptoms**:
- v1_tasks only shows "read_one" permission (maybe)
- Create/update/delete checkboxes missing from UI
- Attempting task creation with token returns 401 Unauthorized

---

### Step 2: Write Failing Tests (15 min)

**Test File**: `pkg/services/api_tokens_test.go`

Add tests that currently FAIL:

```go
func TestAPITokenPermissionRegistration(t *testing.T) {
    // Register test routes
    registerTestAPIRoutes()
    
    routes := models.GetAPITokenRoutes()
    
    // Test v1 tasks has all CRUD operations
    v1Routes, hasV1 := routes["v1"]
    assert.True(t, hasV1, "Should have v1 routes")
    
    taskRoutes, hasTasks := v1Routes["tasks"]
    assert.True(t, hasTasks, "Should have tasks routes")
    
    // THESE SHOULD FAIL before fix:
    assert.NotNil(t, taskRoutes["create"], "Should have create permission")
    assert.NotNil(t, taskRoutes["update"], "Should have update permission")
    assert.NotNil(t, taskRoutes["delete"], "Should have delete permission")
    assert.NotNil(t, taskRoutes["read_one"], "Should have read_one permission")
}

func TestAPITokenCanCreateTask(t *testing.T) {
    s := db.NewSession()
    defer s.Close()
    
    // Create token with v1_tasks create permission
    token := createTokenWithPermissions(t, s, map[string][]string{
        "v1_tasks": {"create"},
    })
    
    // Mock Echo context for PUT /api/v1/projects/:project/tasks
    c := createMockContext("PUT", "/api/v1/projects/1/tasks")
    
    // SHOULD FAIL before fix:
    can := models.CanDoAPIRoute(c, token)
    assert.True(t, can, "Token with create permission should be able to create tasks")
}
```

**Run tests** (should FAIL):
```bash
cd /home/aron/projects/vikunja
mage test:feature -run TestAPITokenPermission
```

---

### Step 3: Implement Fix (30 min)

**Primary Changes**:

#### 3.1: Make CollectRoute Idempotent

**File**: `pkg/models/api_routes.go`

**Change**: Modify `CollectRoutesForAPITokenUsage` to skip routes already explicitly registered:

```go
func CollectRoutesForAPITokenUsage(route echo.Route, middlewares []echo.MiddlewareFunc) {
    // ... existing version/group extraction ...
    
    // NEW: Check if route already explicitly registered
    if existingRoutes, hasGroup := apiTokenRoutes[apiVersion][routeGroupName]; hasGroup {
        // Check if we already have an explicit registration for this path+method
        for _, detail := range existingRoutes {
            if detail != nil && detail.Path == route.Path && detail.Method == route.Method {
                // Already explicitly registered, skip legacy detection
                log.Debugf("[routes] Skipping legacy detection for %s %s (already explicit)", route.Method, route.Path)
                return
            }
        }
    }
    
    // ... rest of existing logic for legacy detection ...
}
```

#### 3.2: Add Logging to CollectRoute

**File**: `pkg/models/api_routes.go`

**Change**: Add debug logging to verify explicit registration:

```go
func CollectRoute(method, path, permissionScope string) {
    // ... existing extraction logic ...
    
    ensureAPITokenRoutesGroup(apiVersion, routeGroupName)
    routeDetail := &RouteDetail{
        Path:   path,
        Method: method,
    }
    apiTokenRoutes[apiVersion][routeGroupName][permissionScope] = routeDetail
    
    // NEW: Log successful registration
    log.Debugf("[routes] Explicitly registered: %s %s → %s_%s.%s", 
        method, path, apiVersion, routeGroupName, permissionScope)
}
```

#### 3.3: Enhance CanDoAPIRoute Logging

**File**: `pkg/models/api_routes.go`

**Change**: Improve error logging when token lacks permission:

```go
func CanDoAPIRoute(c echo.Context, token *APIToken) (can bool) {
    // ... existing logic ...
    
    // Enhanced logging at the end:
    availableScopes := []string{}
    for scope := range routes {
        availableScopes = append(availableScopes, scope)
    }
    
    log.Debugf("[auth] Token %d tried to use route %s which requires permission %s but has only %v. Available scopes for %s: %v", 
        token.ID, path, route, token.APIPermissions, routeGroupName, availableScopes)
    
    return false
}
```

#### 3.4: Verify v1 Task Routes Registration

**File**: `pkg/routes/api/v1/task.go`

**Verify**: TaskRoutes array is complete (should already be correct):

```go
var TaskRoutes = []APIRoute{
    {Method: "PUT", Path: "/projects/:project/tasks", Handler: handler.WithDBAndUser(createTaskLogic, true), PermissionScope: "create"},
    {Method: "GET", Path: "/tasks/:taskid", Handler: handler.WithDBAndUser(getTaskLogic, false), PermissionScope: "read_one"},
    {Method: "POST", Path: "/tasks/:taskid", Handler: handler.WithDBAndUser(updateTaskLogic, true), PermissionScope: "update"},
    {Method: "DELETE", Path: "/tasks/:taskid", Handler: handler.WithDBAndUser(deleteTaskLogic, true), PermissionScope: "delete"},
}
```

**Verify**: RegisterTasks calls registerRoutes:

```go
func RegisterTasks(a *echo.Group) {
    registerRoutes(a, TaskRoutes)
}
```

#### 3.5: Convert v2 Routes to Declarative Pattern (Optional but Recommended)

**File**: `pkg/routes/api/v2/tasks.go`

**Change**: Convert manual registration to declarative:

```go
var TaskRoutes = []apiv1.APIRoute{
    {Method: "GET", Path: "/tasks", Handler: GetTasks, PermissionScope: "read_all"},
}

func RegisterTasks(a *echo.Group) {
    apiv1.registerRoutes(a, TaskRoutes)  // Use v1's helper
}
```

**Alternative**: Add explicit CollectRoute calls:

```go
func RegisterTasks(a *echo.Group) {
    a.GET("/tasks", GetTasks)
    models.CollectRoute("GET", "/tasks", "read_all")
}
```

---

### Step 4: Run Tests (5 min)

**Backend Tests**:
```bash
# Run specific test
mage test:feature -run TestAPITokenPermission

# Run all service tests
mage test:feature

# Run web tests
mage test:web
```

**Expected Result**: All tests should now PASS.

---

### Step 5: Manual Verification (10 min)

**Test GET /routes endpoint**:
```bash
curl -X GET http://localhost:3456/api/v1/routes \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" | jq '.v1.tasks'
```

**Expected Output** (after fix):
```json
{
  "create": {
    "path": "/api/v1/projects/:project/tasks",
    "method": "PUT"
  },
  "read_one": {
    "path": "/api/v1/tasks/:taskid",
    "method": "GET"
  },
  "update": {
    "path": "/api/v1/tasks/:taskid",
    "method": "POST"
  },
  "delete": {
    "path": "/api/v1/tasks/:taskid",
    "method": "DELETE"
  }
}
```

**Test API Token Creation in UI**:
1. Navigate to Settings → API Tokens
2. Click "Create new token"
3. Verify "tasks" section shows: create, read, update, delete checkboxes
4. Create token with "all permissions"
5. Verify token can create/update/delete tasks

**Test API Token Authentication**:
```bash
# Create a task with API token
curl -X PUT http://localhost:3456/api/v1/projects/1/tasks \
  -H "Authorization: Bearer tk_xxxxx..." \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Task", "description": "Created via API token"}'

# Expected: 201 Created (not 401 Unauthorized)

# Update the task
curl -X POST http://localhost:3456/api/v1/tasks/123 \
  -H "Authorization: Bearer tk_xxxxx..." \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Task"}'

# Expected: 200 OK

# Delete the task
curl -X DELETE http://localhost:3456/api/v1/tasks/123 \
  -H "Authorization: Bearer tk_xxxxx..."

# Expected: 200 OK
```

---

### Step 6: Lint and Format (5 min)

**Backend**:
```bash
mage fmt
mage lint:fix
```

**Frontend** (if UI changes):
```bash
cd frontend
pnpm lint:fix
pnpm lint:styles:fix
```

---

### Step 7: Commit (2 min)

```bash
git add .
git commit -m "fix: register complete API token permission scopes for v1 tasks

- Make CollectRoutesForAPITokenUsage skip routes already explicitly registered
- Add debug logging to CollectRoute and CanDoAPIRoute
- Verify TaskRoutes array includes all CRUD operations
- Ensure GET /routes returns complete permission scopes
- Add tests for API token permission validation

Fixes API tokens unable to create/update/delete tasks despite having
'all permissions' selected. The issue was incomplete route registration
where explicit CollectRoute calls were being overridden by legacy detection.

Refs: specs/005-fix-api-token-permissions/spec.md"
```

---

## Common Issues & Solutions

### Issue 1: Tests still failing after fix

**Symptom**: TestAPITokenPermissionRegistration fails, routes map is empty

**Cause**: Test not calling registerTestAPIRoutes()

**Solution**: Ensure test calls helper before assertions:
```go
func TestAPITokenPermissionRegistration(t *testing.T) {
    registerTestAPIRoutes()  // ← Add this
    routes := models.GetAPITokenRoutes()
    // ... assertions ...
}
```

---

### Issue 2: GET /routes returns empty for v1_tasks

**Symptom**: Endpoint returns `{"v1": {}}` or v1_tasks is missing

**Cause**: RegisterTasks not being called in routes.go

**Solution**: Verify in `pkg/routes/routes.go`:
```go
func registerAPIRoutes(a *echo.Group) {
    // ... other setup ...
    
    apiv1.RegisterTasks(a)  // ← Verify this is present
    
    // ... rest of routes ...
}
```

---

### Issue 3: Token still can't create tasks (401 Unauthorized)

**Symptom**: API returns 401 even with token having v1_tasks create permission

**Cause**: Token was created before fix, doesn't have new permission structure

**Solution**: Delete old token and create new one via UI after fix is applied.

---

### Issue 4: V2 routes not showing up

**Symptom**: GET /routes returns empty v2 object

**Cause**: V2 routes not using CollectRoute (still relying on deprecated detection)

**Solution**: Convert v2 routes to declarative pattern or add explicit CollectRoute calls (see Step 3.5).

---

## Validation Checklist

Before considering this fix complete, verify:

- [ ] GET /routes returns v1_tasks with create, update, delete, read_one
- [ ] Frontend token creation UI shows all four task permissions
- [ ] API token with create permission can create tasks (PUT /projects/:project/tasks)
- [ ] API token with update permission can update tasks (POST /tasks/:taskid)
- [ ] API token with delete permission can delete tasks (DELETE /tasks/:taskid)
- [ ] API token without specific permission is rejected (401 Unauthorized)
- [ ] Existing tokens created before fix still work (backward compatibility)
- [ ] V2 routes show up in GET /routes (if Step 3.5 implemented)
- [ ] All tests pass: `mage test:feature && mage test:web`
- [ ] Linting passes: `mage lint` (no errors)
- [ ] No new warnings in server logs

---

## Admin-Level Tokens (Phase 3 Enhancement)

### Overview

Some operations require elevated privileges beyond standard API token capabilities. Admin-level tokens provide access to sensitive operations like webhook management and team administration.

### Token Levels

**Standard Token** (default):
- Access to regular operations: tasks, projects, labels, comments, attachments, etc.
- Cannot access webhooks or team management
- Suitable for most integrations and automation

**Admin Token**:
- All standard token capabilities
- **Plus**: Webhook management (create, read, update, delete webhooks)
- **Plus**: Team management (create teams, manage members)
- Requires explicit admin level selection during token creation
- Recommended only for trusted integrations

### Creating Admin Tokens

**Via UI** (frontend must be updated with T076):
1. Navigate to Settings → API Tokens
2. Click "Create a token"
3. Enter token title
4. **Select token level**: Choose "Admin" (shows security warning)
5. Select permissions (must include webhook/team permissions for admin operations)
6. Click "Create token"

**Via API**:
```bash
curl -X PUT http://localhost:3456/api/v1/tokens \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Admin Integration Token",
    "token_level": "admin",
    "permissions": {
      "v1_projects_webhooks": ["read_all", "create", "update", "delete"],
      "v1_teams": ["read_all", "create", "update", "delete", "add_member"]
    },
    "expires_at": "2026-01-01T00:00:00Z"
  }'
```

### Admin-Only Routes

The following routes **require admin-level tokens**:

**Webhooks** (`v1_projects_webhooks`):
- `GET /api/v1/projects/:project/webhooks` - List webhooks
- `PUT /api/v1/projects/:project/webhooks` - Create webhook
- `POST /api/v1/projects/:project/webhooks/:webhook` - Update webhook
- `DELETE /api/v1/projects/:project/webhooks/:webhook` - Delete webhook

**Teams** (`v1_teams`):
- `GET /api/v1/teams` - List all teams
- `GET /api/v1/teams/:team` - Get team details
- `PUT /api/v1/teams` - Create team
- `POST /api/v1/teams/:team` - Update team
- `DELETE /api/v1/teams/:team` - Delete team
- `PUT /api/v1/teams/:team/members` - Add team member
- `DELETE /api/v1/teams/:team/members/:user` - Remove team member
- `POST /api/v1/teams/:team/members/:user/admin` - Update member role

### Security Considerations

**Why admin tokens?**
- Webhooks can send data to external endpoints (potential data leaks)
- Team membership affects permissions across multiple projects
- Separation of concerns: not all integrations need these capabilities

**Best practices**:
1. **Use standard tokens by default** - Only create admin tokens when absolutely necessary
2. **Minimize admin token permissions** - Only grant webhook/team permissions if needed
3. **Short expiration times** - Admin tokens should have shorter lifespans
4. **Audit regularly** - Review admin token usage in logs
5. **Secure storage** - Store admin tokens with extra protection (secrets managers)
6. **Rotate frequently** - Replace admin tokens periodically

**Checking route requirements**:
```bash
# Get routes with admin_only flag
curl -X GET http://localhost:3456/api/v1/routes \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" | jq '
  .[] | to_entries[] | select(.value.admin_only == true) | {
    route: .key,
    admin_only: .value.admin_only,
    path: .value.path,
    method: .value.method
  }'
```

### Example: Webhook Integration

**Standard token attempt** (will fail):
```bash
# This FAILS even with correct permissions
curl -X GET http://localhost:3456/api/v1/projects/1/webhooks \
  -H "Authorization: Bearer tk_standard_token_here"

# Response: 401 Unauthorized
```

**Admin token** (succeeds):
```bash
# This WORKS with admin level + permissions
curl -X GET http://localhost:3456/api/v1/projects/1/webhooks \
  -H "Authorization: Bearer tk_admin_token_here"

# Response: [{"id": 1, "url": "https://...", ...}]
```

### Backward Compatibility

- **Existing tokens**: All tokens created before admin-level feature default to `standard` level
- **No breaking changes**: Standard tokens continue to work for all non-admin routes
- **Opt-in**: Admin level must be explicitly selected, never automatic

---

## Next Steps

After this fix is complete:

1. **Phase 2**: Generate implementation tasks with `/speckit.tasks`
2. **Testing**: Add E2E tests in `frontend/cypress/` for token permission UI
3. **Documentation**: Update API documentation to clarify permission scopes
4. **Monitoring**: Add metrics for API token authentication failures by permission type

---

## Architecture Notes

**Why this approach?**

1. **Preserves service layer architecture**: No changes to business logic
2. **Uses established patterns**: Leverages existing APIRoute/CollectRoute infrastructure
3. **Backward compatible**: Old tokens continue working
4. **Defensive**: Explicit registrations take precedence over legacy detection
5. **Minimal scope**: <200 LOC changes, focused on infrastructure only

**What we're NOT doing**:

- ❌ Reverting to deprecated getRouteDetail "magic" detection
- ❌ Moving business logic from services back to models
- ❌ Breaking existing API tokens
- ❌ Modifying database schema
- ❌ Changing frontend components (just using existing data correctly)

This fix completes the partially-done route refactoring that was started during the service layer migration.
